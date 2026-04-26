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
	// Each blank import runs that package's init() to self-register.
	_ "github.com/Real-Fruit-Snacks/topsail/applets/basename"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/cat"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/cp"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/cut"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/dirname"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/echo"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/expr"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/false"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/head"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/mkdir"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/mv"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/printf"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/pwd"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/rev"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/rm"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/rmdir"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/seq"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/sleep"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/sort"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/tac"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/tail"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/tee"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/test"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/touch"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/tr"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/true"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/uniq"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/wc"
	_ "github.com/Real-Fruit-Snacks/topsail/applets/yes"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
