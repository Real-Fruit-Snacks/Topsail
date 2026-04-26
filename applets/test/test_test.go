package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestEmpty(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
}

func TestSingleArg(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "hello"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", ""}); rc != 1 {
		t.Errorf("empty string rc = %d", rc)
	}
}

func TestStringEqual(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "abc", "=", "abc"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "abc", "=", "xyz"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
}

func TestStringNotEqual(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "abc", "!=", "xyz"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestZAndN(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "-z", ""}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-z", "x"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-n", "x"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestIntegerCompare(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "5", "-eq", "5"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "5", "-lt", "10"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "5", "-gt", "10"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "-e", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-e", filepath.Join(dir, "missing")}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
}

func TestFileType(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "-f", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-d", p}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-d", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

func TestFileSize(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	_ = os.WriteFile(a, []byte("data"), 0o644)
	_ = os.WriteFile(b, nil, 0o644)
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "-s", a}); rc != 0 {
		t.Errorf("-s on non-empty: %d", rc)
	}
	if rc := Main([]string{"test", "-s", b}); rc != 1 {
		t.Errorf("-s on empty: %d", rc)
	}
}

func TestNot(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "!", "-z", "x"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "!", "5", "-eq", "5"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
}

func TestAndOr(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "-n", "x", "-a", "-z", ""}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-z", "x", "-o", "-n", "y"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"test", "-z", "x", "-a", "-n", "y"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
}

func TestBracketAlias(t *testing.T) {
	testutil.CaptureStdio(t)
	if rc := Main([]string{"[", "5", "-eq", "5", "]"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if rc := Main([]string{"[", "5", "-eq", "5"}); rc != 2 {
		t.Errorf("missing ']' rc = %d", rc)
	}
}

func TestSyntaxError(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"test", "5", "-eq", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
