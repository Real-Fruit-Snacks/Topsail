package gzip

import (
	"bytes"
	gzipPkg "compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestGzipStreamRoundtrip(t *testing.T) {
	original := "hello world from topsail gzip"
	testutil.SetStdin(t, original)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"gzip"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	gr, err := gzipPkg.NewReader(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("not gzip output: %v", err)
	}
	got, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Errorf("got %q; want %q", got, original)
	}
}

func TestGunzipStream(t *testing.T) {
	var buf bytes.Buffer
	gw := gzipPkg.NewWriter(&buf)
	_, _ = gw.Write([]byte("decompressed content"))
	_ = gw.Close()

	testutil.SetStdin(t, buf.String())
	out := testutil.CaptureStdout(t)
	if rc := GunzipMain([]string{"gunzip"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "decompressed content" {
		t.Errorf("got %q", got)
	}
}

func TestGzipFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(src, []byte("contents"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"gzip", "-k", src}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(src + ".gz"); err != nil {
		t.Errorf("missing .gz: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source removed despite -k: %v", err)
	}
}

func TestGzipFileNoKeep(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(src, []byte("contents"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"gzip", src}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be removed: %v", err)
	}
}

func TestGzipDecompressFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(src, []byte("contents"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	Main([]string{"gzip", src})
	// Now decompress.
	if rc := GunzipMain([]string{"gunzip", src + ".gz"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(src)
	if string(got) != "contents" {
		t.Errorf("got %q", got)
	}
}

func TestGzipInvalidGunzipName(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "no_suffix")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errBuf := testutil.CaptureStdio(t)
	if rc := GunzipMain([]string{"gunzip", src}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), ".gz suffix") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestGzipUnknownFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"gzip", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
