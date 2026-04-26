// Package cli implements the multi-call dispatcher: argv[0] basename
// resolution, the topsail wrapper command, and global flags (--help,
// --version, --list).
package cli

import (
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

// Process exit codes mirroring POSIX conventions.
const (
	ExitSuccess        = 0
	ExitRuntimeErr     = 1
	ExitUsageErr       = 2
	ExitAppletNotFound = 127
)

// wrapperName is the canonical name the binary uses in wrapper mode.
const wrapperName = "topsail"

// Run is the process entry point. It receives the full os.Args slice and
// returns the exit code for the process. main() should pass the result to
// os.Exit directly.
func Run(args []string) int {
	if len(args) == 0 {
		ioutil.Errf("topsail: empty argv")
		return ExitUsageErr
	}

	invocation := basename(args[0])

	// Multi-call dispatch: argv[0] basename matches a registered applet.
	// The wrapper name itself never matches an applet so we always fall
	// through to runWrapper for `topsail ...`.
	if invocation != wrapperName {
		if a, ok := applet.Get(invocation); ok {
			return runApplet(a, args)
		}
	}

	return runWrapper(args)
}

// runApplet invokes the applet directly. In multi-call mode we honor only
// --help (long form). Short -h stays available so applets like `df -h`
// (human-readable) keep working.
func runApplet(a applet.Applet, args []string) int {
	for _, arg := range args[1:] {
		if arg == "--" {
			break
		}
		if arg == "--help" {
			printAppletHelp(ioutil.Stdout, a)
			return ExitSuccess
		}
	}
	return a.Main(args)
}

// runWrapper handles `topsail ...`:
//
//	topsail                  -> top-level help
//	topsail --help           -> top-level help
//	topsail --help <applet>  -> per-applet help
//	topsail --version        -> version banner
//	topsail --list           -> applet list
//	topsail <applet> [args]  -> dispatch with argv[0]=<applet>
func runWrapper(args []string) int {
	if len(args) < 2 {
		printTopHelp(ioutil.Stdout)
		return ExitSuccess
	}

	switch args[1] {
	case "-h", "--help":
		if len(args) >= 3 {
			target := args[2]
			a, ok := applet.Get(target)
			if !ok {
				ioutil.Errf("topsail: unknown applet: %s", target)
				return ExitAppletNotFound
			}
			printAppletHelp(ioutil.Stdout, a)
			return ExitSuccess
		}
		printTopHelp(ioutil.Stdout)
		return ExitSuccess

	case "-V", "--version":
		printVersion(ioutil.Stdout)
		return ExitSuccess

	case "--list":
		printList(ioutil.Stdout, applet.All())
		return ExitSuccess
	}

	if strings.HasPrefix(args[1], "-") {
		ioutil.Errf("topsail: unknown option: %s", args[1])
		ioutil.Errf("Try 'topsail --help' for usage.")
		return ExitUsageErr
	}

	name := args[1]
	a, ok := applet.Get(name)
	if !ok {
		ioutil.Errf("topsail: unknown applet: %s", name)
		ioutil.Errf("Try 'topsail --list' to see all applets.")
		return ExitAppletNotFound
	}

	// Reform argv with the applet name as argv[0] so the applet sees the
	// same shape as a symlink invocation.
	sub := append([]string{name}, args[2:]...)
	return runApplet(a, sub)
}

// basename returns the base name of a path with any trailing .exe stripped
// (case-insensitive) so Windows symlink/copy invocations match the registry.
func basename(p string) string {
	b := filepath.Base(p)
	if len(b) > 4 {
		tail := strings.ToLower(b[len(b)-4:])
		if tail == ".exe" {
			b = b[:len(b)-4]
		}
	}
	return b
}
