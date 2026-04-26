// Package tar implements a minimal `tar` applet: create, extract, list.
//
// Supported flags: -c create, -x extract, -t list, -f FILE, -z gzip,
// -v verbose. Long options and many GNU extensions are deferred.
package tar

import (
	"archive/tar"
	"compress/gzip"
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
		Name:  "tar",
		Help:  "create / extract / list tar archives",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tar -c[zvf] [ARCHIVE] FILE...
       tar -x[zvf] [ARCHIVE]
       tar -t[zvf] [ARCHIVE]
Create, extract, or list contents of a tar archive.

Mode (exactly one):
  -c, --create     create a new archive
  -x, --extract    extract files from an archive
  -t, --list       list the contents of an archive

Options:
  -f FILE          read/write archive to FILE (default: stdin/stdout)
  -z, --gzip       (de)compress the archive through gzip
  -v, --verbose    list files processed
  -C DIR           change to DIR before processing
`

type mode int

const (
	modeNone mode = iota
	modeCreate
	modeExtract
	modeList
)

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		m       mode
		archive string
		gzipped bool
		verbose bool
		cdir    string
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c", a == "--create":
			m = modeCreate
			args = args[1:]
		case a == "-x", a == "--extract":
			m = modeExtract
			args = args[1:]
		case a == "-t", a == "--list":
			m = modeList
			args = args[1:]
		case a == "-f":
			if len(args) < 2 {
				ioutil.Errf("tar: option requires an argument -- 'f'")
				return 2
			}
			archive = args[1]
			args = args[2:]
		case a == "-z", a == "--gzip":
			gzipped = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case a == "-C":
			if len(args) < 2 {
				ioutil.Errf("tar: option requires an argument -- 'C'")
				return 2
			}
			cdir = args[1]
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Combined flags like -cvzf or -xvf
			combo := a[1:]
			i := 0
			for i < len(combo) {
				c := combo[i]
				switch c {
				case 'c':
					m = modeCreate
				case 'x':
					m = modeExtract
				case 't':
					m = modeList
				case 'z':
					gzipped = true
				case 'v':
					verbose = true
				case 'f':
					if i != len(combo)-1 {
						ioutil.Errf("tar: -f must be last in combined flag")
						return 2
					}
					if len(args) < 2 {
						ioutil.Errf("tar: option requires an argument -- 'f'")
						return 2
					}
					archive = args[1]
					args = args[1:]
				default:
					ioutil.Errf("tar: invalid option -- '%c'", c)
					return 2
				}
				i++
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if m == modeNone {
		ioutil.Errf("tar: no mode specified (use -c, -x, or -t)")
		return 2
	}

	if cdir != "" {
		if err := os.Chdir(cdir); err != nil {
			ioutil.Errf("tar: -C %s: %v", cdir, err)
			return 1
		}
	}

	switch m {
	case modeCreate:
		return doCreate(archive, gzipped, verbose, args)
	case modeExtract:
		return doExtract(archive, gzipped, verbose)
	case modeList:
		return doList(archive, gzipped, verbose)
	}
	return 0
}

func doCreate(archive string, gzipped, verbose bool, files []string) int {
	if len(files) == 0 {
		ioutil.Errf("tar: missing files to archive")
		return 2
	}
	var out io.Writer
	out = ioutil.Stdout
	if archive != "" && archive != "-" {
		f, err := os.Create(archive) //nolint:gosec // user-supplied path
		if err != nil {
			ioutil.Errf("tar: %s: %v", archive, err)
			return 1
		}
		defer func() { _ = f.Close() }()
		out = f
	}
	var w io.Writer
	w = out
	var gw *gzip.Writer
	if gzipped {
		gw = gzip.NewWriter(out)
		w = gw
	}
	tw := tar.NewWriter(w)
	defer func() {
		_ = tw.Close()
		if gw != nil {
			_ = gw.Close()
		}
	}()
	rc := 0
	for _, name := range files {
		if err := addToArchive(tw, name, verbose); err != nil {
			ioutil.Errf("tar: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func addToArchive(tw *tar.Writer, root string, verbose bool) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(path)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if verbose {
			ioutil.Errf("%s", path)
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path) //nolint:gosec // walking user-supplied root
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()
		}
		return nil
	})
}

func openArchive(archive string, gzipped bool) (io.Reader, func(), error) {
	var r io.Reader
	r = ioutil.Stdin
	closer := func() {}
	if archive != "" && archive != "-" {
		f, err := os.Open(archive) //nolint:gosec // user-supplied path
		if err != nil {
			return nil, nil, err
		}
		r = f
		closer = func() { _ = f.Close() }
	}
	if gzipped {
		gr, err := gzip.NewReader(r)
		if err != nil {
			closer()
			return nil, nil, err
		}
		oldCloser := closer
		closer = func() {
			_ = gr.Close()
			oldCloser()
		}
		r = gr
	}
	return r, closer, nil
}

func doExtract(archive string, gzipped, verbose bool) int {
	r, closer, err := openArchive(archive, gzipped)
	if err != nil {
		ioutil.Errf("tar: %v", err)
		return 1
	}
	defer closer()

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return 0
		}
		if err != nil {
			ioutil.Errf("tar: %v", err)
			return 1
		}
		// Sanitize: refuse paths that escape the cwd.
		clean := filepath.Clean(hdr.Name)
		if strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
			ioutil.Errf("tar: refusing to extract %q (path traversal)", hdr.Name)
			return 1
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(clean, os.FileMode(hdr.Mode)); err != nil {
				ioutil.Errf("tar: %v", err)
				return 1
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(clean), 0o750); err != nil {
				ioutil.Errf("tar: %v", err)
				return 1
			}
			f, err := os.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode)) //nolint:gosec // archive contents are user-supplied
			if err != nil {
				ioutil.Errf("tar: %v", err)
				return 1
			}
			if _, err := io.Copy(f, tr); err != nil { //nolint:gosec // tar reader limited to archive size
				_ = f.Close()
				ioutil.Errf("tar: %v", err)
				return 1
			}
			_ = f.Close()
		}
		if verbose {
			_, _ = fmt.Fprintln(ioutil.Stdout, hdr.Name)
		}
	}
}

func doList(archive string, gzipped, verbose bool) int {
	r, closer, err := openArchive(archive, gzipped)
	if err != nil {
		ioutil.Errf("tar: %v", err)
		return 1
	}
	defer closer()
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return 0
		}
		if err != nil {
			ioutil.Errf("tar: %v", err)
			return 1
		}
		if verbose {
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s %d %s\n",
				os.FileMode(hdr.Mode), hdr.Size, hdr.Name)
		} else {
			_, _ = fmt.Fprintln(ioutil.Stdout, hdr.Name)
		}
	}
}
