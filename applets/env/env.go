// Package env implements the `env` applet: print or modify the environment.
package env

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "env",
		Help:  "print or modify the environment",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: env [OPTION]... [NAME=VALUE]... [COMMAND [ARG]...]
Print the environment, or set NAME=VALUE pairs and run COMMAND.

Options:
  -i, --ignore-environment   start with an empty environment
  -u NAME, --unset=NAME      remove NAME from the environment
  -0                         end output lines with NUL, not newline
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var clearEnv, nul bool
	var unsets []string

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-i", a == "--ignore-environment":
			clearEnv = true
			args = args[1:]
		case a == "-u":
			if len(args) < 2 {
				ioutil.Errf("env: option requires an argument -- 'u'")
				return 2
			}
			unsets = append(unsets, args[1])
			args = args[2:]
		case strings.HasPrefix(a, "--unset="):
			unsets = append(unsets, a[len("--unset="):])
			args = args[1:]
		case a == "-0":
			nul = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-" && !strings.Contains(a, "="):
			ioutil.Errf("env: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	envMap := map[string]string{}
	if !clearEnv {
		for _, kv := range os.Environ() {
			if i := strings.Index(kv, "="); i >= 0 {
				envMap[kv[:i]] = kv[i+1:]
			}
		}
	}
	for _, u := range unsets {
		delete(envMap, u)
	}

	// Apply NAME=VALUE prefix; first arg without '=' marks the command.
	cmdStart := 0
	for i, a := range args {
		if !strings.Contains(a, "=") {
			cmdStart = i
			break
		}
		eq := strings.Index(a, "=")
		envMap[a[:eq]] = a[eq+1:]
		cmdStart = i + 1
	}

	if cmdStart >= len(args) {
		// Print environment
		keys := make([]string, 0, len(envMap))
		for k := range envMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		term := byte('\n')
		if nul {
			term = 0
		}
		for _, k := range keys {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s=%s%c", k, envMap[k], term)
		}
		return 0
	}

	// Run command
	cmdArgs := args[cmdStart:]
	envSlice := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envSlice = append(envSlice, k+"="+v)
	}
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec // env runs user-named commands by design
	c.Env = envSlice
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
		ioutil.Errf("env: %v", err)
		return 1
	}
	return 0
}
