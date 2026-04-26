// Package unexpand implements the `unexpand` applet: convert leading
// (or all) blank runs to tabs.
package unexpand

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
		Name:  "unexpand",
		Help:  "convert spaces to tabs",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: unexpand [OPTION]... [FILE]...
Convert blanks in each FILE to tabs, writing to standard output.

By default unexpand only converts leading runs of blanks. Use -a to
convert blanks anywhere they would compress.

Options:
  -a, --all             convert all blanks, not just leading
  --first-only          convert only leading blanks (default)
  -t, --tabs=N          tab stops every N columns (default 8)
  -t, --tabs=LIST       comma-separated explicit tab stops
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		allBlanks bool
		stops     = []int{8}
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
		case a == "-a", a == "--all":
			allBlanks = true
			args = args[1:]
		case a == "--first-only":
			allBlanks = false
			args = args[1:]
		case a == "-t":
			if len(args) < 2 {
				ioutil.Errf("unexpand: option requires an argument -- 't'")
				return 2
			}
			s, err := parseTabs(args[1])
			if err != nil {
				ioutil.Errf("unexpand: %v", err)
				return 2
			}
			stops = s
			// -t implies -a per GNU unexpand.
			allBlanks = true
			args = args[2:]
		case strings.HasPrefix(a, "--tabs="):
			s, err := parseTabs(a[len("--tabs="):])
			if err != nil {
				ioutil.Errf("unexpand: %v", err)
				return 2
			}
			stops = s
			allBlanks = true
			args = args[1:]
		case strings.HasPrefix(a, "-t") && len(a) > 2:
			s, err := parseTabs(a[2:])
			if err != nil {
				ioutil.Errf("unexpand: %v", err)
				return 2
			}
			stops = s
			allBlanks = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("unexpand: unknown option: %s", a)
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
		if err := unexpandOne(name, w, stops, allBlanks); err != nil {
			ioutil.Errf("unexpand: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func unexpandOne(name string, w *bufio.Writer, stops []int, allBlanks bool) error {
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

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		emit(w, sc.Text(), stops, allBlanks)
		_ = w.WriteByte('\n')
	}
	return sc.Err()
}

// emit writes the unexpanded form of line to w.
//
// Strategy: walk the line column by column, accumulating runs of blanks
// (spaces or tabs already present) and replacing each maximal run with
// the smallest equivalent (tabs to the next stop, then any leftover
// spaces). When allBlanks is false, runs are only compressed if the
// run starts at the beginning of the line.
func emit(w *bufio.Writer, line string, stops []int, allBlanks bool) {
	col := 0
	i := 0
	leading := true
	for i < len(line) {
		c := line[i]
		if c == ' ' || c == '\t' {
			if !leading && !allBlanks {
				// Preserve the run as-is.
				j := i
				for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
					col = advanceCol(col, line[j], stops)
					j++
				}
				_, _ = w.WriteString(line[i:j])
				i = j
				continue
			}
			// Compress the run.
			start := col
			j := i
			for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
				col = advanceCol(col, line[j], stops)
				j++
			}
			emitRun(w, start, col, stops)
			i = j
			continue
		}
		_ = w.WriteByte(c)
		col++
		leading = false
		i++
	}
}

// advanceCol returns the new column after consuming one whitespace byte.
func advanceCol(col int, c byte, stops []int) int {
	if c == '\t' {
		return nextStop(col, stops)
	}
	return col + 1
}

// emitRun writes the smallest tab+space sequence that advances from
// start column to end column.
func emitRun(w *bufio.Writer, start, end int, stops []int) {
	cur := start
	for cur < end {
		ns := nextStop(cur, stops)
		if ns <= end {
			_ = w.WriteByte('\t')
			cur = ns
		} else {
			_ = w.WriteByte(' ')
			cur++
		}
	}
}

func nextStop(pos int, stops []int) int {
	if len(stops) == 1 {
		stride := stops[0]
		return ((pos / stride) + 1) * stride
	}
	for _, s := range stops {
		if s > pos {
			return s
		}
	}
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
