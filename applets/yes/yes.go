// Package yes implements the `yes` applet: print a string repeatedly.
package yes

import (
	"io"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "yes",
		Help:  "output a string repeatedly until killed",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: yes [STRING]...
Repeatedly output STRING(s), one per line, until killed.

If no STRING is given, output "y". Multiple STRINGs are joined with a
single space. Exits 0 on broken pipe (the conventional "consumer
hung up" termination).
`

// Main is the applet entry point.
func Main(argv []string) int {
	var s string
	if len(argv) <= 1 {
		s = "y"
	} else {
		s = strings.Join(argv[1:], " ")
	}
	line := s + "\n"

	for {
		if _, err := io.WriteString(ioutil.Stdout, line); err != nil {
			// Broken pipe (or any write error) is the standard
			// "consumer is gone" signal — exit cleanly.
			return 0
		}
	}
}
