// Package chmod implements the `chmod` applet: change file permission bits.
//
// Both octal modes (e.g. 0644, 755) and POSIX symbolic modes
// (u+rwx, go-w, a=rx, u=g, +X) are supported via internal/filemode.
// On Windows file modes are mostly cosmetic; the call is still passed
// through to os.Chmod which handles the read-only bit.
package chmod

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/filemode"
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

MODE is octal (e.g. 0644) or symbolic ([ugoa]*[+-=][rwxXst]+[,...]).
Symbolic forms are evaluated against each file's current mode, so
"chmod u+x prog" adds the owner-execute bit while leaving the rest
of the mode untouched.

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
			// Modes can start with - or + (e.g. chmod -w file). Only
			// treat the token as an option group if every byte is a
			// known short flag — otherwise fall through to MODE.
			if !looksLikeFlagGroup(a) {
				stop = true
				break
			}
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
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("chmod: missing operand")
		return 2
	}
	modeArg := args[0]
	files := args[1:]

	// Up-front syntax check. Symbolic modes need each file's current
	// mode to compute the result, but the parser reports syntactic
	// errors (e.g. "u@x", "9988") consistently against any base.
	if _, err := filemode.Parse(modeArg, 0); err != nil {
		ioutil.Errf("chmod: invalid mode: %s", modeArg)
		return 2
	}

	apply := func(p string) error {
		st, err := os.Stat(p)
		if err != nil {
			return err
		}
		newMode, err := filemode.Parse(modeArg, st.Mode())
		if err != nil {
			return err
		}
		// Preserve the type bits (dir, symlink, etc.) — Parse strips
		// them, but os.Chmod only honors the permission bits anyway.
		if err := os.Chmod(p, newMode); err != nil { //nolint:gosec // chmod operates on the user's specified path by design
			return err
		}
		if verbose {
			ioutil.Errf("mode of '%s' changed to %04o", p, newMode&0o7777)
		}
		return nil
	}

	rc := 0
	for _, name := range files {
		if recursive {
			err := filepath.WalkDir(name, func(p string, _ fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				return apply(p)
			})
			if err != nil {
				ioutil.Errf("chmod: %s: %v", name, err)
				rc = 1
			}
			continue
		}
		if err := apply(name); err != nil {
			ioutil.Errf("chmod: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

// looksLikeFlagGroup returns true if every byte after the leading '-'
// is one of chmod's short option letters. This avoids confusing
// `chmod -w file` (mode argument starting with '-') with an unknown
// flag.
func looksLikeFlagGroup(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case 'R', 'v':
		default:
			return false
		}
	}
	return true
}
