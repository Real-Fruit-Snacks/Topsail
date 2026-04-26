package unexpand

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestUnexpandLeading(t *testing.T) {
	// Default: only leading blanks compressed; embedded run preserved.
	testutil.SetStdin(t, "        word    here\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"unexpand"})
	got := out.String()
	if !strings.HasPrefix(got, "\tword") {
		t.Errorf("expected leading 8 spaces -> tab; got %q", got)
	}
	// The embedded 4 spaces should NOT be compressed (default mode).
	if !strings.Contains(got, "word    here") {
		t.Errorf("embedded spaces should be preserved; got %q", got)
	}
}

func TestUnexpandAll(t *testing.T) {
	testutil.SetStdin(t, "a       b       c\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"unexpand", "-a"})
	// "a" col 0; 7 spaces to col 8 (tab); "b" col 9; 7 spaces to col 16 (tab); "c".
	if got, want := out.String(), "a\tb\tc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUnexpandCustomTabs(t *testing.T) {
	testutil.SetStdin(t, "    word\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"unexpand", "-t", "4"})
	if got, want := out.String(), "\tword\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUnexpandFirstOnly(t *testing.T) {
	testutil.SetStdin(t, "        word        here\n")
	out := testutil.CaptureStdout(t)
	// --first-only is the default, but we exercise the explicit flag too.
	Main([]string{"unexpand", "--first-only"})
	got := out.String()
	if !strings.HasPrefix(got, "\tword") {
		t.Errorf("got %q; want leading tab", got)
	}
}

func TestUnexpandInvalidTabs(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"unexpand", "-t", "0"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUnexpandUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"unexpand", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
