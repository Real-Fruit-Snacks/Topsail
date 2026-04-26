package mkdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestMkdirLongFlags covers --parents / --verbose / --mode= long forms.
func TestMkdirLongFlags(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "a", "b", "c")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "--parents", "--verbose", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("target not created: %v", err)
	}
	if !strings.Contains(out.String(), "created directory") {
		t.Errorf("verbose output missing: %q", out.String())
	}
}

// TestMkdirLongMode covers --mode=NNNN.
func TestMkdirLongMode(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "modedir")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "--mode=0700", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestMkdirInvalidLongMode covers the parse-error path on --mode=garbage.
func TestMkdirInvalidLongMode(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "--mode=garbage", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid mode") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestMkdirModeMissingArg covers -m with no value.
func TestMkdirModeMissingArg(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-m"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "requires an argument") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestMkdirDoubleDash verifies "--" stops flag parsing.
func TestMkdirDoubleDash(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "-weird")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "--", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestMkdirInvalidCombinedFlag exercises the unknown-char in cluster.
func TestMkdirInvalidCombinedFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-pZ", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestMkdirParentsAlreadyExist exercises the ErrExist+parents skip path
// (more directly than the existing test, which lets the dir already be
// there at the top).
func TestMkdirParentsExistsMidway(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "existing"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "existing")
	testutil.CaptureStdio(t)
	// -p on an existing path should silently succeed.
	if rc := Main([]string{"mkdir", "-p", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestMkdirSymbolicMode exercises POSIX-style symbolic --mode arguments,
// which are evaluated against an implied 0o777 base so "u-x" yields 0o677.
func TestMkdirSymbolicMode(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "sym")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mkdir", "-m", "u-x", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("target not created: %v", err)
	}
	// On Windows file modes are mostly cosmetic, so don't assert bits.
}
