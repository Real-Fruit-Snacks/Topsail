package printf

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestPrintfDecodeEscapeAll covers every branch of decodeEscape.
func TestPrintfDecodeEscapeAll(t *testing.T) {
	cases := []struct {
		src      string
		want     string
		consumed int
	}{
		{"\\", "\\", 1},
		{"a", "\a", 1},
		{"b", "\b", 1},
		{"f", "\f", 1},
		{"n", "\n", 1},
		{"r", "\r", 1},
		{"t", "\t", 1},
		{"v", "\v", 1},
		{"0101", "A", 4}, // octal
		{"0", "\x00", 1}, // bare \0
		{"x", "\\x", 1},  // unknown -> passes through with backslash
		{"", "\\", 0},    // empty src
	}
	for _, tc := range cases {
		got, n := decodeEscape(tc.src)
		if got != tc.want {
			t.Errorf("decodeEscape(%q) = %q; want %q", tc.src, got, tc.want)
		}
		if n != tc.consumed {
			t.Errorf("decodeEscape(%q) consumed = %d; want %d", tc.src, n, tc.consumed)
		}
	}
}

// TestPrintfDirectiveU covers the %u branch of emit.
func TestPrintfDirectiveU(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"printf", "%u\n", "42"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "42\n" {
		t.Errorf("got %q", got)
	}
}

// TestPrintfDirectiveI covers the %i (alias for %d) branch.
func TestPrintfDirectiveI(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"printf", "%i\n", "-7"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "-7\n" {
		t.Errorf("got %q", got)
	}
}

// TestPrintfInvalidUnsigned covers the %u / %o / %x error path.
func TestPrintfInvalidUnsigned(t *testing.T) {
	out, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"printf", "%u/%o/%x\n", "abc", "abc", "abc"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid number") {
		t.Errorf("stderr = %q", errBuf.String())
	}
	if got := out.String(); got != "0/0/0\n" {
		t.Errorf("got %q (zeros expected on parse failure)", got)
	}
}

// TestPrintfMissingTrailingPercent triggers the "invalid conversion
// specification" error path (a bare '%' at end of format). emit logs
// the error, increments errs, and returns; Main turns errs>0 into rc=1.
func TestPrintfMissingTrailingPercent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"printf", "tail %"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid conversion") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestPrintfBackslashAtEnd verifies a trailing backslash is preserved
// without invoking decodeEscape (covers the "no second char" branch).
func TestPrintfBackslashAtEnd(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"printf", `tail\`})
	if got := out.String(); got != `tail\` {
		t.Errorf("got %q", got)
	}
}

// TestPrintfBEmpty covers the %b branch with an empty arg.
func TestPrintfBEmpty(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"printf", "%b\n"})
	if got := out.String(); got != "\n" {
		t.Errorf("got %q", got)
	}
}

// TestPrintfBLastBackslash covers the %b branch where the input ends
// with a bare backslash (no second char to decode).
func TestPrintfBLastBackslash(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"printf", "%b\n", `end\`})
	if got := out.String(); got != "end\\\n" {
		t.Errorf("got %q", got)
	}
}

// TestPrintfCharEmpty covers the %c branch with an empty arg.
func TestPrintfCharEmpty(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"printf", "[%c]"})
	if got := out.String(); got != "[]" {
		t.Errorf("got %q", got)
	}
}
