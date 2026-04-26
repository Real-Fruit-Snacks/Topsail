// Package md5sum implements the `md5sum` applet.
package md5sum

import (
	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/hashing"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "md5sum",
		Help:  "compute and check MD5 checksums",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: md5sum [OPTION]... [FILE]...
Print or check MD5 (128-bit) checksums.

Options:
  -c, --check       read sums from FILEs and verify them
  -b, --binary      read in binary mode (accepted; identical on Go)
  -t, --text        read in text mode (accepted; identical on Go)
      --quiet       with -c, suppress per-file 'OK' lines

Note: MD5 is collision-broken. Use only for non-security checksums.
`

// Main is the applet entry point.
func Main(argv []string) int {
	return hashing.Run("md5sum", "md5", argv)
}
