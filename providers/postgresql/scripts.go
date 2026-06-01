package postgresql

import (
	"fmt"
	"strings"
)

const (
	SandboxUser      = "rsandbox"
	SandboxPassword  = "rsandbox"
	SandboxDatabase  = "rsandbox"
)

type ScriptOptions struct {
	SandboxDir string
	DataDir    string
	BinDir     string
	LibDir     string
	Port       int
	LogFile    string
}

const envPreamble = `#!/bin/bash
export LD_LIBRARY_PATH="%s"
unset PGDATA PGPORT PGHOST PGUSER PGDATABASE
`

// ShellEnvPreamble returns LD_LIBRARY_PATH setup lines for sandbox shell scripts.
func ShellEnvPreamble(libDir string) string {
	return fmt.Sprintf("export LD_LIBRARY_PATH=\"%s\"\nunset PGDATA PGPORT PGHOST PGUSER PGDATABASE\n", libDir)
}

// SandboxUserGrantsSQL returns idempotent SQL to create the sandbox user and database.
func SandboxUserGrantsSQL() string {
	return fmt.Sprintf(`DO $$ BEGIN
    CREATE USER %s WITH PASSWORD '%s';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
SELECT 'CREATE DATABASE %s'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '%s')\gexec
GRANT CONNECT ON DATABASE postgres TO %s;
GRANT CONNECT ON DATABASE %s TO %s;
GRANT CREATE ON SCHEMA public TO %s;`, SandboxUser, SandboxPassword, SandboxDatabase, SandboxDatabase, SandboxUser, SandboxDatabase, SandboxUser, SandboxUser)
}

// SandboxDatabaseGrantsSQL returns grants applied inside the sandbox database.
func SandboxDatabaseGrantsSQL() string {
	return fmt.Sprintf("GRANT CREATE ON SCHEMA public TO %s;", SandboxUser)
}

func GenerateInitSandboxUserScript(opts ScriptOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	return fmt.Sprintf(`%srecovery=$("%s/psql" -h 127.0.0.1 -p %d -U postgres -tAc "SELECT pg_is_in_recovery();")
if [ "$recovery" = "t" ]; then
    exit 0
fi
"%s/psql" -h 127.0.0.1 -p %d -U postgres -v ON_ERROR_STOP=1 <<'EOSQL'
%s
EOSQL
"%s/psql" -h 127.0.0.1 -p %d -U postgres -d %s -v ON_ERROR_STOP=1 <<'EOSQL'
%s
EOSQL
`, preamble, opts.BinDir, opts.Port, opts.BinDir, opts.Port, SandboxUserGrantsSQL(), opts.BinDir, opts.Port, SandboxDatabase, SandboxDatabaseGrantsSQL())
}

func GenerateScripts(opts ScriptOptions) map[string]string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	initUser := GenerateInitSandboxUserScript(opts)

	return map[string]string{
		"init_sandbox_user": initUser,

		"start": fmt.Sprintf("%s%s/pg_ctl -w -D %s -l %s start\nbash %s/init_sandbox_user\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile, opts.SandboxDir),

		"stop": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast\n",
			preamble, opts.BinDir, opts.DataDir),

		"status": fmt.Sprintf("%s%s/pg_ctl -D %s status\n",
			preamble, opts.BinDir, opts.DataDir),

		"restart": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s restart\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"use": fmt.Sprintf("%s%s/psql -h 127.0.0.1 -p %d -U postgres \"$@\"\n",
			preamble, opts.BinDir, opts.Port),

		"bench": GenerateBenchScript(opts),

		"clear": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast 2>/dev/null\nrm -rf %s\n%s/initdb -D %s --auth=trust --username=postgres\necho \"Sandbox cleared.\"\n",
			preamble, opts.BinDir, opts.DataDir, opts.DataDir, opts.BinDir, opts.DataDir),
	}
}

// GenerateBenchScript returns a pgbench helper for sandbox demos.
// With no arguments it initializes a small pgbench dataset (if needed) and
// runs a 30-second TPC-B style benchmark. Any arguments are passed to pgbench.
func GenerateBenchScript(opts ScriptOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	return fmt.Sprintf(`%sset -e
PGBENCH="%s/pgbench"
PSQL="%s/psql"
export PGPASSWORD=%s
HOST=127.0.0.1
PORT=%d
USER=%s
DB=%s
SCALE=5
CLIENTS=10
JOBS=2
TIME=30

if [ ! -x "$PGBENCH" ]; then
    echo "pgbench not found at $PGBENCH" >&2
    exit 1
fi

IS_RECOVERY=$("$PSQL" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -tAc "SELECT pg_is_in_recovery();" 2>/dev/null || echo "f")

if [ $# -gt 0 ]; then
    exec "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" "$@"
fi

TABLES=$("$PSQL" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -tAc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'pgbench_%%';")
if [ "${TABLES:-0}" -lt 4 ]; then
    if [ "$IS_RECOVERY" = "t" ]; then
        echo "Error: pgbench tables are not initialized, and this server is a read-only replica." >&2
        echo "Please run the bench script on the primary server first to initialize the database." >&2
        exit 1
    fi
    echo "# initializing pgbench tables (scale=${SCALE})"
    "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -i -s "$SCALE"
fi

if [ "$IS_RECOVERY" = "t" ]; then
    echo "# running pgbench (select-only mode on replica): ${CLIENTS} clients, ${JOBS} jobs, ${TIME}s"
    exec "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c "$CLIENTS" -j "$JOBS" -T "$TIME" -S
else
    echo "# running pgbench: ${CLIENTS} clients, ${JOBS} jobs, ${TIME}s"
    exec "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c "$CLIENTS" -j "$JOBS" -T "$TIME"
fi
`, preamble, opts.BinDir, opts.BinDir, SandboxPassword, opts.Port, SandboxUser, SandboxDatabase)
}

// TopologyBenchOptions configures a topology-level bench script (replication root).
type TopologyBenchOptions struct {
	BinDir      string
	LibDir      string
	PrimaryPort int
	ProxyPort   int // 0 when ProxySQL is not deployed
}

// GenerateTopologyBenchScript returns a bench helper at the replication topology root.
// Pass --proxysql as the first argument to benchmark through the ProxySQL proxy port.
func GenerateTopologyBenchScript(opts TopologyBenchOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	proxyPortLine := "PROXY_PORT=0"
	if opts.ProxyPort > 0 {
		proxyPortLine = fmt.Sprintf("PROXY_PORT=%d", opts.ProxyPort)
	}
	return fmt.Sprintf(`%sset -e
PGBENCH="%s/pgbench"
PSQL="%s/psql"
export PGPASSWORD=%s
HOST=127.0.0.1
PRIMARY_PORT=%d
%s
USER=%s
DB=%s
SCALE=5
CLIENTS=10
JOBS=2
TIME=30

PORT=$PRIMARY_PORT
if [ "${1:-}" = "--proxysql" ]; then
    shift
    if [ "$PROXY_PORT" -eq 0 ]; then
        echo "ProxySQL is not deployed in this topology" >&2
        exit 1
    fi
    PORT=$PROXY_PORT
    echo "# benchmarking via ProxySQL proxy port ${PORT}"
fi

if [ ! -x "$PGBENCH" ]; then
    echo "pgbench not found at $PGBENCH" >&2
    exit 1
fi

if [ $# -gt 0 ]; then
    exec "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" "$@"
fi

TABLES=$("$PSQL" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -tAc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'pgbench_%%';")
if [ "${TABLES:-0}" -lt 4 ]; then
    echo "# initializing pgbench tables on primary (scale=${SCALE})"
    "$PGBENCH" -h "$HOST" -p "$PRIMARY_PORT" -U "$USER" -d "$DB" -i -s "$SCALE"
fi

echo "# running pgbench on port ${PORT}: ${CLIENTS} clients, ${JOBS} jobs, ${TIME}s"
exec "$PGBENCH" -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c "$CLIENTS" -j "$JOBS" -T "$TIME"
`, preamble, opts.BinDir, opts.BinDir, SandboxPassword, opts.PrimaryPort, proxyPortLine, SandboxUser, SandboxDatabase)
}

func GenerateCheckReplicationScript(opts ScriptOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	return fmt.Sprintf(`%s%s/psql -h 127.0.0.1 -p %d -U postgres -c \
  "SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication;"
`, preamble, opts.BinDir, opts.Port)
}

func GenerateCheckRecoveryScript(opts ScriptOptions, replicaPorts []int) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	var b strings.Builder
	b.WriteString(preamble)
	for _, port := range replicaPorts {
		b.WriteString(fmt.Sprintf("echo \"=== Replica port %d ===\"\n", port))
		b.WriteString(fmt.Sprintf("%s/psql -h 127.0.0.1 -p %d -U postgres -c \"SELECT pg_is_in_recovery();\"\n", opts.BinDir, port))
	}
	return b.String()
}
