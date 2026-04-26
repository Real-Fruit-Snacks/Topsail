package dirname

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestDirnameBasic(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "/foo/bar/baz.txt"})
	if got, want := out.String(), "/foo/bar\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDirnameNoSlash(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "file.txt"})
	if got, want := out.String(), ".\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDirnameRoot(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "/"})
	if got, want := out.String(), "/\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDirnameMultiple(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "/a/b", "/c/d/e", "x"})
	if got, want := out.String(), "/a\n/c/d\n.\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDirnameZero(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "-z", "/a/b", "/c/d"})
	if got, want := out.String(), "/a\x00/c\x00"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDirnameMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"dirname"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestDirnameUnknown(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"dirname", "--bogus"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestDirnameDoubleDash(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"dirname", "--", "-weird/path"})
	if got, want := out.String(), "-weird\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
