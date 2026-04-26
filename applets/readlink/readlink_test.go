package readlink

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestReadlinkSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may require privilege on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "tgt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "lnk")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink: %v", err)
	}
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"readlink", link}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := strings.TrimSpace(out.String()); got != target {
		t.Errorf("got %q; want %q", got, target)
	}
}

func TestReadlinkCanonicalize(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"readlink", "-f", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected canonical path")
	}
}

func TestReadlinkMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"readlink"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestReadlinkInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"readlink", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
