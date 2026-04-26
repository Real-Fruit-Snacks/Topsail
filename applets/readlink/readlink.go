// Package readlink implements the `readlink` applet.
package readlink

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "readlink",
		Help:  "print the resolved target of a symbolic link",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: readlink [OPTION]... FILE...
Print the value of each symbolic link FILE.

Options:
  -f, --canonicalize   resolve all components, even if some are missing
  -e                   resolve all components; all must exist
  -n                   do not output the trailing newline
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var canon, mustExist, noNL bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-f", a == "--canonicalize":
			canon = true
			args = args[1:]
		case a == "-e", a == "--canonicalize-existing":
			canon = true
			mustExist = true
			args = args[1:]
		case a == "-n", a == "--no-newline":
			noNL = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'f':
					canon = true
				case 'e':
					canon = true
					mustExist = true
				case 'n':
					noNL = true
				default:
					ioutil.Errf("readlink: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("readlink: missing operand")
		return 2
	}

	rc := 0
	for _, name := range args {
		var resolved string
		var err error
		switch {
		case canon:
			resolved, err = filepath.EvalSymlinks(name)
			if err != nil && !mustExist {
				resolved, _ = filepath.Abs(name)
				err = nil
			}
		default:
			resolved, err = os.Readlink(name)
		}
		if err != nil {
			ioutil.Errf("readlink: %s: %v", name, err)
			rc = 1
			continue
		}
		_, _ = ioutil.Stdout.Write([]byte(resolved))
		if !noNL {
			_, _ = ioutil.Stdout.Write([]byte("\n"))
		}
	}
	return rc
}
