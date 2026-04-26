// Package ln implements the `ln` applet: create file links.
package ln

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "ln",
		Help:  "create links between files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: ln [OPTION]... TARGET LINK_NAME
       ln [OPTION]... TARGET... DIRECTORY
Create LINK_NAME (or, in the second form, links inside DIRECTORY)
that point at TARGET.

Options:
  -s, --symbolic    create symbolic links instead of hard links
  -f, --force       remove existing destination files
  -v, --verbose     print a message for each link created
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var symbolic, force, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-s", a == "--symbolic":
			symbolic = true
			args = args[1:]
		case a == "-f", a == "--force":
			force = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 's':
					symbolic = true
				case 'f':
					force = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("ln: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("ln: missing operand")
		return 2
	}

	dest := args[len(args)-1]
	sources := args[:len(args)-1]

	destInfo, err := os.Stat(dest)
	destIsDir := err == nil && destInfo.IsDir()

	if len(sources) > 1 && !destIsDir {
		ioutil.Errf("ln: target '%s' is not a directory", dest)
		return 2
	}

	rc := 0
	for _, src := range sources {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}
		if force {
			_ = os.Remove(target)
		}
		var linkErr error
		if symbolic {
			linkErr = os.Symlink(src, target)
		} else {
			linkErr = os.Link(src, target)
		}
		if linkErr != nil {
			ioutil.Errf("ln: failed to link '%s' -> '%s': %v", target, src, linkErr)
			rc = 1
			continue
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "'%s' -> '%s'\n", target, src)
		}
	}
	return rc
}
