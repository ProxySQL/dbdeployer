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
	"strings"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/providers"
)

// DeployProxySQLForTopology creates a ProxySQL sandbox configured for a MySQL topology.
//
// Parameters:
//   - sandboxDir: parent sandbox directory (e.g. ~/sandboxes/rsandbox_8_4_4)
//   - masterPort: MySQL master port
//   - slavePorts: MySQL slave ports (empty for single topology)
//   - proxysqlPort: port for ProxySQL admin interface (0 = auto-assign)
//   - host: bind address (typically "127.0.0.1")
func DeployProxySQLForTopology(sandboxDir string, masterPort int, slavePorts []int, proxysqlPort int, host string) error {
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
		proxysqlPort = 6032
	}
	// Find 2 consecutive free ports (admin + mysql) to avoid TIME_WAIT conflicts
	freePort, err := common.FindFreePort(proxysqlPort, []int{}, 2)
	if err == nil {
		proxysqlPort = freePort
	}

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
			"monitor_user":     "msandbox",
			"monitor_password": "msandbox",
			"backends":         strings.Join(backendParts, ","),
		},
	}

	_, err = p.CreateSandbox(config)
	if err != nil {
		return fmt.Errorf("creating ProxySQL sandbox: %w", err)
	}

	// Start ProxySQL
	if err := p.StartSandbox(proxysqlDir); err != nil {
		return fmt.Errorf("starting ProxySQL: %w", err)
	}

	fmt.Printf("ProxySQL deployed in %s (admin port: %d, mysql port: %d)\n", proxysqlDir, proxysqlPort, proxysqlPort+1)
	return nil
}
