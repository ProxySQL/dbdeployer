# Phase 2b: ProxySQL Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add ProxySQL as the first non-MySQL provider in dbdeployer, supporting standalone ProxySQL sandboxes and topology-aware deployment with MySQL replication.

**Architecture:** ProxySQL provider uses system-installed binaries (no tarball management). Deploys local ProxySQL instances with generated config files, data directories, and lifecycle scripts. Topology-aware deployment (`--with-proxysql`) automatically configures ProxySQL backends based on the MySQL topology type.

**Tech Stack:** Go 1.22+, ProxySQL admin interface (MySQL protocol), existing Cobra CLI

**Spec:** `docs/superpowers/specs/2026-03-23-dbdeployer-revitalization-design.md`

---

## Key Design Decisions

### Binary management
- First iteration: ProxySQL must be installed on the system (deb/rpm/compiled)
- Provider locates `proxysql` binary in PATH or user-configured location
- `Unpack()` is a no-op — tarball support deferred to when ProxySQL distributes tarballs

### ProxySQL sandbox structure
```
~/sandboxes/proxysql_2_7_0/
  proxysql.cnf          # generated config
  data/                  # ProxySQL SQLite datadir
  start                  # lifecycle script
  stop                   #
  status                 #
  use                    # connects to admin interface via mysql client
  use_proxy              # connects through ProxySQL's MySQL port
  my.proxy.cnf           # client defaults for admin connection
```

### Topology-aware config generation
ProxySQL config varies by MySQL topology:

| MySQL Topology | Hostgroups | Monitoring |
|---------------|-----------|------------|
| Single | HG 0 only (one backend) | Basic health check |
| Replication | HG 0 = writer (master), HG 1 = readers (slaves) | read_only + replication lag |
| Group Replication | HG 0 = writer, HG 1 = readers | group_replication monitoring |

No query rules are generated — users configure those themselves.

### Monitor user
Uses the existing `msandbox` user for backend monitoring (already has SELECT privileges on all nodes).

---

## File Structure

### New files:
```
providers/proxysql/
  proxysql.go           # ProxySQLProvider implementing Provider
  proxysql_test.go      # unit tests
  config.go             # config file generation for different topologies
  config_test.go        # config generation tests
  templates/
    proxysql.cnf.gotxt  # ProxySQL config template
    start.gotxt         # start script template
    stop.gotxt          # stop script template
    status.gotxt        # status script template
    use.gotxt           # admin connection script
    use_proxy.gotxt     # proxy connection script
```

### Files to modify:
```
providers/provider.go      # Extend Provider interface with CreateSandbox, Start, Stop
cmd/root.go               # Register ProxySQL provider
cmd/single.go             # Add --with-proxysql flag
cmd/replication.go        # Add --with-proxysql flag
sandbox/replication.go    # Hook for post-deploy ProxySQL wiring
```

---

### Task 1: Extend Provider interface with lifecycle methods

**Files:**
- Modify: `providers/provider.go`
- Modify: `providers/provider_test.go`
- Modify: `providers/mysql/mysql.go`

The Phase 2a interface only has `Name`, `ValidateVersion`, `DefaultPorts`. Now add the methods needed for ProxySQL to actually deploy sandboxes.

- [ ] **Step 1: Add SandboxConfig and lifecycle methods to Provider interface**

In `providers/provider.go`, add:

```go
// SandboxConfig holds provider-agnostic sandbox configuration.
type SandboxConfig struct {
	Version    string
	Dir        string            // sandbox directory path
	Port       int               // primary port
	AdminPort  int               // admin/management port (0 if not applicable)
	Host       string            // bind address
	DbUser     string            // admin username
	DbPassword string            // admin password
	Options    map[string]string // provider-specific key-value options
}

// SandboxInfo describes a running sandbox instance.
type SandboxInfo struct {
	Dir    string
	Port   int
	Socket string
	Status string // "running", "stopped"
}
```

Extend the Provider interface:

```go
type Provider interface {
	Name() string
	ValidateVersion(version string) error
	DefaultPorts() PortRange
	// FindBinary returns the path to the provider's main binary, or error if not found.
	FindBinary(version string) (string, error)
	// CreateSandbox deploys a new sandbox instance.
	CreateSandbox(config SandboxConfig) (*SandboxInfo, error)
	// StartSandbox starts a stopped sandbox.
	StartSandbox(dir string) error
	// StopSandbox stops a running sandbox.
	StopSandbox(dir string) error
}
```

- [ ] **Step 2: Add stub implementations to MySQLProvider**

In `providers/mysql/mysql.go`, add no-op stubs so it still compiles:

```go
func (p *MySQLProvider) FindBinary(version string) (string, error) {
	return "", fmt.Errorf("MySQLProvider.FindBinary: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, fmt.Errorf("MySQLProvider.CreateSandbox: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) StartSandbox(dir string) error {
	return fmt.Errorf("MySQLProvider.StartSandbox: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) StopSandbox(dir string) error {
	return fmt.Errorf("MySQLProvider.StopSandbox: use sandbox package directly (not yet migrated)")
}
```

- [ ] **Step 3: Update mock in provider_test.go**

Add stub methods to the mock provider so tests compile.

- [ ] **Step 4: Verify all tests pass**

Run: `go test ./providers/... -v`

- [ ] **Step 5: Commit**

```bash
git add providers/
git commit -m "feat: extend Provider interface with FindBinary, CreateSandbox, Start, Stop"
```

---

### Task 2: Create ProxySQL provider — binary detection and registration

**Files:**
- Create: `providers/proxysql/proxysql.go`
- Create: `providers/proxysql/proxysql_test.go`
- Modify: `cmd/root.go` (register proxysql provider)

- [ ] **Step 1: Create `providers/proxysql/proxysql.go`**

```go
package proxysql

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "proxysql"

type ProxySQLProvider struct{}

func NewProxySQLProvider() *ProxySQLProvider {
	return &ProxySQLProvider{}
}

func (p *ProxySQLProvider) Name() string { return ProviderName }

func (p *ProxySQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid ProxySQL version format: %q", version)
	}
	return nil
}

func (p *ProxySQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{
		BasePort:         6032, // admin port
		PortsPerInstance: 2,    // admin port + mysql port
	}
}

// FindBinary locates the proxysql binary on the system.
func (p *ProxySQLProvider) FindBinary(version string) (string, error) {
	path, err := exec.LookPath("proxysql")
	if err != nil {
		return "", fmt.Errorf("proxysql binary not found in PATH: %w", err)
	}
	return path, nil
}

func (p *ProxySQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	// Implemented in Task 3
	return nil, fmt.Errorf("not yet implemented")
}

func (p *ProxySQLProvider) StartSandbox(dir string) error {
	return fmt.Errorf("not yet implemented")
}

func (p *ProxySQLProvider) StopSandbox(dir string) error {
	return fmt.Errorf("not yet implemented")
}

func Register(reg *providers.Registry) error {
	return reg.Register(NewProxySQLProvider())
}
```

- [ ] **Step 2: Create `providers/proxysql/proxysql_test.go`**

```go
package proxysql

import (
	"testing"

	"github.com/ProxySQL/dbdeployer/providers"
)

func TestProxySQLProviderName(t *testing.T) {
	p := NewProxySQLProvider()
	if p.Name() != "proxysql" {
		t.Errorf("expected 'proxysql', got %q", p.Name())
	}
}

func TestProxySQLProviderValidateVersion(t *testing.T) {
	p := NewProxySQLProvider()
	tests := []struct {
		version string
		wantErr bool
	}{
		{"2.7.0", false},
		{"3.0.0", false},
		{"invalid", true},
	}
	for _, tt := range tests {
		err := p.ValidateVersion(tt.version)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
		}
	}
}

func TestProxySQLProviderRegister(t *testing.T) {
	reg := providers.NewRegistry()
	if err := Register(reg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	p, err := reg.Get("proxysql")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name() != "proxysql" {
		t.Errorf("expected 'proxysql', got %q", p.Name())
	}
}

func TestProxySQLFindBinary(t *testing.T) {
	p := NewProxySQLProvider()
	path, err := p.FindBinary("2.7.0")
	if err != nil {
		t.Skipf("proxysql not installed, skipping: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}
```

- [ ] **Step 3: Register ProxySQL provider in cmd/root.go**

Add alongside the MySQL registration:

```go
import proxysqlprovider "github.com/ProxySQL/dbdeployer/providers/proxysql"

// In init():
// ProxySQL registration is non-fatal — it's OK if proxysql isn't installed
_ = proxysqlprovider.Register(providers.DefaultRegistry)
```

- [ ] **Step 4: Verify**

```bash
go build -o dbdeployer .
./dbdeployer providers
```
Expected:
```
mysql           (base port: 3306, ports per instance: 3)
proxysql        (base port: 6032, ports per instance: 2)
```

- [ ] **Step 5: Commit**

```bash
git add providers/proxysql/ cmd/root.go
git commit -m "feat: add ProxySQL provider with binary detection"
```

---

### Task 3: ProxySQL sandbox creation — config generation and lifecycle scripts

**Files:**
- Create: `providers/proxysql/config.go`
- Create: `providers/proxysql/config_test.go`
- Modify: `providers/proxysql/proxysql.go` (implement CreateSandbox, StartSandbox, StopSandbox)

This is the core of the ProxySQL provider. It generates a proxysql.cnf, creates the sandbox directory structure, and writes lifecycle scripts.

- [ ] **Step 1: Create `providers/proxysql/config.go`**

Config generation function that builds a proxysql.cnf string:

```go
package proxysql

import (
	"fmt"
	"strings"
)

// BackendServer represents a MySQL backend for ProxySQL configuration.
type BackendServer struct {
	Host       string
	Port       int
	Hostgroup  int
	MaxConns   int
	Weight     int
}

// ProxySQLConfig holds all settings needed to generate proxysql.cnf.
type ProxySQLConfig struct {
	AdminHost     string
	AdminPort     int
	AdminUser     string
	AdminPassword string
	MySQLPort     int
	DataDir       string
	Backends      []BackendServer
	MonitorUser   string
	MonitorPass   string
}

// GenerateConfig produces a proxysql.cnf file content.
func GenerateConfig(cfg ProxySQLConfig) string {
	var b strings.Builder

	b.WriteString("datadir=\"" + cfg.DataDir + "\"\n\n")

	b.WriteString("admin_variables=\n{\n")
	b.WriteString(fmt.Sprintf("    admin_credentials=\"%s:%s\"\n", cfg.AdminUser, cfg.AdminPassword))
	b.WriteString(fmt.Sprintf("    mysql_ifaces=\"%s:%d\"\n", cfg.AdminHost, cfg.AdminPort))
	b.WriteString("}\n\n")

	b.WriteString("mysql_variables=\n{\n")
	b.WriteString(fmt.Sprintf("    interfaces=\"%s:%d\"\n", cfg.AdminHost, cfg.MySQLPort))
	b.WriteString(fmt.Sprintf("    monitor_username=\"%s\"\n", cfg.MonitorUser))
	b.WriteString(fmt.Sprintf("    monitor_password=\"%s\"\n", cfg.MonitorPass))
	b.WriteString("    monitor_connect_interval=2000\n")
	b.WriteString("    monitor_ping_interval=2000\n")
	b.WriteString("}\n\n")

	if len(cfg.Backends) > 0 {
		b.WriteString("mysql_servers=\n(\n")
		for i, srv := range cfg.Backends {
			b.WriteString("    {\n")
			b.WriteString(fmt.Sprintf("        address=\"%s\"\n", srv.Host))
			b.WriteString(fmt.Sprintf("        port=%d\n", srv.Port))
			b.WriteString(fmt.Sprintf("        hostgroup=%d\n", srv.Hostgroup))
			maxConns := srv.MaxConns
			if maxConns == 0 {
				maxConns = 200
			}
			b.WriteString(fmt.Sprintf("        max_connections=%d\n", maxConns))
			b.WriteString("    }")
			if i < len(cfg.Backends)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(")\n\n")
	}

	b.WriteString("mysql_users=\n(\n")
	b.WriteString("    {\n")
	b.WriteString(fmt.Sprintf("        username=\"%s\"\n", cfg.MonitorUser))
	b.WriteString(fmt.Sprintf("        password=\"%s\"\n", cfg.MonitorPass))
	b.WriteString("        default_hostgroup=0\n")
	b.WriteString("    }\n")
	b.WriteString(")\n")

	return b.String()
}
```

- [ ] **Step 2: Create `providers/proxysql/config_test.go`**

```go
package proxysql

import (
	"strings"
	"testing"
)

func TestGenerateConfigBasic(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:     "127.0.0.1",
		AdminPort:     6032,
		AdminUser:     "admin",
		AdminPassword: "admin",
		MySQLPort:     6033,
		DataDir:       "/tmp/proxysql-test",
		MonitorUser:   "msandbox",
		MonitorPass:   "msandbox",
	}
	result := GenerateConfig(cfg)
	if !strings.Contains(result, `admin_credentials="admin:admin"`) {
		t.Error("missing admin credentials")
	}
	if !strings.Contains(result, `interfaces="127.0.0.1:6033"`) {
		t.Error("missing mysql interfaces")
	}
	if !strings.Contains(result, `monitor_username="msandbox"`) {
		t.Error("missing monitor username")
	}
}

func TestGenerateConfigWithBackends(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:     "127.0.0.1",
		AdminPort:     6032,
		AdminUser:     "admin",
		AdminPassword: "admin",
		MySQLPort:     6033,
		DataDir:       "/tmp/proxysql-test",
		MonitorUser:   "msandbox",
		MonitorPass:   "msandbox",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 100},
			{Host: "127.0.0.1", Port: 3307, Hostgroup: 1, MaxConns: 100},
		},
	}
	result := GenerateConfig(cfg)
	if !strings.Contains(result, "mysql_servers=") {
		t.Error("missing mysql_servers section")
	}
	if !strings.Contains(result, "port=3306") {
		t.Error("missing first backend port")
	}
	if !strings.Contains(result, "hostgroup=1") {
		t.Error("missing reader hostgroup")
	}
}
```

- [ ] **Step 3: Implement CreateSandbox, StartSandbox, StopSandbox in proxysql.go**

Update the provider to actually create sandbox directories with config and scripts:

```go
func (p *ProxySQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	binaryPath, err := p.FindBinary(config.Version)
	if err != nil {
		return nil, err
	}

	// Create directory structure
	dataDir := filepath.Join(config.Dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	adminPort := config.AdminPort
	if adminPort == 0 {
		adminPort = config.Port
	}
	mysqlPort := adminPort + 1

	// Generate config
	proxyCfg := ProxySQLConfig{
		AdminHost:     config.Host,
		AdminPort:     adminPort,
		AdminUser:     config.DbUser,
		AdminPassword: config.DbPassword,
		MySQLPort:     mysqlPort,
		DataDir:       dataDir,
		MonitorUser:   config.Options["monitor_user"],
		MonitorPass:   config.Options["monitor_password"],
	}

	// Parse backends from options if provided
	// (populated by topology-aware deployment)
	proxyCfg.Backends = parseBackends(config.Options)

	cfgContent := GenerateConfig(proxyCfg)
	cfgPath := filepath.Join(config.Dir, "proxysql.cnf")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	// Write lifecycle scripts
	writeScript(config.Dir, "start", fmt.Sprintf(
		"#!/bin/bash\n%s --config %s -D %s &\necho $! > %s/proxysql.pid\necho 'ProxySQL started'\n",
		binaryPath, cfgPath, dataDir, config.Dir))

	writeScript(config.Dir, "stop", fmt.Sprintf(
		"#!/bin/bash\nif [ -f %s/proxysql.pid ]; then\n  kill $(cat %s/proxysql.pid) 2>/dev/null\n  rm -f %s/proxysql.pid\n  echo 'ProxySQL stopped'\nfi\n",
		config.Dir, config.Dir, config.Dir))

	writeScript(config.Dir, "status", fmt.Sprintf(
		"#!/bin/bash\nif [ -f %s/proxysql.pid ] && kill -0 $(cat %s/proxysql.pid) 2>/dev/null; then\n  echo 'ProxySQL running (pid '$(cat %s/proxysql.pid)')'\nelse\n  echo 'ProxySQL not running'\n  exit 1\nfi\n",
		config.Dir, config.Dir, config.Dir))

	writeScript(config.Dir, "use", fmt.Sprintf(
		"#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL Admin> ' \"$@\"\n",
		config.Host, adminPort, config.DbUser, config.DbPassword))

	writeScript(config.Dir, "use_proxy", fmt.Sprintf(
		"#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL> ' \"$@\"\n",
		config.Host, mysqlPort, config.Options["monitor_user"], config.Options["monitor_password"]))

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   adminPort,
		Status: "stopped",
	}, nil
}

func (p *ProxySQLProvider) StartSandbox(dir string) error {
	startScript := filepath.Join(dir, "start")
	cmd := exec.Command("bash", startScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *ProxySQLProvider) StopSandbox(dir string) error {
	stopScript := filepath.Join(dir, "stop")
	cmd := exec.Command("bash", stopScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop failed: %s: %w", string(output), err)
	}
	return nil
}

func writeScript(dir, name, content string) error {
	path := filepath.Join(dir, name)
	return os.WriteFile(path, []byte(content), 0755)
}

func parseBackends(options map[string]string) []BackendServer {
	// Format: "host1:port1:hg1,host2:port2:hg2"
	raw, ok := options["backends"]
	if !ok || raw == "" {
		return nil
	}
	var backends []BackendServer
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.Split(entry, ":")
		if len(parts) >= 3 {
			port, _ := strconv.Atoi(parts[1])
			hg, _ := strconv.Atoi(parts[2])
			backends = append(backends, BackendServer{
				Host:      parts[0],
				Port:      port,
				Hostgroup: hg,
				MaxConns:  200,
			})
		}
	}
	return backends
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./providers/... -v`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add providers/proxysql/
git commit -m "feat: implement ProxySQL sandbox creation with config generation and lifecycle scripts"
```

---

### Task 4: Add `dbdeployer deploy proxysql` command

**Files:**
- Create: `cmd/deploy_proxysql.go`

A new subcommand that deploys a standalone ProxySQL sandbox using the system-installed binary.

- [ ] **Step 1: Create `cmd/deploy_proxysql.go`**

```go
package cmd

// Adds a "dbdeployer deploy proxysql" command that:
// 1. Looks up the ProxySQL provider from the registry
// 2. Finds the proxysql binary on the system
// 3. Creates a sandbox directory in ~/sandboxes/proxysql_<port>/
// 4. Generates proxysql.cnf with admin/mysql ports
// 5. Writes lifecycle scripts (start, stop, status, use)
// 6. Optionally starts the sandbox
//
// Usage: dbdeployer deploy proxysql [--port=6032] [--admin-user=admin] [--admin-password=admin]
```

The command should use `providers.DefaultRegistry.Get("proxysql")` and call `CreateSandbox()`.

- [ ] **Step 2: Verify**

```bash
go build -o dbdeployer .
./dbdeployer deploy proxysql --port 6032
ls ~/sandboxes/proxysql_6032/
cat ~/sandboxes/proxysql_6032/proxysql.cnf
~/sandboxes/proxysql_6032/start
~/sandboxes/proxysql_6032/use -e "SELECT 1"
~/sandboxes/proxysql_6032/stop
```

- [ ] **Step 3: Commit**

```bash
git add cmd/deploy_proxysql.go
git commit -m "feat: add 'dbdeployer deploy proxysql' command"
```

---

### Task 5: Add `--with-proxysql` flag to replication deployment

**Files:**
- Modify: `cmd/replication.go` (add flag)
- Create: `sandbox/proxysql_topology.go` (topology wiring logic)

This is the topology-aware deployment. When `--with-proxysql` is passed to `dbdeployer deploy replication`, after the MySQL replication sandbox is created, a ProxySQL sandbox is deployed and configured with the MySQL backends.

- [ ] **Step 1: Create `sandbox/proxysql_topology.go`**

Logic to wire ProxySQL to a MySQL replication sandbox:

```go
package sandbox

// DeployProxySQLForReplication creates a ProxySQL sandbox configured
// for a MySQL replication topology.
//
// Parameters:
//   - replicationDir: path to the MySQL replication sandbox (e.g. ~/sandboxes/rsandbox_8_4_4)
//   - masterPort: MySQL master port
//   - slavePorts: MySQL slave ports
//   - proxysqlPort: port for ProxySQL admin interface
//
// ProxySQL configuration:
//   - Hostgroup 0: writer (master)
//   - Hostgroup 1: readers (slaves)
//   - Monitor user: msandbox/msandbox
//   - No query rules (user configures)
```

- [ ] **Step 2: Add `--with-proxysql` flag to cmd/replication.go**

Add a `--with-proxysql` boolean flag. When set, after the replication sandbox deploys successfully, call the topology wiring function to deploy ProxySQL alongside it.

- [ ] **Step 3: Test end-to-end**

```bash
go build -o dbdeployer .
./dbdeployer deploy replication 8.4.4 --sandbox-binary=$HOME/opt/mysql --with-proxysql
# Verify MySQL replication works
~/sandboxes/rsandbox_8_4_4/check_slaves
# Verify ProxySQL sandbox exists
ls ~/sandboxes/rsandbox_8_4_4/proxysql/
# Verify ProxySQL is running and has backends
~/sandboxes/rsandbox_8_4_4/proxysql/use -e "SELECT * FROM mysql_servers"
# Connect through ProxySQL to MySQL
~/sandboxes/rsandbox_8_4_4/proxysql/use_proxy -e "SELECT @@hostname, @@port"
# Cleanup
./dbdeployer delete all --skip-confirm
```

- [ ] **Step 4: Commit**

```bash
git add sandbox/proxysql_topology.go cmd/replication.go
git commit -m "feat: add --with-proxysql flag for topology-aware ProxySQL deployment"
```

---

### Task 6: Add `--with-proxysql` to single deployment

**Files:**
- Modify: `cmd/single.go` (add flag)

Simpler than replication — just one backend in hostgroup 0.

- [ ] **Step 1: Add `--with-proxysql` flag to cmd/single.go**

When set, deploy a ProxySQL sandbox alongside the single MySQL sandbox with one backend.

- [ ] **Step 2: Test**

```bash
./dbdeployer deploy single 8.4.4 --sandbox-binary=$HOME/opt/mysql --with-proxysql
~/sandboxes/msb_8_4_4/proxysql/use -e "SELECT * FROM mysql_servers"
./dbdeployer delete all --skip-confirm
```

- [ ] **Step 3: Commit**

```bash
git add cmd/single.go
git commit -m "feat: add --with-proxysql flag for single sandbox deployment"
```

---

### Task 7: Update sandbox deletion to handle ProxySQL

**Files:**
- Modify: `cmd/delete.go` or sandbox deletion logic

Ensure `dbdeployer delete` properly stops and removes ProxySQL sandboxes alongside MySQL ones.

- [ ] **Step 1: Update deletion to check for ProxySQL sub-sandbox**

When deleting a sandbox that has a `proxysql/` subdirectory, run `proxysql/stop` first.

- [ ] **Step 2: Test**

```bash
./dbdeployer deploy replication 8.4.4 --sandbox-binary=$HOME/opt/mysql --with-proxysql
./dbdeployer delete all --skip-confirm
# Verify no stale proxysql processes
ps aux | grep proxysql | grep -v grep
```

- [ ] **Step 3: Commit**

```bash
git add cmd/delete.go
git commit -m "feat: handle ProxySQL cleanup during sandbox deletion"
```

---

### Task 8: Final validation and documentation

- [ ] **Step 1: Run all unit tests**

```bash
go test ./providers/... ./cmd/... ./common/... -timeout 30m
```

- [ ] **Step 2: Full integration test**

```bash
# Standalone ProxySQL
./dbdeployer deploy proxysql --port 16032
./dbdeployer delete all --skip-confirm

# Single MySQL + ProxySQL
./dbdeployer deploy single 8.4.4 --sandbox-binary=$HOME/opt/mysql --with-proxysql
./dbdeployer delete all --skip-confirm

# Replication + ProxySQL
./dbdeployer deploy replication 9.1.0 --sandbox-binary=$HOME/opt/mysql --with-proxysql
~/sandboxes/rsandbox_9_1_0/proxysql/use -e "SELECT * FROM mysql_servers"
~/sandboxes/rsandbox_9_1_0/check_slaves
./dbdeployer delete all --skip-confirm
```

- [ ] **Step 3: Verify `dbdeployer providers` shows both**

```bash
./dbdeployer providers
```
Expected:
```
mysql           (base port: 3306, ports per instance: 3)
proxysql        (base port: 6032, ports per instance: 2)
```

- [ ] **Step 4: Update README with ProxySQL usage examples**

---

## What Phase 2b Does NOT Do (Deferred)

- No Orchestrator provider (separate Phase 2c)
- No tarball management for ProxySQL (no tarballs distributed yet)
- No query rules in generated config (users configure manually)
- No `--with-proxysql` for group replication (can be added incrementally)
- No ProxySQL version detection from system binary (uses user-specified or "system")
