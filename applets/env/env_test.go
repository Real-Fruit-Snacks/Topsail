package env

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestEnvPrint(t *testing.T) {
	t.Setenv("TOPSAIL_TEST_VAR", "hello")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"env"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "TOPSAIL_TEST_VAR=hello") {
		t.Errorf("missing var in output")
	}
}

func TestEnvIgnoreEnvironment(t *testing.T) {
	t.Setenv("TOPSAIL_KEEP_ME", "x")
	out := testutil.CaptureStdout(t)
	Main([]string{"env", "-i"})
	if strings.Contains(out.String(), "TOPSAIL_KEEP_ME") {
		t.Errorf("env -i should hide existing vars; got %q", out.String())
	}
}

func TestEnvSetVar(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"env", "-i", "FOO=bar"})
	if !strings.Contains(out.String(), "FOO=bar") {
		t.Errorf("expected FOO=bar; got %q", out.String())
	}
}

func TestEnvUnset(t *testing.T) {
	t.Setenv("TOPSAIL_REMOVE_ME", "x")
	out := testutil.CaptureStdout(t)
	Main([]string{"env", "-u", "TOPSAIL_REMOVE_ME"})
	if strings.Contains(out.String(), "TOPSAIL_REMOVE_ME") {
		t.Errorf("expected TOPSAIL_REMOVE_ME removed; got %q", out.String())
	}
}

func TestEnvUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"env", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
