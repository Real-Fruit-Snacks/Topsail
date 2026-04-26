package join

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

func TestJoinDefault(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1 alpha\n2 beta\n3 gamma\n")
	b := writeFile(t, dir, "b", "1 one\n2 two\n3 three\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"join", a, b}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "1 alpha one") {
		t.Errorf("got %q", got)
	}
}

func TestJoinTooFewFiles(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"join", "a"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "exactly two") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestJoinInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"join", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
