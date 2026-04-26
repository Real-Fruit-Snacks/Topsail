// Package mkdir implements the `mkdir` applet: make directories.
package mkdir

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "mkdir",
		Help:  "make directories",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: mkdir [OPTION]... DIRECTORY...
Create the DIRECTORY(ies), if they do not already exist.

Options:
  -p, --parents       no error if existing, make parent directories as needed
  -m, --mode=MODE     set file mode (octal), not a=rwx - umask
  -v, --verbose       print a message for each created directory

Symbolic mode strings (u+rwx, etc.) are not yet supported; use octal.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		parents, verbose, modeSet bool
		mode                      = os.FileMode(0o777)
	)

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
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case a == "-m":
			if len(args) < 2 {
				ioutil.Errf("mkdir: option requires an argument -- 'm'")
				return 2
			}
			m, err := parseMode(args[1])
			if err != nil {
				ioutil.Errf("mkdir: invalid mode: %s", args[1])
				return 2
			}
			mode = m
			modeSet = true
			args = args[2:]
		case strings.HasPrefix(a, "--mode="):
			val := a[len("--mode="):]
			m, err := parseMode(val)
			if err != nil {
				ioutil.Errf("mkdir: invalid mode: %s", val)
				return 2
			}
			mode = m
			modeSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'p':
					parents = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("mkdir: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("mkdir: missing operand")
		return 2
	}

	rc := 0
	for _, dir := range args {
		var err error
		if parents {
			err = os.MkdirAll(dir, mode)
		} else {
			err = os.Mkdir(dir, mode)
		}
		if err != nil {
			if parents && errors.Is(err, fs.ErrExist) {
				continue
			}
			ioutil.Errf("mkdir: %s: %v", dir, err)
			rc = 1
			continue
		}
		// Re-apply mode in case the umask masked bits we wanted.
		if modeSet {
			_ = os.Chmod(dir, mode)
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "mkdir: created directory '%s'\n", dir)
		}
	}
	return rc
}

// parseMode parses an octal file mode. Symbolic modes are not supported here.
func parseMode(s string) (os.FileMode, error) {
	n, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(n), nil
}
