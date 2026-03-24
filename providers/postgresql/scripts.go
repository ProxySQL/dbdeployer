package postgresql

import "fmt"

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
