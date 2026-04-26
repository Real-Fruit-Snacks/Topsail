// Command topsail is the multi-call BusyBox-like binary entry point.
//
// Applets self-register in their package init() functions. To add an applet,
// create applets/<name>/, give it an init() that calls applet.Register, and
// add a blank import to the import list below.
package main

import (
	"os"

	"github.com/Real-Fruit-Snacks/topsail/internal/cli"
	// Applet imports — one line per applet package, alphabetized.
	// Wave 0 ships with no applets registered; Wave 1 begins populating.
)

func main() {
	os.Exit(cli.Run(os.Args))
}
