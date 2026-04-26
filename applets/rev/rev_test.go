package rev

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

func TestRevBasic(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "hello\nworld\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"rev", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "olleh\ndlrow\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevStdin(t *testing.T) {
	testutil.SetStdin(t, "abc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"rev"})
	if got, want := out.String(), "cba\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevUnicode(t *testing.T) {
	testutil.SetStdin(t, "héllo\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"rev"})
	if got, want := out.String(), "olléh\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevEmpty(t *testing.T) {
	testutil.SetStdin(t, "")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"rev"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestRevMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rev", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
