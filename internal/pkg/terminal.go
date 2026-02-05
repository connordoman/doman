package pkg

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

// detectTerminalWidth returns the current terminal width in columns.
// If width cannot be determined, it falls back to a sensible default of 80.
func DetectTerminalWidth() int {
	// Respect COLUMNS if set and valid
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if cols, err := strconv.Atoi(colsStr); err == nil && cols > 0 {
			return max(cols-1, 1)
		}
	}

	// Try stdout
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			return max(w-1, 1)
		}
	}

	// Try stderr
	if term.IsTerminal(int(os.Stderr.Fd())) {
		if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
			return max(w-1, 1)
		}
	}

	return 79
}
