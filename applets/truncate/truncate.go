// Package truncate implements the `truncate` applet: shrink or extend
// a file to a specified size.
package truncate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "truncate",
		Help:  "shrink or extend a file to a specified size",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: truncate -s [+|-|<|>|/|%]SIZE [OPTION]... FILE...
Shrink or extend the size of each FILE to SIZE.

A FILE that does not exist is created unless --no-create is given.
SIZE may be followed by one of K, M, G (powers of 1024) or KB, MB, GB
(powers of 1000), and may be prefixed with:
  +    extend by SIZE bytes
  -    shrink by SIZE bytes (clamped to 0)
  <    only if FILE is larger than SIZE
  >    only if FILE is smaller than SIZE
With no prefix, SIZE is the absolute new file size.

Options:
  -c, --no-create   do not create files that do not exist
  -s, --size=SIZE   the new size (or relative spec)
`

type sizeOp struct {
	op  byte // 0 (absolute), '+', '-', '<', '>'
	val int64
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		sizeSet  bool
		spec     sizeOp
		noCreate bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c", a == "--no-create":
			noCreate = true
			args = args[1:]
		case a == "-s", a == "--size":
			if len(args) < 2 {
				ioutil.Errf("truncate: option requires an argument -- 's'")
				return 2
			}
			s, err := parseSize(args[1])
			if err != nil {
				ioutil.Errf("truncate: invalid size: %v", err)
				return 2
			}
			spec = s
			sizeSet = true
			args = args[2:]
		case strings.HasPrefix(a, "--size="):
			s, err := parseSize(a[len("--size="):])
			if err != nil {
				ioutil.Errf("truncate: invalid size: %v", err)
				return 2
			}
			spec = s
			sizeSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-s") && len(a) > 2:
			s, err := parseSize(a[2:])
			if err != nil {
				ioutil.Errf("truncate: invalid size: %v", err)
				return 2
			}
			spec = s
			sizeSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			ioutil.Errf("truncate: unknown option: %s", a)
			return 2
		default:
			stop = true
		}
	}
	if !sizeSet {
		ioutil.Errf("truncate: -s SIZE is required")
		return 2
	}
	if len(args) == 0 {
		ioutil.Errf("truncate: missing file operand")
		return 2
	}

	rc := 0
	for _, name := range args {
		if err := truncateOne(name, spec, noCreate); err != nil {
			ioutil.Errf("truncate: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func truncateOne(name string, spec sizeOp, noCreate bool) error {
	var current int64
	st, err := os.Stat(name)
	switch {
	case err == nil:
		current = st.Size()
	case errors.Is(err, fs.ErrNotExist):
		if noCreate {
			return nil
		}
		f, ferr := os.Create(name) //nolint:gosec // path is user-supplied by design
		if ferr != nil {
			return ferr
		}
		_ = f.Close()
		current = 0
	default:
		return err
	}

	target, err := resolveTarget(current, spec)
	if err != nil {
		return err
	}
	if target == current {
		return nil
	}
	return os.Truncate(name, target)
}

func resolveTarget(current int64, spec sizeOp) (int64, error) {
	switch spec.op {
	case 0:
		return spec.val, nil
	case '+':
		return current + spec.val, nil
	case '-':
		v := current - spec.val
		if v < 0 {
			v = 0
		}
		return v, nil
	case '<':
		// Only if file is larger than SIZE; if smaller/equal, no-op.
		if current <= spec.val {
			return current, nil
		}
		return spec.val, nil
	case '>':
		// Only if file is smaller than SIZE; if larger/equal, no-op.
		if current >= spec.val {
			return current, nil
		}
		return spec.val, nil
	}
	return 0, fmt.Errorf("internal: unknown op %q", string(spec.op))
}

// parseSize accepts [+-<>]?DIGITS[K|M|G|T|KB|MB|GB|TB]. K-suffixes are
// powers of 1024; KB-suffixes are powers of 1000.
func parseSize(s string) (sizeOp, error) {
	if s == "" {
		return sizeOp{}, fmt.Errorf("empty size")
	}
	var op byte
	switch s[0] {
	case '+', '-', '<', '>':
		op = s[0]
		s = s[1:]
	}
	if s == "" {
		return sizeOp{}, fmt.Errorf("missing digits")
	}
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return sizeOp{}, fmt.Errorf("missing digits in %q", s)
	}
	digits := s[:end]
	suf := s[end:]
	mult, err := suffixMultiplier(suf)
	if err != nil {
		return sizeOp{}, err
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return sizeOp{}, err
	}
	if n < 0 {
		return sizeOp{}, fmt.Errorf("negative size")
	}
	if mult > 1 {
		n *= mult
	}
	return sizeOp{op: op, val: n}, nil
}

func suffixMultiplier(suf string) (int64, error) {
	switch strings.ToUpper(suf) {
	case "":
		return 1, nil
	case "K":
		return 1 << 10, nil
	case "M":
		return 1 << 20, nil
	case "G":
		return 1 << 30, nil
	case "T":
		return 1 << 40, nil
	case "KB":
		return 1000, nil
	case "MB":
		return 1000 * 1000, nil
	case "GB":
		return 1000 * 1000 * 1000, nil
	case "TB":
		return 1000 * 1000 * 1000 * 1000, nil
	}
	return 0, fmt.Errorf("unknown suffix %q", suf)
}
