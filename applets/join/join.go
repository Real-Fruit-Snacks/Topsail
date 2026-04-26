// Package join implements a simplified `join` applet: relational join
// of two files on a key field.
package join

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
		Name:  "join",
		Help:  "join lines of two files on a common field",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: join [OPTION]... FILE1 FILE2
Output a line for each pair of input lines having identical join fields.

Options:
  -1 N      use field N from FILE1 as the join key (default 1)
  -2 N      use field N from FILE2 as the join key (default 1)
  -t SEP    field separator (default: any whitespace)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	keyA, keyB := 1, 1
	var sep string

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-1":
			if len(args) < 2 {
				ioutil.Errf("join: option requires an argument -- '1'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("join: invalid field: %s", args[1])
				return 2
			}
			keyA = n
			args = args[2:]
		case a == "-2":
			if len(args) < 2 {
				ioutil.Errf("join: option requires an argument -- '2'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("join: invalid field: %s", args[1])
				return 2
			}
			keyB = n
			args = args[2:]
		case a == "-t":
			if len(args) < 2 {
				ioutil.Errf("join: option requires an argument -- 't'")
				return 2
			}
			sep = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("join: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) != 2 {
		ioutil.Errf("join: requires exactly two files")
		return 2
	}

	rowsA, err := readRows(args[0], sep)
	if err != nil {
		ioutil.Errf("join: %s: %v", args[0], err)
		return 1
	}
	rowsB, err := readRows(args[1], sep)
	if err != nil {
		ioutil.Errf("join: %s: %v", args[1], err)
		return 1
	}

	joinSep := " "
	if sep != "" {
		joinSep = sep
	}

	for _, ra := range rowsA {
		if keyA-1 >= len(ra) {
			continue
		}
		k := ra[keyA-1]
		for _, rb := range rowsB {
			if keyB-1 >= len(rb) {
				continue
			}
			if rb[keyB-1] != k {
				continue
			}
			rest := append([]string{k}, fields(ra, keyA-1)...)
			rest = append(rest, fields(rb, keyB-1)...)
			_, _ = fmt.Fprintln(ioutil.Stdout, strings.Join(rest, joinSep))
		}
	}
	return 0
}

func fields(row []string, skip int) []string {
	out := make([]string, 0, len(row))
	for i, v := range row {
		if i == skip {
			continue
		}
		out = append(out, v)
	}
	return out
}

func readRows(name, sep string) ([][]string, error) {
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
		var row []string
		if sep == "" {
			row = strings.Fields(sc.Text())
		} else {
			row = strings.Split(sc.Text(), sep)
		}
		rows = append(rows, row)
	}
	return rows, sc.Err()
}
