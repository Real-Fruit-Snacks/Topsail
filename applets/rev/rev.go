// Package rev implements the `rev` applet: reverse characters on each line.
package rev

import (
	"bufio"
	"io"
	"os"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "rev",
		Help:  "reverse characters on each line",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: rev [FILE]...
Reverse the characters on each line of FILE(s).
With no FILE, or when FILE is -, read standard input.

Operates on Unicode code points (runes), not bytes.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		args = []string{"-"}
	}

	rc := 0
	for _, name := range args {
		if err := revOne(name); err != nil {
			ioutil.Errf("rev: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func revOne(name string) error {
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
		runes := []rune(sc.Text())
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		if _, err := io.WriteString(ioutil.Stdout, string(runes)+"\n"); err != nil {
			return err
		}
	}
	return sc.Err()
}
