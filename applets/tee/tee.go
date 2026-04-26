// Package tee implements the `tee` applet: copy stdin to stdout and files.
package tee

import (
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "tee",
		Help:  "read from stdin and write to stdout and files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tee [OPTION]... [FILE]...
Copy standard input to standard output AND to each FILE.

Options:
  -a, --append      append to each FILE rather than overwriting
  -i, --ignore-interrupts   (accepted; signals are not handled in this build)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var appendMode bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-a", a == "--append":
			appendMode = true
			args = args[1:]
		case a == "-i", a == "--ignore-interrupts":
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'a':
					appendMode = true
				case 'i':
					// no-op
				default:
					ioutil.Errf("tee: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if appendMode {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	writers := []io.Writer{ioutil.Stdout}
	files := make([]*os.File, 0, len(args))
	rc := 0
	for _, name := range args {
		f, err := os.OpenFile(name, flag, 0o644) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			ioutil.Errf("tee: %s: %v", name, err)
			rc = 1
			continue
		}
		files = append(files, f)
		writers = append(writers, f)
	}
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()

	mw := io.MultiWriter(writers...)
	if _, err := io.Copy(mw, ioutil.Stdin); err != nil {
		ioutil.Errf("tee: %v", err)
		rc = 1
	}
	return rc
}
