package grep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGrepBasic(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\ngamma\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"grep", "et"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "beta\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepNoMatch(t *testing.T) {
	testutil.SetStdin(t, "alpha\nbeta\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"grep", "zzz"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if got := out.String(); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestGrepIgnoreCase(t *testing.T) {
	testutil.SetStdin(t, "Hello\nWorld\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-i", "hello"})
	if got, want := out.String(), "Hello\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepInvert(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-v", "b"})
	if got, want := out.String(), "a\nc\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepCount(t *testing.T) {
	testutil.SetStdin(t, "x\ny\nx\nx\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-c", "x"})
	if got, want := strings.TrimSpace(out.String()), "3"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepLineNum(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-n", "b"})
	if got, want := out.String(), "2:b\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepFile(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "alpha\nbeta\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "alpha", p})
	if got, want := out.String(), "alpha\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepMultipleFilesHeader(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "alpha\n")
	b := writeFile(t, dir, "b", "alpha\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "alpha", a, b})
	got := out.String()
	if !strings.Contains(got, ":alpha") {
		t.Errorf("expected filename headers; got %q", got)
	}
}

func TestGrepFixed(t *testing.T) {
	testutil.SetStdin(t, "1+2\n3+4\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-F", "1+2"})
	if got, want := out.String(), "1+2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestGrepFilesOnly(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "alpha\n")
	b := writeFile(t, dir, "b", "beta\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"grep", "-l", "alpha", a, b})
	if !strings.Contains(out.String(), filepath.Base(a)) {
		t.Errorf("expected %s in output; got %q", a, out.String())
	}
	if strings.Contains(out.String(), filepath.Base(b)+"\n") {
		t.Errorf("did not want %s in output", b)
	}
}

func TestGrepRecursive(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub"), "f", "needle\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"grep", "-r", "needle", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "needle") {
		t.Errorf("got %q", out.String())
	}
}

func TestGrepInvalidRegex(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"grep", "[unclosed"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestGrepMissingPattern(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"grep"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
