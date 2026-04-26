package cp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func writeFile(t *testing.T, p, contents string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCpFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	writeFile(t, src, "alpha")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "alpha" {
		t.Errorf("got %q", got)
	}
}

func TestCpIntoDirectory(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	writeFile(t, src, "x")
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", src, subdir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(filepath.Join(subdir, "a")); err != nil {
		t.Errorf("not copied: %v", err)
	}
}

func TestCpMultiple(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "1")
	writeFile(t, b, "2")
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", a, b, subdir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, name := range []string{"a", "b"} {
		if _, err := os.Stat(filepath.Join(subdir, name)); err != nil {
			t.Errorf("%s not copied: %v", name, err)
		}
	}
}

func TestCpRecursive(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(srcDir, "f1"), "1")
	writeFile(t, filepath.Join(srcDir, "nested", "f2"), "2")

	dstDir := filepath.Join(dir, "dst")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-r", srcDir, dstDir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, p := range []string{
		filepath.Join(dstDir, "f1"),
		filepath.Join(dstDir, "nested", "f2"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("not copied: %s (%v)", p, err)
		}
	}
}

func TestCpDirectoryWithoutR(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dstDir := filepath.Join(dir, "dst")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", srcDir, dstDir}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "-r") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCpNoClobber(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-n", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "old" {
		t.Errorf("dst overwritten despite -n: %q", got)
	}
}

func TestCpMissingDest(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCpVerbose(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	writeFile(t, src, "x")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-v", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "->") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestCpPreserve(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a")
	dst := filepath.Join(dir, "b")
	writeFile(t, src, "x")
	// Set a known mtime on src.
	want := time.Unix(1577836800, 0) // 2020-01-01 UTC
	if err := os.Chtimes(src, want, want); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"cp", "-p", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(want) {
		t.Errorf("mtime = %v; want %v", info.ModTime(), want)
	}
}
