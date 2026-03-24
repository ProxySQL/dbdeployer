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
