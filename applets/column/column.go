// Package column implements a minimal `column` applet: format input
// into columns.
package column

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
		Name:  "column",
		Help:  "columnate lists",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: column [OPTION]... [FILE]...
Pretty-print input as evenly-spaced columns.

Options:
  -t                use the input field separator and align fields into a table
  -s SEP            input field separator for -t (default: TAB)
  -o SEP            output field separator (default: two spaces)
`

// readRows is split out so the deferred Close stays out of the per-file
// loop in Main.
func readRows(name string, table bool, inSep string) ([][]string, error) {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var rows [][]string
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			rows = append(rows, []string{""})
			continue
		}
		if table {
			rows = append(rows, strings.Split(line, inSep))
		} else {
			rows = append(rows, []string{line})
		}
	}
	return rows, sc.Err()
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var table bool
	inSep := "\t"
	outSep := "  "

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-t":
			table = true
			args = args[1:]
		case a == "-s":
			if len(args) < 2 {
				ioutil.Errf("column: option requires an argument -- 's'")
				return 2
			}
			inSep = args[1]
			args = args[2:]
		case a == "-o":
			if len(args) < 2 {
				ioutil.Errf("column: option requires an argument -- 'o'")
				return 2
			}
			outSep = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("column: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	var rows [][]string
	for _, name := range files {
		got, err := readRows(name, table, inSep)
		if err != nil {
			ioutil.Errf("column: %s: %v", name, err)
			return 1
		}
		rows = append(rows, got...)
	}

	if !table {
		for _, r := range rows {
			_, _ = fmt.Fprintln(ioutil.Stdout, r[0])
		}
		return 0
	}

	// Compute column widths.
	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	widths := make([]int, cols)
	for _, r := range rows {
		for i, f := range r {
			if len(f) > widths[i] {
				widths[i] = len(f)
			}
		}
	}
	for _, r := range rows {
		var b strings.Builder
		for i, f := range r {
			if i > 0 {
				b.WriteString(outSep)
			}
			if i < len(widths) {
				_, _ = fmt.Fprintf(&b, "%-*s", widths[i], f)
			} else {
				b.WriteString(f)
			}
		}
		_, _ = fmt.Fprintln(ioutil.Stdout, strings.TrimRight(b.String(), " "))
	}
	return 0
}
