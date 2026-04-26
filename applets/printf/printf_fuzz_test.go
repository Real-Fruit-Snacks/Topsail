package printf

import (
	"bytes"
	"testing"
)

// FuzzEmit feeds arbitrary format strings into the directive parser and
// asserts it never panics. emit handles its own conversion errors and
// reports them via the errs counter; the fuzz target only catches
// outright runtime failures (slice bounds, nil derefs, etc.).
func FuzzEmit(f *testing.F) {
	seeds := []string{
		"",
		"hello\n",
		"%s",
		"%d",
		"%-10s|",
		"%05.2f",
		"%%",
		"%",
		"%q",
		"%x",
		"%c",
		"\\n\\t",
		"\\x41",
		"\\0123",
		"%-",
		"%5",
		"%5.",
		"%.5",
		"%5.5",
		"%9999s",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, format string) {
		var buf bytes.Buffer
		// Pass a generous arg pool so directives don't stall on empties.
		_, _, _ = emit(&buf, format, []string{"a", "b", "c", "1", "-1", "1.5"})
	})
}

// FuzzDecodeEscape exercises the backslash-escape decoder. Inputs may
// be truncated or malformed; decodeEscape must always return without
// panicking.
func FuzzDecodeEscape(f *testing.F) {
	seeds := []string{"n", "t", "x4", "x41", "0", "01", "012", "0123", "z", "\\", ""}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = decodeEscape(s)
	})
}
