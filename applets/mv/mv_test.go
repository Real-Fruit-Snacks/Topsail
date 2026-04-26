package mv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMvRename(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "alpha")
	dst := filepath.Join(dir, "b")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src still exists: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "alpha" {
		t.Errorf("dst content = %q", got)
	}
}

func TestMvIntoDirectory(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "x")
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", src, subdir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	moved := filepath.Join(subdir, "a")
	if _, err := os.Stat(moved); err != nil {
		t.Errorf("file not in subdir: %v", err)
	}
}

func TestMvMultipleIntoDir(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1")
	b := writeFile(t, dir, "b", "2")
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", a, b, subdir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, name := range []string{"a", "b"} {
		if _, err := os.Stat(filepath.Join(subdir, name)); err != nil {
			t.Errorf("%s not moved: %v", name, err)
		}
	}
}

func TestMvMultipleNonDir(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "1")
	b := writeFile(t, dir, "b", "2")
	dst := writeFile(t, dir, "c", "3")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", a, b, dst}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "is not a directory") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMvNoClobber(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "alpha")
	dst := writeFile(t, dir, "b", "existing")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "-n", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "existing" {
		t.Errorf("dst overwritten despite -n: %q", got)
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("src missing: %v", err)
	}
}

func TestMvVerbose(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "x", "1")
	dst := filepath.Join(dir, "y")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "-v", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "renamed") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestMvMissingDest(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing destination") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMvNonexistentSource(t *testing.T) {
	dir := t.TempDir()
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", filepath.Join(dir, "missing"), filepath.Join(dir, "x")}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

func TestMvOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "a", "new")
	dst := writeFile(t, dir, "b", "old")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"mv", src, dst}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "new" {
		t.Errorf("overwrite failed: %q", got)
	}
}
