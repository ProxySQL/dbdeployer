package proxysql

import (
	"strings"
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

func TestGeneratePsqlUseScript(t *testing.T) {
	script := generatePsqlUseScript(
		"/opt/postgresql/18.4/lib:/opt/postgresql/18.4/lib/postgresql/18/lib",
		"/opt/postgresql/18.4/bin",
		"127.0.0.1", 6133, "rsandbox", "rsandbox", "rsandbox",
	)
	if !strings.Contains(script, "LD_LIBRARY_PATH") {
		t.Error("missing LD_LIBRARY_PATH")
	}
	if !strings.Contains(script, "unset PGDATA") {
		t.Error("missing PGDATA unset")
	}
	if !strings.Contains(script, "/opt/postgresql/18.4/bin/psql") {
		t.Error("missing bundled psql path")
	}
	if !strings.Contains(script, "PGPASSWORD=rsandbox") {
		t.Error("missing PGPASSWORD")
	}
	if !strings.Contains(script, "-p 6133") {
		t.Error("missing port")
	}
	if !strings.Contains(script, "-d rsandbox") {
		t.Error("missing database")
	}
}
