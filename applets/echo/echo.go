// Package echo implements the `echo` applet: write arguments to stdout.
package echo

import (
	"io"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "echo",
		Help:  "write arguments to stdout",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: echo [-neE] [STRING]...
Echo the STRING(s) to standard output.

Options:
  -n     do not output the trailing newline
  -e     enable interpretation of backslash escapes
  -E     disable interpretation of backslash escapes (default)

If -e is in effect, these escape sequences are recognized:
  \\     backslash
  \a     alert (BEL)
  \b     backspace
  \c     suppress further output (including the trailing newline)
  \f     form feed
  \n     new line
  \r     carriage return
  \t     horizontal tab
  \v     vertical tab
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var noNewline, interpret bool

	// Parse leading flags. echo is unusual: an unrecognized flag-like
	// argument terminates parsing and is treated as data, matching
	// GNU echo behavior.
	for len(args) > 0 {
		a := args[0]
		if len(a) < 2 || a[0] != '-' {
			break
		}
		valid := true
		for _, c := range a[1:] {
			if c != 'n' && c != 'e' && c != 'E' {
				valid = false
				break
			}
		}
		if !valid {
			break
		}
		for _, c := range a[1:] {
			switch c {
			case 'n':
				noNewline = true
			case 'e':
				interpret = true
			case 'E':
				interpret = false
			}
		}
		args = args[1:]
	}

	out := strings.Join(args, " ")
	if interpret {
		var stop bool
		out, stop = expandEscapes(out)
		if stop {
			noNewline = true
		}
	}
	if !noNewline {
		out += "\n"
	}
	_, _ = io.WriteString(ioutil.Stdout, out)
	return 0
}

// expandEscapes processes backslash sequences. The bool return is true
// if a \c sequence was encountered, signaling "stop and emit no further
// output (including the trailing newline)."
func expandEscapes(s string) (string, bool) {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		i++
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case 'a':
			b.WriteByte('\a')
		case 'b':
			b.WriteByte('\b')
		case 'c':
			return b.String(), true
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'v':
			b.WriteByte('\v')
		default:
			// Unknown escape: emit literally including the backslash.
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String(), false
}
