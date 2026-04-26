package curl

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestCurlGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))
	defer srv.Close()
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"curl", srv.URL}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "hello world" {
		t.Errorf("got %q", got)
	}
}

func TestCurlOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("written to file"))
	}))
	defer srv.Close()
	dir := t.TempDir()
	dst := filepath.Join(dir, "out")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"curl", "-o", dst, srv.URL}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "written to file" {
		t.Errorf("got %q", got)
	}
}

func TestCurlHead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Foo", "bar")
		_, _ = w.Write([]byte("body"))
	}))
	defer srv.Close()
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"curl", "-I", srv.URL}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out.String(), "X-Foo: bar") {
		t.Errorf("got %q", out.String())
	}
}

func TestCurl404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer srv.Close()
	_, _ = testutil.CaptureStdio(t)
	if rc := Main([]string{"curl", srv.URL}); rc != 22 {
		t.Errorf("rc = %d; want 22", rc)
	}
}

func TestCurlMissingURL(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"curl"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "no URL") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCurlInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"curl", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCurlHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Test")))
	}))
	defer srv.Close()
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"curl", "-H", "X-Test: hello", srv.URL}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "hello" {
		t.Errorf("got %q", got)
	}
}
