package realpath

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestRealpathBasic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"realpath", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimSpace(out.String())
	want, _ := filepath.EvalSymlinks(p)
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRealpathMultiple(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a")
	p2 := filepath.Join(dir, "b")
	if err := os.WriteFile(p1, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"realpath", p1, p2})
	if !strings.Contains(out.String(), "a") || !strings.Contains(out.String(), "b") {
		t.Errorf("output missing entries: %q", out.String())
	}
}

func TestRealpathMissingFails(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"realpath", "/no/such/path"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestRealpathMissingTolerated(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"realpath", "-m", "/no/such/path"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "no") {
		t.Errorf("output: %q", out.String())
	}
}

func TestRealpathQuiet(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"realpath", "-q", "/no/such/path"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() != 0 {
		t.Errorf("quiet should suppress stderr; got %q", errBuf.String())
	}
}

func TestRealpathZeroTerminator(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"realpath", "-z", p})
	if !strings.HasSuffix(out.String(), "\x00") {
		t.Errorf("expected NUL terminator: %q", out.String())
	}
}

func TestRealpathStripSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks not reliably available on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks unavailable:", err)
	}

	// Without -s, link resolves to target.
	out := testutil.CaptureStdout(t)
	Main([]string{"realpath", link})
	got := strings.TrimSpace(out.String())
	wantResolved, _ := filepath.EvalSymlinks(link)
	if got != wantResolved {
		t.Errorf("default: got %q; want %q", got, wantResolved)
	}

	// With -s, link is left in place.
	out = testutil.CaptureStdout(t)
	Main([]string{"realpath", "-s", link})
	got = strings.TrimSpace(out.String())
	if !strings.HasSuffix(got, "link") {
		t.Errorf("-s: got %q; want path ending in 'link'", got)
	}
}

func TestRealpathMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"realpath"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
