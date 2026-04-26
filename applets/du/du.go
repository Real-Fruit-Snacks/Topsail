// Package du implements the `du` applet: estimate file space usage.
package du

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "du",
		Help:  "estimate file space usage",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: du [OPTION]... [FILE]...
Summarize disk usage of the set of FILEs, recursively for directories.

Options:
  -s, --summarize        display only a total for each argument
  -h, --human-readable   print sizes in human-readable form (e.g. 1.2K)
  -a, --all              write counts for all files, not just directories
  -k                     report sizes in 1024-byte blocks (default)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var summary, human, all bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-s", a == "--summarize":
			summary = true
			args = args[1:]
		case a == "-h", a == "--human-readable":
			human = true
			args = args[1:]
		case a == "-a", a == "--all":
			all = true
			args = args[1:]
		case a == "-k":
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 's':
					summary = true
				case 'h':
					human = true
				case 'a':
					all = true
				case 'k':
				default:
					ioutil.Errf("du: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		args = []string{"."}
	}

	rc := 0
	for _, name := range args {
		_, err := walk(name, summary, human, all)
		if err != nil {
			ioutil.Errf("du: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

// walk returns the total size in bytes under root.
func walk(root string, summary, human, all bool) (int64, error) {
	// Surface "doesn't exist" up front rather than swallowing it inside
	// the WalkDir callback (where returning nil would mask the error).
	if _, err := os.Stat(root); err != nil {
		return 0, err
	}
	var total int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			ioutil.Errf("du: %s: %v", path, err)
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			total += info.Size()
			if !summary && all {
				printSize(path, info.Size(), human)
			}
			return nil
		}
		// Print directory totals on the way back up.
		// We track and emit when leaving the directory in a separate
		// pass below. For simplicity here we emit non-summary at the
		// root only.
		return nil
	})
	if err != nil {
		return total, err
	}
	// Whether or not -s, we print at least the root total; per-subdir
	// totals are deferred (Wave 6 candidate).
	printSize(root, total, human)
	return total, nil
}

func printSize(path string, size int64, human bool) {
	if human {
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s\t%s\n", humanSize(size), path)
		return
	}
	// POSIX du reports in 1024-byte blocks (rounded up).
	blocks := (size + 1023) / 1024
	_, _ = fmt.Fprintf(ioutil.Stdout, "%d\t%s\n", blocks, path)
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
