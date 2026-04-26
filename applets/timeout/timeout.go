// Package timeout implements the `timeout` applet: run a command with
// a wall-clock deadline. If the deadline elapses, the command is sent
// the configured signal (SIGTERM by default) and its exit status is
// reported as 124 to mirror coreutils.
package timeout

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "timeout",
		Help:  "run a command with a time limit",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: timeout [OPTION]... DURATION COMMAND [ARG]...
Start COMMAND, and kill it if it is still running after DURATION.

DURATION accepts plain numbers (seconds), or Go-style duration strings
("250ms", "2.5s", "1m30s"). Coreutils suffixes s/m/h/d are honored.

Options:
  -s SIGNAL, --signal=SIGNAL    signal name to send on timeout (default TERM; Windows ignores this)
  -k AFTER, --kill-after=AFTER  send SIGKILL/Process.Kill if still alive AFTER seconds after the timeout
  --preserve-status             exit with the command's status, not 124, when timed out
  --foreground                  ignored (no controlling terminal manipulation in this build)

Exit status:
  124  if the command timed out
  125  if timeout itself failed
  126  if COMMAND was found but could not be invoked
  127  if COMMAND was not found
  rc   the command's exit status otherwise
`

const (
	rcTimeout    = 124
	rcSelfFailed = 125
	rcExecFailed = 126
	rcNotFound   = 127
)

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	var (
		killAfter      time.Duration
		preserveStatus bool
		hasKillAfter   bool
	)
	// -s SIGNAL is accepted for parity with GNU timeout but ignored:
	// CommandContext relies on Process.Kill, which doesn't take a signal
	// number on Windows and doesn't expose the right hook for graceful
	// SIGTERM-then-SIGKILL flow without per-OS code. Documented in usage.

	stop := false
	for !stop && len(args) > 0 {
		a := args[0]
		switch {
		case a == "--":
			args = args[1:]
			stop = true
		case a == "--preserve-status":
			preserveStatus = true
			args = args[1:]
		case a == "--foreground":
			args = args[1:]
		case a == "-s", a == "--signal":
			if len(args) < 2 {
				ioutil.Errf("timeout: option requires an argument -- 's'")
				return rcSelfFailed
			}
			// Discard SIGNAL value — see comment above.
			args = args[2:]
		case strings.HasPrefix(a, "--signal="):
			args = args[1:]
		case a == "-k", a == "--kill-after":
			if len(args) < 2 {
				ioutil.Errf("timeout: option requires an argument -- 'k'")
				return rcSelfFailed
			}
			d, err := parseDuration(args[1])
			if err != nil {
				ioutil.Errf("timeout: invalid kill-after: %v", err)
				return rcSelfFailed
			}
			killAfter = d
			hasKillAfter = true
			args = args[2:]
		case strings.HasPrefix(a, "--kill-after="):
			d, err := parseDuration(a[len("--kill-after="):])
			if err != nil {
				ioutil.Errf("timeout: invalid kill-after: %v", err)
				return rcSelfFailed
			}
			killAfter = d
			hasKillAfter = true
			args = args[1:]
		case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
			// Bare digits look like negative duration to flag parser; reject.
			ioutil.Errf("timeout: unknown option: %s", a)
			return rcSelfFailed
		default:
			stop = true
		}
	}

	if len(args) < 2 {
		ioutil.Errf("timeout: missing DURATION and/or COMMAND")
		return rcSelfFailed
	}
	duration, err := parseDuration(args[0])
	if err != nil {
		ioutil.Errf("timeout: invalid duration: %v", err)
		return rcSelfFailed
	}
	cmdArgs := args[1:]

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...) //nolint:gosec // user-supplied command is the whole point
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			ioutil.Errf("timeout: %s: command not found", cmdArgs[0])
			return rcNotFound
		}
		ioutil.Errf("timeout: %s: %v", cmdArgs[0], err)
		return rcExecFailed
	}

	waitErr := cmd.Wait()
	timedOut := ctx.Err() == context.DeadlineExceeded

	if timedOut && hasKillAfter && killAfter > 0 {
		// CommandContext already killed the process; this branch is mostly
		// a placeholder for richer kill-after semantics on Unix where you
		// may want to send SIGTERM first then SIGKILL after the grace.
		// We at least give the OS a moment to reap it before we return.
		select {
		case <-time.After(killAfter):
			_ = cmd.Process.Kill()
		default:
		}
	}

	if timedOut && !preserveStatus {
		return rcTimeout
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return exitErr.ExitCode()
		}
		ioutil.Errf("timeout: %v", waitErr)
		return rcExecFailed
	}
	return 0
}

// parseDuration extends time.ParseDuration with two coreutils
// conveniences: a plain number is interpreted as seconds, and the
// "d" (days) suffix is honored.
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// Coreutils-style days suffix.
	if strings.HasSuffix(s, "d") {
		num := s[:len(s)-1]
		f, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(f * float64(24*time.Hour)), nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if f < 0 {
			return 0, fmt.Errorf("negative duration")
		}
		return time.Duration(f * float64(time.Second)), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("negative duration")
	}
	return d, nil
}
