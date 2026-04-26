// Package file implements a minimal `file` applet: detect file types
// by examining magic bytes and content.
package file

import (
	"bytes"
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "file",
		Help:  "determine file type",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: file FILE...
Print a one-line description of FILE's content type.

The detection is best-effort, based on the first 8 KiB of each file.
Recognized: ELF / Mach-O / PE binaries, gzip / bzip2 / xz / zip / tar
archives, common image formats (PNG / JPEG / GIF / WebP), PDF, ASCII
text, UTF-8 text, and "data" as the catch-all.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) == 0 {
		ioutil.Errf("file: missing operand")
		return 2
	}
	rc := 0
	for _, name := range args {
		desc, err := describe(name)
		if err != nil {
			ioutil.Errf("file: %s: %v", name, err)
			rc = 1
			continue
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s: %s\n", name, desc)
	}
	return rc
}

func describe(name string) (string, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "directory", nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, lerr := os.Readlink(name)
		if lerr == nil {
			return "symbolic link to " + target, nil
		}
		return "symbolic link", nil
	}
	if info.Size() == 0 {
		return "empty", nil
	}

	f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	buf = buf[:n]

	for _, m := range magics {
		if bytes.HasPrefix(buf, m.prefix) {
			return m.label, nil
		}
	}
	if isASCII(buf) {
		return "ASCII text", nil
	}
	if utf8.Valid(buf) {
		return "UTF-8 text", nil
	}
	return "data", nil
}

type magic struct {
	prefix []byte
	label  string
}

var magics = []magic{
	{[]byte("\x7fELF"), "ELF binary"},
	{[]byte{0xCA, 0xFE, 0xBA, 0xBE}, "Mach-O fat binary"},
	{[]byte{0xCF, 0xFA, 0xED, 0xFE}, "Mach-O 64-bit binary"},
	{[]byte{0xCE, 0xFA, 0xED, 0xFE}, "Mach-O 32-bit binary"},
	{[]byte("MZ"), "PE binary (Windows executable)"},
	{[]byte{0x1F, 0x8B}, "gzip compressed data"},
	{[]byte("BZh"), "bzip2 compressed data"},
	{[]byte{0xFD, '7', 'z', 'X', 'Z', 0x00}, "xz compressed data"},
	{[]byte("PK\x03\x04"), "zip archive"},
	{[]byte("PK\x05\x06"), "zip archive (empty)"},
	{[]byte("PK\x07\x08"), "zip archive (spanned)"},
	{[]byte("ustar"), "tar archive"},
	{[]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, "PNG image"},
	{[]byte{0xFF, 0xD8, 0xFF}, "JPEG image"},
	{[]byte("GIF87a"), "GIF image (1987a)"},
	{[]byte("GIF89a"), "GIF image (1989a)"},
	{[]byte("RIFF"), "RIFF / WebP container"},
	{[]byte("%PDF-"), "PDF document"},
	{[]byte("#!"), "shell script"},
}

func isASCII(b []byte) bool {
	for _, c := range b {
		if c < 7 || (c > 13 && c < 32) || c > 126 {
			return false
		}
	}
	return true
}
