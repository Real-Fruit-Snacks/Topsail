package host

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// host's behavior depends on a working resolver; on offline runners
// some tests may be flaky. We use localhost for portable cases and
// only test parsing/error paths for the rest.

func TestHostLocalhost(t *testing.T) {
	out := testutil.CaptureStdout(t)
	rc := Main([]string{"host", "localhost"})
	// Even if the system can't resolve localhost (rare), the test
	// runner should at least produce some output or a clear error.
	if rc != 0 && rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if rc == 0 && !strings.Contains(strings.ToLower(got), "127.0.0.1") &&
		!strings.Contains(strings.ToLower(got), "::1") {
		t.Errorf("got %q (expected a localhost address)", got)
	}
}

func TestHostUnsupportedType(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"host", "-t", "BOGUS", "example.com"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unsupported record type") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestHostMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"host"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestHostInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"host", "-Z", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
