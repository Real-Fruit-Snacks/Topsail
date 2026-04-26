// Package dirname implements the `dirname` applet: strip last component.
package dirname

import (
	"fmt"
	"path"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "dirname",
		Help:  "strip last component from a file name",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: dirname [-z] NAME...
Print NAME with the last component removed. If NAME contains no
slashes, output ".". If NAME is "/", output "/".

Options:
  -z, --zero    terminate each output line with NUL, not newline
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	zero := false

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch a {
		case "--":
			args = args[1:]
			stop = true
		case "-z", "--zero":
			zero = true
			args = args[1:]
		default:
			if len(a) > 1 && a[0] == '-' && a != "-" {
				ioutil.Errf("dirname: unknown option: %s", a)
				return 2
			}
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("dirname: missing operand")
		return 2
	}

	term := byte('\n')
	if zero {
		term = 0
	}

	for _, a := range args {
		// Use path.Dir (forward slashes only) rather than filepath.Dir;
		// dirname is a string-manipulation utility per POSIX, not a
		// filesystem one, so '/' is the separator on every platform.
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s%c", path.Dir(a), term)
	}
	return 0
}
