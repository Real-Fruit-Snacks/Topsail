// Package gzip implements the `gzip` and `gunzip` applets.
package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "gzip",
		Help:  "compress files (gzip)",
		Usage: usage,
		Main:  Main,
	})
	applet.Register(applet.Applet{
		Name:  "gunzip",
		Help:  "decompress gzip files",
		Usage: gunzipUsage,
		Main:  GunzipMain,
	})
}

const usage = `Usage: gzip [OPTION]... [FILE]...
Compress each FILE in place, replacing it with FILE.gz. With no FILE,
or when FILE is -, read stdin and write compressed bytes to stdout.

Options:
  -d, --decompress    decompress (same as gunzip)
  -k, --keep          keep input files (don't delete after compressing)
  -c, --stdout        write to stdout, keep input files unchanged
  -f, --force         overwrite existing files without prompting
  -1 ... -9           compression level (1=fast, 9=best, default=6)
`

const gunzipUsage = `Usage: gunzip [OPTION]... [FILE]...
Decompress each FILE.gz in place. With no FILE, or when FILE is -, read
stdin and write decompressed bytes to stdout.

Options:
  -k, --keep          keep input files
  -c, --stdout        write to stdout, keep input files
  -f, --force         overwrite existing files
`

type opts struct {
	decompress, keep, toStdout, force bool
	level                             int
}

// Main is the gzip entry point.
func Main(argv []string) int {
	o, files, rc := parseFlags(argv, false)
	if rc != 0 {
		return rc
	}
	return run(o, files)
}

// GunzipMain is the gunzip entry point.
func GunzipMain(argv []string) int {
	o, files, rc := parseFlags(argv, true)
	if rc != 0 {
		return rc
	}
	o.decompress = true
	return run(o, files)
}

func parseFlags(argv []string, isGunzip bool) (o opts, files []string, rc int) {
	args := argv[1:]
	o = opts{level: gzip.DefaultCompression}
	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-d", a == "--decompress":
			o.decompress = true
			args = args[1:]
		case a == "-k", a == "--keep":
			o.keep = true
			args = args[1:]
		case a == "-c", a == "--stdout", a == "--to-stdout":
			o.toStdout = true
			o.keep = true
			args = args[1:]
		case a == "-f", a == "--force":
			o.force = true
			args = args[1:]
		case len(a) == 2 && a[0] == '-' && a[1] >= '1' && a[1] <= '9':
			o.level = int(a[1] - '0')
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			cmd := "gzip"
			if isGunzip {
				cmd = "gunzip"
			}
			ioutil.Errf("%s: unknown option: %s", cmd, a)
			return o, nil, 2
		default:
			stop = true
		}
	}
	return o, args, 0
}

func run(o opts, files []string) int {
	if len(files) == 0 {
		// stdin -> stdout
		return runStream(o)
	}
	rc := 0
	for _, name := range files {
		if name == "-" {
			rc |= runStream(o)
			continue
		}
		if err := runFile(name, o); err != nil {
			cmd := "gzip"
			if o.decompress {
				cmd = "gunzip"
			}
			ioutil.Errf("%s: %s: %v", cmd, name, err)
			rc = 1
		}
	}
	return rc
}

func runStream(o opts) int {
	if o.decompress {
		gr, err := gzip.NewReader(ioutil.Stdin)
		if err != nil {
			ioutil.Errf("gunzip: %v", err)
			return 1
		}
		if _, err := io.Copy(ioutil.Stdout, gr); err != nil { //nolint:gosec // we trust the user-supplied input; gzip bombs are user's risk
			ioutil.Errf("gunzip: %v", err)
			return 1
		}
		return 0
	}
	gw, err := gzip.NewWriterLevel(ioutil.Stdout, o.level)
	if err != nil {
		ioutil.Errf("gzip: %v", err)
		return 1
	}
	if _, err := io.Copy(gw, ioutil.Stdin); err != nil {
		_ = gw.Close()
		ioutil.Errf("gzip: %v", err)
		return 1
	}
	if err := gw.Close(); err != nil {
		ioutil.Errf("gzip: %v", err)
		return 1
	}
	return 0
}

func runFile(name string, o opts) error {
	in, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	var dstName string
	if o.decompress {
		dstName = strings.TrimSuffix(name, ".gz")
		if dstName == name {
			return fmt.Errorf("file %s does not have .gz suffix", name)
		}
	} else {
		dstName = name + ".gz"
	}

	var out io.Writer
	if o.toStdout {
		out = ioutil.Stdout
	} else {
		flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !o.force {
			flag |= os.O_EXCL
		}
		f, err := os.OpenFile(dstName, flag, 0o644) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		out = f
	}

	if o.decompress {
		gr, err := gzip.NewReader(in)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, gr); err != nil { //nolint:gosec // we trust the user-supplied input; gzip bombs are user's risk
			return err
		}
		_ = gr.Close()
	} else {
		gw, err := gzip.NewWriterLevel(out, o.level)
		if err != nil {
			return err
		}
		if _, err := io.Copy(gw, in); err != nil {
			_ = gw.Close()
			return err
		}
		if err := gw.Close(); err != nil {
			return err
		}
	}

	if !o.keep && !o.toStdout {
		_ = in.Close()
		_ = os.Remove(name)
	}
	return nil
}
