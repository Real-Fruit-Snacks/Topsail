// Package which implements the `which` applet: locate a command in PATH.
package which

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "which",
		Help:  "locate a command in PATH",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: which [-a] COMMAND...
Print the full path of each COMMAND that would be executed if invoked.

Options:
  -a    print all matches in PATH, not just the first
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var all bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-a", a == "--all":
			all = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("which: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("which: missing operand")
		return 2
	}
	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))
	exts := []string{""}
	if runtime.GOOS == "windows" {
		extsEnv := os.Getenv("PATHEXT")
		if extsEnv == "" {
			extsEnv = ".COM;.EXE;.BAT;.CMD"
		}
		exts = append([]string{""}, strings.Split(extsEnv, ";")...)
	}

	rc := 0
	for _, cmd := range args {
		found := false
		for _, dir := range dirs {
			if dir == "" {
				continue
			}
			for _, ext := range exts {
				p := filepath.Join(dir, cmd+ext)
				info, err := os.Stat(p) //nolint:gosec // walking $PATH for a named command is the whole point
				if err != nil || info.IsDir() {
					continue
				}
				if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
					continue
				}
				_, _ = fmt.Fprintln(ioutil.Stdout, p)
				found = true
				if !all {
					break
				}
			}
			if found && !all {
				break
			}
		}
		if !found {
			rc = 1
		}
	}
	return rc
}
