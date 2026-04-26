// Package base64 implements the `base64` applet: encode/decode RFC 4648.
package base64

import (
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "base64",
		Help:  "encode or decode data as base64",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: base64 [OPTION]... [FILE]
Base64-encode (default) or decode FILE (or stdin).

Options:
  -d, --decode      decode rather than encode
  -w COLS, --wrap=COLS   wrap encoded output every COLS characters (default 76; 0 disables)
  -i, --ignore-garbage   when decoding, ignore non-base64 bytes
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var decode, ignoreGarbage bool
	wrap := 76

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-d", a == "--decode":
			decode = true
			args = args[1:]
		case a == "-i", a == "--ignore-garbage":
			ignoreGarbage = true
			args = args[1:]
		case a == "-w":
			if len(args) < 2 {
				ioutil.Errf("base64: option requires an argument -- 'w'")
				return 2
			}
			n, err := parseUint(args[1])
			if err != nil {
				ioutil.Errf("base64: invalid wrap value: %s", args[1])
				return 2
			}
			wrap = n
			args = args[2:]
		case strings.HasPrefix(a, "--wrap="):
			n, err := parseUint(a[len("--wrap="):])
			if err != nil {
				ioutil.Errf("base64: invalid wrap value: %s", a)
				return 2
			}
			wrap = n
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("base64: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	if len(args) > 1 {
		ioutil.Errf("base64: extra operand: %s", args[1])
		return 2
	}
	src := "-"
	if len(args) == 1 {
		src = args[0]
	}

	r, closer, err := openIn(src)
	if err != nil {
		ioutil.Errf("base64: %s: %v", src, err)
		return 1
	}
	defer closer()

	if decode {
		input, rerr := io.ReadAll(r)
		if rerr != nil {
			ioutil.Errf("base64: %v", rerr)
			return 1
		}
		if ignoreGarbage {
			input = stripGarbage(input)
		}
		clean := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, string(input))
		decoded, derr := base64.StdEncoding.DecodeString(clean)
		if derr != nil {
			ioutil.Errf("base64: %v", derr)
			return 1
		}
		if _, werr := ioutil.Stdout.Write(decoded); werr != nil {
			return 1
		}
		return 0
	}

	input, err := io.ReadAll(r)
	if err != nil {
		ioutil.Errf("base64: %v", err)
		return 1
	}
	encoded := base64.StdEncoding.EncodeToString(input)
	if wrap > 0 {
		var b strings.Builder
		for i := 0; i < len(encoded); i += wrap {
			end := i + wrap
			if end > len(encoded) {
				end = len(encoded)
			}
			b.WriteString(encoded[i:end])
			b.WriteByte('\n')
		}
		encoded = b.String()
	} else {
		encoded += "\n"
	}
	_, _ = io.WriteString(ioutil.Stdout, encoded)
	return 0
}

func parseUint(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errBadInt
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

var errBadInt = stringErr("not a non-negative integer")

type stringErr string

func (e stringErr) Error() string { return string(e) }

func stripGarbage(input []byte) []byte {
	// Allow A-Z a-z 0-9 + / = and whitespace.
	out := make([]byte, 0, len(input))
	for _, b := range input {
		switch {
		case b >= 'A' && b <= 'Z',
			b >= 'a' && b <= 'z',
			b >= '0' && b <= '9',
			b == '+', b == '/', b == '=',
			b == '\n', b == '\r', b == ' ', b == '\t':
			out = append(out, b)
		}
	}
	return out
}

func openIn(name string) (io.Reader, func(), error) {
	if name == "-" {
		return ioutil.Stdin, func() {}, nil
	}
	f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}
