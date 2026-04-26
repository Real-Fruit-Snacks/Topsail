// Package cut implements the `cut` applet: extract sections from each line.
package cut

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
		Name:  "cut",
		Help:  "extract sections from each line of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: cut OPTION... [FILE]...
Print selected parts of lines from each FILE to standard output.

Mode (exactly one):
  -b LIST   select only these bytes
  -c LIST   select only these characters (rune-aware)
  -f LIST   select only these fields (default delimiter: TAB)

Options:
  -d DELIM  use DELIM (one byte) instead of TAB as the field delimiter
  -s        with -f, suppress lines that don't contain the delimiter
  --output-delimiter=STRING   use STRING in place of the input delimiter

LIST is a comma-separated list of ranges (1, 1-3, 5-, -2).
`

type mode int

const (
	modeBytes mode = iota
	modeChars
	modeFields
)

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		listStr      string
		modeSet      bool
		m            mode
		delim        byte = '\t'
		outDelim     string
		outDelimSet  bool
		onlyDelimSet bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-b":
			if len(args) < 2 {
				ioutil.Errf("cut: option requires an argument -- 'b'")
				return 2
			}
			listStr, m, modeSet = args[1], modeBytes, true
			args = args[2:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("cut: option requires an argument -- 'c'")
				return 2
			}
			listStr, m, modeSet = args[1], modeChars, true
			args = args[2:]
		case a == "-f":
			if len(args) < 2 {
				ioutil.Errf("cut: option requires an argument -- 'f'")
				return 2
			}
			listStr, m, modeSet = args[1], modeFields, true
			args = args[2:]
		case a == "-d":
			if len(args) < 2 || len(args[1]) != 1 {
				ioutil.Errf("cut: the delimiter must be a single character")
				return 2
			}
			delim = args[1][0]
			args = args[2:]
		case a == "-s":
			onlyDelimSet = true
			args = args[1:]
		case strings.HasPrefix(a, "--output-delimiter="):
			outDelim = a[len("--output-delimiter="):]
			outDelimSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("cut: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if !modeSet {
		ioutil.Errf("cut: you must specify a list of bytes, characters, or fields")
		return 2
	}
	ranges, err := parseList(listStr)
	if err != nil {
		ioutil.Errf("cut: %v", err)
		return 2
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	odelim := string(delim)
	if outDelimSet {
		odelim = outDelim
	}

	rc := 0
	for _, name := range files {
		if err := cutOne(name, m, ranges, delim, odelim, onlyDelimSet); err != nil {
			ioutil.Errf("cut: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

type rng struct {
	lo, hi int // 1-based, inclusive; hi==0 means "to end"
}

func parseList(s string) ([]rng, error) {
	if s == "" {
		return nil, fmt.Errorf("invalid list")
	}
	var out []rng
	for _, part := range strings.Split(s, ",") {
		if part == "" {
			return nil, fmt.Errorf("invalid range")
		}
		var r rng
		if i := strings.Index(part, "-"); i >= 0 {
			lo, hi := part[:i], part[i+1:]
			if lo == "" {
				r.lo = 1
			} else {
				n, err := strconv.Atoi(lo)
				if err != nil || n < 1 {
					return nil, fmt.Errorf("invalid range: %q", part)
				}
				r.lo = n
			}
			if hi == "" {
				r.hi = 0 // open-ended
			} else {
				n, err := strconv.Atoi(hi)
				if err != nil || n < 1 {
					return nil, fmt.Errorf("invalid range: %q", part)
				}
				r.hi = n
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 {
				return nil, fmt.Errorf("invalid number: %q", part)
			}
			r.lo, r.hi = n, n
		}
		out = append(out, r)
	}
	return out, nil
}

func includes(ranges []rng, idx int) bool {
	for _, r := range ranges {
		if idx >= r.lo && (r.hi == 0 || idx <= r.hi) {
			return true
		}
	}
	return false
}

func cutOne(name string, m mode, ranges []rng, delim byte, outDelim string, onlyDelim bool) error {
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
		var emitted string
		switch m {
		case modeBytes:
			emitted = cutByIndex([]byte(line), ranges)
		case modeChars:
			emitted = cutByRune([]rune(line), ranges)
		case modeFields:
			fields := strings.Split(line, string(delim))
			if len(fields) == 1 {
				if onlyDelim {
					continue
				}
				emitted = line
				break
			}
			var keep []string
			for i, fld := range fields {
				if includes(ranges, i+1) {
					keep = append(keep, fld)
				}
			}
			emitted = strings.Join(keep, outDelim)
		}
		if _, err := io.WriteString(ioutil.Stdout, emitted+"\n"); err != nil {
			return err
		}
	}
	return sc.Err()
}

func cutByIndex(b []byte, ranges []rng) string {
	var out []byte
	for i := range b {
		if includes(ranges, i+1) {
			out = append(out, b[i])
		}
	}
	return string(out)
}

func cutByRune(rs []rune, ranges []rng) string {
	var out []rune
	for i := range rs {
		if includes(ranges, i+1) {
			out = append(out, rs[i])
		}
	}
	return string(out)
}
