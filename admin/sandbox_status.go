package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/defaults"
)

// SandboxStatus represents a sandbox with its current running status.
type SandboxStatus struct {
	Name     string        `json:"name"`
	Dir      string        `json:"dir"`
	SBType   string        `json:"type"`
	Version  string        `json:"version"`
	Flavor   string        `json:"flavor,omitempty"`
	Ports    []int         `json:"ports"`
	Nodes    []SandboxNode `json:"nodes,omitempty"`
	Running  bool          `json:"running"`
	Provider string        `json:"provider,omitempty"`
}

// SandboxNode represents a node within a multi-node sandbox.
type SandboxNode struct {
	Name    string `json:"name"`
	Dir     string `json:"dir"`
	Port    int    `json:"port"`
	Running bool   `json:"running"`
}

// GetAllSandboxes reads the sandbox catalog and checks running status for each sandbox.
func GetAllSandboxes() ([]SandboxStatus, error) {
	catalog, err := defaults.ReadCatalog()
	if err != nil {
		return nil, fmt.Errorf("reading sandbox catalog: %w", err)
	}

	sandboxHome := defaults.Defaults().SandboxHome

	var result []SandboxStatus
	for name, item := range catalog {
		// The catalog key is the full path (Destination), use the base name as the display name.
		sbName := name
		sbDir := item.Destination
		if sbDir == "" {
			sbDir = filepath.Join(sandboxHome, name)
		}
		// If the name looks like a full path, extract just the base name.
		if strings.Contains(sbName, "/") {
			sbName = filepath.Base(sbName)
		}

		status := SandboxStatus{
			Name:    sbName,
			Dir:     sbDir,
			SBType:  item.SBType,
			Version: item.Version,
			Flavor:  item.Flavor,
			Ports:   item.Port,
		}

		// Determine provider from flavor.
		switch strings.ToLower(item.Flavor) {
		case "postgresql", "postgres":
			status.Provider = "postgresql"
		case "tidb":
			status.Provider = "tidb"
		case "ndb":
			status.Provider = "ndb"
		case "pxc":
			status.Provider = "pxc"
		case "mariadb":
			status.Provider = "mariadb"
		default:
			status.Provider = "mysql"
		}

		isMulti := len(item.Nodes) > 0

		if isMulti {
			// Multi-node sandbox: check each node.
			for _, nodeRelPath := range item.Nodes {
				nodeDir := filepath.Join(sbDir, nodeRelPath)
				nodeName := filepath.Base(nodeRelPath)

				// Try to get node port from sandbox_description in node dir.
				nodePort := 0
				if desc, err := common.ReadSandboxDescription(nodeDir); err == nil {
					if len(desc.Port) > 0 {
						nodePort = desc.Port[0]
					}
				}

				nodeRunning := isRunning(nodeDir)
				status.Nodes = append(status.Nodes, SandboxNode{
					Name:    nodeName,
					Dir:     nodeDir,
					Port:    nodePort,
					Running: nodeRunning,
				})
			}
			// Overall running = at least one node running.
			for _, n := range status.Nodes {
				if n.Running {
					status.Running = true
					break
				}
			}
		} else {
			status.Running = isRunning(sbDir)
		}

		result = append(result, status)
	}

	return result, nil
}

// isRunning checks whether a sandbox is running by looking for a live PID file.
func isRunning(sandboxDir string) bool {
	pidFile := filepath.Join(sandboxDir, "data", "mysql.pid")
	if _, err := os.Stat(pidFile); err != nil {
		// Try alternate pid file names.
		pidFile = filepath.Join(sandboxDir, "data", "mysqld.pid")
		if _, err := os.Stat(pidFile); err != nil {
			return false
		}
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		return false
	}

	// Check if the process exists by sending signal 0.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
