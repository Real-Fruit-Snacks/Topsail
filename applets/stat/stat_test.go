package stat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestStatDefault(t *testing.T) {
	p := setup(t)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"stat", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	for _, want := range []string{"File:", "Size:", "Type:", "regular file"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestStatTerse(t *testing.T) {
	p := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"stat", "-t", p})
	if !strings.Contains(out.String(), p) {
		t.Errorf("missing path in %q", out.String())
	}
}

func TestStatFormat(t *testing.T) {
	p := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"stat", "-c", "%s %F %n", p})
	if got, want := strings.TrimSpace(out.String()), "5 regular file "+p; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestStatNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"stat", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestStatMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"stat"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
