// Package ls implements a minimal `ls` applet.
//
// Supported flags: -a, -A, -l, -h, -1, -R. Per-applet long format is
// best-effort cross-platform; full POSIX `ls -l` (with major/minor for
// devices, etc.) is deferred to a later wave.
package ls

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
	"github.com/Real-Fruit-Snacks/topsail/internal/platform"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "ls",
		Help:  "list directory contents",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: ls [OPTION]... [FILE]...
List information about FILE(s) (the current directory by default).

Options:
  -a, --all               include entries starting with '.'
  -A, --almost-all        like -a but exclude . and ..
  -l                      use long listing format
  -h, --human-readable    with -l, print sizes like 1.2K, 3.4M
  -1                      list one entry per line
  -R, --recursive         list subdirectories recursively
`

type opts struct {
	showAll, almostAll, longFmt, human, oneLine, recursive bool
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var o opts

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-a", a == "--all":
			o.showAll = true
			args = args[1:]
		case a == "-A", a == "--almost-all":
			o.almostAll = true
			args = args[1:]
		case a == "-l":
			o.longFmt = true
			args = args[1:]
		case a == "-h", a == "--human-readable":
			o.human = true
			args = args[1:]
		case a == "-1":
			o.oneLine = true
			args = args[1:]
		case a == "-R", a == "--recursive":
			o.recursive = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'a':
					o.showAll = true
				case 'A':
					o.almostAll = true
				case 'l':
					o.longFmt = true
				case 'h':
					o.human = true
				case '1':
					o.oneLine = true
				case 'R':
					o.recursive = true
				default:
					ioutil.Errf("ls: invalid option -- '%c'", c)
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

	rc := 0
	for i, p := range paths {
		if i > 0 {
			_, _ = fmt.Fprintln(ioutil.Stdout)
		}
		if err := listOne(p, len(paths) > 1 || o.recursive, o); err != nil {
			ioutil.Errf("ls: %s: %v", p, err)
			rc = 1
		}
	}
	return rc
}

func listOne(path string, withHeader bool, o opts) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		printEntry(path, info, o)
		return nil
	}
	if withHeader {
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s:\n", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		name := e.Name()
		if !o.showAll && !o.almostAll && strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(path, name)
		info, err := os.Lstat(full)
		if err != nil {
			ioutil.Errf("ls: %s: %v", full, err)
			continue
		}
		printEntry(name, info, o)
	}
	if o.recursive {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if !o.showAll && strings.HasPrefix(name, ".") {
				continue
			}
			_, _ = fmt.Fprintln(ioutil.Stdout)
			if err := listOne(filepath.Join(path, name), true, o); err != nil {
				ioutil.Errf("ls: %v", err)
			}
		}
	}
	return nil
}

func printEntry(name string, info fs.FileInfo, o opts) {
	if o.longFmt {
		mode := info.Mode().String()
		size := fmt.Sprintf("%d", info.Size())
		if o.human {
			size = humanSize(info.Size())
		}
		mtime := info.ModTime().Format("2006-01-02 15:04")
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s %s %s %s %s\n",
			mode,
			platform.UserName("0"),
			platform.GroupName("0"),
			size,
			mtime+" "+name)
		return
	}
	_, _ = fmt.Fprintln(ioutil.Stdout, name)
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
