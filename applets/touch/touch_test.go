package touch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestTouchCreate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "new.txt")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestTouchNoCreate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "x.txt")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-c", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("file should not exist with -c: %v", err)
	}
}

func TestTouchUpdatesMtime(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "f")
	if err := os.WriteFile(target, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(target, old, old); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.ModTime().Before(old.Add(30 * time.Minute)) {
		t.Errorf("mtime not updated: %v (was %v)", info.ModTime(), old)
	}
}

func TestTouchReference(t *testing.T) {
	dir := t.TempDir()
	ref := filepath.Join(dir, "ref")
	target := filepath.Join(dir, "tgt")
	if err := os.WriteFile(ref, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	refTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(ref, refTime, refTime); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-r", ref, target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	got := info.ModTime().UTC().Truncate(time.Second)
	if !got.Equal(refTime) {
		t.Errorf("mtime = %v; want %v", got, refTime)
	}
}

func TestTouchDate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dated")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-d", "2024-06-15T10:30:00Z", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if !info.ModTime().UTC().Equal(want) {
		t.Errorf("mtime = %v; want %v", info.ModTime().UTC(), want)
	}
}

func TestTouchInvalidDate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "x")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-d", "not a date", target}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid date") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTouchMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "missing operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTouchReferenceMissing(t *testing.T) {
	dir := t.TempDir()
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-r", "/no/such/ref", filepath.Join(dir, "x")}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr message")
	}
}
