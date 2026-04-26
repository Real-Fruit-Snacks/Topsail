// Package pwd implements the `pwd` applet: print the working directory.
package pwd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "pwd",
		Help:  "print the working directory",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: pwd [-LP]
Print the absolute pathname of the current working directory.

Options:
  -L     do not resolve symlinks (default)
  -P     resolve symlinks to print the canonical physical directory
`

// Main is the applet entry point.
func Main(argv []string) int {
	physical := false
	for _, a := range argv[1:] {
		switch a {
		case "-L":
			physical = false
		case "-P":
			physical = true
		case "--":
			// Accept and ignore — pwd has no positional operands.
		default:
			ioutil.Errf("pwd: invalid option: %s", a)
			return 2
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		ioutil.Errf("pwd: %v", err)
		return 1
	}
	if physical {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			dir = resolved
		}
	}
	_, _ = io.WriteString(ioutil.Stdout, dir+"\n")
	return 0
}
