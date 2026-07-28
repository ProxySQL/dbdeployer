package postgresql

import (
	"fmt"
	"strings"
)

type ScriptOptions struct {
	SandboxDir string
	DataDir    string
	BinDir     string
	LibDir     string
	Port       int
	LogFile    string
	// DbUser is the database user embedded in the generated psql/initdb
	// invocations. When empty, it defaults to the PostgreSQL superuser
	// "postgres" to preserve the historical sandbox behavior.
	DbUser string
}

const envPreamble = `#!/bin/bash
export LD_LIBRARY_PATH="%s"
unset PGDATA PGPORT PGHOST PGUSER PGDATABASE
`

func GenerateScripts(opts ScriptOptions) map[string]string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	dbUser := opts.DbUser
	if dbUser == "" {
		dbUser = "postgres"
	}

	return map[string]string{
		"start": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s start\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"stop": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast\n",
			preamble, opts.BinDir, opts.DataDir),

		"status": fmt.Sprintf("%s%s/pg_ctl -D %s status\n",
			preamble, opts.BinDir, opts.DataDir),

		"restart": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s restart\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"use": fmt.Sprintf("%s%s/psql -h 127.0.0.1 -p %d -U %s \"$@\"\n",
			preamble, opts.BinDir, opts.Port, dbUser),

		"clear": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast 2>/dev/null\nrm -rf %s\n%s/initdb -D %s --auth=trust --username=%s\necho \"Sandbox cleared.\"\n",
			preamble, opts.BinDir, opts.DataDir, opts.DataDir, opts.BinDir, opts.DataDir, dbUser),
	}
}

func GenerateCheckReplicationScript(opts ScriptOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	dbUser := opts.DbUser
	if dbUser == "" {
		dbUser = "postgres"
	}
	return fmt.Sprintf(`%s%s/psql -h 127.0.0.1 -p %d -U %s -c \
  "SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication;"
`, preamble, opts.BinDir, opts.Port, dbUser)
}

func GenerateCheckRecoveryScript(opts ScriptOptions, replicaPorts []int) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	dbUser := opts.DbUser
	if dbUser == "" {
		dbUser = "postgres"
	}
	var b strings.Builder
	b.WriteString(preamble)
	for _, port := range replicaPorts {
		b.WriteString(fmt.Sprintf("echo \"=== Replica port %d ===\"\n", port))
		b.WriteString(fmt.Sprintf("%s/psql -h 127.0.0.1 -p %d -U %s -c \"SELECT pg_is_in_recovery();\"\n", opts.BinDir, port, dbUser))
	}
	return b.String()
}

// GenerateTestReplicationScript emits a topology-level test_replication
// script that verifies replication is actually working: it creates a
// table on the primary, inserts rows, captures the primary's WAL LSN,
// then queries each replica to confirm the LSN matches and the row
// count is in sync. Modeled on the MySQL `test_replication` template
// (sandbox/templates/replication/test_replication.gotxt) but adapted to
// PostgreSQL primitives (pg_current_wal_lsn / pg_last_wal_replay_lsn).
//
// numReplicas is the count of replica sub-sandboxes laid down by
// `deploy replication --provider=postgresql` (i.e. one less than --nodes).
// The generated script invokes ./primary/use and ./replicaN/use, which
// the PostgreSQL provider already creates as per-sandbox `use` scripts.
func GenerateTestReplicationScript(opts ScriptOptions, numReplicas int) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	var b strings.Builder
	b.WriteString(preamble)
	b.WriteString(fmt.Sprintf(`SBDIR=%s
cd "$SBDIR"

if [ ! -x ./primary/use ]; then
    echo "# No primary found at ./primary/use"
    exit 1
fi
PRIMARY=./primary/use

FAILED=0
PASSED=0

function ok_equal {
    fact="$1"
    expected="$2"
    msg="$3"
    if [ "$fact" = "$expected" ]; then
        echo "ok - (expected: <$expected>) - $msg"
        PASSED=$((PASSED+1))
    else
        echo "not ok - (expected: <$expected> found: <$fact>) - $msg"
        FAILED=$((FAILED+1))
    fi
}

function test_summary {
    TESTS=$((PASSED+FAILED))
    printf "# Tests : %%5d\n" $TESTS
    printf "# Passed: %%5d\n" $PASSED
    printf "# Failed: %%5d\n" $FAILED
    if [ "$FAILED" != "0" ]; then exit 1; fi
    exit 0
}

YYYY_MM=$(date +%%Y-%%m)
echo "# Creating and populating table on primary..."
$PRIMARY -qc 'DROP TABLE IF EXISTS dbdeployer_test_replication'
$PRIMARY -qc 'CREATE TABLE dbdeployer_test_replication (i INT NOT NULL PRIMARY KEY, msg VARCHAR(50), d DATE, t TIME, dt TIMESTAMP)'
for N in $(seq -f '%%02.0f' 1 20); do
    $PRIMARY -qc "INSERT INTO dbdeployer_test_replication VALUES ($N, 'test sandbox $N', '${YYYY_MM}-$N', '11:23:$N', '${YYYY_MM}-01 12:34:$N')"
done

PRIMARY_RECS=$($PRIMARY -Atqc 'SELECT count(*) FROM dbdeployer_test_replication')
PRIMARY_LSN=$($PRIMARY -Atqc 'SELECT pg_current_wal_lsn()')
echo "# Primary: rows=$PRIMARY_RECS  LSN=$PRIMARY_LSN"

# Allow a brief window for streaming replication to catch up
sleep 1

`, opts.SandboxDir))

	for n := 1; n <= numReplicas; n++ {
		b.WriteString(fmt.Sprintf(`echo "# Testing replica %d"
if [ -x ./replica%d/use ]; then
    REPLICA_LSN=$(./replica%d/use -Atqc 'SELECT pg_last_wal_replay_lsn()')
    REPLICA_RECS=$(./replica%d/use -Atqc 'SELECT count(*) FROM dbdeployer_test_replication')
    REPLICA_IS_REPLICA=$(./replica%d/use -Atqc 'SELECT pg_is_in_recovery()')
    ok_equal "$REPLICA_IS_REPLICA" "t" "replica %d is in recovery mode"
    ok_equal "$REPLICA_LSN" "$PRIMARY_LSN" "replica %d LSN matches primary"
    ok_equal "$REPLICA_RECS" "$PRIMARY_RECS" "replica %d row count matches primary"
else
    echo "not ok - replica %d/use not found"
    FAILED=$((FAILED+1))
fi

`, n, n, n, n, n, n, n, n, n))
	}

	b.WriteString("test_summary\n")
	return b.String()
}
