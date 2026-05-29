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

package postgresql

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LibraryPath returns LD_LIBRARY_PATH entries for a PostgreSQL basedir.
// Client tools (psql) need libpq from <basedir>/lib; server binaries and
// extensions need <basedir>/lib/postgresql/<major>/lib on Linux.
func LibraryPath(basedir, version string) string {
	flatLib := filepath.Join(basedir, "lib")
	if runtime.GOOS != "linux" {
		return flatLib
	}
	major := strings.Split(version, ".")[0]
	pgLib := filepath.Join(basedir, "lib", "postgresql", major, "lib")
	return flatLib + string(filepath.ListSeparator) + pgLib
}

// PgBinDir returns the directory containing Debian-packaged PostgreSQL client tools.
func PgBinDir(basedir, version string) string {
	if runtime.GOOS != "linux" {
		return filepath.Join(basedir, "bin")
	}
	major := strings.Split(version, ".")[0]
	return filepath.Join(basedir, "lib", "postgresql", major, "bin")
}

// ResolvePsqlBinary locates the unpacked psql binary, skipping Debian pg_wrapper scripts.
func ResolvePsqlBinary(basedir, version string) (string, error) {
	candidates := []string{
		filepath.Join(PgBinDir(basedir, version), "psql"),
		filepath.Join(basedir, "bin", "psql"),
	}
	for _, candidate := range candidates {
		ok, err := validPsqlBinary(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf(
		"bundled psql not found under %s — re-run unpack with server, client, and libpq5 debs:\n"+
			"    dbdeployer unpack --provider=postgresql postgresql-NN_*.deb postgresql-client-NN_*.deb libpq5_*.deb",
		basedir,
	)
}

func validPsqlBinary(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("%s is a directory", path)
	}

	resolved := path
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err = filepath.EvalSymlinks(path)
		if err != nil {
			return false, err
		}
	}
	if isShellScript(resolved) {
		return false, nil
	}
	return true, nil
}

func isShellScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 2)
	n, _ := f.Read(buf)
	return n >= 2 && buf[0] == '#' && buf[1] == '!'
}
