package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ProxySQL/dbdeployer/providers"
)

func (p *PostgreSQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	basedir, err := p.resolveBasedir(config)
	if err != nil {
		return nil, err
	}
	binDir := filepath.Join(basedir, "bin")
	libDir := filepath.Join(basedir, "lib")
	dataDir := filepath.Join(config.Dir, "data")
	logDir := filepath.Join(dataDir, "log")
	logFile := filepath.Join(config.Dir, "postgresql.log")

	replication := config.Options["replication"] == "true"

	// Run initdb (data dir must not exist or must be empty)
	// Use -L to point to our extracted share directory. Deb-packaged initdb
	// looks for share data relative to its compiled --prefix (/usr), which
	// won't match our extracted layout at ~/opt/postgresql/<version>/share/.
	shareDir := filepath.Join(basedir, "share")
	initdbPath := filepath.Join(binDir, "initdb")
	initCmd := exec.Command(initdbPath, "-D", dataDir, "--auth=trust", "--username=postgres", "-L", shareDir)
	initCmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir))
	if output, err := initCmd.CombinedOutput(); err != nil {
		os.RemoveAll(config.Dir) // cleanup on failure
		return nil, fmt.Errorf("initdb failed: %s: %w", string(output), err)
	}

	// Create log directory (after initdb, which requires empty data dir)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Generate and write postgresql.conf
	pgConf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:            config.Port,
		ListenAddresses: "127.0.0.1",
		UnixSocketDir:   dataDir,
		LogDir:          logDir,
		Replication:     replication,
	})
	confPath := filepath.Join(dataDir, "postgresql.conf")
	if err := os.WriteFile(confPath, []byte(pgConf), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing postgresql.conf: %w", err)
	}

	// Generate and write pg_hba.conf
	hbaConf := GeneratePgHbaConf(replication)
	hbaPath := filepath.Join(dataDir, "pg_hba.conf")
	if err := os.WriteFile(hbaPath, []byte(hbaConf), 0644); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing pg_hba.conf: %w", err)
	}

	// Generate and write lifecycle scripts
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

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   config.Port,
		Status: "stopped",
	}, nil
}

// resolveBasedir determines the PostgreSQL base directory.
func (p *PostgreSQLProvider) resolveBasedir(config providers.SandboxConfig) (string, error) {
	if bd, ok := config.Options["basedir"]; ok && bd != "" {
		return bd, nil
	}
	return basedirFromVersion(config.Version)
}
