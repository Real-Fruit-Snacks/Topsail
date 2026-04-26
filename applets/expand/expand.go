// Package expand implements the `expand` applet: convert tabs in
// each FILE to the proper number of spaces.
package expand

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "expand",
		Help:  "convert tabs to spaces",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: expand [OPTION]... [FILE]...
Convert tabs in each FILE to the proper number of spaces.

With no FILE, or when FILE is -, read standard input.

Options:
  -i, --initial         only convert leading tabs on each line
  -t, --tabs=N          tab stops every N columns (default 8)
  -t, --tabs=LIST       comma-separated explicit tab stops
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		initialOnly bool
		stops       = []int{8}
	)

	parseTabs := func(s string) ([]int, error) {
		if s == "" {
			return nil, fmt.Errorf("empty -t value")
		}
		parts := strings.Split(s, ",")
		out := make([]int, 0, len(parts))
		for _, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || n < 1 {
				return nil, fmt.Errorf("invalid tab stop %q", p)
			}
			out = append(out, n)
		}
		return out, nil
	}

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-i", a == "--initial":
			initialOnly = true
			args = args[1:]
		case a == "-t":
			if len(args) < 2 {
				ioutil.Errf("expand: option requires an argument -- 't'")
				return 2
			}
			s, err := parseTabs(args[1])
			if err != nil {
				ioutil.Errf("expand: %v", err)
				return 2
			}
			stops = s
			args = args[2:]
		case strings.HasPrefix(a, "--tabs="):
			s, err := parseTabs(a[len("--tabs="):])
			if err != nil {
				ioutil.Errf("expand: %v", err)
				return 2
			}
			stops = s
			args = args[1:]
		case strings.HasPrefix(a, "-t") && len(a) > 2:
			s, err := parseTabs(a[2:])
			if err != nil {
				ioutil.Errf("expand: %v", err)
				return 2
			}
			stops = s
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("expand: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	w := bufio.NewWriter(ioutil.Stdout)
	defer func() { _ = w.Flush() }()

	rc := 0
	for _, name := range files {
		if err := expandOne(name, w, stops, initialOnly); err != nil {
			ioutil.Errf("expand: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func expandOne(name string, w *bufio.Writer, stops []int, initialOnly bool) error {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	br := bufio.NewReader(r)
	col := 0
	leading := true
	for {
		c, _, err := br.ReadRune()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if c == '\n' {
			_, _ = w.WriteRune(c)
			col = 0
			leading = true
			continue
		}
		if c == '\t' && (!initialOnly || leading) {
			next := nextStop(col, stops)
			for col < next {
				_ = w.WriteByte(' ')
				col++
			}
			continue
		}
		if c != ' ' && c != '\t' {
			leading = false
		}
		_, _ = w.WriteRune(c)
		col++
	}
}

// nextStop returns the next column after pos given the configured tab stops.
// stops is either a single stride (one element: every N) or an explicit list.
func nextStop(pos int, stops []int) int {
	if len(stops) == 1 {
		stride := stops[0]
		// Round up to the next multiple of stride.
		return ((pos / stride) + 1) * stride
	}
	for _, s := range stops {
		if s > pos {
			return s
		}
	}
	// Past the last explicit stop: advance by the last delta.
	last := stops[len(stops)-1]
	if len(stops) >= 2 {
		stride := stops[len(stops)-1] - stops[len(stops)-2]
		if stride < 1 {
			stride = 1
		}
		return last + stride*((pos-last)/stride+1)
	}
	return last + 1
}
