package tac

import (
	"os"
	"path/filepath"
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

func TestTacBasic(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "a\nb\nc\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tac", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "c\nb\na\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTacStdin(t *testing.T) {
	testutil.SetStdin(t, "1\n2\n3\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tac"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "3\n2\n1\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTacMultiple(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n2\n")
	b := writeFile(t, dir, "b", "3\n4\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"tac", a, b})
	if got, want := out.String(), "2\n1\n4\n3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestTacEmpty(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tac", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestTacMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tac", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
