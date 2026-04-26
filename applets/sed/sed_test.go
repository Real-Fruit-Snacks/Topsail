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
	// y/// transliteration is intentionally not supported in this build.
	testutil.SetStdin(t, "x\n")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sed", "y/abc/xyz/"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unsupported") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSedDelete(t *testing.T) {
	testutil.SetStdin(t, "keep\nremove\nkeep\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sed", "/remove/d"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "keep\nkeep\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedDeleteByLineNumber(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "2d"})
	if got, want := out.String(), "a\nc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedDeleteRange(t *testing.T) {
	testutil.SetStdin(t, "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "2,4d"})
	if got, want := out.String(), "1\n5\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedDeleteLastLine(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "$d"})
	if got, want := out.String(), "a\nb\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedPrintWithQuiet(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\ngamma\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "-n", "/eta/p"})
	if got, want := out.String(), "beta\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedQuit(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\nd\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "2q"})
	if got, want := out.String(), "a\nb\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedMultipleE(t *testing.T) {
	testutil.SetStdin(t, "abc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "-e", "s/a/A/", "-e", "s/c/C/"})
	if got, want := out.String(), "AbC\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedMultiCommandSemicolon(t *testing.T) {
	testutil.SetStdin(t, "abc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"sed", "s/a/A/; s/c/C/"})
	if got, want := out.String(), "AbC\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedNegateAddress(t *testing.T) {
	testutil.SetStdin(t, "# comment\nreal\n# more\nrest\n")
	out := testutil.CaptureStdout(t)
	// Print everything that isn't a comment line; -n suppresses default print.
	Main([]string{"sed", "-n", "/^#/!p"})
	if got, want := out.String(), "real\nrest\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSedRangeRegex(t *testing.T) {
	testutil.SetStdin(t, "head\n--start\nin1\nin2\n--end\ntail\n")
	out := testutil.CaptureStdout(t)
	// Delete lines from --start through --end inclusive.
	Main([]string{"sed", "/--start/,/--end/d"})
	if got, want := out.String(), "head\ntail\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
