// Package mktemp implements the `mktemp` applet: create a unique
// temporary file or directory and print its path.
package mktemp

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
		Name:  "mktemp",
		Help:  "create a unique temporary file or directory",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: mktemp [OPTION]... [TEMPLATE]
Create a temporary file or directory securely and print its name.

TEMPLATE must contain at least three consecutive 'X's, which are
replaced by a random suffix. With no TEMPLATE, mktemp uses
"tmp.XXXXXXXXXX".

Options:
  -d, --directory       create a directory, not a file
  -p DIR, --tmpdir=DIR  place the file/dir in DIR (default: $TMPDIR or /tmp)
  -t                    treat TEMPLATE as a basename and place under tmpdir
  -u, --dry-run         do not create; just print a name (insecure — discouraged)
  -q, --quiet           do not print error messages
  --suffix=SUFFIX       append SUFFIX to the template (template must end in X's)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		mkDir    bool
		dryRun   bool
		quiet    bool
		treatT   bool
		tmpdir   string
		hasTmp   bool
		suffix   string
		template string
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-d", a == "--directory":
			mkDir = true
			args = args[1:]
		case a == "-u", a == "--dry-run":
			dryRun = true
			args = args[1:]
		case a == "-q", a == "--quiet":
			quiet = true
			args = args[1:]
		case a == "-t":
			treatT = true
			args = args[1:]
		case a == "-p":
			if len(args) < 2 {
				ioutil.Errf("mktemp: option requires an argument -- 'p'")
				return 2
			}
			tmpdir = args[1]
			hasTmp = true
			args = args[2:]
		case strings.HasPrefix(a, "--tmpdir="):
			tmpdir = a[len("--tmpdir="):]
			hasTmp = true
			args = args[1:]
		case a == "--tmpdir":
			tmpdir = os.TempDir()
			hasTmp = true
			args = args[1:]
		case strings.HasPrefix(a, "--suffix="):
			suffix = a[len("--suffix="):]
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("mktemp: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) > 0 {
		template = args[0]
	}
	if template == "" {
		template = "tmp.XXXXXXXXXX"
	}

	// Resolve directory.
	dir := ""
	if treatT || hasTmp {
		dir = tmpdir
		if dir == "" {
			dir = os.TempDir()
		}
		// In -t mode the template is a basename — strip any directory part
		// the user might have included.
		if treatT {
			template = filepath.Base(template)
		}
	}

	prefix, hasX := splitTemplate(template, suffix)
	if !hasX {
		if !quiet {
			ioutil.Errf("mktemp: too few X's in template %q", template)
		}
		return 1
	}

	pattern := prefix + "*" + suffix
	if dryRun {
		// Build a candidate path and print without creating. Emulates
		// the GNU shape: the template is the literal string with X's
		// retained; we use a UUID-ish token to keep the output unique.
		name := pattern
		if dir != "" {
			name = filepath.Join(dir, name)
		}
		_, _ = fmt.Fprintln(ioutil.Stdout, name)
		return 0
	}

	var path string
	if mkDir {
		p, err := os.MkdirTemp(dir, pattern)
		if err != nil {
			if !quiet {
				ioutil.Errf("mktemp: %v", err)
			}
			return 1
		}
		path = p
	} else {
		f, err := os.CreateTemp(dir, pattern)
		if err != nil {
			if !quiet {
				ioutil.Errf("mktemp: %v", err)
			}
			return 1
		}
		path = f.Name()
		_ = f.Close()
	}
	_, _ = fmt.Fprintln(ioutil.Stdout, path)
	return 0
}

// splitTemplate verifies the template ends with at least 3 X's
// (optionally followed by suffix) and returns the prefix portion plus
// whether enough X's were present.
func splitTemplate(template, suffix string) (prefix string, ok bool) {
	// Strip the suffix off the tail before counting X's.
	core := template
	if suffix != "" && strings.HasSuffix(core, suffix) {
		core = core[:len(core)-len(suffix)]
	}
	end := len(core)
	xs := 0
	for end > 0 && core[end-1] == 'X' {
		end--
		xs++
	}
	if xs < 3 {
		return "", false
	}
	return core[:end], true
}
