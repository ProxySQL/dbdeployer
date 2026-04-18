// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/compare"
)

type boolMap map[bool]string
type VersionCollectionInfo struct {
	foundVersions         []string
	requestedShortVersion string
	expected              boolMap
}

func getShortVersion(s string) string {

	pieces := strings.Split(s, ".")
	return fmt.Sprintf("%s.%s", pieces[0], pieces[1])
}

func makeTarballCollection(info VersionCollectionInfo) []TarballDescription {
	var tbd []TarballDescription

	for _, v := range info.foundVersions {
		tbd = append(tbd, TarballDescription{
			Name:            "mysql-" + v,
			Checksum:        "",
			OperatingSystem: "linux",
			Url:             "",
			Flavor:          common.MySQLFlavor,
			Minimal:         false,
			Size:            0,
			ShortVersion:    getShortVersion(v),
			Version:         v,
			UpdatedBy:       "",
			Notes:           "",
		})
	}

	return tbd
}

func TestFindOrGuessTarballByVersionFlavorOS(t *testing.T) {

	var versionCollections = map[string][]string{
		"8.0":     []string{"8.0.19", "8.0.20", "8.0.22", "8.0.23"},
		"5.6":     []string{"5.6.31", "5.6.33"},
		"5.7":     []string{"5.7.31", "5.7.33"},
		"5.7-8.0": []string{"5.7.31", "5.7.33", "8.0.19", "8.0.20", "8.0.22", "8.0.23"},
	}
	var versionCollectionData = []VersionCollectionInfo{
		{
			foundVersions:         versionCollections["8.0"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "8.0.24", false: "8.0.23"},
		},
		{
			foundVersions:         versionCollections["5.7-8.0"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "8.0.24", false: "8.0.23"},
		},
		{
			foundVersions:         versionCollections["5.7-8.0"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "5.7.34", false: "5.7.33"},
		},
		{
			foundVersions:         versionCollections["5.7"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "5.7.34", false: "5.7.33"},
		},
		{
			foundVersions:         versionCollections["8.0"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.7"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "8.0",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "5.7",
			expected:              boolMap{true: "", false: ""},
		},
		{
			foundVersions:         versionCollections["5.6"],
			requestedShortVersion: "5.6",
			expected:              boolMap{true: "", false: "5.6.33"},
		},
	}
	saveTarballCollection := DefaultTarballRegistry.Tarballs

	for _, data := range versionCollectionData {
		tbd := makeTarballCollection(data)
		DefaultTarballRegistry.Tarballs = tbd

		for guess, expected := range data.expected {

			tb, _ := FindOrGuessTarballByVersionFlavorOS(
				data.requestedShortVersion,
				common.MySQLFlavor,
				"linux", "amd64", false, !guess, guess)
			label := fmt.Sprintf("versions %s - requested '%s' - guess '%v'",
				data.foundVersions,
				data.requestedShortVersion,
				guess)
			compare.OkEqualString(
				label,
				tb.Version, expected, t)
		}

	}

	DefaultTarballRegistry.Tarballs = saveTarballCollection
}

func TestNewMySQLVersionsRecognized(t *testing.T) {
	versions := []string{"8.4", "9.0", "9.1", "9.2", "9.3", "9.4", "9.5"}
	for _, v := range versions {
		result := isAllowedForGuessing(v)
		compare.OkEqualBool(fmt.Sprintf("version %s allowed for guessing", v), result, true, t)
	}
}

func TestTarballRegistry(t *testing.T) {
	// CDN checks are inherently flaky (timeouts, EOF, rate limiting).
	// Allow a small number of transient failures without failing the test.
	maxAllowedFailures := 3
	failures := 0
	transient403s := 0

	for _, tarball := range DefaultTarballRegistry.Tarballs {
		size, err := CheckRemoteUrl(tarball.Url)
		if err != nil {
			// HTTP 403 from MySQL CDN is rate-limiting, not a broken URL.
			// Count separately and don't let it fail the test.
			if strings.Contains(err.Error(), "received code 403") {
				transient403s++
				t.Logf("WARN - tarball %s rate-limited by CDN (403): %s", tarball.Name, err)
			} else {
				failures++
				t.Logf("WARN - tarball %s check failed (%d/%d allowed): %s", tarball.Name, failures, maxAllowedFailures, err)
			}
		} else {
			t.Logf("ok - tarball %s found", tarball.Name)
			if size == 0 {
				t.Logf("note - size 0 for tarball %s (size not recorded in registry)", tarball.Name)
			}
		}
		// Small delay to avoid triggering CDN rate limits
		time.Sleep(50 * time.Millisecond)
	}

	if transient403s > 0 {
		t.Logf("INFO: %d tarballs returned HTTP 403 (CDN rate limit) — not counted as failures", transient403s)
	}
	if failures > maxAllowedFailures {
		t.Errorf("too many tarball URL failures: %d (max allowed: %d)", failures, maxAllowedFailures)
	}
}

func TestMergeCollection(t *testing.T) {
	type args struct {
		oldest TarballCollection
		newest TarballCollection
	}
	var (
		oneItem         = TarballCollection{Tarballs: []TarballDescription{{Name: "one"}}}
		anotherItem     = TarballCollection{Tarballs: []TarballDescription{{Name: "first"}}}
		twoItems        = TarballCollection{Tarballs: []TarballDescription{{Name: "one"}, {Name: "two"}}}
		anotherTwoItems = TarballCollection{Tarballs: []TarballDescription{{Name: "first"}, {Name: "second"}}}
		threeItems      = TarballCollection{Tarballs: []TarballDescription{{Name: "one"}, {Name: "two"}, {Name: "three"}}}

		twoItemsSameResult      = TarballCollection{DbdeployerVersion: common.VersionDef, Tarballs: []TarballDescription{{Name: "one"}}}
		twoItemsDifferentResult = TarballCollection{DbdeployerVersion: common.VersionDef, Tarballs: []TarballDescription{{Name: "one"}, {Name: "first"}}}
		threeItemsResult        = TarballCollection{DbdeployerVersion: common.VersionDef, Tarballs: []TarballDescription{{Name: "one"}, {Name: "two"}, {Name: "three"}}}
		fiveItemsResult         = TarballCollection{DbdeployerVersion: common.VersionDef, Tarballs: []TarballDescription{{Name: "one"}, {Name: "two"}, {Name: "three"}, {Name: "first"}, {Name: "second"}}}
	)
	tests := []struct {
		name    string
		args    args
		want    TarballCollection
		wantErr bool
	}{
		{"both-empty", args{TarballCollection{}, TarballCollection{}}, TarballCollection{}, true},
		{"origin-empty", args{TarballCollection{}, TarballCollection{}}, TarballCollection{Tarballs: []TarballDescription{{}}}, true},
		{"additional-empty", args{TarballCollection{Tarballs: []TarballDescription{{}}}, TarballCollection{}}, TarballCollection{}, true},
		{"one-item-same", args{oneItem, oneItem}, twoItemsSameResult, false},
		{"one-item-different", args{oneItem, anotherItem}, twoItemsDifferentResult, false},
		{"two-three-items-common", args{twoItems, threeItems}, threeItemsResult, false},
		{"two-three-items-different", args{threeItems, anotherTwoItems}, fiveItemsResult, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeTarballCollection(tt.args.oldest, tt.args.newest)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) && err == nil {
				t.Errorf("MergeCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTarballUrlInfo(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantName    string
		wantVersion string
		wantShort   string
		wantOS      string
		wantArch    string
		wantFlavor  string
		wantMinimal bool
		wantErr     bool
	}{
		{
			name:        "mysql-linux-amd64",
			url:         "https://cdn.mysql.com/Downloads/MySQL-8.4/mysql-8.4.8-linux-glibc2.17-x86_64.tar.xz",
			wantName:    "mysql-8.4.8-linux-glibc2.17-x86_64.tar.xz",
			wantVersion: "8.4.8",
			wantShort:   "8.4",
			wantOS:      "Linux",
			wantArch:    "amd64",
			wantFlavor:  "mysql",
			wantMinimal: false,
		},
		{
			name:        "mysql-linux-amd64-minimal",
			url:         "https://cdn.mysql.com/Downloads/MySQL-8.4/mysql-8.4.8-linux-glibc2.17-x86_64-minimal.tar.xz",
			wantName:    "mysql-8.4.8-linux-glibc2.17-x86_64-minimal.tar.xz",
			wantVersion: "8.4.8",
			wantShort:   "8.4",
			wantOS:      "Linux",
			wantArch:    "amd64",
			wantFlavor:  "mysql",
			wantMinimal: true,
		},
		{
			name:        "mysql-macos-arm64",
			url:         "https://cdn.mysql.com/Downloads/MySQL-8.4/mysql-8.4.8-macos15-arm64.tar.gz",
			wantName:    "mysql-8.4.8-macos15-arm64.tar.gz",
			wantVersion: "8.4.8",
			wantShort:   "8.4",
			wantOS:      "Darwin",
			wantArch:    "arm64",
			wantFlavor:  "mysql",
			wantMinimal: false,
		},
		{
			name:        "percona-linux-amd64-minimal",
			url:         "https://downloads.percona.com/downloads/Percona-Server-8.0/Percona-Server-8.0.35-27/binary/tarball/Percona-Server-8.0.35-27-Linux.x86_64.glibc2.17-minimal.tar.gz",
			wantName:    "Percona-Server-8.0.35-27-Linux.x86_64.glibc2.17-minimal.tar.gz",
			wantVersion: "8.0.35",
			wantShort:   "8.0",
			wantOS:      "Linux",
			wantArch:    "amd64",
			wantFlavor:  "percona",
			wantMinimal: true,
		},
		{
			name:    "invalid-no-version",
			url:     "https://example.com/some-tarball-without-version.tar.gz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTarballUrlInfo(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTarballUrlInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", got.Version, tt.wantVersion)
			}
			if got.ShortVersion != tt.wantShort {
				t.Errorf("ShortVersion = %q, want %q", got.ShortVersion, tt.wantShort)
			}
			if got.OperatingSystem != tt.wantOS {
				t.Errorf("OperatingSystem = %q, want %q", got.OperatingSystem, tt.wantOS)
			}
			if got.Arch != tt.wantArch {
				t.Errorf("Arch = %q, want %q", got.Arch, tt.wantArch)
			}
			if got.Flavor != tt.wantFlavor {
				t.Errorf("Flavor = %q, want %q", got.Flavor, tt.wantFlavor)
			}
			if got.Minimal != tt.wantMinimal {
				t.Errorf("Minimal = %v, want %v", got.Minimal, tt.wantMinimal)
			}
			if got.Url != tt.url {
				t.Errorf("Url = %q, want %q", got.Url, tt.url)
			}
		})
	}
}
