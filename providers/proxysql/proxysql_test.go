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
