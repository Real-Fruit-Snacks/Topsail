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

func TestSortKeySingleField(t *testing.T) {
	testutil.SetStdin(t, "bob 30\nalice 25\ncarol 40\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-k", "2", "-n"})
	if got, want := out.String(), "alice 25\nbob 30\ncarol 40\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortKeyAttachedField(t *testing.T) {
	testutil.SetStdin(t, "b 1\na 3\nc 2\n")
	out := testutil.CaptureStdout(t)
	// -k2 (no space) should equal -k 2.
	Main([]string{"sort", "-k2n"})
	if got, want := out.String(), "b 1\nc 2\na 3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortKeyRange(t *testing.T) {
	testutil.SetStdin(t, "x b 1\nx a 2\nx a 1\n")
	out := testutil.CaptureStdout(t)
	// Sort by fields 2..3 ascending; key1 ("x") is identical so it's a tiebreaker.
	Main([]string{"sort", "-k", "2,3"})
	if got, want := out.String(), "x a 1\nx a 2\nx b 1\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortKeyPerKeyFlags(t *testing.T) {
	testutil.SetStdin(t, "a 10\nb 2\nc 9\n")
	out := testutil.CaptureStdout(t)
	// "2nr" = numeric reverse on key 2.
	Main([]string{"sort", "-k", "2nr"})
	if got, want := out.String(), "a 10\nc 9\nb 2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortMultipleKeys(t *testing.T) {
	testutil.SetStdin(t, "alice b\nbob a\nalice a\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-k", "1,1", "-k", "2,2"})
	if got, want := out.String(), "alice a\nalice b\nbob a\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortFieldSeparator(t *testing.T) {
	testutil.SetStdin(t, "user:101\nadmin:1\nguest:50\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sort", "-t", ":", "-k", "2", "-n"})
	if got, want := out.String(), "admin:1\nguest:50\nuser:101\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortFieldSeparatorAttached(t *testing.T) {
	testutil.SetStdin(t, "a|2\nb|1\n")
	out := testutil.CaptureStdout(t)
	// -t| with attached separator.
	Main([]string{"sort", "-t|", "-k", "2", "-n"})
	if got, want := out.String(), "b|1\na|2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSortRejectsCharacterOffset(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sort", "-k", "1.3"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "not supported") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSortRejectsBadSep(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sort", "-t", "ab"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "single character") {
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
