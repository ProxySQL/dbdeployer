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
	"github.com/spf13/cobra"
)

func deploySandboxProxySQL(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	port, _ := flags.GetInt("port")
	adminUser, _ := flags.GetString("admin-user")
	adminPassword, _ := flags.GetString("admin-password")
	skipStart, _ := flags.GetBool("skip-start")

	p, err := providers.DefaultRegistry.Get("proxysql")
	if err != nil {
		common.Exitf(1, "ProxySQL provider not available: %s", err)
	}

	if _, err := p.FindBinary(""); err != nil {
		common.Exitf(1, "proxysql binary not found in PATH: %s", err)
	}

	sandboxHome := defaults.Defaults().SandboxHome
	sandboxDir := path.Join(sandboxHome, fmt.Sprintf("proxysql_%d", port))

	if common.DirExists(sandboxDir) {
		common.Exitf(1, "sandbox directory %s already exists", sandboxDir)
	}

	config := providers.SandboxConfig{
		Version:    "system",
		Dir:        sandboxDir,
		Port:       port,
		AdminPort:  port,
		Host:       "127.0.0.1",
		DbUser:     adminUser,
		DbPassword: adminPassword,
		Options:    map[string]string{},
	}

	_, err = p.CreateSandbox(config)
	if err != nil {
		common.Exitf(1, "error creating ProxySQL sandbox: %s", err)
	}

	if !skipStart {
		if err := p.StartSandbox(sandboxDir); err != nil {
			common.Exitf(1, "error starting ProxySQL: %s", err)
		}
	}

	fmt.Printf("ProxySQL sandbox deployed in %s (admin port: %d, mysql port: %d)\n", sandboxDir, port, port+1)
}

var deployProxySQLCmd = &cobra.Command{
	Use:   "proxysql",
	Short: "deploys a ProxySQL sandbox",
	Long: `proxysql deploys a standalone ProxySQL instance as a sandbox.
It creates a sandbox directory with configuration, start/stop scripts, and a
client script for administration.

Example:
    dbdeployer deploy proxysql
    dbdeployer deploy proxysql --port=16032
    dbdeployer deploy proxysql --admin-user=myadmin --admin-password=secret
`,
	Run: deploySandboxProxySQL,
}

func init() {
	deployCmd.AddCommand(deployProxySQLCmd)
	deployProxySQLCmd.Flags().Int("port", 6032, "ProxySQL admin port")
	deployProxySQLCmd.Flags().String("admin-user", "admin", "ProxySQL admin user")
	deployProxySQLCmd.Flags().String("admin-password", "admin", "ProxySQL admin password")
	deployProxySQLCmd.Flags().Bool("skip-start", false, "Do not start ProxySQL after deployment")
}
