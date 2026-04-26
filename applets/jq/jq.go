// Package jq implements the `jq` applet by embedding itchyny/gojq.
package jq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itchyny/gojq"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "jq",
		Help:  "filter and transform JSON",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: jq [OPTION]... FILTER [FILE]...
Apply gojq FILTER to each input JSON value and print the result.

Options:
  -r, --raw-output   output strings without quoting
  -c, --compact      compact (single-line) output
  -n, --null-input   ignore stdin; use null as the input value
  -s, --slurp        read all inputs into a single array

This applet wraps github.com/itchyny/gojq; see that project's README
for the supported jq language and built-in functions.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var raw, compact, nullInput, slurp bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-r", a == "--raw-output":
			raw = true
			args = args[1:]
		case a == "-c", a == "--compact":
			compact = true
			args = args[1:]
		case a == "-n", a == "--null-input":
			nullInput = true
			args = args[1:]
		case a == "-s", a == "--slurp":
			slurp = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("jq: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("jq: missing filter")
		return 2
	}
	filterSrc := args[0]
	files := args[1:]

	q, err := gojq.Parse(filterSrc)
	if err != nil {
		ioutil.Errf("jq: parse error: %v", err)
		return 3
	}
	code, err := gojq.Compile(q)
	if err != nil {
		ioutil.Errf("jq: compile error: %v", err)
		return 3
	}

	var inputs []any
	switch {
	case nullInput:
		inputs = []any{nil}
	case len(files) == 0:
		ins, err := readInputs(ioutil.Stdin, slurp)
		if err != nil {
			ioutil.Errf("jq: %v", err)
			return 2
		}
		inputs = ins
	default:
		for _, name := range files {
			f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
			if err != nil {
				ioutil.Errf("jq: %s: %v", name, err)
				return 2
			}
			ins, err := readInputs(f, slurp)
			_ = f.Close()
			if err != nil {
				ioutil.Errf("jq: %s: %v", name, err)
				return 2
			}
			inputs = append(inputs, ins...)
		}
	}

	rc := 0
	for _, in := range inputs {
		iter := code.Run(in)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				ioutil.Errf("jq: %v", err)
				rc = 5
				continue
			}
			if err := emit(v, raw, compact); err != nil {
				ioutil.Errf("jq: %v", err)
				rc = 2
			}
		}
	}
	return rc
}

func readInputs(r io.Reader, slurp bool) ([]any, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var out []any
	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		out = append(out, v)
	}
	if slurp {
		out = []any{any(out)}
	}
	return out, nil
}

func emit(v any, raw, compact bool) error {
	if raw {
		if s, ok := v.(string); ok {
			_, err := fmt.Fprintln(ioutil.Stdout, s)
			return err
		}
	}
	enc := json.NewEncoder(ioutil.Stdout)
	if !compact {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}
