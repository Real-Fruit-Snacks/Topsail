// Package xxd implements the `xxd` applet: hex-dump or hex-revert.
//
// The default output style matches xxd's canonical layout:
//
//	00000000: 4865 6c6c 6f2c 2074 6f70 7361 696c 210a  Hello, topsail!.
//
// Options supported: -p (plain hex / postscript), -r (revert hex back
// to bytes), -c COLS (bytes per line), -g GROUP (bytes per group). The
// less-common modes (-i C-include, -b binary, -s seek) are deferred.
package xxd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "xxd",
		Help:  "make a hex dump or revert one back to binary",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: xxd [OPTION]... [FILE]
Make a hexadecimal dump of FILE (or stdin), or revert one back to binary.

Options:
  -p, --postscript     plain hex output (no offsets, no ASCII column)
  -r, --revert         revert plain or canonical hex back to binary
  -c COLS              bytes per output line (default 16)
  -g GROUP             bytes per group within a line (default 2; 0 = no grouping)
  -u                   uppercase hex digits
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		plain     bool
		revert    bool
		uppercase bool
		cols      = 16
		group     = 2
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-p", a == "--postscript", a == "-ps":
			plain = true
			args = args[1:]
		case a == "-r", a == "--revert":
			revert = true
			args = args[1:]
		case a == "-u":
			uppercase = true
			args = args[1:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("xxd: option requires an argument -- 'c'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("xxd: invalid -c: %s", args[1])
				return 2
			}
			cols = n
			args = args[2:]
		case a == "-g":
			if len(args) < 2 {
				ioutil.Errf("xxd: option requires an argument -- 'g'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("xxd: invalid -g: %s", args[1])
				return 2
			}
			group = n
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("xxd: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	var (
		r io.Reader
		w *bufio.Writer
	)
	if len(args) == 0 || args[0] == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(args[0]) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			ioutil.Errf("xxd: %s: %v", args[0], err)
			return 1
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	w = bufio.NewWriter(ioutil.Stdout)
	defer func() { _ = w.Flush() }()

	if revert {
		if err := doRevert(r, w, plain); err != nil {
			ioutil.Errf("xxd: %v", err)
			return 1
		}
		return 0
	}

	if plain {
		if err := doPlain(r, w, cols, uppercase); err != nil {
			ioutil.Errf("xxd: %v", err)
			return 1
		}
		return 0
	}

	if err := doCanonical(r, w, cols, group, uppercase); err != nil {
		ioutil.Errf("xxd: %v", err)
		return 1
	}
	return 0
}

func doCanonical(r io.Reader, w *bufio.Writer, cols, group int, uppercase bool) error {
	buf := make([]byte, cols)
	offset := int64(0)
	hexFmt := "%02x"
	if uppercase {
		hexFmt = "%02X"
	}
	for {
		n, err := io.ReadFull(r, buf)
		if n == 0 {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}
		// Offset: 8-digit lowercase hex, ":" + space.
		_, _ = fmt.Fprintf(w, "%08x: ", offset)
		// Hex columns.
		for i := 0; i < cols; i++ {
			if i < n {
				_, _ = fmt.Fprintf(w, hexFmt, buf[i])
			} else {
				_, _ = w.WriteString("  ")
			}
			if group > 0 && (i+1)%group == 0 {
				_ = w.WriteByte(' ')
			}
		}
		// ASCII column.
		_ = w.WriteByte(' ')
		for i := 0; i < n; i++ {
			c := buf[i]
			if c < 0x20 || c >= 0x7f {
				c = '.'
			}
			_ = w.WriteByte(c)
		}
		_ = w.WriteByte('\n')
		offset += int64(n)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func doPlain(r io.Reader, w *bufio.Writer, cols int, uppercase bool) error {
	buf := make([]byte, cols)
	hexFmt := "%02x"
	if uppercase {
		hexFmt = "%02X"
	}
	for {
		n, err := io.ReadFull(r, buf)
		if n == 0 {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}
		for i := 0; i < n; i++ {
			_, _ = fmt.Fprintf(w, hexFmt, buf[i])
		}
		_ = w.WriteByte('\n')
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func doRevert(r io.Reader, w *bufio.Writer, plain bool) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		hexPart := line
		if !plain {
			// Canonical line: "OFFSET: HEXBYTES  ASCII"
			if i := strings.Index(line, ":"); i >= 0 {
				rest := line[i+1:]
				// Take everything up to the first run of two spaces (which
				// separates the hex columns from the ASCII column).
				if idx := strings.Index(rest, "  "); idx >= 0 {
					rest = rest[:idx]
				}
				hexPart = rest
			}
		}
		// Drop whitespace from the hex part.
		hexPart = stripWhitespace(hexPart)
		if hexPart == "" {
			continue
		}
		if len(hexPart)%2 == 1 {
			return fmt.Errorf("revert: odd hex digit count on line %q", line)
		}
		raw, err := hex.DecodeString(hexPart)
		if err != nil {
			return err
		}
		if _, err := w.Write(raw); err != nil {
			return err
		}
	}
	return sc.Err()
}

func stripWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
