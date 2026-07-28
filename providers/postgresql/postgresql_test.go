package postgresql

import (
	"strings"
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

func TestPostgreSQLProxySQLAdminPort(t *testing.T) {
	tests := []struct {
		version  string
		expected int
	}{
		{"16.14", 6132},
		{"16.13", 6131},
		{"16.3", 6121},
	}
	for _, tt := range tests {
		primaryPort, err := VersionToPort(tt.version)
		if err != nil {
			t.Fatalf("VersionToPort(%q): %v", tt.version, err)
		}
		if got := ProxySQLAdminPort(primaryPort); got != tt.expected {
			t.Errorf("ProxySQLAdminPort(%q) = %d, want %d", tt.version, got, tt.expected)
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

func TestGenerateTestReplicationScript(t *testing.T) {
	script := GenerateTestReplicationScript(ScriptOptions{
		SandboxDir: "/home/test/sandboxes/postgresql_repl_16614",
		BinDir:     "/opt/postgresql/16.13/bin",
		LibDir:     "/opt/postgresql/16.13/lib",
		Port:       16614,
	}, 2)
	expectedSubstrings := []string{
		`SBDIR=/home/test/sandboxes/postgresql_repl_16614`,
		"./primary/use",
		"pg_current_wal_lsn()",
		"pg_last_wal_replay_lsn()",
		"pg_is_in_recovery()",
		"./replica1/use",
		"./replica2/use",
		"test_summary",
	}
	for _, want := range expectedSubstrings {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q", want)
		}
	}
	// numReplicas=2 should NOT mention replica3
	if strings.Contains(script, "./replica3/use") {
		t.Error("script unexpectedly references replica3")
	}
}

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

	expectedScripts := []string{"start", "stop", "status", "restart", "use", "clear"}
	for _, name := range expectedScripts {
		if _, ok := scripts[name]; !ok {
			t.Errorf("missing script %q", name)
		}
	}

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

	use := scripts["use"]
	if !strings.Contains(use, "psql") {
		t.Error("use script missing psql")
	}
	if !strings.Contains(use, "16613") {
		t.Error("use script missing port")
	}
}

// TestGenerateScriptsDbUserPropagation ensures the configured db-user is embedded
// in the generated psql/initdb invocations, and that an empty DbUser falls back
// to the historical "postgres" default.
func TestGenerateScriptsDbUserPropagation(t *testing.T) {
	// Explicit user: scripts must reference it and not the default.
	custom := GenerateScripts(ScriptOptions{
		BinDir:  "/opt/postgresql/16.13/bin",
		DataDir: "/tmp/pg/data",
		Port:    16613,
		DbUser:  "wse",
	})
	if !strings.Contains(custom["use"], "-U wse") {
		t.Errorf("use script must embed -U wse; got: %q", custom["use"])
	}
	if strings.Contains(custom["use"], "-U postgres") {
		t.Errorf("use script must not reference -U postgres; got: %q", custom["use"])
	}
	if !strings.Contains(custom["clear"], "--username=wse") {
		t.Errorf("clear script must embed --username=wse; got: %q", custom["clear"])
	}

	// Empty user: falls back to "postgres" (backward compatible).
	def := GenerateScripts(ScriptOptions{
		BinDir:  "/opt/postgresql/16.13/bin",
		DataDir: "/tmp/pg/data",
		Port:    16613,
	})
	if !strings.Contains(def["use"], "-U postgres") {
		t.Errorf("default use script must embed -U postgres; got: %q", def["use"])
	}
	if !strings.Contains(def["clear"], "--username=postgres") {
		t.Errorf("default clear script must embed --username=postgres; got: %q", def["clear"])
	}

	// Replication scripts honor DbUser too.
	repl := GenerateCheckReplicationScript(ScriptOptions{
		BinDir: "/opt/postgresql/16.13/bin",
		Port:   16613,
		DbUser: "wse",
	})
	if !strings.Contains(repl, "-U wse") {
		t.Errorf("replication script must embed -U wse; got: %q", repl)
	}
}
