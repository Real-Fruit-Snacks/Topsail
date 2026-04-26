package comm

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

func TestComm(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "alpha\nbeta\ngamma\n")
	b := writeFile(t, dir, "b", "beta\ndelta\ngamma\n") // unsorted on purpose
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"comm", a, b}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "alpha") {
		t.Errorf("expected 'alpha' (only in a)")
	}
}

func TestCommTwoFiles(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"comm", "a"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "exactly two") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCommSuppress(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "alpha\nbeta\n")
	b := writeFile(t, dir, "b", "beta\ngamma\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"comm", "-12", a, b})
	if got := strings.TrimSpace(out.String()); got != "beta" {
		t.Errorf("got %q; want beta", got)
	}
}

func TestCommInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"comm", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
