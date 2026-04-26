package tee

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestTeeStdoutOnly(t *testing.T) {
	testutil.SetStdin(t, "hello\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tee"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "hello\n" {
		t.Errorf("got %q", got)
	}
}

func TestTeeToFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out")
	testutil.SetStdin(t, "x\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tee", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "x\n" {
		t.Errorf("stdout = %q", got)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "x\n" {
		t.Errorf("file = %q", got)
	}
}

func TestTeeMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	testutil.SetStdin(t, "data\n")
	testutil.CaptureStdout(t)
	if rc := Main([]string{"tee", a, b}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, p := range []string{a, b} {
		got, _ := os.ReadFile(p)
		if string(got) != "data\n" {
			t.Errorf("%s = %q", p, got)
		}
	}
}

func TestTeeAppend(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log")
	if err := os.WriteFile(p, []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.SetStdin(t, "second\n")
	testutil.CaptureStdout(t)
	if rc := Main([]string{"tee", "-a", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "first\nsecond\n" {
		t.Errorf("got %q", got)
	}
}

func TestTeeOverwriteByDefault(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log")
	if err := os.WriteFile(p, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.SetStdin(t, "new\n")
	testutil.CaptureStdout(t)
	Main([]string{"tee", p})
	got, _ := os.ReadFile(p)
	if string(got) != "new\n" {
		t.Errorf("got %q; want \"new\\n\"", got)
	}
}

func TestTeeBadPath(t *testing.T) {
	testutil.SetStdin(t, "x")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tee", "/this/does/not/exist/file"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "tee:") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTeeInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tee", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
