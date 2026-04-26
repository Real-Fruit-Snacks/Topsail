// Package tr implements the `tr` applet: translate, delete, or squeeze characters.
package tr

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "tr",
		Help:  "translate, delete, or squeeze characters",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tr [OPTION]... SET1 [SET2]
Translate, squeeze, or delete characters from standard input.

Options:
  -c, --complement     use the complement of SET1
  -d, --delete         delete characters in SET1
  -s, --squeeze-repeats  squeeze each run of a character in SET1 (or SET2)
                       into a single instance

SETs accept:
  - literal characters
  - ranges (a-z)
  - escapes (\\n \\t \\r \\\\ \\ooo octal)
  - character classes [:alpha:] [:digit:] [:upper:] [:lower:] [:space:]
                      [:alnum:] [:punct:] [:print:] [:cntrl:] [:xdigit:]
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var complement, deleteMode, squeeze bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c", a == "--complement":
			complement = true
			args = args[1:]
		case a == "-d", a == "--delete":
			deleteMode = true
			args = args[1:]
		case a == "-s", a == "--squeeze-repeats":
			squeeze = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'c':
					complement = true
				case 'd':
					deleteMode = true
				case 's':
					squeeze = true
				default:
					ioutil.Errf("tr: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("tr: missing operand")
		return 2
	}
	if !deleteMode && len(args) < 2 && !squeeze {
		ioutil.Errf("tr: missing operand after %q", args[0])
		return 2
	}

	set1, err := expandSet(args[0])
	if err != nil {
		ioutil.Errf("tr: %v", err)
		return 2
	}
	var set2 []rune
	if len(args) >= 2 {
		set2, err = expandSet(args[1])
		if err != nil {
			ioutil.Errf("tr: %v", err)
			return 2
		}
	}
	if complement {
		set1 = complementOf(set1)
	}

	br := bufio.NewReader(ioutil.Stdin)
	bw := bufio.NewWriter(ioutil.Stdout)
	defer func() { _ = bw.Flush() }()

	var prev rune = -1
	for {
		r, _, err := br.ReadRune()
		if err == io.EOF {
			return 0
		}
		if err != nil {
			ioutil.Errf("tr: %v", err)
			return 1
		}
		out, drop := transform(r, set1, set2, deleteMode)
		if drop {
			prev = -1
			continue
		}
		if squeeze && out == prev && (deleteMode || inSet(out, set1) || (len(set2) > 0 && inSet(out, set2))) {
			continue
		}
		_, _ = bw.WriteRune(out)
		prev = out
	}
}

// transform returns (newRune, drop). If drop is true, the rune should be
// omitted from output entirely (delete mode).
func transform(r rune, set1, set2 []rune, deleteMode bool) (rune, bool) {
	if deleteMode {
		if inSet(r, set1) {
			return r, true
		}
		return r, false
	}
	if len(set2) == 0 {
		return r, false
	}
	for i, c := range set1 {
		if c == r {
			if i < len(set2) {
				return set2[i], false
			}
			// SET1 longer than SET2: pad with last char of SET2.
			return set2[len(set2)-1], false
		}
	}
	return r, false
}

func inSet(r rune, set []rune) bool {
	for _, c := range set {
		if c == r {
			return true
		}
	}
	return false
}

// expandSet expands ranges, escapes, and POSIX character classes into a
// flat list of runes.
func expandSet(s string) ([]rune, error) {
	var out []rune
	for i := 0; i < len(s); {
		// Character class: [:alpha:] etc.
		if i+1 < len(s) && s[i] == '[' && s[i+1] == ':' {
			end := strings.Index(s[i:], ":]")
			if end > 0 {
				name := s[i+2 : i+end]
				if cls := classRunes(name); cls != nil {
					out = append(out, cls...)
					i += end + 2
					continue
				}
			}
		}
		r, n, err := readChar(s[i:])
		if err != nil {
			return nil, err
		}
		// Range: c-c
		if i+n < len(s) && s[i+n] == '-' && i+n+1 < len(s) {
			r2, n2, err := readChar(s[i+n+1:])
			if err != nil {
				return nil, err
			}
			if r2 < r {
				return nil, &expandErr{msg: "invalid range"}
			}
			for c := r; c <= r2; c++ {
				out = append(out, c)
			}
			i += n + 1 + n2
			continue
		}
		out = append(out, r)
		i += n
	}
	return out, nil
}

type expandErr struct{ msg string }

func (e *expandErr) Error() string { return e.msg }

// readChar reads one possibly-escaped character.
func readChar(s string) (r rune, consumed int, err error) {
	if s == "" {
		return 0, 0, &expandErr{msg: "unexpected end of set"}
	}
	if s[0] != '\\' {
		decoded, size := utf8.DecodeRuneInString(s)
		return decoded, size, nil
	}
	if len(s) < 2 {
		return '\\', 1, nil
	}
	switch s[1] {
	case 'n':
		return '\n', 2, nil
	case 't':
		return '\t', 2, nil
	case 'r':
		return '\r', 2, nil
	case 'f':
		return '\f', 2, nil
	case 'v':
		return '\v', 2, nil
	case 'b':
		return '\b', 2, nil
	case '\\':
		return '\\', 2, nil
	case '/':
		return '/', 2, nil
	}
	// Octal escape: \NNN
	if s[1] >= '0' && s[1] <= '7' {
		end := 2
		for end < 4 && end < len(s) && s[end] >= '0' && s[end] <= '7' {
			end++
		}
		var n int
		for j := 1; j < end; j++ {
			n = n*8 + int(s[j]-'0')
		}
		return rune(n), end, nil
	}
	// Unknown escape: pass through literally.
	return rune(s[1]), 2, nil
}

func classRunes(name string) []rune {
	switch name {
	case "alpha":
		return rangeRunes('A', 'Z', 'a', 'z')
	case "upper":
		return rangeRunes('A', 'Z')
	case "lower":
		return rangeRunes('a', 'z')
	case "digit":
		return rangeRunes('0', '9')
	case "alnum":
		return rangeRunes('A', 'Z', 'a', 'z', '0', '9')
	case "xdigit":
		return rangeRunes('0', '9', 'A', 'F', 'a', 'f')
	case "space":
		return []rune{' ', '\t', '\n', '\v', '\f', '\r'}
	case "blank":
		return []rune{' ', '\t'}
	case "cntrl":
		out := make([]rune, 0, 33)
		for i := 0; i < 32; i++ {
			out = append(out, rune(i))
		}
		out = append(out, 127)
		return out
	case "punct":
		var out []rune
		for r := rune(33); r <= 126; r++ {
			if !isAlnum(r) {
				out = append(out, r)
			}
		}
		return out
	case "print":
		out := make([]rune, 0, 95)
		for r := rune(32); r <= 126; r++ {
			out = append(out, r)
		}
		return out
	case "graph":
		out := make([]rune, 0, 94)
		for r := rune(33); r <= 126; r++ {
			out = append(out, r)
		}
		return out
	}
	return nil
}

func isAlnum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func rangeRunes(pairs ...rune) []rune {
	var out []rune
	for i := 0; i+1 < len(pairs); i += 2 {
		for r := pairs[i]; r <= pairs[i+1]; r++ {
			out = append(out, r)
		}
	}
	return out
}

func complementOf(set []rune) []rune {
	in := make(map[rune]bool, len(set))
	for _, r := range set {
		in[r] = true
	}
	out := make([]rune, 0, 256)
	for r := rune(0); r < 256; r++ {
		if !in[r] {
			out = append(out, r)
		}
	}
	return out
}
