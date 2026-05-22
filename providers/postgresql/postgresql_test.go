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

func TestGenerateTestReplicationScript(t *testing.T) {
	script, err := GenerateTestReplicationScript(TestReplicationOptions{
		SandboxDir: "/home/user/sandboxes/postgresql_repl_16614",
		ShellPath:  "/bin/bash",
	})
	if err != nil {
		t.Fatalf("GenerateTestReplicationScript: %v", err)
	}
	for _, want := range []string{
		"#!/bin/bash",
		"/home/user/sandboxes/postgresql_repl_16614",
		"./primary/use",
		"./replica$N/use",
		"pg_current_wal_lsn",
		"pg_last_wal_replay_lsn",
		"test_summary",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("test_replication script missing %q", want)
		}
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

func TestSandboxUserGrantsSQL(t *testing.T) {
	sql := SandboxUserGrantsSQL()
	for _, want := range []string{
		"CREATE USER rsandbox WITH PASSWORD 'rsandbox'",
		"CREATE DATABASE rsandbox",
		"GRANT CONNECT ON DATABASE postgres TO rsandbox",
		"GRANT CONNECT ON DATABASE rsandbox TO rsandbox",
		"GRANT CREATE ON SCHEMA public TO rsandbox",
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("SandboxUserGrantsSQL() missing %q", want)
		}
	}
}

func TestSandboxDatabaseGrantsSQL(t *testing.T) {
	sql := SandboxDatabaseGrantsSQL()
	if sql != "GRANT CREATE ON SCHEMA public TO rsandbox;" {
		t.Errorf("unexpected SandboxDatabaseGrantsSQL: %q", sql)
	}
}

func TestGenerateInitSandboxUserScript(t *testing.T) {
	script := GenerateInitSandboxUserScript(ScriptOptions{
		BinDir: "/opt/postgresql/16.13/bin",
		LibDir: "/opt/postgresql/16.13/lib",
		Port:   16613,
	})
	if !strings.Contains(script, "pg_is_in_recovery") {
		t.Error("init script should skip on recovery standbys")
	}
	if !strings.Contains(script, "CREATE USER rsandbox") {
		t.Error("init script should create rsandbox user")
	}
	if strings.Contains(script, `$('`) {
		t.Error("init script has invalid shell quoting around psql path")
	}
	if !strings.Contains(script, `recovery=$("`) {
		t.Error("init script should quote psql binary path")
	}
	if !strings.Contains(script, "-d rsandbox") {
		t.Error("init script should connect to rsandbox database for schema grants")
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

	expectedScripts := []string{"start", "stop", "status", "restart", "use", "clear", "init_sandbox_user"}
	for _, name := range expectedScripts {
		if _, ok := scripts[name]; !ok {
			t.Errorf("missing script %q", name)
		}
	}

	start := scripts["start"]
	if !strings.Contains(start, "pg_ctl") {
		t.Error("start script missing pg_ctl")
	}
	if !strings.Contains(start, "init_sandbox_user") {
		t.Error("start script should run init_sandbox_user")
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
