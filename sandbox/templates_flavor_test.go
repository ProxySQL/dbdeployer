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

func baseTemplateData(flavor string) common.StringMap {
	return common.StringMap{
		"ShellPath":      "/bin/bash",
		"SandboxDir":     "/tmp/sandbox/rsandbox_1234",
		"Basedir":        "/opt/mysql/11.4",
		"ClientBasedir":  "/opt/mysql/11.4",
		"Copyright":      "# test",
		"CustomMysqld":  "",
		"Port":           "1234",
		"Flavor":         flavor,
		"Version":        "11.4.10",
		"VersionMajor":   "11",
		"VersionMinor":   "4",
		"VersionRev":     "10",
		"SortableVersion": "011004010",
		"HistoryDir":     "/tmp/sandbox/rsandbox_1234",
		"SbHost":         "127.0.0.1",
		"SBType":         "replication-node",
		"EngineClause":   "",
		"TemplateName":   "test",
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
