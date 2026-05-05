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

func deployReplicationNonMySQL(cmd *cobra.Command, args []string, providerName string) {
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

	if !providers.ContainsString(p.SupportedTopologies(), "replication") {
		common.Exitf(1, "provider %q does not support topology \"replication\"\nSupported topologies: %s",
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
	topologyDir := path.Join(sandboxHome, fmt.Sprintf("%s_repl_%d", providerName, basePort))
	if common.DirExists(topologyDir) {
		common.Exitf(1, "sandbox directory %s already exists", topologyDir)
	}
	if err := os.MkdirAll(topologyDir, 0755); err != nil {
		common.Exitf(1, "error creating topology directory %s: %s", topologyDir, err)
	}

	primaryPort := basePort

	// Create and start primary with replication options
	primaryDir := path.Join(topologyDir, "primary")
	primaryConfig := providers.SandboxConfig{
		Version: version,
		Dir:     primaryDir,
		Port:    primaryPort,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{"replication": "true"},
	}

	if _, err := p.CreateSandbox(primaryConfig); err != nil {
		common.Exitf(1, "error creating primary: %s", err)
	}

	skipStart, _ := flags.GetBool(globals.SkipStartLabel)
	if !skipStart {
		if err := p.StartSandbox(primaryDir); err != nil {
			common.Exitf(1, "error starting primary: %s", err)
		}
	}

	fmt.Printf("  Primary deployed in %s (port: %d)\n", primaryDir, primaryPort)

	primaryInfo := providers.SandboxInfo{Dir: primaryDir, Port: primaryPort, Status: "running"}

	// Create replicas sequentially
	var replicaPorts []int
	for i := 1; i <= nodes-1; i++ {
		replicaPort := primaryPort + i
		freePort, err := common.FindFreePort(replicaPort, []int{}, 1)
		if err == nil {
			replicaPort = freePort
		}

		replicaDir := path.Join(topologyDir, fmt.Sprintf("replica%d", i))
		replicaConfig := providers.SandboxConfig{
			Version: version,
			Dir:     replicaDir,
			Port:    replicaPort,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}

		if _, err := p.CreateReplica(primaryInfo, replicaConfig); err != nil {
			// Cleanup on failure
			_ = p.StopSandbox(primaryDir)
			for j := 1; j < i; j++ {
				_ = p.StopSandbox(path.Join(topologyDir, fmt.Sprintf("replica%d", j)))
			}
			common.Exitf(1, "error creating replica %d: %s", i, err)
		}

		replicaPorts = append(replicaPorts, replicaPort)
		fmt.Printf("  Replica %d deployed in %s (port: %d)\n", i, replicaDir, replicaPort)
	}

	// Generate monitoring scripts
	home, _ := os.UserHomeDir()
	basedir := path.Join(home, "opt", "postgresql", version)
	binDir := path.Join(basedir, "bin")
	libDir := path.Join(basedir, "lib")

	scriptOpts := postgresql.ScriptOptions{
		BinDir: binDir,
		LibDir: libDir,
		Port:   primaryPort,
	}

	checkReplScript := postgresql.GenerateCheckReplicationScript(scriptOpts)
	if err := os.WriteFile(path.Join(topologyDir, "check_replication"), []byte(checkReplScript), 0755); err != nil { //nolint:gosec // scripts must be executable
		fmt.Printf("WARNING: could not write check_replication script: %s\n", err)
	}

	checkRecovScript := postgresql.GenerateCheckRecoveryScript(scriptOpts, replicaPorts)
	if err := os.WriteFile(path.Join(topologyDir, "check_recovery"), []byte(checkRecovScript), 0755); err != nil { //nolint:gosec // scripts must be executable
		fmt.Printf("WARNING: could not write check_recovery script: %s\n", err)
	}

	// Handle --with-proxysql
	withProxySQL, _ := flags.GetBool("with-proxysql")
	if withProxySQL {
		if !providers.ContainsString(providers.CompatibleAddons["proxysql"], providerName) {
			common.Exitf(1, "--with-proxysql is not compatible with provider %q", providerName)
		}
		err := sandbox.DeployProxySQLForTopology(topologyDir, primaryPort, replicaPorts, 0, "127.0.0.1", providerName)
		if err != nil {
			common.Exitf(1, "ProxySQL deployment failed: %s", err)
		}
	}

	fmt.Printf("%s replication sandbox (1 primary + %d replicas) deployed in %s\n",
		providerName, nodes-1, topologyDir)
}

func replicationSandbox(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	providerName, _ := flags.GetString(globals.ProviderLabel)

	if providerName != "mysql" {
		deployReplicationNonMySQL(cmd, args, providerName)
		return
	}

	var sd sandbox.SandboxDef
	var semisync bool
	common.CheckOrigin(args)
	sd, err := fillSandboxDefinition(cmd, args, false)
	common.ErrCheckExitf(err, 1, "error filling sandbox definition : %s", err)
	// Validate version with provider
	// TODO: Phase 2b — determine provider from sd.Flavor instead of hardcoding "mysql"
	p, provErr := providers.DefaultRegistry.Get("mysql")
	if provErr != nil {
		common.Exitf(1, "provider error: %s", provErr)
	}
	if provErr = p.ValidateVersion(sd.Version); provErr != nil {
		common.Exitf(1, "version validation failed: %s", provErr)
	}
	if sd.Flavor == common.TiDbFlavor {
		common.Exitf(1, "flavor '%s' is not suitable to create replication sandboxes", common.TiDbFlavor)
	}
	sd.ReplOptions = sandbox.SingleTemplates[globals.TmplReplicationOptions].Contents
	semisync, _ = flags.GetBool(globals.SemiSyncLabel)
	ndbNodes, _ := flags.GetInt(globals.NdbNodesLabel)
	nodes, _ := flags.GetInt(globals.NodesLabel)
	topology, _ := flags.GetString(globals.TopologyLabel)
	masterIp, _ := flags.GetString(globals.MasterIpLabel)
	masterList, _ := flags.GetString(globals.MasterListLabel)
	slaveList, _ := flags.GetString(globals.SlaveListLabel)
	withProxySQL, _ := flags.GetBool("with-proxysql")
	if withProxySQL && topology == globals.NdbLabel {
		common.Exitf(1, "--with-proxysql is not supported with topology %q", topology)
	}
	sd.SinglePrimary, _ = flags.GetBool(globals.SinglePrimaryLabel)
	replHistoryDir, _ := flags.GetBool(globals.ReplHistoryDirLabel)
	if replHistoryDir {
		sd.HistoryDir = "REPL_DIR"
	}
	if topology != globals.FanInLabel && topology != globals.AllMastersLabel {
		masterList = ""
		slaveList = ""
	}
	if semisync {
		if topology != globals.MasterSlaveLabel {
			common.Exit(1, "--semi-sync is only available with master/slave topology")
		}
		// 5.5.1

		// isMinimumSync, err := common.GreaterOrEqualVersion(sd.Version, globals.MinimumSemiSyncVersion)
		isMinimumSync, err := common.HasCapability(sd.Flavor, common.SemiSynch, sd.Version)
		common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
		if isMinimumSync {
			// MySQL 9.2+ removed semisync_master.so/semisync_slave.so;
			// use semisync_source.so/semisync_replica.so instead.
			useSourcePlugin, _ := common.GreaterOrEqualVersion(sd.Version, globals.MinimumSemisyncSourcePluginVersion)
			if useSourcePlugin {
				sd.SemiSyncOptions = sandbox.SingleTemplates[globals.TmplSemisyncSourceOptions].Contents
			} else {
				sd.SemiSyncOptions = sandbox.SingleTemplates[globals.TmplSemisyncMasterOptions].Contents
			}
		} else {
			common.Exitf(1, "--%s requires version %s+",
				globals.SemiSyncLabel,
				common.IntSliceToDottedString(globals.MinimumSemiSyncVersion))
		}
	}
	if sd.SinglePrimary && topology != globals.GroupLabel {
		common.Exitf(1, "option '%s' can only be used with '%s' topology ",
			globals.SinglePrimaryLabel,
			globals.GroupLabel)
	}
	if ndbNodes != globals.NdbNodesValue && topology != globals.NdbLabel {
		common.Exitf(1, "option '%s' can only be used with '%s' topology ",
			globals.NdbNodesLabel,
			globals.NdbLabel)

	}
	skipRouter, _ := flags.GetBool(globals.SkipRouterLabel)

	origin := args[0]
	if args[0] != sd.BasedirName {
		origin = sd.BasedirName
	}
	err = sandbox.CreateReplicationSandbox(sd, origin,
		sandbox.ReplicationData{
			Topology:   topology,
			Nodes:      nodes,
			NdbNodes:   ndbNodes,
			MasterIp:   masterIp,
			MasterList: masterList,
			SlaveList:  slaveList,
			SkipRouter: skipRouter})
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}

	if withProxySQL {
		// Determine the sandbox directory that was created
		var sandboxDir string
		switch topology {
		case globals.InnoDBClusterLabel:
			sandboxDir = path.Join(sd.SandboxDir, defaults.Defaults().InnoDBClusterPrefix+common.VersionToName(origin))
		case globals.PxcLabel:
			sandboxDir = path.Join(sd.SandboxDir, defaults.Defaults().PxcPrefix+common.VersionToName(origin))
		case globals.GaleraLabel:
			sandboxDir = path.Join(sd.SandboxDir, defaults.Defaults().GaleraPrefix+common.VersionToName(origin))
		case globals.NdbLabel:
			sandboxDir = path.Join(sd.SandboxDir, defaults.Defaults().NdbPrefix+common.VersionToName(origin))
		default:
			sandboxDir = path.Join(sd.SandboxDir, defaults.Defaults().MasterSlavePrefix+common.VersionToName(origin))
		}
		var masterPort int
		var slavePorts []int

		if topology == globals.InnoDBClusterLabel || topology == globals.PxcLabel || topology == globals.GaleraLabel {
			// InnoDB Cluster: node1 is primary, node2..N are secondaries
			primaryDesc, err := common.ReadSandboxDescription(path.Join(sandboxDir, fmt.Sprintf("%s%d", defaults.Defaults().NodePrefix, 1)))
			if err != nil {
				common.Exitf(1, "could not read primary (node1) sandbox description: %s", err)
			}
			masterPort = primaryDesc.Port[0]
			for i := 2; i <= nodes; i++ {
				nodeDir := path.Join(sandboxDir, fmt.Sprintf("%s%d", defaults.Defaults().NodePrefix, i))
				nodeDesc, err := common.ReadSandboxDescription(nodeDir)
				if err != nil {
					common.Exitf(1, "could not read node%d sandbox description: %s", i, err)
				}
				slavePorts = append(slavePorts, nodeDesc.Port[0])
			}
		} else {
			// Standard replication: master + node1..N-1 as slaves
			masterDesc, err := common.ReadSandboxDescription(path.Join(sandboxDir, defaults.Defaults().MasterName))
			if err != nil {
				common.Exitf(1, "could not read master sandbox description: %s", err)
			}
			masterPort = masterDesc.Port[0]
			for i := 1; i < nodes; i++ {
				nodeDir := path.Join(sandboxDir, fmt.Sprintf("%s%d", defaults.Defaults().NodePrefix, i))
				nodeDesc, err := common.ReadSandboxDescription(nodeDir)
				if err != nil {
					common.Exitf(1, "could not read node%d sandbox description: %s", i, err)
				}
				slavePorts = append(slavePorts, nodeDesc.Port[0])
			}
		}

		err = sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, slavePorts, 0, "127.0.0.1", "", topology)
		if err != nil {
			common.Exitf(1, "ProxySQL deployment failed: %s", err)
		}
	}
}

var replicationCmd = &cobra.Command{
	Use: "replication MySQL-Version",
	//Args:  cobra.ExactArgs(1),
	Short: "create replication sandbox",
	Long: `The replication command allows you to deploy several nodes in replication.
Allowed topologies are "master-slave" for all versions, and  "group", "all-masters", "fan-in"
for  5.7.17+.
Topologies "pxc", "galera", and "ndb" are available for binaries of type Percona Xtradb Cluster,
MariaDB Galera, and MySQL Cluster.
Topology "innodb-cluster" deploys Group Replication managed by MySQL Shell AdminAPI with optional
MySQL Router for connection routing (requires MySQL 8.0.11+ and mysqlsh).
For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	//Allowed topologies are "master-slave", "group" (requires 5.7.17+),
	//"fan-in" and "all-masters" (require 5.7.9+)
	// pxc (requires PXC tarball),
	// ndb (Requires NDB tarball)
	Run: replicationSandbox,
	Example: `
		$ dbdeployer deploy replication 5.7    # deploys highest revision for 5.7
		$ dbdeployer deploy replication 5.7.21 # deploys a specific revision
		$ dbdeployer deploy replication /path/to/5.7.21 # deploys a specific revision in a given path
		# (implies topology = master-slave)

		$ dbdeployer deploy --topology=master-slave replication 5.7
		# (explicitly setting topology)

		$ dbdeployer deploy --topology=group replication 5.7
		$ dbdeployer deploy --topology=group replication 8.0 --single-primary
		$ dbdeployer deploy --topology=all-masters replication 5.7
		$ dbdeployer deploy --topology=fan-in replication 5.7
		$ dbdeployer deploy --topology=pxc replication pxc5.7.25
		$ dbdeployer deploy --topology=galera replication 10.11.21
		$ dbdeployer deploy --topology=ndb replication ndb8.0.14
		$ dbdeployer deploy --topology=innodb-cluster replication 8.4.4
		$ dbdeployer deploy --topology=innodb-cluster replication 8.4.4 --skip-router
	`,
	Annotations: map[string]string{"export": ExportAnnotationToJson(ReplicationExport)},
}

func init() {
	deployCmd.AddCommand(replicationCmd)
	replicationCmd.PersistentFlags().StringP(globals.MasterListLabel, "", "", "Which nodes are masters in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP(globals.SlaveListLabel, "", "", "Which nodes are slaves in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP(globals.MasterIpLabel, "", globals.MasterIpValue, "Which IP the slaves will connect to")
	replicationCmd.PersistentFlags().StringP(globals.TopologyLabel, "t", globals.TopologyValue, "Which topology will be installed")
	replicationCmd.PersistentFlags().IntP(globals.NodesLabel, "n", globals.NodesValue, "How many nodes will be installed")
	replicationCmd.PersistentFlags().IntP(globals.NdbNodesLabel, "", globals.NdbNodesValue, "How many NDB nodes will be installed")
	replicationCmd.PersistentFlags().BoolP(globals.SinglePrimaryLabel, "", false, "Using single primary for group replication")
	replicationCmd.PersistentFlags().BoolP(globals.SemiSyncLabel, "", false, "Use semi-synchronous plugin")
	replicationCmd.PersistentFlags().BoolP(globals.ReadOnlyLabel, "", false, "Set read-only for slaves")
	replicationCmd.PersistentFlags().BoolP(globals.SuperReadOnlyLabel, "", false, "Set super-read-only for slaves")
	replicationCmd.PersistentFlags().Bool(globals.ReplHistoryDirLabel, false, "uses the replication directory to store mysql client history")
	setPflag(replicationCmd, globals.ChangeMasterOptions, "", "CHANGE_MASTER_OPTIONS", "", "options to add to CHANGE MASTER TO", true)
	replicationCmd.PersistentFlags().Bool(globals.SkipRouterLabel, false, "Skip MySQL Router deployment for InnoDB Cluster topology")
	replicationCmd.PersistentFlags().Bool("with-proxysql", false, "Deploy ProxySQL alongside the replication sandbox")
	replicationCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
}
