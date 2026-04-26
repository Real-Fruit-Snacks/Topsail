// Package truecmd implements the `true` applet: do nothing, exit 0.
//
// The package is named truecmd rather than true to avoid shadowing
// Go's predeclared boolean identifier within the package scope.
package truecmd

import "github.com/Real-Fruit-Snacks/topsail/internal/applet"

func init() {
	applet.Register(applet.Applet{
		Name:  "true",
		Help:  "exit with success status",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: true
Exit with status 0 (success).

The classic POSIX placeholder for "do nothing, succeed."
All arguments are silently ignored.
`

// Main is the applet entry point. It always returns 0.
func Main(argv []string) int {
	_ = argv
	return 0
}
