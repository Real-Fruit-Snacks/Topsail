package tr

import "testing"

// FuzzExpandSet exercises the SET parser used by tr's set arguments,
// including character ranges (a-z), POSIX classes ([:alpha:]), and
// backslash escapes. Past tr clones have crashed on truncated escapes
// or inverted ranges; the parser is required to fail gracefully.
func FuzzExpandSet(f *testing.F) {
	seeds := []string{
		"",
		"abc",
		"a-z",
		"A-Z",
		"0-9",
		"[:alpha:]",
		"[:digit:]",
		"[:upper:]",
		"[:lower:]",
		"[:space:]",
		"\\n",
		"\\t",
		"\\\\",
		"\\077",
		"a-",
		"-z",
		"-",
		"\\",
		"[:",
		"[:bogus:]",
		"[:alpha",
		"a-zA-Z0-9",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		if len(s) > 4096 {
			return
		}
		_, _ = expandSet(s)
	})
}
