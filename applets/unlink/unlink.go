// Package unlink implements the `unlink` applet: remove a single file
// via the unlink(2) system call. Distinct from `rm` in that it accepts
// exactly one operand and refuses to recurse or remove directories.
package unlink

import (
	"os"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "unlink",
		Help:  "remove a single file",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: unlink FILE
Call the unlink function to remove FILE.

Exactly one FILE is required; unlink does not recurse, does not accept
flags, and refuses to remove directories. Use rm for richer behavior.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) != 1 {
		ioutil.Errf("unlink: expected exactly one operand")
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" {
		// Help is handled by the wrapper; multi-call invocation can
		// still hit this path. Print the usage and exit cleanly.
		_, _ = ioutil.Stdout.Write([]byte(usage))
		return 0
	}
	// Refuse to unlink directories. POSIX unlink(2) returns EISDIR for
	// directories; Go's os.Remove falls back to rmdir(2) on empty dirs,
	// which would mask that distinction. We Lstat first to keep the
	// strict "files only" contract this applet promises.
	st, err := os.Lstat(args[0])
	if err != nil {
		ioutil.Errf("unlink: cannot unlink '%s': %v", args[0], err)
		return 1
	}
	if st.IsDir() {
		ioutil.Errf("unlink: cannot unlink '%s': Is a directory", args[0])
		return 1
	}
	if err := os.Remove(args[0]); err != nil {
		ioutil.Errf("unlink: cannot unlink '%s': %v", args[0], err)
		return 1
	}
	return 0
}
