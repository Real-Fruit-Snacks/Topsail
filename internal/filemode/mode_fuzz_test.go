package filemode

import "testing"

// FuzzParse exercises the octal/symbolic mode parser. The parser is
// expected to return errors for malformed input but must never panic.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"0", "7", "777", "0644", "4755",
		"u+x", "go-w", "a=rwx",
		"u=rwx,go=rx",
		"+x", "+s", "+t",
		"u+s", "g+s", "o+t",
		"a+X",
		"g=u",
		"",
		",",
		"u",
		"u@x",
		"u+z",
		"99",
		"0xff",
		"u+x,",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		if len(s) > 256 {
			return
		}
		_, _ = Parse(s, 0o644)
	})
}
