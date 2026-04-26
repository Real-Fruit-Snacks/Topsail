package nproc

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestNprocDefault(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"nproc"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		t.Fatalf("parse output %q: %v", out.String(), err)
	}
	if got != runtime.NumCPU() {
		t.Errorf("got %d; want %d", got, runtime.NumCPU())
	}
}

func TestNprocIgnore(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"nproc", "--ignore=1"})
	got, _ := strconv.Atoi(strings.TrimSpace(out.String()))
	want := runtime.NumCPU() - 1
	if want < 1 {
		want = 1
	}
	if got != want {
		t.Errorf("got %d; want %d", got, want)
	}
}

func TestNprocIgnoreLargeClamped(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"nproc", "--ignore=9999"})
	if got := strings.TrimSpace(out.String()); got != "1" {
		t.Errorf("got %q; want 1", got)
	}
}

func TestNprocAll(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"nproc", "--all"})
	if got := strings.TrimSpace(out.String()); got == "" {
		t.Error("expected non-empty output")
	}
}

func TestNprocBadIgnore(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"nproc", "--ignore=foo"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestNprocUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"nproc", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
