package timeout

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// helperShell returns a quick command that exits 0 immediately, suitable
// for testing the success path on every platform.
func helperShell(t *testing.T) []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c", "exit", "0"}
	}
	return []string{"true"}
}

// helperSleep returns a command that sleeps for d seconds.
func helperSleep(t *testing.T, d float64) []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		// PowerShell's Start-Sleep is reliable on stock Windows.
		return []string{"powershell", "-NoProfile", "-Command",
			"Start-Sleep", "-Seconds", durationToString(d)}
	}
	return []string{"sleep", durationToString(d)}
}

func durationToString(d float64) string {
	if d == float64(int(d)) {
		return string(rune('0' + int(d))) // single-digit int seconds
	}
	// Fall back: caller will pass simple values.
	return "1"
}

func TestTimeoutSuccess(t *testing.T) {
	args := append([]string{"timeout", "5"}, helperShell(t)...)
	testutil.CaptureStdio(t)
	if rc := Main(args); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestTimeoutTimesOut(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-sensitive; skipping in -short")
	}
	cmd := helperSleep(t, 5)
	args := append([]string{"timeout", "0.2"}, cmd...)
	testutil.CaptureStdio(t)
	start := time.Now()
	rc := Main(args)
	elapsed := time.Since(start)
	if rc != 124 {
		t.Errorf("rc = %d; want 124 (timed out)", rc)
	}
	if elapsed > 4*time.Second {
		t.Errorf("did not actually time out fast: %v", elapsed)
	}
}

func TestTimeoutMissingArgs(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"timeout"}); rc != 125 {
		t.Errorf("rc = %d; want 125", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTimeoutInvalidDuration(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"timeout", "not-a-duration", "true"}); rc != 125 {
		t.Errorf("rc = %d; want 125", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid duration") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTimeoutCommandNotFound(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	rc := Main([]string{"timeout", "5", "this-command-does-not-exist-xyz"})
	if rc != 127 && rc != 126 {
		t.Errorf("rc = %d; want 127 or 126", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestTimeoutPreserveStatus(t *testing.T) {
	if testing.Short() || runtime.GOOS == "windows" {
		t.Skip("timing-sensitive or Windows-specific")
	}
	args := []string{"timeout", "--preserve-status", "0.2", "sleep", "5"}
	testutil.CaptureStdio(t)
	rc := Main(args)
	// With --preserve-status we get whatever exit signal the killed
	// process reports; on Unix that is typically -1 (signal) which
	// exec reports as a non-zero, non-124 status.
	if rc == 124 {
		t.Errorf("rc = 124; --preserve-status should bypass the 124 override")
	}
}

func TestTimeoutDurationSuffixes(t *testing.T) {
	for _, in := range []string{"1", "1s", "100ms", "0.5", "1d"} {
		if _, err := parseDuration(in); err != nil {
			t.Errorf("%s: %v", in, err)
		}
	}
	for _, in := range []string{"", "-1", "garbage"} {
		if _, err := parseDuration(in); err == nil {
			t.Errorf("%s: expected error", in)
		}
	}
}
