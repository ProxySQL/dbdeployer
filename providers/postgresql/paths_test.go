package postgresql

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLibraryPath(t *testing.T) {
	basedir := "/opt/postgresql/18.4"
	got := LibraryPath(basedir, "18.4")
	if !strings.HasPrefix(got, basedir+"/lib") {
		t.Fatalf("LibraryPath = %q, expected to start with %q/lib", got, basedir)
	}
	if runtime.GOOS == "linux" {
		want := basedir + "/lib" + string(filepath.ListSeparator) + basedir + "/lib/postgresql/18/lib"
		if got != want {
			t.Fatalf("LibraryPath = %q, want %q", got, want)
		}
	}
}

func TestCopyClientLibraries(t *testing.T) {
	tmpDir := t.TempDir()
	usrLib := filepath.Join(tmpDir, "usr", "lib", "aarch64-linux-gnu")
	if err := os.MkdirAll(usrLib, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usrLib, "libpq.so.5.24"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("libpq.so.5.24", filepath.Join(usrLib, "libpq.so.5")); err != nil {
		t.Fatal(err)
	}

	dstLib := filepath.Join(tmpDir, "basedir", "lib")
	if err := copyClientLibraries(tmpDir, dstLib); err != nil {
		t.Fatalf("copyClientLibraries: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dstLib, "libpq.so.5")); err != nil {
		t.Fatalf("libpq.so.5 not copied: %v", err)
	}
}
