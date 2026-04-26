package xxd

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestXxdCanonical(t *testing.T) {
	testutil.SetStdin(t, "Hello, topsail!\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"xxd"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "00000000:") {
		t.Errorf("expected canonical offset prefix; got %q", got)
	}
	if !strings.Contains(got, "Hello, topsail!.") {
		t.Errorf("expected ASCII column; got %q", got)
	}
}

func TestXxdPlain(t *testing.T) {
	testutil.SetStdin(t, "abc")
	out := testutil.CaptureStdout(t)
	Main([]string{"xxd", "-p"})
	got := strings.TrimSpace(out.String())
	if got != "616263" {
		t.Errorf("got %q; want 616263", got)
	}
}

func TestXxdUppercase(t *testing.T) {
	testutil.SetStdin(t, "ABC")
	out := testutil.CaptureStdout(t)
	Main([]string{"xxd", "-p", "-u"})
	got := strings.TrimSpace(out.String())
	if got != "414243" {
		t.Errorf("got %q; want 414243", got)
	}
}

func TestXxdRevertPlain(t *testing.T) {
	testutil.SetStdin(t, "48656c6c6f0a")
	out := testutil.CaptureStdout(t)
	Main([]string{"xxd", "-r", "-p"})
	if got := out.String(); got != "Hello\n" {
		t.Errorf("got %q; want Hello\\n", got)
	}
}

func TestXxdRevertCanonical(t *testing.T) {
	// Round-trip a known string.
	in := "Hello, topsail!\n"
	testutil.SetStdin(t, in)
	dump := testutil.CaptureStdout(t)
	Main([]string{"xxd"})

	// Now revert the dump.
	testutil.SetStdin(t, dump.String())
	roundtrip := testutil.CaptureStdout(t)
	Main([]string{"xxd", "-r"})
	if got := roundtrip.String(); got != in {
		t.Errorf("round-trip got %q; want %q", got, in)
	}
}

func TestXxdCustomCols(t *testing.T) {
	testutil.SetStdin(t, "0123456789abcdef0123456789abcdef")
	out := testutil.CaptureStdout(t)
	Main([]string{"xxd", "-c", "8"})
	// 32 bytes / 8 cols = 4 lines.
	got := out.String()
	if strings.Count(got, "\n") != 4 {
		t.Errorf("expected 4 lines for -c 8; got %d:\n%s", strings.Count(got, "\n"), got)
	}
}

func TestXxdInvalidCols(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xxd", "-c", "0"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid -c") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestXxdMissingFile(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xxd", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
