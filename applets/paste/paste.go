// Package paste implements the `paste` applet: merge lines of files.
package paste

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "paste",
		Help:  "merge lines of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: paste [OPTION]... [FILE]...
Write lines of FILE(s) merged side-by-side, separated by TAB.

Options:
  -d DELIMS   use characters from DELIMS as separators (cycles through)
  -s          paste each file's lines onto a single output line
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	delims := []rune{'\t'}
	var serial bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-d":
			if len(args) < 2 {
				ioutil.Errf("paste: option requires an argument -- 'd'")
				return 2
			}
			delims = []rune(args[1])
			if len(delims) == 0 {
				delims = []rune{'\t'}
			}
			args = args[2:]
		case a == "-s", a == "--serial":
			serial = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("paste: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	if serial {
		rc := 0
		for _, name := range files {
			if err := pasteSerial(name, delims); err != nil {
				ioutil.Errf("paste: %s: %v", name, err)
				rc = 1
			}
		}
		return rc
	}
	return pasteParallel(files, delims)
}

func pasteSerial(name string, delims []rune) error {
	r, closer, err := openIn(name)
	if err != nil {
		return err
	}
	defer closer()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	first := true
	di := 0
	for sc.Scan() {
		if !first {
			_, _ = ioutil.Stdout.Write([]byte(string(delims[di%len(delims)])))
			di++
		}
		_, _ = ioutil.Stdout.Write([]byte(sc.Text()))
		first = false
	}
	_, _ = ioutil.Stdout.Write([]byte("\n"))
	return sc.Err()
}

func pasteParallel(files []string, delims []rune) int {
	scs := make([]*bufio.Scanner, len(files))
	closers := make([]func(), len(files))
	rc := 0
	for i, name := range files {
		r, closer, err := openIn(name)
		if err != nil {
			ioutil.Errf("paste: %s: %v", name, err)
			return 1
		}
		closers[i] = closer
		scs[i] = bufio.NewScanner(r)
		scs[i].Buffer(make([]byte, 64*1024), 8*1024*1024)
	}
	defer func() {
		for _, c := range closers {
			c()
		}
	}()

	for {
		anyHit := false
		var parts []string
		for _, sc := range scs {
			if sc.Scan() {
				parts = append(parts, sc.Text())
				anyHit = true
			} else {
				parts = append(parts, "")
			}
		}
		if !anyHit {
			return rc
		}
		var b strings.Builder
		for i, p := range parts {
			if i > 0 {
				b.WriteRune(delims[(i-1)%len(delims)])
			}
			b.WriteString(p)
		}
		_, _ = ioutil.Stdout.Write([]byte(b.String() + "\n"))
	}
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
