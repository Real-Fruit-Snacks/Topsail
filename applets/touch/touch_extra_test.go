package touch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// TestTouchOnlyAccess covers the onlyAccess branch in touchOne.
func TestTouchOnlyAccess(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "f")
	if err := os.WriteFile(target, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(target, old, old); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-a", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	// With -a only, mtime should be left at the old value (we read
	// info.ModTime() and pass it in as the new mtime).
	if !info.ModTime().UTC().Equal(old) {
		t.Errorf("mtime changed despite -a: got %v; want %v", info.ModTime().UTC(), old)
	}
}

// TestTouchOnlyModification covers the onlyMod branch in touchOne.
func TestTouchOnlyModification(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "f")
	if err := os.WriteFile(target, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(target, old, old); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-m", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	// -m should bump the mtime to "now" (≥ old + significant delta).
	if info.ModTime().Before(old.Add(24 * time.Hour)) {
		t.Errorf("mtime not bumped: %v vs old %v", info.ModTime(), old)
	}
}

// TestTouchLongFlags covers --no-create / --reference= / --date=.
func TestTouchLongFlags(t *testing.T) {
	dir := t.TempDir()
	ref := filepath.Join(dir, "ref")
	if err := os.WriteFile(ref, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	refTime := time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(ref, refTime, refTime); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "tgt")
	if err := os.WriteFile(target, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "--reference=" + ref, target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(target)
	if !info.ModTime().UTC().Truncate(time.Second).Equal(refTime) {
		t.Errorf("--reference= ignored: %v", info.ModTime())
	}
}

// TestTouchLongDate covers --date= long form.
func TestTouchLongDate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "f")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "--date=2024-06-15T10:30:00Z", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	info, _ := os.Stat(target)
	want := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if !info.ModTime().UTC().Equal(want) {
		t.Errorf("mtime = %v; want %v", info.ModTime().UTC(), want)
	}
}

// TestTouchCombinedFlags exercises -amc as a cluster.
func TestTouchCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "missing")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-amc", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("file should not be created with -c: %v", err)
	}
}

// TestTouchInvalidCombinedFlag exercises the unknown-char error in cluster.
func TestTouchInvalidCombinedFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-aZ", "/tmp/x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestTouchDoubleDash verifies "--" stops flag parsing.
func TestTouchDoubleDash(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "-weird")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "--", target}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
}

// TestTouchRefMissingArg / TestTouchDateMissingArg cover the missing-arg paths.
func TestTouchRefMissingArg(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-r"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "requires an argument") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTouchDateMissingArg(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"touch", "-d"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "requires an argument") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

// TestTouchInvalidParseDate hits parseDate's "no formats matched" path
// directly via the public Main entry point.
func TestTouchInvalidParseDate(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "x")
	_, errBuf := testutil.CaptureStdio(t)
	// completely garbled date should fail to parse with all candidate formats
	if rc := Main([]string{"touch", "-d", "garbage-date-string", target}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid date") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
