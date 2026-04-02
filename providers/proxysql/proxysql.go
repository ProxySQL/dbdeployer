package proxysql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ProxySQL/dbdeployer/providers"
)

const ProviderName = "proxysql"

type ProxySQLProvider struct{}

func NewProxySQLProvider() *ProxySQLProvider { return &ProxySQLProvider{} }

func (p *ProxySQLProvider) Name() string { return ProviderName }

func (p *ProxySQLProvider) ValidateVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid ProxySQL version format: %q (expected X.Y or X.Y.Z)", version)
	}
	return nil
}

func (p *ProxySQLProvider) DefaultPorts() providers.PortRange {
	return providers.PortRange{BasePort: 6032, PortsPerInstance: 2}
}

func (p *ProxySQLProvider) FindBinary(version string) (string, error) {
	path, err := exec.LookPath("proxysql")
	if err != nil {
		return "", fmt.Errorf("proxysql binary not found in PATH: %w", err)
	}
	return path, nil
}

func (p *ProxySQLProvider) CreateSandbox(config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	binaryPath, err := p.FindBinary(config.Version)
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(config.Dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	adminPort := config.AdminPort
	if adminPort == 0 {
		adminPort = config.Port
	}
	mysqlPort := adminPort + 1

	adminUser := config.DbUser
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPassword := config.DbPassword
	if adminPassword == "" {
		adminPassword = "admin"
	}

	monitorUser := config.Options["monitor_user"]
	if monitorUser == "" {
		monitorUser = "msandbox"
	}
	monitorPass := config.Options["monitor_password"]
	if monitorPass == "" {
		monitorPass = "msandbox"
	}

	host := config.Host
	if host == "" {
		host = "127.0.0.1"
	}

	proxyCfg := ProxySQLConfig{
		AdminHost:       host,
		AdminPort:       adminPort,
		AdminUser:       adminUser,
		AdminPassword:   adminPassword,
		MySQLPort:       mysqlPort,
		DataDir:         dataDir,
		MonitorUser:     monitorUser,
		MonitorPass:     monitorPass,
		Backends:        parseBackends(config.Options),
		BackendProvider: config.Options["backend_provider"],
		Topology:        config.Options["topology"],
	}

	cfgContent := GenerateConfig(proxyCfg)
	cfgPath := filepath.Join(config.Dir, "proxysql.cnf")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	// Write lifecycle scripts
	// ProxySQL daemonizes by default — it forks and the parent exits.
	// We use ProxySQL's own PID file (written to datadir/proxysql.pid) to track the process.
	pidFile := filepath.Join(dataDir, "proxysql.pid")
	scripts := map[string]string{
		"start": fmt.Sprintf("#!/bin/bash\ncd %s\n%s --initial -c %s -D %s 2>&1 | grep -v '^profiling:' || true\n# ProxySQL daemonizes — wait for PID file\nfor i in $(seq 1 10); do\n  if [ -f %s ]; then\n    PID=$(cat %s)\n    if kill -0 $PID 2>/dev/null; then\n      sleep 2\n      echo \"ProxySQL started (pid $PID)\"\n      exit 0\n    fi\n  fi\n  sleep 1\ndone\necho 'ProxySQL failed to start'\nexit 1\n",
			config.Dir, binaryPath, cfgPath, dataDir, pidFile, pidFile),
		"stop": fmt.Sprintf("#!/bin/bash\nPIDFILE=%s\nCONFIG=%s\nif [ -f $PIDFILE ]; then\n  PID=$(cat $PIDFILE)\n  # Kill main process and any children matching our config\n  kill $PID 2>/dev/null\n  pkill -f \"proxysql.*$CONFIG\" 2>/dev/null\n  for i in $(seq 1 5); do\n    kill -0 $PID 2>/dev/null || break\n    sleep 1\n  done\n  kill -0 $PID 2>/dev/null && kill -9 $PID 2>/dev/null\n  pkill -9 -f \"proxysql.*$CONFIG\" 2>/dev/null\n  rm -f $PIDFILE\n  echo 'ProxySQL stopped'\nelse\n  # Try to find and kill by config file pattern\n  pkill -f \"proxysql.*$CONFIG\" 2>/dev/null\n  echo 'ProxySQL stopped (no pid file)'\nfi\n",
			pidFile, cfgPath),
		"status": fmt.Sprintf("#!/bin/bash\nPIDFILE=%s\nif [ -f $PIDFILE ] && kill -0 $(cat $PIDFILE) 2>/dev/null; then\n  echo \"ProxySQL running (pid $(cat $PIDFILE))\"\nelse\n  echo 'ProxySQL not running'\n  exit 1\nfi\n",
			pidFile),
		"use": fmt.Sprintf("#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL Admin> ' \"$@\"\n",
			host, adminPort, adminUser, adminPassword),
	}
	if config.Options["backend_provider"] == "postgresql" {
		scripts["use_proxy"] = fmt.Sprintf("#!/bin/bash\npsql -h %s -p %d -U %s \"$@\"\n",
			host, mysqlPort, monitorUser)
	} else {
		scripts["use_proxy"] = fmt.Sprintf("#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL> ' \"$@\"\n",
			host, mysqlPort, monitorUser, monitorPass)
	}

	for name, content := range scripts {
		scriptPath := filepath.Join(config.Dir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil { //nolint:gosec // scripts must be executable
			return nil, fmt.Errorf("writing script %s: %w", name, err)
		}
	}

	return &providers.SandboxInfo{
		Dir:    config.Dir,
		Port:   adminPort,
		Status: "stopped",
	}, nil
}

func (p *ProxySQLProvider) StartSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "start")) //nolint:gosec // path is from trusted sandbox directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *ProxySQLProvider) StopSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "stop")) //nolint:gosec // path is from trusted sandbox directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *ProxySQLProvider) SupportedTopologies() []string {
	return []string{"single"}
}

func (p *ProxySQLProvider) CreateReplica(primary providers.SandboxInfo, config providers.SandboxConfig) (*providers.SandboxInfo, error) {
	return nil, providers.ErrNotSupported
}

func Register(reg *providers.Registry) error {
	return reg.Register(NewProxySQLProvider())
}

func parseBackends(options map[string]string) []BackendServer {
	raw, ok := options["backends"]
	if !ok || raw == "" {
		return nil
	}
	var backends []BackendServer
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.Split(entry, ":")
		if len(parts) >= 3 {
			port, _ := strconv.Atoi(parts[1])
			hg, _ := strconv.Atoi(parts[2])
			backends = append(backends, BackendServer{
				Host: parts[0], Port: port, Hostgroup: hg, MaxConns: 200,
			})
		}
	}
	return backends
}
