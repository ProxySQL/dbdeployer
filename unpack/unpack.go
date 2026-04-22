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

/* Originally copyrighted as
// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/
/*
 Code adapted and enhanced from examples to the book:
 Programming in Go by Mark Summerfield
 http://www.qtrac.eu/gobook.html

 Original author: Mark Summerfield
 Converted to package by Giuseppe Maxia in 2018

 The original code was a stand-alone program, and it
 had a few bugs:
 * when extracting from a tar file: when there
 isn't a separate item for each directory, the
 extraction fails.
 * The attributes of the files were not reproduced
 in the extracted files.
 This code fixes those problems and introduces a
 destination directory and verbosity
 levels for the extraction

*/

package unpack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xi2/xz"

	"github.com/ProxySQL/dbdeployer/common"
	"github.com/ProxySQL/dbdeployer/globals"
)

const (
	SILENT  = iota // No output
	VERBOSE        // Minimal feedback about extraction operations
	CHATTY         // Full details of what is being extracted
)

var Verbose int

func condPrint(s string, nl bool, level int) {
	if Verbose >= level {
		if nl {
			fmt.Println(s)
		} else {
			fmt.Print(s)
		}
	}
}

func validSuffix(filename string) bool {
	for _, suffix := range []string{globals.TgzExt, globals.TarExt, globals.TarGzExt, globals.TarXzExt} {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}

// canonicalExtractDir returns an absolute, symlink-resolved path for the
// extraction destination. It must be called before os.Chdir so that a relative
// destination is resolved against the original working directory, not the
// post-Chdir one.
func canonicalExtractDir(destination string) (string, error) {
	absDir, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("error defining the absolute path of '%s': %s", destination, err)
	}
	if resolved, err := filepath.EvalSymlinks(absDir); err == nil {
		return resolved, nil
	}
	return absDir, nil
}

func UnpackXzTar(filename string, destination string, verbosityLevel int) (err error) {
	Verbose = verbosityLevel
	if !common.FileExists(filename) {
		return fmt.Errorf("file %s not found", filename)
	}
	if !common.DirExists(destination) {
		return fmt.Errorf("directory %s not found", destination)
	}
	filename, err = common.AbsolutePath(filename)
	if err != nil {
		return err
	}
	destinationAbs, err := canonicalExtractDir(destination)
	if err != nil {
		return err
	}
	err = os.Chdir(destinationAbs)
	if err != nil {
		return errors.Wrapf(err, "error changing directory to %s", destination)
	}

	f, err := os.Open(filename) // #nosec G304
	if err != nil {
		return err
	}
	defer f.Close() // #nosec G307
	// Create an xz Reader
	r, err := xz.NewReader(f, 0)
	if err != nil {
		return err
	}
	// Create a tar Reader
	tr := tar.NewReader(r)
	return unpackTarFiles(tr, destinationAbs)
}

func UnpackTar(filename string, destination string, verbosityLevel int) (err error) {
	Verbose = verbosityLevel
	f, err := os.Stat(destination)
	if os.IsNotExist(err) {
		return fmt.Errorf("destination directory '%s' does not exist", destination)
	}
	filemode := f.Mode()
	if !filemode.IsDir() {
		return fmt.Errorf("destination '%s' is not a directory", destination)
	}
	if !validSuffix(filename) {
		return fmt.Errorf("unrecognized archive suffix")
	}
	var file *os.File
	// #nosec G304
	if file, err = os.Open(filename); err != nil {
		return err
	}
	defer file.Close() // #nosec G307
	destinationAbs, err := canonicalExtractDir(destination)
	if err != nil {
		return err
	}
	err = os.Chdir(destinationAbs)
	if err != nil {
		return errors.Wrapf(err, "error changing directory to %s", destination)
	}
	var fileReader io.Reader = file
	var decompressor *gzip.Reader
	if strings.HasSuffix(filename, globals.GzExt) {
		if decompressor, err = gzip.NewReader(file); err != nil {
			return err
		}
		defer decompressor.Close()
	}
	var reader *tar.Reader
	if decompressor != nil {
		reader = tar.NewReader(decompressor)
	} else {
		reader = tar.NewReader(fileReader)
	}
	return unpackTarFiles(reader, destinationAbs)
}

// unpackTarFiles extracts reader's entries into extractAbsDir. The caller must
// supply an absolute, symlink-resolved path (see canonicalExtractDir) so the
// validation helpers can compare canonical paths for containment.
func unpackTarFiles(reader *tar.Reader, extractAbsDir string) error {
	const errLinkedDirectoryOutside = "linked directory '%s' is outside the extraction directory"
	const errDirectoryOutside = "directory for entry '%s' is outside the extraction directory"
	var err error
	var header *tar.Header
	var count int = 0
	var reSlash = regexp.MustCompile(`/.*`)

	innerDir := ""
	for {
		if header, err = reader.Next(); err != nil {
			if err == io.EOF {
				condPrint("Files ", false, CHATTY)
				condPrint(strconv.Itoa(count), true, 1)
				return nil // OK
			}
			return err
		}
		// cond_print(fmt.Sprintf("%#v\n", header), true, CHATTY)
		/*
			tar.Header{
				Typeflag:0x30,
				Name:"mysql-8.0.11-macos10.13-x86_64/docs/INFO_SRC",
				Linkname:"",
				Size:185,
				Mode:420,
				Uid:7161,
				Gid:10,
				Uname:"pb2user",
				Gname:"owner",
				ModTime:time.Time{wall:0x0, ext:63658769207, loc:(*time.Location)(0x13730e0)},
				AccessTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				ChangeTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				Devmajor:0, Devminor:0,
				Xattrs:map[string]string(nil),
				PAXRecords:map[string]string(nil),
				Format:0}
			tar.Header{
				Typeflag:0x32,
				Name:"mysql-8.0.11-macos10.13-x86_64/lib/libssl.dylib",
				Linkname:"libssl.1.0.0.dylib",
				Size:0,
				Mode:493,
				Uid:7161,
				Gid:10,
				Uname:"pb2user",
				Gname:"owner",
				ModTime:time.Time{wall:0x0, ext:63658772525, loc:(*time.Location)(0x13730e0)},
				AccessTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				ChangeTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				Devmajor:0,
				Devminor:0,
				Xattrs:map[string]string(nil),
				PAXRecords:map[string]string(nil),
				Format:0}
		*/
		filemode := os.FileMode(header.Mode)
		filename := sanitizedName(header.Name)
		fileDir := path.Dir(filename)
		upperDir := reSlash.ReplaceAllString(fileDir, "")
		if innerDir != "" {
			if upperDir != innerDir {
				return fmt.Errorf("found more than one directory inside the tarball\n"+
					"<%s> and <%s>", upperDir, innerDir)
			}
		} else {
			innerDir = upperDir
		}

		absFilePath := filepath.Join(extractAbsDir, filename)
		absFileDir := filepath.Dir(absFilePath)

		// Validate that the entry's parent directory (after resolving any symlinks
		// created by previous tar entries) stays inside the extraction directory.
		// This closes the chain-symlink traversal bypass where an earlier entry
		// creates a symlink whose realpath escapes extractAbsDir.
		if _, err := resolveInsideExtractDir(absFileDir, extractAbsDir); err != nil {
			return fmt.Errorf(errDirectoryOutside, filename)
		}

		if _, err = os.Lstat(fileDir); os.IsNotExist(err) {
			if err = os.MkdirAll(fileDir, globals.PublicDirectoryAttr); err != nil {
				return err
			}
			condPrint(" + "+fileDir+" ", true, CHATTY)
		}
		if header.Typeflag == 0 {
			header.Typeflag = tar.TypeReg
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(filename, globals.PublicDirectoryAttr); err != nil {
				return err
			}
		case tar.TypeReg:
			// Refuse to write through a pre-existing symlink at the target name:
			// os.Create would follow it and write outside the extraction directory.
			if info, lerr := os.Lstat(filename); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to overwrite existing symlink at '%s'", filename)
			}
			if err = unpackTarFile(filename, reader); err != nil {
				return err
			}
			err = os.Chmod(filename, filemode)
			if err != nil {
				return err
			}
			count++
			condPrint(filename, true, CHATTY)
			if count%10 == 0 {
				mark := "."
				if count%100 == 0 {
					mark = strconv.Itoa(count)
				}
				if Verbose < CHATTY {
					condPrint(mark, false, 1)
				}
			}
		case tar.TypeSymlink:
			if header.Linkname == "" {
				return fmt.Errorf("file %s is a symlink, but no link information was provided", filename)
			}
			// Build the absolute path the symlink would point to. We concatenate
			// with a raw separator instead of filepath.Join so that ".." components
			// in Linkname are preserved: EvalSymlinks must walk through any
			// intermediate symlinks before evaluating ".." against their real
			// targets. filepath.Join would lexically collapse the ".." and miss
			// chain-symlink escapes.
			var targetPath string
			if filepath.IsAbs(header.Linkname) {
				targetPath = header.Linkname
			} else {
				targetPath = absFileDir + string(os.PathSeparator) + header.Linkname
			}
			if _, err := resolveInsideExtractDir(targetPath, extractAbsDir); err != nil {
				return fmt.Errorf(errLinkedDirectoryOutside, header.Linkname)
			}
			condPrint(fmt.Sprintf("%s -> %s", filename, header.Linkname), true, CHATTY)
			err = os.Symlink(header.Linkname, filename)
			if err != nil {
				return fmt.Errorf("%#v\n#ERROR: %s", header, err)
			}
		}
	}
	// return nil
}

// resolveInsideExtractDir resolves target through the filesystem (following any
// existing symlinks, resolving ".." components *after* symlink expansion) and
// confirms the result is inside extractAbsDir. When the full target does not
// exist yet, the deepest existing ancestor is resolved and the remaining
// lexical suffix is appended; this lets us validate symlinks whose targets
// have not been created yet without losing chain-traversal detection (any
// ".." that would cross a symlink lives in an existing ancestor, so it gets
// resolved through the filesystem rather than lexically).
func resolveInsideExtractDir(target, extractAbsDir string) (string, error) {
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		if !pathInside(resolved, extractAbsDir) {
			return "", fmt.Errorf("path '%s' resolves to '%s' outside extraction directory '%s'", target, resolved, extractAbsDir)
		}
		return resolved, nil
	}
	// Walk up one directory at a time using filepath.Dir so volume roots
	// (e.g. "/" on POSIX, "C:\" on Windows) are handled portably. Terminate
	// at the fixed point, where filepath.Dir no longer shortens the path.
	parent := filepath.Dir(target)
	for {
		if resolved, err := filepath.EvalSymlinks(parent); err == nil {
			if !pathInside(resolved, extractAbsDir) {
				return "", fmt.Errorf("ancestor of '%s' resolves to '%s' outside extraction directory '%s'", target, resolved, extractAbsDir)
			}
			rel, err := filepath.Rel(parent, target)
			if err != nil {
				return "", err
			}
			combined := filepath.Join(resolved, rel)
			if !pathInside(combined, extractAbsDir) {
				return "", fmt.Errorf("path '%s' would resolve to '%s' outside extraction directory '%s'", target, combined, extractAbsDir)
			}
			return combined, nil
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", fmt.Errorf("cannot resolve any ancestor of '%s'", target)
		}
		parent = next
	}
}

func pathInside(candidate, dir string) bool {
	if candidate == dir {
		return true
	}
	sep := string(os.PathSeparator)
	if !strings.HasSuffix(dir, sep) {
		dir += sep
	}
	return strings.HasPrefix(candidate, dir)
}

func unpackTarFile(filename string,
	reader *tar.Reader) (err error) {
	var writer *os.File
	// #nosec G304
	if writer, err = os.Create(filename); err != nil {
		return err
	}
	defer writer.Close() // #nosec G307
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	return nil
}

func sanitizedName(filename string) string {
	if len(filename) > 1 && filename[1] == ':' {
		filename = filename[2:]
	}
	filename = strings.TrimLeft(filename, "\\/.")
	filename = strings.Replace(filename, "../", "", -1)
	return strings.Replace(filename, "..\\", "", -1)
}

func VerifyTarFile(fileName string) error {
	if !validSuffix(fileName) {
		return fmt.Errorf("unrecognized archive suffix %s", fileName)
	}
	var file *os.File
	var err error
	// #nosec G304
	if file, err = os.Open(fileName); err != nil {
		return fmt.Errorf("[open file Validation] %s", err)
	}
	defer file.Close() // #nosec G307
	var fileReader io.Reader = file
	var decompressor *gzip.Reader
	var xzDecompressor *xz.Reader

	if strings.HasSuffix(fileName, globals.GzExt) {
		if decompressor, err = gzip.NewReader(file); err != nil {
			return fmt.Errorf("[gz Validation] %s", err)
		}
		defer decompressor.Close()
	} else {
		if strings.HasSuffix(fileName, globals.TarXzExt) {
			if xzDecompressor, err = xz.NewReader(file, 0); err != nil {
				return fmt.Errorf("[xz Validation] %s", err)
			}
		}
	}
	var reader *tar.Reader
	if decompressor != nil {
		reader = tar.NewReader(decompressor)
	} else {
		if xzDecompressor != nil {
			reader = tar.NewReader(xzDecompressor)
		} else {
			reader = tar.NewReader(fileReader)
		}
	}
	var header *tar.Header
	expectedDirName := common.BaseName(fileName)
	reExt := regexp.MustCompile(`\.(?:tar(?:\.gz|\.xz)?)$`)
	expectedDirName = reExt.ReplaceAllString(expectedDirName, "")

	if header, err = reader.Next(); err != nil {
		if err == io.EOF {
			return fmt.Errorf("[EOF Validation] file %s is empty", fileName)
		}
		return fmt.Errorf("[header validation] %s", err)
	}
	innerFileName := sanitizedName(header.Name)
	fileDir := path.Dir(innerFileName)

	reSlash := regexp.MustCompile(`/.*`)
	fileDir = reSlash.ReplaceAllString(fileDir, "")

	if fileDir != expectedDirName {
		return fmt.Errorf("inner directory name different from tarball name\n"+
			"Expected: %s - Found: %s", expectedDirName, fileDir)
	}
	return nil
}
