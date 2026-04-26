package sed

import "testing"

// FuzzParseScript stresses the sed script parser: the splitter that
// recognizes ';' / '\n' separators while skipping inside regex
// addresses and s-command bodies, plus the per-command parser for
// addresses, ranges, negation, and command bodies. The parser is
// allowed to return errors but must never panic.
func FuzzParseScript(f *testing.F) {
	seeds := []string{
		"",
		"s/a/b/",
		"s|a|b|g",
		"/foo/d",
		"$d",
		"1,3d",
		"2q",
		"-n s/x/y/p",
		"/^#/!p",
		"/start/,/end/d",
		"s/a/b/;s/c/d/",
		"s/a/b/\ns/c/d/",
		"s/;/X/",
		"s|/|\\\\|g",
		"s",
		"s/",
		"s//",
		"s/a/",
		",d",
		"0d",
		"1,",
		",1d",
		"!d",
		"s/[/x/", // unterminated character class
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		if len(s) > 4096 {
			return
		}
		_, _ = parseScript(s)
	})
}
