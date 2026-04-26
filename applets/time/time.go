// Package time implements the `time` applet: run a command and report
// how long it took.
//
// Coreutils' /usr/bin/time supports a rich format string and reports
// user/system CPU time pulled from getrusage(2). For cross-platform
// portability this build reports wall-clock real time always; user
// and system CPU time are reported when Go can extract them from
// cmd.ProcessState (which works on Unix; on Windows they're zeros).
package time

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	stdtime "time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "time",
		Help:  "run a command and report how long it took",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: time COMMAND [ARG]...
Run COMMAND and write a summary to stderr after it exits:

  real    Ns
  user    Ns
  sys     Ns

real is wall-clock elapsed time. user/sys are pulled from the
ProcessState if the OS supports them (Unix); on Windows they are
reported as 0s.

The exit status is the command's exit status, or 127 if the
command was not found and 126 if it could not be invoked.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if len(args) == 0 {
		ioutil.Errf("time: missing COMMAND")
		return 2
	}
	if args[0] == "--" {
		args = args[1:]
		if len(args) == 0 {
			ioutil.Errf("time: missing COMMAND")
			return 2
		}
	}

	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // user-supplied command is the whole point
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := stdtime.Now()
	err := cmd.Run()
	elapsed := stdtime.Since(start)

	rc := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			rc = exitErr.ExitCode()
		} else if errors.Is(err, exec.ErrNotFound) {
			ioutil.Errf("time: %s: command not found", args[0])
			return 127
		} else {
			ioutil.Errf("time: %s: %v", args[0], err)
			return 126
		}
	}

	var userT, sysT stdtime.Duration
	if cmd.ProcessState != nil {
		userT = cmd.ProcessState.UserTime()
		sysT = cmd.ProcessState.SystemTime()
	}

	_, _ = fmt.Fprintf(ioutil.Stderr, "\nreal\t%s\nuser\t%s\nsys\t%s\n",
		formatDuration(elapsed),
		formatDuration(userT),
		formatDuration(sysT),
	)
	return rc
}

// formatDuration mirrors `time(1)`'s "0m1.234s" style.
func formatDuration(d stdtime.Duration) string {
	if d < 0 {
		d = 0
	}
	mins := int64(d / stdtime.Minute)
	rem := d - stdtime.Duration(mins)*stdtime.Minute
	secs := float64(rem) / float64(stdtime.Second)
	return fmt.Sprintf("%dm%.3fs", mins, secs)
}
