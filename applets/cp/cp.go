// Package cp implements the `cp` applet: copy files and directories.
package cp

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
		Name:  "cp",
		Help:  "copy files and directories",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: cp [OPTION]... SOURCE DEST
       cp [OPTION]... SOURCE... DIRECTORY
Copy SOURCE to DEST, or multiple SOURCEs into DIRECTORY.

Options:
  -r, -R, --recursive    copy directories recursively
  -p, --preserve         preserve mode and modification time
  -f, --force            never prompt (default)
  -i, --interactive      accepted but never prompts
  -n, --no-clobber       do not overwrite an existing file
  -v, --verbose          print a message for each copy
`

type options struct {
	recursive, preserve, noClobber, verbose bool
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var opts options

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-r", a == "-R", a == "--recursive":
			opts.recursive = true
			args = args[1:]
		case a == "-p", a == "--preserve":
			opts.preserve = true
			args = args[1:]
		case a == "-f", a == "--force":
			opts.noClobber = false
			args = args[1:]
		case a == "-i", a == "--interactive":
			args = args[1:]
		case a == "-n", a == "--no-clobber":
			opts.noClobber = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			opts.verbose = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'r', 'R':
					opts.recursive = true
				case 'p':
					opts.preserve = true
				case 'f':
					opts.noClobber = false
				case 'i':
					// no-op
				case 'n':
					opts.noClobber = true
				case 'v':
					opts.verbose = true
				default:
					ioutil.Errf("cp: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("cp: missing destination operand")
		return 2
	}

	dest := args[len(args)-1]
	sources := args[:len(args)-1]

	destInfo, err := os.Stat(dest)
	destIsDir := err == nil && destInfo.IsDir()

	if len(sources) > 1 && !destIsDir {
		ioutil.Errf("cp: target '%s' is not a directory", dest)
		return 2
	}

	rc := 0
	for _, src := range sources {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}
		if err := copyEntry(src, target, opts); err != nil {
			ioutil.Errf("cp: %v", err)
			rc = 1
		}
	}
	return rc
}

func copyEntry(src, dst string, opts options) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if !opts.recursive {
			return fmt.Errorf("-r not specified; omitting directory '%s'", src)
		}
		return copyDir(src, dst, opts)
	}
	// Default behavior dereferences symlinks (matches GNU cp without -P).
	return copyFile(src, dst, info.Mode().Perm(), opts)
}

func copyFile(src, dst string, mode fs.FileMode, opts options) error {
	if opts.noClobber {
		if _, err := os.Stat(dst); err == nil {
			return nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	in, err := os.Open(src) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if opts.preserve {
		if info, err := os.Stat(src); err == nil {
			_ = os.Chtimes(dst, info.ModTime(), info.ModTime())
		}
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(ioutil.Stdout, "'%s' -> '%s'\n", src, dst)
	}
	return nil
}

func copyDir(src, dst string, opts options) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dst, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := copyEntry(s, d, opts); err != nil {
			return err
		}
	}
	if opts.preserve {
		_ = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(ioutil.Stdout, "'%s' -> '%s'\n", src, dst)
	}
	return nil
}
