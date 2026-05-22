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
