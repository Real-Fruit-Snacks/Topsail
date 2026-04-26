package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

// captureStdio swaps the package-level stdio for buffers and restores them
// when the test ends.
func captureStdio(t *testing.T) (out, errBuf *bytes.Buffer) {
	t.Helper()
	origOut, origErr := ioutil.Stdout, ioutil.Stderr
	out = &bytes.Buffer{}
	errBuf = &bytes.Buffer{}
	ioutil.Stdout = out
	ioutil.Stderr = errBuf
	t.Cleanup(func() {
		ioutil.Stdout = origOut
		ioutil.Stderr = origErr
	})
	return out, errBuf
}

func registerEcho(t *testing.T) {
	t.Helper()
	applet.ResetForTesting()
	applet.Register(applet.Applet{
		Name:    "echo",
		Aliases: []string{"e"},
		Help:    "echo args",
		Usage:   "Usage: echo [args...]\n",
		Main: func(argv []string) int {
			for i, a := range argv[1:] {
				if i > 0 {
					_, _ = io.WriteString(ioutil.Stdout, " ")
				}
				_, _ = io.WriteString(ioutil.Stdout, a)
			}
			_, _ = io.WriteString(ioutil.Stdout, "\n")
			return 0
		},
	})
}

func TestRunNoArgs(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("expected top-level help; got %q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "--help"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "topsail") {
		t.Errorf(`expected "topsail" in help; got %q`, out.String())
	}
}

func TestRunHelpShortForm(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "-h"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("expected top-level help; got %q", out.String())
	}
}

func TestRunHelpApplet(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "--help", "echo"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "Usage: echo") {
		t.Errorf("expected applet usage; got %q", out.String())
	}
}

func TestRunHelpUnknownApplet(t *testing.T) {
	applet.ResetForTesting()
	_, errBuf := captureStdio(t)
	if rc := Run([]string{"topsail", "--help", "nope"}); rc != ExitAppletNotFound {
		t.Errorf("rc = %d; want %d", rc, ExitAppletNotFound)
	}
	if !strings.Contains(errBuf.String(), "unknown applet") {
		t.Errorf("expected 'unknown applet'; got %q", errBuf.String())
	}
}

func TestRunVersion(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "--version"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "topsail ") {
		t.Errorf(`expected "topsail " in version; got %q`, out.String())
	}
}

func TestRunVersionShortForm(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "-V"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "topsail ") {
		t.Errorf(`expected "topsail " in version; got %q`, out.String())
	}
}

func TestRunList(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "--list"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "echo") {
		t.Errorf("expected 'echo' in list; got %q", out.String())
	}
}

func TestRunListEmpty(t *testing.T) {
	applet.ResetForTesting()
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "--list"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "no applets") {
		t.Errorf("expected empty notice; got %q", out.String())
	}
}

func TestRunWrapperApplet(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail", "echo", "hello", "world"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if got := strings.TrimSpace(out.String()); got != "hello world" {
		t.Errorf("output = %q; want %q", got, "hello world")
	}
}

func TestRunWrapperUnknownApplet(t *testing.T) {
	applet.ResetForTesting()
	_, errBuf := captureStdio(t)
	if rc := Run([]string{"topsail", "nope"}); rc != ExitAppletNotFound {
		t.Errorf("rc = %d; want %d", rc, ExitAppletNotFound)
	}
	if !strings.Contains(errBuf.String(), "unknown applet") {
		t.Errorf("expected 'unknown applet'; got %q", errBuf.String())
	}
}

func TestRunUnknownOption(t *testing.T) {
	applet.ResetForTesting()
	_, errBuf := captureStdio(t)
	if rc := Run([]string{"topsail", "--bogus"}); rc != ExitUsageErr {
		t.Errorf("rc = %d; want %d", rc, ExitUsageErr)
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("expected 'unknown option'; got %q", errBuf.String())
	}
}

func TestRunMultiCallDispatch(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"echo", "from", "symlink"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if got := strings.TrimSpace(out.String()); got != "from symlink" {
		t.Errorf("output = %q; want %q", got, "from symlink")
	}
}

func TestRunMultiCallDispatchAlias(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"e", "via", "alias"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if got := strings.TrimSpace(out.String()); got != "via alias" {
		t.Errorf("output = %q; want %q", got, "via alias")
	}
}

func TestRunMultiCallStripsExe(t *testing.T) {
	// argv[0]="echo.exe" (no directory) is the cross-platform case: on
	// every OS, filepath.Base returns "echo.exe" verbatim, then our
	// dispatcher strips the .exe suffix and matches the registered
	// applet. Paths with backslashes are intentionally not exercised
	// here because they're only separators on Windows.
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"echo.exe", "windows", "path"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if got := strings.TrimSpace(out.String()); got != "windows path" {
		t.Errorf("output = %q; want %q", got, "windows path")
	}
}

func TestRunMultiCallHelpLongForm(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"echo", "--help"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "Usage: echo") {
		t.Errorf("expected applet usage; got %q", out.String())
	}
}

// In multi-call mode -h MUST reach the applet (e.g. `df -h` is
// human-readable, not help).
func TestRunMultiCallShortHelpFallsThrough(t *testing.T) {
	applet.ResetForTesting()
	var sawShortH bool
	applet.Register(applet.Applet{
		Name: "df",
		Help: "disk free",
		Main: func(argv []string) int {
			for _, a := range argv[1:] {
				if a == "-h" {
					sawShortH = true
				}
			}
			return 0
		},
	})
	captureStdio(t)
	if rc := Run([]string{"df", "-h"}); rc != 0 {
		t.Errorf("rc = %d; want 0", rc)
	}
	if !sawShortH {
		t.Error("applet did not receive -h; dispatcher swallowed it")
	}
}

func TestRunMultiCallDoubleDashStopsHelp(t *testing.T) {
	registerEcho(t)
	out, _ := captureStdio(t)
	if rc := Run([]string{"echo", "--", "--help"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want %d", rc, ExitSuccess)
	}
	// After --, --help is data and should be echoed.
	if got := strings.TrimSpace(out.String()); got != "-- --help" {
		t.Errorf("output = %q; want %q", got, "-- --help")
	}
}

func TestRunWrapperNameNeverMatchesApplet(t *testing.T) {
	// Even if some pathological applet were named "topsail", invoking
	// the binary by its wrapper name still goes through wrapper mode.
	applet.ResetForTesting()
	applet.Register(applet.Applet{
		Name: "topsail",
		Help: "shadow",
		Main: func([]string) int { return 99 },
	})
	out, _ := captureStdio(t)
	if rc := Run([]string{"topsail"}); rc != ExitSuccess {
		t.Errorf("rc = %d; want wrapper to take over with %d", rc, ExitSuccess)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("expected wrapper help; got %q", out.String())
	}
}

func TestRunEmptyArgs(t *testing.T) {
	applet.ResetForTesting()
	_, errBuf := captureStdio(t)
	if rc := Run(nil); rc != ExitUsageErr {
		t.Errorf("rc = %d; want %d", rc, ExitUsageErr)
	}
	if !strings.Contains(errBuf.String(), "empty argv") {
		t.Errorf("expected 'empty argv'; got %q", errBuf.String())
	}
}

func TestRunAppletPropagatesExitCode(t *testing.T) {
	applet.ResetForTesting()
	applet.Register(applet.Applet{
		Name: "fail",
		Help: "always fails",
		Main: func([]string) int { return 42 },
	})
	captureStdio(t)
	if rc := Run([]string{"topsail", "fail"}); rc != 42 {
		t.Errorf("rc = %d; want 42", rc)
	}
}
