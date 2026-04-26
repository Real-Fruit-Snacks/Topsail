// Package sort implements the `sort` applet: sort lines of files.
//
// This Wave 2 build supports the most common flags: -r, -n, -u, -f, -b
// and reading from multiple files. Complex key handling (-k, -t) is
// deferred to a later wave.
package sort

import (
	"bufio"
	"io"
	gosort "sort"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sort",
		Help:  "sort lines of text files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sort [OPTION]... [FILE]...
Write sorted concatenation of all FILE(s) to standard output.

Options:
  -r, --reverse           reverse the result of comparisons
  -n, --numeric-sort      compare according to string numerical value
  -u, --unique            output only the first of an equal run
  -f, --ignore-case       fold lower case to upper case characters
  -b, --ignore-leading-blanks   ignore leading blanks
  -k, --key=KEY           (not supported in this build)
  -t, --field-separator=SEP (not supported in this build)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var reverse, numeric, unique, foldCase, trimLeft bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-r", a == "--reverse":
			reverse = true
			args = args[1:]
		case a == "-n", a == "--numeric-sort":
			numeric = true
			args = args[1:]
		case a == "-u", a == "--unique":
			unique = true
			args = args[1:]
		case a == "-f", a == "--ignore-case":
			foldCase = true
			args = args[1:]
		case a == "-b", a == "--ignore-leading-blanks":
			trimLeft = true
			args = args[1:]
		case a == "-k", a == "-t":
			ioutil.Errf("sort: %s is not supported in this build", a)
			return 2
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'r':
					reverse = true
				case 'n':
					numeric = true
				case 'u':
					unique = true
				case 'f':
					foldCase = true
				case 'b':
					trimLeft = true
				default:
					ioutil.Errf("sort: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}
	var lines []string
	rc := 0
	for _, name := range files {
		ls, err := readLines(name)
		if err != nil {
			ioutil.Errf("sort: %s: %v", name, err)
			rc = 1
			continue
		}
		lines = append(lines, ls...)
	}

	cmpKey := func(s string) string {
		if trimLeft {
			s = strings.TrimLeft(s, " \t")
		}
		if foldCase {
			s = strings.ToUpper(s)
		}
		return s
	}

	gosort.SliceStable(lines, func(i, j int) bool {
		ai, aj := cmpKey(lines[i]), cmpKey(lines[j])
		var less bool
		if numeric {
			ni := parseLeadingFloat(ai)
			nj := parseLeadingFloat(aj)
			if ni != nj {
				less = ni < nj
			} else {
				less = ai < aj
			}
		} else {
			less = ai < aj
		}
		if reverse {
			return !less
		}
		return less
	})

	if unique {
		out := lines[:0]
		var prev string
		first := true
		for _, l := range lines {
			k := cmpKey(l)
			if first || k != prev {
				out = append(out, l)
				prev = k
				first = false
			}
		}
		lines = out
	}

	bw := bufio.NewWriter(ioutil.Stdout)
	defer func() { _ = bw.Flush() }()
	for _, l := range lines {
		_, _ = bw.WriteString(l)
		_, _ = bw.WriteString("\n")
	}
	return rc
}

func readLines(name string) ([]string, error) {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := openFile(name)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var out []string
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out, sc.Err()
}

// parseLeadingFloat extracts the leading numeric prefix of s and returns
// it as a float64. Used for -n comparisons; non-numeric input yields 0,
// which matches GNU sort's behavior of treating empty/garbage as zero.
func parseLeadingFloat(s string) float64 {
	s = strings.TrimLeft(s, " \t")
	end := 0
	if end < len(s) && (s[end] == '+' || s[end] == '-') {
		end++
	}
	for end < len(s) && (s[end] >= '0' && s[end] <= '9' || s[end] == '.') {
		end++
	}
	if end == 0 {
		return 0
	}
	n, err := strconv.ParseFloat(s[:end], 64)
	if err != nil {
		return 0
	}
	return n
}
