package df

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestDfBasic(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"df", "."}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "Filesystem") {
		t.Errorf("missing header in %q", got)
	}
}

func TestDfHuman(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"df", "-h", "."}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "Size") {
		t.Errorf("missing human header: %q", got)
	}
}

func TestDfNonexistent(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"df", "/no/such/path/at/all"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestDfInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"df", "-Z"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
