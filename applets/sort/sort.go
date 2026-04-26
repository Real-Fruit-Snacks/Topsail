// Package sort implements the `sort` applet: sort lines of files.
//
// Beyond the simple flag set (-r, -n, -u, -f, -b), this build supports
// repeatable -k field keys and -t custom separators. Per-key option
// suffixes (e.g. `-k 2nr`) override the global flags for that key.
//
// Out of scope for this build: in-field character offsets (-k 1.3,1.5),
// stable per-key ordering distinct from the global -s flag, locale-aware
// collation, and -h/-V/-V human/version comparators.
package sort

import (
	"bufio"
	"fmt"
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
  -r, --reverse                reverse the result of comparisons
  -n, --numeric-sort           compare according to string numerical value
  -u, --unique                 output only the first of an equal run
  -f, --ignore-case            fold lower case to upper case characters
  -b, --ignore-leading-blanks  ignore leading blanks
  -k, --key=KEYDEF             sort by a key; KEYDEF is F1[OPTS][,F2[OPTS]]
                               (F1, F2 are field numbers, OPTS are nrfb)
  -t, --field-separator=SEP    use SEP (single character) as the field separator
                               (default: runs of blanks)

Multiple -k keys are honored in order; the first non-equal comparison wins.
`

type keyFlags struct {
	numeric, reverse, foldCase, trimLeft bool
	set                                  bool // any of the above explicitly set
}

type keySpec struct {
	start, end int // 1-indexed field numbers; end == 0 means "to end of line"
	flags      keyFlags
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		reverse, numeric, unique, foldCase, trimLeft bool
		keys                                         []keySpec
		sep                                          byte
		hasSep                                       bool
	)

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
		case a == "-k":
			if len(args) < 2 {
				ioutil.Errf("sort: option requires an argument -- 'k'")
				return 2
			}
			k, err := parseKey(args[1])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			keys = append(keys, k)
			args = args[2:]
		case strings.HasPrefix(a, "-k") && len(a) > 2:
			k, err := parseKey(a[2:])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			keys = append(keys, k)
			args = args[1:]
		case strings.HasPrefix(a, "--key="):
			k, err := parseKey(a[len("--key="):])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			keys = append(keys, k)
			args = args[1:]
		case a == "-t":
			if len(args) < 2 {
				ioutil.Errf("sort: option requires an argument -- 't'")
				return 2
			}
			s, err := parseSep(args[1])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			sep = s
			hasSep = true
			args = args[2:]
		case strings.HasPrefix(a, "-t") && len(a) > 2:
			s, err := parseSep(a[2:])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			sep = s
			hasSep = true
			args = args[1:]
		case strings.HasPrefix(a, "--field-separator="):
			s, err := parseSep(a[len("--field-separator="):])
			if err != nil {
				ioutil.Errf("sort: %v", err)
				return 2
			}
			sep = s
			hasSep = true
			args = args[1:]
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

	globals := keyFlags{
		numeric:  numeric,
		reverse:  reverse,
		foldCase: foldCase,
		trimLeft: trimLeft,
	}

	cmp := func(a, b string) int {
		if len(keys) == 0 {
			return compareWhole(a, b, globals)
		}
		for _, k := range keys {
			ka := extractFieldKey(a, k, sep, hasSep)
			kb := extractFieldKey(b, k, sep, hasSep)
			flags := k.flags
			if !flags.set {
				flags = globals
			}
			if c := compareStrings(ka, kb, flags); c != 0 {
				return c
			}
		}
		return 0
	}

	gosort.SliceStable(lines, func(i, j int) bool {
		return cmp(lines[i], lines[j]) < 0
	})

	if unique {
		out := lines[:0]
		first := true
		var prev string
		for _, l := range lines {
			if first || cmp(prev, l) != 0 {
				out = append(out, l)
				prev = l
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

func parseSep(s string) (byte, error) {
	if len(s) != 1 {
		return 0, fmt.Errorf("the separator must be a single character: %q", s)
	}
	return s[0], nil
}

// parseKey parses a `-k F1[OPTS][,F2[OPTS]]` argument.
func parseKey(s string) (keySpec, error) {
	if s == "" {
		return keySpec{}, fmt.Errorf("empty key")
	}
	parts := strings.SplitN(s, ",", 2)
	startN, startFlags, err := parseKeyPart(parts[0])
	if err != nil {
		return keySpec{}, err
	}
	k := keySpec{start: startN, flags: startFlags}
	if len(parts) == 2 {
		endN, endFlags, err := parseKeyPart(parts[1])
		if err != nil {
			return keySpec{}, err
		}
		k.end = endN
		// End-side flags take precedence (matches GNU's "last one wins").
		if endFlags.set {
			k.flags = endFlags
		}
	}
	return k, nil
}

func parseKeyPart(s string) (int, keyFlags, error) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, keyFlags{}, fmt.Errorf("invalid key field: %q", s)
	}
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0, keyFlags{}, err
	}
	if n < 1 {
		return 0, keyFlags{}, fmt.Errorf("field numbers start at 1: %q", s)
	}
	rest := s[end:]
	var f keyFlags
	for _, c := range rest {
		switch c {
		case 'n':
			f.numeric = true
			f.set = true
		case 'r':
			f.reverse = true
			f.set = true
		case 'f':
			f.foldCase = true
			f.set = true
		case 'b':
			f.trimLeft = true
			f.set = true
		case '.':
			return 0, keyFlags{}, fmt.Errorf("character offsets in -k are not supported in this build")
		default:
			return 0, keyFlags{}, fmt.Errorf("invalid key option: %q", string(c))
		}
	}
	return n, f, nil
}

// extractFieldKey returns the substring of line spanning fields
// [k.start, k.end] under the given separator.
func extractFieldKey(line string, k keySpec, sep byte, hasSep bool) string {
	var fields []string
	if hasSep {
		fields = strings.Split(line, string(sep))
	} else {
		fields = strings.Fields(line)
	}
	if k.start > len(fields) {
		return ""
	}
	from := k.start - 1
	if from < 0 {
		from = 0
	}
	to := len(fields)
	if k.end > 0 && k.end < to {
		to = k.end
	}
	if to <= from {
		return ""
	}
	join := " "
	if hasSep {
		join = string(sep)
	}
	return strings.Join(fields[from:to], join)
}

func compareStrings(a, b string, f keyFlags) int {
	if f.trimLeft {
		a = strings.TrimLeft(a, " \t")
		b = strings.TrimLeft(b, " \t")
	}
	if f.foldCase {
		a = strings.ToUpper(a)
		b = strings.ToUpper(b)
	}
	var c int
	if f.numeric {
		na := parseLeadingFloat(a)
		nb := parseLeadingFloat(b)
		switch {
		case na < nb:
			c = -1
		case na > nb:
			c = 1
		default:
			c = strings.Compare(a, b)
		}
	} else {
		c = strings.Compare(a, b)
	}
	if f.reverse {
		c = -c
	}
	return c
}

func compareWhole(a, b string, f keyFlags) int {
	return compareStrings(a, b, f)
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
