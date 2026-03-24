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
}

const envPreamble = `#!/bin/bash
export LD_LIBRARY_PATH="%s"
unset PGDATA PGPORT PGHOST PGUSER PGDATABASE
`

func GenerateScripts(opts ScriptOptions) map[string]string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)

	return map[string]string{
		"start": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s start\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

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
