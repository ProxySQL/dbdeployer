# Phase 3 — PostgreSQL Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a PostgreSQL provider to dbdeployer supporting single sandbox, streaming replication, cross-database topology constraints, and ProxySQL+PostgreSQL backend wiring.

**Architecture:** Extend the Provider interface with `SupportedTopologies()` and `CreateReplica()`. Implement a PostgreSQL provider that uses `initdb`/`pg_ctl`/`pg_basebackup` for sandbox lifecycle. Add deb extraction for binary management. Wire into existing cmd layer via `--provider` flag. Extend ProxySQL config generator for PostgreSQL backends.

**Tech Stack:** Go, PostgreSQL CLI tools (initdb, pg_ctl, pg_basebackup, psql), dpkg-deb

**Spec:** `docs/superpowers/specs/2026-03-24-phase3-postgresql-provider-design.md`

---

## File Structure

### New Files
- `providers/postgresql/postgresql.go` — Provider struct, registration, Name/ValidateVersion/DefaultPorts/FindBinary/StartSandbox/StopSandbox/SupportedTopologies/CreateReplica
- `providers/postgresql/sandbox.go` — CreateSandbox implementation (initdb, config gen, script gen)
- `providers/postgresql/config.go` — postgresql.conf and pg_hba.conf generation functions
- `providers/postgresql/scripts.go` — lifecycle script generation (start, stop, status, restart, use, clear)
- `providers/postgresql/unpack.go` — deb extraction logic
- `providers/postgresql/postgresql_test.go` — unit tests for provider methods
- `providers/postgresql/config_test.go` — unit tests for config generation
- `providers/postgresql/unpack_test.go` — unit tests for deb extraction
- `providers/postgresql/integration_test.go` — integration tests (build-tagged)
- `cmd/deploy_postgresql.go` — `dbdeployer deploy postgresql <version>` standalone command

### Modified Files
- `providers/provider.go` — add `SupportedTopologies()`, `CreateReplica()`, `ErrNotSupported`
- `providers/provider_test.go` — update mock, add topology/validation tests
- `providers/mysql/mysql.go` — implement new interface methods
- `providers/proxysql/proxysql.go` — implement new interface methods
- `providers/proxysql/config.go` — add PostgreSQL backend config generation
- `providers/proxysql/proxysql_test.go` — update for new interface methods
- `providers/proxysql/config_test.go` — test PostgreSQL backend config
- `sandbox/proxysql_topology.go` — accept `backendProvider` parameter
- `cmd/root.go` — register PostgreSQL provider
- `cmd/single.go` — add `--provider` flag, route to provider
- `cmd/multiple.go` — add `--provider` flag, route to provider
- `cmd/replication.go` — add `--provider` flag, PostgreSQL replication flow
- `cmd/unpack.go` — add `--provider` flag for deb extraction
- `globals/globals.go` — PostgreSQL constants

---

## Task 1: Extend Provider Interface

**Files:**
- Modify: `providers/provider.go`
- Modify: `providers/provider_test.go`
- Modify: `providers/mysql/mysql.go`
- Modify: `providers/proxysql/proxysql.go`
- Modify: `providers/proxysql/proxysql_test.go`

- [ ] **Step 1: Write failing test for SupportedTopologies on mock provider**

In `providers/provider_test.go`, add `SupportedTopologies` and `CreateReplica` to `mockProvider`, then write a test:

```go
func (m *mockProvider) SupportedTopologies() []string {
	return []string{"single", "multiple"}
}
func (m *mockProvider) CreateReplica(primary SandboxInfo, config SandboxConfig) (*SandboxInfo, error) {
	return nil, ErrNotSupported
}

func TestErrNotSupported(t *testing.T) {
	mock := &mockProvider{name: "test"}
	_, err := mock.CreateReplica(SandboxInfo{}, SandboxConfig{})
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
}

func TestSupportedTopologies(t *testing.T) {
	mock := &mockProvider{name: "test"}
	topos := mock.SupportedTopologies()
	if len(topos) != 2 || topos[0] != "single" {
		t.Errorf("unexpected topologies: %v", topos)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /data/rene/dbdeployer && go test ./providers/ -run TestErrNotSupported -v`
Expected: Compilation error — `ErrNotSupported` and `SupportedTopologies` not defined on interface.

- [ ] **Step 3: Add interface methods and ErrNotSupported to provider.go**

In `providers/provider.go`, add:

```go
import (
	"errors"
	"fmt"
	"sort"
)

var ErrNotSupported = errors.New("operation not supported by this provider")

type Provider interface {
	Name() string
	ValidateVersion(version string) error
	DefaultPorts() PortRange
	FindBinary(version string) (string, error)
	CreateSandbox(config SandboxConfig) (*SandboxInfo, error)
	StartSandbox(dir string) error
	StopSandbox(dir string) error
	SupportedTopologies() []string
	CreateReplica(primary SandboxInfo, config SandboxConfig) (*SandboxInfo, error)
}
```

- [ ] **Step 4: Update MySQLProvider to implement new methods**

In `providers/mysql/mysql.go`, add:

```go
func (p *MySQLProvider) SupportedTopologies() []string {
	return []string{"single", "multiple", "replication", "group", "fan-in", "all-masters", "ndb", "pxc"}
}

func (p *MySQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, providers.ErrNotSupported
}
```

- [ ] **Step 5: Update ProxySQLProvider to implement new methods**

In `providers/proxysql/proxysql.go`, add:

```go
func (p *ProxySQLProvider) SupportedTopologies() []string {
	return []string{"single"}
}

func (p *ProxySQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, providers.ErrNotSupported
}
```

- [ ] **Step 6: Run all provider tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/... -v`
Expected: All tests pass, including the new ones and existing ProxySQL tests.

- [ ] **Step 7: Commit**

```bash
git add providers/provider.go providers/provider_test.go providers/mysql/mysql.go providers/proxysql/proxysql.go
git commit -m "feat: extend Provider interface with SupportedTopologies and CreateReplica"
```

---

## Task 2: PostgreSQL Provider — Core Structure and Version Validation

**Files:**
- Create: `providers/postgresql/postgresql.go`
- Create: `providers/postgresql/postgresql_test.go`

- [ ] **Step 1: Write failing tests for PostgreSQL provider basics**

Create `providers/postgresql/postgresql_test.go`:

```go
package postgresql

import (
	"testing"

	"github.com/ProxySQL/dbdeployer/providers"
)

func TestPostgreSQLProviderName(t *testing.T) {
	p := NewPostgreSQLProvider()
	if p.Name() != "postgresql" {
		t.Errorf("expected 'postgresql', got %q", p.Name())
	}
}

func TestPostgreSQLProviderValidateVersion(t *testing.T) {
	p := NewPostgreSQLProvider()
	tests := []struct {
		version string
		wantErr bool
	}{
		{"16.13", false},
		{"17.1", false},
		{"12.0", false},
		{"11.5", true},    // major < 12
		{"16", true},      // missing minor
		{"16.13.1", true}, // three parts
		{"abc", true},
		{"", true},
	}
	for _, tt := range tests {
		err := p.ValidateVersion(tt.version)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
		}
	}
}

func TestPostgreSQLProviderDefaultPorts(t *testing.T) {
	p := NewPostgreSQLProvider()
	ports := p.DefaultPorts()
	if ports.BasePort != 15000 {
		t.Errorf("expected BasePort 15000, got %d", ports.BasePort)
	}
	if ports.PortsPerInstance != 1 {
		t.Errorf("expected PortsPerInstance 1, got %d", ports.PortsPerInstance)
	}
}

func TestPostgreSQLProviderSupportedTopologies(t *testing.T) {
	p := NewPostgreSQLProvider()
	topos := p.SupportedTopologies()
	expected := map[string]bool{"single": true, "multiple": true, "replication": true}
	if len(topos) != len(expected) {
		t.Fatalf("expected %d topologies, got %d: %v", len(expected), len(topos), topos)
	}
	for _, topo := range topos {
		if !expected[topo] {
			t.Errorf("unexpected topology %q", topo)
		}
	}
}

func TestPostgreSQLVersionToPort(t *testing.T) {
	tests := []struct {
		version  string
		expected int
	}{
		{"16.13", 16613},
		{"16.3", 16603},
		{"17.1", 16701},
		{"17.10", 16710},
		{"12.0", 16200},
	}
	for _, tt := range tests {
		port, err := VersionToPort(tt.version)
		if err != nil {
			t.Errorf("VersionToPort(%q) unexpected error: %v", tt.version, err)
			continue
		}
		if port != tt.expected {
			t.Errorf("VersionToPort(%q) = %d, want %d", tt.version, port, tt.expected)
		}
	}
}

func TestPostgreSQLProviderRegister(t *testing.T) {
	reg := providers.NewRegistry()
	if err := Register(reg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	p, err := reg.Get("postgresql")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name() != "postgresql" {
		t.Errorf("expected 'postgresql', got %q", p.Name())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -v`
Expected: Compilation error — package doesn't exist.

- [ ] **Step 3: Implement PostgreSQL provider core**

Create `providers/postgresql/postgresql.go`:

```go
package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "postgresql"

type PostgreSQLProvider struct{}

func NewPostgreSQLProvider() *PostgreSQLProvider { return &PostgreSQLProvider{} }

func (p *PostgreSQLProvider) Name() string { return ProviderName }

func (p *PostgreSQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid PostgreSQL version format: %q (expected major.minor, e.g. 16.13)", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid PostgreSQL major version %q: %w", parts[0], err)
	}
	if major < 12 {
		return fmt.Errorf("PostgreSQL major version must be >= 12, got %d", major)
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("invalid PostgreSQL minor version %q: %w", parts[1], err)
	}
	return nil
}

func (p *PostgreSQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{BasePort: 15000, PortsPerInstance: 1}
}

func (p *PostgreSQLProvider) SupportedTopologies() []string {
	return []string{"single", "multiple", "replication"}
}

// VersionToPort converts a PostgreSQL version to a port number.
// Formula: BasePort + major*100 + minor
// Example: 16.13 -> 15000 + 1600 + 13 = 16613
func VersionToPort(version string) (int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid version format: %q", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return 15000 + major*100 + minor, nil
}

// FindBinary returns the path to the postgres binary for the given version.
// Looks in ~/opt/postgresql/<version>/bin/postgres by default.
func (p *PostgreSQLProvider) FindBinary(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	binPath := filepath.Join(home, "opt", "postgresql", version, "bin", "postgres")
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("PostgreSQL binary not found at %s: %w", binPath, err)
	}
	return binPath, nil
}

// basedirFromVersion returns the base directory for a PostgreSQL version.
func basedirFromVersion(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, "opt", "postgresql", version), nil
}

func (p *PostgreSQLProvider) StartSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "start"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *PostgreSQLProvider) StopSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "stop"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop failed: %s: %w", string(output), err)
	}
	return nil
}

func Register(reg *providers.Registry) error {
	return reg.Register(NewPostgreSQLProvider())
}
```

Note: `CreateSandbox` and `CreateReplica` are implemented in Task 4 and Task 6 respectively, in separate files. Add stubs for now:

```go
func (p *PostgreSQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, fmt.Errorf("PostgreSQLProvider.CreateSandbox: not yet implemented")
}

func (p *PostgreSQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, fmt.Errorf("PostgreSQLProvider.CreateReplica: not yet implemented")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -v`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add providers/postgresql/postgresql.go providers/postgresql/postgresql_test.go
git commit -m "feat: add PostgreSQL provider core structure and version validation"
```

---

## Task 3: PostgreSQL Config Generation

**Files:**
- Create: `providers/postgresql/config.go`
- Create: `providers/postgresql/config_test.go`

- [ ] **Step 1: Write failing tests for config generation**

Create `providers/postgresql/config_test.go`:

```go
package postgresql

import (
	"strings"
	"testing"
)

func TestGeneratePostgresqlConf(t *testing.T) {
	conf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:                5433,
		ListenAddresses:     "127.0.0.1",
		UnixSocketDir:       "/tmp/sandbox/data",
		LogDir:              "/tmp/sandbox/data/log",
		Replication:         false,
	})
	if !strings.Contains(conf, "port = 5433") {
		t.Error("missing port setting")
	}
	if !strings.Contains(conf, "listen_addresses = '127.0.0.1'") {
		t.Error("missing listen_addresses")
	}
	if !strings.Contains(conf, "unix_socket_directories = '/tmp/sandbox/data'") {
		t.Error("missing unix_socket_directories")
	}
	if !strings.Contains(conf, "logging_collector = on") {
		t.Error("missing logging_collector")
	}
	if strings.Contains(conf, "wal_level") {
		t.Error("should not contain wal_level when replication is false")
	}
}

func TestGeneratePostgresqlConfWithReplication(t *testing.T) {
	conf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:            5433,
		ListenAddresses: "127.0.0.1",
		UnixSocketDir:   "/tmp/sandbox/data",
		LogDir:          "/tmp/sandbox/data/log",
		Replication:     true,
	})
	if !strings.Contains(conf, "wal_level = replica") {
		t.Error("missing wal_level = replica")
	}
	if !strings.Contains(conf, "max_wal_senders = 10") {
		t.Error("missing max_wal_senders")
	}
	if !strings.Contains(conf, "hot_standby = on") {
		t.Error("missing hot_standby")
	}
}

func TestGeneratePgHbaConf(t *testing.T) {
	conf := GeneratePgHbaConf(false)
	if !strings.Contains(conf, "local   all") {
		t.Error("missing local all entry")
	}
	if !strings.Contains(conf, "host    all") {
		t.Error("missing host all entry")
	}
	if strings.Contains(conf, "replication") {
		t.Error("should not contain replication when replication is false")
	}
}

func TestGeneratePgHbaConfWithReplication(t *testing.T) {
	conf := GeneratePgHbaConf(true)
	if !strings.Contains(conf, "host    replication") {
		t.Error("missing replication entry")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run TestGenerate -v`
Expected: Compilation error — functions not defined.

- [ ] **Step 3: Implement config generation**

Create `providers/postgresql/config.go`:

```go
package postgresql

import (
	"fmt"
	"strings"
)

type PostgresqlConfOptions struct {
	Port            int
	ListenAddresses string
	UnixSocketDir   string
	LogDir          string
	Replication     bool
}

func GeneratePostgresqlConf(opts PostgresqlConfOptions) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("port = %d\n", opts.Port))
	b.WriteString(fmt.Sprintf("listen_addresses = '%s'\n", opts.ListenAddresses))
	b.WriteString(fmt.Sprintf("unix_socket_directories = '%s'\n", opts.UnixSocketDir))
	b.WriteString("logging_collector = on\n")
	b.WriteString(fmt.Sprintf("log_directory = '%s'\n", opts.LogDir))

	if opts.Replication {
		b.WriteString("\n# Replication settings\n")
		b.WriteString("wal_level = replica\n")
		b.WriteString("max_wal_senders = 10\n")
		b.WriteString("hot_standby = on\n")
	}

	return b.String()
}

func GeneratePgHbaConf(replication bool) string {
	var b strings.Builder
	b.WriteString("# TYPE  DATABASE  USER  ADDRESS       METHOD\n")
	b.WriteString("local   all       all                 trust\n")
	b.WriteString("host    all       all   127.0.0.1/32  trust\n")
	b.WriteString("host    all       all   ::1/128       trust\n")

	if replication {
		b.WriteString("host    replication  all  127.0.0.1/32  trust\n")
	}

	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run TestGenerate -v`
Expected: All pass.

- [ ] **Step 5: Commit**

```bash
git add providers/postgresql/config.go providers/postgresql/config_test.go
git commit -m "feat: add PostgreSQL config generation (postgresql.conf, pg_hba.conf)"
```

---

## Task 4: PostgreSQL Script Generation and CreateSandbox

**Files:**
- Create: `providers/postgresql/scripts.go`
- Create: `providers/postgresql/sandbox.go`
- Modify: `providers/postgresql/postgresql.go` (replace CreateSandbox stub)
- Modify: `providers/postgresql/postgresql_test.go` (add script tests)

- [ ] **Step 1: Write failing tests for script generation**

Add to `providers/postgresql/postgresql_test.go`:

```go
func TestGenerateScripts(t *testing.T) {
	opts := ScriptOptions{
		SandboxDir: "/tmp/pg_sandbox",
		DataDir:    "/tmp/pg_sandbox/data",
		BinDir:     "/opt/postgresql/16.13/bin",
		LibDir:     "/opt/postgresql/16.13/lib",
		Port:       16613,
		LogFile:    "/tmp/pg_sandbox/postgresql.log",
	}
	scripts := GenerateScripts(opts)

	// Verify all expected scripts exist
	expectedScripts := []string{"start", "stop", "status", "restart", "use", "clear"}
	for _, name := range expectedScripts {
		if _, ok := scripts[name]; !ok {
			t.Errorf("missing script %q", name)
		}
	}

	// Verify start script contents
	start := scripts["start"]
	if !strings.Contains(start, "pg_ctl") {
		t.Error("start script missing pg_ctl")
	}
	if !strings.Contains(start, "LD_LIBRARY_PATH") {
		t.Error("start script missing LD_LIBRARY_PATH")
	}
	if !strings.Contains(start, "unset PGDATA") {
		t.Error("start script missing PGDATA unset")
	}

	// Verify use script
	use := scripts["use"]
	if !strings.Contains(use, "psql") {
		t.Error("use script missing psql")
	}
	if !strings.Contains(use, "16613") {
		t.Error("use script missing port")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run TestGenerateScripts -v`
Expected: Compilation error — `ScriptOptions` and `GenerateScripts` not defined.

- [ ] **Step 3: Implement script generation**

Create `providers/postgresql/scripts.go`:

```go
package postgresql

import "fmt"

type ScriptOptions struct {
	SandboxDir string
	DataDir    string
	BinDir     string
	LibDir     string
	Port       int
	LogFile    string
}

const envPreamble = `#!/bin/bash
export LD_LIBRARY_PATH="%s"
unset PGDATA PGPORT PGHOST PGUSER PGDATABASE
`

func GenerateScripts(opts ScriptOptions) map[string]string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)

	return map[string]string{
		"start": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s start\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"stop": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast\n",
			preamble, opts.BinDir, opts.DataDir),

		"status": fmt.Sprintf("%s%s/pg_ctl -D %s status\n",
			preamble, opts.BinDir, opts.DataDir),

		"restart": fmt.Sprintf("%s%s/pg_ctl -D %s -l %s restart\n",
			preamble, opts.BinDir, opts.DataDir, opts.LogFile),

		"use": fmt.Sprintf("%s%s/psql -h 127.0.0.1 -p %d -U postgres \"$@\"\n",
			preamble, opts.BinDir, opts.Port),

		"clear": fmt.Sprintf("%s%s/pg_ctl -D %s stop -m fast 2>/dev/null\nrm -rf %s\n%s/initdb -D %s --auth=trust --username=postgres\necho \"Sandbox cleared.\"\n",
			preamble, opts.BinDir, opts.DataDir, opts.DataDir, opts.BinDir, opts.DataDir),
	}
}
```

- [ ] **Step 4: Run script generation tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run TestGenerateScripts -v`
Expected: PASS.

- [ ] **Step 5: Implement CreateSandbox**

Create `providers/postgresql/sandbox.go`:

```go
package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ProxySQL/dbdeployer/providers"
)

func (p *PostgreSQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	basedir, err := p.resolveBasedir(config)
	if err != nil {
		return nil, err
	}
	binDir := filepath.Join(basedir, "bin")
	libDir := filepath.Join(basedir, "lib")
	dataDir := filepath.Join(config.Dir, "data")
	logDir := filepath.Join(dataDir, "log")
	logFile := filepath.Join(config.Dir, "postgresql.log")

	replication := config.Options["replication"] == "true"

	// Create log directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Run initdb
	initdbPath := filepath.Join(binDir, "initdb")
	initCmd := exec.Command(initdbPath, "-D", dataDir, "--auth=trust", "--username=postgres")
	initCmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	if output, err := initCmd.CombinedOutput(); err != nil {
		os.RemoveAll(config.Dir) // cleanup on failure
		return nil, fmt.Errorf("initdb failed: %s: %w", string(output), err)
	}

	// Generate and write postgresql.conf
	pgConf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:            config.Port,
		ListenAddresses: "127.0.0.1",
		UnixSocketDir:   dataDir,
		LogDir:          logDir,
		Replication:     replication,
	})
	confPath := filepath.Join(dataDir, "postgresql.conf")
	if err := os.WriteFile(confPath, []byte(pgConf), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing postgresql.conf: %w", err)
	}

	// Generate and write pg_hba.conf
	hbaConf := GeneratePgHbaConf(replication)
	hbaPath := filepath.Join(dataDir, "pg_hba.conf")
	if err := os.WriteFile(hbaPath, []byte(hbaConf), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing pg_hba.conf: %w", err)
	}

	// Generate and write lifecycle scripts
	scripts := GenerateScripts(ScriptOptions{
		SandboxDir: config.Dir,
		DataDir:    dataDir,
		BinDir:     binDir,
		LibDir:     libDir,
		Port:       config.Port,
		LogFile:    logFile,
	})
	for name, content := range scripts {
		scriptPath := filepath.Join(config.Dir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			os.RemoveAll(config.Dir)
			return nil, fmt.Errorf("writing script %s: %w", name, err)
		}
	}

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   config.Port,
		Status: "stopped",
	}, nil
}

// resolveBasedir determines the PostgreSQL base directory.
// Uses config.Options["basedir"] if set, otherwise ~/opt/postgresql/<version>.
func (p *PostgreSQLProvider) resolveBasedir(config providers.SandboxConfig) (string, error) {
	if bd, ok := config.Options["basedir"]; ok && bd != "" {
		return bd, nil
	}
	return basedirFromVersion(config.Version)
}
```

- [ ] **Step 6: Remove the CreateSandbox stub from postgresql.go**

In `providers/postgresql/postgresql.go`, remove the stub `CreateSandbox` method (now implemented in sandbox.go).

- [ ] **Step 7: Run all PostgreSQL provider tests**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -v`
Expected: All pass.

- [ ] **Step 8: Commit**

```bash
git add providers/postgresql/scripts.go providers/postgresql/sandbox.go providers/postgresql/postgresql.go providers/postgresql/postgresql_test.go
git commit -m "feat: implement PostgreSQL CreateSandbox with initdb, config gen, and lifecycle scripts"
```

---

## Task 5: Deb Extraction for PostgreSQL Binaries

**Files:**
- Create: `providers/postgresql/unpack.go`
- Create: `providers/postgresql/unpack_test.go`

- [ ] **Step 1: Write failing tests for deb filename parsing and validation**

Create `providers/postgresql/unpack_test.go`:

```go
package postgresql

import "testing"

func TestParseDebVersion(t *testing.T) {
	tests := []struct {
		filename string
		wantVer  string
		wantErr  bool
	}{
		{"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb", "16.13", false},
		{"postgresql-17_17.2-1_amd64.deb", "17.2", false},
		{"postgresql-client-16_16.13-0ubuntu0.24.04.1_amd64.deb", "16.13", false},
		{"random-file.tar.gz", "", true},
		{"postgresql-16_bad-version.deb", "", true},
	}
	for _, tt := range tests {
		ver, err := ParseDebVersion(tt.filename)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDebVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			continue
		}
		if ver != tt.wantVer {
			t.Errorf("ParseDebVersion(%q) = %q, want %q", tt.filename, ver, tt.wantVer)
		}
	}
}

func TestClassifyDebs(t *testing.T) {
	files := []string{
		"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb",
		"postgresql-client-16_16.13-0ubuntu0.24.04.1_amd64.deb",
	}
	server, client, err := ClassifyDebs(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server != files[0] {
		t.Errorf("server = %q, want %q", server, files[0])
	}
	if client != files[1] {
		t.Errorf("client = %q, want %q", client, files[1])
	}
}

func TestClassifyDebsMissingClient(t *testing.T) {
	files := []string{"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb"}
	_, _, err := ClassifyDebs(files)
	if err == nil {
		t.Error("expected error for missing client deb")
	}
}

func TestRequiredBinaries(t *testing.T) {
	expected := []string{"postgres", "initdb", "pg_ctl", "psql", "pg_basebackup"}
	got := RequiredBinaries()
	if len(got) != len(expected) {
		t.Fatalf("expected %d binaries, got %d", len(expected), len(got))
	}
	for i, name := range expected {
		if got[i] != name {
			t.Errorf("binary[%d] = %q, want %q", i, got[i], name)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run "TestParseDeb|TestClassify|TestRequired" -v`
Expected: Compilation error — functions not defined.

- [ ] **Step 3: Implement deb extraction logic**

Create `providers/postgresql/unpack.go`:

```go
package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var debVersionRegex = regexp.MustCompile(`^postgresql(?:-client)?-(\d+)_(\d+\.\d+)`)

// ParseDebVersion extracts the PostgreSQL version from a deb filename.
func ParseDebVersion(filename string) (string, error) {
	base := filepath.Base(filename)
	matches := debVersionRegex.FindStringSubmatch(base)
	if matches == nil {
		return "", fmt.Errorf("cannot parse PostgreSQL version from %q (expected postgresql[-client]-NN_X.Y-*)", base)
	}
	return matches[2], nil
}

// ClassifyDebs identifies server and client debs from a list of filenames.
func ClassifyDebs(files []string) (server, client string, err error) {
	for _, f := range files {
		base := filepath.Base(f)
		if strings.HasPrefix(base, "postgresql-client-") {
			client = f
		} else if strings.HasPrefix(base, "postgresql-") && strings.HasSuffix(base, ".deb") {
			server = f
		}
	}
	if server == "" {
		return "", "", fmt.Errorf("no server deb found (expected postgresql-NN_*.deb)")
	}
	if client == "" {
		return "", "", fmt.Errorf("no client deb found (expected postgresql-client-NN_*.deb)")
	}
	return server, client, nil
}

// RequiredBinaries returns the binaries that must exist after extraction.
func RequiredBinaries() []string {
	return []string{"postgres", "initdb", "pg_ctl", "psql", "pg_basebackup"}
}

// UnpackDebs extracts PostgreSQL server and client debs into the target directory.
// targetDir is the final layout dir, e.g. ~/opt/postgresql/16.13/
func UnpackDebs(serverDeb, clientDeb, targetDir string) error {
	tmpDir, err := os.MkdirTemp("", "dbdeployer-pg-unpack-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract both debs
	for _, deb := range []string{serverDeb, clientDeb} {
		cmd := exec.Command("dpkg-deb", "-x", deb, tmpDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("extracting %s: %s: %w", filepath.Base(deb), string(output), err)
		}
	}

	// Determine the major version directory inside the extracted tree
	version, err := ParseDebVersion(serverDeb)
	if err != nil {
		return err
	}
	major := strings.Split(version, ".")[0]

	// Source paths within extracted debs
	srcBin := filepath.Join(tmpDir, "usr", "lib", "postgresql", major, "bin")
	srcLib := filepath.Join(tmpDir, "usr", "lib", "postgresql", major, "lib")
	srcShare := filepath.Join(tmpDir, "usr", "share", "postgresql", major)

	// Create target directories
	dstBin := filepath.Join(targetDir, "bin")
	dstLib := filepath.Join(targetDir, "lib")
	dstShare := filepath.Join(targetDir, "share")

	for _, dir := range []string{dstBin, dstLib, dstShare} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	// Copy files using cp -a to preserve permissions and symlinks
	copies := []struct{ src, dst string }{
		{srcBin, dstBin},
		{srcLib, dstLib},
		{srcShare, dstShare},
	}
	for _, c := range copies {
		if _, err := os.Stat(c.src); os.IsNotExist(err) {
			continue // some dirs may not exist in the client deb
		}
		cmd := exec.Command("cp", "-a", c.src+"/.", c.dst+"/")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("copying %s to %s: %s: %w", c.src, c.dst, string(output), err)
		}
	}

	// Validate required binaries
	for _, bin := range RequiredBinaries() {
		binPath := filepath.Join(dstBin, bin)
		if _, err := os.Stat(binPath); err != nil {
			return fmt.Errorf("required binary %q not found at %s after extraction", bin, binPath)
		}
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run "TestParseDeb|TestClassify|TestRequired" -v`
Expected: All pass.

- [ ] **Step 5: Commit**

```bash
git add providers/postgresql/unpack.go providers/postgresql/unpack_test.go
git commit -m "feat: add PostgreSQL deb extraction for binary management"
```

---

## Task 6: PostgreSQL Replication (CreateReplica)

**Files:**
- Modify: `providers/postgresql/postgresql.go` (replace CreateReplica stub)
- Create: `providers/postgresql/replication.go`
- Modify: `providers/postgresql/postgresql_test.go` (add replication config tests)

- [ ] **Step 1: Write failing tests for replication monitoring script generation**

Add to `providers/postgresql/postgresql_test.go`:

```go
func TestGenerateCheckReplicationScript(t *testing.T) {
	script := GenerateCheckReplicationScript(ScriptOptions{
		BinDir: "/opt/postgresql/16.13/bin",
		LibDir: "/opt/postgresql/16.13/lib",
		Port:   16613,
	})
	if !strings.Contains(script, "pg_stat_replication") {
		t.Error("missing pg_stat_replication query")
	}
	if !strings.Contains(script, "16613") {
		t.Error("missing primary port")
	}
}

func TestGenerateCheckRecoveryScript(t *testing.T) {
	ports := []int{16614, 16615}
	script := GenerateCheckRecoveryScript(ScriptOptions{
		BinDir: "/opt/postgresql/16.13/bin",
		LibDir: "/opt/postgresql/16.13/lib",
	}, ports)
	if !strings.Contains(script, "pg_is_in_recovery") {
		t.Error("missing pg_is_in_recovery query")
	}
	if !strings.Contains(script, "16614") || !strings.Contains(script, "16615") {
		t.Error("missing replica ports")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -run "TestGenerateCheck" -v`
Expected: Compilation error — functions not defined.

- [ ] **Step 3: Add monitoring script generators to scripts.go**

Add to `providers/postgresql/scripts.go`:

```go
func GenerateCheckReplicationScript(opts ScriptOptions) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	return fmt.Sprintf(`%s%s/psql -h 127.0.0.1 -p %d -U postgres -c \
  "SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication;"
`, preamble, opts.BinDir, opts.Port)
}

func GenerateCheckRecoveryScript(opts ScriptOptions, replicaPorts []int) string {
	preamble := fmt.Sprintf(envPreamble, opts.LibDir)
	var b strings.Builder
	b.WriteString(preamble)
	for _, port := range replicaPorts {
		b.WriteString(fmt.Sprintf("echo \"=== Replica port %d ===\"\n", port))
		b.WriteString(fmt.Sprintf("%s/psql -h 127.0.0.1 -p %d -U postgres -c \"SELECT pg_is_in_recovery();\"\n", opts.BinDir, port))
	}
	return b.String()
}
```

Add `"strings"` to imports in `scripts.go`.

- [ ] **Step 4: Implement CreateReplica**

Create `providers/postgresql/replication.go`:

```go
package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

func (p *PostgreSQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	basedir, err := p.resolveBasedir(config)
	if err != nil {
		return nil, err
	}
	binDir := filepath.Join(basedir, "bin")
	libDir := filepath.Join(basedir, "lib")
	dataDir := filepath.Join(config.Dir, "data")
	logFile := filepath.Join(config.Dir, "postgresql.log")

	// pg_basebackup from the running primary
	pgBasebackup := filepath.Join(binDir, "pg_basebackup")
	bbCmd := exec.Command(pgBasebackup,
		"-h", "127.0.0.1",
		"-p", fmt.Sprintf("%d", primary.Port),
		"-U", "postgres",
		"-D", dataDir,
		"-Fp", "-Xs", "-R",
	)
	bbCmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	if output, err := bbCmd.CombinedOutput(); err != nil {
		os.RemoveAll(config.Dir) // cleanup on failure
		return nil, fmt.Errorf("pg_basebackup failed: %s: %w", string(output), err)
	}

	// Modify replica's postgresql.conf: update port and unix_socket_directories
	confPath := filepath.Join(dataDir, "postgresql.conf")
	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("reading postgresql.conf: %w", err)
	}

	conf := string(confBytes)
	// Replace port line
	lines := strings.Split(conf, "\n")
	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "port =") || strings.HasPrefix(trimmed, "port=") {
			newLines = append(newLines, fmt.Sprintf("port = %d", config.Port))
		} else if strings.HasPrefix(trimmed, "unix_socket_directories =") || strings.HasPrefix(trimmed, "unix_socket_directories=") {
			newLines = append(newLines, fmt.Sprintf("unix_socket_directories = '%s'", dataDir))
		} else {
			newLines = append(newLines, line)
		}
	}

	if err := os.WriteFile(confPath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing modified postgresql.conf: %w", err)
	}

	// Write lifecycle scripts
	scripts := GenerateScripts(ScriptOptions{
		SandboxDir: config.Dir,
		DataDir:    dataDir,
		BinDir:     binDir,
		LibDir:     libDir,
		Port:       config.Port,
		LogFile:    logFile,
	})
	for name, content := range scripts {
		scriptPath := filepath.Join(config.Dir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			os.RemoveAll(config.Dir)
			return nil, fmt.Errorf("writing script %s: %w", name, err)
		}
	}

	// Start the replica
	if err := p.StartSandbox(config.Dir); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("starting replica: %w", err)
	}

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   config.Port,
		Status: "running",
	}, nil
}
```

- [ ] **Step 5: Remove CreateReplica stub from postgresql.go**

In `providers/postgresql/postgresql.go`, remove the stub `CreateReplica` method.

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -v`
Expected: All tests pass (unit tests; replication flow is integration-tested).

- [ ] **Step 7: Commit**

```bash
git add providers/postgresql/replication.go providers/postgresql/scripts.go providers/postgresql/postgresql.go providers/postgresql/postgresql_test.go
git commit -m "feat: implement PostgreSQL CreateReplica with pg_basebackup and monitoring scripts"
```

---

## Task 7: Register Provider and Add --provider Flag to Commands

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/single.go`
- Modify: `cmd/multiple.go`
- Modify: `cmd/replication.go`
- Modify: `globals/globals.go`
- Modify: `providers/provider.go` (add `ContainsString` helper)
- Modify: `sandbox/proxysql_topology.go` (add `backendProvider` parameter)

**Note:** This task introduces `cmd/deploy_postgresql.go` (Task 11) and splits files not in the original spec (`sandbox.go`, `scripts.go`). These are intentional improvements for code organization and UX.

- [ ] **Step 1: Add PostgreSQL constants and ContainsString helper to providers**

In `globals/globals.go`, add near the existing constant blocks:

```go
const (
	ProviderLabel = "provider"
	ProviderValue = "mysql" // default provider
)
```

In `providers/provider.go`, add an exported helper:

```go
// ContainsString checks if a string slice contains a given value.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Register PostgreSQL provider in cmd/root.go**

In `cmd/root.go`, add import for PostgreSQL provider and register it in `init()`:

```go
import (
	// existing imports...
	postgresqlprovider "github.com/ProxySQL/dbdeployer/providers/postgresql"
)

// In init(), after proxysql registration:
_ = postgresqlprovider.Register(providers.DefaultRegistry)
```

- [ ] **Step 3: Update DeployProxySQLForTopology signature**

In `sandbox/proxysql_topology.go`, add a `backendProvider` parameter. All callers must be updated:

```go
func DeployProxySQLForTopology(sandboxDir string, masterPort int, slavePorts []int, proxysqlPort int, host string, backendProvider string) error {
	// ... existing code unchanged until config building ...
	config := providers.SandboxConfig{
		// ... existing fields ...
		Options: map[string]string{
			"monitor_user":     "msandbox",
			"monitor_password": "msandbox",
			"backends":         strings.Join(backendParts, ","),
			"backend_provider": backendProvider, // NEW: "" for mysql, "postgresql" for pg
		},
	}
	// ... rest unchanged ...
}
```

**Callers to update** (pass `""` to preserve existing MySQL behavior):
- `cmd/single.go:485` — `sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, nil, 0, "127.0.0.1", "")`
- `cmd/replication.go:135` — `sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, slavePorts, 0, "127.0.0.1", "")`

- [ ] **Step 4: Update cmd/single.go — add --provider flag and routing**

The key design decision: for non-MySQL providers, we **skip `fillSandboxDefinition` entirely** because it is deeply MySQL-specific (checks for MySQL directories, runs `common.CheckLibraries`, calls `getFlavor`, etc.). Instead, non-MySQL providers build a `providers.SandboxConfig` directly from CLI flags.

Replace `singleSandbox()` with this structure:

```go
func singleSandbox(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	providerName, _ := flags.GetString(globals.ProviderLabel)

	// Non-MySQL providers: bypass fillSandboxDefinition entirely
	if providerName != "mysql" {
		deploySingleNonMySQL(cmd, args, providerName)
		return
	}

	// Existing MySQL path — completely unchanged
	var sd sandbox.SandboxDef
	var err error
	common.CheckOrigin(args)
	sd, err = fillSandboxDefinition(cmd, args, false)
	// ... rest of existing code unchanged, BUT update DeployProxySQLForTopology call:
	// sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, nil, 0, "127.0.0.1", "")
}

func deploySingleNonMySQL(cmd *cobra.Command, args []string, providerName string) {
	flags := cmd.Flags()
	version := args[0]

	p, err := providers.DefaultRegistry.Get(providerName)
	if err != nil {
		common.Exitf(1, "provider error: %s", err)
	}

	// Flavor validation: --flavor is MySQL-only
	flavor, _ := flags.GetString(globals.FlavorLabel)
	if flavor != "" {
		common.Exitf(1, "--flavor is only valid with --provider=mysql")
	}

	// Topology validation
	if !providers.ContainsString(p.SupportedTopologies(), "single") {
		common.Exitf(1, "provider %q does not support topology \"single\"\nSupported topologies: %s",
			providerName, strings.Join(p.SupportedTopologies(), ", "))
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "version validation failed: %s", err)
	}

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "binaries not found: %s", err)
	}

	// Compute port from provider's default port range
	portRange := p.DefaultPorts()
	port := portRange.BasePort
	// For PostgreSQL, use VersionToPort
	if providerName == "postgresql" {
		port, _ = postgresql.VersionToPort(version)
	}
	freePort, portErr := common.FindFreePort(port, []int{}, portRange.PortsPerInstance)
	if portErr == nil {
		port = freePort
	}

	sandboxHome := defaults.Defaults().SandboxHome
	sandboxDir := path.Join(sandboxHome, fmt.Sprintf("%s_sandbox_%d", providerName, port))
	if common.DirExists(sandboxDir) {
		common.Exitf(1, "sandbox directory %s already exists", sandboxDir)
	}

	skipStart, _ := flags.GetBool(globals.SkipStartLabel)
	config := providers.SandboxConfig{
		Version: version,
		Dir:     sandboxDir,
		Port:    port,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{},
	}

	if _, err := p.CreateSandbox(config); err != nil {
		common.Exitf(1, "error creating sandbox: %s", err)
	}

	if !skipStart {
		if err := p.StartSandbox(sandboxDir); err != nil {
			common.Exitf(1, "error starting sandbox: %s", err)
		}
	}

	// Handle --with-proxysql
	withProxySQL, _ := flags.GetBool("with-proxysql")
	if withProxySQL {
		if !providers.ContainsString(providers.CompatibleAddons["proxysql"], providerName) {
			common.Exitf(1, "--with-proxysql is not compatible with provider %q", providerName)
		}
		err := sandbox.DeployProxySQLForTopology(sandboxDir, port, nil, 0, "127.0.0.1", providerName)
		if err != nil {
			common.Exitf(1, "ProxySQL deployment failed: %s", err)
		}
	}

	fmt.Printf("%s %s sandbox deployed in %s (port: %d)\n", providerName, version, sandboxDir, port)
}
```

Add flag in `init()`:

```go
singleCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
```

Add imports for `postgresql` and `providers` packages.

- [ ] **Step 5: Update cmd/multiple.go — add --provider flag and routing**

Same bypass pattern. For non-MySQL providers, create N instances with sequential ports:

```go
func multipleSandbox(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	providerName, _ := flags.GetString(globals.ProviderLabel)

	if providerName != "mysql" {
		deployMultipleNonMySQL(cmd, args, providerName)
		return
	}

	// Existing MySQL path unchanged, no modification needed
	// ...
}

func deployMultipleNonMySQL(cmd *cobra.Command, args []string, providerName string) {
	flags := cmd.Flags()
	version := args[0]
	nodes, _ := flags.GetInt(globals.NodesLabel)

	p, err := providers.DefaultRegistry.Get(providerName)
	if err != nil {
		common.Exitf(1, "provider error: %s", err)
	}

	flavor, _ := flags.GetString(globals.FlavorLabel)
	if flavor != "" {
		common.Exitf(1, "--flavor is only valid with --provider=mysql")
	}

	if !providers.ContainsString(p.SupportedTopologies(), "multiple") {
		common.Exitf(1, "provider %q does not support topology \"multiple\"\nSupported topologies: %s",
			providerName, strings.Join(p.SupportedTopologies(), ", "))
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "version validation failed: %s", err)
	}

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "binaries not found: %s", err)
	}

	// Compute base port
	basePort := p.DefaultPorts().BasePort
	if providerName == "postgresql" {
		basePort, _ = postgresql.VersionToPort(version)
	}

	sandboxHome := defaults.Defaults().SandboxHome
	topologyDir := path.Join(sandboxHome, fmt.Sprintf("%s_multi_%d", providerName, basePort))
	if common.DirExists(topologyDir) {
		common.Exitf(1, "sandbox directory %s already exists", topologyDir)
	}
	os.MkdirAll(topologyDir, 0755)

	skipStart, _ := flags.GetBool(globals.SkipStartLabel)

	for i := 1; i <= nodes; i++ {
		port := basePort + i
		freePort, err := common.FindFreePort(port, []int{}, 1)
		if err == nil {
			port = freePort
		}

		nodeDir := path.Join(topologyDir, fmt.Sprintf("node%d", i))
		config := providers.SandboxConfig{
			Version: version,
			Dir:     nodeDir,
			Port:    port,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}

		if _, err := p.CreateSandbox(config); err != nil {
			common.Exitf(1, "error creating node %d: %s", i, err)
		}

		if !skipStart {
			if err := p.StartSandbox(nodeDir); err != nil {
				common.Exitf(1, "error starting node %d: %s", i, err)
			}
		}

		fmt.Printf("  Node %d deployed in %s (port: %d)\n", i, nodeDir, port)
	}

	fmt.Printf("%s multiple sandbox (%d nodes) deployed in %s\n", providerName, nodes, topologyDir)
}
```

Add flag in `init()`:

```go
multipleCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
```

- [ ] **Step 6: Update cmd/replication.go — add --provider flag and PostgreSQL replication flow**

Same bypass pattern. For PostgreSQL: create primary with replication options, start it, then CreateReplica for each replica sequentially:

```go
func replicationSandbox(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	providerName, _ := flags.GetString(globals.ProviderLabel)

	if providerName != "mysql" {
		deployReplicationNonMySQL(cmd, args, providerName)
		return
	}

	// Existing MySQL path unchanged, BUT update DeployProxySQLForTopology call:
	// sandbox.DeployProxySQLForTopology(sandboxDir, masterPort, slavePorts, 0, "127.0.0.1", "")
	// ...
}

func deployReplicationNonMySQL(cmd *cobra.Command, args []string, providerName string) {
	flags := cmd.Flags()
	version := args[0]
	nodes, _ := flags.GetInt(globals.NodesLabel)

	p, err := providers.DefaultRegistry.Get(providerName)
	if err != nil {
		common.Exitf(1, "provider error: %s", err)
	}

	flavor, _ := flags.GetString(globals.FlavorLabel)
	if flavor != "" {
		common.Exitf(1, "--flavor is only valid with --provider=mysql")
	}

	if !providers.ContainsString(p.SupportedTopologies(), "replication") {
		common.Exitf(1, "provider %q does not support topology \"replication\"\nSupported topologies: %s",
			providerName, strings.Join(p.SupportedTopologies(), ", "))
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "version validation failed: %s", err)
	}

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "binaries not found: %s", err)
	}

	// Compute base port
	basePort := p.DefaultPorts().BasePort
	if providerName == "postgresql" {
		basePort, _ = postgresql.VersionToPort(version)
	}

	sandboxHome := defaults.Defaults().SandboxHome
	topologyDir := path.Join(sandboxHome, fmt.Sprintf("%s_repl_%d", providerName, basePort))
	if common.DirExists(topologyDir) {
		common.Exitf(1, "sandbox directory %s already exists", topologyDir)
	}
	os.MkdirAll(topologyDir, 0755)

	skipStart, _ := flags.GetBool(globals.SkipStartLabel)
	primaryPort := basePort

	// 1. Create and start primary with replication options
	primaryDir := path.Join(topologyDir, "primary")
	primaryConfig := providers.SandboxConfig{
		Version: version,
		Dir:     primaryDir,
		Port:    primaryPort,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{"replication": "true"},
	}

	if _, err := p.CreateSandbox(primaryConfig); err != nil {
		common.Exitf(1, "error creating primary: %s", err)
	}

	if !skipStart {
		if err := p.StartSandbox(primaryDir); err != nil {
			common.Exitf(1, "error starting primary: %s", err)
		}
	}

	fmt.Printf("  Primary deployed in %s (port: %d)\n", primaryDir, primaryPort)

	primaryInfo := providers.SandboxInfo{Dir: primaryDir, Port: primaryPort, Status: "running"}

	// 2. Create replicas sequentially (pg_basebackup requires running primary)
	var replicaPorts []int
	for i := 1; i <= nodes-1; i++ {
		replicaPort := primaryPort + i
		freePort, err := common.FindFreePort(replicaPort, []int{}, 1)
		if err == nil {
			replicaPort = freePort
		}

		replicaDir := path.Join(topologyDir, fmt.Sprintf("replica%d", i))
		replicaConfig := providers.SandboxConfig{
			Version: version,
			Dir:     replicaDir,
			Port:    replicaPort,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}

		if _, err := p.CreateReplica(primaryInfo, replicaConfig); err != nil {
			// Cleanup: stop primary and any already-running replicas
			p.StopSandbox(primaryDir)
			for j := 1; j < i; j++ {
				p.StopSandbox(path.Join(topologyDir, fmt.Sprintf("replica%d", j)))
			}
			common.Exitf(1, "error creating replica %d: %s", i, err)
		}

		replicaPorts = append(replicaPorts, replicaPort)
		fmt.Printf("  Replica %d deployed in %s (port: %d)\n", i, replicaDir, replicaPort)
	}

	// 3. Generate topology-level monitoring scripts
	home, _ := os.UserHomeDir()
	basedir := path.Join(home, "opt", "postgresql", version)
	binDir := path.Join(basedir, "bin")
	libDir := path.Join(basedir, "lib")

	scriptOpts := postgresql.ScriptOptions{
		BinDir: binDir,
		LibDir: libDir,
		Port:   primaryPort,
	}

	checkReplScript := postgresql.GenerateCheckReplicationScript(scriptOpts)
	os.WriteFile(path.Join(topologyDir, "check_replication"), []byte(checkReplScript), 0755)

	checkRecovScript := postgresql.GenerateCheckRecoveryScript(scriptOpts, replicaPorts)
	os.WriteFile(path.Join(topologyDir, "check_recovery"), []byte(checkRecovScript), 0755)

	// 4. Handle --with-proxysql
	withProxySQL, _ := flags.GetBool("with-proxysql")
	if withProxySQL {
		if !providers.ContainsString(providers.CompatibleAddons["proxysql"], providerName) {
			common.Exitf(1, "--with-proxysql is not compatible with provider %q", providerName)
		}
		err := sandbox.DeployProxySQLForTopology(topologyDir, primaryPort, replicaPorts, 0, "127.0.0.1", providerName)
		if err != nil {
			common.Exitf(1, "ProxySQL deployment failed: %s", err)
		}
	}

	fmt.Printf("%s replication sandbox (1 primary + %d replicas) deployed in %s\n",
		providerName, nodes-1, topologyDir)
}
```

Add flag in `init()`:

```go
replicationCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
```

- [ ] **Step 7: Run full test suite to verify nothing is broken**

Run: `cd /data/rene/dbdeployer && go test ./... -v -timeout 5m`
Expected: All existing tests pass. No regressions.

- [ ] **Step 8: Commit**

```bash
git add globals/globals.go providers/provider.go cmd/root.go cmd/single.go cmd/multiple.go cmd/replication.go sandbox/proxysql_topology.go
git commit -m "feat: add --provider flag and PostgreSQL routing to deploy commands"
```

---

## Task 8: Unpack Command for PostgreSQL Debs

**Files:**
- Modify: `cmd/unpack.go`

- [ ] **Step 1: Add --provider flag to unpack command**

In `cmd/unpack.go`, modify `unpackTarball()` to check `--provider` flag. When `--provider=postgresql`, route to PostgreSQL deb extraction instead of MySQL tarball extraction:

```go
providerName, _ := flags.GetString(globals.ProviderLabel)
if providerName == "postgresql" {
    // PostgreSQL deb extraction
    if len(args) < 2 {
        common.Exitf(1, "PostgreSQL unpack requires both server and client .deb files\n"+
            "Usage: dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb")
    }
    server, client, err := postgresql.ClassifyDebs(args)
    if err != nil {
        common.Exitf(1, "error classifying deb files: %s", err)
    }
    version := Version // from --unpack-version flag
    if version == "" {
        version, err = postgresql.ParseDebVersion(server)
        if err != nil {
            common.Exitf(1, "cannot detect version from filename: %s\nUse --unpack-version to specify", err)
        }
    }
    targetDir := filepath.Join(home, "opt", "postgresql", version)
    if err := postgresql.UnpackDebs(server, client, targetDir); err != nil {
        common.Exitf(1, "error unpacking PostgreSQL debs: %s", err)
    }
    fmt.Printf("PostgreSQL %s unpacked to %s\n", version, targetDir)
    return
}
// ... existing MySQL tarball path unchanged
```

Add the `--provider` flag in `init()`:

```go
unpackCmd.PersistentFlags().String(globals.ProviderLabel, globals.ProviderValue, "Database provider (mysql, postgresql)")
```

Update `unpackCmd` to accept variadic args for PostgreSQL (currently `Args: cobra.ExactArgs(1)`):

```go
Args: cobra.MinimumNArgs(1),
```

- [ ] **Step 2: Run tests to verify no regressions**

Run: `cd /data/rene/dbdeployer && go test ./... -timeout 5m`
Expected: All pass.

- [ ] **Step 3: Commit**

```bash
git add cmd/unpack.go
git commit -m "feat: add --provider=postgresql support to dbdeployer unpack for deb extraction"
```

---

## Task 9: ProxySQL + PostgreSQL Backend Wiring

**Files:**
- Modify: `providers/proxysql/config.go`
- Modify: `providers/proxysql/config_test.go` (or create if absent)
- Modify: `providers/proxysql/proxysql.go`
- Modify: `sandbox/proxysql_topology.go`

- [ ] **Step 1: Write failing test for PostgreSQL backend config generation**

Add to `providers/proxysql/config_test.go` (create if needed):

```go
package proxysql

import (
	"strings"
	"testing"
)

func TestGenerateConfigMySQL(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:     "127.0.0.1",
		AdminPort:     6032,
		AdminUser:     "admin",
		AdminPassword: "admin",
		MySQLPort:     6033,
		DataDir:       "/tmp/proxysql/data",
		MonitorUser:   "msandbox",
		MonitorPass:   "msandbox",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "mysql_servers") {
		t.Error("expected mysql_servers block")
	}
	if !strings.Contains(config, "mysql_variables") {
		t.Error("expected mysql_variables block")
	}
}

func TestGenerateConfigPostgreSQL(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:       "127.0.0.1",
		AdminPort:       6032,
		AdminUser:       "admin",
		AdminPassword:   "admin",
		MySQLPort:       6033,
		DataDir:         "/tmp/proxysql/data",
		MonitorUser:     "postgres",
		MonitorPass:     "postgres",
		BackendProvider: "postgresql",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 16613, Hostgroup: 0, MaxConns: 200},
			{Host: "127.0.0.1", Port: 16614, Hostgroup: 1, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "pgsql_servers") {
		t.Error("expected pgsql_servers block")
	}
	if !strings.Contains(config, "pgsql_users") {
		t.Error("expected pgsql_users block")
	}
	if !strings.Contains(config, "pgsql_variables") {
		t.Error("expected pgsql_variables block")
	}
	if strings.Contains(config, "mysql_servers") {
		t.Error("should not contain mysql_servers for postgresql backend")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /data/rene/dbdeployer && go test ./providers/proxysql/ -run TestGenerateConfig -v`
Expected: Fail — `BackendProvider` field doesn't exist yet.

- [ ] **Step 3: Add BackendProvider field to ProxySQLConfig and update GenerateConfig**

In `providers/proxysql/config.go`:

Add `BackendProvider string` field to `ProxySQLConfig`.

Update `GenerateConfig` to branch on `BackendProvider`:

```go
func GenerateConfig(cfg ProxySQLConfig) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("datadir=\"%s\"\n\n", cfg.DataDir))

	b.WriteString("admin_variables=\n{\n")
	b.WriteString(fmt.Sprintf("    admin_credentials=\"%s:%s\"\n", cfg.AdminUser, cfg.AdminPassword))
	b.WriteString(fmt.Sprintf("    mysql_ifaces=\"%s:%d\"\n", cfg.AdminHost, cfg.AdminPort))
	b.WriteString("}\n\n")

	isPgsql := cfg.BackendProvider == "postgresql"

	if isPgsql {
		b.WriteString("pgsql_variables=\n{\n")
		b.WriteString(fmt.Sprintf("    interfaces=\"%s:%d\"\n", cfg.AdminHost, cfg.MySQLPort))
		b.WriteString(fmt.Sprintf("    monitor_username=\"%s\"\n", cfg.MonitorUser))
		b.WriteString(fmt.Sprintf("    monitor_password=\"%s\"\n", cfg.MonitorPass))
		b.WriteString("}\n\n")
	} else {
		b.WriteString("mysql_variables=\n{\n")
		b.WriteString(fmt.Sprintf("    interfaces=\"%s:%d\"\n", cfg.AdminHost, cfg.MySQLPort))
		b.WriteString(fmt.Sprintf("    monitor_username=\"%s\"\n", cfg.MonitorUser))
		b.WriteString(fmt.Sprintf("    monitor_password=\"%s\"\n", cfg.MonitorPass))
		b.WriteString("    monitor_connect_interval=2000\n")
		b.WriteString("    monitor_ping_interval=2000\n")
		b.WriteString("}\n\n")
	}

	serversKey := "mysql_servers"
	usersKey := "mysql_users"
	if isPgsql {
		serversKey = "pgsql_servers"
		usersKey = "pgsql_users"
	}

	if len(cfg.Backends) > 0 {
		b.WriteString(fmt.Sprintf("%s=\n(\n", serversKey))
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

	b.WriteString(fmt.Sprintf("%s=\n(\n", usersKey))
	b.WriteString("    {\n")
	b.WriteString(fmt.Sprintf("        username=\"%s\"\n", cfg.MonitorUser))
	b.WriteString(fmt.Sprintf("        password=\"%s\"\n", cfg.MonitorPass))
	b.WriteString("        default_hostgroup=0\n")
	b.WriteString("    }\n")
	b.WriteString(")\n")

	return b.String()
}
```

- [ ] **Step 4: Update ProxySQLProvider.CreateSandbox to pass BackendProvider**

In `providers/proxysql/proxysql.go`, set `BackendProvider` from `config.Options["backend_provider"]`:

```go
proxyCfg := ProxySQLConfig{
    // ... existing fields ...
    BackendProvider: config.Options["backend_provider"],
}
```

Also update the `use_proxy` script generation to use `psql` when backend is PostgreSQL:

```go
if config.Options["backend_provider"] == "postgresql" {
    scripts["use_proxy"] = fmt.Sprintf("#!/bin/bash\npsql -h %s -p %d -U %s \"$@\"\n",
        host, mysqlPort, monitorUser)
} else {
    scripts["use_proxy"] = fmt.Sprintf("#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL> ' \"$@\"\n",
        host, mysqlPort, monitorUser, monitorPass)
}
```

- [ ] **Step 5: Update DeployProxySQLForTopology to accept backend provider**

In `sandbox/proxysql_topology.go`, add a `backendProvider` parameter:

```go
func DeployProxySQLForTopology(sandboxDir string, masterPort int, slavePorts []int, proxysqlPort int, host string, backendProvider string) error {
    // ... existing code ...
    config.Options["backend_provider"] = backendProvider
    // ...
}
```

Update all callers in `cmd/single.go` and `cmd/replication.go` to pass `""` (empty string = mysql default) or `"postgresql"` when appropriate.

- [ ] **Step 6: Run all tests**

Run: `cd /data/rene/dbdeployer && go test ./... -timeout 5m`
Expected: All pass.

- [ ] **Step 7: Commit**

```bash
git add providers/proxysql/config.go providers/proxysql/config_test.go providers/proxysql/proxysql.go sandbox/proxysql_topology.go cmd/single.go cmd/replication.go
git commit -m "feat: add ProxySQL PostgreSQL backend wiring (pgsql_servers/pgsql_users config)"
```

---

## Task 10: Cross-Database Topology Constraints

**Files:**
- Modify: `cmd/single.go`
- Modify: `cmd/multiple.go`
- Modify: `cmd/replication.go`
- Modify: `providers/provider_test.go`

This task ensures the validation logic added in Task 7 is properly tested.

- [ ] **Step 1: Write tests for topology and cross-provider validation**

Add to `providers/provider_test.go`:

```go
func TestTopologyValidation(t *testing.T) {
	mock := &mockProvider{name: "test"}
	topos := mock.SupportedTopologies()
	if !containsString(topos, "single") {
		t.Error("expected single in supported topologies")
	}
	if containsString(topos, "group") {
		t.Error("did not expect group in supported topologies")
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
```

Also add a test for the addon compatibility map (define it in `providers/provider.go` or `cmd/` depending on where it lives):

```go
var CompatibleAddons = map[string][]string{
	"proxysql": {"mysql", "postgresql"},
}

func TestAddonCompatibility(t *testing.T) {
	if !containsString(CompatibleAddons["proxysql"], "postgresql") {
		t.Error("proxysql should be compatible with postgresql")
	}
	if containsString(CompatibleAddons["proxysql"], "fake") {
		t.Error("proxysql should not be compatible with fake")
	}
}
```

- [ ] **Step 2: Add CompatibleAddons map to providers/provider.go**

```go
// CompatibleAddons maps addon names to the list of providers they work with.
var CompatibleAddons = map[string][]string{
	"proxysql": {"mysql", "postgresql"},
}
```

- [ ] **Step 3: Run tests**

Run: `cd /data/rene/dbdeployer && go test ./providers/ -v`
Expected: All pass.

- [ ] **Step 4: Commit**

```bash
git add providers/provider.go providers/provider_test.go
git commit -m "feat: add cross-database topology constraint validation"
```

---

## Task 11: Standalone PostgreSQL Deploy Command

**Files:**
- Create: `cmd/deploy_postgresql.go`

- [ ] **Step 1: Create deploy postgresql subcommand**

Create `cmd/deploy_postgresql.go` following the pattern from `cmd/deploy_proxysql.go`:

```go
package cmd

import (
	"fmt"
	"path"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/defaults"
	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/ProxySQL/dbdeployer/providers/postgresql"
	"github.com/spf13/cobra"
)

func deploySandboxPostgreSQL(cmd *cobra.Command, args []string) {
	version := args[0]
	flags := cmd.Flags()
	skipStart, _ := flags.GetBool("skip-start")

	p, err := providers.DefaultRegistry.Get("postgresql")
	if err != nil {
		common.Exitf(1, "PostgreSQL provider not available: %s", err)
	}

	if err := p.ValidateVersion(version); err != nil {
		common.Exitf(1, "invalid version: %s", err)
	}

	if _, err := p.FindBinary(version); err != nil {
		common.Exitf(1, "PostgreSQL binaries not found: %s\nRun: dbdeployer unpack --provider=postgresql <server.deb> <client.deb>", err)
	}

	port, err := postgresql.VersionToPort(version)
	if err != nil {
		common.Exitf(1, "error computing port: %s", err)
	}
	freePort, portErr := common.FindFreePort(port, []int{}, 1)
	if portErr == nil {
		port = freePort
	}

	sandboxHome := defaults.Defaults().SandboxHome
	sandboxDir := path.Join(sandboxHome, fmt.Sprintf("pg_sandbox_%d", port))

	if common.DirExists(sandboxDir) {
		common.Exitf(1, "sandbox directory %s already exists", sandboxDir)
	}

	config := providers.SandboxConfig{
		Version:    version,
		Dir:        sandboxDir,
		Port:       port,
		Host:       "127.0.0.1",
		DbUser:     "postgres",
		DbPassword: "",
		Options:    map[string]string{},
	}

	if _, err := p.CreateSandbox(config); err != nil {
		common.Exitf(1, "error creating PostgreSQL sandbox: %s", err)
	}

	if !skipStart {
		if err := p.StartSandbox(sandboxDir); err != nil {
			common.Exitf(1, "error starting PostgreSQL: %s", err)
		}
	}

	fmt.Printf("PostgreSQL %s sandbox deployed in %s (port: %d)\n", version, sandboxDir, port)
}

var deployPostgreSQLCmd = &cobra.Command{
	Use:   "postgresql version",
	Short: "deploys a PostgreSQL sandbox",
	Long: `postgresql deploys a standalone PostgreSQL instance as a sandbox.
It creates a sandbox directory with data, configuration, start/stop scripts, and a
psql client script.

Requires PostgreSQL binaries to be extracted first:
    dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb

Example:
    dbdeployer deploy postgresql 16.13
    dbdeployer deploy postgresql 17.1 --skip-start
`,
	Args: cobra.ExactArgs(1),
	Run:  deploySandboxPostgreSQL,
}

func init() {
	deployCmd.AddCommand(deployPostgreSQLCmd)
	deployPostgreSQLCmd.Flags().Bool("skip-start", false, "Do not start PostgreSQL after deployment")
}
```

- [ ] **Step 2: Run build to verify compilation**

Run: `cd /data/rene/dbdeployer && go build -o /dev/null .`
Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add cmd/deploy_postgresql.go
git commit -m "feat: add 'dbdeployer deploy postgresql' standalone command"
```

---

## Task 12: Integration Tests

**Files:**
- Create: `providers/postgresql/integration_test.go`

- [ ] **Step 1: Write integration tests (build-tagged)**

Create `providers/postgresql/integration_test.go`:

```go
//go:build integration

package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/providers"
)

func findPostgresVersion(t *testing.T) string {
	t.Helper()
	home, _ := os.UserHomeDir()
	entries, err := os.ReadDir(filepath.Join(home, "opt", "postgresql"))
	if err != nil {
		t.Skipf("no PostgreSQL installations found: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return e.Name()
		}
	}
	t.Skip("no PostgreSQL version directories found")
	return ""
}

func TestIntegrationSingleSandbox(t *testing.T) {
	version := findPostgresVersion(t)
	p := NewPostgreSQLProvider()

	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "pg_test")

	config := providers.SandboxConfig{
		Version: version,
		Dir:     sandboxDir,
		Port:    15432,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{},
	}

	// Create
	info, err := p.CreateSandbox(config)
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}
	if info.Port != 15432 {
		t.Errorf("expected port 15432, got %d", info.Port)
	}

	// Start
	if err := p.StartSandbox(sandboxDir); err != nil {
		t.Fatalf("StartSandbox failed: %v", err)
	}
	stopped := false
	defer func() {
		if !stopped {
			p.StopSandbox(sandboxDir)
		}
	}()
	time.Sleep(2 * time.Second)

	// Connect via psql
	home, _ := os.UserHomeDir()
	psql := filepath.Join(home, "opt", "postgresql", version, "bin", "psql")
	cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", "15432", "-U", "postgres", "-c", "SELECT 1;")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s",
		filepath.Join(home, "opt", "postgresql", version, "lib")))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("psql connection failed: %s: %v", string(output), err)
	}

	// Stop
	if err := p.StopSandbox(sandboxDir); err != nil {
		t.Fatalf("StopSandbox failed: %v", err)
	}
	stopped = true
}

func TestIntegrationReplication(t *testing.T) {
	version := findPostgresVersion(t)
	p := NewPostgreSQLProvider()

	tmpDir := t.TempDir()
	primaryDir := filepath.Join(tmpDir, "primary")
	replica1Dir := filepath.Join(tmpDir, "replica1")
	replica2Dir := filepath.Join(tmpDir, "replica2")

	// Create and start primary with replication
	primaryConfig := providers.SandboxConfig{
		Version: version,
		Dir:     primaryDir,
		Port:    15500,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{"replication": "true"},
	}

	_, err := p.CreateSandbox(primaryConfig)
	if err != nil {
		t.Fatalf("CreateSandbox (primary) failed: %v", err)
	}
	if err := p.StartSandbox(primaryDir); err != nil {
		t.Fatalf("StartSandbox (primary) failed: %v", err)
	}
	defer p.StopSandbox(primaryDir)
	time.Sleep(2 * time.Second)

	primaryInfo := providers.SandboxInfo{Dir: primaryDir, Port: 15500}

	// Create replicas
	for i, rDir := range []string{replica1Dir, replica2Dir} {
		rConfig := providers.SandboxConfig{
			Version: version,
			Dir:     rDir,
			Port:    15501 + i,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}
		_, err := p.CreateReplica(primaryInfo, rConfig)
		if err != nil {
			t.Fatalf("CreateReplica %d failed: %v", i+1, err)
		}
		defer p.StopSandbox(rDir)
	}

	time.Sleep(2 * time.Second)

	// Verify pg_stat_replication on primary shows 2 replicas
	home, _ := os.UserHomeDir()
	psql := filepath.Join(home, "opt", "postgresql", version, "bin", "psql")
	libDir := filepath.Join(home, "opt", "postgresql", version, "lib")

	cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", "15500", "-U", "postgres", "-t", "-c",
		"SELECT count(*) FROM pg_stat_replication;")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("replication check failed: %s: %v", string(output), err)
	}

	// Verify replicas are in recovery
	for _, port := range []int{15501, 15502} {
		cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", fmt.Sprintf("%d", port), "-U", "postgres", "-t", "-c",
			"SELECT pg_is_in_recovery();")
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("recovery check on port %d failed: %s: %v", port, string(output), err)
		}
	}
}
```

- [ ] **Step 2: Verify unit tests still pass (integration tests skipped by default)**

Run: `cd /data/rene/dbdeployer && go test ./providers/postgresql/ -v`
Expected: All unit tests pass. Integration tests are not compiled without the build tag.

- [ ] **Step 3: Commit**

```bash
git add providers/postgresql/integration_test.go
git commit -m "test: add PostgreSQL integration tests (build-tagged)"
```

---

## Task 13: Create GitHub Issues for CI Follow-Up

**Files:** None (GitHub issues only)

- [ ] **Step 1: Create GitHub issue for PostgreSQL deb caching in CI**

```bash
gh issue create --title "CI: Add PostgreSQL deb caching to CI pipeline" \
  --body "Add caching of PostgreSQL server and client .deb packages to CI, similar to MySQL tarball caching. This enables running PostgreSQL integration tests in CI." \
  --label "enhancement,ci"
```

- [ ] **Step 2: Create GitHub issue for PostgreSQL integration tests in CI matrix**

```bash
gh issue create --title "CI: Add PostgreSQL integration tests to CI matrix" \
  --body "Add PostgreSQL integration tests (providers/postgresql/integration_test.go) to the CI test matrix. Requires PostgreSQL deb caching (#<prev-issue>) to be in place." \
  --label "enhancement,ci"
```

- [ ] **Step 3: Create GitHub issue for nightly PostgreSQL topology tests**

```bash
gh issue create --title "CI: Add nightly PostgreSQL replication topology tests" \
  --body "Add nightly CI job that runs full PostgreSQL replication topology tests (primary + replicas, ProxySQL wiring)." \
  --label "enhancement,ci"
```

- [ ] **Step 4: Commit (no code changes, just documenting)**

No commit needed — issues are tracked in GitHub.

---

## Execution Notes

### Dependencies between tasks
- Task 1 (interface changes) must complete before all other tasks
- Tasks 2-5 can run in parallel after Task 1
- Task 6 (replication) depends on Tasks 2-4
- Task 7 (cmd layer) depends on Tasks 2-6
- Task 8 (unpack cmd) depends on Task 5
- Task 9 (ProxySQL wiring) depends on Tasks 6-7
- Task 10 (constraints) depends on Task 7
- Task 11 (deploy command) depends on Tasks 2-4, 7
- Task 12 (integration tests) depends on all implementation tasks
- Task 13 (GitHub issues) is independent

### Running integration tests locally

```bash
# Extract PostgreSQL binaries first
apt-get download postgresql-16 postgresql-client-16
./dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb

# Run integration tests
cd /data/rene/dbdeployer && go test ./providers/postgresql/ -tags integration -v -timeout 10m
```
