// Package sed implements a useful subset of `sed`: addresses, ranges,
// multi-command scripts, and the s/d/p/q commands.
//
// Out of scope for this build: hold space (h, H, g, G), labels and
// branches (b, t, :), append/insert/change (a, i, c), text-input (r,
// w), and y/// transliteration. Anchored at GNU sed's command grammar
// where they overlap.
package sed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sed",
		Help:  "stream editor (s, d, p, q with addresses)",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sed [OPTION]... SCRIPT [FILE]...
       sed [OPTION]... -e SCRIPT [-e SCRIPT]... [FILE]...

Apply SCRIPT to each line of FILE(s) (or stdin). Multiple commands may
be separated by ';' or newlines.

Supported commands (each may be prefixed by an address or address range):
  s/PATTERN/REPLACEMENT/[gi]   substitute (g = all occurrences, i = case-insensitive)
  d                            delete the pattern space (skip auto-print)
  p                            print the pattern space (in addition to auto-print)
  q                            quit (after auto-print of the current line)

Addresses:
  N            line number N
  $            the last line
  /REGEX/      lines matching REGEX (RE2 syntax)
  ADDR1,ADDR2  range from ADDR1 to ADDR2 (inclusive)
  ADDR!CMD     apply CMD to lines NOT matching ADDR

Options:
  -e, --expression=SCRIPT   add SCRIPT to the commands (combinable)
  -E, --regexp-extended     parity with GNU sed; RE2 is the default
  -n, --quiet               suppress automatic printing of the pattern space
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var quiet bool
	var scripts []string

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
		case a == "-e", a == "--expression":
			if len(args) < 2 {
				ioutil.Errf("sed: option requires an argument -- 'e'")
				return 2
			}
			scripts = append(scripts, args[1])
			args = args[2:]
		case strings.HasPrefix(a, "--expression="):
			scripts = append(scripts, a[len("--expression="):])
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("sed: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	var files []string
	if len(scripts) == 0 {
		if len(args) == 0 {
			ioutil.Errf("sed: missing script")
			return 2
		}
		scripts = []string{args[0]}
		files = args[1:]
	} else {
		files = args
	}

	cmds, err := parseScript(strings.Join(scripts, "\n"))
	if err != nil {
		ioutil.Errf("sed: %v", err)
		return 2
	}

	if len(files) == 0 {
		files = []string{"-"}
	}
	rc := 0
	for _, name := range files {
		stopAll, err := sedOne(name, cmds, quiet)
		if err != nil {
			ioutil.Errf("sed: %s: %v", name, err)
			rc = 1
		}
		if stopAll {
			break
		}
	}
	return rc
}

// addrKind identifies the type of an address selector.
type addrKind int

const (
	addrNone addrKind = iota
	addrLine
	addrLast
	addrRegex
)

// addr describes one side of a command's address (or address range).
type addr struct {
	kind addrKind
	line int
	re   *regexp.Regexp
}

// command is one parsed sed command with its (possibly empty) address
// selector. inRange tracks range-mode state across input lines and is
// the only mutable field during execution.
type command struct {
	a1, a2  *addr
	negate  bool
	op      byte    // 's', 'd', 'p', 'q'
	sub     *subCmd // populated for op == 's'
	inRange bool
}

type subCmd struct {
	re          *regexp.Regexp
	replacement string
	global      bool
}

// parseScript splits a script on ';' or '\n' and parses each command.
// Splitting respects backslash escapes and the body delimiter of an s
// command so things like 's/;/X/' don't get misinterpreted.
func parseScript(script string) ([]*command, error) {
	parts := splitCommands(script)
	out := make([]*command, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		c, err := parseCommand(p)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty script")
	}
	return out, nil
}

// splitCommands walks s and slices on ';' or '\n', skipping any byte
// inside a regex address (/.../) or an s command body. Backslash
// escapes are honored everywhere.
func splitCommands(s string) []string {
	var parts []string
	var cur []byte

	flush := func() {
		parts = append(parts, string(cur))
		cur = cur[:0]
	}

	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s):
			cur = append(cur, c, s[i+1])
			i += 2
		case c == ';' || c == '\n':
			flush()
			i++
		case c == '/':
			// Regex address: scan to the closing '/'.
			cur = append(cur, c)
			i++
			for i < len(s) {
				if s[i] == '\\' && i+1 < len(s) {
					cur = append(cur, s[i], s[i+1])
					i += 2
					continue
				}
				cur = append(cur, s[i])
				if s[i] == '/' {
					i++
					break
				}
				i++
			}
		case c == 's' && i+1 < len(s) && isDelim(s[i+1]):
			// s command body: capture s + 3 delim-bounded fields + flags.
			cur = append(cur, c)
			i++
			delim := s[i]
			cur = append(cur, delim)
			i++
			fields := 0
			for i < len(s) && fields < 2 {
				if s[i] == '\\' && i+1 < len(s) {
					cur = append(cur, s[i], s[i+1])
					i += 2
					continue
				}
				cur = append(cur, s[i])
				if s[i] == delim {
					fields++
				}
				i++
			}
			// Now flags up to the next ';', '\n', or end.
			for i < len(s) && s[i] != ';' && s[i] != '\n' {
				cur = append(cur, s[i])
				i++
			}
		default:
			cur = append(cur, c)
			i++
		}
	}
	flush()
	return parts
}

func isDelim(b byte) bool {
	// Any non-alphanumeric byte that's not whitespace can be a sed delimiter.
	if b == '\n' || b == ';' || b == ' ' || b == '\t' {
		return false
	}
	return true
}

// parseCommand parses "[addr1[,addr2][!]] op [args]".
func parseCommand(s string) (*command, error) {
	c := &command{}
	rest := s

	// First address.
	a1, after, err := parseAddr(rest)
	if err != nil {
		return nil, err
	}
	if a1 != nil {
		c.a1 = a1
		rest = after
		// Optional second address.
		if rest != "" && rest[0] == ',' {
			rest = rest[1:]
			a2, after, err := parseAddr(rest)
			if err != nil {
				return nil, err
			}
			if a2 == nil {
				return nil, fmt.Errorf("missing second address")
			}
			c.a2 = a2
			rest = after
		}
	}
	rest = strings.TrimLeft(rest, " \t")
	if rest != "" && rest[0] == '!' {
		c.negate = true
		rest = strings.TrimLeft(rest[1:], " \t")
	}
	if rest == "" {
		return nil, fmt.Errorf("missing command")
	}

	op := rest[0]
	body := rest[1:]
	switch op {
	case 's':
		sc, err := parseSubBody(body)
		if err != nil {
			return nil, err
		}
		c.op = 's'
		c.sub = sc
	case 'd', 'p', 'q':
		c.op = op
		// Trailing characters are ignored to match sed's tolerance.
	default:
		return nil, fmt.Errorf("unsupported command: %q", string(op))
	}
	return c, nil
}

// parseAddr parses one address from the front of s and returns the
// address (or nil if none) and the remainder.
func parseAddr(s string) (*addr, string, error) {
	if s == "" {
		return nil, s, nil
	}
	if s[0] == '$' {
		return &addr{kind: addrLast}, s[1:], nil
	}
	if s[0] >= '0' && s[0] <= '9' {
		end := 0
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		n, err := strconv.Atoi(s[:end])
		if err != nil {
			return nil, s, err
		}
		return &addr{kind: addrLine, line: n}, s[end:], nil
	}
	if s[0] == '/' {
		// Find the closing '/', honoring backslash escapes.
		i := 1
		for i < len(s) {
			if s[i] == '\\' && i+1 < len(s) {
				i += 2
				continue
			}
			if s[i] == '/' {
				break
			}
			i++
		}
		if i >= len(s) {
			return nil, s, fmt.Errorf("unterminated regex address")
		}
		pat := s[1:i]
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, s, fmt.Errorf("address regex: %w", err)
		}
		return &addr{kind: addrRegex, re: re}, s[i+1:], nil
	}
	return nil, s, nil
}

// parseSubBody parses the body of an s command: delim PAT delim REPL delim FLAGS.
func parseSubBody(body string) (*subCmd, error) {
	if body == "" {
		return nil, fmt.Errorf("missing delimiter for s")
	}
	delim := body[0]
	rest := body[1:]
	parts := splitDelim(rest, delim)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed s command")
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
	return &subCmd{re: re, replacement: convertReplacement(repl), global: global}, nil
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

// convertReplacement rewrites sed-style \1..\9 to Go's $1..$9 and & to
// $0, escaping any literal $ in input so Go's regexp.Expand doesn't
// reinterpret it.
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
			b.WriteByte(n)
			i++
			continue
		}
		if c == '$' {
			b.WriteString("$$")
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// matches reports whether c's address selector applies to (line, text)
// at line position (1-indexed). isLast is true for the final line.
func (c *command) matches(line int, text string, isLast bool) bool {
	hit := c.rawMatch(line, text, isLast)
	if c.negate {
		hit = !hit
	}
	return hit
}

func (c *command) rawMatch(line int, text string, isLast bool) bool {
	if c.a1 == nil {
		return true
	}
	if c.a2 == nil {
		return c.a1.matches(line, text, isLast)
	}
	if c.inRange {
		// Already in range: a2 may close it (still apply on the closing line).
		if c.a2.matches(line, text, isLast) {
			c.inRange = false
		}
		return true
	}
	if c.a1.matches(line, text, isLast) {
		// Range opens. Whether to also close on this same line depends on
		// a2: a line-number a2 closes only if line >= a2.line; a regex a2
		// closes if it matches the same line. We mimic GNU's "if a2 also
		// matches on this line, close immediately" for non-line a2.
		if c.a2.matches(line, text, isLast) {
			return true
		}
		c.inRange = true
		return true
	}
	return false
}

func (a *addr) matches(line int, text string, isLast bool) bool {
	switch a.kind {
	case addrLine:
		return line == a.line
	case addrLast:
		return isLast
	case addrRegex:
		return a.re.MatchString(text)
	}
	return false
}

// sedOne reads name (or stdin) and applies cmds. The bool return is
// true when a 'q' command was executed and the caller should stop
// processing further input files.
func sedOne(name string, cmds []*command, quiet bool) (bool, error) {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path
		if err != nil {
			return false, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)

	// Peek-ahead so we know which line is "$".
	var prev string
	have := false
	lineNo := 0
	stopAll := false
	for sc.Scan() {
		if have {
			lineNo++
			stop, err := applyLine(cmds, prev, lineNo, false, quiet)
			if err != nil {
				return false, err
			}
			if stop {
				return true, nil
			}
		}
		prev = sc.Text()
		have = true
	}
	if err := sc.Err(); err != nil {
		return false, err
	}
	if have {
		lineNo++
		stop, err := applyLine(cmds, prev, lineNo, true, quiet)
		if err != nil {
			return false, err
		}
		if stop {
			stopAll = true
		}
	}
	return stopAll, nil
}

// applyLine walks every command for one input line, applying those
// whose address matches. It returns whether 'q' was triggered.
func applyLine(cmds []*command, line string, lineNo int, isLast, quiet bool) (bool, error) {
	pat := line
	deleted := false
	quit := false

	for _, c := range cmds {
		if !c.matches(lineNo, pat, isLast) {
			continue
		}
		switch c.op {
		case 's':
			pat = applySub(pat, c.sub)
		case 'd':
			deleted = true
			// "d" terminates the cycle for this line per sed semantics.
			break //nolint:staticcheck // explicit cycle break
		case 'p':
			if _, err := fmt.Fprintln(ioutil.Stdout, pat); err != nil {
				return false, err
			}
		case 'q':
			quit = true
		}
		if deleted {
			break
		}
	}

	if !deleted && !quiet {
		if _, err := fmt.Fprintln(ioutil.Stdout, pat); err != nil {
			return false, err
		}
	}
	return quit, nil
}

func applySub(line string, sc *subCmd) string {
	if sc.global {
		return sc.re.ReplaceAllString(line, sc.replacement)
	}
	loc := sc.re.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return line[:loc[0]] +
		sc.re.ReplaceAllString(line[loc[0]:loc[1]], sc.replacement) +
		line[loc[1]:]
}
