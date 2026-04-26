// Package sum implements the legacy `sum` applet (BSD checksum).
package sum

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sum",
		Help:  "compute legacy BSD/SysV checksums and block counts",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sum [OPTION]... [FILE]...
Print BSD-style 16-bit checksums and 1024-byte block counts.

Options:
  -s, --sysv    use System V checksum (32-bit) instead of BSD
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	sysv := false

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-s", a == "--sysv":
			sysv = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("sum: unknown option: %s", a)
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
		if err := sumOne(name, sysv); err != nil {
			ioutil.Errf("sum: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func sumOne(name string, sysv bool) error {
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
	buf := make([]byte, 4096)
	var size int64
	if sysv {
		var s uint32
		for {
			n, err := r.Read(buf)
			for i := 0; i < n; i++ {
				s += uint32(buf[i])
			}
			size += int64(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		s = (s & 0xFFFF) + (s >> 16)
		s = (s & 0xFFFF) + (s >> 16)
		blocks := (size + 511) / 512
		_, _ = fmt.Fprintf(ioutil.Stdout, "%d %d", s, blocks)
	} else {
		var s uint16
		for {
			n, err := r.Read(buf)
			for i := 0; i < n; i++ {
				s = (s >> 1) | (s << 15)
				s += uint16(buf[i])
			}
			size += int64(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		blocks := (size + 1023) / 1024
		_, _ = fmt.Fprintf(ioutil.Stdout, "%05d %d", s, blocks)
	}
	if name != "-" {
		_, _ = fmt.Fprintf(ioutil.Stdout, " %s", name)
	}
	_, _ = ioutil.Stdout.Write([]byte("\n"))
	return nil
}
