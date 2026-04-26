// Package filemode parses file mode strings — both octal (e.g. "0755",
// "644") and POSIX symbolic forms (e.g. "u+x", "go-w", "a=rwx,u+s") —
// into an os.FileMode.
//
// Symbolic parsing requires a "current" mode so operations like u+x
// can layer on top of existing bits. Callers that don't have a real
// current mode (mkdir creating a fresh directory, for example) should
// pass 0o777 to start from a "fully permissive" baseline; the umask
// will trim it back at creation time.
package filemode

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Bit masks shared across the parser. Keep them named so the symbolic
// arithmetic below stays legible.
const (
	allRead    = 0o444 // r across u, g, o
	allWrite   = 0o222 // w across u, g, o
	allExec    = 0o111 // x across u, g, o
	bitSetuid  = 0o4000
	bitSetgid  = 0o2000
	bitSticky  = 0o1000
	whoUser    = 0o4700 // owner perm bits + setuid
	whoGroup   = 0o2070 // group perm bits + setgid
	whoOther   = 0o1007 // other perm bits + sticky
	whoAll     = 0o7777
	permLowAll = 0o7777 // every settable bit
)

// Parse interprets s as either an octal mode (digits only) or a POSIX
// symbolic mode applied on top of current. Returns the resulting mode.
func Parse(s string, current os.FileMode) (os.FileMode, error) {
	if s == "" {
		return 0, fmt.Errorf("empty mode")
	}
	if isOctal(s) {
		n, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid octal mode %q: %w", s, err)
		}
		return os.FileMode(n), nil
	}
	return parseSymbolic(s, current)
}

func isOctal(s string) bool {
	// Treat any string consisting only of digits 0-7 as octal. Mixed
	// strings or anything containing letters routes to the symbolic parser.
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '7' {
			return false
		}
	}
	return true
}

// parseSymbolic walks one or more comma-separated clauses, each of the
// form `[ugoa]*[+-=][rwxXst]*` (perms list may be empty for `=` to clear).
func parseSymbolic(s string, current os.FileMode) (os.FileMode, error) {
	mode := current & permLowAll
	for _, clause := range strings.Split(s, ",") {
		if clause == "" {
			return 0, fmt.Errorf("empty clause in mode %q", s)
		}
		next, err := applyClause(clause, mode, current)
		if err != nil {
			return 0, err
		}
		mode = next
	}
	return mode, nil
}

func applyClause(clause string, mode, original os.FileMode) (os.FileMode, error) {
	i := 0
	var who uint32
	for i < len(clause) {
		switch clause[i] {
		case 'u':
			who |= whoUser
		case 'g':
			who |= whoGroup
		case 'o':
			who |= whoOther
		case 'a':
			who |= whoAll
		default:
			goto op
		}
		i++
	}
op:
	if i >= len(clause) {
		return 0, fmt.Errorf("missing operator in %q", clause)
	}
	if who == 0 {
		// POSIX: empty 'who' is treated as 'a' (umask trimming is the
		// caller's job; we don't consult the process umask here).
		who = whoAll
	}
	op := clause[i]
	switch op {
	case '+', '-', '=':
	default:
		return 0, fmt.Errorf("invalid operator %q in %q", string(op), clause)
	}
	i++

	// Optional source mode: copy bits from another who (e.g. g=u).
	if i < len(clause) {
		switch clause[i] {
		case 'u', 'g', 'o':
			src := clause[i]
			bits := bitsFromWho(src, mode)
			return commit(mode, who, bits, op), nil
		}
	}

	var bits uint32
	for ; i < len(clause); i++ {
		switch clause[i] {
		case 'r':
			bits |= allRead
		case 'w':
			bits |= allWrite
		case 'x':
			bits |= allExec
		case 's':
			bits |= bitSetuid | bitSetgid
		case 't':
			bits |= bitSticky
		case 'X':
			// X = set x only if the file is a directory or already has
			// any execute bit set somewhere. Evaluated against the
			// original mode, not the in-progress one, so multi-clause
			// scripts don't shift their meaning mid-flight.
			if original.IsDir() || (uint32(original)&allExec) != 0 {
				bits |= allExec
			}
		default:
			return 0, fmt.Errorf("invalid permission %q in %q", string(clause[i]), clause)
		}
	}
	return commit(mode, who, bits, op), nil
}

// bitsFromWho returns the permission bits of mode projected from the
// referenced 'who' (u, g, or o) onto the canonical "all whos" layout.
// E.g. mode is rwxr-x---, src='u' -> rwx pattern (0o111 + 0o222 + 0o444 = 0o777
// actually no — extract owner's r/w/x and re-replicate across all whos).
func bitsFromWho(src byte, mode os.FileMode) uint32 {
	m := uint32(mode)
	var r, w, x uint32
	switch src {
	case 'u':
		if m&0o400 != 0 {
			r = allRead
		}
		if m&0o200 != 0 {
			w = allWrite
		}
		if m&0o100 != 0 {
			x = allExec
		}
	case 'g':
		if m&0o040 != 0 {
			r = allRead
		}
		if m&0o020 != 0 {
			w = allWrite
		}
		if m&0o010 != 0 {
			x = allExec
		}
	case 'o':
		if m&0o004 != 0 {
			r = allRead
		}
		if m&0o002 != 0 {
			w = allWrite
		}
		if m&0o001 != 0 {
			x = allExec
		}
	}
	return r | w | x
}

func commit(mode os.FileMode, who, bits uint32, op byte) os.FileMode {
	change := bits & who
	switch op {
	case '+':
		return mode | os.FileMode(change)
	case '-':
		return mode &^ os.FileMode(change)
	case '=':
		// Clear all bits in 'who', then set the change bits.
		return (mode &^ os.FileMode(who)) | os.FileMode(change)
	}
	return mode
}
