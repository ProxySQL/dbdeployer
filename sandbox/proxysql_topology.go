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

package sandbox

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/ProxySQL/dbdeployer/providers/postgresql"
)

// DeployProxySQLForTopology creates a ProxySQL sandbox configured for a MySQL topology.
//
// Parameters:
//   - sandboxDir: parent sandbox directory (e.g. ~/sandboxes/rsandbox_8_4_4)
//   - masterPort: MySQL master port
//   - slavePorts: MySQL slave ports (empty for single topology)
//   - proxysqlPort: port for ProxySQL admin interface (0 = auto-assign)
//   - host: bind address (typically "127.0.0.1")
//   - backendProvider: database provider name (e.g. "mysql", "postgresql", or "" for mysql default)
// DeployProxySQLForTopology creates a ProxySQL sandbox configured for a database topology.
//
// Parameters:
//   - sandboxDir: parent sandbox directory (e.g. ~/sandboxes/rsandbox_8_4_4)
//   - masterPort: primary/master port
//   - slavePorts: replica/slave ports (empty for single topology)
//   - proxysqlPort: port for ProxySQL admin interface (0 = auto-assign)
//   - host: bind address (typically "127.0.0.1")
//   - backendProvider: database provider name ("mysql", "postgresql", or "" for mysql default)
//   - topology: the deployment topology ("replication", "innodb-cluster", "group", etc.)
//
// For Group Replication and InnoDB Cluster topologies, ProxySQL is configured with
// mysql_group_replication_hostgroups for automatic failover-aware routing.
func DeployProxySQLForTopology(sandboxDir string, masterPort int, slavePorts []int, proxysqlPort int, host string, backendProvider string, topology ...string) error {
	reg := providers.DefaultRegistry
	p, err := reg.Get("proxysql")
	if err != nil {
		return fmt.Errorf("ProxySQL provider not available: %w", err)
	}

	if _, err := p.FindBinary(""); err != nil {
		return fmt.Errorf("proxysql binary not found: %w", err)
	}

	proxysqlDir := path.Join(sandboxDir, "proxysql")

	if proxysqlPort == 0 {
		if backendProvider == postgresql.ProviderName {
			proxysqlPort = postgresql.ProxySQLAdminPort(masterPort)
		} else {
			proxysqlPort = 6032
		}
	}
	installedPorts := []int{masterPort}
	installedPorts = append(installedPorts, slavePorts...)
	// Find 2 consecutive free ports (admin + proxy) to avoid TIME_WAIT conflicts
	freePort, err := common.FindFreePort(proxysqlPort, installedPorts, 2)
	if err == nil {
		proxysqlPort = freePort
	}
	common.DeployDebugf("ProxySQL ports: admin=%d proxy=%d dir=%s\n", proxysqlPort, proxysqlPort+1, proxysqlDir)

	// Build backends: master = HG 0, slaves = HG 1
	var backendParts []string
	backendParts = append(backendParts, fmt.Sprintf("%s:%d:0", host, masterPort))
	for _, slavePort := range slavePorts {
		backendParts = append(backendParts, fmt.Sprintf("%s:%d:1", host, slavePort))
	}

	config := providers.SandboxConfig{
		Version:    "system",
		Dir:        proxysqlDir,
		Port:       proxysqlPort,
		AdminPort:  proxysqlPort,
		Host:       host,
		DbUser:     "admin",
		DbPassword: "admin",
		Options: map[string]string{
			"monitor_user":     "rsandbox",
			"monitor_password": "rsandbox",
			"backends":         strings.Join(backendParts, ","),
			"backend_provider": backendProvider,
			"topology":         topologyName(topology),
		},
	}
	if backendProvider == postgresql.ProviderName {
		primaryDir := path.Join(sandboxDir, "primary")
		desc, err := common.ReadSandboxDescription(primaryDir)
		if err != nil {
			return fmt.Errorf("reading primary sandbox description for PostgreSQL client paths: %w", err)
		}
		psqlPath, err := postgresql.ResolvePsqlBinary(desc.Basedir, desc.Version)
		if err != nil {
			return err
		}
		config.Options["pg_psql"] = psqlPath
		config.Options["pg_bindir"] = filepath.Dir(psqlPath)
		config.Options["pg_libdir"] = postgresql.LibraryPath(desc.Basedir, desc.Version)
	}

	_, err = p.CreateSandbox(config)
	if err != nil {
		return fmt.Errorf("creating ProxySQL sandbox: %w", err)
	}
	common.DeployDebugf("ProxySQL CreateSandbox complete, starting\n")
	proxysqlStart := time.Now()

	// Start ProxySQL
	if err := p.StartSandbox(proxysqlDir); err != nil {
		return fmt.Errorf("starting ProxySQL: %w", err)
	}
	common.DeployDebugSince("ProxySQL StartSandbox complete", proxysqlStart)

	fmt.Printf("ProxySQL deployed in %s (admin port: %d, mysql port: %d)\n", proxysqlDir, proxysqlPort, proxysqlPort+1)
	return nil
}

func topologyName(topology []string) string {
	if len(topology) > 0 {
		return topology[0]
	}
	return ""
}
