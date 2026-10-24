// Package sysu provides system utilities.
package sysu

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether file is connected to a terminal.
func IsTerminal(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
}

// TerminalColorEnabled reports whether color output for the terminal is enabled.
func TerminalColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return IsTerminal(os.Stdout)
}

// ReadStdin returns the content of stdin, if something was piped in.
func ReadStdin() (string, error) {
	if IsTerminal(os.Stdin) {
		return "", nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}
	return string(data), nil
}
