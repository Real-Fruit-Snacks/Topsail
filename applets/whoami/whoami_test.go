package whoami

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestWhoami(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"whoami"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected non-empty username")
	}
}

func TestWhoamiExtraOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"whoami", "extra"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "extra operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
