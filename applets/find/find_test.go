package find

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub", "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "b.log"),
		filepath.Join(dir, "sub", "c.txt"),
		filepath.Join(dir, "sub", "deeper", "d.log"),
	} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFindBasic(t *testing.T) {
	dir := setup(t)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"find", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	for _, want := range []string{"a.txt", "b.log", "c.txt", "d.log"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %q", want, got)
		}
	}
}

func TestFindName(t *testing.T) {
	dir := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"find", dir, "-name", "*.txt"})
	got := out.String()
	if !strings.Contains(got, "a.txt") || !strings.Contains(got, "c.txt") {
		t.Errorf("missing matches in %q", got)
	}
	if strings.Contains(got, "b.log") {
		t.Errorf("did not want .log in %q", got)
	}
}

func TestFindType(t *testing.T) {
	dir := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"find", dir, "-type", "d"})
	got := out.String()
	if !strings.Contains(got, "sub") {
		t.Errorf("missing sub in %q", got)
	}
	if strings.Contains(got, "a.txt") {
		t.Errorf("file should not match -type d in %q", got)
	}
}

func TestFindMaxDepth(t *testing.T) {
	dir := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"find", dir, "-maxdepth", "1"})
	got := out.String()
	if strings.Contains(got, "deeper") {
		t.Errorf("deeper should be excluded with -maxdepth 1: %q", got)
	}
}

func TestFindMinDepth(t *testing.T) {
	dir := setup(t)
	out := testutil.CaptureStdout(t)
	Main([]string{"find", dir, "-mindepth", "1"})
	got := out.String()
	// At depth 0, dir itself shouldn't appear with mindepth 1.
	for _, line := range strings.Split(got, "\n") {
		if line == dir {
			t.Errorf("root should not appear with -mindepth 1: %q", got)
		}
	}
}

func TestFindUnknownPredicate(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"find", ".", "-bogus"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown predicate") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestFindNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"find", "/no/such/path"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
