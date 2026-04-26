package truecmd

import "testing"

func TestMain(t *testing.T) {
	if rc := Main([]string{"true"}); rc != 0 {
		t.Errorf("Main = %d; want 0", rc)
	}
}

func TestMainIgnoresArgs(t *testing.T) {
	cases := [][]string{
		{"true"},
		{"true", "anything"},
		{"true", "--bogus", "--help", "--version"},
		{"true", "", "", ""},
	}
	for _, argv := range cases {
		if rc := Main(argv); rc != 0 {
			t.Errorf("Main(%v) = %d; want 0", argv, rc)
		}
	}
}
