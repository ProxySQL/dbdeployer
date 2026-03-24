package providers

import "testing"

type mockProvider struct{ name string }

func (m *mockProvider) Name() string                         { return m.name }
func (m *mockProvider) ValidateVersion(version string) error { return nil }
func (m *mockProvider) DefaultPorts() PortRange              { return PortRange{BasePort: 9999, PortsPerInstance: 1} }
func (m *mockProvider) FindBinary(version string) (string, error) { return "/usr/bin/mock", nil }
func (m *mockProvider) CreateSandbox(config SandboxConfig) (*SandboxInfo, error) { return &SandboxInfo{Dir: "/tmp/mock"}, nil }
func (m *mockProvider) StartSandbox(dir string) error { return nil }
func (m *mockProvider) StopSandbox(dir string) error  { return nil }
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
	if err := reg.Register(mock); err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()
	if _, err := reg.Get("nonexistent"); err == nil {
		t.Fatal("expected error on missing provider")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&mockProvider{name: "b"})
	_ = reg.Register(&mockProvider{name: "a"})
	names := reg.List()
	if len(names) != 2 {
		t.Errorf("expected 2 providers, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("expected sorted [a b], got %v", names)
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"single", "multiple", "replication"}
	if !ContainsString(slice, "single") {
		t.Error("expected single to be found")
	}
	if ContainsString(slice, "group") {
		t.Error("did not expect group to be found")
	}
	if ContainsString(nil, "single") {
		t.Error("did not expect match in nil slice")
	}
}

func TestCompatibleAddons(t *testing.T) {
	if !ContainsString(CompatibleAddons["proxysql"], "mysql") {
		t.Error("proxysql should be compatible with mysql")
	}
	if !ContainsString(CompatibleAddons["proxysql"], "postgresql") {
		t.Error("proxysql should be compatible with postgresql")
	}
	if ContainsString(CompatibleAddons["proxysql"], "fake") {
		t.Error("proxysql should not be compatible with fake")
	}
}

func TestSupportedTopologiesMySQL(t *testing.T) {
	reg := NewRegistry()
	mock := &mockProvider{name: "mysql-like"}
	_ = reg.Register(mock)
	p, _ := reg.Get("mysql-like")
	topos := p.SupportedTopologies()
	if !ContainsString(topos, "single") {
		t.Error("expected single in topologies")
	}
	if ContainsString(topos, "group") {
		t.Error("mock should not support group")
	}
}
