package yes

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

// limitedWriter forwards writes to w until it has accepted limit bytes,
// after which every Write returns errBroken — simulating a closed pipe.
type limitedWriter struct {
	w     io.Writer
	limit int
	wrote int
}

var errBroken = errors.New("simulated broken pipe")

func (lw *limitedWriter) Write(p []byte) (int, error) {
	if lw.wrote >= lw.limit {
		return 0, errBroken
	}
	remaining := lw.limit - lw.wrote
	if len(p) > remaining {
		p = p[:remaining]
	}
	n, _ := lw.w.Write(p)
	lw.wrote += n
	return n, nil
}

func runYes(t *testing.T, limit int, argv []string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	lw := &limitedWriter{w: buf, limit: limit}
	testutil.SwapStdio(t, nil, lw, nil)
	if rc := Main(argv); rc != 0 {
		t.Errorf("rc = %d; want 0 on broken pipe", rc)
	}
	return buf.String()
}

func TestYesDefault(t *testing.T) {
	got := runYes(t, 6, []string{"yes"})
	if want := "y\ny\ny\n"; got != want {
		t.Errorf("yes output = %q; want %q", got, want)
	}
}

func TestYesWithArgs(t *testing.T) {
	got := runYes(t, 24, []string{"yes", "hello", "world"})
	if want := "hello world\nhello world\n"; got != want {
		t.Errorf("yes output = %q; want %q", got, want)
	}
}

func TestYesEmptyString(t *testing.T) {
	got := runYes(t, 3, []string{"yes", ""})
	if want := "\n\n\n"; got != want {
		t.Errorf("yes \"\" output = %q; want %q", got, want)
	}
}

func TestYesMultiArgsSpaceJoined(t *testing.T) {
	got := runYes(t, 8, []string{"yes", "a", "b"})
	if want := "a b\na b\n"; got != want {
		t.Errorf("yes a b output = %q; want %q", got, want)
	}
}
