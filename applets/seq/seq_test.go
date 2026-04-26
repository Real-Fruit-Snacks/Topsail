package seq

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestSeqLast(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "3"})
	if got, want := out.String(), "1\n2\n3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqFirstLast(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "5", "7"})
	if got, want := out.String(), "5\n6\n7\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqIncrement(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "0", "2", "10"})
	if got, want := out.String(), "0\n2\n4\n6\n8\n10\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqDescending(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "5", "-1", "3"})
	if got, want := out.String(), "5\n4\n3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqSeparator(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "-s", ",", "3"})
	if got, want := out.String(), "1,2,3\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqEqualWidth(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "-w", "8", "10"})
	if got, want := out.String(), "08\n09\n10\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqFormat(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "-f", "%03.0f", "1", "3"})
	if got, want := out.String(), "001\n002\n003\n"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSeqInvalid(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"seq", "abc"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSeqZeroIncrement(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"seq", "1", "0", "10"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "Zero increment") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSeqEmpty(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"seq", "5", "3"}) // first > last with positive inc -> empty
	if got := out.String(); got != "" {
		t.Errorf("got %q; want empty", got)
	}
}

func TestSeqMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"seq"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}
