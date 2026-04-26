package ping

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestPingLocalListener(t *testing.T) {
	// Spin up a TCP listener so the probe always succeeds.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port

	// Accept any connections so they don't pile up.
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"ping", "-c", "1", "-p", strconv.Itoa(port), "-i", "0", "127.0.0.1"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got := out.String()
	if !strings.Contains(got, "Reply from") {
		t.Errorf("got %q", got)
	}
}

func TestPingClosedPort(t *testing.T) {
	// Pick a port unlikely to be open.
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"ping", "-c", "1", "-p", "1", "-W", "1", "-i", "0", "127.0.0.1"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if !strings.Contains(out.String(), "Request") {
		t.Errorf("got %q", out.String())
	}
}

func TestPingMissing(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ping"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "missing host") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestPingInvalidFlag(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ping", "-Z", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestPingInvalidCount(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"ping", "-c", "abc", "x"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
