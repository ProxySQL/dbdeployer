// Package postgresql — macOS Postgres.app installer.
//
// Unlike Linux, where PostgreSQL is distributed as two .deb files that we
// extract with dpkg-deb, macOS has no official PostgreSQL binary tarball.
// The closest to a "drop-in" binary distribution is Postgres.app, which
// bundles multiple PostgreSQL major versions inside a .dmg, each under
// Postgres.app/Contents/Versions/<major>/{bin,lib,share,include}.
//
// This file provides a helper that downloads a single-version Postgres.app
// .dmg from GitHub, mounts it headlessly with hdiutil, copies the
// bin/lib/share/include tree into the dbdeployer sandbox-binary directory,
// and detaches the image. No sudo, no GUI, no Homebrew required.
package postgresql

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// PostgresAppReleasesAPI is the GitHub API endpoint for the Postgres.app
// project. We read the latest release to discover the current version
// number and the .dmg asset URLs.
const PostgresAppReleasesAPI = "https://api.github.com/repos/PostgresApp/PostgresApp/releases/latest"

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []githubAsset `json:"assets"`
}

// PostgresAppAsset describes a single-major-version .dmg from Postgres.app.
type PostgresAppAsset struct {
	// AppVersion is the Postgres.app release tag, e.g. "2.9.4".
	AppVersion string
	// Major is the PostgreSQL major version, e.g. "16".
	Major string
	// URL is the direct download URL for the .dmg.
	URL string
	// Size is the .dmg size in bytes.
	Size int64
}

// LatestPostgresAppAssets fetches the latest Postgres.app release and
// returns one asset per PostgreSQL major version (the small single-version
// .dmg files, not the bundle). Assets are ordered by major version
// descending (newest first).
func LatestPostgresAppAssets() ([]PostgresAppAsset, error) {
	return parseAssetsFromURL(PostgresAppReleasesAPI)
}

// parseAssetsFromURL is factored out from LatestPostgresAppAssets so tests
// can point it at a local httptest server.
func parseAssetsFromURL(apiURL string) ([]PostgresAppAsset, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building GitHub request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dbdeployer")
	// Unauthenticated GitHub API is rate-limited to 60 req/hour per IP.
	// Shared CI runners routinely exhaust this. If GITHUB_TOKEN is set
	// in the environment (as it is on every GitHub Actions runner by
	// default), send it to get a 5000 req/hour limit.
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching Postgres.app releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned %d for %s", resp.StatusCode, apiURL)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decoding release JSON: %w", err)
	}
	appVersion := strings.TrimPrefix(rel.TagName, "v")

	// Single-version assets are named Postgres-<appver>-<major>.dmg,
	// distinguished from the multi-version bundle Postgres-<appver>-14-15-16-...dmg
	// by having exactly one numeric segment after the app version.
	var assets []PostgresAppAsset
	for _, a := range rel.Assets {
		if !strings.HasSuffix(a.Name, ".dmg") {
			continue
		}
		base := strings.TrimSuffix(a.Name, ".dmg")
		// e.g. "Postgres-2.9.4-16"
		prefix := "Postgres-" + appVersion + "-"
		if !strings.HasPrefix(base, prefix) {
			continue
		}
		tail := strings.TrimPrefix(base, prefix)
		if strings.Contains(tail, "-") {
			// multi-version bundle like "14-15-16-17-18" — skip
			continue
		}
		assets = append(assets, PostgresAppAsset{
			AppVersion: appVersion,
			Major:      tail,
			URL:        a.BrowserDownloadURL,
			Size:       a.Size,
		})
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("no single-version Postgres.app .dmg assets found in release %s", rel.TagName)
	}
	// Sort by major version descending (string sort works because all
	// majors are 2 digits).
	for i := 0; i < len(assets)-1; i++ {
		for j := i + 1; j < len(assets); j++ {
			if assets[j].Major > assets[i].Major {
				assets[i], assets[j] = assets[j], assets[i]
			}
		}
	}
	return assets, nil
}

// InstallFromPostgresAppDMG downloads the given Postgres.app .dmg, mounts
// it with hdiutil, queries the bundled postgres binary for its exact
// X.Y version, copies bin/lib/share/include into
// sandboxBinaryDir/<X.Y>/, and detaches.
//
// Returns the detected full version (e.g. "16.4") on success so callers
// can feed it to `dbdeployer deploy postgresql <version>`.
//
// Only supported on darwin. On other platforms, returns an error.
func InstallFromPostgresAppDMG(asset PostgresAppAsset, sandboxBinaryDir string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("Postgres.app installation is only supported on macOS (current GOOS=%s)", runtime.GOOS)
	}
	if asset.URL == "" || asset.Major == "" {
		return "", fmt.Errorf("invalid Postgres.app asset: %+v", asset)
	}
	if _, err := exec.LookPath("hdiutil"); err != nil {
		return "", fmt.Errorf("hdiutil not found in PATH (this should ship with macOS): %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "dbdeployer-pgapp-*")
	if err != nil {
		return "", fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dmgPath := filepath.Join(tmpDir, "Postgres.dmg")
	if err := downloadFile(asset.URL, dmgPath); err != nil {
		return "", fmt.Errorf("downloading %s: %w", asset.URL, err)
	}

	mountPoint := filepath.Join(tmpDir, "mnt")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", fmt.Errorf("creating mount point: %w", err)
	}

	// Mount headlessly (-nobrowse keeps it out of Finder, -readonly avoids
	// any write attempts, -noverify skips the checksum dance).
	mountCmd := exec.Command("hdiutil", "attach",
		"-nobrowse", "-readonly", "-noverify",
		"-mountpoint", mountPoint,
		dmgPath)
	if out, err := mountCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("hdiutil attach failed: %s: %w", string(out), err)
	}

	// Regardless of whether the copy succeeds, detach on the way out.
	detach := func() {
		detachCmd := exec.Command("hdiutil", "detach", mountPoint, "-quiet")
		_ = detachCmd.Run()
	}
	defer detach()

	versionDir := filepath.Join(mountPoint, "Postgres.app", "Contents", "Versions", asset.Major)
	binDir := filepath.Join(versionDir, "bin")
	if _, err := os.Stat(binDir); err != nil {
		return "", fmt.Errorf("expected %s in mounted DMG but it is missing: %w", binDir, err)
	}

	// Query the bundled postgres binary for the exact X.Y version.
	// dbdeployer's VersionToPort expects X.Y format (e.g. "16.4"), and
	// deploy postgresql looks up the binary at ~/opt/postgresql/<X.Y>/.
	fullVersion, err := detectPostgresFullVersion(filepath.Join(binDir, "postgres"))
	if err != nil {
		return "", fmt.Errorf("detecting bundled PostgreSQL version: %w", err)
	}

	targetDir := filepath.Join(sandboxBinaryDir, fullVersion)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("creating target directory %s: %w", targetDir, err)
	}

	// Copy bin/, lib/, share/, include/ preserving symlinks and modes.
	// Postgres.app's `share/postgresql/` contains the datadir contents
	// (postgres.bki, timezonesets, ...) — this is where the `postgres`
	// binary looks for them at runtime after compiled-in path relocation.
	for _, sub := range []string{"bin", "lib", "share", "include"} {
		src := filepath.Join(versionDir, sub)
		dst := filepath.Join(targetDir, sub)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		if err := os.MkdirAll(dst, 0755); err != nil {
			return "", fmt.Errorf("creating %s: %w", dst, err)
		}
		cpCmd := exec.Command("cp", "-a", src+"/.", dst+"/") //nolint:gosec // paths come from mounted DMG
		if out, err := cpCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("copying %s to %s: %s: %w", src, dst, string(out), err)
		}
	}

	// dbdeployer invokes `initdb -L <basedir>/share` (see
	// providers/postgresql/sandbox.go). `initdb` then looks for
	// postgres.bki directly under that path, but Postgres.app keeps
	// datadir files at `share/postgresql/`. Flatten them into
	// `<basedir>/share/` alongside the nested copy so both initdb (flat)
	// and the postgres runtime (nested, via compiled-in sharedir
	// relocation) find what they need.
	srcDataDir := filepath.Join(versionDir, "share", "postgresql")
	dstShare := filepath.Join(targetDir, "share")
	if _, err := os.Stat(srcDataDir); err == nil {
		flatCmd := exec.Command("cp", "-a", srcDataDir+"/.", dstShare+"/") //nolint:gosec // paths come from mounted DMG
		if out, err := flatCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("flattening share/postgresql/ into share/: %s: %w", string(out), err)
		}
	}

	// Sanity check that the required binaries are present.
	for _, bin := range RequiredBinaries() {
		if _, err := os.Stat(filepath.Join(targetDir, "bin", bin)); err != nil {
			return "", fmt.Errorf("required binary %q not found at %s after extraction", bin, targetDir)
		}
	}
	return fullVersion, nil
}

// detectPostgresFullVersion runs `postgres --version` and parses out an
// X.Y version string. Postgres prints something like:
//
//	postgres (PostgreSQL) 16.4 (Postgres.app)
//
// We extract "16.4".
func detectPostgresFullVersion(postgresBinary string) (string, error) {
	out, err := exec.Command(postgresBinary, "--version").Output() //nolint:gosec // binary path constructed from mounted DMG
	if err != nil {
		return "", fmt.Errorf("running %s --version: %w", postgresBinary, err)
	}
	// Regex matches the first X.Y (or X.Y.Z) number in the output.
	re := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
	m := re.FindStringSubmatch(string(out))
	if len(m) < 2 {
		return "", fmt.Errorf("could not parse version from %q", strings.TrimSpace(string(out)))
	}
	// Normalize to X.Y (drop any patch component — dbdeployer's
	// VersionToPort only accepts two segments).
	parts := strings.SplitN(m[1], ".", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected version format %q", m[1])
	}
	return parts[0] + "." + parts[1], nil
}

// downloadFile streams a URL to a local path.
func downloadFile(url, dst string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "dbdeployer")

	// Long timeout: DMGs are ~100 MB and GitHub Releases occasionally throttles.
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.Create(dst) //nolint:gosec // dst is a path we constructed in a tempdir
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
