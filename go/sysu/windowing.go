package sysu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

// IsWayland reports whether the current session looks like Wayland.
func IsWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	return strings.EqualFold(os.Getenv("XDG_SESSION_TYPE"), "wayland")
}

// toolError joins the tool failure with its stderr when present, so errors carry actionable causes
// such as display connection failures instead of bare exit statuses.
func toolError(tool string, err error, stderr *bytes.Buffer) error {
	trimmed := strings.TrimSpace(stderr.String())
	if trimmed == "" {
		return fmt.Errorf("%s: %w", tool, err)
	}
	return fmt.Errorf("%s: %w (stderr: %s)", tool, err, trimmed)
}

///////////////
// Clipboard //

// clipboardCommands returns clipboard reader commands in preferred order.
// Wayland sessions prefer wl-paste, X11 sessions prefer xclip, with fallback to the other in case
// of misconfigured environment or missing binaries.
func clipboardCommands() [][]string {
	xclip := []string{"xclip", "-o", "-selection", "clipboard"}
	xsel := []string{"xsel", "--clipboard", "--output"}
	wlPaste := []string{"wl-paste", "--no-newline"}

	if IsWayland() {
		return [][]string{wlPaste, xclip, xsel}
	}
	return [][]string{xclip, xsel, wlPaste}
}

// ReadClipboard returns the content of the clipboard.
func ReadClipboard() (string, error) {
	var tried []string
	var errs []error

	clipboardCmds := clipboardCommands()
	for _, args := range clipboardCmds {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		var stderr bytes.Buffer
		tried = append(tried, args[0])
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		cancel()
		if err == nil {
			return string(output), nil
		} // Tool exists but failed (e.g. empty clipboard), simply try the next one.
		errs = append(errs, toolError(args[0], err, &stderr))
	}

	if len(tried) == 0 {
		cmds := lie.Map(func(el []string) string { return el[0] }, clipboardCmds)
		return "", fmt.Errorf("failed to get clipboard content: no clipboard tool found (tried %s)",
			strings.Join(cmds, ","))
	}
	return "", fmt.Errorf("failed to get clipboard content (tried %s):\n%w",
		strings.Join(tried, ", "), errors.Join(errs...))
}

// clipboardWriteCommands returns clipboard writer commands in preferred order.
// Wayland sessions prefer wl-copy, X11 sessions prefer xclip, with fallback to the other in case
// of misconfigured environment or missing binaries.
func clipboardWriteCommands() [][]string {
	xclip := []string{"xclip", "-i", "-selection", "clipboard"}
	xsel := []string{"xsel", "--clipboard", "--input"}
	wlCopy := []string{"wl-copy"}

	if IsWayland() {
		return [][]string{wlCopy, xclip, xsel}
	}
	return [][]string{xclip, xsel, wlCopy}
}

// WriteClipboard writes content to the clipboard.
func WriteClipboard(content string) error {
	var tried []string
	var errs []error

	clipboardCmds := clipboardWriteCommands()
	for _, args := range clipboardCmds {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		tried = append(tried, args[0])
		var stderr bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec
		cmd.Stdin = strings.NewReader(content)
		cmd.Stderr = &stderr

		// Clipboard tools tend to daemonize and block forever (at least wl-copy), so the timeout
		// kills them.
		cmd.WaitDelay = 400 * time.Millisecond

		err := cmd.Run()
		ctxErr := ctx.Err()
		cancel()
		if err == nil {
			return nil
		}
		if strings.TrimSpace(stderr.String()) == "" &&
			(ctxErr == context.DeadlineExceeded || strings.Contains(err.Error(), "signal: killed")) {
			return nil
		}
		// Tool exists but failed, simply try the next one.
		errs = append(errs, toolError(args[0], err, &stderr))
	}

	if len(tried) == 0 {
		cmds := lie.Map(func(el []string) string { return el[0] }, clipboardCmds)
		return fmt.Errorf("failed to set clipboard content: no clipboard tool found (tried %s)",
			strings.Join(cmds, ","))
	}
	return fmt.Errorf("failed to set clipboard content (tried %s):\n%w",
		strings.Join(tried, ", "), errors.Join(errs...))
}
