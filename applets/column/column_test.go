package column

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestColumnTable(t *testing.T) {
	testutil.SetStdin(t, "alice\t30\tNY\nbob\t25\tCA\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"column", "-t"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	for _, want := range []string{"alice", "bob", "NY", "CA"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %q", want, got)
		}
	}
}

func TestColumnSep(t *testing.T) {
	testutil.SetStdin(t, "a:b:c\n1:2:3\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"column", "-t", "-s", ":"})
	got := out.String()
	if !strings.Contains(got, "a") || !strings.Contains(got, "3") {
		t.Errorf("got %q", got)
	}
}

func TestColumnPassthrough(t *testing.T) {
	testutil.SetStdin(t, "line one\nline two\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"column"})
	if got, want := out.String(), "line one\nline two\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestColumnInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"column", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
