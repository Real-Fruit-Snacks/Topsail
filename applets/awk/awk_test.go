package awk

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// awkOut returns awk's output with CRLF normalized to LF. goawk emits
// platform-native line endings on Windows; our tests want a portable
// assertion target.
func awkOut(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func TestAwkPrint(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"awk", "{ print }"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := awkOut(out.String()), "alpha\nbeta\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkFields(t *testing.T) {
	testutil.SetStdin(t, "a b c\n1 2 3\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"awk", "{ print $2 }"})
	if got, want := awkOut(out.String()), "b\n2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkFieldSeparator(t *testing.T) {
	testutil.SetStdin(t, "a:b:c\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"awk", "-F", ":", "{ print $2 }"})
	if got, want := awkOut(out.String()), "b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkVar(t *testing.T) {
	testutil.SetStdin(t, "")
	out := testutil.CaptureStdout(t)
	Main([]string{"awk", "-v", "name=World", "BEGIN { print \"Hello,\", name }"})
	if got, want := awkOut(out.String()), "Hello, World\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkPattern(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\ngamma\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"awk", "/^[ag]/ { print }"})
	if got, want := awkOut(out.String()), "alpha\ngamma\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkArithmetic(t *testing.T) {
	testutil.SetStdin(t, "1 2\n3 4\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"awk", "{ print $1 + $2 }"})
	if got, want := awkOut(out.String()), "3\n7\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestAwkParseError(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"awk", "{ print"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestAwkMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"awk"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
