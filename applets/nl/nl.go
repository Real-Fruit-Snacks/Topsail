// Package nl implements a simplified `nl` applet: number lines.
package nl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "nl",
		Help:  "number lines of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: nl [OPTION]... [FILE]...
Number the lines of each FILE (or stdin) to standard output.

Options:
  -b a    number all lines (default)
  -b t    number only non-empty lines
  -b n    do not number lines
  -i N    line number increment (default 1)
  -s SEP  separator between number and line (default TAB)
  -w N    line number width (default 6)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	mode := "a"
	inc := 1
	sep := "\t"
	width := 6

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-b":
			if len(args) < 2 {
				ioutil.Errf("nl: option requires an argument -- 'b'")
				return 2
			}
			if args[1] != "a" && args[1] != "t" && args[1] != "n" {
				ioutil.Errf("nl: invalid line numbering type: %s", args[1])
				return 2
			}
			mode = args[1]
			args = args[2:]
		case a == "-i":
			if len(args) < 2 {
				ioutil.Errf("nl: option requires an argument -- 'i'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil {
				ioutil.Errf("nl: invalid increment: %s", args[1])
				return 2
			}
			inc = n
			args = args[2:]
		case a == "-s":
			if len(args) < 2 {
				ioutil.Errf("nl: option requires an argument -- 's'")
				return 2
			}
			sep = args[1]
			args = args[2:]
		case a == "-w":
			if len(args) < 2 {
				ioutil.Errf("nl: option requires an argument -- 'w'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("nl: invalid width: %s", args[1])
				return 2
			}
			width = n
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("nl: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}
	rc := 0
	n := 1
	for _, name := range files {
		if err := nlOne(name, &n, inc, mode, sep, width); err != nil {
			ioutil.Errf("nl: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func nlOne(name string, n *int, inc int, mode, sep string, width int) error {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		showNum := false
		switch mode {
		case "a":
			showNum = true
		case "t":
			showNum = strings.TrimSpace(line) != ""
		case "n":
			showNum = false
		}
		if showNum {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%*d%s%s\n", width, *n, sep, line)
			*n += inc
		} else {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%*s%s%s\n", width, "", sep, line)
		}
	}
	return sc.Err()
}
