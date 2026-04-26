package falsecmd

import "testing"

func TestMain(t *testing.T) {
	if rc := Main([]string{"false"}); rc != 1 {
		t.Errorf("Main = %d; want 1", rc)
	}
}

func TestMainIgnoresArgs(t *testing.T) {
	cases := [][]string{
		{"false"},
		{"false", "anything"},
		{"false", "--bogus", "--help"},
	}
	for _, argv := range cases {
		if rc := Main(argv); rc != 1 {
			t.Errorf("Main(%v) = %d; want 1", argv, rc)
		}
	}
}
