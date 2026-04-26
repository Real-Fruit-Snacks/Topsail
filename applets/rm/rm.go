// Package rm implements the `rm` applet: remove files and directories.
package rm

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "rm",
		Help:  "remove files and directories",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: rm [OPTION]... FILE...
Remove (unlink) the FILE(s).

Options:
  -r, -R, --recursive    remove directories and their contents recursively
  -f, --force            ignore nonexistent files; never prompt
  -i, --interactive      accepted but never prompts
  -d, --dir              remove empty directories
  -v, --verbose          print a message for each removed file
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var recursive, force, emptyDir, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-r", a == "-R", a == "--recursive":
			recursive = true
			args = args[1:]
		case a == "-f", a == "--force":
			force = true
			args = args[1:]
		case a == "-i", a == "--interactive":
			args = args[1:]
		case a == "-d", a == "--dir":
			emptyDir = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'r', 'R':
					recursive = true
				case 'f':
					force = true
				case 'i':
					// no-op
				case 'd':
					emptyDir = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("rm: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		if force {
			return 0
		}
		ioutil.Errf("rm: missing operand")
		return 2
	}

	rc := 0
	for _, name := range args {
		info, err := os.Lstat(name)
		if err != nil {
			if force && errors.Is(err, fs.ErrNotExist) {
				continue
			}
			ioutil.Errf("rm: %s: %v", name, err)
			rc = 1
			continue
		}
		switch {
		case info.IsDir() && recursive:
			if err := os.RemoveAll(name); err != nil {
				ioutil.Errf("rm: %s: %v", name, err)
				rc = 1
				continue
			}
		case info.IsDir() && emptyDir:
			if err := os.Remove(name); err != nil {
				ioutil.Errf("rm: %s: %v", name, err)
				rc = 1
				continue
			}
		case info.IsDir():
			ioutil.Errf("rm: cannot remove '%s': Is a directory", name)
			rc = 1
			continue
		default:
			if err := os.Remove(name); err != nil {
				ioutil.Errf("rm: %s: %v", name, err)
				rc = 1
				continue
			}
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "removed '%s'\n", name)
		}
	}
	return rc
}
