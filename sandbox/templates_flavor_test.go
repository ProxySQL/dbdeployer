// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"strings"
	"testing"

	"github.com/ProxySQL/dbdeployer/common"
)

// maxAllowedWaitSeconds is the maximum total wait time (in seconds) that any
// generated shell script should impose during normal operation on non-Galera
// MySQL. Setting this too high hides regressions (e.g. the 120s wsrep spin on
// vanilla MySQL from commit a2caf94); setting it too low causes flaky failures
// on slow CI runners. This value must be ≥ the per-node replica-ready poll.
const maxAllowedWaitSeconds = 30

func baseTemplateData(flavor string) common.StringMap {
	return common.StringMap{
		"ShellPath":       "/bin/bash",
		"SandboxDir":      "/tmp/sandbox/rsandbox_1234",
		"Basedir":         "/opt/mysql/11.4",
		"ClientBasedir":   "/opt/mysql/11.4",
		"Copyright":       "# test",
		"CustomMysqld":    "",
		"Port":            "1234",
		"MysqlXPort":      "0",
		"MysqlXSocket":    "",
		"AdminPort":       "0",
		"ServerId":        "100",
		"SocketFile":      "/tmp/sandbox/mysql_sandbox.sock",
		"Flavor":          flavor,
		"Version":         "11.4.10",
		"VersionMajor":    "11",
		"VersionMinor":    "4",
		"VersionRev":      "10",
		"SortableVersion": "011004010",
		"HistoryDir":      "/tmp/sandbox/rsandbox_1234",
		"SbHost":          "127.0.0.1",
		"SBType":          "replication-node",
		"SandboxType":     "single",
		"EngineClause":    "",
		"TemplateName":    "test",
	}
}

func renderTemplate(t *testing.T, tmpl string, data common.StringMap) string {
	t.Helper()
	result, err := common.SafeTemplateFill("test", tmpl, data)
	if err != nil {
		t.Fatalf("template rendering failed: %v", err)
	}
	return result
}

func TestStartTemplate_MySQL(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, startTemplate, data)
	if !strings.Contains(result, `MYSQLD_SAFE="bin/mysqld_safe"`) {
		t.Error("mysql start template should use mysqld_safe")
	}
	if strings.Contains(result, "mariadbd-safe") {
		t.Error("mysql start template should not reference mariadbd-safe")
	}
}

func TestStartTemplate_MariaDB(t *testing.T) {
	data := baseTemplateData("mariadb")
	result := renderTemplate(t, startTemplate, data)
	if !strings.Contains(result, `MYSQLD_SAFE="bin/mariadbd-safe"`) {
		t.Error("mariadb start template should use mariadbd-safe")
	}
}

func TestStopTemplate_MySQL(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, stopTemplate, data)
	if !strings.Contains(result, `$CLIENT_BASEDIR/bin/mysqladmin`) {
		t.Error("mysql stop template should use mysqladmin")
	}
	if strings.Contains(result, "mariadb-admin") {
		t.Error("mysql stop template should not reference mariadb-admin")
	}
}

func TestStopTemplate_MariaDB(t *testing.T) {
	data := baseTemplateData("mariadb")
	result := renderTemplate(t, stopTemplate, data)
	if !strings.Contains(result, `$CLIENT_BASEDIR/bin/mariadb-admin`) {
		t.Error("mariadb stop template should use mariadb-admin")
	}
}

func TestUseTemplate_MySQL(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, useTemplate, data)
	if !strings.Contains(result, `$CLIENT_BASEDIR/bin/mysql"`) {
		t.Error("mysql use template should use mysql client")
	}
	if strings.Contains(result, `bin/mariadb`) {
		t.Error("mysql use template should not reference mariadb client")
	}
}

func TestUseTemplate_MariaDB(t *testing.T) {
	data := baseTemplateData("mariadb")
	result := renderTemplate(t, useTemplate, data)
	if !strings.Contains(result, `$CLIENT_BASEDIR/bin/mariadb"`) {
		t.Error("mariadb use template should use mariadb client")
	}
}

func TestReplicationTopology_UsesSingleStartTemplate(t *testing.T) {
	// Replication creates per-node sandboxes via CreateChildSandbox,
	// which uses the same single templates. Verify the start template
	// (which replication/start_all.gotxt delegates to) handles both flavors.
	for _, flavor := range []string{"mysql", "mariadb"} {
		data := baseTemplateData(flavor)
		result := renderTemplate(t, startTemplate, data)

		if flavor == "mariadb" {
			if !strings.Contains(result, "mariadbd-safe") {
				t.Errorf("replication node start for mariadb should use mariadbd-safe")
			}
		} else {
			if !strings.Contains(result, "mysqld_safe") {
				t.Errorf("replication node start for mysql should use mysqld_safe")
			}
		}
	}
}

func TestGaleraTopology_UsesSingleStartTemplate(t *testing.T) {
	// Galera also creates per-node sandboxes that delegate to single templates.
	// The galera_start.gotxt calls $SBDIR/nodeN/start, which is the single start template.
	for _, flavor := range []string{"mysql", "mariadb"} {
		data := baseTemplateData(flavor)
		result := renderTemplate(t, startTemplate, data)

		if flavor == "mariadb" {
			if !strings.Contains(result, "mariadbd-safe") {
				t.Errorf("galera node start for mariadb should use mariadbd-safe")
			}
		} else {
			if !strings.Contains(result, "mysqld_safe") {
				t.Errorf("galera node start for mysql should use mysqld_safe")
			}
		}
	}
}

func TestReplicationStopAndUse_DelegateToSingleTemplates(t *testing.T) {
	// replication/stop_all.gotxt calls $SBDIR/nodeN/stop for each node
	// replication/use_all.gotxt calls $SBDIR/master/use etc.
	// Both delegate to the single stop.gotxt and use.gotxt.
	for _, flavor := range []string{"mysql", "mariadb"} {
		data := baseTemplateData(flavor)

		stopResult := renderTemplate(t, stopTemplate, data)
		useResult := renderTemplate(t, useTemplate, data)

		if flavor == "mariadb" {
			if !strings.Contains(stopResult, "mariadb-admin") {
				t.Errorf("replication node stop for mariadb should use mariadb-admin")
			}
			if !strings.Contains(useResult, "bin/mariadb") {
				t.Errorf("replication node use for mariadb should use mariadb client")
			}
		} else {
			if !strings.Contains(stopResult, "mysqladmin") {
				t.Errorf("replication node stop for mysql should use mysqladmin")
			}
			if !strings.Contains(useResult, "bin/mysql") {
				t.Errorf("replication node use for mysql should use mysql client")
			}
		}
	}
}

// TestInitSlavesTemplates_IncludeReplicaReadyWait ensures that the
// replication initialization templates (used by "deploy replication")
// contain the wait_until_replica_ready logic introduced to fix #131.
// This is a pure template test and does not require any MySQL binaries.
func TestInitSlavesTemplates_IncludeReplicaReadyWait(t *testing.T) {
	for name, tmplContent := range map[string]string{
		"init_slaves":    initSlavesTemplate,
		"init_slaves_84": initSlaves84Template,
	} {
		data := common.StringMap{
			"ShellPath":           "/bin/bash",
			"Copyright":           "# test",
			"AppVersion":          "test",
			"DateTime":            "now",
			"TemplateName":        name,
			"SandboxDir":          "/tmp/rsandbox_1234",
			"MasterLabel":         "master",
			"MasterIp":            "127.0.0.1",
			"RplUser":             "rsandbox",
			"RplPassword":         "rsandbox",
			"MasterAutoPosition":  "",
			"ChangeMasterExtra":   "",
			"StartReplica":        "START REPLICA",
			"NodeLabel":           "n",
			"SlaveLabel":          "slave",
			"ChangeMasterTo":      "CHANGE MASTER TO",
			"MasterHostParam":     "MASTER_HOST",
			"MasterPort":          "12345",
			"MasterPortParam":     "MASTER_PORT",
			"MasterUserParam":     "MASTER_USER",
			"MasterPasswordParam": "MASTER_PASSWORD",
			"Slaves": []common.StringMap{
				{"NodeLabel": "n", "Node": 1, "SlaveLabel": "slave"},
				{"NodeLabel": "n", "Node": 2, "SlaveLabel": "slave"},
			},
		}
		result := renderTemplate(t, tmplContent, data)

		if !strings.Contains(result, "wait_until_replica_ready") {
			t.Errorf("%s template must define the wait_until_replica_ready helper (regression for #131)", name)
		}
		// The call must appear for each slave after the START REPLICA line.
		if !strings.Contains(result, `wait_until_replica_ready "$SBDIR/n1/use" 20 1`) {
			t.Errorf("%s template must invoke wait for the first replica", name)
		}
		if !strings.Contains(result, `wait_until_replica_ready "$SBDIR/n2/use" 20 1`) {
			t.Errorf("%s template must invoke wait for the second replica", name)
		}
		// Verify the timeout warning message reflects the reduced max_attempts.
		// This ensures the generated script won't wait longer than expected.
		expectedWarning := "after $((max_attempts * sleep_sec))s"
		if !strings.Contains(result, expectedWarning) {
			t.Errorf("%s template must include a timeout warning message", name)
		}
	}
}

// TestWaitWsrepAfterStart_DoesNotBlockNonGalera verifies that the
// wait_wsrep_after_start template (which runs for every sandbox node
// after start, regardless of flavor) exits quickly on non-Galera MySQL.
// On vanilla MySQL the wsrep status variable does not exist; the function
// must detect this and return immediately. This is a pure template test.
func TestWaitWsrepAfterStart_DoesNotBlockNonGalera(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, waitWsrepAfterStartTemplate, data)

	if !strings.Contains(result, "wait_until_wsrep_ready") {
		t.Error("wait_wsrep_after_start template must call wait_until_wsrep_ready")
	}
	// The template appends '|| true' to make the wait best-effort.
	if !strings.Contains(result, "|| true") {
		t.Error("wait_wsrep_after_start must use '|| true' for best-effort semantics")
	}
}

// TestSbInclude_WaitUntilWsrepReady_DetectsNonGalera verifies that the
// wait_until_wsrep_ready function in sb_include.gotxt checks for the
// existence of the wsrep_ready status variable before polling. Without
// this guard, every non-Galera node would spin for the full timeout
// (up to 120s per node), which caused the 5-minute total delay reported
// on issue #131 after the v2.4.0 release.
func TestSbInclude_WaitUntilWsrepReady_DetectsNonGalera(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, sbIncludeTemplate, data)

	// The early-return guard: grep for 'wsrep_ready' string in the output
	// of SHOW STATUS. On non-Galera the output is empty and the function
	// returns 0 immediately.
	if !strings.Contains(result, `grep -q 'wsrep_ready'`) {
		t.Error("wait_until_wsrep_ready must check if wsrep_ready variable exists (regression: #131 2-min delay on non-Galera)")
	}
	// The default max_attempts for wsrep should not exceed the budget.
	if !strings.Contains(result, "local max_attempts=${2:-60}") {
		t.Error("wait_until_wsrep_ready default max_attempts must remain 60 for Galera nodes, but the guard above must short-circuit on non-Galera")
	}
}

// TestSbInclude_WaitUntilReplicaReady_BoundedWait verifies that the
// wait_until_replica_ready function in sb_include.gotxt has a default
// max_attempts that keeps the total wait within the allowed budget.
// This prevents the 60s-per-replica delay that was part of the #131
// v2.4.0 regression.
func TestSbInclude_WaitUntilReplicaReady_BoundedWait(t *testing.T) {
	data := baseTemplateData("mysql")
	result := renderTemplate(t, sbIncludeTemplate, data)

	// wait_until_replica_ready must have max_attempts=${2:-20}
	// (wait_until_wsrep_ready keeps ${2:-60} for Galera nodes).
	if !strings.Contains(result, `max_attempts=${2:-20}`) {
		t.Errorf("wait_until_replica_ready max_attempts should be 20 to keep total wait ≤ %ds",
			maxAllowedWaitSeconds)
	}
}
