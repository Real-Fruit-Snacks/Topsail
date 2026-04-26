package ls

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func setupDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range []string{"alpha", "beta", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestLsBasic(t *testing.T) {
	dir := setupDir(t)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"ls", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") {
		t.Errorf("missing entries in %q", got)
	}
	if strings.Contains(got, ".hidden") {
		t.Errorf("hidden file shown without -a in %q", got)
	}
}

func TestLsAll(t *testing.T) {
	dir := setupDir(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"ls", "-a", dir})
	if !strings.Contains(out.String(), ".hidden") {
		t.Errorf("missing .hidden with -a: %q", out.String())
	}
}

func TestLsLong(t *testing.T) {
	dir := setupDir(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"ls", "-l", dir})
	got := out.String()
	if !strings.Contains(got, "alpha") {
		t.Errorf("missing entry in %q", got)
	}
	// Long format includes a date like "2026-04-26".
	if !strings.Contains(got, "20") {
		t.Errorf("expected date in long format: %q", got)
	}
}

func TestLsHuman(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "big"), make([]byte, 5000), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"ls", "-lh", dir})
	if !strings.Contains(out.String(), "K") {
		t.Errorf("expected K suffix in human size; got %q", out.String())
	}
}

func TestLsNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ls", "/no/such/path"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestLsInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ls", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestLsFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	if err := os.WriteFile(p, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"ls", p})
	if !strings.Contains(out.String(), "x") {
		t.Errorf("got %q", out.String())
	}
}
