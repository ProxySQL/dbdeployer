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
