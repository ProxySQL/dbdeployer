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
