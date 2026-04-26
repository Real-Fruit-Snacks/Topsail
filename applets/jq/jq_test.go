package jq

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestJqIdentity(t *testing.T) {
	testutil.SetStdin(t, `{"name":"alice","age":30}`)
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"jq", "."}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "alice") || !strings.Contains(got, "30") {
		t.Errorf("got %q", got)
	}
}

func TestJqField(t *testing.T) {
	testutil.SetStdin(t, `{"name":"alice","age":30}`)
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", ".name"})
	if !strings.Contains(out.String(), "alice") {
		t.Errorf("got %q", out.String())
	}
}

func TestJqRaw(t *testing.T) {
	testutil.SetStdin(t, `{"name":"alice"}`)
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", "-r", ".name"})
	if got := strings.TrimSpace(out.String()); got != "alice" {
		t.Errorf("got %q (want alice without quotes)", got)
	}
}

func TestJqCompact(t *testing.T) {
	testutil.SetStdin(t, `{"a":1,"b":2}`)
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", "-c", "."})
	got := strings.TrimSpace(out.String())
	if strings.Contains(got, "  ") {
		t.Errorf("expected compact, got %q", got)
	}
}

func TestJqNullInput(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", "-n", "1+2"})
	if got := strings.TrimSpace(out.String()); got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestJqSlurp(t *testing.T) {
	testutil.SetStdin(t, `1 2 3`)
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", "-s", "length"})
	if got := strings.TrimSpace(out.String()); got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestJqArray(t *testing.T) {
	testutil.SetStdin(t, `[1,2,3]`)
	out := testutil.CaptureStdout(t)
	Main([]string{"jq", ".[1]"})
	if got := strings.TrimSpace(out.String()); got != "2" {
		t.Errorf("got %q", got)
	}
}

func TestJqMissingFilter(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"jq"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing filter") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestJqParseError(t *testing.T) {
	testutil.SetStdin(t, `{}`)
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"jq", "..."}); rc != 3 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestJqInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"jq", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
