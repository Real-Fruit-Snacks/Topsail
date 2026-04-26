// Package whoami implements the `whoami` applet.
package whoami

import (
	"os/user"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "whoami",
		Help:  "print the current effective username",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: whoami
Print the current effective username.
`

// Main is the applet entry point.
func Main(argv []string) int {
	if len(argv) > 1 && argv[1] != "--" {
		ioutil.Errf("whoami: extra operand: %s", argv[1])
		return 2
	}
	u, err := user.Current()
	if err != nil {
		ioutil.Errf("whoami: %v", err)
		return 1
	}
	_, _ = ioutil.Stdout.Write([]byte(u.Username + "\n"))
	return 0
}
