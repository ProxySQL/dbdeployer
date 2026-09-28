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
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fakeClientScript stands in for the mysql/mariadb client used by
// wait_until_wsrep_ready. It answers the three queries the function issues
// according to environment variables, so each test case can play a different
// kind of server:
//
//	FAKE_WSREP_ON          output row for SHOW VARIABLES LIKE 'wsrep_on'
//	                       ("" = no row, as on MySQL)
//	FAKE_WSREP_ON_FAIL     if "1", SHOW VARIABLES exits non-zero
//	FAKE_WSREP_READY       output row for SHOW STATUS LIKE 'wsrep_ready'
//	FAKE_READY_AFTER       if > 0, wsrep_ready turns ON from that call onward
//	FAKE_STATE_DIR         where the call counter is kept
const fakeClientScript = `#!/bin/bash
query=""
while [ $# -gt 0 ]; do
    if [ "$1" = "-e" ]; then query="$2"; shift; fi
    shift
done
case "$query" in
    "SELECT 1")
        echo 1 ;;
    *"SHOW VARIABLES LIKE 'wsrep_on'"*)
        [ "$FAKE_WSREP_ON_FAIL" = "1" ] && exit 1
        [ -n "$FAKE_WSREP_ON" ] && printf '%b\n' "$FAKE_WSREP_ON" ;;
    *"SHOW STATUS LIKE 'wsrep_ready'"*)
        n=$(cat "$FAKE_STATE_DIR/ready_calls" 2>/dev/null || echo 0)
        n=$((n + 1))
        echo $n > "$FAKE_STATE_DIR/ready_calls"
        if [ "${FAKE_READY_AFTER:-0}" -gt 0 ] && [ $n -ge $FAKE_READY_AFTER ]; then
            printf 'wsrep_ready\tON\n'
        elif [ -n "$FAKE_WSREP_READY" ]; then
            printf '%b\n' "$FAKE_WSREP_READY"
        fi ;;
esac
exit 0
`

// TestWaitUntilWsrepReady_Behavior executes the rendered wait_until_wsrep_ready
// function against a fake client. Unlike the substring checks in
// templates_flavor_test.go, it proves how long the function actually waits.
//
// Regression for #142: MariaDB without Galera reports "wsrep_ready OFF"
// (the variable exists), so a guard that only short-circuited on an *empty*
// wsrep_ready result spun for the full 60x2s on every MariaDB node.
func TestWaitUntilWsrepReady_Behavior(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell templates are not used on Windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}

	// Polling budget used by every case: 3 attempts x 1s. A case that
	// short-circuits returns well under 1s; one that polls to exhaustion
	// takes ~3s.
	const attempts, sleepSec = 3, 1
	const fastLimit = 900 * time.Millisecond

	tests := []struct {
		name       string
		flavor     string
		env        map[string]string
		wantRC     int
		wantFast   bool // returned without entering the polling loop
		wantPolled bool // polled wsrep_ready at least once
	}{
		{
			name:     "mysql has no wsrep variables",
			flavor:   "mysql",
			env:      map[string]string{},
			wantRC:   0,
			wantFast: true,
		},
		{
			// The #142 case: variables exist but are OFF.
			name:   "mariadb without galera reports wsrep_on OFF and wsrep_ready OFF",
			flavor: "mariadb",
			env: map[string]string{
				"FAKE_WSREP_ON":    `wsrep_on\tOFF`,
				"FAKE_WSREP_READY": `wsrep_ready\tOFF`,
			},
			wantRC:   0,
			wantFast: true,
		},
		{
			// A Galera node that is still joining: wsrep_ready is OFF for a
			// while. The function must keep waiting, not treat OFF as
			// "not a cluster".
			name:   "galera node joining becomes ready after polling",
			flavor: "mariadb",
			env: map[string]string{
				"FAKE_WSREP_ON":    `wsrep_on\tON`,
				"FAKE_WSREP_READY": `wsrep_ready\tOFF`,
				"FAKE_READY_AFTER": "2",
			},
			wantRC:     0,
			wantPolled: true,
		},
		{
			name:   "galera node that never becomes ready times out",
			flavor: "mariadb",
			env: map[string]string{
				"FAKE_WSREP_ON":    `wsrep_on\tON`,
				"FAKE_WSREP_READY": `wsrep_ready\tOFF`,
			},
			wantRC:     1,
			wantPolled: true,
		},
		{
			// A failing probe must not be misread as "non-Galera".
			name:   "failed wsrep_on probe falls through to polling",
			flavor: "mariadb",
			env: map[string]string{
				"FAKE_WSREP_ON_FAIL": "1",
				"FAKE_WSREP_READY":   `wsrep_ready\tOFF`,
			},
			wantRC:     1,
			wantPolled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			binDir := path.Join(dir, "bin")
			if err := os.Mkdir(binDir, 0755); err != nil {
				t.Fatal(err)
			}
			client := "mysql"
			if tc.flavor == "mariadb" {
				client = "mariadb"
			}
			// #nosec G306 -- the fake client must be executable
			if err := os.WriteFile(path.Join(binDir, client), []byte(fakeClientScript), 0755); err != nil {
				t.Fatal(err)
			}

			data := baseTemplateData(tc.flavor)
			data["SandboxDir"] = dir
			data["Basedir"] = dir
			data["ClientBasedir"] = dir
			data["SocketFile"] = path.Join(dir, "fake.sock")
			include := path.Join(dir, "sb_include")
			// #nosec G306 -- test fixture
			if err := os.WriteFile(include, []byte(renderTemplate(t, sbIncludeTemplate, data)), 0644); err != nil {
				t.Fatal(err)
			}

			script := "source " + include + "\nwait_until_wsrep_ready 3 1\n"
			// #nosec G204 -- fixed test script
			cmd := exec.Command(bash, "-c", script)
			cmd.Env = append(os.Environ(), "FAKE_STATE_DIR="+dir)
			for k, v := range tc.env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			start := time.Now()
			out, err := cmd.CombinedOutput()
			elapsed := time.Since(start)

			rc := 0
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatalf("running wait_until_wsrep_ready: %v\n%s", err, out)
				}
				rc = exitErr.ExitCode()
			}
			if rc != tc.wantRC {
				t.Errorf("exit code = %d, want %d\noutput: %s", rc, tc.wantRC, out)
			}

			readyCalls := 0
			if b, err := os.ReadFile(path.Join(dir, "ready_calls")); err == nil {
				readyCalls, _ = strconv.Atoi(strings.TrimSpace(string(b)))
			}
			if tc.wantFast {
				if elapsed > fastLimit {
					t.Errorf("took %v, want < %v (budget %dx%ds must not be spent on a non-Galera server)",
						elapsed, fastLimit, attempts, sleepSec)
				}
				if readyCalls != 0 {
					t.Errorf("polled wsrep_ready %d times, want 0 on a non-Galera server", readyCalls)
				}
			}
			if tc.wantPolled && readyCalls == 0 {
				t.Errorf("did not poll wsrep_ready; a possible Galera node must be waited on")
			}
		})
	}
}
