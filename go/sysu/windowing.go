package sysu

import (
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

// clipboardCommands returns clipboard reader commands in preferred order.
// Wayland sessions prefer wl-paste, X11 sessions prefer xclip, with fallback to the other in case
// of misconfigured environment or missing binaries.
func clipboardCommands() [][]string {
	xclip := []string{"xclip", "-o", "-selection", "clipboard"}
	xsel := []string{"xsel", "--clipboard", "--output"}
	wlPaste := []string{"wl-paste", "-n"}

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
		tried = append(tried, args[0])

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		output, err := exec.CommandContext(ctx, args[0], args[1:]...).Output() //nolint:gosec
		cancel()
		if err == nil {
			return string(output), nil
		} // Tool exists but failed (e.g. empty clipboard), simply try the next one.
		errs = append(errs, fmt.Errorf("%s: %w", args[0], err))
	}

	if len(tried) == 0 {
		cmds := lie.Map(func(el []string) string { return el[0] }, clipboardCmds)
		return "", fmt.Errorf("failed to get clipboard content: no clipboard tool found (tried %s)",
			strings.Join(cmds, ","))
	}
	return "", fmt.Errorf("failed to get clipboard content (tried %s): %w",
		strings.Join(tried, ", "), errors.Join(errs...))
}
