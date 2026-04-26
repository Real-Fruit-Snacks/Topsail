package paste

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

func TestPasteParallel(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n2\n3\n")
	b := writeFile(t, dir, "b", "x\ny\nz\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"paste", a, b})
	if got, want := out.String(), "1\tx\n2\ty\n3\tz\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestPasteDelim(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n2\n")
	b := writeFile(t, dir, "b", "x\ny\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"paste", "-d", ",", a, b})
	if got, want := out.String(), "1,x\n2,y\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestPasteSerial(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1\n2\n3\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"paste", "-s", a})
	if got, want := out.String(), "1\t2\t3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestPasteInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"paste", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
