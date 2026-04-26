package platform

import (
	"os"
	"testing"
)

func TestIsTerminalNil(t *testing.T) {
	if IsTerminal(nil) {
		t.Error("IsTerminal(nil) = true; want false")
	}
}

func TestTerminalSizeNil(t *testing.T) {
	w, h, ok := TerminalSize(nil)
	if w != 0 || h != 0 || ok {
		t.Errorf("TerminalSize(nil) = (%d, %d, %t); want (0, 0, false)", w, h, ok)
	}
}

// openDevNull returns an *os.File pointing at the platform null device, or
// skips the test if it cannot be opened.
func openDevNull(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Skipf("cannot open %s: %v", os.DevNull, err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestIsTerminalNonTerminalFile(t *testing.T) {
	f := openDevNull(t)
	if IsTerminal(f) {
		t.Errorf("IsTerminal(%s) = true; want false", os.DevNull)
	}
}

func TestTerminalSizeNonTerminalFile(t *testing.T) {
	f := openDevNull(t)
	w, h, ok := TerminalSize(f)
	if ok {
		t.Errorf("TerminalSize(%s) = (%d, %d, true); want ok=false", os.DevNull, w, h)
	}
}

// IsTerminal/TerminalSize on a real terminal can't be portably asserted in
// CI: the test runner's stdin/stdout aren't ttys under `go test`. The nil
// and non-tty paths above are what we can portably check; real tty
// behavior is exercised end-to-end by applet integration tests in later
// waves.

func TestUserNameAndGroupNameReturnSomething(t *testing.T) {
	// Sanity: non-empty input always produces non-empty output.
	if got := UserName("0"); got == "" {
		t.Error(`UserName("0") = ""; want non-empty`)
	}
	if got := GroupName("0"); got == "" {
		t.Error(`GroupName("0") = ""; want non-empty`)
	}
}
