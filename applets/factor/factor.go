// Package factor implements the `factor` applet: factor integers.
package factor

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "factor",
		Help:  "factor positive integers",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: factor [NUMBER]...
Print the prime factors of each NUMBER. With no NUMBERs, read from stdin.

Output: "N: F1 F2 ..." per number.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) == 0 {
		sc := bufio.NewScanner(ioutil.Stdin)
		for sc.Scan() {
			for _, tok := range strings.Fields(sc.Text()) {
				if rc := factorOne(tok); rc != 0 {
					return rc
				}
			}
		}
		return 0
	}
	for _, a := range args {
		if rc := factorOne(a); rc != 0 {
			return rc
		}
	}
	return 0
}

func factorOne(s string) int {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		ioutil.Errf("factor: '%s' is not a valid positive integer", s)
		return 1
	}
	_, _ = fmt.Fprintf(ioutil.Stdout, "%d:", n)
	if n < 2 {
		_, _ = ioutil.Stdout.Write([]byte("\n"))
		return 0
	}
	for d := uint64(2); d*d <= n; d++ {
		for n%d == 0 {
			_, _ = fmt.Fprintf(ioutil.Stdout, " %d", d)
			n /= d
		}
	}
	if n > 1 {
		_, _ = fmt.Fprintf(ioutil.Stdout, " %d", n)
	}
	_, _ = ioutil.Stdout.Write([]byte("\n"))
	return 0
}
