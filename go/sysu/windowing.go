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

///////////////
// Clipboard //
///////////////

// toolError joins the tool failure with its stderr when present, so errors carry actionable causes
// such as display connection failures instead of bare exit statuses.
func toolError(tool string, err error, stderr *bytes.Buffer) error {
	trimmed := strings.TrimSpace(stderr.String())
	if trimmed == "" {
		return fmt.Errorf("%s: %w", tool, err)
	}
	return fmt.Errorf("%s: %w (stderr: %s)", tool, err, trimmed)
}

// clipboardCommandOrder orders Wayland and X11 commands according to the current session.
func clipboardCommandOrder(xclip, xsel, wayland []string) [][]string {
	if IsWayland() {
		return [][]string{wayland, xclip, xsel}
	}
	return [][]string{xclip, xsel, wayland}
}

// runClipboardCommands tries available tools in order and reports their failures.
func runClipboardCommands(
	clipboardCmds [][]string, action string, run func([]string) (string, error),
) (string, error) {
	var tried []string
	var errs []error

	for _, args := range clipboardCmds {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		tried = append(tried, args[0])
		output, err := run(args)
		if err == nil {
			return output, nil
		}
		errs = append(errs, err)
	}

	if len(tried) == 0 {
		cmds := lie.Map(func(el []string) string { return el[0] }, clipboardCmds)
		msg := "failed to %s clipboard content: no clipboard tool found (tried %s)"
		return "", fmt.Errorf(msg, action, strings.Join(cmds, ","))
	}

	msg := "failed to %s clipboard content (tried %s):\n%w"
	return "", fmt.Errorf(msg, action, strings.Join(tried, ", "), errors.Join(errs...))
}

////////////////////
// Clipboard read //

// clipboardReaders returns clipboard reader commands in preferred order.
// Wayland sessions prefer wl-paste, X11 sessions prefer xclip, with fallback to the other in case
// of misconfigured environment or missing binaries.
func clipboardReaders() [][]string {
	xclip := []string{"xclip", "-o", "-selection", "clipboard"}
	xsel := []string{"xsel", "--clipboard", "--output"}
	wlPaste := []string{"wl-paste", "--no-newline"}

	return clipboardCommandOrder(xclip, xsel, wlPaste)
}

// ReadClipboard returns the content of the clipboard.
func ReadClipboard() (string, error) {
	return runClipboardCommands(clipboardReaders(), "get", func(args []string) (string, error) {
		var stderr bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		cancel()
		if err != nil {
			return "", toolError(args[0], err, &stderr)
		}
		return string(output), nil
	})
}

/////////////////////
// Clipboard write //

// clipboardWriters returns clipboard writer commands in preferred order.
// Wayland sessions prefer wl-copy, X11 sessions prefer xclip, with fallback to the other in case
// of misconfigured environment or missing binaries.
func clipboardWriters() [][]string {
	xclip := []string{"xclip", "-i", "-selection", "clipboard"}
	xsel := []string{"xsel", "--clipboard", "--input"}
	wlCopy := []string{"wl-copy"}

	return clipboardCommandOrder(xclip, xsel, wlCopy)
}

// WriteClipboard writes content to the clipboard.
func WriteClipboard(content string) error {
	_, err := runClipboardCommands(clipboardWriters(), "set", func(args []string) (string, error) {
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
			return "", nil
		}
		if strings.TrimSpace(stderr.String()) == "" && (ctxErr == context.DeadlineExceeded ||
			strings.Contains(err.Error(), "signal: killed")) {
			return "", nil
		}
		return "", toolError(args[0], err, &stderr)
	})
	return err
}
