// Package shuf implements the `shuf` applet: shuffle lines.
package shuf

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "shuf",
		Help:  "generate a random permutation of input lines",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: shuf [OPTION]... [FILE]
Output a random permutation of the lines of FILE (or stdin).

Options:
  -n N, --head-count=N   output at most N lines
  -e, --echo             treat each ARG as an input line (instead of reading)
  -i LO-HI, --input-range=LO-HI   shuffle integers from LO to HI inclusive
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	headCount := -1
	var echoArgs bool
	rangeLo, rangeHi := 0, 0
	useRange := false

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-n":
			if len(args) < 2 {
				ioutil.Errf("shuf: option requires an argument -- 'n'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("shuf: invalid -n value: %s", args[1])
				return 2
			}
			headCount = n
			args = args[2:]
		case a == "-e", a == "--echo":
			echoArgs = true
			args = args[1:]
		case a == "-i":
			if len(args) < 2 {
				ioutil.Errf("shuf: option requires an argument -- 'i'")
				return 2
			}
			lo, hi, ok := parseRange(args[1])
			if !ok {
				ioutil.Errf("shuf: invalid range: %s", args[1])
				return 2
			}
			rangeLo, rangeHi, useRange = lo, hi, true
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("shuf: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	var lines []string
	switch {
	case useRange:
		for i := rangeLo; i <= rangeHi; i++ {
			lines = append(lines, strconv.Itoa(i))
		}
	case echoArgs:
		lines = args
	default:
		var r io.Reader
		r = ioutil.Stdin
		if len(args) > 0 && args[0] != "-" {
			f, err := os.Open(args[0]) //nolint:gosec // user-supplied path is the whole point
			if err != nil {
				ioutil.Errf("shuf: %s: %v", args[0], err)
				return 1
			}
			defer func() { _ = f.Close() }()
			r = f
		}
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
		for sc.Scan() {
			lines = append(lines, sc.Text())
		}
	}

	shuffle(lines)
	if headCount >= 0 && headCount < len(lines) {
		lines = lines[:headCount]
	}
	for _, l := range lines {
		_, _ = ioutil.Stdout.Write([]byte(l + "\n"))
	}
	return 0
}

func parseRange(s string) (lo, hi int, ok bool) {
	i := strings.Index(s, "-")
	if i < 0 {
		return 0, 0, false
	}
	a, err1 := strconv.Atoi(s[:i])
	b, err2 := strconv.Atoi(s[i+1:])
	if err1 != nil || err2 != nil || a > b {
		return 0, 0, false
	}
	return a, b, true
}

// shuffle uses crypto/rand for portability across platforms; performance
// is not the goal here.
func shuffle(s []string) {
	for i := len(s) - 1; i > 0; i-- {
		var b [8]byte
		_, _ = rand.Read(b[:])
		j := int(binary.LittleEndian.Uint64(b[:]) % uint64(i+1))
		s[i], s[j] = s[j], s[i]
	}
}
