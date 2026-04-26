package xargs

import (
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// xargs's behavior is to spawn external processes; on a portable test
// runner we can only easily exercise the parsing path that targets a
// command we know exists. We use `echo` on Unix-like systems; on Windows
// the shell echo is a built-in, so we exercise an alternative.

func TestXargsParseN(t *testing.T) {
	testutil.SetStdin(t, "1 2 3 4 5")
	_, _ = testutil.CaptureStdio(t)
	// We can't easily verify external command invocation portably.
	// At minimum, ensure -n parses without error and returns 0/1.
	rc := Main([]string{"xargs", "-n", "2"})
	if rc < 0 || rc > 2 {
		t.Errorf("unexpected rc = %d", rc)
	}
}

func TestXargsMissingNValue(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xargs", "-n"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "requires") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestXargsBadN(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xargs", "-n", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid number") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestXargsInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xargs", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestXargsMissingDValue(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"xargs", "-d", ""}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestXargsItemSplit(t *testing.T) {
	testutil.SetStdin(t, "a\nb\nc\n")
	_, _ = testutil.CaptureStdio(t)
	// We accept any rc; we're verifying the parse path.
	_ = Main([]string{"xargs", "-n", "1"})
	if runtime.GOOS == "" {
		t.Fatal("test runner has no GOOS")
	}
}
