// Package integration_test exercises the topsail binary as a subprocess.
//
// Unit tests in applets/* call Main() directly; this package builds the
// real binary and dispatches through argv[0] so we catch regressions in
// the multi-call wiring, the registry, the wrapper-mode flag handling,
// and the .exe-suffix stripping on Windows. Symlink-based dispatch is
// covered on Unix; Windows uses copy-based dispatch since unprivileged
// symlinks are not guaranteed.
package integration_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// binPath is the absolute path to the built topsail binary, populated
// by TestMain. Each test calls run() which invokes it with argv set
// directly rather than going through a shell.
var binPath string

func TestMain(m *testing.M) {
	// Run() wraps the test execution so we can `defer` cleanup and still
	// surface the test exit code. os.Exit jumps over deferred functions,
	// so we keep that call in the outer main shell only.
	os.Exit(runMain(m))
}

func runMain(m *testing.M) (code int) {
	dir, err := os.MkdirTemp("", "topsail-integration-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "tempdir:", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()

	name := "topsail"
	if runtime.GOOS == "windows" {
		name = "topsail.exe"
	}
	binPath = filepath.Join(dir, name)

	// `go build` resolves ./cmd/topsail relative to the test binary's
	// working directory, which is this package's source dir. Walk up
	// to the module root so the build target resolves cleanly.
	repoRoot, err := findModuleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "find module root:", err)
		return 1
	}
	cmd := exec.Command("go", "build", "-trimpath", "-o", binPath, "./cmd/topsail")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, "go build failed:", err, "\n", string(out))
		return 1
	}
	return m.Run()
}

// findModuleRoot walks parent directories looking for go.mod.
func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above %s", wd)
		}
		dir = parent
	}
}

// run executes the test binary with argv and returns (stdout, stderr,
// exit code). exit code 0 indicates clean success; non-zero values
// match what main.go's os.Exit consumed.
func run(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	return runWith(t, "", args...)
}

func runWith(t *testing.T, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		rc := 0
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				rc = exitErr.ExitCode()
			} else {
				t.Fatalf("wait: %v", err)
			}
		}
		return outBuf.String(), errBuf.String(), rc
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("subprocess timed out: %v", args)
		return "", "", -1
	}
}

func TestList(t *testing.T) {
	out, _, rc := run(t, "--list")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	for _, name := range []string{"cat", "echo", "ls", "grep", "wc", "tail"} {
		if !strings.Contains(out, name) {
			t.Errorf("--list missing %q\noutput:\n%s", name, out)
		}
	}
}

func TestVersion(t *testing.T) {
	out, _, rc := run(t, "--version")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out, "topsail") {
		t.Errorf("--version output unexpected: %q", out)
	}
}

func TestTopHelpListsApplets(t *testing.T) {
	out, _, rc := run(t, "--help")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out, "Usage") {
		t.Errorf("--help missing 'Usage': %q", out)
	}
}

func TestPerAppletHelp(t *testing.T) {
	out, _, rc := run(t, "--help", "echo")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(strings.ToLower(out), "echo") {
		t.Errorf("--help echo missing 'echo': %q", out)
	}
}

func TestDispatchEcho(t *testing.T) {
	out, _, rc := run(t, "echo", "hello", "world")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := strings.TrimRight(out, "\r\n"); got != "hello world" {
		t.Errorf("got %q; want %q", got, "hello world")
	}
}

func TestDispatchCatStdin(t *testing.T) {
	out, _, rc := runWith(t, "alpha\nbeta\n", "cat")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("cat stdin output: %q", out)
	}
}

func TestDispatchWcStdin(t *testing.T) {
	out, _, rc := runWith(t, "one\ntwo\nthree\n", "wc", "-l")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("wc -l on 3-line input: %q", out)
	}
}

func TestUnknownApplet(t *testing.T) {
	_, errOut, rc := run(t, "does-not-exist")
	if rc != 127 {
		t.Errorf("rc = %d; want 127 (applet-not-found)", rc)
	}
	if !strings.Contains(errOut, "unknown applet") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestUnknownGlobalFlag(t *testing.T) {
	_, errOut, rc := run(t, "--no-such-flag")
	if rc != 2 {
		t.Errorf("rc = %d; want 2 (usage error)", rc)
	}
	if !strings.Contains(errOut, "unknown option") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestSymlinkDispatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Unprivileged symlinks aren't guaranteed on Windows. Use copy-based
		// dispatch to cover the same code path.
		copyName := filepath.Join(filepath.Dir(binPath), "echo.exe")
		if err := copyFile(binPath, copyName); err != nil {
			t.Skip("copy-based dispatch unavailable:", err)
		}
		t.Cleanup(func() { _ = os.Remove(copyName) })

		cmd := exec.Command(copyName, "hi")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if got := strings.TrimRight(string(out), "\r\n"); got != "hi" {
			t.Errorf("copy dispatch: got %q; want %q", got, "hi")
		}
		return
	}

	link := filepath.Join(filepath.Dir(binPath), "echo")
	if err := os.Symlink(binPath, link); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	t.Cleanup(func() { _ = os.Remove(link) })

	cmd := exec.Command(link, "hi")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimRight(string(out), "\r\n"); got != "hi" {
		t.Errorf("symlink dispatch: got %q; want %q", got, "hi")
	}
}

func TestPipeline(t *testing.T) {
	// Pipe `echo` through `wc -c` via two subprocesses.
	one := exec.Command(binPath, "echo", "abc")
	two := exec.Command(binPath, "wc", "-c")
	pipe, err := one.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	two.Stdin = pipe
	var out bytes.Buffer
	two.Stdout = &out
	if err := two.Start(); err != nil {
		t.Fatal(err)
	}
	if err := one.Run(); err != nil {
		t.Fatal(err)
	}
	if err := two.Wait(); err != nil {
		t.Fatal(err)
	}
	// "abc\n" = 4 bytes. wc may print whitespace+number; just look for 4.
	if !strings.Contains(out.String(), "4") {
		t.Errorf("pipeline output: %q", out.String())
	}
}

func TestSedThroughBinary(t *testing.T) {
	out, _, rc := runWith(t, "alpha\nbeta\ngamma\n", "sed", "-n", "/eta/p")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := strings.TrimRight(out, "\r\n"); got != "beta" {
		t.Errorf("sed -n /eta/p: got %q; want %q", got, "beta")
	}
}

func TestSortKeyThroughBinary(t *testing.T) {
	out, _, rc := runWith(t, "user:101\nadmin:1\nguest:50\n",
		"sort", "-t", ":", "-k", "2", "-n")
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	want := "admin:1\nguest:50\nuser:101"
	if got := strings.TrimRight(out, "\r\n"); !strings.Contains(got, want) {
		t.Errorf("sort -t: -k2 -n: got %q; want %q", got, want)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}
