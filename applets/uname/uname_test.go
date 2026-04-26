package uname

import (
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestUnameDefault(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"uname"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected non-empty kernel name")
	}
}

func TestUnameAll(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"uname", "-a"})
	got := out.String()
	if !strings.Contains(got, runtime.GOARCH) {
		t.Errorf("expected arch in -a output: %q", got)
	}
}

func TestUnameMachine(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"uname", "-m"})
	if !strings.Contains(out.String(), runtime.GOARCH) {
		t.Errorf("got %q", out.String())
	}
}

func TestUnameInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"uname", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
