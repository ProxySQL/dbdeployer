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
	_ "embed"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/globals"
)

//go:embed templates/test_replication.gotxt
var testReplicationTemplate string

type TestReplicationOptions struct {
	SandboxDir string
	ShellPath  string
}

func GenerateTestReplicationScript(opts TestReplicationOptions) (string, error) {
	data := common.StringMap{
		"ShellPath":    opts.ShellPath,
		"Copyright":    globals.ShellScriptCopyright,
		"AppVersion":   common.VersionDef,
		"DateTime":     time.Now().Format(time.UnixDate),
		"TemplateName": "test_replication",
		"SandboxDir":   opts.SandboxDir,
		"PrimaryAbbr":  "primary",
		"PrimaryLabel": "primary",
		"ReplicaAbbr":  "replica",
		"ReplicaLabel": "replica",
	}
	return common.SafeTemplateFill("test_replication", testReplicationTemplate, data)
}
