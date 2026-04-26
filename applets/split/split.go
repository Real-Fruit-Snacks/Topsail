// Package split implements a minimal `split` applet: split a file by lines.
package split

import (
	"bufio"
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
		Name:  "split",
		Help:  "split a file into pieces",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: split [OPTION]... [FILE [PREFIX]]
Split FILE into pieces named PREFIX{aa,ab,...} (default PREFIX = "x").

Options:
  -l N, --lines=N   put N lines per output file (default 1000)
  -d                use numeric suffixes instead of alphabetic
  -a N              suffix length (default 2)
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	linesPerFile := 1000
	numeric := false
	suffixLen := 2

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-l":
			if len(args) < 2 {
				ioutil.Errf("split: option requires an argument -- 'l'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("split: invalid -l value: %s", args[1])
				return 2
			}
			linesPerFile = n
			args = args[2:]
		case strings.HasPrefix(a, "--lines="):
			n, err := strconv.Atoi(a[len("--lines="):])
			if err != nil || n < 1 {
				ioutil.Errf("split: invalid --lines value")
				return 2
			}
			linesPerFile = n
			args = args[1:]
		case a == "-d":
			numeric = true
			args = args[1:]
		case a == "-a":
			if len(args) < 2 {
				ioutil.Errf("split: option requires an argument -- 'a'")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				ioutil.Errf("split: invalid -a value: %s", args[1])
				return 2
			}
			suffixLen = n
			args = args[2:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("split: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	input := "-"
	prefix := "x"
	if len(args) >= 1 {
		input = args[0]
	}
	if len(args) >= 2 {
		prefix = args[1]
	}
	if len(args) > 2 {
		ioutil.Errf("split: extra operand: %s", args[2])
		return 2
	}

	r, closer, err := openIn(input)
	if err != nil {
		ioutil.Errf("split: %s: %v", input, err)
		return 1
	}
	defer closer()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	pieceIdx := 0
	var current *os.File
	linesInPiece := 0
	for sc.Scan() {
		if linesInPiece == 0 {
			name := prefix + suffix(pieceIdx, suffixLen, numeric)
			f, err := os.Create(name) //nolint:gosec // user-supplied prefix is the whole point
			if err != nil {
				ioutil.Errf("split: %s: %v", name, err)
				return 1
			}
			current = f
			pieceIdx++
		}
		if _, err := current.WriteString(sc.Text() + "\n"); err != nil {
			ioutil.Errf("split: %v", err)
			_ = current.Close()
			return 1
		}
		linesInPiece++
		if linesInPiece >= linesPerFile {
			_ = current.Close()
			current = nil
			linesInPiece = 0
		}
	}
	if current != nil {
		_ = current.Close()
	}
	return 0
}

// suffix produces the n-th suffix at the given length. Alphabetic when
// !numeric (aa, ab, ..., az, ba, ...); numeric otherwise (00, 01, ...).
func suffix(n, length int, numeric bool) string {
	if numeric {
		return fmt.Sprintf("%0*d", length, n)
	}
	out := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		out[i] = byte('a' + n%26)
		n /= 26
	}
	return string(out)
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
