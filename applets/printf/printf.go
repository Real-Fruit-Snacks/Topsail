// Package printf implements the `printf` applet: formatted output.
//
// The format string is reused if there are more arguments than directives,
// matching POSIX/GNU printf semantics. Unknown numeric input is treated
// as 0 with a stderr warning, and the process exits 1 at the end.
package printf

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "printf",
		Help:  "format and print arguments",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: printf FORMAT [ARGUMENT]...
Print ARGUMENT(s) according to FORMAT.

Conversion directives:
  %s            string
  %d, %i        signed decimal integer
  %u            unsigned decimal integer
  %o            unsigned octal integer
  %x, %X        unsigned hexadecimal integer (lower / upper case)
  %c            single character (first byte of arg)
  %b            string with backslash escapes interpreted
  %q            quoted string
  %%            literal percent sign

Width and precision: %[-]?[0]?[N][.M][diouxXs]

Backslash escapes in FORMAT:
  \\  \a  \b  \f  \n  \r  \t  \v  \0NNN (octal)

If there are more ARGUMENTs than directives, FORMAT is reused.
If fewer, missing values default to 0 (numeric) or "" (string).
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) == 0 {
		ioutil.Errf("printf: missing operand")
		return 2
	}
	format := args[0]
	args = args[1:]

	rc := 0
	first := true
	for first || len(args) > 0 {
		first = false
		consumed, hadDirective, errs := emit(ioutil.Stdout, format, args)
		args = args[consumed:]
		if errs > 0 {
			rc = 1
		}
		// If FORMAT is purely literal there's no point looping; otherwise
		// reuse the format string until args are exhausted.
		if !hadDirective {
			break
		}
	}
	return rc
}

// emit walks format once, consuming up to len(args) values.
//
//	consumed     - number of args used
//	hadDirective - true if format contained at least one %... directive
//	errs         - count of conversion errors
func emit(w io.Writer, format string, args []string) (consumed int, hadDirective bool, errs int) {
	for i := 0; i < len(format); {
		c := format[i]
		if c == '\\' && i+1 < len(format) {
			esc, n := decodeEscape(format[i+1:])
			_, _ = io.WriteString(w, esc)
			i += 1 + n
			continue
		}
		if c != '%' {
			_, _ = w.Write([]byte{c})
			i++
			continue
		}
		// %... directive
		j := i + 1
		// flags
		for j < len(format) && (format[j] == '-' || format[j] == '+' ||
			format[j] == ' ' || format[j] == '#' || format[j] == '0') {
			j++
		}
		// width
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			j++
		}
		// precision
		if j < len(format) && format[j] == '.' {
			j++
			for j < len(format) && format[j] >= '0' && format[j] <= '9' {
				j++
			}
		}
		if j >= len(format) {
			ioutil.Errf("printf: %s: invalid conversion specification", format[i:])
			errs++
			return consumed, hadDirective, errs
		}
		verb := format[j]
		spec := format[i : j+1]
		i = j + 1

		if verb == '%' {
			_, _ = io.WriteString(w, "%")
			continue
		}
		hadDirective = true
		var arg string
		if consumed < len(args) {
			arg = args[consumed]
			consumed++
		}

		switch verb {
		case 's':
			_, _ = fmt.Fprintf(w, spec, arg)
		case 'd', 'i':
			n, err := strconv.ParseInt(strings.TrimSpace(arg), 0, 64)
			if err != nil && arg != "" {
				ioutil.Errf("printf: %s: invalid number", arg)
				errs++
			}
			if verb == 'i' {
				spec = spec[:len(spec)-1] + "d"
			}
			_, _ = fmt.Fprintf(w, spec, n)
		case 'u':
			n, err := strconv.ParseUint(strings.TrimSpace(arg), 0, 64)
			if err != nil && arg != "" {
				ioutil.Errf("printf: %s: invalid number", arg)
				errs++
			}
			spec = spec[:len(spec)-1] + "d"
			_, _ = fmt.Fprintf(w, spec, n)
		case 'o', 'x', 'X':
			n, err := strconv.ParseUint(strings.TrimSpace(arg), 0, 64)
			if err != nil && arg != "" {
				ioutil.Errf("printf: %s: invalid number", arg)
				errs++
			}
			_, _ = fmt.Fprintf(w, spec, n)
		case 'c':
			if arg != "" {
				_, _ = w.Write([]byte{arg[0]})
			}
		case 'b':
			var sb strings.Builder
			for k := 0; k < len(arg); k++ {
				if arg[k] == '\\' && k+1 < len(arg) {
					esc, n := decodeEscape(arg[k+1:])
					sb.WriteString(esc)
					k += n
					continue
				}
				sb.WriteByte(arg[k])
			}
			_, _ = io.WriteString(w, sb.String())
		case 'q':
			_, _ = io.WriteString(w, strconv.Quote(arg))
		default:
			ioutil.Errf("printf: %%%c: invalid directive", verb)
			errs++
		}
	}
	return consumed, hadDirective, errs
}

// decodeEscape decodes a backslash escape, given the byte(s) following the
// leading backslash. It returns the decoded string and the number of source
// bytes consumed.
func decodeEscape(src string) (decoded string, consumed int) {
	if src == "" {
		return `\`, 0
	}
	switch src[0] {
	case '\\':
		return `\`, 1
	case 'a':
		return "\a", 1
	case 'b':
		return "\b", 1
	case 'f':
		return "\f", 1
	case 'n':
		return "\n", 1
	case 'r':
		return "\r", 1
	case 't':
		return "\t", 1
	case 'v':
		return "\v", 1
	case '0':
		// up to 3 octal digits after \0
		end := 1
		for end < 4 && end < len(src) && src[end] >= '0' && src[end] <= '7' {
			end++
		}
		n, err := strconv.ParseUint(src[1:end], 8, 8)
		if err != nil {
			return "\x00", 1
		}
		return string([]byte{byte(n)}), end
	default:
		// Unknown escape: pass through with backslash.
		return `\` + string(src[0]), 1
	}
}
