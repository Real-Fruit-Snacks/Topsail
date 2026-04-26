// Package applet defines the contract every topsail applet implements
// and the registry used by the dispatcher to look them up by name or alias.
package applet

// MainFunc is the entry point of an applet. argv[0] is the invocation name
// (the basename of the multi-call symlink/copy, or the canonical applet name
// when invoked as "topsail <applet> ..."). The return value is the process
// exit status: 0 success, 1 runtime error, 2 usage error, 127 not found.
type MainFunc func(argv []string) int

// Applet is the contract every topsail applet self-registers in init().
type Applet struct {
	// Name is the canonical applet name, e.g. "ls", "wc", "grep".
	Name string

	// Aliases are alternative invocation names. May be empty.
	Aliases []string

	// Help is a one-line description shown by --list.
	Help string

	// Usage is the multi-line help text shown by "<applet> --help" and
	// "topsail --help <applet>". Should follow the standard Usage: form.
	Usage string

	// Main is the applet entry point.
	Main MainFunc
}
