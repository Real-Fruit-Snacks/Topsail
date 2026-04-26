package cp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestCpVerboseDirectory verifies the verbose log line for the directory itself.
func TestCpVerboseDirectory(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(srcDir, "f"), "x")
	dstDir := filepath.Join(dir, "dst")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-rv", srcDir, dstDir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "->") {
		t.Errorf("verbose output missing: %q", out.String())
	}
}

// TestCpRecursivePreserve covers the preserve+recursive Chtimes branches in copyDir.
func TestCpRecursivePreserve(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(srcDir, "f"), "x")
	want := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, p := range []string{srcDir, filepath.Join(srcDir, "nested"), filepath.Join(srcDir, "f")} {
		if err := os.Chtimes(p, want, want); err != nil {
			t.Fatal(err)
		}
	}
	dstDir := filepath.Join(dir, "dst")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-rp", srcDir, dstDir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(filepath.Join(dstDir, "f"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(want) {
		t.Errorf("file mtime = %v; want %v", info.ModTime(), want)
	}
}

// TestCpCombinedFlags exercises -rfvi (interactive is no-op).
func TestCpCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	writeFile(t, src, "x")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-rfvi", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "->") {
		t.Errorf("verbose output missing: %q", out.String())
	}
}

// TestCpInvalidCombinedFlag exercises the unknown-char error in the cluster.
func TestCpInvalidCombinedFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-rZ", "/tmp/a", "/tmp/b"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestCpDoubleDash verifies "--" stops flag parsing.
func TestCpDoubleDash(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "x")
	dst := filepath.Join(dir, "-y")
	writeFile(t, src, "x")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "--", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestCpNoClobberDirect calls copyFile directly with noClobber on an
// existing destination so we cover the "destination exists, skip" branch.
func TestCpNoClobberDirect(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")
	if err := copyFile(src, dst, 0o644, options{noClobber: true}); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "old" {
		t.Errorf("dst overwritten despite noClobber: %q", got)
	}
}

// TestCpFileBadDest covers copyFile's OpenFile error path.
func TestCpFileBadDest(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "x")
	if err := copyFile(src, "/no/such/dir/dst", 0o644, options{}); err == nil {
		t.Error("expected error for unwritable destination")
	}
}

// TestCpDirBadSrc covers copyDir's Stat error path.
func TestCpDirBadSrc(t *testing.T) {
	if err := copyDir("/no/such/source/dir", "/tmp/dst", options{}); err == nil {
		t.Error("expected error for missing source dir")
	}
}

// TestCpEntryBadSrc covers copyEntry's Lstat error path.
func TestCpEntryBadSrc(t *testing.T) {
	if err := copyEntry("/no/such/source", "/tmp/dst", options{}); err == nil {
		t.Error("expected error for missing source")
	}
}
