// Package tac implements the `tac` applet: print files in reverse line order.
package tac

import (
	"bufio"
	"io"
	"os"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "tac",
		Help:  "concatenate and print files in reverse line order",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tac [FILE]...
Write each FILE to standard output with lines in reverse order.
With no FILE, or when FILE is -, read standard input.
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
		if err := tacOne(name); err != nil {
			ioutil.Errf("tac: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func tacOne(name string) error {
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
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return err
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if _, err := io.WriteString(ioutil.Stdout, lines[i]+"\n"); err != nil {
			return err
		}
	}
	return nil
}
