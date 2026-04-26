package id

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestIdDefault(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"id"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "uid=") {
		t.Errorf("got %q", out.String())
	}
}

func TestIdUser(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"id", "-u"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected uid")
	}
}

func TestIdUserName(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"id", "-un"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Error("expected username")
	}
}

func TestIdInvalidUser(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"id", "definitely-no-such-user-asdfgh"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestIdInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"id", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
