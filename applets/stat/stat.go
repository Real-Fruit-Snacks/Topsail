// Package stat implements a minimal `stat` applet.
package stat

import (
	"fmt"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "stat",
		Help:  "display file or filesystem status",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: stat [OPTION]... FILE...
Display information about each FILE.

Options:
  -t, --terse   print in a single-line, space-separated format
  -L            follow symlinks (default: lstat)
  -c FORMAT     printf-style FORMAT with %n=name %s=size %F=type
                %a=octal mode %y=mtime %i=inode (best-effort)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		terse, follow bool
		format        string
		formatSet     bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-t", a == "--terse":
			terse = true
			args = args[1:]
		case a == "-L":
			follow = true
			args = args[1:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("stat: option requires an argument -- 'c'")
				return 2
			}
			format = args[1]
			formatSet = true
			args = args[2:]
		case strings.HasPrefix(a, "--format="):
			format = a[len("--format="):]
			formatSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("stat: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("stat: missing operand")
		return 2
	}

	rc := 0
	for _, name := range args {
		var info os.FileInfo
		var err error
		if follow {
			info, err = os.Stat(name)
		} else {
			info, err = os.Lstat(name)
		}
		if err != nil {
			ioutil.Errf("stat: %s: %v", name, err)
			rc = 1
			continue
		}
		switch {
		case formatSet:
			_, _ = fmt.Fprintln(ioutil.Stdout, applyFormat(format, name, info))
		case terse:
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s %d %o %s\n",
				name, info.Size(), info.Mode().Perm(), info.ModTime().Format("2006-01-02T15:04:05"))
		default:
			printDefault(name, info)
		}
	}
	return rc
}

func printDefault(name string, info os.FileInfo) {
	_, _ = fmt.Fprintf(ioutil.Stdout, "  File: %s\n", name)
	_, _ = fmt.Fprintf(ioutil.Stdout, "  Size: %d\n", info.Size())
	_, _ = fmt.Fprintf(ioutil.Stdout, "  Type: %s\n", typeLabel(info))
	_, _ = fmt.Fprintf(ioutil.Stdout, " Modes: %s (%04o)\n", info.Mode().String(), info.Mode().Perm())
	_, _ = fmt.Fprintf(ioutil.Stdout, "Modify: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
}

func typeLabel(info os.FileInfo) string {
	mode := info.Mode()
	switch {
	case mode.IsDir():
		return "directory"
	case mode&os.ModeSymlink != 0:
		return "symbolic link"
	case mode&os.ModeNamedPipe != 0:
		return "fifo"
	case mode&os.ModeSocket != 0:
		return "socket"
	case mode&os.ModeDevice != 0:
		return "device"
	case mode.IsRegular():
		return "regular file"
	}
	return "other"
}

func applyFormat(format, name string, info os.FileInfo) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		switch format[i] {
		case 'n':
			b.WriteString(name)
		case 's':
			fmt.Fprintf(&b, "%d", info.Size())
		case 'F':
			b.WriteString(typeLabel(info))
		case 'a':
			fmt.Fprintf(&b, "%o", info.Mode().Perm())
		case 'y':
			b.WriteString(info.ModTime().Format("2006-01-02 15:04:05"))
		case 'i':
			b.WriteString("0") // inode best-effort: not portable.
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}
