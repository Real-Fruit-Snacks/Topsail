package expr

import (
	"strings"
	"testing"
)

// FuzzExprParse runs random space-separated token streams through the
// recursive-descent parser. The parser is allowed to return an error,
// but it must never panic — past bugs in similar parsers have
// included unbounded recursion, slice-out-of-range during precedence
// climbing, and integer overflow in arithmetic ops.
func FuzzExprParse(f *testing.F) {
	seeds := []string{
		"1",
		"1 + 2",
		"3 \\* 4",
		"a + b",
		"( 1 + 2 ) \\* 3",
		"1 = 1",
		"foo : foo",
		"length string",
		"substr abcdef 2 3",
		"index abcdef cd",
		"match abc abc",
		"1 + ( 2 \\* ( 3 - 4 / 5 ) )",
		"",
		"+ + +",
		"((((",
		")))))",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		// Cap input to keep the corpus small and to avoid pathological
		// memory growth from synthesized mega-strings.
		if len(s) > 4096 {
			return
		}
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			return
		}
		p := &parser{tokens: tokens}
		_, _ = p.expr()
	})
}
