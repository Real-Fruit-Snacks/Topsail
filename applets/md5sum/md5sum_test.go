package md5sum

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestMd5SumStdin(t *testing.T) {
	testutil.SetStdin(t, "hello")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"md5sum"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "5d41402abc4b2a76b9719d911017c592") {
		t.Errorf("got %q", out.String())
	}
}

func TestMd5SumInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"md5sum", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
