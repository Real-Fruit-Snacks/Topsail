// Package head implements the `head` applet: output the first part of files.
package head

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
		Name:  "head",
		Help:  "output the first part of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: head [OPTION]... [FILE]...
Print the first 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header.
With no FILE, or when FILE is -, read standard input.

Options:
  -n N, --lines=N    print the first N lines instead of 10
  -c N, --bytes=N    print the first N bytes
  -q, --quiet        never print headers
  -v, --verbose      always print headers
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	nLines, nBytes := 10, 0
	var byBytes, quiet, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-n":
			if len(args) < 2 {
				ioutil.Errf("head: option requires an argument -- 'n'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("head: invalid number of lines: %s", args[1])
				return 2
			}
			nLines = n
			byBytes = false
			args = args[2:]
		case strings.HasPrefix(a, "--lines="):
			n, err := strconv.Atoi(a[len("--lines="):])
			if err != nil || n < 0 {
				ioutil.Errf("head: invalid number of lines: %s", a)
				return 2
			}
			nLines = n
			byBytes = false
			args = args[1:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("head: option requires an argument -- 'c'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("head: invalid number of bytes: %s", args[1])
				return 2
			}
			nBytes = n
			byBytes = true
			args = args[2:]
		case strings.HasPrefix(a, "--bytes="):
			n, err := strconv.Atoi(a[len("--bytes="):])
			if err != nil || n < 0 {
				ioutil.Errf("head: invalid number of bytes: %s", a)
				return 2
			}
			nBytes = n
			byBytes = true
			args = args[1:]
		case a == "-q", a == "--quiet", a == "--silent":
			quiet = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Allow -<digits> shorthand: -5 == -n 5
			if n, err := strconv.Atoi(a[1:]); err == nil && n >= 0 {
				nLines = n
				byBytes = false
				args = args[1:]
				continue
			}
			ioutil.Errf("head: invalid option: %s", a)
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
		if err := headOne(name, nLines, nBytes, byBytes); err != nil {
			ioutil.Errf("head: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func headOne(name string, nLines, nBytes int, byBytes bool) error {
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
		_, err := io.CopyN(ioutil.Stdout, r, int64(nBytes))
		if err == io.EOF {
			return nil
		}
		return err
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	count := 0
	for sc.Scan() && count < nLines {
		if _, err := io.WriteString(ioutil.Stdout, sc.Text()+"\n"); err != nil {
			return err
		}
		count++
	}
	return sc.Err()
}

func labelFor(name string) string {
	if name == "-" {
		return "standard input"
	}
	return name
}
