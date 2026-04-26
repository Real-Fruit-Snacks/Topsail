// Package tail implements the `tail` applet: output the last part of files.
//
// The follow mode (-f) is intentionally deferred to a later wave; this build
// supports the static read-all-then-print-tail flavor that covers the
// overwhelming majority of pipeline use cases.
package tail

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
		Name:  "tail",
		Help:  "output the last part of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tail [OPTION]... [FILE]...
Print the last 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header.
With no FILE, or when FILE is -, read standard input.

Options:
  -n N, --lines=N    print the last N lines instead of 10 (or +N for "starting at line N")
  -c N, --bytes=N    print the last N bytes
  -q, --quiet        never print headers
  -v, --verbose      always print headers

The follow option (-f) is not implemented in this build.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	nLines := 10
	nBytes := 0
	var byBytes, fromStart, quiet, verbose bool

	parseN := func(s string) (n int, fs, ok bool) {
		if strings.HasPrefix(s, "+") {
			fs = true
			s = s[1:]
		} else if strings.HasPrefix(s, "-") {
			s = s[1:]
		}
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			return 0, false, false
		}
		return v, fs, true
	}

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-n":
			if len(args) < 2 {
				ioutil.Errf("tail: option requires an argument -- 'n'")
				return 2
			}
			n, fs, ok := parseN(args[1])
			if !ok {
				ioutil.Errf("tail: invalid number of lines: %s", args[1])
				return 2
			}
			nLines = n
			fromStart = fs
			byBytes = false
			args = args[2:]
		case strings.HasPrefix(a, "--lines="):
			n, fs, ok := parseN(a[len("--lines="):])
			if !ok {
				ioutil.Errf("tail: invalid number of lines: %s", a)
				return 2
			}
			nLines = n
			fromStart = fs
			byBytes = false
			args = args[1:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("tail: option requires an argument -- 'c'")
				return 2
			}
			n, fs, ok := parseN(args[1])
			if !ok {
				ioutil.Errf("tail: invalid number of bytes: %s", args[1])
				return 2
			}
			nBytes = n
			fromStart = fs
			byBytes = true
			args = args[2:]
		case strings.HasPrefix(a, "--bytes="):
			n, fs, ok := parseN(a[len("--bytes="):])
			if !ok {
				ioutil.Errf("tail: invalid number of bytes: %s", a)
				return 2
			}
			nBytes = n
			fromStart = fs
			byBytes = true
			args = args[1:]
		case a == "-q", a == "--quiet", a == "--silent":
			quiet = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case a == "-f", a == "--follow":
			ioutil.Errf("tail: -f is not implemented in this build")
			return 2
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Allow -<digits> shorthand: -5 == -n 5
			if n, err := strconv.Atoi(a[1:]); err == nil && n >= 0 {
				nLines = n
				byBytes = false
				args = args[1:]
				continue
			}
			ioutil.Errf("tail: invalid option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}
	showHeader := (len(files) > 1 && !quiet) || verbose

	rc := 0
	for i, name := range files {
		if showHeader {
			if i > 0 {
				_, _ = fmt.Fprintln(ioutil.Stdout)
			}
			_, _ = fmt.Fprintf(ioutil.Stdout, "==> %s <==\n", labelFor(name))
		}
		if err := tailOne(name, nLines, nBytes, byBytes, fromStart); err != nil {
			ioutil.Errf("tail: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func tailOne(name string, nLines, nBytes int, byBytes, fromStart bool) error {
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

	if byBytes {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if fromStart {
			start := nBytes - 1
			if start < 0 {
				start = 0
			}
			if start > len(data) {
				return nil
			}
			_, err = ioutil.Stdout.Write(data[start:])
			return err
		}
		if nBytes >= len(data) {
			_, err = ioutil.Stdout.Write(data)
		} else {
			_, err = ioutil.Stdout.Write(data[len(data)-nBytes:])
		}
		return err
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	if fromStart {
		idx := 0
		for sc.Scan() {
			idx++
			if idx >= nLines {
				if _, err := io.WriteString(ioutil.Stdout, sc.Text()+"\n"); err != nil {
					return err
				}
			}
		}
		return sc.Err()
	}

	// Ring buffer of the last nLines lines.
	buf := make([]string, 0, nLines)
	for sc.Scan() {
		if len(buf) < nLines {
			buf = append(buf, sc.Text())
		} else {
			copy(buf, buf[1:])
			buf[len(buf)-1] = sc.Text()
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	for _, line := range buf {
		if _, err := io.WriteString(ioutil.Stdout, line+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func labelFor(name string) string {
	if name == "-" {
		return "standard input"
	}
	return name
}
