package shuf

import (
	"sort"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestShufBasic(t *testing.T) {
	testutil.SetStdin(t, "1\n2\n3\n4\n5\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"shuf"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines; want 5", len(lines))
	}
	sort.Strings(lines)
	for i, want := range []string{"1", "2", "3", "4", "5"} {
		if lines[i] != want {
			t.Errorf("missing %s in shuffled output", want)
		}
	}
}

func TestShufRange(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"shuf", "-i", "1-5"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 5 {
		t.Errorf("got %d lines; want 5", len(lines))
	}
}

func TestShufHead(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"shuf", "-n", "2", "-i", "1-10"})
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines; want 2", len(lines))
	}
}

func TestShufEcho(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"shuf", "-e", "alpha", "beta", "gamma"})
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines; want 3", len(lines))
	}
}

func TestShufInvalidRange(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"shuf", "-i", "10-1"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid range") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
