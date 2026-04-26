// Package chmod implements the `chmod` applet: change file permission bits.
//
// This Wave 3 build supports octal mode strings (e.g. 0644, 755). Symbolic
// modes (u+rwx, g-w) are deferred to a later wave. On Windows file modes
// are mostly cosmetic; the call is still passed through to os.Chmod which
// handles the read-only bit.
package chmod

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "chmod",
		Help:  "change file mode bits",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: chmod [OPTION]... MODE FILE...
Change file mode bits of each FILE to MODE.

MODE must be octal (e.g. 0644). Symbolic modes are not yet supported.

Options:
  -R, --recursive   change files and directories recursively
  -v, --verbose     print a message for each changed file
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var recursive, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-R", a == "--recursive":
			recursive = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ok := true
			for _, c := range a[1:] {
				switch c {
				case 'R':
					recursive = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("chmod: invalid option -- '%c'", c)
					return 2
				}
			}
			if !ok {
				return 2
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("chmod: missing operand")
		return 2
	}
	mode, err := parseMode(args[0])
	if err != nil {
		ioutil.Errf("chmod: invalid mode: %s", args[0])
		return 2
	}

	rc := 0
	for _, name := range args[1:] {
		if recursive {
			err := filepath.WalkDir(name, func(p string, _ fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if err := os.Chmod(p, mode); err != nil { //nolint:gosec // chmod -R operates on the user's specified subtree by design
					return err
				}
				if verbose {
					ioutil.Errf("mode of '%s' changed to %04o", p, mode)
				}
				return nil
			})
			if err != nil {
				ioutil.Errf("chmod: %s: %v", name, err)
				rc = 1
			}
			continue
		}
		if err := os.Chmod(name, mode); err != nil {
			ioutil.Errf("chmod: %s: %v", name, err)
			rc = 1
			continue
		}
		if verbose {
			ioutil.Errf("mode of '%s' changed to %04o", name, mode)
		}
	}
	return rc
}

func parseMode(s string) (os.FileMode, error) {
	n, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(n), nil
}
