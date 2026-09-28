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

import "testing"

// legacyReplicationCommands is the syntax every MariaDB release line uses.
// MariaDB accepts CHANGE MASTER TO and rejects CHANGE REPLICATION SOURCE TO,
// so emitting the MySQL 8.0.23+ forms breaks replication (issue #82).
var legacyReplicationCommands = map[string]string{
	"ShowMasterStatus":    "show master status",
	"ShowSlaveStatus":     "show slave status",
	"ChangeMasterTo":      "CHANGE MASTER TO",
	"StartReplica":        "START SLAVE",
	"StopReplica":         "STOP SLAVE",
	"ResetReplica":        "RESET SLAVE",
	"ResetMasterCmd":      "reset master",
	"MasterPosWaitFunc":   "master_pos_wait",
	"MasterHostParam":     "master_host",
	"MasterPortParam":     "master_port",
	"MasterUserParam":     "master_user",
	"MasterPasswordParam": "master_password",
}

// TestReplicationCommands_MariaDB pins the replication syntax generated for
// every MariaDB release line dbdeployer is verified against. Each of these
// versions is also deployed for real by the mariadb-test CI job.
func TestReplicationCommands_MariaDB(t *testing.T) {
	versions := []string{
		"10.11.9",
		"11.4.9",
		"11.4.13", // issue #141
		"12.3.3",  // issue #139
		"13.0.2",  // issue #140
	}
	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			cmds := replicationCommands(version)
			for key, want := range legacyReplicationCommands {
				if got := cmds[key]; got != want {
					t.Errorf("%s: got %q, want %q", key, got, want)
				}
			}
		})
	}
}

// TestReplicationCommands_MySQL checks the MySQL version thresholds, so the
// MariaDB guard above cannot pass by breaking MySQL detection altogether.
func TestReplicationCommands_MySQL(t *testing.T) {
	tests := []struct {
		version string
		key     string
		want    string
	}{
		{"5.7.44", "ChangeMasterTo", "CHANGE MASTER TO"},
		{"8.0.22", "ShowSlaveStatus", "show replica status"},
		{"8.0.22", "ChangeMasterTo", "CHANGE MASTER TO"},
		{"8.0.23", "ChangeMasterTo", "CHANGE REPLICATION SOURCE TO"},
		{"8.0.23", "MasterHostParam", "source_host"},
		{"8.2.0", "ShowMasterStatus", "show binary log status"},
		{"8.4.4", "ResetMasterCmd", "RESET BINARY LOGS AND GTIDS"},
		{"9.5.0", "ChangeMasterTo", "CHANGE REPLICATION SOURCE TO"},
	}
	for _, tc := range tests {
		if got := replicationCommands(tc.version)[tc.key]; got != tc.want {
			t.Errorf("%s %s: got %q, want %q", tc.version, tc.key, got, tc.want)
		}
	}
}
