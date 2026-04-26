package uniq

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestUniqBasic(t *testing.T) {
	testutil.SetStdin(t, "a\na\nb\nc\nc\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"uniq"})
	if got, want := out.String(), "a\nb\nc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUniqCount(t *testing.T) {
	testutil.SetStdin(t, "a\na\nb\nc\nc\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"uniq", "-c"})
	got := out.String()
	if !strings.Contains(got, "      2 a") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "      3 c") {
		t.Errorf("got %q", got)
	}
}

func TestUniqRepeated(t *testing.T) {
	testutil.SetStdin(t, "a\na\nb\nc\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"uniq", "-d"})
	if got, want := out.String(), "a\nc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUniqUnique(t *testing.T) {
	testutil.SetStdin(t, "a\na\nb\nc\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"uniq", "-u"})
	if got, want := out.String(), "b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUniqIgnoreCase(t *testing.T) {
	testutil.SetStdin(t, "Hello\nhello\nHELLO\nbye\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"uniq", "-i"})
	if got, want := out.String(), "Hello\nbye\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestUniqEmpty(t *testing.T) {
	testutil.SetStdin(t, "")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"uniq"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestUniqInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"uniq", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUniqExtraOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"uniq", "a", "b", "c"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "extra operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
