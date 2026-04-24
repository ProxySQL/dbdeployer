// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unpack

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTar builds a tar archive at path from the supplied entries. Each entry's
// Size is set to len(Body) when the entry is a regular file; symlinks carry a
// Linkname and zero size.
type tarEntry struct {
	Name     string
	Linkname string
	Typeflag byte
	Body     []byte
}

func writeTar(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tar: %v", err)
	}
	defer f.Close()
	tw := tar.NewWriter(f)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.Name,
			Linkname: e.Linkname,
			Typeflag: e.Typeflag,
			Mode:     0o644,
		}
		if e.Typeflag == tar.TypeReg {
			hdr.Size = int64(len(e.Body))
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header for %s: %v", e.Name, err)
		}
		if e.Typeflag == tar.TypeReg && len(e.Body) > 0 {
			if _, err := io.Copy(tw, bytes.NewReader(e.Body)); err != nil {
				t.Fatalf("write body for %s: %v", e.Name, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
}

// unpackTarForTest wraps UnpackTar so that the process cwd is restored after
// the call. UnpackTar does os.Chdir(destination); without a restore, tests
// that use t.TempDir() would leak a deleted cwd into subsequent tests.
func unpackTarForTest(t *testing.T, tarPath, dest string) error {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	return UnpackTar(tarPath, dest, SILENT)
}

// Test_symlinkChainEscape reproduces the PoC from the report: a chain of
// dirN -> dirN-1/.. symlinks whose cumulative realpath climbs above the
// extraction directory, followed by a pivot symlink with a path depth equal
// to the entry name (so the previous pathDepth heuristic cannot catch it)
// and a regular file written through the pivot. The extraction must be
// rejected and no file may appear outside the extraction directory.
func Test_symlinkChainEscape(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	victim := filepath.Join(root, "victim")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}

	// Three ".." hops are enough to climb from dest up to `root`, where
	// sibling `victim` lives. The pivot then names `dir3/victim`, which has
	// one path separator — the same as the symlink name `test/myVictim`.
	entries := []tarEntry{
		{Name: "test/dir0/baseFile.txt", Typeflag: tar.TypeReg, Body: []byte("base")},
		{Name: "test/dir1", Linkname: "dir0/..", Typeflag: tar.TypeSymlink},
		{Name: "test/dir2", Linkname: "dir1/..", Typeflag: tar.TypeSymlink},
		{Name: "test/dir3", Linkname: "dir2/..", Typeflag: tar.TypeSymlink},
		{Name: "test/myVictim", Linkname: "dir3/victim", Typeflag: tar.TypeSymlink},
		{Name: "test/myVictim/Exp.txt", Typeflag: tar.TypeReg, Body: []byte("Malicious Text\n")},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	err := unpackTarForTest(t, tarPath, dest)
	if err == nil {
		t.Fatalf("expected extraction to fail, but it succeeded")
	}
	if !strings.Contains(err.Error(), "outside the extraction directory") {
		t.Fatalf("expected 'outside the extraction directory' error, got: %v", err)
	}

	exp := filepath.Join(victim, "Exp.txt")
	if _, err := os.Lstat(exp); err == nil {
		t.Fatalf("attacker file was written to %s despite error — fix did not block the write", exp)
	}
}

// Test_symlinkSingleHopEscape covers the simpler case of a single symlink
// whose target escapes via "..", which the previous pathDepth heuristic
// already caught. Kept as a regression guard so the new code still rejects it.
func Test_symlinkSingleHopEscape(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{
		{Name: "test/placeholder.txt", Typeflag: tar.TypeReg, Body: []byte("x")},
		{Name: "test/escape", Linkname: "../../etc", Typeflag: tar.TypeSymlink},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	if err := unpackTarForTest(t, tarPath, dest); err == nil {
		t.Fatalf("expected extraction to fail for escaping symlink")
	}
}

// Test_symlinkAbsoluteTarget rejects an absolute-path symlink that points
// outside the extraction directory.
func Test_symlinkAbsoluteTarget(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{
		{Name: "test/placeholder.txt", Typeflag: tar.TypeReg, Body: []byte("x")},
		{Name: "test/passwd", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	if err := unpackTarForTest(t, tarPath, dest); err == nil {
		t.Fatalf("expected extraction to fail for absolute-path symlink")
	}
}

// Test_legitimateSymlinkPreserved confirms the fix does not regress normal
// same-directory symlinks of the kind real MySQL tarballs contain
// (e.g. lib/libssl.dylib -> libssl.1.0.0.dylib).
func Test_legitimateSymlinkPreserved(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{
		{Name: "mysql/lib/libssl.1.0.0.dylib", Typeflag: tar.TypeReg, Body: []byte("real")},
		{Name: "mysql/lib/libssl.dylib", Linkname: "libssl.1.0.0.dylib", Typeflag: tar.TypeSymlink},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	if err := unpackTarForTest(t, tarPath, dest); err != nil {
		t.Fatalf("unexpected extraction failure: %v", err)
	}
	linkPath := filepath.Join(dest, "mysql", "lib", "libssl.dylib")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", linkPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to be a symlink", linkPath)
	}
}

// Test_refuseOverwriteSymlinkWithRegular covers the belt-and-suspenders
// protection: even if a malicious entry created a symlink that passed
// validation, a subsequent regular-file entry with the same name must not
// be allowed to write through it. Here we exercise the guard by extracting
// a legitimate intra-archive symlink followed by a regular file at the
// symlink's own path; that latter entry would otherwise follow the symlink.
func Test_refuseOverwriteSymlinkWithRegular(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{
		{Name: "test/target.txt", Typeflag: tar.TypeReg, Body: []byte("original")},
		{Name: "test/alias", Linkname: "target.txt", Typeflag: tar.TypeSymlink},
		{Name: "test/alias", Typeflag: tar.TypeReg, Body: []byte("overwrite attempt")},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	if err := unpackTarForTest(t, tarPath, dest); err == nil {
		t.Fatalf("expected extraction to fail when a regular file would overwrite a symlink")
	}
	body, err := os.ReadFile(filepath.Join(dest, "test", "target.txt"))
	if err != nil {
		t.Fatalf("target.txt missing: %v", err)
	}
	if string(body) != "original" {
		t.Fatalf("target.txt was modified through symlink: got %q", string(body))
	}
}

// Test_relativeDestination verifies that a caller may pass a relative
// destination path. The canonicalization must happen before the internal
// os.Chdir, otherwise the resolved extraction directory would be incorrect.
func Test_relativeDestination(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dest"), 0o755); err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{
		{Name: "mysql/lib/libssl.1.0.0.dylib", Typeflag: tar.TypeReg, Body: []byte("real")},
		{Name: "mysql/lib/libssl.dylib", Linkname: "libssl.1.0.0.dylib", Typeflag: tar.TypeSymlink},
	}
	tarPath := filepath.Join(root, "archive.tar")
	writeTar(t, tarPath, entries)

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to root: %v", err)
	}

	if err := UnpackTar(tarPath, "dest", SILENT); err != nil {
		t.Fatalf("unexpected extraction failure with relative destination: %v", err)
	}
	linkPath := filepath.Join(root, "dest", "mysql", "lib", "libssl.dylib")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("expected symlink at %s: %v", linkPath, err)
	}
}
