package chown

import (
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestChownInvalidSpec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chown is a stub on Windows")
	}
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chown", "abc", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestChownMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chown"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestChownInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chown", "-Z", "0:0", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestChownWindowsStub(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only meaningful on Windows")
	}
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"chown", "0:0", "."}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "not supported") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
