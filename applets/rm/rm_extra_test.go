package rm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestRmCombinedFlags exercises the -rfvi cluster + verbose + dir + force.
func TestRmCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.MkdirAll(filepath.Join(d, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(d, "f"), "")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-rfvi", d}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "removed") {
		t.Errorf("expected 'removed' in stdout: %q", out.String())
	}
}

// TestRmInvalidCombinedFlag exercises the unknown-char in cluster path.
func TestRmInvalidCombinedFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-rZ", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestRmDoubleDash verifies "--" stops flag parsing so a "-name" file can be removed.
func TestRmDoubleDash(t *testing.T) {
	dir := t.TempDir()
	odd := filepath.Join(dir, "-weird")
	writeFile(t, odd, "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "--", odd}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(odd); !os.IsNotExist(err) {
		t.Errorf("file still exists: %v", err)
	}
}

// TestRmEmptyDirNonEmpty fires the os.Remove failure path with -d on a non-empty dir.
func TestRmEmptyDirNonEmpty(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(d, "blocker"), "")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-d", d}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

// TestRmRegularFileFailure exercises the os.Remove failure path on a missing file.
func TestRmRegularFileFailure(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "ghost")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", missing}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

// TestRmLongFlags covers the long-option spellings.
func TestRmLongFlags(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	writeFile(t, a, "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "--verbose", "--force", "--recursive", a}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestRmInteractiveNoOp confirms -i and --interactive are accepted.
func TestRmInteractiveNoOp(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	writeFile(t, a, "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "--interactive", a}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}
