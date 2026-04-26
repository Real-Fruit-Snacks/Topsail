package ioutil

import (
	"bytes"
	"testing"
)

func swapStderr(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	orig := Stderr
	Stderr = buf
	t.Cleanup(func() { Stderr = orig })
	return buf
}

func TestErrfAddsNewline(t *testing.T) {
	buf := swapStderr(t)
	Errf("hello %s", "world")
	if got, want := buf.String(), "hello world\n"; got != want {
		t.Errorf("Errf wrote %q; want %q", got, want)
	}
}

func TestErrfPreservesExistingNewline(t *testing.T) {
	buf := swapStderr(t)
	Errf("already terminated\n")
	if got, want := buf.String(), "already terminated\n"; got != want {
		t.Errorf("Errf wrote %q; want %q", got, want)
	}
}

func TestErrfEmptyFormat(t *testing.T) {
	buf := swapStderr(t)
	Errf("")
	if got, want := buf.String(), "\n"; got != want {
		t.Errorf(`Errf("") wrote %q; want %q`, got, want)
	}
}

func TestErrfNoArgs(t *testing.T) {
	buf := swapStderr(t)
	Errf("plain message")
	if got, want := buf.String(), "plain message\n"; got != want {
		t.Errorf("Errf wrote %q; want %q", got, want)
	}
}
