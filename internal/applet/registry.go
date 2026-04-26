package applet

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]Applet)
)

// Register adds an applet to the global registry. It panics if the applet
// is invalid (empty Name or nil Main) or if the Name or any Alias is
// already taken — these are programmer errors that must surface at startup.
//
// Applets call this from their init() function.
func Register(a Applet) {
	if a.Name == "" {
		panic("applet: Register called with empty Name")
	}
	if a.Main == nil {
		panic(fmt.Sprintf("applet: Register(%q) with nil Main", a.Name))
	}

	mu.Lock()
	defer mu.Unlock()

	if existing, dup := registry[a.Name]; dup {
		panic(fmt.Sprintf("applet: %q already registered (existing=%q)", a.Name, existing.Name))
	}
	registry[a.Name] = a

	for _, alias := range a.Aliases {
		if alias == "" {
			panic(fmt.Sprintf("applet: %q has empty alias", a.Name))
		}
		if existing, dup := registry[alias]; dup {
			panic(fmt.Sprintf("applet: alias %q for %q collides with %q", alias, a.Name, existing.Name))
		}
		registry[alias] = a
	}
}

// Get looks up an applet by name or alias. The boolean is true if found.
func Get(name string) (Applet, bool) {
	mu.RLock()
	defer mu.RUnlock()
	a, ok := registry[name]
	return a, ok
}

// All returns every registered applet, deduplicated by Name and sorted
// alphabetically. Aliases do not produce duplicate entries.
func All() []Applet {
	mu.RLock()
	defer mu.RUnlock()

	seen := make(map[string]struct{}, len(registry))
	out := make([]Applet, 0, len(registry))
	for _, a := range registry {
		if _, dup := seen[a.Name]; dup {
			continue
		}
		seen[a.Name] = struct{}{}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Names returns every registered applet name, sorted. Aliases are not included.
func Names() []string {
	all := All()
	out := make([]string, len(all))
	for i, a := range all {
		out[i] = a.Name
	}
	return out
}

// ResetForTesting empties the registry. It is exported so other internal
// packages' tests can produce a clean fixture; production code MUST NOT call it.
func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Applet)
}
