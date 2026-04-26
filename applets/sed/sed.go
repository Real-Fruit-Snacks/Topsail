// Package sed implements a minimal `sed` applet: substitution only.
//
// This Wave 3 build supports the `s/PATTERN/REPLACEMENT/FLAGS` command
// (with -e, -i deferred). Address selectors, multi-command scripts, and
// the full sed command set are deferred to a later wave.
package sed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sed",
		Help:  "stream editor (s/// substitution only in this build)",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sed [OPTION]... 's/PATTERN/REPLACEMENT/[FLAGS]' [FILE]...
Apply a substitution command to each line of FILE(s) (or stdin).

This Wave 3 build supports only the substitution command:

  s/PATTERN/REPLACEMENT/[gi]
    g  replace all occurrences on each line (default: first only)
    i  case-insensitive match

Options:
  -E, --regexp-extended   treat PATTERN as ERE (RE2 is the default — flag is parity)
  -n, --quiet             suppress automatic printing (only print 'p' substitutions)

Replacement supports & (whole match) and \1-\9 (capture groups).
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var quiet bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-E", a == "--regexp-extended":
			args = args[1:]
		case a == "-n", a == "--quiet", a == "--silent":
			quiet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("sed: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("sed: missing script")
		return 2
	}
	script := args[0]
	files := args[1:]

	cmd, err := parseScript(script)
	if err != nil {
		ioutil.Errf("sed: %v", err)
		return 2
	}

	if len(files) == 0 {
		files = []string{"-"}
	}
	rc := 0
	for _, name := range files {
		if err := sedOne(name, cmd, quiet); err != nil {
			ioutil.Errf("sed: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

type subCmd struct {
	re          *regexp.Regexp
	replacement string
	global      bool
}

// parseScript handles a single s/PATTERN/REPLACEMENT/FLAGS command.
func parseScript(script string) (*subCmd, error) {
	if !strings.HasPrefix(script, "s") {
		return nil, fmt.Errorf("only 's/.../.../' is supported in this build")
	}
	if len(script) < 2 {
		return nil, fmt.Errorf("malformed substitute command")
	}
	delim := script[1]
	rest := script[2:]
	parts := splitDelim(rest, delim)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed substitute command")
	}
	pat := parts[0]
	repl := parts[1]
	flags := ""
	if len(parts) >= 3 {
		flags = parts[2]
	}

	caseI := strings.Contains(flags, "i")
	global := strings.Contains(flags, "g")
	if caseI {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	// Convert sed-style \1..\9 to Go's $1..$9. & stays via $0.
	repl = convertReplacement(repl)
	return &subCmd{re: re, replacement: repl, global: global}, nil
}

// splitDelim splits s by delim, respecting backslash-escapes.
func splitDelim(s string, delim byte) []string {
	var parts []string
	var cur []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			cur = append(cur, s[i], s[i+1])
			i++
			continue
		}
		if s[i] == delim {
			parts = append(parts, string(cur))
			cur = cur[:0]
			continue
		}
		cur = append(cur, s[i])
	}
	parts = append(parts, string(cur))
	return parts
}

func convertReplacement(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '&' {
			b.WriteString("${0}")
			continue
		}
		if c == '\\' && i+1 < len(s) {
			n := s[i+1]
			if n >= '0' && n <= '9' {
				b.WriteString("${")
				b.WriteByte(n)
				b.WriteString("}")
				i++
				continue
			}
			// Pass through escape literally.
			b.WriteByte(n)
			i++
			continue
		}
		// Escape literal $ in input so Go's Expand doesn't interpret it.
		if c == '$' {
			b.WriteString("$$")
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func sedOne(name string, cmd *subCmd, quiet bool) error {
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
		var newLine string
		if cmd.global {
			newLine = cmd.re.ReplaceAllString(line, cmd.replacement)
		} else {
			loc := cmd.re.FindStringIndex(line)
			if loc != nil {
				newLine = line[:loc[0]] +
					cmd.re.ReplaceAllString(line[loc[0]:loc[1]], cmd.replacement) +
					line[loc[1]:]
			} else {
				newLine = line
			}
		}
		if !quiet {
			_, _ = fmt.Fprintln(ioutil.Stdout, newLine)
		}
	}
	return sc.Err()
}
