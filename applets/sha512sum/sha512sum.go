// Package sha512sum implements the `sha512sum` applet.
package sha512sum

import (
	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/hashing"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sha512sum",
		Help:  "compute and check SHA-512 checksums",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sha512sum [OPTION]... [FILE]...
Print or check SHA-512 (512-bit) checksums.

Options:
  -c, --check       read sums from FILEs and verify them
  -b, --binary      read in binary mode (accepted; identical on Go)
  -t, --text        read in text mode (accepted; identical on Go)
      --quiet       with -c, suppress per-file 'OK' lines
`

// Main is the applet entry point.
func Main(argv []string) int {
	return hashing.Run("sha512sum", "sha512", argv)
}
