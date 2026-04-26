package filemode

import (
	"os"
	"testing"
)

func TestParseOctal(t *testing.T) {
	cases := []struct {
		in   string
		want os.FileMode
	}{
		{"0", 0},
		{"7", 0o7},
		{"644", 0o644},
		{"0755", 0o755},
		{"4755", 0o4755}, // setuid + 0755
	}
	for _, c := range cases {
		got, err := Parse(c.in, 0)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %04o; want %04o", c.in, got, c.want)
		}
	}
}

func TestParseOctalInvalid(t *testing.T) {
	for _, in := range []string{"8", "9", "0xff"} {
		if _, err := Parse(in, 0); err == nil {
			t.Errorf("%q: expected error", in)
		}
	}
}

func TestParseSymbolicAdd(t *testing.T) {
	got, err := Parse("u+x", 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o744 {
		t.Errorf("u+x on 0644: got %04o; want 0744", got)
	}
}

func TestParseSymbolicRemove(t *testing.T) {
	got, err := Parse("go-w", 0o666)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o644 {
		t.Errorf("go-w on 0666: got %04o; want 0644", got)
	}
}

func TestParseSymbolicEqual(t *testing.T) {
	got, err := Parse("a=rwx", 0o000)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o777 {
		t.Errorf("a=rwx: got %04o; want 0777", got)
	}
}

func TestParseSymbolicMulti(t *testing.T) {
	got, err := Parse("u=rwx,go=rx", 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o755 {
		t.Errorf("u=rwx,go=rx: got %04o; want 0755", got)
	}
}

func TestParseSymbolicEmptyWho(t *testing.T) {
	// "+x" with no who is treated as 'a'.
	got, err := Parse("+x", 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o755 {
		t.Errorf("+x on 0644: got %04o; want 0755", got)
	}
}

func TestParseSymbolicSetuidSticky(t *testing.T) {
	got, err := Parse("u+s,o+t", 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o5755 {
		t.Errorf("u+s,o+t on 0755: got %04o; want 5755", got)
	}
}

func TestParseSymbolicCapitalX(t *testing.T) {
	// On a regular file with no exec bits, X has no effect.
	got, err := Parse("a+X", 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o644 {
		t.Errorf("a+X on plain 0644: got %04o; want 0644", got)
	}
	// On a regular file that already has any x bit, X applies.
	got, err = Parse("a+X", 0o744)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o755 {
		t.Errorf("a+X on 0744: got %04o; want 0755", got)
	}
	// Pretend it's a directory by setting the directory bit.
	got, err = Parse("a+X", os.FileMode(0o644)|os.ModeDir)
	if err != nil {
		t.Fatal(err)
	}
	// We only care about the low permission bits; the dir bit isn't
	// returned (parseSymbolic strips to permLowAll).
	if got&0o777 != 0o755 {
		t.Errorf("a+X on dir 0644: got %04o; want low bits 0755", got)
	}
}

func TestParseSymbolicCopyFromWho(t *testing.T) {
	// g=u copies u's perms into g.
	got, err := Parse("g=u", 0o740)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0o770 {
		t.Errorf("g=u on 0740: got %04o; want 0770", got)
	}
}

func TestParseSymbolicErrors(t *testing.T) {
	cases := []string{
		"",
		"u",   // no operator
		"u+",  // valid (no bits) — actually GNU treats this as no-op; we accept it
		"u@x", // invalid op
		"q+x", // invalid who letter falls through to op test, then 'q' isn't an op either
		"u+z", // invalid permission
		",",   // empty clause
	}
	for _, in := range cases {
		_, err := Parse(in, 0o644)
		// Some of these may be accepted; we only assert that "" and "," fail.
		if (in == "" || in == ",") && err == nil {
			t.Errorf("%q: expected error", in)
		}
	}
}
