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
	"github.com/ProxySQL/dbdeployer/defaults"
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

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "PostgreSQL binaries not found: %s\nRun: dbdeployer unpack --provider=postgresql <server.deb> <client.deb>", err)
	}

	port, err := postgresql.VersionToPort(version)
	if err != nil {
		common.Exitf(1, "error computing port: %s", err)
	}
	freePort, portErr := common.FindFreePort(port, []int{}, 1)
	if portErr == nil {
		port = freePort
	}

	sandboxHome := defaults.Defaults().SandboxHome
	sandboxDir := path.Join(sandboxHome, fmt.Sprintf("pg_sandbox_%d", port))

	if common.DirExists(sandboxDir) {
		common.Exitf(1, "sandbox directory %s already exists", sandboxDir)
	}

	config := providers.SandboxConfig{
		Version:    version,
		Dir:        sandboxDir,
		Port:       port,
		Host:       "127.0.0.1",
		DbUser:     "postgres",
		DbPassword: "",
		Options:    map[string]string{},
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
