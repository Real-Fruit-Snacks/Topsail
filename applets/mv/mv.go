// Package mv implements the `mv` applet: move (rename) files.
package mv

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "mv",
		Help:  "move (rename) files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: mv [OPTION]... SOURCE... DEST
Rename SOURCE to DEST, or move SOURCE(s) into DEST directory.

Options:
  -f, --force          do not prompt before overwriting (default)
  -i, --interactive    accepted but never prompts
  -n, --no-clobber     do not overwrite an existing file
  -v, --verbose        print a message for each move
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var noClobber, verbose bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-f", a == "--force":
			noClobber = false
			args = args[1:]
		case a == "-i", a == "--interactive":
			args = args[1:]
		case a == "-n", a == "--no-clobber":
			noClobber = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'f':
					noClobber = false
				case 'i':
					// no-op (we never prompt)
				case 'n':
					noClobber = true
				case 'v':
					verbose = true
				default:
					ioutil.Errf("mv: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("mv: missing destination operand")
		return 2
	}

	dest := args[len(args)-1]
	sources := args[:len(args)-1]

	destInfo, err := os.Stat(dest)
	destIsDir := err == nil && destInfo.IsDir()

	if len(sources) > 1 && !destIsDir {
		ioutil.Errf("mv: target '%s' is not a directory", dest)
		return 2
	}

	rc := 0
	for _, src := range sources {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}
		if noClobber {
			if _, err := os.Stat(target); err == nil {
				continue
			} else if !errors.Is(err, fs.ErrNotExist) {
				ioutil.Errf("mv: %s: %v", target, err)
				rc = 1
				continue
			}
		}
		if err := move(src, target); err != nil {
			ioutil.Errf("mv: cannot move '%s' to '%s': %v", src, target, err)
			rc = 1
			continue
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "renamed '%s' -> '%s'\n", src, target)
		}
	}
	return rc
}

// renameFunc is the rename primitive move() calls first. It is a package
// var so tests can simulate cross-device failures without needing a real
// cross-device setup.
var renameFunc = os.Rename

// move tries renameFunc first (atomic, fast) and falls back to copy+delete
// when rename fails — typically because the source and destination are
// on different filesystems.
func move(src, dst string) error {
	if err := renameFunc(src, dst); err == nil {
		return nil
	}
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return errors.New("cross-device directory move not supported")
	}
	if err := copyFile(src, dst, srcInfo.Mode()); err != nil {
		return err
	}
	return os.Remove(src)
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm()) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
