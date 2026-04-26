// Package find implements a minimal `find` applet: walk a directory tree
// and print or filter entries by name, type, and depth.
//
// More complex predicates (-mtime, -size, -exec, -prune, etc.) are
// deferred to a later wave. This Wave 3 build covers the common
// scripting use of `find DIR -name 'glob' -type f`.
package find

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "find",
		Help:  "walk a directory tree and print matching paths",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: find [PATH]... [EXPRESSION]
Walk each PATH (default: '.') and print entries matching EXPRESSION.

Supported predicates:
  -name PATTERN       basename matches glob PATTERN
  -iname PATTERN      same, case-insensitive
  -type [fdl]         f=regular file, d=directory, l=symlink
  -maxdepth N         descend at most N directory levels
  -mindepth N         apply tests only at depth >= N
  -print              print path (default action)
  -print0             print path NUL-terminated

Predicates are AND-ed together. -o for OR is not yet supported.
`

type predicate struct {
	name    string // "name", "iname", "type", ...
	pattern string
}

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var paths []string
	var preds []predicate
	maxDepth := -1
	minDepth := 0
	print0 := false

	// Paths come first, until we hit something starting with '-' or a
	// recognized predicate keyword.
	for len(args) > 0 {
		a := args[0]
		if strings.HasPrefix(a, "-") || isPredicateKeyword(a) {
			break
		}
		paths = append(paths, a)
		args = args[1:]
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for len(args) > 0 {
		a := args[0]
		switch a {
		case "-name", "-iname":
			if len(args) < 2 {
				ioutil.Errf("find: %s: missing argument", a)
				return 2
			}
			preds = append(preds, predicate{name: strings.TrimPrefix(a, "-"), pattern: args[1]})
			args = args[2:]
		case "-type":
			if len(args) < 2 {
				ioutil.Errf("find: -type: missing argument")
				return 2
			}
			preds = append(preds, predicate{name: "type", pattern: args[1]})
			args = args[2:]
		case "-maxdepth":
			if len(args) < 2 {
				ioutil.Errf("find: -maxdepth: missing argument")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("find: -maxdepth: invalid value: %s", args[1])
				return 2
			}
			maxDepth = n
			args = args[2:]
		case "-mindepth":
			if len(args) < 2 {
				ioutil.Errf("find: -mindepth: missing argument")
				return 2
			}
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				ioutil.Errf("find: -mindepth: invalid value: %s", args[1])
				return 2
			}
			minDepth = n
			args = args[2:]
		case "-print":
			args = args[1:]
		case "-print0":
			print0 = true
			args = args[1:]
		default:
			ioutil.Errf("find: unknown predicate: %s", a)
			return 2
		}
	}

	rc := 0
	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				ioutil.Errf("find: %s: %v", path, err)
				rc = 1
				return nil
			}
			depth := strings.Count(strings.TrimPrefix(path, root), string(filepath.Separator))
			if maxDepth >= 0 && depth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if depth < minDepth {
				return nil
			}
			if !match(path, d, preds) {
				return nil
			}
			term := byte('\n')
			if print0 {
				term = 0
			}
			_, _ = fmt.Fprintf(ioutil.Stdout, "%s%c", path, term)
			return nil
		})
		if err != nil {
			ioutil.Errf("find: %s: %v", root, err)
			rc = 1
		}
	}
	return rc
}

func isPredicateKeyword(s string) bool {
	switch s {
	case "-name", "-iname", "-type", "-maxdepth", "-mindepth", "-print", "-print0":
		return true
	}
	return false
}

func match(path string, _ fs.DirEntry, preds []predicate) bool {
	for _, p := range preds {
		switch p.name {
		case "name":
			ok, _ := filepath.Match(p.pattern, filepath.Base(path))
			if !ok {
				return false
			}
		case "iname":
			ok, _ := filepath.Match(strings.ToLower(p.pattern), strings.ToLower(filepath.Base(path)))
			if !ok {
				return false
			}
		case "type":
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			switch p.pattern {
			case "f":
				if !info.Mode().IsRegular() {
					return false
				}
			case "d":
				if !info.IsDir() {
					return false
				}
			case "l":
				if info.Mode()&os.ModeSymlink == 0 {
					return false
				}
			default:
				return false
			}
		}
	}
	return true
}
