// Package hashing provides shared logic for the md5sum / sha256sum /
// sha512sum applets. Each applet is a thin wrapper that picks a hash
// constructor and a default name, then delegates to Run.
package hashing

import (
	"crypto/md5" //nolint:gosec // md5 is required for md5sum compatibility
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

// New returns a hash constructor by algorithm name.
func New(algo string) func() hash.Hash {
	switch algo {
	case "md5":
		return md5.New
	case "sha256":
		return sha256.New
	case "sha512":
		return sha512.New
	}
	panic("hashing: unknown algorithm " + algo)
}

// Run is the shared Main implementation. cmdName is the applet name
// (used for diagnostics); algo is one of "md5", "sha256", "sha512".
func Run(cmdName, algo string, argv []string) int {
	args := argv[1:]
	var (
		check, binary, text, quiet bool
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-c", a == "--check":
			check = true
			args = args[1:]
		case a == "-b", a == "--binary":
			binary = true
			args = args[1:]
		case a == "-t", a == "--text":
			text = true
			args = args[1:]
		case a == "--quiet":
			quiet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			for _, c := range a[1:] {
				switch c {
				case 'c':
					check = true
				case 'b':
					binary = true
				case 't':
					text = true
				default:
					ioutil.Errf("%s: invalid option -- '%c'", cmdName, c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	_ = binary
	_ = text

	hashCtor := New(algo)

	if check {
		return runCheck(cmdName, hashCtor, args, quiet)
	}

	if len(args) == 0 {
		args = []string{"-"}
	}
	rc := 0
	for _, name := range args {
		sum, err := hashOne(name, hashCtor)
		if err != nil {
			ioutil.Errf("%s: %s: %v", cmdName, name, err)
			rc = 1
			continue
		}
		_, _ = fmt.Fprintf(ioutil.Stdout, "%s  %s\n", sum, name)
	}
	return rc
}

func hashOne(name string, ctor func() hash.Hash) (string, error) {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return "", err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	h := ctor()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// runCheck verifies sums listed in input files (or stdin). Each line
// must look like "HEX  FILENAME" (two spaces or "  " or " *").
func runCheck(cmdName string, ctor func() hash.Hash, files []string, quiet bool) int {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var ok, fail int
	for _, listFile := range files {
		data, err := readListFile(listFile)
		if err != nil {
			ioutil.Errf("%s: %s: %v", cmdName, listFile, err)
			continue
		}
		for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			if line == "" {
				continue
			}
			expected, name, ok2 := parseCheckLine(line)
			if !ok2 {
				ioutil.Errf("%s: bad check line: %q", cmdName, line)
				fail++
				continue
			}
			actual, err := hashOne(name, ctor)
			if err != nil {
				_, _ = fmt.Fprintf(ioutil.Stdout, "%s: FAILED open or read\n", name)
				fail++
				continue
			}
			if actual != expected {
				_, _ = fmt.Fprintf(ioutil.Stdout, "%s: FAILED\n", name)
				fail++
				continue
			}
			ok++
			if !quiet {
				_, _ = fmt.Fprintf(ioutil.Stdout, "%s: OK\n", name)
			}
		}
	}
	if fail > 0 {
		ioutil.Errf("%s: WARNING: %d computed checksum(s) did NOT match", cmdName, fail)
		return 1
	}
	return 0
}

// readListFile reads a checksum listing from disk or stdin. It is split
// out so the deferred Close stays out of the per-file loop in runCheck.
func readListFile(name string) ([]byte, error) {
	if name == "-" {
		return io.ReadAll(ioutil.Stdin)
	}
	f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}

func parseCheckLine(line string) (expected, name string, ok bool) {
	// Format: "HEX  NAME" (two spaces) or "HEX *NAME" (binary mark).
	for sep := range []string{"  ", " *"} {
		_ = sep
	}
	for _, sep := range []string{"  ", " *"} {
		if i := strings.Index(line, sep); i > 0 {
			return line[:i], line[i+2:], true
		}
	}
	return "", "", false
}
