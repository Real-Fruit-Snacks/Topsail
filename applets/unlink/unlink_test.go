package unlink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestUnlinkFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"unlink", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("file still exists: %v", err)
	}
}

func TestUnlinkMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"unlink", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "cannot unlink") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUnlinkExactlyOneArg(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"unlink"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "expected exactly one operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
	_, errBuf = testutil.CaptureStdio(t)
	if rc := Main([]string{"unlink", "a", "b"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "expected exactly one operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUnlinkRefusesDirectory(t *testing.T) {
	dir := t.TempDir()
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"unlink", dir}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected error output for directory unlink")
	}
}
