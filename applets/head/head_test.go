package head

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

func TestHeadDefault(t *testing.T) {
	dir := t.TempDir()
	var sb strings.Builder
	for i := 1; i <= 15; i++ {
		sb.WriteString("L\n")
	}
	p := writeFile(t, dir, "f", sb.String())
	out := testutil.CaptureStdout(t)
	Main([]string{"head", p})
	got := out.String()
	if strings.Count(got, "\n") != 10 {
		t.Errorf("expected 10 lines, got %d in %q", strings.Count(got, "\n"), got)
	}
}

func TestHeadN(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-n", "2", p})
	if got, want := out.String(), "1\n2\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestHeadLongN(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "--lines=1", p})
	if got, want := out.String(), "1\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestHeadShortHandN(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-3", p})
	if got, want := out.String(), "1\n2\n3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestHeadBytes(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "abcdefghij")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-c", "4", p})
	if got, want := out.String(), "abcd"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestHeadStdin(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\nd\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-n", "2"})
	if got, want := out.String(), "a\nb\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestHeadMultipleHeaders(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n")
	b := writeFile(t, dir, "b", "2\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", a, b})
	got := out.String()
	if !strings.Contains(got, "==> ") {
		t.Errorf("expected headers; got %q", got)
	}
}

func TestHeadQuiet(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n")
	b := writeFile(t, dir, "b", "2\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-q", a, b})
	if strings.Contains(out.String(), "==>") {
		t.Errorf("expected no headers with -q; got %q", out.String())
	}
}

func TestHeadVerbose(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"head", "-v", a})
	if !strings.Contains(out.String(), "==>") {
		t.Errorf("expected header with -v; got %q", out.String())
	}
}

func TestHeadInvalidN(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"head", "-n", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestHeadMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"head", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
