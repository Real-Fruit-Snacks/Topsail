package sed

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSedSimple(t *testing.T) {
	testutil.SetStdin(t, "hello world\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sed", "s/world/everyone/"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "hello everyone\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedGlobal(t *testing.T) {
	testutil.SetStdin(t, "a a a\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s/a/b/g"})
	if got, want := out.String(), "b b b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedFirstOnly(t *testing.T) {
	testutil.SetStdin(t, "a a a\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s/a/b/"})
	if got, want := out.String(), "b a a\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedIgnoreCase(t *testing.T) {
	testutil.SetStdin(t, "Hello World\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s/world/X/i"})
	if got, want := out.String(), "Hello X\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedAmpersand(t *testing.T) {
	testutil.SetStdin(t, "abc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s/abc/[&]/"})
	if got, want := out.String(), "[abc]\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedCaptureGroup(t *testing.T) {
	testutil.SetStdin(t, "John Smith\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", `s/(\w+) (\w+)/\2 \1/`})
	if got, want := out.String(), "Smith John\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedAlternateDelim(t *testing.T) {
	testutil.SetStdin(t, "/usr/local/bin\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s|/usr/local|/opt|"})
	if got, want := out.String(), "/opt/bin\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedQuiet(t *testing.T) {
	testutil.SetStdin(t, "hello\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "-n", "s/h/H/"})
	if got := out.String(); got != "" {
		t.Errorf("got %q; want empty (quiet)", got)
	}
}

func TestSedMissingScript(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sed"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSedUnsupportedCommand(t *testing.T) {
	testutil.SetStdin(t, "x\n")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sed", "d"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "not") && !strings.Contains(errBuf.String(), "supported") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
