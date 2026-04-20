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
	"strings"
	"testing"
)

func TestCheckMysqlShellCompatibility(t *testing.T) {
	cases := []struct {
		label     string
		shell     string
		server    string
		wantErr   bool
		wantMatch string // substring expected in error when wantErr is true
	}{
		// The exact scenario from issue #87
		{"shell 8.0.36 rejects server 8.4.8", "8.0.36", "8.4.8", true, "too old"},
		{"shell 8.0.36 rejects server 9.1.0", "8.0.36", "9.1.0", true, "too old"},

		// Matching major.minor — allowed
		{"shell 8.4.8 accepts server 8.4.8", "8.4.8", "8.4.8", false, ""},
		{"shell 8.0.36 accepts server 8.0.42", "8.0.36", "8.0.42", false, ""},
		{"shell 9.5.0 accepts server 9.5.0", "9.5.0", "9.5.0", false, ""},

		// Shell newer than server — allowed
		{"shell 8.4.0 accepts server 8.0.42", "8.4.0", "8.0.42", false, ""},
		{"shell 9.5.0 accepts server 8.4.8", "9.5.0", "8.4.8", false, ""},

		// Shell older than server by minor — rejected
		{"shell 8.3.0 rejects server 8.4.0", "8.3.0", "8.4.0", true, "too old"},

		// Malformed inputs
		{"invalid shell version", "not-a-version", "8.4.8", true, "invalid mysqlsh version"},
		{"invalid server version", "8.4.8", "nope", true, "invalid server version"},
	}
	for _, c := range cases {
		err := checkMysqlShellCompatibility(c.shell, c.server)
		if c.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got nil", c.label)
				continue
			}
			if c.wantMatch != "" && !strings.Contains(err.Error(), c.wantMatch) {
				t.Errorf("%s: error %q does not contain %q", c.label, err.Error(), c.wantMatch)
			}
		} else if err != nil {
			t.Errorf("%s: unexpected error: %s", c.label, err)
		}
	}
}

func TestMysqlShellVersionRegexp(t *testing.T) {
	cases := []struct {
		label string
		out   string
		want  string
	}{
		{
			"classic 8.0.36 output",
			"mysqlsh   Ver 8.0.36 for Linux on x86_64 - for MySQL 8.0.36 (Source distribution)",
			"8.0.36",
		},
		{
			"8.4.8 output",
			"mysqlsh   Ver 8.4.8 for Linux on x86_64 - for MySQL 8.4.8 (MySQL Community Server)",
			"8.4.8",
		},
		{
			"9.5.0 output",
			"MySQL Shell Ver 9.5.0 for Linux on x86_64 - for MySQL 9.5.0",
			"9.5.0",
		},
	}
	for _, c := range cases {
		m := mysqlShellVersionRegexp.FindStringSubmatch(c.out)
		if len(m) < 2 {
			t.Errorf("%s: no match in %q", c.label, c.out)
			continue
		}
		if m[1] != c.want {
			t.Errorf("%s: got %q, want %q", c.label, m[1], c.want)
		}
	}
}
