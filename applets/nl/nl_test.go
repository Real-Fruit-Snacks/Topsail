package nl

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestNlAll(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\ngamma\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"nl"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "1\talpha") {
		t.Errorf("got %q", got)
	}
}

func TestNlNonEmpty(t *testing.T) {
	testutil.SetStdin(t, "a\n\nb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"nl", "-b", "t"})
	got := out.String()
	if !strings.Contains(got, "1\ta") || !strings.Contains(got, "2\tb") {
		t.Errorf("got %q", got)
	}
}

func TestNlSeparator(t *testing.T) {
	testutil.SetStdin(t, "x\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"nl", "-s", "|", "-w", "1"})
	if got, want := out.String(), "1|x\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestNlInvalidMode(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"nl", "-b", "z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid line numbering") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestNlInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"nl", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
