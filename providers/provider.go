package providers

import (
	"fmt"
	"sort"
)

// Provider is the core abstraction for deploying database infrastructure.
// Phase 2a defines a minimal interface (Name, ValidateVersion, DefaultPorts).
// Phase 2b will add CreateSandbox, Start, Stop, Destroy, HealthCheck when
// ProxySQL and other providers need them.
type Provider interface {
	Name() string
	ValidateVersion(version string) error
	DefaultPorts() PortRange
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
