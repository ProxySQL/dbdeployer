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

//go:build go1.16
// +build go1.16

package sandbox

import (
	_ "embed"

	"github.com/ProxySQL/dbdeployer/globals"
)

// Templates for InnoDB Cluster

var (
	//go:embed templates/cluster/innodb_cluster_options.gotxt
	innodbClusterOptionsTemplate string

	//go:embed templates/cluster/init_cluster.gotxt
	initClusterTemplate string

	//go:embed templates/cluster/check_cluster.gotxt
	checkClusterTemplate string

	//go:embed templates/cluster/router_start.gotxt
	routerStartTemplate string

	//go:embed templates/cluster/router_stop.gotxt
	routerStopTemplate string

	InnoDBClusterTemplates = TemplateCollection{
		globals.TmplInnoDBClusterOptions: TemplateDesc{
			Description: "MySQL server options for InnoDB Cluster nodes",
			Notes:       "Enables GTID, report_host, and disables non-InnoDB storage engines",
			Contents:    innodbClusterOptionsTemplate,
		},
		globals.TmplInitCluster: TemplateDesc{
			Description: "Initialize InnoDB Cluster via MySQL Shell",
			Notes:       "Uses dba.createCluster() and cluster.addInstance()",
			Contents:    initClusterTemplate,
		},
		globals.TmplCheckCluster: TemplateDesc{
			Description: "Check InnoDB Cluster status via MySQL Shell",
			Notes:       "",
			Contents:    checkClusterTemplate,
		},
		globals.TmplRouterStart: TemplateDesc{
			Description: "Start MySQL Router for InnoDB Cluster",
			Notes:       "",
			Contents:    routerStartTemplate,
		},
		globals.TmplRouterStop: TemplateDesc{
			Description: "Stop MySQL Router for InnoDB Cluster",
			Notes:       "",
			Contents:    routerStopTemplate,
		},
	}
)
