// Package cat implements the `cat` applet: concatenate files to stdout.
package cat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "cat",
		Help:  "concatenate files to stdout",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: cat [OPTION]... [FILE]...
Concatenate FILE(s) to standard output.

With no FILE, or when FILE is "-", read standard input.

Options:
  -n     number all output lines
  -b     number non-empty output lines (overrides -n)
  -E     display $ at end of each line
  -T     display TAB characters as ^I
  -s     suppress repeated empty output lines
`

type options struct {
	number, numberNonBlank, showEnds, showTabs, squeeze bool
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var opts options

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-":
			stop = true
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'n':
					opts.number = true
				case 'b':
					opts.numberNonBlank = true
				case 'E':
					opts.showEnds = true
				case 'T':
					opts.showTabs = true
				case 's':
					opts.squeeze = true
				default:
					ioutil.Errf("cat: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}
	transform := opts.number || opts.numberNonBlank || opts.showEnds || opts.showTabs || opts.squeeze

	rc := 0
	state := &transformState{}
	for _, name := range files {
		if err := processFile(name, transform, opts, state); err != nil {
			ioutil.Errf("cat: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

type transformState struct {
	lineNum   int
	prevBlank bool
}

func processFile(name string, transform bool, opts options, state *transformState) error {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the entire point
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	if !transform {
		_, err := io.Copy(ioutil.Stdout, r)
		return err
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		blank := line == ""
		if opts.squeeze && blank && state.prevBlank {
			continue
		}
		state.prevBlank = blank

		if opts.showTabs {
			line = strings.ReplaceAll(line, "\t", "^I")
		}
		switch {
		case opts.numberNonBlank:
			if !blank {
				state.lineNum++
				_, _ = fmt.Fprintf(ioutil.Stdout, "%6d\t", state.lineNum)
			}
		case opts.number:
			state.lineNum++
			_, _ = fmt.Fprintf(ioutil.Stdout, "%6d\t", state.lineNum)
		}
		_, _ = io.WriteString(ioutil.Stdout, line)
		if opts.showEnds {
			_, _ = io.WriteString(ioutil.Stdout, "$")
		}
		_, _ = io.WriteString(ioutil.Stdout, "\n")
	}
	return sc.Err()
}
