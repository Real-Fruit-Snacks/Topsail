// Package df implements a minimal `df` applet: report filesystem usage.
//
// Cross-platform filesystem statistics in pure Go without cgo are
// limited; this build prints best-effort totals for the path(s) given,
// or a placeholder line for `df` with no args.
package df

import (
	"fmt"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "df",
		Help:  "report file system disk space usage",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: df [OPTION]... [PATH]...
Show how much disk space is available on each filesystem containing PATH
(or, with no PATH, on every mounted filesystem).

Options:
  -h, --human-readable   print sizes in human-readable form
  -k                     report sizes in 1024-byte blocks (default)

This Wave 3 build uses platform-specific syscalls for the listed PATHs
and prints them. Full mount-table enumeration is deferred.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var human bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-h", a == "--human-readable":
			human = true
			args = args[1:]
		case a == "-k":
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'h':
					human = true
				case 'k':
				default:
					ioutil.Errf("df: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	_, _ = fmt.Fprintln(ioutil.Stdout, headerLine(human))
	rc := 0
	for _, p := range paths {
		stats, err := statFS(p)
		if err != nil {
			ioutil.Errf("df: %s: %v", p, err)
			rc = 1
			continue
		}
		_, _ = fmt.Fprintln(ioutil.Stdout, formatLine(p, stats, human))
	}
	return rc
}

func headerLine(human bool) string {
	if human {
		return "Filesystem      Size  Used Avail Use% Mounted on"
	}
	return "Filesystem     1K-blocks      Used Available Use% Mounted on"
}

type fsStats struct {
	totalBytes uint64
	freeBytes  uint64
}

func formatLine(path string, s fsStats, human bool) string {
	used := int64(s.totalBytes) - int64(s.freeBytes)
	if used < 0 {
		used = 0
	}
	var pct int
	if s.totalBytes > 0 {
		pct = int(uint64(used) * 100 / s.totalBytes)
	}
	if human {
		return fmt.Sprintf("%-15s %5s %5s %5s %3d%% %s",
			"-", humanSize(int64(s.totalBytes)), humanSize(used),
			humanSize(int64(s.freeBytes)), pct, path)
	}
	return fmt.Sprintf("%-15s %9d %9d %9d %3d%% %s",
		"-", s.totalBytes/1024, used/1024, s.freeBytes/1024, pct, path)
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(n)/float64(div), "KMGTPE"[exp])
}
