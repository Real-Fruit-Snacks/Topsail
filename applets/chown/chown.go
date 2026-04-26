// Package chown implements the `chown` applet: change file ownership.
//
// On Windows, file ownership is a separate model (ACLs / SIDs) that the
// portable os.Chown wrapper does not handle, so this build prints a
// "not supported" diagnostic on Windows and returns 1.
package chown

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "chown",
		Help:  "change file owner and group",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: chown [OPTION]... [USER][:GROUP] FILE...
Change the owner and/or group of each FILE to USER (and optionally GROUP).

USER and GROUP must be numeric uid/gid in this build; named lookups
are deferred to a later wave (they require platform-specific NSS
integration not in the standard library on every OS).

Options:
  -R, --recursive   change files and directories recursively
  -v, --verbose     print a message for each changed file

On Windows, this applet is a no-op that prints a stub diagnostic.
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
			for _, c := range a[1:] {
				switch c {
				case 'R':
					recursive = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("chown: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("chown: missing operand")
		return 2
	}

	if runtime.GOOS == "windows" {
		ioutil.Errf("chown: not supported on Windows in this build")
		return 1
	}

	uid, gid, err := parseSpec(args[0])
	if err != nil {
		ioutil.Errf("chown: %v", err)
		return 2
	}

	rc := 0
	for _, name := range args[1:] {
		if recursive {
			err := filepath.WalkDir(name, func(p string, _ fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if err := os.Chown(p, uid, gid); err != nil { //nolint:gosec // chown -R operates on the user's specified subtree by design
					return err
				}
				if verbose {
					ioutil.Errf("changed ownership of '%s' to %d:%d", p, uid, gid)
				}
				return nil
			})
			if err != nil {
				ioutil.Errf("chown: %s: %v", name, err)
				rc = 1
			}
			continue
		}
		if err := os.Chown(name, uid, gid); err != nil {
			ioutil.Errf("chown: %s: %v", name, err)
			rc = 1
			continue
		}
		if verbose {
			ioutil.Errf("changed ownership of '%s' to %d:%d", name, uid, gid)
		}
	}
	return rc
}

// parseSpec parses USER, USER:, USER:GROUP, :GROUP. -1 means "leave alone".
func parseSpec(s string) (uid, gid int, err error) {
	uid, gid = -1, -1
	if i := strings.Index(s, ":"); i >= 0 {
		left, right := s[:i], s[i+1:]
		if left != "" {
			n, perr := strconv.Atoi(left)
			if perr != nil {
				return 0, 0, perr
			}
			uid = n
		}
		if right != "" {
			n, perr := strconv.Atoi(right)
			if perr != nil {
				return 0, 0, perr
			}
			gid = n
		}
		return uid, gid, nil
	}
	n, perr := strconv.Atoi(s)
	if perr != nil {
		return 0, 0, perr
	}
	return n, gid, nil
}
