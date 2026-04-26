// Package testutil provides shared test helpers for topsail.
//
// The most common need across applet tests is swapping ioutil.Stdin /
// Stdout / Stderr for in-memory buffers so a Main() call can be
// asserted on without subprocess plumbing.
package testutil

import (
	"bytes"
	"io"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

// SwapStdio replaces ioutil.Stdin / Stdout / Stderr for the duration of
// the test. Pass nil for any stream you do not want to mock; the
// originals are restored via t.Cleanup.
func SwapStdio(t *testing.T, in io.Reader, out, errBuf io.Writer) {
	t.Helper()
	origIn, origOut, origErr := ioutil.Stdin, ioutil.Stdout, ioutil.Stderr
	if in != nil {
		ioutil.Stdin = in
	}
	if out != nil {
		ioutil.Stdout = out
	}
	if errBuf != nil {
		ioutil.Stderr = errBuf
	}
	t.Cleanup(func() {
		ioutil.Stdin = origIn
		ioutil.Stdout = origOut
		ioutil.Stderr = origErr
	})
}

// CaptureStdout swaps Stdout for a fresh *bytes.Buffer and returns it.
func CaptureStdout(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	SwapStdio(t, nil, buf, nil)
	return buf
}

// CaptureStderr swaps Stderr for a fresh *bytes.Buffer and returns it.
func CaptureStderr(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	SwapStdio(t, nil, nil, buf)
	return buf
}

// CaptureStdio swaps Stdout AND Stderr for fresh buffers, returning
// both. Convenient when an applet writes to both and the test needs
// to assert on each independently.
func CaptureStdio(t *testing.T) (out, errBuf *bytes.Buffer) {
	t.Helper()
	out = &bytes.Buffer{}
	errBuf = &bytes.Buffer{}
	SwapStdio(t, nil, out, errBuf)
	return out, errBuf
}

// SetStdin replaces Stdin with a reader serving s for the test's lifetime.
func SetStdin(t *testing.T, s string) {
	t.Helper()
	SwapStdio(t, bytes.NewBufferString(s), nil, nil)
}
