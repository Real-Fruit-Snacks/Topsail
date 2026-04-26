// Package touch implements the `touch` applet: change file timestamps.
package touch

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "touch",
		Help:  "change file timestamps",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: touch [OPTION]... FILE...
Update access and modification times of each FILE to the current time.
A FILE that does not exist is created empty unless -c is given.

Options:
  -a                       change only the access time
  -m                       change only the modification time
  -c, --no-create          do not create files
  -r, --reference=FILE     use FILE's times instead of current time
  -d, --date=STRING        use STRING (RFC 3339 / common formats) for time
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		noCreate, onlyAccess, onlyMod, refSet, dateSet bool
		ref, dateStr                                   string
	)

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-a":
			onlyAccess = true
			args = args[1:]
		case a == "-m":
			onlyMod = true
			args = args[1:]
		case a == "-c", a == "--no-create":
			noCreate = true
			args = args[1:]
		case a == "-r":
			if len(args) < 2 {
				ioutil.Errf("touch: option requires an argument -- 'r'")
				return 2
			}
			ref = args[1]
			refSet = true
			args = args[2:]
		case strings.HasPrefix(a, "--reference="):
			ref = a[len("--reference="):]
			refSet = true
			args = args[1:]
		case a == "-d":
			if len(args) < 2 {
				ioutil.Errf("touch: option requires an argument -- 'd'")
				return 2
			}
			dateStr = args[1]
			dateSet = true
			args = args[2:]
		case strings.HasPrefix(a, "--date="):
			dateStr = a[len("--date="):]
			dateSet = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1:
			for _, c := range a[1:] {
				switch c {
				case 'a':
					onlyAccess = true
				case 'm':
					onlyMod = true
				case 'c':
					noCreate = true
				default:
					ioutil.Errf("touch: invalid option -- '%c'", c)
					return 2
				}
			}
			args = args[1:]
		default:
			stop = true
		}
	}

	if len(args) == 0 {
		ioutil.Errf("touch: missing operand")
		return 2
	}

	now := time.Now()
	atime, mtime := now, now
	if refSet {
		info, err := os.Stat(ref)
		if err != nil {
			ioutil.Errf("touch: %s: %v", ref, err)
			return 1
		}
		mtime = info.ModTime()
		atime = mtime
	}
	if dateSet {
		t, err := parseDate(dateStr)
		if err != nil {
			ioutil.Errf("touch: invalid date: %s", dateStr)
			return 1
		}
		atime, mtime = t, t
	}

	rc := 0
	for _, name := range args {
		if err := touchOne(name, atime, mtime, onlyAccess, onlyMod, noCreate); err != nil {
			ioutil.Errf("touch: %s: %v", name, err)
			rc = 1
		}
	}
	return rc
}

func touchOne(name string, atime, mtime time.Time, onlyAccess, onlyMod, noCreate bool) error {
	info, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		if noCreate {
			return nil
		}
		var f *os.File
		f, err = os.Create(name) //nolint:gosec // touch creates user-named files by design
		if err != nil {
			return err
		}
		_ = f.Close()
		info, err = os.Stat(name)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// If only one of -a / -m was given, leave the other timestamp alone
	// (best-effort: Go's stdlib doesn't portably expose atime, so we
	// fall back to existing ModTime as a safe approximation).
	if onlyAccess && !onlyMod {
		mtime = info.ModTime()
	}
	if onlyMod && !onlyAccess {
		atime = info.ModTime()
	}
	return os.Chtimes(name, atime, mtime)
}

// parseDate accepts a small set of common date formats. POSIX-style
// "MMDDhhmm[CC]YY[.ss]" arguments are deferred to a later wave.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unrecognized format")
}
