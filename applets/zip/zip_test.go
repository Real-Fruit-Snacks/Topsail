package zip

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestZipRoundtrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(src, []byte("zipped!"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(dir, "out.zip")
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)

	testutil.CaptureStdio(t)
	if rc := Main([]string{"zip", archive, "f.txt"}); rc != 0 {
		t.Errorf("zip rc = %d", rc)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive not created: %v", err)
	}

	out := filepath.Join(dir, "extract")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if rc := UnzipMain([]string{"unzip", "-d", out, archive}); rc != 0 {
		t.Errorf("unzip rc = %d", rc)
	}
	got, err := os.ReadFile(filepath.Join(out, "f.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "zipped!" {
		t.Errorf("got %q", got)
	}
}

func TestZipRecursive(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "sub", "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "r.zip")
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)
	testutil.CaptureStdio(t)
	if rc := Main([]string{"zip", "-r", archive, "src"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Errorf("missing archive: %v", err)
	}
}

func TestZipRecursiveRequired(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"zip", "out.zip", "src"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "-r required") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestZipMissingArgs(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"zip", "out.zip"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUnzipList(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "a.zip")
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	_ = os.Chdir(dir)
	testutil.CaptureStdio(t)
	Main([]string{"zip", archive, "f.txt"})

	out := testutil.CaptureStdout(t)
	if rc := UnzipMain([]string{"unzip", "-l", archive}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "f.txt") {
		t.Errorf("got %q", out.String())
	}
}

func TestUnzipMissingArchive(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := UnzipMain([]string{"unzip"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
