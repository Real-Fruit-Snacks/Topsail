package tsort

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestTsortLinear(t *testing.T) {
	testutil.SetStdin(t, "a b\nb c\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"tsort"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	posA, posB, posC := -1, -1, -1
	for i, l := range lines {
		switch l {
		case "a":
			posA = i
		case "b":
			posB = i
		case "c":
			posC = i
		}
	}
	if posA == -1 || posB == -1 || posC == -1 {
		t.Errorf("missing nodes: %v", lines)
	}
	if posA >= posB || posB >= posC {
		t.Errorf("order violated: %v", lines)
	}
}

func TestTsortCycle(t *testing.T) {
	testutil.SetStdin(t, "a b\nb a\n")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tsort"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(errBuf.String(), "cycle") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTsortOddTokens(t *testing.T) {
	testutil.SetStdin(t, "a b c\n")
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"tsort"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "odd number") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
