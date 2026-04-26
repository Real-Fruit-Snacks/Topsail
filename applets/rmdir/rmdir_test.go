package rmdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestRmdirEmpty(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "empty")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", d}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(d); !os.IsNotExist(err) {
		t.Errorf("dir still exists: %v", err)
	}
}

func TestRmdirNonEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", root}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

func TestRmdirParents(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", "-p", deep}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Errorf("parent 'a' still exists: %v", err)
	}
}

func TestRmdirIgnoreNonEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", "--ignore-fail-on-non-empty", root}); rc != 0 {
		t.Errorf("rc = %d; want 0 with --ignore-fail-on-non-empty", rc)
	}
}

func TestRmdirVerbose(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "v")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", "-v", d}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "removing") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestRmdirMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRmdirInvalidOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", "-x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRmdirNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rmdir", "/this/does/not/exist"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}
