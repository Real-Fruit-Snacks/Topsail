package expand

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestExpandDefaultEightSpaces(t *testing.T) {
	testutil.SetStdin(t, "a\tb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"expand"})
	// "a" is column 0, tab advances to column 8 (7 spaces), then "b".
	if got, want := out.String(), "a       b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestExpandCustomTabWidth(t *testing.T) {
	testutil.SetStdin(t, "a\tb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"expand", "-t", "4"})
	if got, want := out.String(), "a   b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestExpandInitialOnly(t *testing.T) {
	testutil.SetStdin(t, "\tword\there\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"expand", "-i", "-t", "4"})
	// Leading tab expands to 4 spaces; embedded tab is preserved as a literal tab.
	got := out.String()
	if !strings.HasPrefix(got, "    word") {
		t.Errorf("got %q; want prefix '    word'", got)
	}
	if !strings.Contains(got, "word\there") {
		t.Errorf("embedded tab should be preserved: %q", got)
	}
}

func TestExpandExplicitStops(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"expand", "-t", "3,8"})
	// 'a' at col 0; tab -> col 3 (2 spaces). 'b' at col 3; tab -> col 8 (4 spaces). 'c' at col 8.
	if got, want := out.String(), "a  b    c\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestExpandLongFlagTabs(t *testing.T) {
	testutil.SetStdin(t, "a\tb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"expand", "--tabs=2"})
	if got, want := out.String(), "a b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestExpandInvalidTabs(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"expand", "-t", "0"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestExpandUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"expand", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
