// Package basename implements the `basename` applet: strip directory and suffix.
package basename

import (
	"fmt"
	"path"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "basename",
		Help:  "strip directory and suffix from filenames",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: basename NAME [SUFFIX]
       basename OPTION... NAME...
Print NAME with any leading directory components removed. If SUFFIX
is given (in the two-argument form), also strip a trailing SUFFIX.

Options:
  -a, --multiple        treat each argument as a NAME
  -s, --suffix=SUFFIX   remove a trailing SUFFIX from each NAME (implies -a)
  -z, --zero            terminate each output line with NUL, not newline
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		multiple, zero, suffixSet bool
		suffix                    string
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-a", a == "--multiple":
			multiple = true
			args = args[1:]
		case a == "-z", a == "--zero":
			zero = true
			args = args[1:]
		case a == "-s":
			if len(args) < 2 {
				ioutil.Errf("basename: option requires an argument -- 's'")
				return 2
			}
			suffix = args[1]
			suffixSet = true
			multiple = true
			args = args[2:]
		case strings.HasPrefix(a, "--suffix="):
			suffix = a[len("--suffix="):]
			suffixSet = true
			multiple = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && a != "-":
			ioutil.Errf("basename: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("basename: missing operand")
		return 2
	}

	term := byte('\n')
	if zero {
		term = 0
	}

	if !multiple {
		if len(args) > 2 {
			ioutil.Errf("basename: extra operand %q", args[2])
			return 2
		}
		name := path.Base(args[0])
		if len(args) == 2 {
			name = strings.TrimSuffix(name, args[1])
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s%c", name, term)
		return 0
	}

	for _, a := range args {
		name := path.Base(a)
		if suffixSet && suffix != "" {
			name = strings.TrimSuffix(name, suffix)
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s%c", name, term)
	}
	return 0
}
