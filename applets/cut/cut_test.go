package cut

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestCutFields(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\n1\t2\t3\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-f", "2"})
	if got, want := out.String(), "b\n2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutFieldsRange(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\td\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-f", "2-3"})
	if got, want := out.String(), "b\tc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutFieldsOpenEnd(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\td\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-f", "2-"})
	if got, want := out.String(), "b\tc\td\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutFieldsList(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\td\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-f", "1,3"})
	if got, want := out.String(), "a\tc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutDelim(t *testing.T) {
	testutil.SetStdin(t, "a:b:c\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-d", ":", "-f", "2"})
	if got, want := out.String(), "b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutOnlyDelim(t *testing.T) {
	testutil.SetStdin(t, "no_delim_here\nhas\ttab\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-s", "-f", "1"})
	if got, want := out.String(), "has\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutBytes(t *testing.T) {
	testutil.SetStdin(t, "abcdef\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-b", "1-3"})
	if got, want := out.String(), "abc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutChars(t *testing.T) {
	testutil.SetStdin(t, "héllo\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-c", "1-3"})
	if got, want := out.String(), "hél\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutOutputDelim(t *testing.T) {
	testutil.SetStdin(t, "a\tb\tc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cut", "-f", "1,3", "--output-delimiter=,"})
	if got, want := out.String(), "a,c\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCutInvalidList(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cut", "-f", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCutNoMode(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cut"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "must specify") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
