package tail

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

func TestTailDefault(t *testing.T) {
	dir := t.TempDir()
	var sb strings.Builder
	for i := 1; i <= 15; i++ {
		sb.WriteString("L\n")
	}
	p := writeFile(t, dir, "f", sb.String())
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", p})
	if strings.Count(out.String(), "\n") != 10 {
		t.Errorf("expected 10 lines, got %q", out.String())
	}
}

func TestTailN(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-n", "2", p})
	if got, want := out.String(), "4\n5\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTailFromStart(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-n", "+3", p})
	if got, want := out.String(), "3\n4\n5\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTailBytes(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "abcdefghij")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-c", "3", p})
	if got, want := out.String(), "hij"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTailStdin(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\nd\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-n", "2"})
	if got, want := out.String(), "c\nd\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTailHeaders(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "x\n")
	b := writeFile(t, dir, "b", "y\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", a, b})
	if !strings.Contains(out.String(), "==>") {
		t.Errorf("expected headers; got %q", out.String())
	}
}

func TestTailFollowStdinWarns(t *testing.T) {
	// tail -f - has no real file to follow; we expect a diagnostic and a
	// clean exit (the initial tail of stdin still happens).
	testutil.SetStdin(t, "x\n")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tail", "-f", "-"}); rc != 0 {
		t.Errorf("rc = %d; want 0", rc)
	}
	if !strings.Contains(errBuf.String(), "ineffective on standard input") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTailInvalidN(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tail", "-n", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestTailShorthand(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-3", p})
	if got, want := out.String(), "3\n4\n5\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTailFewerThanN(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "x\ny\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tail", "-n", "10", p})
	if got, want := out.String(), "x\ny\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
