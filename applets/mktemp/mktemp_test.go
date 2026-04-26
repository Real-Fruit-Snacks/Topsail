package mktemp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestMktempCreatesFile(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"mktemp", "-p", dir}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimSpace(out.String())
	st, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat created file: %v", err)
	}
	if st.IsDir() {
		t.Errorf("expected file, got directory: %s", got)
	}
	if !strings.HasPrefix(got, dir) {
		t.Errorf("created path %q not under %q", got, dir)
	}
}

func TestMktempCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	Main([]string{"mktemp", "-d", "-p", dir})
	got := strings.TrimSpace(out.String())
	st, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat created dir: %v", err)
	}
	if !st.IsDir() {
		t.Errorf("expected directory, got file: %s", got)
	}
}

func TestMktempTemplate(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	Main([]string{"mktemp", "-p", dir, "build.XXXXXX"})
	got := strings.TrimSpace(out.String())
	if !strings.Contains(filepath.Base(got), "build.") {
		t.Errorf("expected basename to start with 'build.': %s", got)
	}
}

func TestMktempSuffix(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	Main([]string{"mktemp", "-p", dir, "--suffix=.log", "stage.XXXXXX"})
	got := strings.TrimSpace(out.String())
	if !strings.HasSuffix(got, ".log") {
		t.Errorf("expected .log suffix: %s", got)
	}
}

func TestMktempTooFewX(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mktemp", "bad.XX"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "too few X") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMktempDryRun(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"mktemp", "-u", "-p", dir, "stage.XXXXXX"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := strings.TrimSpace(out.String())
	if got == "" {
		t.Error("expected dry-run output")
	}
	// Dry run should NOT create the file/dir.
	if _, err := os.Stat(got); err == nil {
		t.Errorf("dry-run should not create %s", got)
	}
}

func TestMktempUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"mktemp", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestMktempQuiet(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	// Bad template silenced by -q.
	Main([]string{"mktemp", "-q", "bad"})
	if errBuf.Len() != 0 {
		t.Errorf("quiet should suppress stderr; got %q", errBuf.String())
	}
}
