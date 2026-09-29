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

	"github.com/ProxySQL/dbdeployer/globals"
)

// fakeClientScript stands in for the mysql/mariadb client used by
// wait_until_wsrep_ready. Every invocation is counted in $FAKE_STATE_DIR/calls.
// SHOW STATUS LIKE 'wsrep_ready' is answered according to:
//
//	FAKE_WSREP_READY   output row ("" = no row, as on a server without wsrep)
//	FAKE_READY_AFTER   if > 0, wsrep_ready turns ON from that call onward
const fakeClientScript = `#!/bin/bash
calls=$(cat "$FAKE_STATE_DIR/calls" 2>/dev/null || echo 0)
echo $((calls + 1)) > "$FAKE_STATE_DIR/calls"
query=""
while [ $# -gt 0 ]; do
    if [ "$1" = "-e" ]; then query="$2"; shift; fi
    shift
done
case "$query" in
    "SELECT 1")
        echo 1 ;;
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

type bashResult struct {
	rc      int
	out     string
	elapsed time.Duration
	calls   int // client invocations of any kind
}

func requireBash(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell templates are not used on Windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available")
	}
	return bash
}

// setupFakeSandbox writes a fake client and the rendered sb_include into a
// temporary sandbox directory, and returns the directory.
func setupFakeSandbox(t *testing.T, flavor string) string {
	t.Helper()
	dir := t.TempDir()
	binDir := path.Join(dir, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, client := range []string{"mysql", "mariadb"} {
		// #nosec G306 -- the fake client must be executable
		if err := os.WriteFile(path.Join(binDir, client), []byte(fakeClientScript), 0755); err != nil {
			t.Fatal(err)
		}
	}
	data := baseTemplateData(flavor)
	data["SandboxDir"] = dir
	data["Basedir"] = dir
	data["ClientBasedir"] = dir
	data["SocketFile"] = path.Join(dir, "fake.sock")
	// #nosec G306 -- test fixture
	if err := os.WriteFile(path.Join(dir, "sb_include"), []byte(renderTemplate(t, sbIncludeTemplate, data)), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runBash(t *testing.T, bash, dir, script string, env map[string]string) bashResult {
	t.Helper()
	// #nosec G204 -- fixed test script
	cmd := exec.Command(bash, "-c", script)
	cmd.Env = append(os.Environ(), "FAKE_STATE_DIR="+dir)
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	start := time.Now()
	out, err := cmd.CombinedOutput()
	res := bashResult{out: string(out), elapsed: time.Since(start)}
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running bash: %v\n%s", err, out)
		}
		res.rc = exitErr.ExitCode()
	}
	if b, err := os.ReadFile(path.Join(dir, "calls")); err == nil {
		res.calls, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	return res
}

// TestWaitUntilWsrepReady_Behavior executes the rendered wait_until_wsrep_ready
// function against a fake client with a 3x1s budget. The function only runs on
// Galera/PXC nodes, so it must wait for wsrep_ready ON and never decide on its
// own that the server is "not a cluster".
func TestWaitUntilWsrepReady_Behavior(t *testing.T) {
	bash := requireBash(t)
	const fastLimit = 900 * time.Millisecond

	tests := []struct {
		name     string
		env      map[string]string
		wantRC   int
		wantFast bool
	}{
		{
			name:     "node already ready",
			env:      map[string]string{"FAKE_READY_AFTER": "1"},
			wantRC:   0,
			wantFast: true,
		},
		{
			name: "node joining becomes ready after polling",
			env: map[string]string{
				"FAKE_WSREP_READY": `wsrep_ready\tOFF`,
				"FAKE_READY_AFTER": "2",
			},
			wantRC: 0,
		},
		{
			name:   "node that never becomes ready times out",
			env:    map[string]string{"FAKE_WSREP_READY": `wsrep_ready\tOFF`},
			wantRC: 1,
		},
		{
			// A cluster node reporting no wsrep variables is broken (e.g. the
			// provider failed to load). Treating that as "not Galera" would
			// hide the failure.
			name:   "cluster node without wsrep variables fails",
			env:    map[string]string{},
			wantRC: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := setupFakeSandbox(t, "mariadb")
			res := runBash(t, bash, dir, "source "+path.Join(dir, "sb_include")+"\nwait_until_wsrep_ready 3 1\n", tc.env)
			if res.rc != tc.wantRC {
				t.Errorf("exit code = %d, want %d\noutput: %s", res.rc, tc.wantRC, res.out)
			}
			if tc.wantFast && res.elapsed > fastLimit {
				t.Errorf("took %v, want < %v", res.elapsed, fastLimit)
			}
			if tc.wantRC != 0 && !strings.Contains(res.out, "wsrep_ready not ON after 3s") {
				t.Errorf("timeout must be reported, got: %s", res.out)
			}
		})
	}
}

// TestWaitWsrepAfterStart_Script executes the generated wait_wsrep_after_start
// script. On non-cluster sandboxes it must exit 0 without contacting the
// server at all (#131, #142). On Galera/PXC nodes a failed wait must fail the
// script, so the deploy stops with a clear message instead of failing later.
func TestWaitWsrepAfterStart_Script(t *testing.T) {
	bash := requireBash(t)

	writeScript := func(t *testing.T, dir, sbType string) string {
		t.Helper()
		data := baseTemplateData("mariadb")
		data["SandboxDir"] = dir
		data["SandboxType"] = sbType
		script := path.Join(dir, "wait_wsrep_after_start")
		// #nosec G306 -- test fixture
		if err := os.WriteFile(script, []byte(renderTemplate(t, waitWsrepAfterStartTemplate, data)), 0755); err != nil {
			t.Fatal(err)
		}
		return script
	}

	for _, sbType := range []string{globals.SbTypeSingle, "replication-node"} {
		t.Run(sbType+" never contacts the server", func(t *testing.T) {
			dir := setupFakeSandbox(t, "mariadb")
			script := writeScript(t, dir, sbType)
			// MariaDB without Galera: the variable exists and is OFF.
			res := runBash(t, bash, dir, script, map[string]string{"FAKE_WSREP_READY": `wsrep_ready\tOFF`})
			if res.rc != 0 {
				t.Errorf("exit code = %d, want 0\noutput: %s", res.rc, res.out)
			}
			if res.calls != 0 {
				t.Errorf("client invoked %d times, want 0 on a non-cluster sandbox", res.calls)
			}
			if res.elapsed > 900*time.Millisecond {
				t.Errorf("took %v, want immediate exit", res.elapsed)
			}
		})
	}

	// For cluster nodes, replace wait_until_wsrep_ready with a stub so the
	// test does not spend the real 60x2s budget.
	for _, sbType := range []string{globals.SbTypeGaleraNode, globals.SbTypePxcNode} {
		for _, stubRC := range []int{0, 1} {
			t.Run(sbType+" wait rc "+strconv.Itoa(stubRC), func(t *testing.T) {
				dir := t.TempDir()
				stub := "export SBDIR=" + dir + "\nfunction wait_until_wsrep_ready { return " + strconv.Itoa(stubRC) + "; }\n"
				// #nosec G306 -- test fixture
				if err := os.WriteFile(path.Join(dir, "sb_include"), []byte(stub), 0644); err != nil {
					t.Fatal(err)
				}
				script := writeScript(t, dir, sbType)
				res := runBash(t, bash, dir, script, nil)
				if stubRC == 0 {
					if res.rc != 0 {
						t.Errorf("exit code = %d, want 0\noutput: %s", res.rc, res.out)
					}
					return
				}
				if res.rc == 0 {
					t.Errorf("a cluster node that never becomes ready must fail the script")
				}
				if !strings.Contains(res.out, dir) || !strings.Contains(res.out, "msandbox.err") {
					t.Errorf("failure message must name the node and its error log, got: %s", res.out)
				}
			})
		}
	}
}
