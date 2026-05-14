package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

	if err := checkLinuxLayout(basedir, binDir); err != nil {
		return nil, err
	}

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
	if err := os.WriteFile(confPath, []byte(pgConf), 0600); err != nil {
		os.RemoveAll(config.Dir)
		return nil, fmt.Errorf("writing postgresql.conf: %w", err)
	}

	// Generate and write pg_hba.conf
	hbaConf := GeneratePgHbaConf(replication)
	hbaPath := filepath.Join(dataDir, "pg_hba.conf")
	if err := os.WriteFile(hbaPath, []byte(hbaConf), 0600); err != nil {
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
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil { //nolint:gosec // scripts must be executable
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

// checkLinuxLayout detects PostgreSQL extractions produced by older
// versions of dbdeployer (before issue #112 was fixed). In the old layout,
// <basedir>/bin/<binary> were regular files; PG's make_relative_path() then
// failed to relocate the compiled-in SHAREDIR (`/usr/share/postgresql/<major>`),
// so initdb died with "could not open directory /usr/share/postgresql/<major>/timezonesets".
//
// New extractions place the real binaries under
// <basedir>/lib/postgresql/<major>/bin/ and expose them via symlinks at
// <basedir>/bin/. If we find a regular file there on Linux, the user
// needs to re-run unpack.
func checkLinuxLayout(basedir, binDir string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	fi, err := os.Lstat(filepath.Join(binDir, "postgres"))
	if err != nil {
		// Missing binary will be reported by the initdb step below with a
		// clearer error than we could produce here.
		return nil
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	return fmt.Errorf(
		"PostgreSQL binaries at %s were unpacked by an older dbdeployer using\n"+
			"a layout incompatible with deb-packaged PostgreSQL — initdb would fail\n"+
			"to find share files (see issue #112). Re-run unpack to fix:\n"+
			"    dbdeployer unpack --provider=postgresql <server.deb> <client.deb>",
		basedir,
	)
}
