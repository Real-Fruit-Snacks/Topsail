// Package fold implements the `fold` applet: wrap each line to width.
package fold

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "fold",
		Help:  "wrap each input line to fit a width",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: fold [OPTION]... [FILE]...
Wrap each input line to fit a fixed width.

Options:
  -w COLS, --width=COLS   use COLS columns instead of 80
  -s, --spaces            break at spaces (don't break inside words)
  -b, --bytes             count bytes rather than runes
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	width := 80
	var spaces, byBytes bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-w":
			if len(args) < 2 {
				ioutil.Errf("fold: option requires an argument -- 'w'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("fold: invalid width: %s", args[1])
				return 2
			}
			width = n
			args = args[2:]
		case strings.HasPrefix(a, "--width="):
			n, err := strconv.Atoi(a[len("--width="):])
			if err != nil || n < 1 {
				ioutil.Errf("fold: invalid width")
				return 2
			}
			width = n
			args = args[1:]
		case a == "-s", a == "--spaces":
			spaces = true
			args = args[1:]
		case a == "-b", a == "--bytes":
			byBytes = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("fold: unknown option: %s", a)
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
	for _, name := range files {
		if err := foldOne(name, width, spaces, byBytes); err != nil {
			ioutil.Errf("fold: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func foldOne(name string, width int, spaces, byBytes bool) error {
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
		var seq []string
		if byBytes {
			seq = wrapBytes(line, width, spaces)
		} else {
			seq = wrapRunes(line, width, spaces)
		}
		for _, p := range seq {
			_, _ = ioutil.Stdout.Write([]byte(p + "\n"))
		}
	}
	return sc.Err()
}

func wrapRunes(line string, width int, atSpaces bool) []string {
	rs := []rune(line)
	var out []string
	for len(rs) > width {
		breakAt := width
		if atSpaces {
			for i := width - 1; i > 0; i-- {
				if rs[i] == ' ' || rs[i] == '\t' {
					breakAt = i + 1
					break
				}
			}
		}
		out = append(out, string(rs[:breakAt]))
		rs = rs[breakAt:]
	}
	out = append(out, string(rs))
	return out
}

func wrapBytes(line string, width int, atSpaces bool) []string {
	b := []byte(line)
	var out []string
	for len(b) > width {
		breakAt := width
		if atSpaces {
			for i := width - 1; i > 0; i-- {
				if b[i] == ' ' || b[i] == '\t' {
					breakAt = i + 1
					break
				}
			}
		}
		out = append(out, string(b[:breakAt]))
		b = b[breakAt:]
	}
	out = append(out, string(b))
	return out
}
