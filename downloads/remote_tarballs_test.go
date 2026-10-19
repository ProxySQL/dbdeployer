// DBDeployer - The MySQL Sandbox
// Copyright © 2026 René Cannaò
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

import (
	"strings"
	"testing"
)

func TestValidateTarballRequestReleaseSeries(t *testing.T) {
	for _, tarballType := range []TarballType{TtMysql, TtCluster, TtShell} {
		for _, version := range []string{"5.7", "8.0", "8.4", "9.2", "9.7", "26.7", "27.1"} {
			for _, os := range []string{"linux", "darwin"} {
				if _, err := validateTarballRequest(tarballType, version, os); err != nil {
					t.Errorf("%s %s on %s should be accepted: %v", tarballType, version, os, err)
				}
			}
		}
		for _, version := range []string{"", "9", "9.7.2", "26.7.0", "9.x", "v9.7", "9.7 "} {
			if _, err := validateTarballRequest(tarballType, version, "linux"); err == nil {
				t.Errorf("%s %q should be rejected", tarballType, version)
			}
		}
	}
	if _, err := validateTarballRequest("nosuchtype", "8.4", "linux"); err == nil {
		t.Errorf("unknown tarball type should be rejected")
	}
	if _, err := validateTarballRequest(TtMysql, "8.4", "windows"); err == nil {
		t.Errorf("unknown OS should be rejected")
	}
}

// Text shaped like the downloads page once converted by html2text.
const sampleMysqlPage = `
| Linux - Generic (glibc 2.28) (x86, 64-bit), Compressed TAR Archive | 9.7.2 | 1.1G | Download |
| (mysql-9.7.2-linux-glibc2.28-x86_64.tar.xz) | MD5: 0123456789abcdef0123456789abcdef | Signature |
| Linux - Generic (glibc 2.28) (x86, 64-bit), Compressed TAR Archive Minimal Install | 9.7.2 | 60M | Download |
| (mysql-9.7.2-linux-glibc2.28-x86_64-minimal.tar.xz) | MD5: fedcba9876543210fedcba9876543210 | Signature |
| Linux - Generic (glibc 2.28) (ARM, 64-bit), Compressed TAR Archive Minimal Install | 9.7.2 | 58M | Download |
| (mysql-9.7.2-linux-glibc2.28-aarch64-minimal.tar.xz) | MD5: 00112233445566778899aabbccddeeff | Signature |
`

func TestParseRemoteTarballPage(t *testing.T) {
	list, err := parseRemoteTarballPage(TtMysql, "9.7", OsLinux, sampleMysqlPage, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 tarballs, got %d", len(list))
	}
	minimal := 0
	for _, tb := range list {
		if tb.ShortVersion != "9.7" || tb.Version != "9.7.2" {
			t.Errorf("unexpected version %s/%s for %s", tb.ShortVersion, tb.Version, tb.Name)
		}
		if tb.Flavor != "mysql" || tb.OperatingSystem != "Linux" {
			t.Errorf("unexpected flavor/OS %s/%s for %s", tb.Flavor, tb.OperatingSystem, tb.Name)
		}
		wantUrl := baseDownloadUrl + "/MySQL-9.7/" + tb.Name
		if tb.Url != wantUrl {
			t.Errorf("expected URL %s, got %s", wantUrl, tb.Url)
		}
		if tb.Minimal {
			minimal++
		}
	}
	if minimal != 2 {
		t.Errorf("expected 2 minimal tarballs, got %d", minimal)
	}
	if list[0].Arch != "amd64" || list[2].Arch != "aarch64" {
		t.Errorf("unexpected arch normalization: %s, %s", list[0].Arch, list[2].Arch)
	}
}

// The download site answers a request for a series it does not list with the
// page of the current default release. That must not be registered as the
// requested series.
func TestParseRemoteTarballPageOtherSeries(t *testing.T) {
	for _, version := range []string{"9.3", "9.2", "99.9"} {
		list, err := parseRemoteTarballPage(TtMysql, version, OsLinux, sampleMysqlPage, "", false)
		if err == nil {
			t.Fatalf("series %s: expected an error, got %d tarballs", version, len(list))
		}
		if !strings.Contains(err.Error(), "[9.7]") {
			t.Errorf("series %s: error should name the series found on the page: %v", version, err)
		}
	}
}

func TestParseRemoteTarballPageNoMatches(t *testing.T) {
	if _, err := parseRemoteTarballPage(TtMysql, "9.7", OsLinux, "nothing to see here", "", false); err == nil {
		t.Fatalf("expected an error for a page without tarballs")
	}
}
