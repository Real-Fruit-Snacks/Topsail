package truncate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Real-Fruit-Snacks/topsail/internal/testutil"
)

func TestTruncateAbsoluteShrink(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "-s", "5", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	st, _ := os.Stat(p)
	if st.Size() != 5 {
		t.Errorf("size = %d; want 5", st.Size())
	}
}

func TestTruncateAbsoluteExtend(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "-s", "100", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	st, _ := os.Stat(p)
	if st.Size() != 100 {
		t.Errorf("size = %d; want 100", st.Size())
	}
}

func TestTruncateRelativeAdd(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, make([]byte, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	Main([]string{"truncate", "-s", "+5", p})
	st, _ := os.Stat(p)
	if st.Size() != 15 {
		t.Errorf("size = %d; want 15", st.Size())
	}
}

func TestTruncateRelativeSub(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, make([]byte, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	Main([]string{"truncate", "-s", "-3", p})
	st, _ := os.Stat(p)
	if st.Size() != 7 {
		t.Errorf("size = %d; want 7", st.Size())
	}
}

func TestTruncateClampLessThan(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, make([]byte, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	// "<5" means: only if larger than 5. We're 10, so go to 5.
	Main([]string{"truncate", "-s", "<5", p})
	st, _ := os.Stat(p)
	if st.Size() != 5 {
		t.Errorf("size = %d; want 5", st.Size())
	}
}

func TestTruncateClampGreaterThan(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, make([]byte, 3), 0o644); err != nil {
		t.Fatal(err)
	}
	testutil.CaptureStdio(t)
	// ">10" means: only if smaller than 10. We're 3, so extend to 10.
	Main([]string{"truncate", "-s", ">10", p})
	st, _ := os.Stat(p)
	if st.Size() != 10 {
		t.Errorf("size = %d; want 10", st.Size())
	}
}

func TestTruncateCreatesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fresh")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "-s", "1024", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() != 1024 {
		t.Errorf("size = %d; want 1024", st.Size())
	}
}

func TestTruncateNoCreate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "absent")
	testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "-c", "-s", "10", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("file should not exist: %v", err)
	}
}

func TestTruncateSizeSuffixes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	testutil.CaptureStdio(t)
	Main([]string{"truncate", "-s", "1K", p})
	st, _ := os.Stat(p)
	if st.Size() != 1024 {
		t.Errorf("1K = %d; want 1024", st.Size())
	}
	Main([]string{"truncate", "-s", "1KB", p})
	st, _ = os.Stat(p)
	if st.Size() != 1000 {
		t.Errorf("1KB = %d; want 1000", st.Size())
	}
}

func TestTruncateAttachedSize(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	testutil.CaptureStdio(t)
	// -s5 attached form.
	if rc := Main([]string{"truncate", "-s5", p}); rc != 0 {
		t.Errorf("rc = %d", rc)
	}
	st, _ := os.Stat(p)
	if st.Size() != 5 {
		t.Errorf("size = %d; want 5", st.Size())
	}
}

func TestTruncateMissingSize(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "f"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "required") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestTruncateInvalidSize(t *testing.T) {
	_, errBuf := testutil.CaptureStdio(t)
	if rc := Main([]string{"truncate", "-s", "garbage", "f"}); rc != 2 {
		t.Errorf("rc = %d", rc)
	}
	if !strings.Contains(errBuf.String(), "invalid size") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
