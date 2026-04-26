// Package rmdir implements the `rmdir` applet: remove empty directories.
package rmdir

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "rmdir",
		Help:  "remove empty directories",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: rmdir [OPTION]... DIRECTORY...
Remove the DIRECTORY(ies), if they are empty.

Options:
  -p, --parents                    also remove parents that become empty
      --ignore-fail-on-non-empty   ignore failure when a directory is not empty
  -v, --verbose                    print a message for each removed directory
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var parents, ignoreNonEmpty, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-p", a == "--parents":
			parents = true
			args = args[1:]
		case a == "--ignore-fail-on-non-empty":
			ignoreNonEmpty = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'p':
					parents = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("rmdir: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("rmdir: missing operand")
		return 2
	}

	rc := 0
	for _, dir := range args {
		if !removeChain(dir, parents, ignoreNonEmpty, verbose) {
			rc = 1
		}
	}
	return rc
}

// removeChain removes dir, then optionally walks up its parents removing
// any that have become empty. Failure on the explicit dir is reported and
// returns false (rc=1); failures while walking parents are silent — we just
// stop, matching GNU rmdir -p semantics.
func removeChain(dir string, parents, ignoreNonEmpty, verbose bool) bool {
	if err := os.Remove(dir); err != nil {
		if ignoreNonEmpty && isNotEmpty(err) {
			return true
		}
		ioutil.Errf("rmdir: %s: %v", dir, err)
		return false
	}
	if verbose {
		_, _ = fmt.Fprintf(ioutil.Stdout, "rmdir: removing directory, '%s'\n", dir)
	}
	if !parents {
		return true
	}
	for {
		next := filepath.Dir(dir)
		if next == dir || next == "." {
			return true
		}
		if err := os.Remove(next); err != nil {
			return true // parent not empty or other error: stop silently.
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "rmdir: removing directory, '%s'\n", next)
		}
		dir = next
	}
}

// isNotEmpty is best-effort cross-OS detection of ENOTEMPTY-equivalents.
func isNotEmpty(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrExist) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "not empty") || strings.Contains(s, "directory not empty")
}
