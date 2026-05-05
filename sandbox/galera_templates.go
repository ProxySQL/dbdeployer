// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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

//go:build go1.16
// +build go1.16

package sandbox

import (
	_ "embed"

	"github.com/ProxySQL/dbdeployer/globals"
)

var (
	//go:embed templates/galera/galera_start.gotxt
	galeraStartTemplate string

	//go:embed templates/galera/check_galera_nodes.gotxt
	galeraCheckNodesTemplate string

	//go:embed templates/galera/galera_replication.gotxt
	galeraReplicationTemplate string

	GaleraTemplates = TemplateCollection{
		globals.TmplGaleraCheckNodes: TemplateDesc{
			Description: "Checks the status of Galera replication",
			Notes:       "",
			Contents:    galeraCheckNodesTemplate,
		},
		globals.TmplGaleraReplication: TemplateDesc{
			Description: "Replication options for MariaDB Galera",
			Notes:       "",
			Contents:    galeraReplicationTemplate,
		},
		globals.TmplGaleraStart: TemplateDesc{
			Description: "start all nodes in a Galera cluster",
			Notes:       "",
			Contents:    galeraStartTemplate,
		},
	}
)
