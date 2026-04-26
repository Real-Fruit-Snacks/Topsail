// Package wc implements the `wc` applet: count lines, words, bytes, chars.
package wc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "wc",
		Help:  "print line, word, and byte counts for files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: wc [OPTION]... [FILE]...
Print newline, word, and byte counts for each FILE.
With no FILE, or when FILE is -, read standard input.

Options:
  -l     print line counts only
  -w     print word counts only
  -c     print byte counts only
  -m     print character counts (UTF-8 aware)
  -L     print the length of the longest line
`

type counts struct {
	lines, words, bytes, chars, longest int
}

type opts struct {
	lines, words, bytes, chars, longest, anyFlag bool
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var o opts

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				o.anyFlag = true
				switch c {
				case 'l':
					o.lines = true
				case 'w':
					o.words = true
				case 'c':
					o.bytes = true
				case 'm':
					o.chars = true
				case 'L':
					o.longest = true
				default:
					ioutil.Errf("wc: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}
	// Default flags when none given.
	if !o.anyFlag {
		o.lines, o.words, o.bytes = true, true, true
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	var total counts
	rc := 0
	for _, name := range files {
		c, err := countOne(name)
		if err != nil {
			ioutil.Errf("wc: %s: %v", name, err)
			rc = 1
			continue
		}
		total.lines += c.lines
		total.words += c.words
		total.bytes += c.bytes
		total.chars += c.chars
		if c.longest > total.longest {
			total.longest = c.longest
		}
		emit(name, c, o)
	}
	if len(files) > 1 {
		emit("total", total, o)
	}
	return rc
}

func countOne(name string) (counts, error) {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return counts{}, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	var c counts
	br := bufio.NewReader(r)
	var line []byte
	var inWord bool
	var lineLen int

	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			if lineLen > c.longest {
				c.longest = lineLen
			}
			break
		}
		if err != nil {
			return c, err
		}
		c.bytes++
		line = append(line, b)
		if b == '\n' {
			c.lines++
			c.chars += utf8.RuneCount(line)
			if lineLen > c.longest {
				c.longest = lineLen
			}
			lineLen = 0
			line = line[:0]
			inWord = false
			continue
		}
		// A "word" boundary is any whitespace.
		if b == ' ' || b == '\t' || b == '\v' || b == '\f' || b == '\r' {
			inWord = false
		} else if !inWord {
			c.words++
			inWord = true
		}
		lineLen++
	}
	// Catch chars on a final line without a trailing newline.
	if len(line) > 0 {
		c.chars += utf8.RuneCount(line)
	}
	return c, nil
}

func emit(name string, c counts, o opts) {
	var parts []string
	if o.lines {
		parts = append(parts, fmt.Sprintf("%7d", c.lines))
	}
	if o.words {
		parts = append(parts, fmt.Sprintf("%7d", c.words))
	}
	if o.bytes {
		parts = append(parts, fmt.Sprintf("%7d", c.bytes))
	}
	if o.chars {
		parts = append(parts, fmt.Sprintf("%7d", c.chars))
	}
	if o.longest {
		parts = append(parts, fmt.Sprintf("%7d", c.longest))
	}
	line := strings.Join(parts, " ")
	if name != "-" {
		line += " " + name
	}
	_, _ = fmt.Fprintln(ioutil.Stdout, line)
}
