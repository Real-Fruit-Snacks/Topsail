// Package sha256sum implements the `sha256sum` applet.
package sha256sum

import (
	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/hashing"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sha256sum",
		Help:  "compute and check SHA-256 checksums",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sha256sum [OPTION]... [FILE]...
Print or check SHA-256 (256-bit) checksums.

Options:
  -c, --check       read sums from FILEs and verify them
  -b, --binary      read in binary mode (accepted; identical on Go)
  -t, --text        read in text mode (accepted; identical on Go)
      --quiet       with -c, suppress per-file 'OK' lines
`

// Main is the applet entry point.
func Main(argv []string) int {
	return hashing.Run("sha256sum", "sha256", argv)
}
