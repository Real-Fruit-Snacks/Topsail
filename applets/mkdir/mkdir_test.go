package mkdir

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestMkdirBasic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "newdir")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("target not created: %v", err)
	}
}

func TestMkdirAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", dir}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

func TestMkdirParents(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "a", "b", "c")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-p", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if info, err := os.Stat(target); err != nil || !info.IsDir() {
		t.Errorf("target not created: err=%v", err)
	}
}

func TestMkdirParentsExisting(t *testing.T) {
	dir := t.TempDir()
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-p", dir}); rc != 0 {
		t.Errorf("rc = %d; -p should be silent for existing dirs", rc)
	}
}

func TestMkdirNoParents(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "missing", "child")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", target}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

func TestMkdirVerbose(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "vdir")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-v", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "vdir") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestMkdirMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	root := t.TempDir()
	target := filepath.Join(root, "modedir")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-m", "0700", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("mode = %o; want 0700", got)
	}
}

func TestMkdirInvalidMode(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-m", "bogus", "x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid mode") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMkdirMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMkdirCombinedFlags(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "a", "b")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-pv", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "a") {
		t.Errorf("stdout = %q (expected verbose output)", out.String())
	}
}
