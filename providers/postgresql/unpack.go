package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var debVersionRegex = regexp.MustCompile(`^postgresql(?:-client)?-(\d+)_(\d+\.\d+)`)

func ParseDebVersion(filename string) (string, error) {
	base := filepath.Base(filename)
	matches := debVersionRegex.FindStringSubmatch(base)
	if matches == nil {
		return "", fmt.Errorf("cannot parse PostgreSQL version from %q (expected postgresql[-client]-NN_X.Y-*)", base)
	}
	return matches[2], nil
}

func ClassifyDebs(files []string) (server, client string, err error) {
	for _, f := range files {
		base := filepath.Base(f)
		if strings.HasPrefix(base, "postgresql-client-") {
			client = f
		} else if strings.HasPrefix(base, "postgresql-") && strings.HasSuffix(base, ".deb") {
			server = f
		}
	}
	if server == "" {
		return "", "", fmt.Errorf("no server deb found (expected postgresql-NN_*.deb)")
	}
	if client == "" {
		return "", "", fmt.Errorf("no client deb found (expected postgresql-client-NN_*.deb)")
	}
	return server, client, nil
}

func RequiredBinaries() []string {
	return []string{"postgres", "initdb", "pg_ctl", "psql", "pg_basebackup"}
}

func UnpackDebs(serverDeb, clientDeb, targetDir string) error {
	tmpDir, err := os.MkdirTemp("", "dbdeployer-pg-unpack-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, deb := range []string{serverDeb, clientDeb} {
		cmd := exec.Command("dpkg-deb", "-x", deb, tmpDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("extracting %s: %s: %w", filepath.Base(deb), string(output), err)
		}
	}

	version, err := ParseDebVersion(serverDeb)
	if err != nil {
		return err
	}
	major := strings.Split(version, ".")[0]

	srcBin := filepath.Join(tmpDir, "usr", "lib", "postgresql", major, "bin")
	srcLib := filepath.Join(tmpDir, "usr", "lib", "postgresql", major, "lib")
	srcShare := filepath.Join(tmpDir, "usr", "share", "postgresql", major)

	dstBin := filepath.Join(targetDir, "bin")
	dstLib := filepath.Join(targetDir, "lib")
	dstShare := filepath.Join(targetDir, "share")

	for _, dir := range []string{dstBin, dstLib, dstShare} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	copies := []struct{ src, dst string }{
		{srcBin, dstBin},
		{srcLib, dstLib},
		{srcShare, dstShare},
	}
	for _, c := range copies {
		if _, err := os.Stat(c.src); os.IsNotExist(err) {
			continue
		}
		cmd := exec.Command("cp", "-a", c.src+"/.", c.dst+"/") //nolint:gosec // paths are from controlled deb extraction
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("copying %s to %s: %s: %w", c.src, c.dst, string(output), err)
		}
	}

	// The postgres binary resolves share data relative to its own binary as
	// ../share/postgresql/<major>/ (compiled-in prefix from deb packaging).
	// Copy share files there too so both initdb (-L share/) and the postgres
	// server binary can find timezonesets and other share data.
	pgShareCompat := filepath.Join(dstShare, "postgresql", major)
	if err := os.MkdirAll(pgShareCompat, 0755); err != nil {
		return fmt.Errorf("creating compat share dir: %w", err)
	}
	compatCmd := exec.Command("cp", "-a", srcShare+"/.", pgShareCompat+"/") //nolint:gosec // paths are from controlled deb extraction
	if output, err := compatCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copying share to compat path: %s: %w", string(output), err)
	}

	for _, bin := range RequiredBinaries() {
		binPath := filepath.Join(dstBin, bin)
		if _, err := os.Stat(binPath); err != nil {
			return fmt.Errorf("required binary %q not found at %s after extraction", bin, binPath)
		}
	}

	return nil
}
