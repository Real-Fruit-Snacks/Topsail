# Architecture

This document captures the design choices that make `topsail` a maintainable single-binary multi-call utility, and the documented divergences from its Python sibling [mainsail](https://github.com/Real-Fruit-Snacks/mainsail).

## Table of contents

1. [Layout](#layout)
2. [The applet contract](#the-applet-contract)
3. [Registry](#registry)
4. [Dispatch flow](#dispatch-flow)
5. [I/O abstraction](#io-abstraction)
6. [Cross-platform shims](#cross-platform-shims)
7. [Build identification](#build-identification)
8. [Quality stack](#quality-stack)
9. [Documented divergences from mainsail](#documented-divergences-from-mainsail)
10. [Adding an applet](#adding-an-applet)

## Layout

```
topsail/
├── cmd/topsail/                        process entry point
│   └── main.go                         imports applet packages, calls cli.Run
├── applets/                            one package per applet
│   ├── cat/cat.go cat_test.go
│   ├── grep/grep.go grep_test.go
│   ├── awk/awk.go readfile.go awk_test.go   (vendors benhoyt/goawk)
│   └── ...                             (~64 packages, ~74 registered names)
├── internal/
│   ├── applet/                         the contract: Applet struct + registry
│   ├── cli/                            dispatcher: argv[0] match, --help/--version/--list
│   ├── ioutil/                         pluggable Stdin / Stdout / Stderr
│   ├── platform/                       per-OS user/group lookup, terminal size
│   ├── hashing/                        shared md5/sha256/sha512 logic
│   └── testutil/                       stdio capture helpers used by every applet test
├── .github/workflows/                  CI / CodeQL / govulncheck / Release
├── .golangci.yml                       lint stack (v2 config)
├── .goreleaser.yaml                    release config (cross-compile, cosign, SBOM)
└── Makefile                            local mirror of the CI gauntlet
```

## The applet contract

Every applet lives in its own package under `applets/<name>/` and self-registers in `init()`:

```go
package cat

import "github.com/Real-Fruit-Snacks/topsail/internal/applet"

func init() {
    applet.Register(applet.Applet{
        Name:    "cat",
        Aliases: nil,                   // optional alternate invocation names
        Help:    "concatenate files to stdout",
        Usage:   usage,                 // multi-line help string
        Main:    Main,                  // func(argv []string) int
    })
}

func Main(argv []string) int {
    // argv[0] is the invocation name (basename of the symlink, or canonical
    // applet name when invoked via `topsail <name> ...`).
    // Return value is the process exit code.
    return 0
}
```

Exit-code conventions across applets:

| Code  | Meaning                                  |
| ----- | ---------------------------------------- |
| `0`   | Success.                                 |
| `1`   | Runtime error (I/O, permission, etc.)    |
| `2`   | Usage error (bad flag, missing operand). |
| `127` | Applet name not found in the registry.   |

## Registry

`internal/applet/registry.go` holds a `sync.RWMutex`-protected `map[string]Applet`. `Register` panics on:

- empty `Name` or nil `Main` (programmer error),
- duplicate `Name`,
- alias collision with an existing `Name` or alias.

These are deliberate panics: they fire at process startup before any user-facing work happens, so the failure mode is loud and immediate rather than a silent shadow.

`Get(name)` returns the `Applet` and a found bool; aliases resolve to the same `Applet` value as the canonical name. `All()` deduplicates by `Name` and returns a sorted slice — used by `--list` and `topsail --help`.

## Dispatch flow

The binary's entry point is `cmd/topsail/main.go`, which imports every applet package for side effects and then calls `cli.Run(os.Args)`. `cli.Run` is the single decision point:

```
                    cli.Run(os.Args)
                          │
                  basename(argv[0])
                          │
       ┌──────────────────┴──────────────────┐
       │                                      │
   matches an applet?                  doesn't match
   (and ≠ "topsail")                          │
       │                                      ▼
       ▼                              wrapper mode:
multi-call mode:                      topsail [global flags] [APPLET [args...]]
applet.Main(argv)                             │
                                              ├── --help        (full top-level help)
                                              ├── --help APPLET (per-applet help)
                                              ├── --version     (build banner)
                                              ├── --list        (registered applets)
                                              ├── -h           (same as --help)
                                              └── APPLET args... (recurse via runApplet)
```

A few subtleties worth noting:

- **`.exe` is stripped** from the basename before lookup, so a Windows `cp.exe` symlink/copy resolves to the `cp` applet.
- **`-h` reaches the applet in multi-call mode** (so `df -h` keeps human-readable behavior). Only the long form `--help` is intercepted there.
- **`--`** stops `--help` interception after parsing; `echo -- --help` echoes `-- --help` literally.
- **The wrapper name (`topsail`)** never matches an applet, even if a pathological registration claimed that name; wrapper mode wins.

## I/O abstraction

Every applet writes to `internal/ioutil.Stdout` / `Stderr` and reads from `internal/ioutil.Stdin` rather than the `os.*` globals. This costs nothing in production (the package-level vars default to `os.Stdin/Stdout/Stderr`) and makes tests trivial: `internal/testutil.CaptureStdio` swaps in `*bytes.Buffer`s that the test owns, runs `Main`, and restores the originals via `t.Cleanup`. No subprocess plumbing, no flaky pipe handling.

`ioutil.Errf(format, args...)` is the canonical one-liner for diagnostics:

```go
ioutil.Errf("cat: %s: %v", name, err)
```

It writes to `Stderr` and ensures a trailing newline.

## Cross-platform shims

`internal/platform/` covers the small set of operations where the standard library does not give a uniform answer:

| Function          | Unix                                      | Windows                              |
| ----------------- | ----------------------------------------- | ------------------------------------ |
| `IsTerminal(f)`   | `golang.org/x/term.IsTerminal`            | same                                 |
| `TerminalSize(f)` | `term.GetSize`                            | same                                 |
| `UserName(uid)`   | `os/user.LookupId(uid).Username`          | returns `"-"` (parity with mainsail) |
| `GroupName(gid)`  | `os/user.LookupGroupId(gid).Name`         | returns `"-"`                        |

Filesystem-statistics applets (`df`) split between `df_unix.go` (`golang.org/x/sys/unix.Statfs`) and `df_windows.go` (`golang.org/x/sys/windows.GetDiskFreeSpaceEx`), gated by `//go:build` constraints.

## Build identification

Three `var` symbols in `internal/cli/version.go` are populated at link time via `-ldflags -X`:

```go
var (
    Version = "dev"      // semver tag (e.g. "v1.0.0")
    Commit  = "none"     // short git SHA
    Date    = "unknown"  // RFC 3339 timestamp
)
```

The `Makefile` and `.goreleaser.yaml` both pass these explicitly. `topsail --version` prints `topsail <Version> (commit <Commit>, built <Date>)`.

## Quality stack

The `make ci` target — and the `CI` GitHub Actions workflow — run the same gauntlet:

1. `go mod tidy` drift check (the worktree must be clean afterward)
2. `goimports -local github.com/Real-Fruit-Snacks/topsail` formatting check
3. `go vet ./...`
4. `golangci-lint run` (errcheck, gocritic, gosec, govet, ineffassign, misspell, revive, staticcheck, unparam, unused)
5. `go test -race -cover ./...` across Linux/macOS/Windows × Go {1.25, 1.26}
6. `govulncheck ./...` (with a weekly cron in addition to per-push)
7. `go build` for the host platform
8. Cross-compile to all six release targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64)

Releases additionally sign `checksums.txt` via cosign keyless (GitHub OIDC) and emit an SBOM per archive via syft.

## Documented divergences from mainsail

These are intentional behavioral differences. They are surfaced both here and at runtime where appropriate.

- **`grep` uses RE2, not POSIX BRE/ERE.** Go's `regexp` package is RE2-flavored, so `\b`, `[[:alpha:]]`, lookarounds, and backreferences behave per RE2's spec. `grep -E` is accepted as a parity flag (RE2 is always extended).
- **`sed` substitution-only.** This release ships only the `s/PATTERN/REPL/[gi]` command. Address selectors (`/re/d`), append/insert, and multi-command scripts are deferred. Unsupported scripts produce a clear "only 's/.../.../' is supported in this build" error.
- **`awk` runs with `system()` disabled.** The embedded `goawk` interpreter is configured with `NoExec = true`, blocking shell-injection paths. Pure data-processing awk is unaffected.
- **`ping` is TCP, not ICMP.** Pure-Go raw-socket ICMP requires elevated privileges; we do a TCP connect to a probe port (default 80) and report connect latency. The output format mirrors `ping`'s for muscle memory.
- **`chown` on Windows is a stub.** Windows uses ACLs/SIDs, which Go's `os.Chown` cannot manipulate portably. The applet prints a clear "not supported on Windows in this build" diagnostic and exits 1.
- **`basename` / `dirname` use the `path` package, not `path/filepath`.** They are POSIX string-manipulation utilities; `/` is the separator on every OS, so `dirname /a/b` returns `/a` even on Windows.
- **Symbolic file modes (`u+rwx`).** `mkdir -m` and `chmod` accept octal modes only in this release; symbolic-mode parsing is a follow-up.
- **`tail -f` follow mode.** Static `tail` works; the inotify/poll loop for `-f` is deferred. The applet errors out clearly if `-f` is passed.

When porting future applets from mainsail, treat the Python implementation as the reference; mirror flag names, error messages, and exit codes unless there is a documented reason to diverge. Add the divergence to this section.

## Adding an applet

The smallest correct applet is ~25 lines:

```go
// applets/foo/foo.go
package foo

import "github.com/Real-Fruit-Snacks/topsail/internal/applet"

func init() {
    applet.Register(applet.Applet{
        Name: "foo",
        Help: "do the foo thing",
        Usage: usage,
        Main: Main,
    })
}

const usage = `Usage: foo [OPTION]... [FILE]...

Description.
`

func Main(argv []string) int {
    // implementation
    return 0
}
```

Then add one line to `cmd/topsail/main.go`:

```go
import _ "github.com/Real-Fruit-Snacks/topsail/applets/foo"
```

That's the entire wiring. Tests go in `applets/foo/foo_test.go` using `internal/testutil.CaptureStdio` for stdio capture; aim for ≥85% statement coverage on every applet package. See `CONTRIBUTING.md` for style and lint conventions.
