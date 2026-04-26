package du

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestDuFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, make([]byte, 2048), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"du", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), p) {
		t.Errorf("got %q", out.String())
	}
}

func TestDuHuman(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big")
	if err := os.WriteFile(p, make([]byte, 5000), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"du", "-h", p})
	if !strings.Contains(out.String(), "K") {
		t.Errorf("expected K suffix; got %q", out.String())
	}
}

func TestDuSummary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"du", "-s", dir})
	got := out.String()
	if strings.Count(got, "\n") != 1 {
		t.Errorf("expected 1 line; got %q", got)
	}
}

func TestDuNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"du", "/no/such/path"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestDuInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"du", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
