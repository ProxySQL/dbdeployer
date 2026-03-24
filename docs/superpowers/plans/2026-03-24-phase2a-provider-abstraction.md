# Phase 2a: Provider Abstraction & MySQL Refactor

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the Provider abstraction layer and refactor the existing MySQL sandbox code behind it, so that new providers (ProxySQL, Orchestrator, PostgreSQL) can be added cleanly in Phase 2b.

**Architecture:** Create a `providers/` package with the `Provider` interface and `ProviderRegistry`. Move MySQL-specific sandbox logic into `providers/mysql/`, keeping the existing `sandbox/` package as a thinner orchestration layer that works through the registry. The `cmd/` layer routes through the registry. All existing functionality must continue to work identically — this is a pure refactoring.

**Tech Stack:** Go 1.22+, existing Cobra CLI framework

**Spec:** `docs/superpowers/specs/2026-03-23-dbdeployer-revitalization-design.md`

**Key constraint:** Every task must leave the codebase in a compilable, test-passing state. No big-bang refactor.

---

## File Structure

### New files to create:
```
providers/
  provider.go          # Provider interface, Instance, PortRange, ProviderRegistry
  provider_test.go     # Registry tests with mock provider
  mysql/
    mysql.go           # MySQLProvider implementing Provider interface
    mysql_test.go      # MySQL provider unit tests
```

### Files to modify:
```
cmd/root.go                  # Register MySQL provider in existing init()
cmd/single.go               # Add provider validation before sandbox creation
cmd/replication.go           # Add provider validation before sandbox creation
cmd/multiple.go              # Add provider validation before sandbox creation
```

Note: `sandbox/sandbox.go` and `sandbox/replication.go` are NOT modified in Phase 2a. Moving sandbox logic behind the provider interface is deferred to Phase 2b when ProxySQL needs it.

### Files that stay as-is (no changes needed in Phase 2a):
```
sandbox/templates/           # All .gotxt files unchanged
sandbox/templates.go         # Template collections unchanged
sandbox/repl_templates.go    # Template collections unchanged
sandbox/group_replication.go # Touched minimally (registry lookup)
sandbox/multiple.go          # Touched minimally
sandbox/multi-source-replication.go
sandbox/ndb_replication.go
sandbox/pxc_replication.go
```

---

### Task 1: Define Provider interface and ProviderRegistry

**Files:**
- Create: `providers/provider.go`
- Create: `providers/provider_test.go`

This is the foundation. The interface is intentionally minimal for Phase 2a — just `Name()`, `ValidateVersion()`, and `DefaultPorts()`. The full interface from the spec (with `CreateSandbox`, `Start`, `Stop`, `Destroy`, `HealthCheck`) will be added in Phase 2b when ProxySQL needs it. This establishes the registry pattern first.

- [ ] **Step 1: Create `providers/provider.go` with interface and registry**

```go
package providers

import (
	"fmt"
	"sort"
)

// Provider is the core abstraction for deploying database infrastructure.
type Provider interface {
	// Name returns the provider identifier (e.g., "mysql", "proxysql").
	Name() string

	// ValidateVersion checks if the given version string is valid for this provider.
	ValidateVersion(version string) error

	// DefaultPorts returns the port allocation strategy for this provider.
	DefaultPorts() PortRange
}

// PortRange defines a provider's default port allocation.
type PortRange struct {
	BasePort         int // default starting port (e.g., 3306 for MySQL)
	PortsPerInstance int // how many ports each instance needs
}

// Registry manages available providers.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) error {
	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	r.providers[name] = p
	return nil
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}

// List returns names of all registered providers (sorted).
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DefaultRegistry is the global provider registry.
var DefaultRegistry = NewRegistry()
```

- [ ] **Step 2: Create `providers/provider_test.go`**

```go
package providers

import "testing"

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string                          { return m.name }
func (m *mockProvider) ValidateVersion(version string) error  { return nil }
func (m *mockProvider) DefaultPorts() PortRange               { return PortRange{BasePort: 9999, PortsPerInstance: 1} }

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	mock := &mockProvider{name: "test"}

	if err := reg.Register(mock); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	p, err := reg.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name() != "test" {
		t.Errorf("expected name 'test', got %q", p.Name())
	}
}

func TestRegistryDuplicateRegister(t *testing.T) {
	reg := NewRegistry()
	mock := &mockProvider{name: "test"}
	_ = reg.Register(mock)
	err := reg.Register(mock)
	if err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error on missing provider")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&mockProvider{name: "a"})
	_ = reg.Register(&mockProvider{name: "b"})
	names := reg.List()
	if len(names) != 2 {
		t.Errorf("expected 2 providers, got %d", len(names))
	}
}
```

- [ ] **Step 3: Verify tests pass**

Run: `go test ./providers/... -v`
Expected: All 4 tests pass.

- [ ] **Step 4: Commit**

```bash
git add providers/
git commit -m "feat: add Provider interface and ProviderRegistry"
```

---

### Task 2: Create MySQLProvider implementing the Provider interface

**Files:**
- Create: `providers/mysql/mysql.go`
- Create: `providers/mysql/mysql_test.go`

The MySQL provider starts minimal — just implementing the interface. It doesn't replace any existing functionality yet. That happens in Task 3.

- [ ] **Step 1: Create `providers/mysql/mysql.go`**

```go
package mysql

import (
	"fmt"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "mysql"

// MySQLProvider implements the Provider interface for MySQL and its flavors
// (Percona, MariaDB, NDB, PXC, TiDB).
type MySQLProvider struct{}

// NewMySQLProvider creates a new MySQL provider.
func NewMySQLProvider() *MySQLProvider {
	return &MySQLProvider{}
}

func (p *MySQLProvider) Name() string { return ProviderName }

func (p *MySQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid MySQL version format: %q (expected X.Y or X.Y.Z)", version)
	}
	return nil
}

func (p *MySQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{
		BasePort:         3306,
		PortsPerInstance: 3, // main port + mysqlx port + admin port
	}
}

// Register adds the MySQL provider to the given registry.
func Register(reg *providers.Registry) error {
	return reg.Register(NewMySQLProvider())
}
```

- [ ] **Step 2: Create `providers/mysql/mysql_test.go`**

```go
package mysql

import (
	"testing"

	"github.com/ProxySQL/dbdeployer/providers"
)

func TestMySQLProviderName(t *testing.T) {
	p := NewMySQLProvider()
	if p.Name() != "mysql" {
		t.Errorf("expected 'mysql', got %q", p.Name())
	}
}

func TestMySQLProviderValidateVersion(t *testing.T) {
	p := NewMySQLProvider()
	tests := []struct {
		version string
		wantErr bool
	}{
		{"8.4.4", false},
		{"9.1.0", false},
		{"5.7", false},
		{"invalid", true},
	}
	for _, tt := range tests {
		err := p.ValidateVersion(tt.version)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
		}
	}
}

func TestMySQLProviderRegister(t *testing.T) {
	reg := providers.NewRegistry()
	if err := Register(reg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	p, err := reg.Get("mysql")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name() != "mysql" {
		t.Errorf("expected 'mysql', got %q", p.Name())
	}
}
```

- [ ] **Step 3: Verify tests pass**

Run: `go test ./providers/... -v`
Expected: All tests pass (both providers/ and providers/mysql/).

- [ ] **Step 4: Commit**

```bash
git add providers/mysql/
git commit -m "feat: add MySQLProvider implementing Provider interface"
```

---

### Task 3: Register MySQLProvider at startup and wire into cmd/root.go

**Files:**
- Modify: `cmd/root.go` (add provider registration to existing init function)

This wires the provider registry into the application lifecycle without changing any existing behavior. No change to `main.go` is needed since it already imports `cmd`.

- [ ] **Step 1: Add MySQL provider registration to the existing init() in cmd/root.go**

`cmd/root.go` already has an `init()` function (around line 145). Add the provider registration at the top of that existing function:

```go
import (
	"github.com/ProxySQL/dbdeployer/providers"
	mysqlprovider "github.com/ProxySQL/dbdeployer/providers/mysql"
)

func init() {
	// Register built-in providers
	if err := mysqlprovider.Register(providers.DefaultRegistry); err != nil {
		// This should never happen at startup
		panic(fmt.Sprintf("failed to register MySQL provider: %v", err))
	}
}
```

- [ ] **Step 2: Verify the application still builds and runs**

```bash
go build -o dbdeployer .
./dbdeployer --version
```
Expected: Outputs version 1.74.1 (or current). No behavior change.

- [ ] **Step 3: Run all unit tests**

Run: `go test ./... -timeout 30m 2>&1 | grep -E "^(ok|FAIL)" | grep -v "sandbox\|ts\b"`
Expected: All packages pass.

- [ ] **Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "feat: register MySQLProvider at startup via DefaultRegistry"
```

---

### Task 4: Add provider lookup to cmd/single.go

**Files:**
- Modify: `cmd/single.go`

This is the first cmd/ file to use the registry. It looks up the MySQL provider and validates the version before calling the existing sandbox creation. Minimal change — just adds a validation step.

- [ ] **Step 1: Read cmd/single.go and understand the current flow**

Find the function that handles `dbdeployer deploy single <version>`. It calls into `sandbox.CreateStandaloneSandbox()`. Add a provider lookup + validation before that call.

- [ ] **Step 2: Add provider validation**

After `fillSandboxDefinition()` returns and before `CreateStandaloneSandbox()` is called, add provider validation using `sd.Version` (which is the resolved version, not the raw CLI argument):

```go
// Validate version with provider
// TODO: Phase 2b — determine provider from sd.Flavor instead of hardcoding "mysql"
p, err := providers.DefaultRegistry.Get("mysql")
if err != nil {
	common.Exitf(1, "provider error: %s", err)
}
if err := p.ValidateVersion(sd.Version); err != nil {
	common.Exitf(1, "version validation failed: %s", err)
}
```

This is additive — existing code continues to work, we just add a validation gate. The `ValidateVersion` call is a seam for future use; the existing code already does extensive version checking.

- [ ] **Step 3: Verify single sandbox deployment still works**

```bash
go build -o dbdeployer .
./dbdeployer deploy single 8.4.4 --sandbox-binary=$HOME/opt/mysql
~/sandboxes/msb_8_4_4/use -e "SELECT VERSION()"
./dbdeployer delete all --skip-confirm
```

- [ ] **Step 4: Commit**

```bash
git add cmd/single.go
git commit -m "feat: add provider validation to single sandbox deployment"
```

---

### Task 5: Add provider lookup to cmd/replication.go and cmd/multiple.go

**Files:**
- Modify: `cmd/replication.go`
- Modify: `cmd/multiple.go`

Same pattern as Task 4 — add provider validation before existing sandbox creation calls.

- [ ] **Step 1: Add provider validation to cmd/replication.go**

Same pattern: look up "mysql" provider, validate version, then proceed with existing flow.

- [ ] **Step 2: Add provider validation to cmd/multiple.go**

Same pattern.

- [ ] **Step 3: Verify replication deployment still works**

```bash
go build -o dbdeployer .
./dbdeployer deploy replication 8.4.4 --sandbox-binary=$HOME/opt/mysql
~/sandboxes/rsandbox_8_4_4/check_slaves
./dbdeployer delete all --skip-confirm
```

- [ ] **Step 4: Run all unit tests**

Run: `go test ./cmd/... -v -timeout 30m`
Expected: All cmd tests pass.

- [ ] **Step 5: Commit**

```bash
git add cmd/replication.go cmd/multiple.go
git commit -m "feat: add provider validation to replication and multiple deployments"
```

---

### Task 6: Add `dbdeployer providers list` command

**Files:**
- Create: `cmd/providers.go`

A new CLI command that lists registered providers. This makes the provider system visible to users and verifies the registry is wired correctly end-to-end.

- [ ] **Step 1: Create `cmd/providers.go`**

```go
package cmd

import (
	"fmt"

	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Shows available deployment providers",
	Long:  "Lists all registered providers that can be used for sandbox deployment",
	Run: func(cmd *cobra.Command, args []string) {
		for _, name := range providers.DefaultRegistry.List() {
			p, _ := providers.DefaultRegistry.Get(name)
			ports := p.DefaultPorts()
			fmt.Printf("%-15s (base port: %d, ports per instance: %d)\n",
				name, ports.BasePort, ports.PortsPerInstance)
		}
	},
}

func init() {
	rootCmd.AddCommand(providersCmd)
}
```

- [ ] **Step 2: Build and test**

```bash
go build -o dbdeployer .
./dbdeployer providers
```
Expected output:
```
mysql           (base port: 3306, ports per instance: 3)
```

- [ ] **Step 3: Commit**

```bash
git add cmd/providers.go
git commit -m "feat: add 'dbdeployer providers' command to list registered providers"
```

---

### Task 7: Final validation and cleanup

- [ ] **Step 1: Run all unit tests**

```bash
go test ./providers/... ./cmd/... ./common/... ./downloads/... ./ops/... -timeout 30m
```
Expected: All pass.

- [ ] **Step 2: Run integration test locally**

```bash
go build -o dbdeployer .
# Single
./dbdeployer deploy single 8.4.4 --sandbox-binary=$HOME/opt/mysql
~/sandboxes/msb_8_4_4/use -e "SELECT VERSION()"
./dbdeployer delete all --skip-confirm
# Replication
./dbdeployer deploy replication 9.1.0 --sandbox-binary=$HOME/opt/mysql
~/sandboxes/rsandbox_9_1_0/check_slaves
./dbdeployer delete all --skip-confirm
# Providers command
./dbdeployer providers
```

- [ ] **Step 3: Verify no regressions in existing behavior**

The provider layer is purely additive in Phase 2a. No existing command syntax or behavior should change. The only new command is `dbdeployer providers`.

- [ ] **Step 4: Commit any final fixes**

---

## What Phase 2a Does NOT Do (Deferred to Phase 2b)

- Does NOT decompose SandboxDef into base + provider-specific structs (that happens when ProxySQL needs a different config shape)
- Does NOT move MySQL sandbox creation logic into providers/mysql/ (the Provider interface is established but MySQL's `CreateSandbox` still lives in `sandbox/`)
- Does NOT add ProxySQL, Orchestrator, or PostgreSQL providers
- Does NOT add topology-aware multi-provider deployment (`--with-proxysql`)
- Does NOT change the sandbox catalog

These are intentionally deferred to keep Phase 2a small, safe, and mergeable. The Provider interface and Registry are the foundation; Phase 2b builds on them.
