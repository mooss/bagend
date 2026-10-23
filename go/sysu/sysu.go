// Package sysu provides system utilities.
package sysu

import (
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
