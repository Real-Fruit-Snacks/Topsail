package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestFileEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"file", p})
	if !strings.Contains(out.String(), "empty") {
		t.Errorf("got %q", out.String())
	}
}

func TestFileText(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "text")
	if err := os.WriteFile(p, []byte("hello world\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"file", p})
	if !strings.Contains(out.String(), "ASCII") && !strings.Contains(out.String(), "UTF-8") {
		t.Errorf("got %q", out.String())
	}
}

func TestFileDirectory(t *testing.T) {
	dir := t.TempDir()
	out := testutil.CaptureStdout(t)
	Main([]string{"file", dir})
	if !strings.Contains(out.String(), "directory") {
		t.Errorf("got %q", out.String())
	}
}

func TestFilePNG(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "img.png")
	pngHeader := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0}
	if err := os.WriteFile(p, pngHeader, 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"file", p})
	if !strings.Contains(out.String(), "PNG") {
		t.Errorf("got %q", out.String())
	}
}

func TestFileGzip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "compressed")
	if err := os.WriteFile(p, []byte{0x1F, 0x8B, 0x08, 0x00}, 0o644); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"file", p})
	if !strings.Contains(out.String(), "gzip") {
		t.Errorf("got %q", out.String())
	}
}

func TestFileShellScript(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "s.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	out := testutil.CaptureStdout(t)
	Main([]string{"file", p})
	if !strings.Contains(out.String(), "shell script") {
		t.Errorf("got %q", out.String())
	}
}

func TestFileMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"file"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestFileNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"file", "/no/such/file"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
