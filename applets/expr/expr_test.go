package expr

import (
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func capture(t *testing.T, argv []string) (output string, rc int) {
	t.Helper()
	out := testutil.CaptureStdout(t)
	rc = Main(argv)
	return strings.TrimRight(out.String(), "\n"), rc
}

func TestExprAdd(t *testing.T) {
	got, rc := capture(t, []string{"expr", "1", "+", "2"})
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestExprSubtract(t *testing.T) {
	got, _ := capture(t, []string{"expr", "10", "-", "4"})
	if got != "6" {
		t.Errorf("got %q", got)
	}
}

func TestExprMul(t *testing.T) {
	got, _ := capture(t, []string{"expr", "3", "*", "4"})
	if got != "12" {
		t.Errorf("got %q", got)
	}
}

func TestExprDiv(t *testing.T) {
	got, _ := capture(t, []string{"expr", "10", "/", "3"})
	if got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestExprMod(t *testing.T) {
	got, _ := capture(t, []string{"expr", "10", "%", "3"})
	if got != "1" {
		t.Errorf("got %q", got)
	}
}

func TestExprPrecedence(t *testing.T) {
	got, _ := capture(t, []string{"expr", "1", "+", "2", "*", "3"})
	if got != "7" {
		t.Errorf("got %q", got)
	}
}

func TestExprParens(t *testing.T) {
	got, _ := capture(t, []string{"expr", "(", "1", "+", "2", ")", "*", "3"})
	if got != "9" {
		t.Errorf("got %q", got)
	}
}

func TestExprCompareEqual(t *testing.T) {
	got, rc := capture(t, []string{"expr", "5", "=", "5"})
	if got != "1" {
		t.Errorf("got %q", got)
	}
	if rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	got, rc = capture(t, []string{"expr", "5", "=", "6"})
	if got != "0" {
		t.Errorf("got %q", got)
	}
	if rc != 1 {
		t.Errorf("rc = %d (zero result should be exit 1)", rc)
	}
}

func TestExprCompareLess(t *testing.T) {
	got, _ := capture(t, []string{"expr", "3", "<", "5"})
	if got != "1" {
		t.Errorf("got %q", got)
	}
}

func TestExprStringEqual(t *testing.T) {
	got, _ := capture(t, []string{"expr", "abc", "=", "abc"})
	if got != "1" {
		t.Errorf("got %q", got)
	}
}

func TestExprAnd(t *testing.T) {
	got, _ := capture(t, []string{"expr", "5", "&", "10"})
	if got != "5" {
		t.Errorf("got %q", got)
	}
	got, _ = capture(t, []string{"expr", "0", "&", "10"})
	if got != "0" {
		t.Errorf("got %q", got)
	}
}

func TestExprOr(t *testing.T) {
	got, _ := capture(t, []string{"expr", "0", "|", "10"})
	if got != "10" {
		t.Errorf("got %q", got)
	}
	got, _ = capture(t, []string{"expr", "5", "|", "10"})
	if got != "5" {
		t.Errorf("got %q", got)
	}
}

func TestExprLength(t *testing.T) {
	got, _ := capture(t, []string{"expr", "length", "hello"})
	if got != "5" {
		t.Errorf("got %q", got)
	}
}

func TestExprSubstr(t *testing.T) {
	got, _ := capture(t, []string{"expr", "substr", "hello", "2", "3"})
	if got != "ell" {
		t.Errorf("got %q", got)
	}
}

func TestExprMissingOperand(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"expr"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if errBuf.Len() == 0 {
		t.Error("expected stderr")
	}
}

func TestExprDivByZero(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"expr", "1", "/", "0"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "division by zero") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestExprNonNumericArith(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"expr", "abc", "+", "1"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "non-numeric") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
