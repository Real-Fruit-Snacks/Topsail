package ln

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestLnSymbolic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tgt.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "lnk")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"ln", "-s", target, link}); rc != 0 {
		// On Windows without admin, symlink creation can fail; allow that.
		if runtime.GOOS == "windows" {
			t.Skip("symlinks may require privilege on Windows")
		}
		t.Errorf("rc = %d", rc)
	}
}

func TestLnHardLink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tgt.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "lnk")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"ln", target, link}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(link); err != nil {
		t.Errorf("hardlink not present: %v", err)
	}
}

func TestLnMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ln", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestLnInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ln", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
