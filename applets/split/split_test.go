package split

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSplitLines(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "input")
	if err := os.WriteFile(src, []byte("a\nb\nc\nd\ne\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)

	testutil.CaptureStdio(t)
	if rc := Main([]string{"split", "-l", "2", src, "piece"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, expected := range []string{"pieceaa", "pieceab", "pieceac"} {
		if _, err := os.Stat(filepath.Join(dir, expected)); err != nil {
			t.Errorf("missing %s: %v", expected, err)
		}
	}
}

func TestSplitNumeric(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "input")
	if err := os.WriteFile(src, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)
	testutil.CaptureStdio(t)
	Main([]string{"split", "-l", "1", "-d", src, "n"})
	for _, expected := range []string{"n00", "n01", "n02"} {
		if _, err := os.Stat(filepath.Join(dir, expected)); err != nil {
			t.Errorf("missing %s: %v", expected, err)
		}
	}
}

func TestSplitInvalidL(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"split", "-l", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid -l") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSplitInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"split", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
