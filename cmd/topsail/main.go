// Command topsail is the multi-call BusyBox-like binary entry point.
//
// Applets self-register in their package init() functions. To add an applet,
// create applets/<name>/, give it an init() that calls applet.Register, and
// add a blank import line to internal/applets/all.go.
package main

import (
	"os"

	_ "github.com/Real-Fruit-Snacks/topsail/internal/applets"
	"github.com/Real-Fruit-Snacks/topsail/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
