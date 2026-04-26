// Package grep implements the `grep` applet: search files for patterns.
//
// Uses Go's regexp package (RE2) — see ARCHITECTURE.md for documented
// divergence from POSIX BRE.
package grep

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:    "grep",
		Aliases: []string{"egrep", "fgrep"},
		Help:    "search files for a regex pattern",
		Usage:   usage,
		Main:    Main,
	})
}

const usage = `Usage: grep [OPTION]... PATTERN [FILE]...
Search each FILE for lines matching PATTERN. With no FILE, or when
FILE is -, read standard input.

Options:
  -i, --ignore-case      case-insensitive match
  -v, --invert-match     select non-matching lines
  -c, --count            print only a count of matching lines per file
  -n, --line-number      prefix each match with its line number
  -l, --files-with-matches  print only the names of files with matches
  -L, --files-without-match  print only the names of files with NO matches
  -H, --with-filename    always print filename headers
  -h, --no-filename      never print filename headers
  -E, --extended-regexp  treat PATTERN as a regex (the RE2 default — present for parity)
  -F, --fixed-strings    treat PATTERN as a literal string
  -r, -R, --recursive    descend into directories
  -q, --quiet            no output; exit 0 on first match
  --                     end of options

Exit status: 0 if any line matched, 1 if none, 2 on error.
`

type opts struct {
	ignoreCase, invert, count, lineNum, filesOnly, filesNone bool
	withFilename, noFilename, recursive, quiet, fixed        bool
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var o opts
	// egrep is a parity alias — RE2 is already extended, no flag toggle
	// is needed. fgrep flips the fixed-string mode by default.
	if argv[0] == "fgrep" {
		o.fixed = true
	}

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-i", a == "--ignore-case":
			o.ignoreCase = true
			args = args[1:]
		case a == "-v", a == "--invert-match":
			o.invert = true
			args = args[1:]
		case a == "-c", a == "--count":
			o.count = true
			args = args[1:]
		case a == "-n", a == "--line-number":
			o.lineNum = true
			args = args[1:]
		case a == "-l", a == "--files-with-matches":
			o.filesOnly = true
			args = args[1:]
		case a == "-L", a == "--files-without-match":
			o.filesNone = true
			args = args[1:]
		case a == "-H", a == "--with-filename":
			o.withFilename = true
			args = args[1:]
		case a == "-h", a == "--no-filename":
			o.noFilename = true
			args = args[1:]
		case a == "-E", a == "--extended-regexp":
			args = args[1:]
		case a == "-F", a == "--fixed-strings":
			o.fixed = true
			args = args[1:]
		case a == "-r", a == "-R", a == "--recursive":
			o.recursive = true
			args = args[1:]
		case a == "-q", a == "--quiet", a == "--silent":
			o.quiet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ok := true
			for _, c := range a[1:] {
				switch c {
				case 'i':
					o.ignoreCase = true
				case 'v':
					o.invert = true
				case 'c':
					o.count = true
				case 'n':
					o.lineNum = true
				case 'l':
					o.filesOnly = true
				case 'L':
					o.filesNone = true
				case 'H':
					o.withFilename = true
				case 'h':
					o.noFilename = true
				case 'E':
				case 'F':
					o.fixed = true
				case 'r', 'R':
					o.recursive = true
				case 'q':
					o.quiet = true
				default:
					ioutil.Errf("grep: invalid option -- '%c'", c)
					return 2
				}
			}
			if !ok {
				return 2
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("grep: missing pattern")
		return 2
	}
	pat := args[0]
	files := args[1:]
	if o.fixed {
		pat = regexp.QuoteMeta(pat)
	}
	if o.ignoreCase {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		ioutil.Errf("grep: %v", err)
		return 2
	}

	if len(files) == 0 {
		files = []string{"-"}
	}
	showHeader := o.withFilename || (len(files) > 1 && !o.noFilename) || o.recursive

	matched := false
	rc := 1
	for _, name := range files {
		if o.recursive && name != "-" {
			info, err := os.Stat(name)
			if err == nil && info.IsDir() {
				err := filepath.WalkDir(name, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if d.IsDir() {
						return nil
					}
					if grepOne(path, re, o, true) {
						matched = true
					}
					return nil
				})
				if err != nil {
					ioutil.Errf("grep: %s: %v", name, err)
					rc = 2
				}
				continue
			}
		}
		if grepOne(name, re, o, showHeader) {
			matched = true
		}
	}
	if matched {
		rc = 0
	}
	return rc
}

func grepOne(name string, re *regexp.Regexp, o opts, showHeader bool) bool {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			ioutil.Errf("grep: %s: %v", name, err)
			return false
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	matches := 0
	lineNo := 0
	for sc.Scan() {
		lineNo++
		match := re.MatchString(sc.Text())
		if o.invert {
			match = !match
		}
		if !match {
			continue
		}
		matches++
		if o.quiet || o.filesOnly || o.filesNone {
			continue
		}
		if o.count {
			continue
		}
		if showHeader && name != "-" {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s:", name)
		}
		if o.lineNum {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%d:", lineNo)
		}
		_, _ = fmt.Fprintln(ioutil.Stdout, sc.Text())
	}

	switch {
	case o.filesOnly && matches > 0:
		_, _ = fmt.Fprintln(ioutil.Stdout, name)
	case o.filesNone && matches == 0:
		_, _ = fmt.Fprintln(ioutil.Stdout, name)
	case o.count:
		if showHeader && name != "-" {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s:", name)
		}
		_, _ = fmt.Fprintln(ioutil.Stdout, matches)
	}
	return matches > 0
}
