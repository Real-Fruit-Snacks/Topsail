package pwd

import (
	"os"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestPwdBasic(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"pwd"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimRight(out.String(), "\n")
	want, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("pwd = %q; want %q", got, want)
	}
}

func TestPwdLogical(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"pwd", "-L"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("pwd -L produced empty output")
	}
}

func TestPwdPhysical(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"pwd", "-P"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("pwd -P produced empty output")
	}
}

func TestPwdInvalidOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"pwd", "--bogus"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestPwdDoubleDashAccepted(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"pwd", "--"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("pwd -- produced empty output")
	}
}
