// Package realpath implements the `realpath` applet: print the
// resolved (absolute, symlink-free) path of each operand.
package realpath

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "realpath",
		Help:  "print the resolved absolute path of each FILE",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: realpath [OPTION]... FILE...
Print the resolved absolute path of each FILE.

By default, all path components must exist and symbolic links are
resolved.

Options:
  -e, --canonicalize-existing   require all components to exist (default)
  -m, --canonicalize-missing    do not require components to exist
  -s, --strip                   make the path absolute but do not resolve symlinks
  -q, --quiet                   suppress most error messages
  -z, --zero                    end each output with NUL instead of newline
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		canonExisting = true
		canonMissing  bool
		stripSymlinks bool
		quiet         bool
		zero          bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-e", a == "--canonicalize-existing":
			canonExisting = true
			canonMissing = false
			args = args[1:]
		case a == "-m", a == "--canonicalize-missing":
			canonMissing = true
			canonExisting = false
			args = args[1:]
		case a == "-s", a == "--strip", a == "--no-symlinks":
			stripSymlinks = true
			args = args[1:]
		case a == "-q", a == "--quiet":
			quiet = true
			args = args[1:]
		case a == "-z", a == "--zero":
			zero = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("realpath: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("realpath: missing operand")
		return 2
	}

	terminator := "\n"
	if zero {
		terminator = "\x00"
	}

	rc := 0
	for _, p := range args {
		resolved, err := resolveOne(p, stripSymlinks, canonMissing, canonExisting)
		if err != nil {
			if !quiet {
				ioutil.Errf("realpath: %s: %v", p, err)
			}
			rc = 1
			continue
		}
		_, _ = fmt.Fprint(ioutil.Stdout, resolved, terminator)
	}
	return rc
}

func resolveOne(p string, strip, missing, existing bool) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if strip {
		return filepath.Clean(abs), nil
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if missing {
		// Some component doesn't exist; clean and return the absolute path.
		return filepath.Clean(abs), nil
	}
	if existing {
		return "", err
	}
	return "", err
}
