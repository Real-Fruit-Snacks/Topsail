package fold

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestFoldBasic(t *testing.T) {
	testutil.SetStdin(t, "abcdefghijklmnop\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"fold", "-w", "5"})
	if got, want := out.String(), "abcde\nfghij\nklmno\np\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestFoldSpaces(t *testing.T) {
	testutil.SetStdin(t, "the quick brown fox\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"fold", "-w", "10", "-s"})
	got := out.String()
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if len(line) > 10 {
			t.Errorf("line %q exceeds width", line)
		}
	}
}

func TestFoldShortLine(t *testing.T) {
	testutil.SetStdin(t, "hi\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"fold", "-w", "100"})
	if got, want := out.String(), "hi\n"; got != want {
		t.Errorf("got %q", got)
	}
}

func TestFoldInvalidWidth(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"fold", "-w", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid width") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestFoldInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"fold", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
