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
	"github.com/ProxySQL/dbdeployer/globals"
	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/ProxySQL/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func replicationSandbox(cmd *cobra.Command, args []string) {
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
	flags := cmd.Flags()
	semisync, _ = flags.GetBool(globals.SemiSyncLabel)
	ndbNodes, _ := flags.GetInt(globals.NdbNodesLabel)
	nodes, _ := flags.GetInt(globals.NodesLabel)
	topology, _ := flags.GetString(globals.TopologyLabel)
	masterIp, _ := flags.GetString(globals.MasterIpLabel)
	masterList, _ := flags.GetString(globals.MasterListLabel)
	slaveList, _ := flags.GetString(globals.SlaveListLabel)
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
			sd.SemiSyncOptions = sandbox.SingleTemplates[globals.TmplSemisyncMasterOptions].Contents
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
			SlaveList:  slaveList})
	if err != nil {
		common.Exitf(1, globals.ErrCreatingSandbox, err)
	}

	withProxySQL, _ := flags.GetBool("with-proxysql")
	if withProxySQL {
		// Determine the sandbox directory that was created
		sandboxDir := path.Join(sd.SandboxDir, defaults.Defaults().MasterSlavePrefix+common.VersionToName(origin))
		if sd.DirName != "" {
			sandboxDir = path.Join(sd.SandboxDir, sd.DirName)
		}

		// Read port info from child sandbox descriptions
		masterDesc, err := common.ReadSandboxDescription(path.Join(sandboxDir, defaults.Defaults().MasterName))
		if err != nil {
			common.Exitf(1, "could not read master sandbox description: %s", err)
		}
		masterPort := masterDesc.Port[0]

		var slavePorts []int
		for i := 1; i < nodes; i++ {
			nodeDir := path.Join(sandboxDir, fmt.Sprintf("%s%d", defaults.Defaults().NodePrefix, i))
			nodeDesc, err := common.ReadSandboxDescription(nodeDir)
			if err != nil {
				common.Exitf(1, "could not read node%d sandbox description: %s", i, err)
			}
			slavePorts = append(slavePorts, nodeDesc.Port[0])
		}

		err = sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, slavePorts, 0, "127.0.0.1")
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
Topologies "pcx" and "ndb" are available for binaries of type Percona Xtradb Cluster and MySQL Cluster.
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
		$ dbdeployer deploy --topology=ndb replication ndb8.0.14
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
	replicationCmd.PersistentFlags().Bool("with-proxysql", false, "Deploy ProxySQL alongside the replication sandbox")
}
