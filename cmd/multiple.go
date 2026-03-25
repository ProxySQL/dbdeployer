// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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
	"os"
	"path"
	"strings"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/defaults"
	"github.com/ProxySQL/dbdeployer/globals"
	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/ProxySQL/dbdeployer/providers/postgresql"
	"github.com/ProxySQL/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func deployMultipleNonMySQL(cmd *cobra.Command, args []string, providerName string) {
	flags := cmd.Flags()
	version := args[0]
	nodes, _ := flags.GetInt(globals.NodesLabel)

	p, err := providers.DefaultRegistry.Get(providerName)
	if err != nil {
		common.Exitf(1, "provider error: %s", err)
	}

	flavor, _ := flags.GetString(globals.FlavorLabel)
	if flavor != "" {
		common.Exitf(1, "--flavor is only valid with --provider=mysql")
	}

	if !providers.ContainsString(p.SupportedTopologies(), "multiple") {
		common.Exitf(1, "provider %q does not support topology \"multiple\"\nSupported topologies: %s",
			providerName, strings.Join(p.SupportedTopologies(), ", "))
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "version validation failed: %s", err)
	}

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "binaries not found: %s", err)
	}

	basePort := p.DefaultPorts().BasePort
	if providerName == "postgresql" {
		basePort, _ = postgresql.VersionToPort(version)
	}

	sandboxHome := defaults.Defaults().SandboxHome
	topologyDir := path.Join(sandboxHome, fmt.Sprintf("%s_multi_%d", providerName, basePort))
	if common.DirExists(topologyDir) {
		common.Exitf(1, "sandbox directory %s already exists", topologyDir)
	}
	if err := os.MkdirAll(topologyDir, 0755); err != nil {
		common.Exitf(1, "error creating topology directory %s: %s", topologyDir, err)
	}

	skipStart, _ := flags.GetBool(globals.SkipStartLabel)

	for i := 1; i <= nodes; i++ {
		port := basePort + i
		freePort, err := common.FindFreePort(port, []int{}, 1)
		if err == nil {
			port = freePort
		}

		nodeDir := path.Join(topologyDir, fmt.Sprintf("node%d", i))
		config := providers.SandboxConfig{
			Version: version,
			Dir:     nodeDir,
			Port:    port,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}

		if _, err := p.CreateSandbox(config); err != nil {
			common.Exitf(1, "error creating node %d: %s", i, err)
		}

		if !skipStart {
			if err := p.StartSandbox(nodeDir); err != nil {
				common.Exitf(1, "error starting node %d: %s", i, err)
			}
		}

		fmt.Printf("  Node %d deployed in %s (port: %d)\n", i, nodeDir, port)
	}

	fmt.Printf("%s multiple sandbox (%d nodes) deployed in %s\n", providerName, nodes, topologyDir)
}

func multipleSandbox(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	providerName, _ := flags.GetString(globals.ProviderLabel)

	if providerName != "mysql" {
		deployMultipleNonMySQL(cmd, args, providerName)
		return
	}

	var sd sandbox.SandboxDef
	common.CheckOrigin(args)
	sd, err := fillSandboxDefinition(cmd, args, false)
	common.ErrCheckExitf(err, 1, "error filling sandbox definition")
	// Validate version with provider
	// TODO: Phase 2b — determine provider from sd.Flavor instead of hardcoding "mysql"
	p, provErr := providers.DefaultRegistry.Get("mysql")
	if provErr != nil {
		common.Exitf(1, "provider error: %s", provErr)
	}
	if provErr = p.ValidateVersion(sd.Version); provErr != nil {
		common.Exitf(1, "version validation failed: %s", provErr)
	}
	nodes, _ := flags.GetInt(globals.NodesLabel)
	sd.SBType = "multiple"
	origin := args[0]
	if args[0] != sd.BasedirName {
		origin = sd.BasedirName
	}
	_, err = sandbox.CreateMultipleSandbox(sd, origin, nodes)
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}
}

var multipleCmd = &cobra.Command{
	Use:   "multiple MySQL-Version",
	Short: "create multiple sandbox",
	Long: `Creates several sandboxes of the same version,
without any replication relationship.
For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	Run: multipleSandbox,
	Example: `
	$ dbdeployer deploy multiple 5.7.21
	`,
	Annotations: map[string]string{"export": makeExportArgs(globals.ExportVersionDir, 1)},
}

func init() {
	deployCmd.AddCommand(multipleCmd)
	multipleCmd.PersistentFlags().IntP(globals.NodesLabel, "n", globals.NodesValue, "How many nodes will be installed")
	multipleCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
}
