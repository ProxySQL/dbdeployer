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

		"start": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s start\nbash %s/init_sandbox_user\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile, opts.SandboxDir),

		"stop": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast\n",
			preamble, opts.BinDir, opts.DataDir),

		"status": fmt.Sprintf("%s%s/pg_ctl -D %s status\n",
			preamble, opts.BinDir, opts.DataDir),

		"restart": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s restart\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"use": fmt.Sprintf("%s%s/psql -h 127.0.0.1 -p %d -U postgres \"$@\"\n",
			preamble, opts.BinDir, opts.Port),

		"clear": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast 2>/dev/null\nrm -rf %s\n%s/initdb -D %s --auth=trust --username=postgres\necho \"Sandbox cleared.\"\n",
			preamble, opts.BinDir, opts.DataDir, opts.DataDir, opts.BinDir, opts.DataDir),
	}
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
