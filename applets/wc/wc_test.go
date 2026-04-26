package wc

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

func TestWcDefault(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "hello world\nfoo bar baz\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", p})
	got := out.String()
	if !strings.Contains(got, "      2") {
		t.Errorf("expected 2 lines in %q", got)
	}
	if !strings.Contains(got, "      5") {
		t.Errorf("expected 5 words in %q", got)
	}
	if !strings.Contains(got, "     24") {
		t.Errorf("expected 24 bytes in %q", got)
	}
}

func TestWcLines(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-l", p})
	if !strings.Contains(out.String(), "      3") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcWords(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "alpha beta gamma\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-w", p})
	if !strings.Contains(out.String(), "      3") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcBytes(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "abcde")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-c", p})
	if !strings.Contains(out.String(), "      5") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcChars(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "héllo") // h=1 é=2 l=1 l=1 o=1 = 6 bytes, 5 chars
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-m", p})
	if !strings.Contains(out.String(), "      5") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcLongest(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "ab\nabcdef\nabc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-L", p})
	if !strings.Contains(out.String(), "      6") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcStdin(t *testing.T) {
	testutil.SetStdin(t, "hello\nworld\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-l"})
	if !strings.Contains(out.String(), "      2") {
		t.Errorf("got %q", out.String())
	}
}

func TestWcMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n2\n")
	b := writeFile(t, dir, "b", "1\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"wc", "-l", a, b})
	got := out.String()
	if !strings.Contains(got, "total") {
		t.Errorf("expected total line in %q", got)
	}
}

func TestWcMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"wc", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestWcInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"wc", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
