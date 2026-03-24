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
		AdminHost:     host,
		AdminPort:     adminPort,
		AdminUser:     adminUser,
		AdminPassword: adminPassword,
		MySQLPort:     mysqlPort,
		DataDir:       dataDir,
		MonitorUser:   monitorUser,
		MonitorPass:   monitorPass,
		Backends:      parseBackends(config.Options),
	}

	cfgContent := GenerateConfig(proxyCfg)
	cfgPath := filepath.Join(config.Dir, "proxysql.cnf")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	// Write lifecycle scripts
	scripts := map[string]string{
		"start": fmt.Sprintf("#!/bin/bash\ncd %s\n%s --initial -c %s -D %s &\nSBPID=$!\necho $SBPID > %s/proxysql.pid\nsleep 2\nif kill -0 $SBPID 2>/dev/null; then\n  echo 'ProxySQL started (pid '$SBPID')'\nelse\n  echo 'ProxySQL failed to start'\n  exit 1\nfi\n",
			config.Dir, binaryPath, cfgPath, dataDir, config.Dir),
		"stop": fmt.Sprintf("#!/bin/bash\nPIDFILE=%s/proxysql.pid\nif [ -f $PIDFILE ]; then\n  PID=$(cat $PIDFILE)\n  kill $PID 2>/dev/null\n  sleep 1\n  kill -0 $PID 2>/dev/null && kill -9 $PID 2>/dev/null\n  rm -f $PIDFILE\n  echo 'ProxySQL stopped'\nelse\n  echo 'ProxySQL not running (no pid file)'\nfi\n",
			config.Dir),
		"status": fmt.Sprintf("#!/bin/bash\nPIDFILE=%s/proxysql.pid\nif [ -f $PIDFILE ] && kill -0 $(cat $PIDFILE) 2>/dev/null; then\n  echo 'ProxySQL running (pid '$(cat $PIDFILE)')'\nelse\n  echo 'ProxySQL not running'\n  exit 1\nfi\n",
			config.Dir),
		"use": fmt.Sprintf("#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL Admin> ' \"$@\"\n",
			host, adminPort, adminUser, adminPassword),
		"use_proxy": fmt.Sprintf("#!/bin/bash\nmysql -h %s -P %d -u %s -p%s --prompt 'ProxySQL> ' \"$@\"\n",
			host, mysqlPort, monitorUser, monitorPass),
	}

	for name, content := range scripts {
		scriptPath := filepath.Join(config.Dir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
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
	cmd := exec.Command("bash", filepath.Join(dir, "start"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("start failed: %s: %w", string(output), err)
	}
	return nil
}

func (p *ProxySQLProvider) StopSandbox(dir string) error {
	cmd := exec.Command("bash", filepath.Join(dir, "stop"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop failed: %s: %w", string(output), err)
	}
	return nil
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
