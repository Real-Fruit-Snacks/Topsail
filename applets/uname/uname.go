// Package uname implements the `uname` applet: print system information.
package uname

import (
	"os"
	"runtime"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "uname",
		Help:  "print system information",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: uname [OPTION]...
Print system information.

Options:
  -a, --all          print all available info
  -s, --kernel-name  kernel name (default; e.g. Linux, Darwin, Windows_NT)
  -n, --nodename     hostname
  -m, --machine      machine architecture (amd64, arm64, ...)
  -o, --operating-system   pretty OS name
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var all, kernel, node, machine, opsys bool

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
		case a == "-s", a == "--kernel-name":
			kernel = true
			args = args[1:]
		case a == "-n", a == "--nodename":
			node = true
			args = args[1:]
		case a == "-m", a == "--machine":
			machine = true
			args = args[1:]
		case a == "-o", a == "--operating-system":
			opsys = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'a':
					all = true
				case 's':
					kernel = true
				case 'n':
					node = true
				case 'm':
					machine = true
				case 'o':
					opsys = true
				default:
					ioutil.Errf("uname: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			ioutil.Errf("uname: extra operand: %s", a)
			return 2
		}
	}

	if !kernel && !node && !machine && !opsys && !all {
		kernel = true
	}

	var parts []string
	kernelName := kernelOf(runtime.GOOS)
	host, _ := os.Hostname()

	if kernel || all {
		parts = append(parts, kernelName)
	}
	if node || all {
		parts = append(parts, host)
	}
	if machine || all {
		parts = append(parts, runtime.GOARCH)
	}
	if opsys || all {
		parts = append(parts, runtime.GOOS)
	}
	_, _ = ioutil.Stdout.Write([]byte(strings.Join(parts, " ") + "\n"))
	return 0
}

func kernelOf(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows_NT"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	case "netbsd":
		return "NetBSD"
	}
	return goos
}
