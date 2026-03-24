package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "postgresql"

type PostgreSQLProvider struct{}

func NewPostgreSQLProvider() *PostgreSQLProvider { return &PostgreSQLProvider{} }

func (p *PostgreSQLProvider) Name() string { return ProviderName }

func (p *PostgreSQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid PostgreSQL version format: %q (expected major.minor, e.g. 16.13)", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid PostgreSQL major version %q: %w", parts[0], err)
	}
	if major < 12 {
		return fmt.Errorf("PostgreSQL major version must be >= 12, got %d", major)
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("invalid PostgreSQL minor version %q: %w", parts[1], err)
	}
	return nil
}

func (p *PostgreSQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{BasePort: 15000, PortsPerInstance: 1}
}

func (p *PostgreSQLProvider) SupportedTopologies() []string {
	return []string{"single", "multiple", "replication"}
}

// VersionToPort converts a PostgreSQL version to a port number.
// Formula: BasePort + major*100 + minor
func VersionToPort(version string) (int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid version format: %q", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return 15000 + major*100 + minor, nil
}

func (p *PostgreSQLProvider) FindBinary(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	binPath := filepath.Join(home, "opt", "postgresql", version, "bin", "postgres")
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("PostgreSQL binary not found at %s: %w", binPath, err)
	}
	return binPath, nil
}

func basedirFromVersion(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, "opt", "postgresql", version), nil
}

func (p *PostgreSQLProvider) StartSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "start"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *PostgreSQLProvider) StopSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "stop"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *PostgreSQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, fmt.Errorf("PostgreSQLProvider.CreateReplica: not yet implemented")
}

func Register(reg *providers.Registry) error {
	return reg.Register(NewPostgreSQLProvider())
}
