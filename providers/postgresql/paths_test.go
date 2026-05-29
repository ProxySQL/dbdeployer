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

func TestPgBinDir(t *testing.T) {
	basedir := "/opt/postgresql/18.4"
	if runtime.GOOS == "linux" {
		want := basedir + "/lib/postgresql/18/bin"
		if got := PgBinDir(basedir, "18.4"); got != want {
			t.Fatalf("PgBinDir = %q, want %q", got, want)
		}
	} else if got := PgBinDir(basedir, "18.4"); got != basedir+"/bin" {
		t.Fatalf("PgBinDir = %q, want %q", got, basedir+"/bin")
	}
}

func TestResolvePsqlBinary(t *testing.T) {
	basedir := filepath.Join(t.TempDir(), "postgresql", "18.4")
	nestedDir := filepath.Join(basedir, "lib", "postgresql", "18", "bin")
	flatDir := PgBinDir(basedir, "18.4")
	for _, dir := range []string{nestedDir, flatDir} {
		if dir == nestedDir && runtime.GOOS != "linux" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	realPsql := filepath.Join(flatDir, "psql")
	if err := os.WriteFile(realPsql, []byte{0x7f, 'E', 'L', 'F'}, 0755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "linux" {
		if err := os.Remove(realPsql); err != nil {
			t.Fatal(err)
		}
		realPsql = filepath.Join(nestedDir, "psql")
		if err := os.WriteFile(realPsql, []byte{0x7f, 'E', 'L', 'F'}, 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ResolvePsqlBinary(basedir, "18.4")
	if err != nil {
		t.Fatalf("ResolvePsqlBinary: %v", err)
	}
	if got != realPsql {
		t.Fatalf("ResolvePsqlBinary = %q, want %q", got, realPsql)
	}

	wrapper := filepath.Join(flatDir, "psql")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec /usr/bin/pg_wrapper\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "linux" {
		if err := os.Remove(realPsql); err != nil {
			t.Fatal(err)
		}
		if _, err := ResolvePsqlBinary(basedir, "18.4"); err == nil {
			t.Fatal("expected error when only pg_wrapper is available")
		}
		return
	}

	got, err = ResolvePsqlBinary(basedir, "18.4")
	if err != nil {
		t.Fatalf("ResolvePsqlBinary with wrapper present: %v", err)
	}
	if got != realPsql {
		t.Fatalf("ResolvePsqlBinary = %q, want nested %q when flat is pg_wrapper", got, realPsql)
	}

	if err := os.Remove(realPsql); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolvePsqlBinary(basedir, "18.4"); err == nil {
		t.Fatal("expected error when only pg_wrapper is available")
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
