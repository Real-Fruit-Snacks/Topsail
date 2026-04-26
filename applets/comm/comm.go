// Package comm implements the `comm` applet: compare two sorted files.
package comm

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
		Name:  "comm",
		Help:  "compare two sorted files line by line",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: comm [OPTION]... FILE1 FILE2
Compare two sorted FILEs line by line, producing three columns of
output: lines unique to FILE1, lines unique to FILE2, and lines
common to both.

Options:
  -1   suppress column 1 (lines unique to FILE1)
  -2   suppress column 2 (lines unique to FILE2)
  -3   suppress column 3 (lines common to both)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var hide1, hide2, hide3 bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-1":
			hide1 = true
			args = args[1:]
		case a == "-2":
			hide2 = true
			args = args[1:]
		case a == "-3":
			hide3 = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Allow combined forms like -12, -13, -23.
			ok := true
			for _, c := range a[1:] {
				switch c {
				case '1':
					hide1 = true
				case '2':
					hide2 = true
				case '3':
					hide3 = true
				default:
					ok = false
				}
				if !ok {
					break
				}
			}
			if !ok {
				ioutil.Errf("comm: unknown option: %s", a)
				return 2
			}
			args = args[1:]
		default:
			stop = true
		}
	}
	if len(args) != 2 {
		ioutil.Errf("comm: comm needs exactly two filenames")
		return 2
	}

	r1, c1, err := openIn(args[0])
	if err != nil {
		ioutil.Errf("comm: %s: %v", args[0], err)
		return 1
	}
	defer c1()
	r2, c2, err := openIn(args[1])
	if err != nil {
		ioutil.Errf("comm: %s: %v", args[1], err)
		return 1
	}
	defer c2()

	s1 := bufio.NewScanner(r1)
	s2 := bufio.NewScanner(r2)
	s1.Buffer(make([]byte, 64*1024), 16*1024*1024)
	s2.Buffer(make([]byte, 64*1024), 16*1024*1024)

	var line1, line2 string
	have1 := s1.Scan()
	if have1 {
		line1 = s1.Text()
	}
	have2 := s2.Scan()
	if have2 {
		line2 = s2.Text()
	}

	emit := func(col int, line string) {
		if (col == 1 && hide1) || (col == 2 && hide2) || (col == 3 && hide3) {
			return
		}
		prefix := ""
		switch col {
		case 2:
			if !hide1 {
				prefix = "\t"
			}
		case 3:
			if !hide1 {
				prefix += "\t"
			}
			if !hide2 {
				prefix += "\t"
			}
		}
		_, _ = ioutil.Stdout.Write([]byte(prefix + line + "\n"))
	}

	for have1 && have2 {
		switch {
		case line1 < line2:
			emit(1, line1)
			have1 = s1.Scan()
			if have1 {
				line1 = s1.Text()
			}
		case line1 > line2:
			emit(2, line2)
			have2 = s2.Scan()
			if have2 {
				line2 = s2.Text()
			}
		default:
			emit(3, line1)
			have1 = s1.Scan()
			if have1 {
				line1 = s1.Text()
			}
			have2 = s2.Scan()
			if have2 {
				line2 = s2.Text()
			}
		}
	}
	for have1 {
		emit(1, line1)
		have1 = s1.Scan()
		if have1 {
			line1 = s1.Text()
		}
	}
	for have2 {
		emit(2, line2)
		have2 = s2.Scan()
		if have2 {
			line2 = s2.Text()
		}
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
