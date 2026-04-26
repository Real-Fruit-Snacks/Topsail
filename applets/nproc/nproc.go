// Package nproc implements the `nproc` applet: print the number of
// processing units available to the current process.
package nproc

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "nproc",
		Help:  "print the number of processing units available",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: nproc [OPTION]
Print the number of processing units available to the current process.
This is the value Go's runtime.NumCPU() reports — it honors any CPU
affinity restrictions imposed on the process.

Options:
  --all              ignore the GOMAXPROCS limit and print runtime.NumCPU()
                     (currently equivalent to the default)
  --ignore=N         exclude N processing units from the count (clamped to 1)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	ignore := 0
	for _, a := range args {
		switch {
		case a == "--all":
			// no-op: runtime.NumCPU is already the visible-to-process count
		case strings.HasPrefix(a, "--ignore="):
			n, err := strconv.Atoi(a[len("--ignore="):])
			if err != nil || n < 0 {
				ioutil.Errf("nproc: invalid --ignore: %s", a)
				return 2
			}
			ignore = n
		case a == "--":
			// no-op
		default:
			ioutil.Errf("nproc: unknown option: %s", a)
			return 2
		}
	}
	n := runtime.NumCPU() - ignore
	if n < 1 {
		n = 1
	}
	_, _ = fmt.Fprintln(ioutil.Stdout, n)
	return 0
}
