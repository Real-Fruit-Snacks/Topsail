package sleep

import (
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSleepShort(t *testing.T) {
	testutil.CaptureStdio(t)
	start := time.Now()
	if rc := Main([]string{"sleep", "0.1"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if elapsed := time.Since(start); elapsed < 80*time.Millisecond {
		t.Errorf("elapsed = %v; want >= ~100ms", elapsed)
	}
}

func TestSleepZero(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"sleep", "0"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestSleepMultiple(t *testing.T) {
	testutil.CaptureStdio(t)
	start := time.Now()
	if rc := Main([]string{"sleep", "0.05", "0.05"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if elapsed := time.Since(start); elapsed < 80*time.Millisecond {
		t.Errorf("elapsed = %v; want >= sum of 50ms+50ms", elapsed)
	}
}

func TestSleepUnits(t *testing.T) {
	testutil.CaptureStdio(t)
	cases := []string{"0.05s", "0"}
	for _, s := range cases {
		if rc := Main([]string{"sleep", s}); rc != 0 {
			t.Errorf("Main %q rc = %d", s, rc)
		}
	}
}

func TestSleepInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sleep", "abc"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestSleepNegative(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sleep", "-1"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestSleepNoArgs(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sleep"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
