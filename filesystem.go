package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	i "github.com/barsanuphe/helpers/ui"
)

// DirectoryExists checks if a directory exists.
func DirectoryExists(path string) (res bool) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		return true
	}
	return
}

// IsDirectoryEmpty checks if files are present in directory.
func IsDirectoryEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// check if at least one file inside
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// AbsoluteFileExists checks if an absolute path is an existing file.
func AbsoluteFileExists(path string) (res bool) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Mode().IsRegular() {
		return true
	}
	return
}

// FileExists checks if a path is valid and returns its absolute path
func FileExists(path string) (absolutePath string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}
	var candidate string
	if filepath.IsAbs(path) {
		candidate = path
	} else {
		candidate = filepath.Join(currentDir, path)
	}

	if AbsoluteFileExists(candidate) {
		absolutePath = candidate
	} else {
		return "", os.ErrNotExist
	}
	return
}

// DeleteEmptyFolders deletes empty folders that may appear after sorting albums.
func DeleteEmptyFolders(root string, ui i.UserInterface) (err error) {
	defer TimeTrack(ui, time.Now(), "Scanning files")

	ui.Debugf("Scanning for empty directories.\n\n")
	deletedDirectories := 0
	deletedDirectoriesThisTime := 0
	atLeastOnce := false

	// loops until all levels of empty directories are deleted
	for !atLeastOnce || deletedDirectoriesThisTime != 0 {
		atLeastOnce = true
		deletedDirectoriesThisTime = 0
		err = filepath.Walk(root, func(path string, fileInfo os.FileInfo, walkError error) (err error) {
			if path == root {
				// do not delete root, even if empty
				return
			}
			// when an directory has just been removed, Walk goes through it a second
			// time with an "file does not exist" error
			if os.IsNotExist(walkError) {
				return
			}
			if fileInfo.IsDir() {
				isEmpty, err := IsDirectoryEmpty(path)
				if err != nil {
					panic(err)
				}
				if isEmpty {
					ui.Debugf("Removing empty directory ", path)
					if err := os.Remove(path); err == nil {
						deletedDirectories++
						deletedDirectoriesThisTime++
					}
				}
			}
			return
		})
		if err != nil {
			ui.Error("Error removing empty directories")
		}
	}

	ui.Debugf("Removed %d directories.", deletedDirectories)
	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.New("source is not a directory")
	}
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return errors.New("destination already exists")
	}
	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}
	return
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CalculateSHA256 calculates a file's current hash
func CalculateSHA256(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hashBytes := sha256.New()
	_, err = io.Copy(hashBytes, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashBytes.Sum(nil)), err
}

// GetUniqueTimestampedFilename for a given filename.
func GetUniqueTimestampedFilename(dir, filename string) (uniqueFilename string, err error) {
	// create dir if necessary
	if !DirectoryExists(dir) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return
		}
	}
	// Mon Jan 2 15:04:05 -0700 MST 2006
	currentTime := time.Now().Local()
	uniqueNameFound := false
	ext := filepath.Ext(filename)
	filenameBase := strings.TrimSuffix(filepath.Base(filename), ext)
	attempts := 0
	for !uniqueNameFound || attempts > 50 {
		suffix := ""
		if attempts > 0 {
			suffix = fmt.Sprintf("_%d", attempts)
		}
		candidate := fmt.Sprintf("%s - %s%s.tar.gz", currentTime.Format("2006-01-02 15:04:05"), filenameBase, suffix)
		// While candidate already exists, change suffix.
		_, err := FileExists(filepath.Join(dir, candidate))
		if err != nil {
			// file not found
			uniqueFilename = filepath.Join(dir, candidate)
			uniqueNameFound = true
		} else {
			attempts++
		}
	}
	return
}
