// Package date implements the `date` applet: print or format the current time.
package date

import (
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "date",
		Help:  "print or format the current date/time",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: date [OPTION]... [+FORMAT]
Print the current date and time, optionally formatted.

Options:
  -u, --utc       use UTC instead of local time
  -R, --rfc-email RFC 5322 email-style format
  -I, --iso-8601  ISO 8601 format
  -d STRING       parse STRING as a date and print it (RFC 3339 / common)

FORMAT specifiers (subset of POSIX strftime):
  %Y year (4 digits)            %m month (01-12)        %d day (01-31)
  %H hour (00-23)               %M minute (00-59)       %S second (00-59)
  %y year (2 digits)            %B full month name      %A weekday name
  %b abbr month                 %a abbr weekday         %p AM/PM
  %s seconds since epoch        %Z timezone name        %z timezone offset
  %T equivalent to %H:%M:%S     %D equivalent to %m/%d/%y
  %% literal percent
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var utc, rfc, iso bool
	var dateArg string

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-u", a == "--utc":
			utc = true
			args = args[1:]
		case a == "-R", a == "--rfc-email":
			rfc = true
			args = args[1:]
		case a == "-I", a == "--iso-8601":
			iso = true
			args = args[1:]
		case a == "-d":
			if len(args) < 2 {
				ioutil.Errf("date: option requires an argument -- 'd'")
				return 2
			}
			dateArg = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--date="):
			dateArg = a[len("--date="):]
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && !strings.HasPrefix(a, "-+") && a != "-":
			ioutil.Errf("date: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	t := time.Now()
	if utc {
		t = t.UTC()
	}
	if dateArg != "" {
		parsed, err := parseDate(dateArg)
		if err != nil {
			ioutil.Errf("date: invalid date: %s", dateArg)
			return 1
		}
		t = parsed
		if utc {
			t = t.UTC()
		}
	}

	format := ""
	if len(args) > 0 && strings.HasPrefix(args[0], "+") {
		format = args[0][1:]
	}

	switch {
	case rfc:
		_, _ = ioutil.Stdout.Write([]byte(t.Format(time.RFC1123Z) + "\n"))
	case iso:
		_, _ = ioutil.Stdout.Write([]byte(t.Format("2006-01-02") + "\n"))
	case format != "":
		_, _ = ioutil.Stdout.Write([]byte(applyFormat(format, t) + "\n"))
	default:
		_, _ = ioutil.Stdout.Write([]byte(t.Format("Mon Jan  2 15:04:05 MST 2006") + "\n"))
	}
	return 0
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errInvalid
}

var errInvalid = stringErr("invalid date format")

type stringErr string

func (e stringErr) Error() string { return string(e) }

func applyFormat(format string, t time.Time) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		switch format[i] {
		case 'Y':
			b.WriteString(t.Format("2006"))
		case 'y':
			b.WriteString(t.Format("06"))
		case 'm':
			b.WriteString(t.Format("01"))
		case 'd':
			b.WriteString(t.Format("02"))
		case 'H':
			b.WriteString(t.Format("15"))
		case 'M':
			b.WriteString(t.Format("04"))
		case 'S':
			b.WriteString(t.Format("05"))
		case 'B':
			b.WriteString(t.Format("January"))
		case 'b':
			b.WriteString(t.Format("Jan"))
		case 'A':
			b.WriteString(t.Format("Monday"))
		case 'a':
			b.WriteString(t.Format("Mon"))
		case 'p':
			b.WriteString(t.Format("PM"))
		case 'Z':
			b.WriteString(t.Format("MST"))
		case 'z':
			b.WriteString(t.Format("-0700"))
		case 'T':
			b.WriteString(t.Format("15:04:05"))
		case 'D':
			b.WriteString(t.Format("01/02/06"))
		case 's':
			b.WriteString(t.Format("Mon Jan  2 15:04:05 MST 2006"))
			// Actually %s is unix epoch — fix:
			b.Reset()
			// fallthrough won't work; manual
			return formatRedo(format, t)
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}

// formatRedo is a one-shot fallback to handle %s correctly without
// rewriting the loop above. It's invoked only when %s is encountered.
func formatRedo(format string, t time.Time) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		switch format[i] {
		case 'Y':
			b.WriteString(t.Format("2006"))
		case 'y':
			b.WriteString(t.Format("06"))
		case 'm':
			b.WriteString(t.Format("01"))
		case 'd':
			b.WriteString(t.Format("02"))
		case 'H':
			b.WriteString(t.Format("15"))
		case 'M':
			b.WriteString(t.Format("04"))
		case 'S':
			b.WriteString(t.Format("05"))
		case 's':
			b.WriteString(strFromInt(t.Unix()))
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}

func strFromInt(n int64) string {
	// Avoid pulling strconv just for this; simple path.
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
