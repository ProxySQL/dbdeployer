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

func RequiredBinaries() []string {
	return []string{"postgres", "initdb", "pg_ctl", "psql", "pg_basebackup"}
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

func findServerDeb(debs []string) (string, error) {
	for _, deb := range debs {
		base := filepath.Base(deb)
		if strings.HasPrefix(base, "postgresql-client-") {
			continue
		}
		if strings.HasPrefix(base, "postgresql-") && strings.HasSuffix(base, ".deb") {
			suffix := base[len("postgresql-"):]
			if len(suffix) > 0 && suffix[0] >= '0' && suffix[0] <= '9' {
				return deb, nil
			}
		}
	}
	return "", fmt.Errorf("no server deb found (expected postgresql-NN_*.deb)")
}

func UnpackDebs(debs []string, targetDir string) error {
	serverDeb, err := findServerDeb(debs)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "dbdeployer-pg-unpack-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, deb := range debs {
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
	srcPgLib := filepath.Join(tmpDir, "usr", "lib", "postgresql", major, "lib")
	srcShare := filepath.Join(tmpDir, "usr", "share", "postgresql", major)

	// Destination layout mirrors Debian's compile-time prefix so that PG's
	// make_relative_path() (src/port/path.c) succeeds at runtime. Debian
	// builds PG with BINDIR=/usr/lib/postgresql/<major>/bin and
	// SHAREDIR=/usr/share/postgresql/<major>; the common prefix is just
	// /usr/, so for PG to relocate SHAREDIR to <basedir>/share/postgresql/<major>
	// the running binary's parent directory must end in
	// "lib/postgresql/<major>/bin". Same logic applies to PKGLIBDIR.
	//
	// We therefore place the real binaries and extension libs under
	// <basedir>/lib/postgresql/<major>/, and expose them via symlinks at
	// <basedir>/bin/ so existing callers (sandbox.go, scripts.go) keep
	// working. On Linux, PG's find_my_exec() reads /proc/self/exe which
	// resolves symlinks to the real file, so launching via the symlink
	// still gives PG the relocatable path it needs.
	dstPgBin := filepath.Join(targetDir, "lib", "postgresql", major, "bin")
	dstPgLib := filepath.Join(targetDir, "lib", "postgresql", major, "lib")
	dstShare := filepath.Join(targetDir, "share")                       // flat, for `initdb -L`
	dstPgShare := filepath.Join(targetDir, "share", "postgresql", major) // nested, for postgres runtime
	dstBin := filepath.Join(targetDir, "bin")                           // symlinks into dstPgBin

	// Make UnpackDebs idempotent: remove any prior extraction artifacts
	// in the subtrees we own so re-running unpack picks up the latest
	// layout cleanly (in particular when migrating from a pre-fix layout
	// where <basedir>/bin/<binary> were real files instead of symlinks).
	for _, dir := range []string{dstBin, filepath.Join(targetDir, "lib"), dstShare} {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("removing stale %s: %w", dir, err)
		}
	}

	for _, dir := range []string{dstPgBin, dstPgLib, dstShare, dstPgShare, dstBin} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	copies := []struct{ src, dst string }{
		{srcBin, dstPgBin},
		{srcPgLib, dstPgLib},
		{srcShare, dstShare},   // flat copy for `initdb -L`
		{srcShare, dstPgShare}, // nested copy for postgres server runtime
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

	flatLib := filepath.Join(targetDir, "lib")
	if err := copyClientLibraries(tmpDir, flatLib); err != nil {
		return err
	}

	// Expose every server-side binary as <basedir>/bin/<name>, symlinked
	// into the nested lib/postgresql/<major>/bin/ directory.
	entries, err := os.ReadDir(dstPgBin)
	if err != nil {
		return fmt.Errorf("reading %s: %w", dstPgBin, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		linkTarget := filepath.Join("..", "lib", "postgresql", major, "bin", entry.Name())
		linkPath := filepath.Join(dstBin, entry.Name())
		if err := os.Symlink(linkTarget, linkPath); err != nil {
			return fmt.Errorf("symlinking %s -> %s: %w", linkPath, linkTarget, err)
		}
	}

	for _, bin := range RequiredBinaries() {
		binPath := filepath.Join(dstBin, bin)
		if _, err := os.Stat(binPath); err != nil {
			return fmt.Errorf("required binary %q not found at %s after extraction", bin, binPath)
		}
	}

	return nil
}

// copyClientLibraries copies libpq from extracted debs into <basedir>/lib.
// On Debian/Ubuntu libpq ships in the separate libpq5 package, not in
// postgresql-client itself.
func copyClientLibraries(tmpDir, dstLib string) error {
	if err := os.MkdirAll(dstLib, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dstLib, err)
	}

	copied := make(map[string]bool)
	err := filepath.WalkDir(tmpDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasPrefix(d.Name(), "libpq.so") {
			return nil
		}
		cmd := exec.Command("cp", "-a", path, dstLib+"/") //nolint:gosec // paths from controlled deb extraction
		if output, cpErr := cmd.CombinedOutput(); cpErr != nil {
			return fmt.Errorf("copying %s to %s: %s: %w", path, dstLib, string(output), cpErr)
		}
		copied[d.Name()] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(copied) == 0 {
		return fmt.Errorf(
			"libpq not found after extracting PostgreSQL debs.\n" +
				"Client tools need the libpq5 package — download it and pass it to unpack:\n" +
				"    apt-get download libpq5\n" +
				"    dbdeployer unpack --provider=postgresql postgresql-NN_*.deb postgresql-client-NN_*.deb libpq5_*.deb")
	}
	return nil
}
