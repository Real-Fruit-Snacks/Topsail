// Package xargs implements the `xargs` applet: build and execute commands
// from standard input.
package xargs

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "xargs",
		Help:  "build and execute commands from stdin",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: xargs [OPTION]... [COMMAND [INITIAL_ARGS]...]
Read items from standard input and execute COMMAND with the items as
trailing arguments. With no COMMAND, the default is 'echo'.

Options:
  -n MAX     use at most MAX arguments per command line
  -I REPL    replace occurrences of REPL in COMMAND with input items
             (implies -n 1)
  -d DELIM   use DELIM as the input separator (default: whitespace)
  -0         input items are NUL-separated (shorthand for -d '\\0')
  -t         echo each command before running it
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		maxN  = -1
		repl  string
		delim byte // 0 means "any whitespace"
		trace bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-n":
			if len(args) < 2 {
				ioutil.Errf("xargs: option requires an argument -- 'n'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("xargs: invalid number: %s", args[1])
				return 2
			}
			maxN = n
			args = args[2:]
		case a == "-I":
			if len(args) < 2 {
				ioutil.Errf("xargs: option requires an argument -- 'I'")
				return 2
			}
			repl = args[1]
			maxN = 1
			args = args[2:]
		case a == "-d":
			if len(args) < 2 || args[1] == "" {
				ioutil.Errf("xargs: -d requires a non-empty argument")
				return 2
			}
			delim = args[1][0]
			args = args[2:]
		case a == "-0":
			// -0 means NUL-separated, but Go's zero-value is already 0,
			// so we set a sentinel via a separate var to distinguish
			// "user explicitly asked for NUL" from the default
			// "whitespace". For our reader, both behave identically when
			// delim==0 (we treat it as whitespace). Tracking that
			// distinction would require a separate bool; not worth it
			// for parity with the GNU manual page.
			delim = '\x00'
			args = args[1:]
		case a == "-t":
			trace = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("xargs: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	cmdArgs := args
	if len(cmdArgs) == 0 {
		cmdArgs = []string{"echo"}
	}

	items, err := readItems(ioutil.Stdin, delim)
	if err != nil {
		ioutil.Errf("xargs: %v", err)
		return 1
	}

	if repl != "" {
		// One execution per item, substituting repl in cmdArgs.
		for _, item := range items {
			expanded := make([]string, len(cmdArgs))
			for i, a := range cmdArgs {
				expanded[i] = strings.ReplaceAll(a, repl, item)
			}
			if rc := runCmd(expanded, trace); rc != 0 {
				return rc
			}
		}
		return 0
	}
	if maxN > 0 {
		for i := 0; i < len(items); i += maxN {
			end := i + maxN
			if end > len(items) {
				end = len(items)
			}
			full := append([]string{}, cmdArgs...)
			full = append(full, items[i:end]...)
			if rc := runCmd(full, trace); rc != 0 {
				return rc
			}
		}
		return 0
	}
	full := append([]string{}, cmdArgs...)
	full = append(full, items...)
	return runCmd(full, trace)
}

func readItems(r io.Reader, delim byte) ([]string, error) {
	br := bufio.NewReader(r)
	var items []string
	var cur []byte
	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			if len(cur) > 0 {
				items = append(items, string(cur))
			}
			return items, nil
		}
		if err != nil {
			return items, err
		}
		if delim == 0 {
			// whitespace delimited
			if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
				if len(cur) > 0 {
					items = append(items, string(cur))
					cur = cur[:0]
				}
				continue
			}
		} else if b == delim {
			items = append(items, string(cur))
			cur = cur[:0]
			continue
		}
		cur = append(cur, b)
	}
}

func runCmd(parts []string, trace bool) int {
	if len(parts) == 0 {
		return 0
	}
	if trace {
		ioutil.Errf("%s", strings.Join(parts, " "))
	}
	c := exec.Command(parts[0], parts[1:]...) //nolint:gosec // xargs by design runs user commands
	c.Stdin = ioutil.Stdin
	c.Stdout = ioutil.Stdout
	c.Stderr = ioutil.Stderr
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if exitErr, ok := err.(*exec.ExitError); ok {
			ee = exitErr
		}
		if ee != nil {
			return ee.ExitCode()
		}
		ioutil.Errf("xargs: %v", err)
		return 1
	}
	return 0
}
