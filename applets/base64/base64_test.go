package base64

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestBase64Encode(t *testing.T) {
	testutil.SetStdin(t, "hello")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"base64"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := strings.TrimSpace(out.String()), "aGVsbG8="; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBase64Decode(t *testing.T) {
	testutil.SetStdin(t, "aGVsbG8=")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"base64", "-d"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "hello" {
		t.Errorf("got %q; want hello", got)
	}
}

func TestBase64Roundtrip(t *testing.T) {
	original := "The quick brown fox jumps over the lazy dog"
	testutil.SetStdin(t, original)
	encoded := testutil.CaptureStdout(t)
	Main([]string{"base64"})

	testutil.SetStdin(t, encoded.String())
	decoded := testutil.CaptureStdout(t)
	Main([]string{"base64", "-d"})
	if got := decoded.String(); got != original {
		t.Errorf("roundtrip got %q; want %q", got, original)
	}
}

func TestBase64Wrap(t *testing.T) {
	testutil.SetStdin(t, strings.Repeat("a", 80))
	out := testutil.CaptureStdout(t)
	Main([]string{"base64", "-w", "20"})
	got := out.String()
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if len(line) > 20 {
			t.Errorf("line longer than wrap: %q", line)
		}
	}
}

func TestBase64NoWrap(t *testing.T) {
	testutil.SetStdin(t, strings.Repeat("a", 200))
	out := testutil.CaptureStdout(t)
	Main([]string{"base64", "-w", "0"})
	got := out.String()
	if strings.Count(got, "\n") != 1 {
		t.Errorf("expected 1 trailing newline, got %d in %q", strings.Count(got, "\n"), got)
	}
}

func TestBase64InvalidInput(t *testing.T) {
	testutil.SetStdin(t, "not-valid-base64!@#")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"base64", "-d"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestBase64IgnoreGarbage(t *testing.T) {
	testutil.SetStdin(t, "aGVs!@bG8=")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"base64", "-d", "-i"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestBase64InvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"base64", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
