package which

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestWhichFound(t *testing.T) {
	dir := t.TempDir()
	name := "topsailtest"
	if runtime.GOOS == "windows" {
		name = "topsailtest.exe"
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte{0x7f, 'E', 'L', 'F'}, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"which", "topsailtest"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "topsailtest") {
		t.Errorf("got %q", out.String())
	}
}

func TestWhichNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, _ = testutil.CaptureStdio(t)
	if rc := Main([]string{"which", "definitely-not-here-asdfgh"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
}

func TestWhichMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"which"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestWhichInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"which", "-Z", "ls"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
