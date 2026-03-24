package mysql

import (
	"fmt"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "mysql"

type MySQLProvider struct{}

func NewMySQLProvider() *MySQLProvider {
	return &MySQLProvider{}
}

func (p *MySQLProvider) Name() string { return ProviderName }

func (p *MySQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid MySQL version format: %q (expected X.Y or X.Y.Z)", version)
	}
	return nil
}

func (p *MySQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{
		BasePort:         3306,
		PortsPerInstance: 3, // main + mysqlx + admin
	}
}

func (p *MySQLProvider) FindBinary(version string) (string, error) {
	return "", fmt.Errorf("MySQLProvider.FindBinary: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, fmt.Errorf("MySQLProvider.CreateSandbox: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) StartSandbox(dir string) error {
	return fmt.Errorf("MySQLProvider.StartSandbox: use sandbox package directly (not yet migrated)")
}

func (p *MySQLProvider) StopSandbox(dir string) error {
	return fmt.Errorf("MySQLProvider.StopSandbox: use sandbox package directly (not yet migrated)")
}

func Register(reg *providers.Registry) error {
	return reg.Register(NewMySQLProvider())
}
