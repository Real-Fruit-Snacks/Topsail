package tr

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestTrSimple(t *testing.T) {
	testutil.SetStdin(t, "abc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "abc", "xyz"})
	if got, want := out.String(), "xyz\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrRange(t *testing.T) {
	testutil.SetStdin(t, "Hello World\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "a-z", "A-Z"})
	if got, want := out.String(), "HELLO WORLD\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrDelete(t *testing.T) {
	testutil.SetStdin(t, "Hello World\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "-d", "lo"})
	if got, want := out.String(), "He Wrd\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrComplement(t *testing.T) {
	testutil.SetStdin(t, "Hello, World 123!\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "-cd", "a-zA-Z"})
	if got, want := out.String(), "HelloWorld"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrSqueeze(t *testing.T) {
	testutil.SetStdin(t, "aabbcc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "-s", "a-c"})
	if got, want := out.String(), "abc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrCharClasses(t *testing.T) {
	testutil.SetStdin(t, "abc123\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "[:lower:]", "[:upper:]"})
	if got, want := out.String(), "ABC123\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrEscapes(t *testing.T) {
	testutil.SetStdin(t, "a\tb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "\\t", " "})
	if got, want := out.String(), "a b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrSet1LongerThanSet2(t *testing.T) {
	// Last char of set2 is repeated for extras.
	testutil.SetStdin(t, "abcde\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tr", "abcde", "xy"})
	if got, want := out.String(), "xyyyy\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTrMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tr"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTrInvalidOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tr", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTrInvalidRange(t *testing.T) {
	testutil.SetStdin(t, "x")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tr", "z-a", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
