package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

func (p *PostgreSQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	basedir, err := p.resolveBasedir(config)
	if err != nil {
		return nil, err
	}
	binDir := filepath.Join(basedir, "bin")
	libDir := filepath.Join(basedir, "lib")
	dataDir := filepath.Join(config.Dir, "data")
	logFile := filepath.Join(config.Dir, "postgresql.log")

	// pg_basebackup from the running primary
	pgBasebackup := filepath.Join(binDir, "pg_basebackup")
	bbCmd := exec.Command(pgBasebackup,
		"-h", "127.0.0.1",
		"-p", fmt.Sprintf("%d", primary.Port),
		"-U", "postgres",
		"-D", dataDir,
		"-Fp", "-Xs", "-R",
	)
	bbCmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	if output, err := bbCmd.CombinedOutput(); err != nil {
		os.RemoveAll(config.Dir) // cleanup on failure
		return nil, fmt.Errorf("pg_basebackup failed: %s: %w", string(output), err)
	}

	// Modify replica's postgresql.conf: update port and unix_socket_directories
	confPath := filepath.Join(dataDir, "postgresql.conf")
	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("reading postgresql.conf: %w", err)
	}

	conf := string(confBytes)
	lines := strings.Split(conf, "\n")
	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "port =") || strings.HasPrefix(trimmed, "port=") {
			newLines = append(newLines, fmt.Sprintf("port = %d", config.Port))
		} else if strings.HasPrefix(trimmed, "unix_socket_directories =") || strings.HasPrefix(trimmed, "unix_socket_directories=") {
			newLines = append(newLines, fmt.Sprintf("unix_socket_directories = '%s'", dataDir))
		} else {
			newLines = append(newLines, line)
		}
	}

	if err := os.WriteFile(confPath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing modified postgresql.conf: %w", err)
	}

	// Write lifecycle scripts
	scripts := GenerateScripts(ScriptOptions{
		SandboxDir: config.Dir,
		DataDir:    dataDir,
		BinDir:     binDir,
		LibDir:     libDir,
		Port:       config.Port,
		LogFile:    logFile,
	})
	for name, content := range scripts {
		scriptPath := filepath.Join(config.Dir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			os.RemoveAll(config.Dir)
			return nil, fmt.Errorf("writing script %s: %w", name, err)
		}
	}

	// Start the replica
	if err := p.StartSandbox(config.Dir); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("starting replica: %w", err)
	}

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   config.Port,
		Status: "running",
	}, nil
}
