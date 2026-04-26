// Package awk implements the `awk` applet by embedding benhoyt/goawk.
package awk

import (
	"strings"

	"github.com/benhoyt/goawk/interp"
	"github.com/benhoyt/goawk/parser"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:    "awk",
		Aliases: []string{"gawk", "nawk"},
		Help:    "pattern scanning and processing language",
		Usage:   usage,
		Main:    Main,
	})
}

const usage = `Usage: awk [OPTION]... 'PROGRAM' [FILE]...
       awk [OPTION]... -f PROGFILE [-f PROGFILE]... [FILE]...
Run an AWK PROGRAM against FILE(s) or stdin.

This applet wraps benhoyt/goawk; see https://github.com/benhoyt/goawk
for the supported language. Common flags are forwarded:

Options:
  -F FS              field separator
  -v VAR=VAL         set a variable before the program runs
  -f PROGFILE        read program from file
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		fs        string
		vars      []string
		progFiles []string
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-F":
			if len(args) < 2 {
				ioutil.Errf("awk: option requires an argument -- 'F'")
				return 2
			}
			fs = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "-F"):
			fs = a[2:]
			args = args[1:]
		case a == "-v":
			if len(args) < 2 {
				ioutil.Errf("awk: option requires an argument -- 'v'")
				return 2
			}
			vars = append(vars, args[1])
			args = args[2:]
		case a == "-f":
			if len(args) < 2 {
				ioutil.Errf("awk: option requires an argument -- 'f'")
				return 2
			}
			progFiles = append(progFiles, args[1])
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("awk: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	var progSrc string
	var files []string
	if len(progFiles) > 0 {
		// Concatenate -f program sources.
		var sb strings.Builder
		for _, pf := range progFiles {
			data, err := readFile(pf)
			if err != nil {
				ioutil.Errf("awk: %s: %v", pf, err)
				return 2
			}
			sb.Write(data)
			sb.WriteByte('\n')
		}
		progSrc = sb.String()
		files = args
	} else {
		if len(args) == 0 {
			ioutil.Errf("awk: missing program")
			return 2
		}
		progSrc = args[0]
		files = args[1:]
	}

	prog, err := parser.ParseProgram([]byte(progSrc), nil)
	if err != nil {
		ioutil.Errf("awk: %v", err)
		return 2
	}

	cfg := &interp.Config{
		Stdin:        ioutil.Stdin,
		Output:       ioutil.Stdout,
		Error:        ioutil.Stderr,
		Args:         files,
		Vars:         flattenVars(vars, fs),
		NoExec:       true, // disable system() / pipes — sandboxing
		NoFileWrites: false,
		NoFileReads:  false,
	}

	status, err := interp.ExecProgram(prog, cfg)
	if err != nil {
		ioutil.Errf("awk: %v", err)
		return 2
	}
	return status
}

func flattenVars(vars []string, fs string) []string {
	out := make([]string, 0, len(vars)*2+2)
	if fs != "" {
		out = append(out, "FS", fs)
	}
	for _, kv := range vars {
		i := strings.Index(kv, "=")
		if i < 0 {
			continue
		}
		out = append(out, kv[:i], kv[i+1:])
	}
	return out
}

// readFile is split out so the //nolint:gosec annotation stays narrow.
func readFile(name string) ([]byte, error) {
	return readFileHelper(name)
}
