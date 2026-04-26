package testutil

import (
	"io"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func TestSwapStdioRestores(t *testing.T) {
	origIn, origOut, origErr := ioutil.Stdin, ioutil.Stdout, ioutil.Stderr
	t.Run("inner", func(t *testing.T) {
		out := CaptureStdout(t)
		_, _ = io.WriteString(ioutil.Stdout, "hello")
		if got := out.String(); got != "hello" {
			t.Errorf("Stdout buffer = %q; want %q", got, "hello")
		}
	})
	if ioutil.Stdin != origIn || ioutil.Stdout != origOut || ioutil.Stderr != origErr {
		t.Error("ioutil globals not restored after subtest")
	}
}

func TestCaptureStderr(t *testing.T) {
	buf := CaptureStderr(t)
	ioutil.Errf("oops")
	if got, want := buf.String(), "oops\n"; got != want {
		t.Errorf("CaptureStderr = %q; want %q", got, want)
	}
}

func TestCaptureStdio(t *testing.T) {
	out, errBuf := CaptureStdio(t)
	_, _ = io.WriteString(ioutil.Stdout, "out")
	ioutil.Errf("err")
	if out.String() != "out" {
		t.Errorf("stdout = %q", out.String())
	}
	if errBuf.String() != "err\n" {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestSetStdin(t *testing.T) {
	SetStdin(t, "hello world")
	got, err := io.ReadAll(ioutil.Stdin)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("Stdin = %q; want hello world", string(got))
	}
}

func TestSwapStdioNilLeavesUnchanged(t *testing.T) {
	origIn := ioutil.Stdin
	SwapStdio(t, nil, nil, nil)
	if ioutil.Stdin != origIn {
		t.Error("nil reader should not replace Stdin")
	}
}
