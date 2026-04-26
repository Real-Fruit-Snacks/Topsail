package sha512sum

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSha512SumStdin(t *testing.T) {
	testutil.SetStdin(t, "hello")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"sha512sum"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.HasPrefix(got, "9b71d224bd62f3785d96d46ad3ea3d73") {
		t.Errorf("got %q", got)
	}
}

func TestSha512SumInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"sha512sum", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
