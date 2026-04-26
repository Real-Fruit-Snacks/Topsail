package echo

import (
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestEchoBasic(t *testing.T) {
	out := testutil.CaptureStdout(t)
	if rc := Main([]string{"echo", "hello", "world"}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got, want := out.String(), "hello world\n"; got != want {
		t.Errorf("echo = %q; want %q", got, want)
	}
}

func TestEchoNoArgs(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo"})
	if got, want := out.String(), "\n"; got != want {
		t.Errorf("echo = %q; want %q", got, want)
	}
}

func TestEchoNoNewline(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "-n", "hello"})
	if got, want := out.String(), "hello"; got != want {
		t.Errorf("echo -n = %q; want %q", got, want)
	}
}

func TestEchoEscapes(t *testing.T) {
	cases := []struct {
		flag, in, want string
	}{
		{"-e", `\n`, "\n\n"},
		{"-e", `tab\there`, "tab\there\n"},
		{"-e", `back\\slash`, "back\\slash\n"},
		{"-e", `bell\a`, "bell\a\n"},
		{"-E", `\n`, `\n` + "\n"},
		{"-e", `stop\chere`, "stop"},
	}
	for _, tc := range cases {
		out := testutil.CaptureStdout(t)
		Main([]string{"echo", tc.flag, tc.in})
		if got := out.String(); got != tc.want {
			t.Errorf("echo %s %q = %q; want %q", tc.flag, tc.in, got, tc.want)
		}
	}
}

func TestEchoUnknownEscape(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "-e", `un\known`})
	if got, want := out.String(), `un\known`+"\n"; got != want {
		t.Errorf("echo -e = %q; want %q", got, want)
	}
}

func TestEchoCombinedFlags(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "-ne", `tab\there`})
	if got, want := out.String(), "tab\there"; got != want {
		t.Errorf("echo -ne = %q; want %q", got, want)
	}
}

func TestEchoUnknownFlagIsData(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "--version"})
	if got, want := out.String(), "--version\n"; got != want {
		t.Errorf("echo --version = %q; want %q", got, want)
	}
}

func TestEchoDoubleDashIsData(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "--", "x"})
	if got, want := out.String(), "-- x\n"; got != want {
		t.Errorf("echo -- x = %q; want %q", got, want)
	}
}

func TestEchoDanglingBackslash(t *testing.T) {
	out := testutil.CaptureStdout(t)
	Main([]string{"echo", "-e", `trail\`})
	if got, want := out.String(), `trail\`+"\n"; got != want {
		t.Errorf("echo -e trail\\ = %q; want %q", got, want)
	}
}
