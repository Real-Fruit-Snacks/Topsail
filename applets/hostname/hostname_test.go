package hostname

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestHostname(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"hostname"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected non-empty hostname")
	}
}

func TestHostnameShort(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"hostname", "-s"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.Contains(out.String(), ".") {
		t.Errorf("expected short hostname; got %q", out.String())
	}
}

func TestHostnameInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"hostname", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
