package admin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ProxySQL/dbdeployer/defaults"
)

// ExecuteSandboxScript runs a lifecycle script (start, stop, restart) in a sandbox directory.
// It tries <script> first, then <script>_all for multi-node sandboxes.
func ExecuteSandboxScript(sandboxName string, script string) error {
	sbDir := sandboxDirFor(sandboxName)

	scriptPath := filepath.Join(sbDir, script)
	if _, err := os.Stat(scriptPath); err != nil {
		// Try the _all variant for multi-node sandboxes.
		scriptPath = filepath.Join(sbDir, script+"_all")
		if _, err2 := os.Stat(scriptPath); err2 != nil {
			return fmt.Errorf("script %q not found in %s", script, sbDir)
		}
	}

	cmd := exec.Command("bash", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %s: %w", script, string(output), err)
	}
	return nil
}

// DestroySandbox stops a sandbox and removes its directory and catalog entry.
func DestroySandbox(sandboxName string) error {
	sbDir := sandboxDirFor(sandboxName)

	// Try stop (log warning if fails — sandbox may already be stopped).
	if err := ExecuteSandboxScript(sandboxName, "stop"); err != nil {
		fmt.Printf("Warning: failed to stop sandbox %s during destruction: %v\n", sandboxName, err)
	}

	// Remove directory.
	if err := os.RemoveAll(sbDir); err != nil {
		return fmt.Errorf("removing %s: %w", sbDir, err)
	}

	// Remove from catalog. The catalog key is the full destination path.
	if err := defaults.DeleteFromCatalog(sbDir); err != nil {
		return fmt.Errorf("removing %s from catalog: %w", sandboxName, err)
	}
	return nil
}

// sandboxDirFor resolves the full path for a sandbox name.
func sandboxDirFor(sandboxName string) string {
	return filepath.Join(defaults.Defaults().SandboxHome, sandboxName)
}
