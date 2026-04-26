package time

import (
	"runtime"
	"strings"
	"testing"
	stdtime "time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func helper(t *testing.T) []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c", "exit", "0"}
	}
	return []string{"true"}
}

func TestTimeRunsAndReports(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main(append([]string{"time"}, helper(t)...)); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := errBuf.String()
	for _, kw := range []string{"real", "user", "sys"} {
		if !strings.Contains(got, kw) {
			t.Errorf("missing %q in summary: %q", kw, got)
		}
	}
}

func TestTimeMissingCommand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"time"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing COMMAND") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTimeCommandNotFound(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	rc := Main([]string{"time", "this-command-does-not-exist-xyz"})
	if rc != 127 && rc != 126 {
		t.Errorf("rc = %d; want 127 or 126", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestTimeFormatDuration(t *testing.T) {
	cases := []struct {
		in   stdtime.Duration
		want string
	}{
		{0, "0m0.000s"},
		{500 * stdtime.Millisecond, "0m0.500s"},
		{stdtime.Second, "0m1.000s"},
		{61 * stdtime.Second, "1m1.000s"},
		{125*stdtime.Second + 250*stdtime.Millisecond, "2m5.250s"},
		{-1 * stdtime.Second, "0m0.000s"},
	}
	for _, c := range cases {
		got := formatDuration(c.in)
		if got != c.want {
			t.Errorf("formatDuration(%v) = %q; want %q", c.in, got, c.want)
		}
	}
}
