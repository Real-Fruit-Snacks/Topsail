// Package uniq implements the `uniq` applet: filter adjacent duplicate lines.
package uniq

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
		Name:  "uniq",
		Help:  "report or filter adjacent duplicate lines",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: uniq [OPTION]... [INPUT [OUTPUT]]
Filter adjacent matching lines from INPUT (or stdin) to OUTPUT (or stdout).

Options:
  -c, --count          prefix lines by their count
  -d, --repeated       only print duplicate lines, one for each group
  -u, --unique         only print unique (non-repeated) lines
  -i, --ignore-case    case-insensitive comparison
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var count, dupOnly, uniqOnly, ignoreCase bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c", a == "--count":
			count = true
			args = args[1:]
		case a == "-d", a == "--repeated":
			dupOnly = true
			args = args[1:]
		case a == "-u", a == "--unique":
			uniqOnly = true
			args = args[1:]
		case a == "-i", a == "--ignore-case":
			ignoreCase = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'c':
					count = true
				case 'd':
					dupOnly = true
				case 'u':
					uniqOnly = true
				case 'i':
					ignoreCase = true
				default:
					ioutil.Errf("uniq: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) > 2 {
		ioutil.Errf("uniq: extra operand %q", args[2])
		return 2
	}
	inName := "-"
	outName := "-"
	if len(args) >= 1 {
		inName = args[0]
	}
	if len(args) >= 2 {
		outName = args[1]
	}

	in, closeIn, err := openIn(inName)
	if err != nil {
		ioutil.Errf("uniq: %s: %v", inName, err)
		return 1
	}
	defer closeIn()
	out, closeOut, err := openOut(outName)
	if err != nil {
		ioutil.Errf("uniq: %s: %v", outName, err)
		return 1
	}
	defer closeOut()

	br := bufio.NewScanner(in)
	br.Buffer(make([]byte, 64*1024), 8*1024*1024)
	bw := bufio.NewWriter(out)
	defer func() { _ = bw.Flush() }()

	keyOf := func(s string) string {
		if ignoreCase {
			return strings.ToLower(s)
		}
		return s
	}
	var prev string
	var prevKey string
	var n int
	first := true
	flush := func() {
		if first {
			return
		}
		isDup := n > 1
		switch {
		case dupOnly && !isDup:
			return
		case uniqOnly && isDup:
			return
		}
		if count {
			_, _ = fmt.Fprintf(bw, "%7d %s\n", n, prev)
		} else {
			_, _ = fmt.Fprintf(bw, "%s\n", prev)
		}
	}

	for br.Scan() {
		line := br.Text()
		k := keyOf(line)
		if first {
			prev, prevKey, n, first = line, k, 1, false
			continue
		}
		if k == prevKey {
			n++
			continue
		}
		flush()
		prev, prevKey, n = line, k, 1
	}
	flush()
	if err := br.Err(); err != nil {
		ioutil.Errf("uniq: %v", err)
		return 1
	}
	return 0
}

func openIn(name string) (io.Reader, func(), error) {
	if name == "-" {
		return ioutil.Stdin, func() {}, nil
	}
	f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}

func openOut(name string) (io.Writer, func(), error) {
	if name == "-" {
		return ioutil.Stdout, func() {}, nil
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}
