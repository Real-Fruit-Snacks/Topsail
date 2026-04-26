package sort

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSortBasic(t *testing.T) {
	testutil.SetStdin(t, "banana\napple\ncherry\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort"})
	if got, want := out.String(), "apple\nbanana\ncherry\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortReverse(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-r"})
	if got, want := out.String(), "c\nb\na\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortNumeric(t *testing.T) {
	testutil.SetStdin(t, "10\n2\n1\n20\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-n"})
	if got, want := out.String(), "1\n2\n10\n20\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortUnique(t *testing.T) {
	testutil.SetStdin(t, "a\nb\na\nc\nb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-u"})
	if got, want := out.String(), "a\nb\nc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortIgnoreCase(t *testing.T) {
	testutil.SetStdin(t, "Banana\napple\nCherry\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-f"})
	if got, want := out.String(), "apple\nBanana\nCherry\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortIgnoreLeadingBlanks(t *testing.T) {
	testutil.SetStdin(t, "  banana\napple\n  cherry\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-b"})
	if got, want := out.String(), "apple\n  banana\n  cherry\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortRejectsK(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sort", "-k", "1"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "not supported") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSortInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sort", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
