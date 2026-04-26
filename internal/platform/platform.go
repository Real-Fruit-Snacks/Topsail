// Package platform centralizes OS-specific behavior so applets stay portable.
// Where the standard library exposes a uniform answer, use it directly;
// where it does not (current user/group lookup on Windows, etc.) the
// per-OS files in this package branch via build tags.
package platform

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether f refers to a terminal.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// TerminalSize returns the (width, height) in cells of the terminal attached
// to f. ok is false if f is not a terminal or the size cannot be queried.
func TerminalSize(f *os.File) (width, height int, ok bool) {
	if f == nil {
		return 0, 0, false
	}
	w, h, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0, 0, false
	}
	return w, h, true
}
