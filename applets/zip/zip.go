// Package zip implements `zip` and `unzip` applets via archive/zip.
package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "zip",
		Help:  "package and compress files into a zip archive",
		Usage: zipUsage,
		Main:  Main,
	})
	applet.Register(applet.Applet{
		Name:  "unzip",
		Help:  "extract files from a zip archive",
		Usage: unzipUsage,
		Main:  UnzipMain,
	})
}

const zipUsage = `Usage: zip [OPTION]... ARCHIVE FILE...
Add FILE(s) to a new ZIP ARCHIVE.

Options:
  -r, --recurse-paths   recurse into directories
  -q, --quiet           suppress per-file output
  -j, --junk-paths      strip path prefix; store basename only
`

const unzipUsage = `Usage: unzip [OPTION]... ARCHIVE [FILE]...
Extract files from a zip ARCHIVE.

Options:
  -d DIR, --directory=DIR   extract into DIR (default: cwd)
  -l, --list                only list archive contents
  -o, --overwrite           overwrite existing files without prompting
  -q, --quiet               suppress per-file output
`

// Main is the zip entry point.
func Main(argv []string) int {
	args := argv[1:]
	var recurse, quiet, junk bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-r", a == "--recurse-paths":
			recurse = true
			args = args[1:]
		case a == "-q", a == "--quiet":
			quiet = true
			args = args[1:]
		case a == "-j", a == "--junk-paths":
			junk = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("zip: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("zip: missing archive name or files")
		return 2
	}
	archive := args[0]
	files := args[1:]

	out, err := os.Create(archive) //nolint:gosec // user-supplied path
	if err != nil {
		ioutil.Errf("zip: %s: %v", archive, err)
		return 1
	}
	defer func() { _ = out.Close() }()

	zw := zip.NewWriter(out)
	defer func() { _ = zw.Close() }()

	rc := 0
	for _, name := range files {
		if err := addToZip(zw, name, recurse, junk, quiet); err != nil {
			ioutil.Errf("zip: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func addToZip(zw *zip.Writer, root string, recurse, junk, quiet bool) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if info.IsDir() && !recurse {
		return fmt.Errorf("not adding directory '%s' (-r required)", root)
	}
	walk := filepath.Walk
	if !info.IsDir() {
		walk = func(p string, fn filepath.WalkFunc) error {
			info, err := os.Lstat(p)
			if err != nil {
				return err
			}
			return fn(p, info, nil)
		}
	}
	return walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := path
		if junk {
			name = filepath.Base(path)
		}
		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		fh.Name = filepath.ToSlash(name)
		fh.Method = zip.Deflate
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}
		f, err := os.Open(path) //nolint:gosec // walking user-supplied root
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		_ = f.Close()
		if err != nil {
			return err
		}
		if !quiet {
			_, _ = fmt.Fprintln(ioutil.Stdout, "  adding: "+name)
		}
		return nil
	})
}

// UnzipMain is the unzip entry point.
func UnzipMain(argv []string) int {
	args := argv[1:]
	var dest string
	var listOnly, overwrite, quiet bool

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-d":
			if len(args) < 2 {
				ioutil.Errf("unzip: option requires an argument -- 'd'")
				return 2
			}
			dest = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "--directory="):
			dest = a[len("--directory="):]
			args = args[1:]
		case a == "-l", a == "--list":
			listOnly = true
			args = args[1:]
		case a == "-o", a == "--overwrite":
			overwrite = true
			args = args[1:]
		case a == "-q", a == "--quiet":
			quiet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("unzip: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("unzip: missing archive name")
		return 2
	}
	archive := args[0]

	zr, err := zip.OpenReader(archive)
	if err != nil {
		ioutil.Errf("unzip: %s: %v", archive, err)
		return 1
	}
	defer func() { _ = zr.Close() }()

	if dest == "" {
		dest = "."
	}

	rc := 0
	for _, f := range zr.File {
		clean := filepath.Clean(f.Name)
		if strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
			ioutil.Errf("unzip: refusing %q (path traversal)", f.Name)
			rc = 1
			continue
		}
		out := filepath.Join(dest, clean)
		if listOnly {
			_, _ = fmt.Fprintln(ioutil.Stdout, f.Name)
			continue
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(out, f.Mode()); err != nil {
				ioutil.Errf("unzip: %v", err)
				rc = 1
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
			ioutil.Errf("unzip: %v", err)
			rc = 1
			continue
		}
		flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !overwrite {
			flag |= os.O_EXCL
		}
		dst, err := os.OpenFile(out, flag, f.Mode()) //nolint:gosec // archive content is user-supplied
		if err != nil {
			ioutil.Errf("unzip: %v", err)
			rc = 1
			continue
		}
		src, err := f.Open()
		if err != nil {
			_ = dst.Close()
			ioutil.Errf("unzip: %v", err)
			rc = 1
			continue
		}
		_, err = io.Copy(dst, src) //nolint:gosec // size is bounded by archive entry
		_ = dst.Close()
		_ = src.Close()
		if err != nil {
			ioutil.Errf("unzip: %v", err)
			rc = 1
			continue
		}
		if !quiet {
			_, _ = fmt.Fprintln(ioutil.Stdout, "  inflating: "+f.Name)
		}
	}
	return rc
}
