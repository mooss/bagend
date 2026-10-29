package pth_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mooss/bagend/go/pth"
	"golang.org/x/sys/unix"
)

func TestMayCreateAbsentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile")

	ok, err := pth.MayCreate(path)
	if err != nil {
		t.Fatalf("MayCreate(%q) returned error: %v", path, err)
	}
	if !ok {
		t.Errorf("MayCreate(%q) = false, want true", path)
	}
}

func TestMayCreateExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists")
	//nolint:gosec
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	ok, err := pth.MayCreate(path)
	if err != nil {
		t.Fatalf("MayCreate(%q) returned error: %v", path, err)
	}
	if ok {
		t.Errorf("MayCreate(%q) = true, want false", path)
	}
}

func TestMayCreateInvalidPath(t *testing.T) {
	for _, path := range []string{"", "foo/"} {
		ok, err := pth.MayCreate(path)
		if ok {
			t.Errorf("MayCreate(%q) = true, want false", path)
		}
		if !errors.Is(err, os.ErrInvalid) {
			t.Errorf("MayCreate(%q) error = %v, want %v", path, err, os.ErrInvalid)
		}
	}
}

func TestMayCreateMissingParent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-such-dir", "file")

	ok, err := pth.MayCreate(path)
	if ok {
		t.Errorf("MayCreate(%q) = true, want false", path)
	}
	if err == nil {
		t.Errorf("MayCreate(%q) error = nil, want non-nil", path)
	}
}

func TestMayCreateParentIsFile(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "file")
	//nolint:gosec
	if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "child")

	ok, err := pth.MayCreate(path)
	if ok {
		t.Errorf("MayCreate(%q) = true, want false", path)
	}
	if !errors.Is(err, unix.ENOTDIR) {
		t.Errorf("MayCreate(%q) error = %v, want %v", path, err, unix.ENOTDIR)
	}
}
