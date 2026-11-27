// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ops

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProxySQL/dbdeployer/common"
)

func TestUnpackDistinctMySQLVersions(t *testing.T) {
	root := t.TempDir()
	archives := filepath.Join(root, "downloads")
	versionedPath := filepath.Join(archives, "1.2.3")
	binaries := filepath.Join(root, "binaries")
	for _, dir := range []string{archives, versionedPath, binaries} {
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, version := range []string{"8.4.13", "8.4.13.1", "9.7.4.1", "26.10.0.1"} {
		name := "mysql-" + version + "-linux-glibc2.28-x86_64-minimal"
		filename := filepath.Join(archives, name+".tar.gz")
		if version == "26.10.0.1" {
			filename = filepath.Join(versionedPath, name+".tar.gz")
		}
		writeVersionArchive(t, filename, name, version)
		if err := UnpackTarball(UnpackOptions{SandboxBinary: binaries, TarballName: filename}); err != nil {
			t.Fatalf("unpack MySQL %s: %v", version, err)
		}
	}
	for _, version := range []string{"8.4.13", "8.4.13.1", "9.7.4.1", "26.10.0.1"} {
		data, err := os.ReadFile(filepath.Join(binaries, version, "VERSION"))
		if err != nil || string(data) != version {
			t.Errorf("unpacked MySQL %s: got %q, err %v", version, data, err)
		}
	}
	if got := common.LatestVersion(binaries, "8.4"); got != "8.4.13.1" {
		t.Errorf("latest installed MySQL 8.4 = %q, want 8.4.13.1", got)
	}
}

func writeVersionArchive(t *testing.T, filename, name, version string) {
	t.Helper()
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name + "/", Typeflag: tar.TypeDir, Mode: 0755}); err != nil {
		t.Fatal(err)
	}
	if err := tw.WriteHeader(&tar.Header{Name: name + "/VERSION", Mode: 0644, Size: int64(len(version))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(version)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}
