// Package pth provides path and file utilities.
package pth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mooss/bagend/go/pipe"
)

// Resolve returns the absolute, cleaned path with symbolic links resolved.
// It returns an error if the path cannot be made absolute or its symbolic links cannot be resolved.
func Resolve(path string) (string, error) {
	return pipe.Pack(filepath.Abs, path).Chain(filepath.EvalSymlinks)()
}

// ResolveRel returns the cleaned path with symbolic links relative to the working directory.
// It returns an error if the path cannot be made absolute or if the working dir cannot be deduced
// or if symbolic links cannot be resolved.
func ResolveRel(path string) (string, error) {
	relative := pipe.Bend2(filepath.Rel)
	workdir := pipe.Bend(os.Getwd).Chain(filepath.EvalSymlinks)
	resolved := pipe.Pack(Resolve, path)
	return relative(workdir, resolved)
}

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
