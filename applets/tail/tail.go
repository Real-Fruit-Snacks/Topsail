// Package tail implements the `tail` applet: output the last part of files.
//
// The follow mode (-f) polls each file's size on a short interval and
// emits any newly-appended bytes. Truncation (size shrinks) is handled
// by re-seeking to the start of the file. stdin and pipes cannot be
// followed — they are skipped with a diagnostic.
package tail

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "tail",
		Help:  "output the last part of files",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: tail [OPTION]... [FILE]...
Print the last 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header.
With no FILE, or when FILE is -, read standard input.

Options:
  -n N, --lines=N    print the last N lines instead of 10 (or +N for "starting at line N")
  -c N, --bytes=N    print the last N bytes
  -q, --quiet        never print headers
  -v, --verbose      always print headers
  -f, --follow       output appended data as the file grows; exits on SIGINT
  -s SECONDS, --sleep-interval=SECONDS
                     time to wait between iterations of the follow loop (default 1)
`

// Test seams: tests override these to drive follow mode deterministically
// without waiting for real wall-clock seconds or real signals.
var (
	followInterval = 1 * time.Second
	newFollowCtx   = func() (context.Context, context.CancelFunc) {
		return signal.NotifyContext(context.Background(), os.Interrupt)
	}
)

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	nLines := 10
	nBytes := 0
	var byBytes, fromStart, quiet, verbose, follow bool

	parseN := func(s string) (n int, fs, ok bool) {
		if strings.HasPrefix(s, "+") {
			fs = true
			s = s[1:]
		} else if strings.HasPrefix(s, "-") {
			s = s[1:]
		}
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			return 0, false, false
		}
		return v, fs, true
	}

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "-n":
			if len(args) < 2 {
				ioutil.Errf("tail: option requires an argument -- 'n'")
				return 2
			}
			n, fs, ok := parseN(args[1])
			if !ok {
				ioutil.Errf("tail: invalid number of lines: %s", args[1])
				return 2
			}
			nLines = n
			fromStart = fs
			byBytes = false
			args = args[2:]
		case strings.HasPrefix(a, "--lines="):
			n, fs, ok := parseN(a[len("--lines="):])
			if !ok {
				ioutil.Errf("tail: invalid number of lines: %s", a)
				return 2
			}
			nLines = n
			fromStart = fs
			byBytes = false
			args = args[1:]
		case a == "-c":
			if len(args) < 2 {
				ioutil.Errf("tail: option requires an argument -- 'c'")
				return 2
			}
			n, fs, ok := parseN(args[1])
			if !ok {
				ioutil.Errf("tail: invalid number of bytes: %s", args[1])
				return 2
			}
			nBytes = n
			fromStart = fs
			byBytes = true
			args = args[2:]
		case strings.HasPrefix(a, "--bytes="):
			n, fs, ok := parseN(a[len("--bytes="):])
			if !ok {
				ioutil.Errf("tail: invalid number of bytes: %s", a)
				return 2
			}
			nBytes = n
			fromStart = fs
			byBytes = true
			args = args[1:]
		case a == "-q", a == "--quiet", a == "--silent":
			quiet = true
			args = args[1:]
		case a == "-v", a == "--verbose":
			verbose = true
			args = args[1:]
		case a == "-f", a == "--follow":
			follow = true
			args = args[1:]
		case a == "-s":
			if len(args) < 2 {
				ioutil.Errf("tail: option requires an argument -- 's'")
				return 2
			}
			d, err := parseSleep(args[1])
			if err != nil {
				ioutil.Errf("tail: invalid sleep interval: %s", args[1])
				return 2
			}
			followInterval = d
			args = args[2:]
		case strings.HasPrefix(a, "--sleep-interval="):
			d, err := parseSleep(a[len("--sleep-interval="):])
			if err != nil {
				ioutil.Errf("tail: invalid sleep interval: %s", a)
				return 2
			}
			followInterval = d
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Allow -<digits> shorthand: -5 == -n 5
			if n, err := strconv.Atoi(a[1:]); err == nil && n >= 0 {
				nLines = n
				byBytes = false
				args = args[1:]
				continue
			}
			ioutil.Errf("tail: invalid option: %s", a)
			return 2
		default:
			stop = true
		}
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}
	showHeader := (len(files) > 1 && !quiet) || verbose

	rc := 0
	for i, name := range files {
		if showHeader {
			if i > 0 {
				_, _ = fmt.Fprintln(ioutil.Stdout)
			}
			_, _ = fmt.Fprintf(ioutil.Stdout, "==> %s <==\n", labelFor(name))
		}
		if err := tailOne(name, nLines, nBytes, byBytes, fromStart); err != nil {
			ioutil.Errf("tail: %s: %v", name, err)
			rc = 1
		}
	}

	if follow {
		realFiles := make([]string, 0, len(files))
		for _, f := range files {
			if f == "-" {
				ioutil.Errf("tail: warning: --follow is ineffective on standard input")
				continue
			}
			realFiles = append(realFiles, f)
		}
		if len(realFiles) > 0 {
			ctx, cancel := newFollowCtx()
			defer cancel()
			if err := followFiles(ctx, realFiles, len(realFiles) > 1 || verbose); err != nil {
				ioutil.Errf("tail: %v", err)
				rc = 1
			}
		}
	}

	return rc
}

func tailOne(name string, nLines, nBytes int, byBytes, fromStart bool) error {
	var r io.Reader
	if name == "-" {
		r = ioutil.Stdin
	} else {
		f, err := os.Open(name) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	if byBytes {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if fromStart {
			start := nBytes - 1
			if start < 0 {
				start = 0
			}
			if start > len(data) {
				return nil
			}
			_, err = ioutil.Stdout.Write(data[start:])
			return err
		}
		if nBytes >= len(data) {
			_, err = ioutil.Stdout.Write(data)
		} else {
			_, err = ioutil.Stdout.Write(data[len(data)-nBytes:])
		}
		return err
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	if fromStart {
		idx := 0
		for sc.Scan() {
			idx++
			if idx >= nLines {
				if _, err := io.WriteString(ioutil.Stdout, sc.Text()+"\n"); err != nil {
					return err
				}
			}
		}
		return sc.Err()
	}

	// Ring buffer of the last nLines lines.
	buf := make([]string, 0, nLines)
	for sc.Scan() {
		if len(buf) < nLines {
			buf = append(buf, sc.Text())
		} else {
			copy(buf, buf[1:])
			buf[len(buf)-1] = sc.Text()
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	for _, line := range buf {
		if _, err := io.WriteString(ioutil.Stdout, line+"\n"); err != nil {
			return err
		}
	}
	return nil
}

// followed tracks one file across the follow loop's iterations: open
// handle, last observed size for truncation detection, and the canonical
// name we report in headers.
type followed struct {
	name string
	f    *os.File
	size int64
}

// followFiles opens each name, seeks to end, and polls every
// followInterval for new bytes. It returns when ctx is canceled or an
// unrecoverable I/O error occurs.
func followFiles(ctx context.Context, names []string, showHeaders bool) error {
	fs := make([]*followed, 0, len(names))
	defer func() {
		for _, fl := range fs {
			_ = fl.f.Close()
		}
	}()
	for _, name := range names {
		f, err := os.Open(name) //nolint:gosec // user-supplied path
		if err != nil {
			ioutil.Errf("tail: cannot follow %s: %v", name, err)
			continue
		}
		st, err := f.Stat()
		if err != nil {
			_ = f.Close()
			ioutil.Errf("tail: cannot stat %s: %v", name, err)
			continue
		}
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			_ = f.Close()
			return err
		}
		fs = append(fs, &followed{name: name, f: f, size: st.Size()})
	}
	if len(fs) == 0 {
		return nil
	}

	var last string
	buf := make([]byte, 4096)
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		progress := false
		for _, fl := range fs {
			st, err := os.Stat(fl.name)
			if err != nil {
				// File temporarily disappeared (rotate, rename) — skip
				// this round; we'll catch the new one if it returns.
				continue
			}
			if st.Size() < fl.size {
				if _, err := fl.f.Seek(0, io.SeekStart); err != nil {
					return err
				}
				fl.size = 0
				ioutil.Errf("tail: %s: file truncated", fl.name)
			}
			for {
				n, err := fl.f.Read(buf)
				if n > 0 {
					if showHeaders && fl.name != last {
						if last != "" {
							_, _ = fmt.Fprintln(ioutil.Stdout)
						}
						_, _ = fmt.Fprintf(ioutil.Stdout, "==> %s <==\n", fl.name)
						last = fl.name
					}
					_, _ = ioutil.Stdout.Write(buf[:n])
					fl.size += int64(n)
					progress = true
				}
				if err == io.EOF || n == 0 {
					break
				}
				if err != nil {
					return err
				}
			}
		}
		if !progress {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(followInterval):
			}
		}
	}
}

func labelFor(name string) string {
	if name == "-" {
		return "standard input"
	}
	return name
}

// parseSleep accepts a bare number (seconds) or a Go duration string.
// "1" -> 1s; "500ms" -> 500ms; "2.5" -> 2.5s.
func parseSleep(s string) (time.Duration, error) {
	if f, err := strconv.ParseFloat(s, 64); err == nil && f >= 0 {
		return time.Duration(f * float64(time.Second)), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("negative interval")
	}
	return d, nil
}
