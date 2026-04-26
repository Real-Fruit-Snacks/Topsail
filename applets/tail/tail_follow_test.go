package tail

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

// syncBuf is a goroutine-safe bytes.Buffer wrapper. Follow tests run
// tail.Main in one goroutine while another goroutine appends to the
// file under test; the same Stdout buffer is read from the test
// goroutine for assertions, so unsynchronised access would race.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// driveFollow installs a fast poll interval and a cancellable context
// for the duration of the test, returning the cancel func. Without this,
// follow tests would either wait the default 1s per iteration or never
// terminate.
func driveFollow(t *testing.T) (cancel context.CancelFunc) {
	t.Helper()
	origInterval := followInterval
	followInterval = 5 * time.Millisecond

	ctx, c := context.WithCancel(context.Background())
	origNew := newFollowCtx
	newFollowCtx = func() (context.Context, context.CancelFunc) {
		return ctx, c
	}

	t.Cleanup(func() {
		c()
		followInterval = origInterval
		newFollowCtx = origNew
	})
	return c
}

// installSyncStdio wires Stdout to a syncBuf for the test's lifetime.
func installSyncStdio(t *testing.T) (out, errBuf *syncBuf) {
	t.Helper()
	out = &syncBuf{}
	errBuf = &syncBuf{}
	origOut, origErr := ioutil.Stdout, ioutil.Stderr
	ioutil.Stdout = out
	ioutil.Stderr = errBuf
	t.Cleanup(func() {
		ioutil.Stdout = origOut
		ioutil.Stderr = origErr
	})
	return out, errBuf
}

// waitFor polls the buffer until it contains s, or fails the test after
// the deadline.
func waitFor(t *testing.T, buf *syncBuf, s string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), s) {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("did not see %q in output within %v\nbuffer: %q", s, timeout, buf.String())
}

func TestTailFollowAppend(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "first\nsecond\n")

	cancel := driveFollow(t)
	out, _ := installSyncStdio(t)

	done := make(chan int, 1)
	go func() {
		done <- Main([]string{"tail", "-f", "-n", "2", p})
	}()

	// Initial tail prints first/second.
	waitFor(t, out, "second\n", 2*time.Second)

	// Append more data and verify it shows up.
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("third\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	waitFor(t, out, "third\n", 2*time.Second)

	cancel()
	select {
	case rc := <-done:
		if rc != 0 {
			t.Errorf("rc = %d; want 0", rc)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tail -f did not exit after cancel")
	}
}

func TestTailFollowTruncate(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "alpha\nbeta\n")

	cancel := driveFollow(t)
	out, errBuf := installSyncStdio(t)

	done := make(chan int, 1)
	go func() {
		done <- Main([]string{"tail", "-f", p})
	}()

	waitFor(t, out, "beta\n", 2*time.Second)

	// Truncate and write fresh content.
	if err := os.WriteFile(p, []byte("gamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	waitFor(t, errBuf, "file truncated", 2*time.Second)
	waitFor(t, out, "gamma\n", 2*time.Second)

	cancel()
	<-done
}

func TestTailFollowMultiFileHeaders(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a.log", "")
	b := writeFile(t, dir, "b.log", "")

	cancel := driveFollow(t)
	out, _ := installSyncStdio(t)

	done := make(chan int, 1)
	go func() {
		done <- Main([]string{"tail", "-f", a, b})
	}()

	// The initial-tail phase emits headers for both empty files; wait for
	// the second header to land so we know tailOne has finished and the
	// follow loop has had a chance to open both files and seek-to-end.
	// Without this sync, an append may race the follow loop's setup and
	// land past the saved position.
	waitFor(t, out, "==> "+b+" <==", 2*time.Second)
	time.Sleep(50 * time.Millisecond)

	// Append to a, then to b — each should produce a header on first write.
	appendStr(t, a, "from-a\n")
	waitFor(t, out, "from-a\n", 2*time.Second)
	appendStr(t, b, "from-b\n")
	waitFor(t, out, "from-b\n", 2*time.Second)

	got := out.String()
	if !strings.Contains(got, "==> "+a+" <==") {
		t.Errorf("missing header for a; got %q", got)
	}
	if !strings.Contains(got, "==> "+b+" <==") {
		t.Errorf("missing header for b; got %q", got)
	}

	cancel()
	<-done
}

func TestTailSleepIntervalParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1", time.Second},
		{"0.5", 500 * time.Millisecond},
		{"250ms", 250 * time.Millisecond},
	}
	for _, c := range cases {
		got, err := parseSleep(c.in)
		if err != nil {
			t.Errorf("%s: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %v; want %v", c.in, got, c.want)
		}
	}
	if _, err := parseSleep("not-a-duration"); err == nil {
		t.Error("expected parse error")
	}
	if _, err := parseSleep("-1"); err == nil {
		t.Error("expected error for negative duration")
	}
}

func appendStr(t *testing.T, p, s string) {
	t.Helper()
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(s); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()
}
