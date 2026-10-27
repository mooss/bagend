package sysu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeClipboardTool creates a small executable shell script that stands in for a clipboard utility
// such as wl-paste or xclip.
//
// The tests put dir at the front of PATH, so when ReadClipboard looks up a utility by name, it
// finds this script instead of a real system installation.
func writeClipboardTool(t *testing.T, dir, name, script string) {
	t.Helper()

	path := filepath.Join(dir, name)
	//nolint:gosec
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestReadClipboardPrefersWaylandTool checks tool selection when a Wayland display is available.
// Both wl-paste and xclip are present, but wl-paste should be tried first and its output returned.
func TestReadClipboardPrefersWaylandTool(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-paste", "printf '%s' 'wayland clipboard'")
	writeClipboardTool(t, dir, "xclip", "printf '%s' 'xclip clipboard'")

	got, err := ReadClipboard()
	if err != nil {
		t.Fatalf("ReadClipboard() error = %v", err)
	}
	if got != "wayland clipboard" {
		t.Fatalf("ReadClipboard() = %q, want %q", got, "wayland clipboard")
	}
}

// TestReadClipboardFallsBackWhenPreferredToolFails verifies that a failed attempt does not prevent
// ReadClipboard from trying another available tool.
func TestReadClipboardFallsBackWhenPreferredToolFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-paste", "exit 1")
	writeClipboardTool(t, dir, "xclip", "printf '%s' 'fallback clipboard'")

	got, err := ReadClipboard()
	if err != nil {
		t.Fatalf("ReadClipboard() error = %v", err)
	}
	if got != "fallback clipboard" {
		t.Fatalf("ReadClipboard() = %q, want %q", got, "fallback clipboard")
	}
}

// TestReadClipboardReportsWhenNoToolsAreFound gives ReadClipboard an empty PATH.
// Executable lookup therefore cannot find any clipboard utility, and the function should return an
// empty result along with an explanatory error.
func TestReadClipboardReportsWhenNoToolsAreFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	got, err := ReadClipboard()
	if err == nil {
		t.Fatal("ReadClipboard() error = nil, want an error")
	}
	if got != "" {
		t.Fatalf("ReadClipboard() = %q, want empty string", got)
	}
	if !strings.Contains(err.Error(), "no clipboard tool found") {
		t.Fatalf("ReadClipboard() error = %q, want it to report no clipboard tool found", err)
	}
}

// TestReadClipboardReportsFailuresFromAllTools distinguishes tools being unavailable from tools
// being present but unable to read the clipboard.
func TestReadClipboardReportsFailuresFromAllTools(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-paste", "exit 1")
	writeClipboardTool(t, dir, "xclip", "exit 1")
	writeClipboardTool(t, dir, "xsel", "exit 1")

	got, err := ReadClipboard()
	if err == nil {
		t.Fatal("ReadClipboard() error = nil, want an error")
	}
	if got != "" {
		t.Fatalf("ReadClipboard() = %q, want empty string", got)
	}
	for _, want := range []string{"wl-paste, xclip, xsel", "wl-paste:"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ReadClipboard() error = %q, want it to contain %q", err, want)
		}
	}
}

// TestWriteClipboardPrefersWaylandTool checks tool selection when a Wayland display is available.
// Both wl-copy and xclip are present, but wl-copy should be tried first and receive the content.
func TestWriteClipboardPrefersWaylandTool(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-copy", "cat > "+filepath.Join(dir, "wl-copy.out"))
	writeClipboardTool(t, dir, "xclip", "cat > "+filepath.Join(dir, "xclip.out"))

	if err := WriteClipboard("wayland clipboard"); err != nil {
		t.Fatalf("WriteClipboard() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "wl-copy.out"))
	if err != nil {
		t.Fatalf("read wl-copy.out: %v", err)
	}
	if string(got) != "wayland clipboard" {
		t.Fatalf("wl-copy received %q, want %q", got, "wayland clipboard")
	}
	if _, err := os.Stat(filepath.Join(dir, "xclip.out")); !os.IsNotExist(err) {
		t.Fatalf("xclip should not have been called, stat error = %v", err)
	}
}

// TestWriteClipboardFallsBackWhenPreferredToolFails verifies that a failed attempt does not prevent
// WriteClipboard from trying another available tool.
func TestWriteClipboardFallsBackWhenPreferredToolFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-copy", "exit 1")
	writeClipboardTool(t, dir, "xclip", "cat > "+filepath.Join(dir, "xclip.out"))

	if err := WriteClipboard("fallback clipboard"); err != nil {
		t.Fatalf("WriteClipboard() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "xclip.out"))
	if err != nil {
		t.Fatalf("read xclip.out: %v", err)
	}
	if string(got) != "fallback clipboard" {
		t.Fatalf("xclip received %q, want %q", got, "fallback clipboard")
	}
}

// TestWriteClipboardReportsWhenNoToolsAreFound gives WriteClipboard an empty PATH.
// Executable lookup therefore cannot find any clipboard utility, and the function should return an
// explanatory error.
func TestWriteClipboardReportsWhenNoToolsAreFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	if err := WriteClipboard("some content"); err == nil {
		t.Fatal("WriteClipboard() error = nil, want an error")
	} else if !strings.Contains(err.Error(), "no clipboard tool found") {
		t.Fatalf("WriteClipboard() error = %q, want it to report no clipboard tool found", err)
	}
}

// TestWriteClipboardReportsFailuresFromAllTools distinguishes tools being unavailable from tools
// being present but unable to write to the clipboard.
func TestWriteClipboardReportsFailuresFromAllTools(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("XDG_SESSION_TYPE", "")

	writeClipboardTool(t, dir, "wl-copy", "exit 1")
	writeClipboardTool(t, dir, "xclip", "exit 1")
	writeClipboardTool(t, dir, "xsel", "exit 1")

	err := WriteClipboard("some content")
	if err == nil {
		t.Fatal("WriteClipboard() error = nil, want an error")
	}
	for _, want := range []string{"wl-copy, xclip, xsel", "wl-copy:"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("WriteClipboard() error = %q, want it to contain %q", err, want)
		}
	}
}
