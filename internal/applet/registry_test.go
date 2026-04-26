package applet

import (
	"testing"
)

func noopMain([]string) int { return 0 }

func TestRegisterAndGet(t *testing.T) {
	ResetForTesting()
	Register(Applet{
		Name:    "foo",
		Aliases: []string{"f", "foo2"},
		Help:    "do foo",
		Main:    noopMain,
	})

	for _, name := range []string{"foo", "f", "foo2"} {
		got, ok := Get(name)
		if !ok {
			t.Errorf("Get(%q) = !ok; want ok", name)
			continue
		}
		if got.Name != "foo" {
			t.Errorf("Get(%q).Name = %q; want %q", name, got.Name, "foo")
		}
	}

	if _, ok := Get("bar"); ok {
		t.Errorf(`Get("bar") = ok; want !ok`)
	}
}

func TestAllAndNames(t *testing.T) {
	ResetForTesting()
	Register(Applet{Name: "b", Main: noopMain})
	Register(Applet{Name: "a", Aliases: []string{"alpha"}, Main: noopMain})
	Register(Applet{Name: "c", Main: noopMain})

	all := All()
	if len(all) != 3 {
		t.Fatalf("len(All()) = %d; want 3 (aliases must not double-count)", len(all))
	}
	if all[0].Name != "a" || all[1].Name != "b" || all[2].Name != "c" {
		t.Errorf("All() not sorted: got %v", []string{all[0].Name, all[1].Name, all[2].Name})
	}

	names := Names()
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("Names() = %v; want [a b c]", names)
	}
}

func TestAllEmpty(t *testing.T) {
	ResetForTesting()
	if got := All(); len(got) != 0 {
		t.Errorf("All() on empty registry = %v; want []", got)
	}
}

func TestRegisterPanicsOnEmptyName(t *testing.T) {
	ResetForTesting()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Register({Name:\"\"}) did not panic")
		}
	}()
	Register(Applet{Main: noopMain})
}

func TestRegisterPanicsOnNilMain(t *testing.T) {
	ResetForTesting()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Register({Main:nil}) did not panic")
		}
	}()
	Register(Applet{Name: "x"})
}

func TestRegisterPanicsOnDuplicateName(t *testing.T) {
	ResetForTesting()
	Register(Applet{Name: "dup", Main: noopMain})
	defer func() {
		if r := recover(); r == nil {
			t.Error("duplicate Register did not panic")
		}
	}()
	Register(Applet{Name: "dup", Main: noopMain})
}

func TestRegisterPanicsOnAliasCollision(t *testing.T) {
	ResetForTesting()
	Register(Applet{Name: "first", Aliases: []string{"shared"}, Main: noopMain})
	defer func() {
		if r := recover(); r == nil {
			t.Error("alias collision did not panic")
		}
	}()
	Register(Applet{Name: "second", Aliases: []string{"shared"}, Main: noopMain})
}

func TestRegisterPanicsOnEmptyAlias(t *testing.T) {
	ResetForTesting()
	defer func() {
		if r := recover(); r == nil {
			t.Error("empty alias did not panic")
		}
	}()
	Register(Applet{Name: "x", Aliases: []string{""}, Main: noopMain})
}

func TestRegisterPanicsOnAliasMatchingExistingName(t *testing.T) {
	ResetForTesting()
	Register(Applet{Name: "a", Main: noopMain})
	defer func() {
		if r := recover(); r == nil {
			t.Error("alias matching an existing Name did not panic")
		}
	}()
	Register(Applet{Name: "b", Aliases: []string{"a"}, Main: noopMain})
}
