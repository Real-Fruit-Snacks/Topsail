package sum

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSumBSD(t *testing.T) {
	testutil.SetStdin(t, "hello\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sum"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.Fields(out.String())
	if len(got) != 2 {
		t.Errorf("expected 'CHKSUM BLOCKS' format; got %q", out.String())
	}
}

func TestSumSysV(t *testing.T) {
	testutil.SetStdin(t, "hello\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sum", "-s"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected output")
	}
}

func TestSumInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sum", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
