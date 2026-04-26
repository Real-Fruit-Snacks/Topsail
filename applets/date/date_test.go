package date

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestDateDefault(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"date"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if strings.TrimSpace(got) == "" {
		t.Errorf("expected non-empty output")
	}
}

func TestDateFormat(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"date", "+%Y-%m-%d"})
	got := strings.TrimSpace(out.String())
	if len(got) != 10 || got[4] != '-' || got[7] != '-' {
		t.Errorf("expected YYYY-MM-DD; got %q", got)
	}
}

func TestDateUTC(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"date", "-u", "+%Z"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimSpace(out.String())
	if got != "UTC" {
		t.Errorf("expected UTC; got %q", got)
	}
}

func TestDateISO(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"date", "-I"})
	got := strings.TrimSpace(out.String())
	if len(got) != 10 || got[4] != '-' {
		t.Errorf("expected ISO date; got %q", got)
	}
}

func TestDateParse(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"date", "-u", "-d", "2024-06-15T10:30:00Z", "+%Y-%m-%d %H:%M"})
	got := strings.TrimSpace(out.String())
	if got != "2024-06-15 10:30" {
		t.Errorf("got %q", got)
	}
}

func TestDateBadParse(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"date", "-d", "not a date"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid date") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestDateInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"date", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
