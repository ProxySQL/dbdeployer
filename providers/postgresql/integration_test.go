//go:build integration

package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProxySQL/dbdeployer/providers"
)

func findPostgresVersion(t *testing.T) string {
	t.Helper()
	home, _ := os.UserHomeDir()
	entries, err := os.ReadDir(filepath.Join(home, "opt", "postgresql"))
	if err != nil {
		t.Skipf("no PostgreSQL installations found: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return e.Name()
		}
	}
	t.Skip("no PostgreSQL version directories found")
	return ""
}

func TestIntegrationSingleSandbox(t *testing.T) {
	version := findPostgresVersion(t)
	p := NewPostgreSQLProvider()

	tmpDir := t.TempDir()
	sandboxDir := filepath.Join(tmpDir, "pg_test")

	config := providers.SandboxConfig{
		Version: version,
		Dir:     sandboxDir,
		Port:    15432,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{},
	}

	// Create
	info, err := p.CreateSandbox(config)
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}
	if info.Port != 15432 {
		t.Errorf("expected port 15432, got %d", info.Port)
	}

	// Start
	if err := p.StartSandbox(sandboxDir); err != nil {
		t.Fatalf("StartSandbox failed: %v", err)
	}
	stopped := false
	defer func() {
		if !stopped {
			p.StopSandbox(sandboxDir)
		}
	}()
	time.Sleep(2 * time.Second)

	// Connect via psql
	home, _ := os.UserHomeDir()
	psql := filepath.Join(home, "opt", "postgresql", version, "bin", "psql")
	cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", "15432", "-U", "postgres", "-c", "SELECT 1;")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s",
		filepath.Join(home, "opt", "postgresql", version, "lib")))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("psql connection failed: %s: %v", string(output), err)
	}

	// Stop
	if err := p.StopSandbox(sandboxDir); err != nil {
		t.Fatalf("StopSandbox failed: %v", err)
	}
	stopped = true
}

func TestIntegrationReplication(t *testing.T) {
	version := findPostgresVersion(t)
	p := NewPostgreSQLProvider()

	tmpDir := t.TempDir()
	primaryDir := filepath.Join(tmpDir, "primary")
	replica1Dir := filepath.Join(tmpDir, "replica1")
	replica2Dir := filepath.Join(tmpDir, "replica2")

	// Create and start primary with replication
	primaryConfig := providers.SandboxConfig{
		Version: version,
		Dir:     primaryDir,
		Port:    15500,
		Host:    "127.0.0.1",
		DbUser:  "postgres",
		Options: map[string]string{"replication": "true"},
	}

	_, err := p.CreateSandbox(primaryConfig)
	if err != nil {
		t.Fatalf("CreateSandbox (primary) failed: %v", err)
	}
	if err := p.StartSandbox(primaryDir); err != nil {
		t.Fatalf("StartSandbox (primary) failed: %v", err)
	}
	defer p.StopSandbox(primaryDir)
	time.Sleep(2 * time.Second)

	primaryInfo := providers.SandboxInfo{Dir: primaryDir, Port: 15500}

	// Create replicas
	for i, rDir := range []string{replica1Dir, replica2Dir} {
		rConfig := providers.SandboxConfig{
			Version: version,
			Dir:     rDir,
			Port:    15501 + i,
			Host:    "127.0.0.1",
			DbUser:  "postgres",
			Options: map[string]string{},
		}
		_, err := p.CreateReplica(primaryInfo, rConfig)
		if err != nil {
			t.Fatalf("CreateReplica %d failed: %v", i+1, err)
		}
		defer p.StopSandbox(rDir)
	}

	time.Sleep(2 * time.Second)

	// Verify replication on primary
	home, _ := os.UserHomeDir()
	psql := filepath.Join(home, "opt", "postgresql", version, "bin", "psql")
	libDir := filepath.Join(home, "opt", "postgresql", version, "lib")

	cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", "15500", "-U", "postgres", "-t", "-c",
		"SELECT count(*) FROM pg_stat_replication;")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("replication check failed: %s: %v", string(output), err)
	}

	// Verify replicas are in recovery
	for _, port := range []int{15501, 15502} {
		cmd := exec.Command(psql, "-h", "127.0.0.1", "-p", fmt.Sprintf("%d", port), "-U", "postgres", "-t", "-c",
			"SELECT pg_is_in_recovery();")
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("recovery check on port %d failed: %s: %v", port, string(output), err)
		}
	}
}
