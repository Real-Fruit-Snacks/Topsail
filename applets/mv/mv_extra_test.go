package mv

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestMvCopyFileDirect exercises copyFile directly so the helper isn't
// only reachable via the cross-device fallback path.
func TestMvCopyFileDirect(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst, 0o644); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "payload" {
		t.Errorf("got %q; want payload", got)
	}
}

// TestMvCopyFileBadSrc covers copyFile's open-error path.
func TestMvCopyFileBadSrc(t *testing.T) {
	if err := copyFile("/no/such/source", "/tmp/x", 0o644); err == nil {
		t.Error("expected error for missing source")
	}
}

// TestMvCopyFileBadDst covers copyFile's destination-create error path.
func TestMvCopyFileBadDst(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, "/no/such/dir/dst", 0o644); err == nil {
		t.Error("expected error for unwritable destination")
	}
}

// TestMvCrossDeviceFallback simulates a rename failure (as if EXDEV)
// and asserts move() falls back to copy+delete.
func TestMvCrossDeviceFallback(t *testing.T) {
	orig := renameFunc
	renameFunc = func(string, string) error {
		return errors.New("simulated EXDEV")
	}
	t.Cleanup(func() { renameFunc = orig })

	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := move(src, dst); err != nil {
		t.Fatalf("move: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "payload" {
		t.Errorf("dst content = %q", got)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src still present: %v", err)
	}
}

// TestMvCrossDeviceDirRefused asserts that the cross-device fallback
// refuses directory moves (we don't recursively copy dirs in the fallback).
func TestMvCrossDeviceDirRefused(t *testing.T) {
	orig := renameFunc
	renameFunc = func(string, string) error {
		return errors.New("simulated EXDEV")
	}
	t.Cleanup(func() { renameFunc = orig })

	dir := t.TempDir()
	src := filepath.Join(dir, "srcdir")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dstdir")
	if err := move(src, dst); err == nil {
		t.Error("expected cross-device directory move to fail")
	} else if !strings.Contains(err.Error(), "cross-device") {
		t.Errorf("err = %v; want 'cross-device'", err)
	}
}

// TestMvCombinedFlags exercises the short-flag cluster -fv path.
func TestMvCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "alpha")
	dst := filepath.Join(dir, "b")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "-fv", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "renamed") {
		t.Errorf("expected verbose output; got %q", out.String())
	}
}

// TestMvCombinedFlagInvalid exercises the short-flag cluster error path.
func TestMvCombinedFlagInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "-fX", "/tmp/a", "/tmp/b"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestMvDoubleDash verifies "--" stops flag parsing.
func TestMvDoubleDash(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "src", "x")
	dst := filepath.Join(dir, "-dash")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "--", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("dst missing: %v", err)
	}
}

// TestMvInteractiveNoOp confirms -i is accepted but never prompts.
func TestMvInteractiveNoOp(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "x")
	dst := filepath.Join(dir, "b")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "-i", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}
