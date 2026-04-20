// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2026 Giuseppe Maxia
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

package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/concurrent"
	"github.com/ProxySQL/dbdeployer/defaults"
	"github.com/ProxySQL/dbdeployer/globals"
	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

// findMysqlShell locates the mysqlsh binary. It first checks the basedir/bin
// directory, then falls back to the system PATH.
func findMysqlShell(basedir string) (string, error) {
	mysqlshPath := path.Join(basedir, "bin", "mysqlsh")
	if common.ExecExists(mysqlshPath) {
		return mysqlshPath, nil
	}
	// Check if mysqlsh is available on PATH
	mysqlshPath = "mysqlsh"
	if common.ExecExists(mysqlshPath) {
		return mysqlshPath, nil
	}
	fullPath := common.FindInPath("mysqlsh")
	if fullPath != "" {
		return fullPath, nil
	}
	return "", fmt.Errorf("mysqlsh not found in %s/bin or in PATH. "+
		"MySQL Shell is required for InnoDB Cluster deployment. "+
		"Install it from https://dev.mysql.com/downloads/shell/", basedir)
}

// mysqlShellVersionRegexp matches the "Ver X.Y.Z" token emitted by
// `mysqlsh --version` (e.g. "mysqlsh   Ver 8.0.36 for Linux on x86_64 ...").
var mysqlShellVersionRegexp = regexp.MustCompile(`Ver\s+(\d+\.\d+\.\d+)`)

// getMysqlShellVersion runs `mysqlsh --version` and returns the parsed
// X.Y.Z version string.
func getMysqlShellVersion(mysqlshPath string) (string, error) {
	out, err := common.RunCmdCtrlWithArgs(mysqlshPath, []string{"--version"}, true)
	if err != nil {
		return "", fmt.Errorf("running '%s --version': %s", mysqlshPath, err)
	}
	m := mysqlShellVersionRegexp.FindStringSubmatch(out)
	if len(m) < 2 {
		return "", fmt.Errorf("could not parse mysqlsh version from: %s", out)
	}
	return m[1], nil
}

// checkMysqlShellCompatibility verifies that a MySQL Shell at shellVersion
// can drive a MySQL Server at serverVersion via AdminAPI. The rule is:
// mysqlsh's major.minor must be >= the server's major.minor. In particular,
// MySQL Shell 8.0.x refuses any server > 8.0 with
// "Unsupported server version: AdminAPI operations in this version of
// MySQL Shell support MySQL Server up to version 8.0".
func checkMysqlShellCompatibility(shellVersion, serverVersion string) error {
	shellList, err := common.VersionToList(shellVersion)
	if err != nil {
		return fmt.Errorf("invalid mysqlsh version '%s': %s", shellVersion, err)
	}
	serverList, err := common.VersionToList(serverVersion)
	if err != nil {
		return fmt.Errorf("invalid server version '%s': %s", serverVersion, err)
	}
	shellMajor, shellMinor := shellList[0], shellList[1]
	serverMajor, serverMinor := serverList[0], serverList[1]
	if shellMajor > serverMajor ||
		(shellMajor == serverMajor && shellMinor >= serverMinor) {
		return nil
	}
	return fmt.Errorf(
		"MySQL Shell %s is too old for MySQL Server %s: "+
			"the AdminAPI in mysqlsh %d.%d only supports MySQL Server up to %d.%d. "+
			"Install mysqlsh >= %d.%d (see https://dev.mysql.com/downloads/shell/) "+
			"and make sure it is the one found on $PATH, "+
			"or place its 'mysqlsh' binary under <basedir>/bin/",
		shellVersion, serverVersion,
		shellMajor, shellMinor, shellMajor, shellMinor,
		serverMajor, serverMinor)
}

// findMysqlRouter locates the mysqlrouter binary. It first checks the basedir/bin
// directory, then falls back to the system PATH.
func findMysqlRouter(basedir string) (string, error) {
	routerPath := path.Join(basedir, "bin", "mysqlrouter")
	if common.ExecExists(routerPath) {
		return routerPath, nil
	}
	routerPath = "mysqlrouter"
	if common.ExecExists(routerPath) {
		return routerPath, nil
	}
	fullPath := common.FindInPath("mysqlrouter")
	if fullPath != "" {
		return fullPath, nil
	}
	return "", fmt.Errorf("mysqlrouter not found in %s/bin or in PATH. "+
		"Use --skip-router to deploy without MySQL Router", basedir)
}

// CreateInnoDBCluster creates an InnoDB Cluster sandbox with the given number of nodes.
// It creates nodes using the same approach as Group Replication, then uses MySQL Shell
// to bootstrap the cluster via the AdminAPI (dba.createCluster / cluster.addInstance).
// Optionally, MySQL Router is bootstrapped for transparent connection routing.
func CreateInnoDBCluster(sandboxDef SandboxDef, origin string, nodes int, masterIp string, skipRouter bool) error {
	var execLists []concurrent.ExecutionList

	var logger *defaults.Logger
	if sandboxDef.Logger != nil {
		logger = sandboxDef.Logger
	} else {
		var fileName string
		var err error
		logger, fileName, err = defaults.NewLogger(common.LogDirName(), "innodb-cluster")
		if err != nil {
			return err
		}
		sandboxDef.LogFileName = common.ReplaceLiteralHome(fileName)
	}

	readOnlyOptions, err := checkReadOnlyFlags(sandboxDef)
	if err != nil {
		return err
	}
	if readOnlyOptions != "" {
		return fmt.Errorf("options --read-only and --super-read-only can't be used for InnoDB Cluster topology\n" +
			"as the cluster software sets it when needed")
	}

	// InnoDB Cluster requires MySQL 8.0+
	isMinimumVersion, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
	if err != nil {
		return err
	}
	if !isMinimumVersion {
		return fmt.Errorf("InnoDB Cluster requires MySQL 8.0.11 or later (current: %s)", sandboxDef.Version)
	}

	// Find mysqlsh - it is required
	mysqlshPath, err := findMysqlShell(sandboxDef.Basedir)
	if err != nil {
		return err
	}
	logger.Printf("Using MySQL Shell: %s\n", mysqlshPath)

	// Pre-flight: fail early if the mysqlsh AdminAPI can't manage this
	// server version (e.g. mysqlsh 8.0.x + server 8.4.x). If parsing the
	// shell's own --version fails for any reason, proceed — we don't want
	// unknown output to block a valid deployment.
	shellVersion, err := getMysqlShellVersion(mysqlshPath)
	if err != nil {
		logger.Printf("Warning: could not determine mysqlsh version: %s\n", err)
	} else {
		logger.Printf("Detected MySQL Shell version: %s\n", shellVersion)
		if err := checkMysqlShellCompatibility(shellVersion, sandboxDef.Version); err != nil {
			return err
		}
	}

	// Find mysqlrouter - optional if --skip-router is set
	var mysqlrouterPath string
	if !skipRouter {
		mysqlrouterPath, err = findMysqlRouter(sandboxDef.Basedir)
		if err != nil {
			return err
		}
		logger.Printf("Using MySQL Router: %s\n", mysqlrouterPath)
	}

	vList, err := common.VersionToList(sandboxDef.Version)
	if err != nil {
		return err
	}
	rev := vList[2]
	basePort := computeBaseport(sandboxDef.Port + defaults.Defaults().InnoDBClusterBasePort + (rev * 100))
	if sandboxDef.BasePort > 0 {
		basePort = sandboxDef.BasePort
	}

	if nodes < 3 {
		return fmt.Errorf("can't run InnoDB Cluster with less than 3 nodes")
	}
	if common.DirExists(sandboxDef.SandboxDir) {
		sandboxDef, err = checkDirectory(sandboxDef)
		if err != nil {
			return err
		}
	}

	// Allocate MySQL ports
	firstPort, err := common.FindFreePort(basePort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error retrieving free port for InnoDB Cluster")
	}
	basePort = firstPort - 1

	// Allocate GR communication ports
	baseGroupPort := basePort + defaults.Defaults().GroupPortDelta
	firstGroupPort, err := common.FindFreePort(baseGroupPort+1, sandboxDef.InstalledPorts, nodes)
	if err != nil {
		return errors.Wrapf(err, "error retrieving group replication free port")
	}
	baseGroupPort = firstGroupPort - 1

	for checkPort := basePort + 1; checkPort < basePort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreateInnoDBCluster", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}
	for checkPort := baseGroupPort + 1; checkPort < baseGroupPort+nodes+1; checkPort++ {
		err = checkPortAvailability("CreateInnoDBCluster-group", sandboxDef.SandboxDir, sandboxDef.InstalledPorts, checkPort)
		if err != nil {
			return err
		}
	}

	baseMysqlxPort, err := getBaseMysqlxPort(basePort, sandboxDef, nodes)
	if err != nil {
		return err
	}
	baseAdminPort, err := getBaseAdminPort(basePort, sandboxDef, nodes)
	if err != nil {
		return err
	}

	err = os.Mkdir(sandboxDef.SandboxDir, globals.PublicDirectoryAttr)
	if err != nil {
		return err
	}
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", sandboxDef.SandboxDir)
	logger.Printf("Creating directory %s\n", sandboxDef.SandboxDir)

	timestamp := time.Now()
	slaveLabel := defaults.Defaults().SlavePrefix
	slaveAbbr := defaults.Defaults().SlaveAbbr
	masterAbbr := defaults.Defaults().MasterAbbr
	masterLabel := defaults.Defaults().MasterName
	// InnoDB Cluster always uses single-primary mode by default
	// The primary is node 1, the rest are secondaries
	masterList := "1"
	slaveList := ""
	for N := 2; N <= nodes; N++ {
		if slaveList != "" {
			slaveList += " "
		}
		slaveList += fmt.Sprintf("%d", N)
	}

	changeMasterExtra := setChangeMasterProperties("", sandboxDef.ChangeMasterOptions, logger)
	nodeLabel := defaults.Defaults().NodePrefix
	stopNodeList := ""
	for i := nodes; i > 0; i-- {
		stopNodeList += fmt.Sprintf(" %d", i)
	}

	replCmds := replicationCommands(sandboxDef.Version)

	// Build the connection string for GR seeds
	connectionString := ""
	for i := 0; i < nodes; i++ {
		groupPort := baseGroupPort + i + 1
		if connectionString != "" {
			connectionString += ","
		}
		connectionString += fmt.Sprintf("127.0.0.1:%d", groupPort)
	}
	logger.Printf("Creating connection string %s\n", connectionString)

	routerDir := path.Join(sandboxDef.SandboxDir, "router")

	var data = common.StringMap{
		"ShellPath":         sandboxDef.ShellPath,
		"Copyright":         globals.ShellScriptCopyright,
		"AppVersion":        common.VersionDef,
		"DateTime":          timestamp.Format(time.UnixDate),
		"SandboxDir":        sandboxDef.SandboxDir,
		"MasterIp":          masterIp,
		"MasterList":        masterList,
		"NodeLabel":         nodeLabel,
		"SlaveList":         slaveList,
		"RplUser":           sandboxDef.RplUser,
		"RplPassword":       sandboxDef.RplPassword,
		"SlaveLabel":        slaveLabel,
		"SlaveAbbr":         slaveAbbr,
		"ChangeMasterExtra": changeMasterExtra,
		"MasterLabel":       masterLabel,
		"MasterAbbr":        masterAbbr,
		"StopNodeList":      stopNodeList,
		"Nodes":             []common.StringMap{},
		// InnoDB Cluster specific
		"Basedir":     sandboxDef.Basedir,
		"MysqlShell":  mysqlshPath,
		"PrimaryPort": basePort + 1,
		"ClusterName": "mySandboxCluster",
		"DbPassword":  sandboxDef.DbPassword,
		"RouterDir":   routerDir,
		"Replicas":    []common.StringMap{},
	}
	data["ChangeMasterTo"] = replCmds["ChangeMasterTo"]
	data["MasterUserParam"] = replCmds["MasterUserParam"]
	data["MasterPasswordParam"] = replCmds["MasterPasswordParam"]
	data["StartReplica"] = replCmds["StartReplica"]
	data["StopReplica"] = replCmds["StopReplica"]
	data["ResetMasterCmd"] = replCmds["ResetMasterCmd"]

	sbType := "innodb-cluster"
	logger.Printf("Defining cluster type %s\n", sbType)

	sbDesc := common.SandboxDescription{
		Basedir: sandboxDef.Basedir,
		SBType:  sbType,
		Version: sandboxDef.Version,
		Flavor:  sandboxDef.Flavor,
		Port:    []int{},
		Nodes:   nodes,
		NodeNum: 0,
		LogFile: sandboxDef.LogFileName,
	}

	sbItem := defaults.SandboxItem{
		Origin:      sbDesc.Basedir,
		SBType:      sbDesc.SBType,
		Version:     sandboxDef.Version,
		Flavor:      sandboxDef.Flavor,
		Port:        []int{},
		Nodes:       []string{},
		Destination: sandboxDef.SandboxDir,
	}

	if sandboxDef.LogFileName != "" {
		sbItem.LogDirectory = common.DirName(sandboxDef.LogFileName)
	}

	// Version-aware group replication init template
	initNodesTmpl := globals.TmplInitNodes
	isMySQL84, _ := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumResetBinaryLogsVersion)
	if isMySQL84 {
		initNodesTmpl = globals.TmplInitNodes84
	}

	for i := 1; i <= nodes; i++ {
		groupPort := baseGroupPort + i
		sandboxDef.Port = basePort + i
		data["Nodes"] = append(data["Nodes"].([]common.StringMap), common.StringMap{
			"ShellPath":           sandboxDef.ShellPath,
			"Copyright":           globals.ShellScriptCopyright,
			"AppVersion":          common.VersionDef,
			"DateTime":            timestamp.Format(time.UnixDate),
			"Node":                i,
			"NodePort":            sandboxDef.Port,
			"MasterIp":            masterIp,
			"NodeLabel":           nodeLabel,
			"SlaveLabel":          slaveLabel,
			"SlaveAbbr":           slaveAbbr,
			"ChangeMasterExtra":   changeMasterExtra,
			"ChangeMasterTo":      replCmds["ChangeMasterTo"],
			"MasterUserParam":     replCmds["MasterUserParam"],
			"MasterPasswordParam": replCmds["MasterPasswordParam"],
			"ResetMasterCmd":      replCmds["ResetMasterCmd"],
			"MasterLabel":         masterLabel,
			"MasterAbbr":          masterAbbr,
			"SandboxDir":          sandboxDef.SandboxDir,
			"StopNodeList":        stopNodeList,
			"RplUser":             sandboxDef.RplUser,
			"RplPassword":         sandboxDef.RplPassword,
		})

		// Build replica list for init_cluster template (nodes 2..N)
		if i > 1 {
			data["Replicas"] = append(data["Replicas"].([]common.StringMap), common.StringMap{
				"Port": sandboxDef.Port,
			})
		}

		sandboxDef.DirName = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.MorePorts = []int{groupPort}
		sandboxDef.ServerId = setServerId(sandboxDef, i)
		sbItem.Nodes = append(sbItem.Nodes, sandboxDef.DirName)
		sbItem.Port = append(sbItem.Port, sandboxDef.Port)
		sbDesc.Port = append(sbDesc.Port, sandboxDef.Port)
		sbItem.Port = append(sbItem.Port, groupPort)
		sbDesc.Port = append(sbDesc.Port, groupPort)

		if !sandboxDef.RunConcurrently {
			installationMessage := "Installing and starting %s %d\n"
			if sandboxDef.SkipStart {
				installationMessage = "Installing %s %d\n"
			}
			common.CondPrintf(installationMessage, nodeLabel, i)
			logger.Printf(installationMessage, nodeLabel, i)
		}

		basePortText := fmt.Sprintf("%08d", basePort)

		// Version-aware options for group replication
		useReplicaUpdates, _ := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumShowReplicaStatusVersion)
		useNoWriteSetExtraction, _ := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumNoWriteSetExtractionVersion)
		useMySQL84GroupOptions, _ := common.GreaterOrEqualVersion(sandboxDef.Version, globals.MinimumResetBinaryLogsVersion)

		replicationData := common.StringMap{
			"BasePort":               basePortText,
			"GroupSeeds":             connectionString,
			"LocalAddresses":         fmt.Sprintf("%s:%d", masterIp, groupPort),
			"PrimaryMode":            "on", // InnoDB Cluster defaults to single-primary
			"UseReplicaUpdates":      useReplicaUpdates,
			"SkipWriteSetExtraction": useNoWriteSetExtraction,
		}

		// Use the same GR options templates as group replication
		groupReplOptionsTmpl := globals.TmplGroupReplOptions
		if useMySQL84GroupOptions {
			groupReplOptionsTmpl = globals.TmplGroupReplOptions84
		}
		replOptionsText, err := common.SafeTemplateFill("innodb_cluster_gr",
			GroupTemplates[groupReplOptionsTmpl].Contents, replicationData)
		if err != nil {
			return err
		}
		sandboxDef.ReplOptions = SingleTemplates[globals.TmplReplicationOptions].Contents + "\n" + replOptionsText

		reMasterIp := regexp.MustCompile(`127\.0\.0\.1`)
		sandboxDef.ReplOptions = reMasterIp.ReplaceAllString(sandboxDef.ReplOptions, masterIp)

		sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates[globals.TmplGtidOptions57].Contents)
		if useMySQL84GroupOptions {
			sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates[globals.TmplReplCrashSafeOptions84].Contents)
		} else {
			sandboxDef.ReplOptions += fmt.Sprintf("\n%s\n", SingleTemplates[globals.TmplReplCrashSafeOptions].Contents)
		}

		// MySQLX port (required for InnoDB Cluster / MySQL Shell)
		isMinimumMySQLXDefault, err := common.HasCapability(sandboxDef.Flavor, common.MySQLXDefault, sandboxDef.Version)
		if err != nil {
			return err
		}
		if isMinimumMySQLXDefault || sandboxDef.EnableMysqlX {
			sandboxDef.MysqlXPort = baseMysqlxPort + i
			if !sandboxDef.DisableMysqlX {
				sbDesc.Port = append(sbDesc.Port, baseMysqlxPort+i)
				sbItem.Port = append(sbItem.Port, baseMysqlxPort+i)
				logger.Printf("adding mysqlx port %d to node %d\n", baseMysqlxPort+i, i)
			}
		}
		if sandboxDef.EnableAdminAddress {
			sandboxDef.AdminPort = baseAdminPort + i
			sbDesc.Port = append(sbDesc.Port, baseAdminPort+i)
			sbItem.Port = append(sbItem.Port, baseAdminPort+i)
			logger.Printf("adding admin port %d to node %d\n", baseAdminPort+i, i)
		}

		sandboxDef.Multi = true
		sandboxDef.LoadGrants = true
		sandboxDef.Prompt = fmt.Sprintf("%s%d", nodeLabel, i)
		sandboxDef.SBType = "innodb-cluster-node"
		sandboxDef.NodeNum = i
		logger.Printf("Create single sandbox for node %d\n", i)
		execList, err := CreateChildSandbox(sandboxDef)
		if err != nil {
			return fmt.Errorf(globals.ErrCreatingSandbox, err)
		}
		execLists = append(execLists, execList...)

		var dataNode = common.StringMap{
			"ShellPath":           sandboxDef.ShellPath,
			"Copyright":           globals.ShellScriptCopyright,
			"AppVersion":          common.VersionDef,
			"DateTime":            timestamp.Format(time.UnixDate),
			"Node":                i,
			"NodePort":            sandboxDef.Port,
			"NodeLabel":           nodeLabel,
			"MasterLabel":         masterLabel,
			"MasterAbbr":          masterAbbr,
			"ChangeMasterExtra":   changeMasterExtra,
			"ChangeMasterTo":      replCmds["ChangeMasterTo"],
			"MasterUserParam":     replCmds["MasterUserParam"],
			"MasterPasswordParam": replCmds["MasterPasswordParam"],
			"ResetMasterCmd":      replCmds["ResetMasterCmd"],
			"SlaveLabel":          slaveLabel,
			"SlaveAbbr":           slaveAbbr,
			"SandboxDir":          sandboxDef.SandboxDir,
		}
		logger.Printf("Create node script for node %d\n", i)
		err = writeScript(logger, MultipleTemplates, fmt.Sprintf("n%d", i), globals.TmplNode, sandboxDef.SandboxDir, dataNode, true)
		if err != nil {
			return err
		}
		if sandboxDef.EnableAdminAddress {
			err = writeScript(logger, MultipleTemplates, fmt.Sprintf("na%d", i), globals.TmplNodeAdmin, sandboxDef.SandboxDir, dataNode, true)
			if err != nil {
				return err
			}
		}
	}

	logger.Printf("Writing sandbox description in %s\n", sandboxDef.SandboxDir)
	err = common.WriteSandboxDescription(sandboxDef.SandboxDir, sbDesc)
	if err != nil {
		return errors.Wrapf(err, "unable to write sandbox description")
	}
	err = defaults.UpdateCatalog(sandboxDef.SandboxDir, sbItem)
	if err != nil {
		return errors.Wrapf(err, "unable to update catalog")
	}

	slavePlural := english.PluralWord(2, slaveLabel, "")
	masterPlural := english.PluralWord(2, masterLabel, "")
	useAllMasters := "use_all_" + masterPlural
	useAllSlaves := "use_all_" + slavePlural
	execAllSlaves := "exec_all_" + slavePlural
	execAllMasters := "exec_all_" + masterPlural

	logger.Printf("Writing InnoDB Cluster scripts\n")
	sbMultiple := ScriptBatch{
		tc:         MultipleTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptStartAll, globals.TmplStartMulti, true},
			{globals.ScriptRestartAll, globals.TmplRestartMulti, true},
			{globals.ScriptStatusAll, globals.TmplStatusMulti, true},
			{globals.ScriptTestSbAll, globals.TmplTestSbMulti, true},
			{globals.ScriptStopAll, globals.TmplStopMulti, true},
			{globals.ScriptClearAll, globals.TmplClearMulti, true},
			{globals.ScriptSendKillAll, globals.TmplSendKillMulti, true},
			{globals.ScriptUseAll, globals.TmplUseMulti, true},
			{globals.ScriptMetadataAll, globals.TmplMetadataMulti, true},
			{globals.ScriptReplicateFrom, globals.TmplReplicateFromMulti, true},
			{globals.ScriptSysbench, globals.TmplSysbenchMulti, true},
			{globals.ScriptSysbenchReady, globals.TmplSysbenchReadyMulti, true},
			{globals.ScriptExecAll, globals.TmplExecMulti, true},
		},
	}
	sbRepl := ScriptBatch{
		tc:         ReplicationTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{useAllSlaves, globals.TmplMultiSourceUseSlaves, true},
			{useAllMasters, globals.TmplMultiSourceUseMasters, true},
			{execAllMasters, globals.TmplMultiSourceExecMasters, true},
			{execAllSlaves, globals.TmplMultiSourceExecSlaves, true},
			{globals.ScriptTestReplication, globals.TmplMultiSourceTest, true},
			{globals.ScriptWipeRestartAll, globals.TmplWipeAndRestartAll, true},
		},
	}
	sbGroup := ScriptBatch{
		tc:         GroupTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptInitializeNodes, initNodesTmpl, true},
			{globals.ScriptCheckNodes, globals.TmplCheckNodes, true},
		},
	}
	// InnoDB Cluster specific scripts
	sbCluster := ScriptBatch{
		tc:         InnoDBClusterTemplates,
		logger:     logger,
		data:       data,
		sandboxDir: sandboxDef.SandboxDir,
		scripts: []ScriptDef{
			{globals.ScriptInitCluster, globals.TmplInitCluster, true},
			{globals.ScriptCheckCluster, globals.TmplCheckCluster, true},
		},
	}

	if !skipRouter {
		sbCluster.scripts = append(sbCluster.scripts,
			ScriptDef{globals.ScriptRouterStart, globals.TmplRouterStart, true},
			ScriptDef{globals.ScriptRouterStop, globals.TmplRouterStop, true},
		)
	}

	for _, sb := range []ScriptBatch{sbMultiple, sbRepl, sbGroup, sbCluster} {
		err := writeScripts(sb)
		if err != nil {
			return err
		}
	}

	if sandboxDef.EnableAdminAddress {
		logger.Printf("Creating admin script for all nodes\n")
		err = writeScript(logger, MultipleTemplates, globals.ScriptUseAllAdmin,
			globals.TmplUseMultiAdmin, sandboxDef.SandboxDir, data, true)
		if err != nil {
			return err
		}
	}

	logger.Printf("Running parallel tasks\n")
	concurrent.RunParallelTasksByPriority(execLists)

	if !sandboxDef.SkipStart {
		// For InnoDB Cluster, skip the standard GR initialization.
		// MySQL Shell's dba.createCluster() manages group replication itself.
		// Running initialize_nodes would start GR before mysqlsh, causing conflicts.

		// Bootstrap the cluster via MySQL Shell
		common.CondPrintln(path.Join(common.ReplaceLiteralHome(sandboxDef.SandboxDir), globals.ScriptInitCluster))
		logger.Printf("Running InnoDB Cluster initialization script\n")
		_, err = common.RunCmd(path.Join(sandboxDef.SandboxDir, globals.ScriptInitCluster))
		if err != nil {
			return fmt.Errorf("error initializing InnoDB Cluster: %s", err)
		}

		// Bootstrap MySQL Router if requested
		if !skipRouter && mysqlrouterPath != "" {
			logger.Printf("Bootstrapping MySQL Router\n")
			err = bootstrapRouter(mysqlrouterPath, routerDir, basePort+1, sandboxDef.DbPassword, logger)
			if err != nil {
				common.CondPrintf("WARNING: MySQL Router bootstrap failed: %s\n", err)
				common.CondPrintln("The cluster is functional without Router. Use mysqlsh to connect directly.")
			}
		}
	}

	common.CondPrintf("InnoDB Cluster directory installed in %s\n", common.ReplaceLiteralHome(sandboxDef.SandboxDir))
	common.CondPrintf("run 'dbdeployer usage multiple' for basic instructions'\n")
	return nil
}

// bootstrapRouter bootstraps MySQL Router against the InnoDB Cluster.
func bootstrapRouter(mysqlrouterPath, routerDir string, primaryPort int, dbPassword string, logger *defaults.Logger) error {
	err := os.MkdirAll(routerDir, globals.PublicDirectoryAttr)
	if err != nil {
		return fmt.Errorf("error creating router directory: %s", err)
	}

	bootstrapURI := fmt.Sprintf("icadmin:icadmin@127.0.0.1:%d", primaryPort)
	args := []string{
		"--bootstrap", bootstrapURI,
		"--directory", routerDir,
		"--force",
		"--conf-use-sockets",
	}

	logger.Printf("Running: %s %v\n", mysqlrouterPath, args)
	_, err = common.RunCmdWithArgs(mysqlrouterPath, args)
	if err != nil {
		return fmt.Errorf("mysqlrouter bootstrap failed: %s", err)
	}

	// Start the router directly (not via start.sh, which backgrounds the
	// process but inherits pipes — causing RunCmd to block forever).
	confFile := path.Join(routerDir, "mysqlrouter.conf")
	if common.FileExists(confFile) {
		cmd := exec.Command(mysqlrouterPath, "-c", confFile)
		cmd.Env = append(os.Environ(), fmt.Sprintf("ROUTER_PID=%s/mysqlrouter.pid", routerDir))
		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("error starting MySQL Router: %s", err)
		}
		common.CondPrintln("MySQL Router started")
	}

	return nil
}
