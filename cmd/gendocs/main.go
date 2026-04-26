// Command gendocs writes shell completions and man pages for every
// registered applet to an output directory. It runs as a goreleaser
// before-hook so each release archive ships with up-to-date docs.
//
// Usage:
//
//	gendocs -out dist/docs
//
// produces:
//
//	dist/docs/completions/topsail.bash
//	dist/docs/completions/_topsail        # zsh
//	dist/docs/completions/topsail.fish
//	dist/docs/man/topsail.1
//	dist/docs/man/<applet>.1              # one per applet
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	_ "github.com/Real-Fruit-Snacks/topsail/internal/applets"
)

// version is overridden at build time via -ldflags.
var version = "dev"

func main() {
	out := flag.String("out", "dist/docs", "output directory")
	flag.Parse()

	if err := os.MkdirAll(filepath.Join(*out, "completions"), 0o755); err != nil { //nolint:gosec // staging dir for world-readable docs; 0o755 is canonical
		fail(err)
	}
	if err := os.MkdirAll(filepath.Join(*out, "man"), 0o755); err != nil { //nolint:gosec // staging dir for world-readable docs; 0o755 is canonical
		fail(err)
	}

	applets := applet.All()
	names := appletNames(applets)

	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(*out, "completions", "topsail.bash"), bashCompletion(names)},
		{filepath.Join(*out, "completions", "_topsail"), zshCompletion(names)},
		{filepath.Join(*out, "completions", "topsail.fish"), fishCompletion(names)},
		{filepath.Join(*out, "man", "topsail.1"), topManPage(applets)},
	}
	for _, a := range applets {
		files = append(files, struct {
			path    string
			content string
		}{
			filepath.Join(*out, "man", a.Name+".1"),
			appletManPage(a),
		})
	}

	for _, f := range files {
		if err := os.WriteFile(f.path, []byte(f.content), 0o644); err != nil { //nolint:gosec // shell completions / man pages need to be world-readable
			fail(err)
		}
	}
	fmt.Printf("gendocs: wrote %d files to %s\n", len(files), *out)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gendocs:", err)
	os.Exit(1)
}

func appletNames(applets []applet.Applet) []string {
	out := make([]string, len(applets))
	for i, a := range applets {
		out[i] = a.Name
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// Shell completions
// ---------------------------------------------------------------------------

func bashCompletion(names []string) string {
	return fmt.Sprintf(`# bash completion for topsail
# Source from /etc/bash_completion.d/ or ~/.bash_completion.d/

_topsail() {
    local cur
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [ "$COMP_CWORD" -eq 1 ]; then
        local applets='--help --version --list %s'
        # shellcheck disable=SC2207
        COMPREPLY=( $(compgen -W "$applets" -- "$cur") )
        return 0
    fi
    # After the applet, fall through to filename completion.
    # shellcheck disable=SC2207
    COMPREPLY=( $(compgen -f -- "$cur") )
    return 0
}

complete -F _topsail topsail
`, strings.Join(names, " "))
}

func zshCompletion(names []string) string {
	return fmt.Sprintf(`#compdef topsail
# zsh completion for topsail
# Drop into a directory on $fpath (e.g. /usr/share/zsh/site-functions/)

_topsail() {
    local -a applets
    applets=( %s )
    if (( CURRENT == 2 )); then
        _describe 'applet' applets
    else
        _files
    fi
}

_topsail "$@"
`, strings.Join(quoteAll(names), " "))
}

func fishCompletion(names []string) string {
	var b strings.Builder
	b.WriteString("# fish completion for topsail\n")
	b.WriteString("# Drop into ~/.config/fish/completions/topsail.fish\n\n")
	b.WriteString("complete -c topsail -f -n '__fish_use_subcommand' -l help -d 'Show top-level help'\n")
	b.WriteString("complete -c topsail -f -n '__fish_use_subcommand' -l version -d 'Print version banner'\n")
	b.WriteString("complete -c topsail -f -n '__fish_use_subcommand' -l list -d 'List every registered applet'\n")
	for _, n := range names {
		fmt.Fprintf(&b, "complete -c topsail -f -n '__fish_use_subcommand' -a '%s' -d 'topsail %s applet'\n", n, n)
	}
	b.WriteString("# After the applet name, complete file paths.\n")
	b.WriteString("complete -c topsail -F -n 'not __fish_use_subcommand'\n")
	return b.String()
}

func quoteAll(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	return out
}

// ---------------------------------------------------------------------------
// Man pages (groff -man format)
// ---------------------------------------------------------------------------

func topManPage(applets []applet.Applet) string {
	date := time.Now().UTC().Format("2006-01-02")
	var b strings.Builder
	fmt.Fprintf(&b, ".TH TOPSAIL 1 \"%s\" \"topsail %s\" \"Topsail Manual\"\n", date, version)
	b.WriteString(".SH NAME\n")
	b.WriteString("topsail \\- BusyBox-style multi-call binary in Go\n")
	b.WriteString(".SH SYNOPSIS\n")
	b.WriteString(".B topsail\n")
	b.WriteString("[\\fB--help\\fR|\\fB--version\\fR|\\fB--list\\fR]\n")
	b.WriteString(".br\n")
	b.WriteString(".B topsail\n")
	b.WriteString("\\fIAPPLET\\fR [\\fIARGS\\fR]...\n")
	b.WriteString(".br\n")
	b.WriteString("\\fIAPPLET\\fR [\\fIARGS\\fR]...\n")
	b.WriteString("(via symlink or copy)\n")
	b.WriteString(".SH DESCRIPTION\n")
	b.WriteString("topsail is a single static Go binary that bundles a large set of POSIX/coreutils applets.\n")
	b.WriteString("Each applet is dispatched by matching argv[0]'s basename against the registered applet names; ")
	b.WriteString("invocations like \\fItopsail cat file\\fR or, via a symlink named \\fIcat\\fR, simply \\fIcat file\\fR both work.\n")
	b.WriteString(".SH GLOBAL OPTIONS\n")
	b.WriteString(".TP\n.B --help\nShow top-level help. With an applet name, show that applet's help.\n")
	b.WriteString(".TP\n.B --version\nPrint the version, commit, and build date.\n")
	b.WriteString(".TP\n.B --list\nList every registered applet.\n")
	b.WriteString(".SH APPLETS\n")
	for _, a := range applets {
		fmt.Fprintf(&b, ".TP\n.B %s\n%s\n", roffEscape(a.Name), roffEscape(a.Help))
	}
	b.WriteString(".SH EXIT STATUS\n")
	b.WriteString(".TP\n.B 0\nSuccess.\n")
	b.WriteString(".TP\n.B 1\nRuntime error.\n")
	b.WriteString(".TP\n.B 2\nUsage error.\n")
	b.WriteString(".TP\n.B 127\nApplet name not found.\n")
	b.WriteString(".SH SEE ALSO\n")
	b.WriteString("Per-applet pages: \\fBcat\\fR(1), \\fBgrep\\fR(1), \\fBawk\\fR(1), ...\n")
	b.WriteString(".SH AUTHORS\n")
	b.WriteString("Written by the topsail contributors. Source: https://github.com/Real-Fruit-Snacks/topsail\n")
	return b.String()
}

func appletManPage(a applet.Applet) string {
	date := time.Now().UTC().Format("2006-01-02")
	upper := strings.ToUpper(a.Name)
	var b strings.Builder
	fmt.Fprintf(&b, ".TH %s 1 \"%s\" \"topsail %s\" \"Topsail Manual\"\n", upper, date, version)
	b.WriteString(".SH NAME\n")
	fmt.Fprintf(&b, "%s \\- %s\n", roffEscape(a.Name), roffEscape(a.Help))
	b.WriteString(".SH SYNOPSIS\n")
	fmt.Fprintf(&b, ".B topsail %s\n[\\fIOPTION\\fR]... [\\fIARGS\\fR]...\n", roffEscape(a.Name))
	b.WriteString(".br\n")
	fmt.Fprintf(&b, ".B %s\n[\\fIOPTION\\fR]... [\\fIARGS\\fR]... (via symlink/copy)\n", roffEscape(a.Name))
	b.WriteString(".SH DESCRIPTION\n")
	usage := strings.TrimSpace(a.Usage)
	if usage == "" {
		fmt.Fprintf(&b, "See \\fBtopsail\\fR(1).\n")
	} else {
		// Render the Usage block as a literal-ish indented block.
		b.WriteString(".nf\n")
		b.WriteString(roffEscape(usage))
		b.WriteString("\n.fi\n")
	}
	b.WriteString(".SH SEE ALSO\n")
	b.WriteString(".BR topsail (1)\n")
	return b.String()
}

// roffEscape escapes the few characters that have special meaning in roff
// (backslashes and dot-at-start-of-line).
func roffEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Lines starting with '.' are roff macros. Quote with zero-width
	// character to keep them literal.
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, ".") {
			lines[i] = `\&` + line
		}
	}
	return strings.Join(lines, "\n")
}
