// Package ioutil provides binary-safe, mockable wrappers around the process
// standard streams. Applets and the dispatcher MUST use the package-level
// Stdin/Stdout/Stderr instead of os.Stdin/os.Stdout/os.Stderr so tests can
// capture output without subprocess plumbing.
package ioutil

import (
	"fmt"
	"io"
	"os"
)

// Stdin is the input stream applets should read from.
var Stdin io.Reader = os.Stdin

// Stdout is the primary output stream.
var Stdout io.Writer = os.Stdout

// Stderr is the diagnostic output stream.
var Stderr io.Writer = os.Stderr

// Errf formats and writes to Stderr, ensuring a trailing newline. It is the
// canonical way for applets to emit usage/error diagnostics.
func Errf(format string, args ...any) {
	if !endsWithNewline(format) {
		format += "\n"
	}
	_, _ = fmt.Fprintf(Stderr, format, args...)
}

func endsWithNewline(s string) bool {
	return s != "" && s[len(s)-1] == '\n'
}
