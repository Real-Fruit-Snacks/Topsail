// Package tsort implements the `tsort` applet: topological sort.
package tsort

import (
	"bufio"
	"io"
	"os"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "tsort",
		Help:  "topological sort of a directed graph",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tsort [FILE]
Read pairs of strings from FILE (or stdin) representing edges A -> B,
and print a linear ordering compatible with all edges.

A cycle in the input is reported on stderr; partial output may still
be produced (matching POSIX tsort behavior). Exit 1 if a cycle exists.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) > 1 {
		ioutil.Errf("tsort: extra operand: %s", args[1])
		return 2
	}
	var r io.Reader
	r = ioutil.Stdin
	if len(args) == 1 {
		name := args[0] //nolint:gosec // length-checked one line above
		if name != "-" {
			f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
			if err != nil {
				ioutil.Errf("tsort: %s: %v", name, err)
				return 1
			}
			defer func() { _ = f.Close() }()
			r = f
		}
	}

	type node struct {
		name  string
		inDeg int
		out   []string
	}
	nodes := map[string]*node{}
	add := func(s string) *node {
		n, ok := nodes[s]
		if !ok {
			n = &node{name: s}
			nodes[s] = n
		}
		return n
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var tokens []string
	for sc.Scan() {
		tokens = append(tokens, whitespaceFields(sc.Text())...)
	}
	if len(tokens)%2 != 0 {
		ioutil.Errf("tsort: input contains an odd number of tokens")
		return 1
	}
	for i := 0; i < len(tokens); i += 2 {
		a, b := tokens[i], tokens[i+1]
		na, nb := add(a), add(b)
		if a != b {
			na.out = append(na.out, b)
			nb.inDeg++
		}
	}

	// Kahn's algorithm.
	var queue []string
	for k, n := range nodes {
		if n.inDeg == 0 {
			queue = append(queue, k)
		}
	}
	var order []string
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		order = append(order, k)
		for _, dep := range nodes[k].out {
			nodes[dep].inDeg--
			if nodes[dep].inDeg == 0 {
				queue = append(queue, dep)
			}
		}
	}
	for _, n := range order {
		_, _ = ioutil.Stdout.Write([]byte(n + "\n"))
	}
	if len(order) != len(nodes) {
		ioutil.Errf("tsort: input contains a cycle")
		return 1
	}
	return 0
}

func whitespaceFields(s string) []string {
	var out []string
	var cur []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\r' {
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = cur[:0]
			}
			continue
		}
		cur = append(cur, c)
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}
