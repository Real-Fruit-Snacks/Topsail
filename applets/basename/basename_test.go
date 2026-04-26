package basename

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestBasenamePlain(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "/foo/bar/baz.txt"})
	if got, want := out.String(), "baz.txt\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameWithSuffix(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "/foo/bar/baz.txt", ".txt"})
	if got, want := out.String(), "baz\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameMultiple(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "-a", "/a/x", "/b/y", "/c/z"})
	if got, want := out.String(), "x\ny\nz\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameSuffixFlag(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "-s", ".log", "/var/log/sys.log", "/var/log/auth.log"})
	if got, want := out.String(), "sys\nauth\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameLongSuffixFlag(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "--suffix=.log", "/var/log/sys.log"})
	if got, want := out.String(), "sys\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameZeroTerminator(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "-z", "-a", "x", "y"})
	if got, want := out.String(), "x\x00y\x00"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"basename"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestBasenameUnknownOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"basename", "--bogus", "x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestBasenameSuffixRequiresArg(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"basename", "-s"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "requires an argument") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestBasenameDoubleDash(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"basename", "--", "-foo.txt", ".txt"})
	if got, want := out.String(), "-foo\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBasenameExtraOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"basename", "/a", "/b", "/c"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "extra operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
