package sha256sum

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

func TestSha256SumStdin(t *testing.T) {
	testutil.SetStdin(t, "hello")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sha256sum"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimSpace(out.String())
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  -"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSha256SumFile(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "hello")
	out := testutil.CaptureStdout(t)
	Main([]string{"sha256sum", p})
	got := out.String()
	if !strings.Contains(got, "2cf24dba") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, p) {
		t.Errorf("missing path in %q", got)
	}
}

func TestSha256SumCheck(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "hello")
	sumLine := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  " + p + "\n"
	listFile := writeFile(t, dir, "list", sumLine)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sha256sum", "-c", listFile}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("got %q", out.String())
	}
}

func TestSha256SumCheckFail(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "different content")
	sumLine := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  " + p + "\n"
	listFile := writeFile(t, dir, "list", sumLine)
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"sha256sum", "-c", listFile}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(out.String(), "FAILED") {
		t.Errorf("got %q", out.String())
	}
}

func TestSha256SumMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sha256sum", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestSha256SumInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sha256sum", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
