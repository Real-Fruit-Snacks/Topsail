package rm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func writeFile(t *testing.T, p, contents string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRmFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	writeFile(t, p, "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("not removed: %v", err)
	}
}

func TestRmMultiple(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "")
	writeFile(t, b, "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", a, b}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestRmDirNoR(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", d}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "directory") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRmDirRecursive(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.MkdirAll(filepath.Join(d, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(d, "f"), "")
	writeFile(t, filepath.Join(d, "nested", "f2"), "")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-r", d}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(d); !os.IsNotExist(err) {
		t.Errorf("not removed: %v", err)
	}
}

func TestRmEmptyDir(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-d", d}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestRmForceMissing(t *testing.T) {
	dir := t.TempDir()
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-f", filepath.Join(dir, "missing")}); rc != 0 {
		t.Errorf("rc = %d; want 0 with -f on missing file", rc)
	}
}

func TestRmMissingNoForce(t *testing.T) {
	dir := t.TempDir()
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", filepath.Join(dir, "missing")}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}

func TestRmMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRmForceNoArgs(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-f"}); rc != 0 {
		t.Errorf("rc = %d; want 0 with -f no args", rc)
	}
}

func TestRmVerbose(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	writeFile(t, p, "")
	out, _ := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-v", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "removed") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestRmInvalidOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"rm", "-z", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
