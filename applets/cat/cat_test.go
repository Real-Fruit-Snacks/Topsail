package cat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCatStdin(t *testing.T) {
	testutil.SetStdin(t, "from stdin\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"cat"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "from stdin\n" {
		t.Errorf("out = %q", got)
	}
}

func TestCatStdinDash(t *testing.T) {
	testutil.SetStdin(t, "via dash\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-"})
	if got := out.String(); got != "via dash\n" {
		t.Errorf("out = %q", got)
	}
}

func TestCatFile(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "a.txt", "alpha\nbeta\n")
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"cat", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got := out.String(); got != "alpha\nbeta\n" {
		t.Errorf("out = %q", got)
	}
}

func TestCatMultiple(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "A\n")
	b := writeFile(t, dir, "b", "B\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", a, b})
	if got := out.String(); got != "A\nB\n" {
		t.Errorf("out = %q", got)
	}
}

func TestCatNumber(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "alpha\nbeta\n\ngamma\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-n", p})
	want := "     1\talpha\n     2\tbeta\n     3\t\n     4\tgamma\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCatNumberNonBlank(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "alpha\n\nbeta\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-b", p})
	want := "     1\talpha\n\n     2\tbeta\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCatShowEnds(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "x\ny\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-E", p})
	want := "x$\ny$\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCatShowTabs(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "a\tb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-T", p})
	want := "a^Ib\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCatSqueeze(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "a\n\n\n\nb\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-s", p})
	want := "a\n\nb\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestCatMissingFile(t *testing.T) {
	out, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cat", "/this/does/not/exist"}); rc != 1 {
		t.Errorf("rc = %d; want 1", rc)
	}
	if out.Len() != 0 {
		t.Errorf("stdout had %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "/this/does/not/exist") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCatInvalidOption(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"cat", "-x"}); rc != 2 {
		t.Errorf("rc = %d; want 2", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestCatCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "f", "a\tb\nc\n")
	out := testutil.CaptureStdout(t)
	Main([]string{"cat", "-nET", p})
	want := "     1\ta^Ib$\n     2\tc$\n"
	if got := out.String(); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
