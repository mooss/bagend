// Package pth provides path and file utilities.
package pth

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mooss/bagend/go/pipe"
	"golang.org/x/sys/unix"
)

// CheckWritable reports whether path names a regular file that can be opened for writing.
// It returns an error if the path does not exist, is not a regular file, or cannot be opened or
// closed successfully.
func CheckWritable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("path %q is not a regular file", path)
	}

	handle, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("file %q is not writable: %w", path, err)
	}
	if err := handle.Close(); err != nil {
		return fmt.Errorf("close error in writable check %q: %w", path, err)
	}

	return nil
}

// Exists reports whether the path itself exists, without following its final symlink.
// False does not necessarily mean the path is absent, it might mean that the permissions were not
// sufficient to check.
func Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// mayCreateUnix reports whether path is absent and the parent-directory permission check succeeds.
// A true result is not a creation guarantee.
// Only compiles on Unix.
func mayCreateUnix(path string) (bool, error) {
	// Require a filename, not an empty path or a trailing separator.
	dir, name := filepath.Split(path)
	if name == "" {
		return false, os.ErrInvalid
	}

	if dir == "" {
		dir = "."
	}

	missing, err := Missing(path)
	if err != nil || !missing {
		return false, err
	}

	info, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		// ENOTDIR: a path component is not a directory.
		return false, unix.ENOTDIR
	}

	err = unix.Faccessat(
		unix.AT_FDCWD, // Resolve dir relative to the current working directory.
		dir,
		unix.W_OK|unix.X_OK, // Need write and search permission.
		unix.AT_EACCESS,     // check using the effective user/group ID, not the real one.
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

// MayCreate reports whether path is absent and the parent-directory permission check succeeds.
// A true result is not a creation guarantee.
func MayCreate(path string) (bool, error) {
	return mayCreateUnix(path)
}

// Missing reports whether no entry exists at path.
// Other errors, including permission errors, are returned.
func Missing(path string) (bool, error) {
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, fs.ErrNotExist):
		return true, nil
	default:
		return false, err
	}
}

// Resolve returns the absolute, cleaned path with symbolic links resolved.
// It returns an error if the path cannot be made absolute or its symbolic links cannot be resolved.
func Resolve(path string) (string, error) {
	return pipe.Pack(filepath.Abs, path).Chain(filepath.EvalSymlinks)()
}

// ResolveRel returns path resolved relative to the resolved working directory.
// It returns an error if path or the working directory cannot be resolved.
func ResolveRel(path string) (string, error) {
	relative := pipe.Bend2(filepath.Rel)
	workdir := pipe.Bend(os.Getwd).Chain(filepath.EvalSymlinks)
	resolved := pipe.Pack(Resolve, path)
	return relative(workdir, resolved)
}

// SplitExists returns the longest existing prefix of the given path, as well as the remaining
// elements.
//
//revive:disable:confusing-results
func SplitExists(path string) (string, string) {
	elements := filepath.SplitList(path)
	for i := len(elements) - 1; i >= 0; i++ {
		candidate := filepath.Join(elements[:i]...)
		if Exists(candidate) {
			return candidate, filepath.Join(elements[i:]...)
		}
	}

	return "", path
}
