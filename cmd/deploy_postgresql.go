// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"path"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/globals"
	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/ProxySQL/dbdeployer/providers/postgresql"
	"github.com/spf13/cobra"
)

func deploySandboxPostgreSQL(cmd *cobra.Command, args []string) {
	version := args[0]
	flags := cmd.Flags()
	skipStart, _ := flags.GetBool("skip-start")

	p, err := providers.DefaultRegistry.Get("postgresql")
	if err != nil {
		common.Exitf(1, "PostgreSQL provider not available: %s", err)
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "invalid version: %s", err)
	}

	// Binary discovery. The --sandbox-binary flag is shared with the MySQL
	// path and defaults to ~/opt/mysql, so it is only applied to PostgreSQL
	// when explicitly provided. When set, binaries are expected at
	// <sandbox-binary>/<version>/bin/postgres and that directory becomes the
	// sandbox basedir (bin/lib/share). When not set, fall back to the strict
	// historical behavior: ~/opt/postgresql/<version>.
	options := map[string]string{}
	if flags.Changed(globals.SandboxBinaryLabel) {
		sandboxBinary, _ := flags.GetString(globals.SandboxBinaryLabel)
		sbAbs, err := common.AbsolutePath(sandboxBinary)
		if err != nil {
			common.Exitf(1, "error defining absolute path for --%s: %s", globals.SandboxBinaryLabel, err)
		}
		basedir := path.Join(sbAbs, version)
		binPath := path.Join(basedir, "bin", "postgres")
		if !common.FileExists(binPath) {
			common.Exitf(1, "PostgreSQL binary not found at %s", binPath)
		}
		options["basedir"] = basedir
	} else if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "PostgreSQL binaries not found: %s\nRun: dbdeployer unpack --provider=postgresql <server.deb> <client.deb>", err)
	}

	// sandbox-home: honor --sandbox-home. The flag is registered on the root
	// command with defaults.Defaults().SandboxHome as its default, so reading it
	// directly already falls back to the configured default when not provided.
	sandboxHome, err := getAbsolutePathFromFlag(cmd, globals.SandboxHomeLabel)
	if err != nil {
		common.Exitf(1, "error defining absolute path for --%s: %s", globals.SandboxHomeLabel, err)
	}

	// port: honor --port; if not provided (<=0), compute it from the version and
	// pick the next free port, matching the previous behavior.
	port, _ := flags.GetInt(globals.PortLabel)
	if port <= 0 {
		port, err = postgresql.VersionToPort(version)
		if err != nil {
			common.Exitf(1, "error computing port: %s", err)
		}
		installedPorts, _ := common.GetInstalledPorts(sandboxHome)
		freePort, portErr := common.FindFreePort(port, installedPorts, 1)
		if portErr == nil {
			port = freePort
		}
	}

	// sandbox-directory: honor --sandbox-directory; fall back to pg_sandbox_<port>.
	sandboxDirectory, _ := flags.GetString(globals.SandboxDirectoryLabel)
	if sandboxDirectory == "" {
		sandboxDirectory = fmt.Sprintf("pg_sandbox_%d", port)
	}
	sandboxDir := path.Join(sandboxHome, sandboxDirectory)

	if common.DirExists(sandboxDir) {
		common.Exitf(1, "sandbox directory %s already exists", sandboxDir)
	}

	// db-user / db-password: these flags are registered on "deploy" with the
	// MySQL default ("msandbox"). To preserve the PostgreSQL default
	// ("postgres" / "") when they are not explicitly provided, only honor them
	// when changed on the command line.
	dbUser := "postgres"
	if flags.Changed(globals.DbUserLabel) {
		dbUser, _ = flags.GetString(globals.DbUserLabel)
	}
	dbPassword := ""
	if flags.Changed(globals.DbPasswordLabel) {
		dbPassword, _ = flags.GetString(globals.DbPasswordLabel)
	}

	config := providers.SandboxConfig{
		Version:    version,
		Dir:        sandboxDir,
		Port:       port,
		Host:       "127.0.0.1",
		DbUser:     dbUser,
		DbPassword: dbPassword,
		Options:    options,
	}

	if _, err := p.CreateSandbox(config); err != nil {
		common.Exitf(1, "error creating PostgreSQL sandbox: %s", err)
	}

	if !skipStart {
		if err := p.StartSandbox(sandboxDir); err != nil {
			common.Exitf(1, "error starting PostgreSQL: %s", err)
		}
	}

	fmt.Printf("PostgreSQL %s sandbox deployed in %s (port: %d)\n", version, sandboxDir, port)
}

var deployPostgreSQLCmd = &cobra.Command{
	Use:   "postgresql version",
	Short: "deploys a PostgreSQL sandbox",
	Long: `postgresql deploys a standalone PostgreSQL instance as a sandbox.
It creates a sandbox directory with data, configuration, start/stop scripts, and a
psql client script.

Requires PostgreSQL binaries to be extracted first:
    dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb

Example:
    dbdeployer deploy postgresql 16.13
    dbdeployer deploy postgresql 17.1 --skip-start
`,
	Args: cobra.ExactArgs(1),
	Run:  deploySandboxPostgreSQL,
}

func init() {
	deployCmd.AddCommand(deployPostgreSQLCmd)
	deployPostgreSQLCmd.Flags().Bool("skip-start", false, "Do not start PostgreSQL after deployment")
}
