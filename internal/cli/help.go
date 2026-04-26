package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
)

// printTopHelp writes the top-level help text shown by `topsail --help`.
func printTopHelp(w io.Writer) {
	_, _ = fmt.Fprintf(w, `topsail %s — single-file BusyBox-like multi-call binary.

Usage:
  topsail [--help|--version|--list]
  topsail <applet> [args...]
  <applet> [args...]                # via symlink or copy named after an applet

Global options:
  --help, -h           Show this help. With an applet name, show its help.
  --version, -V        Print version, commit, and build date.
  --list               List every available applet.

See "topsail --help <applet>" for per-applet help.
`, Version)
}

// printList writes a two-column applet listing with one-line help text.
func printList(w io.Writer, applets []applet.Applet) {
	if len(applets) == 0 {
		_, _ = fmt.Fprintln(w, "no applets registered")
		return
	}
	sort.Slice(applets, func(i, j int) bool { return applets[i].Name < applets[j].Name })

	width := 0
	for _, a := range applets {
		if n := len(a.Name); n > width {
			width = n
		}
	}
	for _, a := range applets {
		_, _ = fmt.Fprintf(w, "  %-*s  %s\n", width, a.Name, a.Help)
	}
}

// printAppletHelp writes the per-applet usage text. Falls back to a stub if
// the applet did not provide one.
func printAppletHelp(w io.Writer, a applet.Applet) {
	usage := strings.TrimRight(a.Usage, "\n")
	if usage == "" {
		usage = fmt.Sprintf("Usage: %s [args...]\n\n%s", a.Name, a.Help)
	}
	_, _ = fmt.Fprintln(w, usage)
}

// printVersion writes the build identification.
func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "topsail %s (commit %s, built %s)\n", Version, Commit, Date)
}
