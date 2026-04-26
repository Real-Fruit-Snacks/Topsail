package factor

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestFactor(t *testing.T) {
	cases := map[string]string{
		"12": "12: 2 2 3",
		"7":  "7: 7",
		"1":  "1:",
		"60": "60: 2 2 3 5",
	}
	for in, want := range cases {
		out := testutil.CaptureStdout(t)
		if rc := Main([]string{"factor", in}); rc != 0 {
			t.Errorf("rc = %d for %q", rc, in)
		}
		if got := strings.TrimSpace(out.String()); got != want {
			t.Errorf("factor %s = %q; want %q", in, got, want)
		}
	}
}

func TestFactorInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"factor", "abc"}); rc != 1 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "valid positive integer") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestFactorStdin(t *testing.T) {
	testutil.SetStdin(t, "10\n15\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"factor"})
	got := out.String()
	if !strings.Contains(got, "10: 2 5") || !strings.Contains(got, "15: 3 5") {
		t.Errorf("got %q", got)
	}
}
