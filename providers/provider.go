package providers

import (
	"errors"
	"fmt"
	"sort"
)

var ErrNotSupported = errors.New("operation not supported by this provider")

// SandboxConfig holds provider-agnostic sandbox configuration.
type SandboxConfig struct {
	Version    string
	Dir        string            // sandbox directory path
	Port       int               // primary port
	AdminPort  int               // admin/management port (0 if not applicable)
	Host       string            // bind address
	DbUser     string            // admin username
	DbPassword string            // admin password
	Options    map[string]string // provider-specific key-value options
}

// SandboxInfo describes a deployed sandbox instance.
type SandboxInfo struct {
	Dir    string
	Port   int
	Socket string
	Status string // "running", "stopped"
}

// Provider is the core abstraction for deploying database infrastructure.
type Provider interface {
	Name() string
	ValidateVersion(version string) error
	DefaultPorts() PortRange
	// FindBinary returns the path to the provider's main binary, or error if not found.
	FindBinary(version string) (string, error)
	// CreateSandbox deploys a new sandbox instance.
	CreateSandbox(config SandboxConfig) (*SandboxInfo, error)
	// StartSandbox starts a stopped sandbox.
	StartSandbox(dir string) error
	// StopSandbox stops a running sandbox.
	StopSandbox(dir string) error
	// SupportedTopologies returns the list of topology names this provider supports.
	SupportedTopologies() []string
	// CreateReplica creates a replica sandbox joined to an existing primary.
	CreateReplica(primary SandboxInfo, config SandboxConfig) (*SandboxInfo, error)
}

type PortRange struct {
	BasePort         int
	PortsPerInstance int
}

type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) error {
	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	r.providers[name] = p
	return nil
}

func (r *Registry) Get(name string) (Provider, error) {
	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var DefaultRegistry = NewRegistry()

// ContainsString checks if a string slice contains a given value.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// CompatibleAddons maps addon names to the list of providers they work with.
var CompatibleAddons = map[string][]string{
	"proxysql": {"mysql", "postgresql"},
}
