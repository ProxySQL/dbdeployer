// DBDeployer - The MySQL Sandbox
// Copyright © 2026 Øystein Grøvlen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package downloads

import "testing"

func TestValidateMysqlReleaseSeries(t *testing.T) {
	for _, version := range []string{"9.3", "9.4", "9.5", "9.6", "9.7", "26.7"} {
		for _, os := range []string{"linux", "darwin"} {
			t.Run(version+"/"+os, func(t *testing.T) {
				if _, err := validateTarballRequest(TtMysql, version, os); err != nil {
					t.Fatalf("MySQL %s on %s should be accepted: %v", version, os, err)
				}
			})
		}
	}

	for _, version := range []string{"invalid", "9.7.2", "26.7.0", "99.99"} {
		t.Run("reject/"+version, func(t *testing.T) {
			if _, err := validateTarballRequest(TtMysql, version, "linux"); err == nil {
				t.Fatalf("invalid release series %q should be rejected", version)
			}
		})
	}
}
