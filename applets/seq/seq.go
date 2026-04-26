// Package seq implements the `seq` applet: print numeric sequences.
package seq

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "seq",
		Help:  "print a numeric sequence",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: seq [OPTION]... LAST
       seq [OPTION]... FIRST LAST
       seq [OPTION]... FIRST INCREMENT LAST
Print numbers from FIRST to LAST in steps of INCREMENT (default 1).

Options:
  -s, --separator=STRING   use STRING between numbers (default: newline)
  -w, --equal-width        equalize widths by padding with leading zeros
  -f, --format=FORMAT      use printf-style FORMAT (default: %g)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	sep := "\n"
	format := "%g"
	var equalWidth bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-s":
			if len(args) < 2 {
				ioutil.Errf("seq: option requires an argument -- 's'")
				return 2
			}
			sep = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--separator="):
			sep = a[len("--separator="):]
			args = args[1:]
		case a == "-w", a == "--equal-width":
			equalWidth = true
			args = args[1:]
		case a == "-f":
			if len(args) < 2 {
				ioutil.Errf("seq: option requires an argument -- 'f'")
				return 2
			}
			format = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--format="):
			format = a[len("--format="):]
			args = args[1:]
		case strings.HasPrefix(a, "-") && a != "-" && !isNumber(a):
			ioutil.Errf("seq: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	first, inc, last := 1.0, 1.0, 0.0
	var err error
	switch len(args) {
	case 1:
		last, err = strconv.ParseFloat(args[0], 64)
	case 2:
		first, err = strconv.ParseFloat(args[0], 64)
		if err == nil {
			last, err = strconv.ParseFloat(args[1], 64)
		}
	case 3:
		first, err = strconv.ParseFloat(args[0], 64)
		if err == nil {
			inc, err = strconv.ParseFloat(args[1], 64)
			if err == nil {
				last, err = strconv.ParseFloat(args[2], 64)
			}
		}
	default:
		ioutil.Errf("seq: missing operand")
		return 2
	}
	if err != nil {
		ioutil.Errf("seq: invalid number")
		return 2
	}
	if inc == 0 {
		ioutil.Errf("seq: invalid Zero increment value")
		return 2
	}

	width := 0
	if equalWidth {
		w1 := len(fmt.Sprintf(format, first))
		w2 := len(fmt.Sprintf(format, last))
		if w1 > width {
			width = w1
		}
		if w2 > width {
			width = w2
		}
	}

	wrote := false
	for n := first; (inc > 0 && n <= last) || (inc < 0 && n >= last); n += inc {
		if wrote {
			_, _ = io.WriteString(ioutil.Stdout, sep)
		}
		wrote = true
		s := fmt.Sprintf(format, n)
		if equalWidth && len(s) < width {
			s = strings.Repeat("0", width-len(s)) + s
		}
		_, _ = io.WriteString(ioutil.Stdout, s)
	}
	if wrote {
		_, _ = io.WriteString(ioutil.Stdout, "\n")
	}
	return 0
}

// isNumber reports whether s parses as a (possibly negative) float so the
// flag parser doesn't mistake "-3" for an option.
func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
