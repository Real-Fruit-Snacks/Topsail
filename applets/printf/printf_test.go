package printf

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func capture(t *testing.T, argv []string) (out, errBuf string, rc int) {
	t.Helper()
	o, e := testutil.CaptureStdio(t)
	rc = Main(argv)
	return o.String(), e.String(), rc
}

func TestPrintfBasic(t *testing.T) {
	out, _, rc := capture(t, []string{"printf", "hello world"})
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if out != "hello world" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfNewline(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", `hello\n`})
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfString(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%s\n", "world"})
	if out != "world\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfDecimal(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%d\n", "42"})
	if out != "42\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfHex(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%x\n", "255"})
	if out != "ff\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfHexUpper(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%X\n", "255"})
	if out != "FF\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfOctal(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%o\n", "8"})
	if out != "10\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfWidth(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "[%5d]\n", "42"})
	if out != "[   42]\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfLeftAlign(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "[%-5s]", "x"})
	if out != "[x    ]" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfZeroPad(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%05d", "42"})
	if out != "00042" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfMultipleArgs(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%s=%d ", "a", "1", "b", "2"})
	if out != "a=1 b=2 " {
		t.Errorf("out = %q (format reuse failed)", out)
	}
}

func TestPrintfPercentLiteral(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%d%%\n", "50"})
	if out != "50%\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfChar(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%c%c%c\n", "abc", "xyz", "123"})
	if out != "ax1\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfQ(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%q\n", `hello "world"`})
	if !strings.HasPrefix(out, `"`) {
		t.Errorf("out = %q; expected quoted", out)
	}
}

func TestPrintfBPercent(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%b\n", `tab\there`})
	if out != "tab\there\n" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfInvalidNumber(t *testing.T) {
	out, errBuf, rc := capture(t, []string{"printf", "%d\n", "abc"})
	if rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if out != "0\n" {
		t.Errorf("out = %q", out)
	}
	if !strings.Contains(errBuf, "invalid number") {
		t.Errorf("stderr = %q", errBuf)
	}
}

func TestPrintfMissingOperand(t *testing.T) {
	_, errBuf, rc := capture(t, []string{"printf"})
	if rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf, "missing operand") {
		t.Errorf("stderr = %q", errBuf)
	}
}

func TestPrintfBackslashEscapes(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`a\tb`, "a\tb"},
		{`a\nb`, "a\nb"},
		{`a\\b`, `a\b`},
		{`\0101`, "A"}, // octal 101 = 'A'
	}
	for _, tc := range cases {
		out, _, _ := capture(t, []string{"printf", tc.in})
		if out != tc.want {
			t.Errorf("printf %q = %q; want %q", tc.in, out, tc.want)
		}
	}
}

func TestPrintfFewerArgsThanDirectives(t *testing.T) {
	out, _, _ := capture(t, []string{"printf", "%s/%s/%s", "a"})
	// Missing args default to "".
	if out != "a//" {
		t.Errorf("out = %q", out)
	}
}

func TestPrintfUnknownDirective(t *testing.T) {
	_, errBuf, rc := capture(t, []string{"printf", "%z", "x"})
	if rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf, "invalid directive") {
		t.Errorf("stderr = %q", errBuf)
	}
}
