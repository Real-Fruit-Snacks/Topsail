package chmod

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestChmodBasic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "0700", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("mode = %o; want 0700", got)
	}
}

func TestChmodInvalidMode(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "bogus", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid mode") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestChmodMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestChmodNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "0644", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestChmodRecursive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "sub", "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "-R", "0700", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("nested mode = %o; want 0700", got)
	}
}

func TestChmodSymbolicAdd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "u+x", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != 0o744 {
		t.Errorf("mode after u+x: got %o; want 0744", got)
	}
}

func TestChmodSymbolicRemove(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "go-w", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != 0o644 {
		t.Errorf("mode after go-w: got %o; want 0644", got)
	}
}

func TestChmodSymbolicEqualMulti(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes not meaningful on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"chmod", "u=rwx,go=rx", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != 0o755 {
		t.Errorf("mode after u=rwx,go=rx: got %o; want 0755", got)
	}
}
