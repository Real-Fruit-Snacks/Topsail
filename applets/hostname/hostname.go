// Package hostname implements the `hostname` applet.
package hostname

import (
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "hostname",
		Help:  "print the system hostname",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: hostname [OPTION]
Print the system hostname.

Options:
  -s, --short   strip the domain part (just the hostname)
  -f, --fqdn    print the fully-qualified domain name (best-effort)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var short, fqdn bool
	for _, a := range args {
		switch a {
		case "-s", "--short":
			short = true
		case "-f", "--fqdn", "--long":
			fqdn = true
		case "--":
			// no-op
		default:
			ioutil.Errf("hostname: unknown option: %s", a)
			return 2
		}
	}

	name, err := os.Hostname()
	if err != nil {
		ioutil.Errf("hostname: %v", err)
		return 1
	}
	switch {
	case short:
		if i := strings.Index(name, "."); i >= 0 {
			name = name[:i]
		}
	case fqdn:
		// os.Hostname is best-effort already; net.LookupCNAME could
		// extend this but we keep it simple in Wave 6.
	}
	_, _ = ioutil.Stdout.Write([]byte(name + "\n"))
	return 0
}
