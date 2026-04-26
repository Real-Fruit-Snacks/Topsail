// Package falsecmd implements the `false` applet: do nothing, exit 1.
//
// The package is named falsecmd rather than false to avoid shadowing
// Go's predeclared boolean identifier within the package scope.
package falsecmd

import "github.com/Real-Fruit-Snacks/topsail/internal/applet"

func init() {
	applet.Register(applet.Applet{
		Name:  "false",
		Help:  "exit with failure status",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: false
Exit with status 1 (failure).

The classic POSIX placeholder for "do nothing, fail."
All arguments are silently ignored.
`

// Main is the applet entry point. It always returns 1.
func Main(argv []string) int {
	_ = argv
	return 1
}
