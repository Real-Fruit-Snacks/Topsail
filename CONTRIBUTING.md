# Contributing to topsail

Thank you for your interest in `topsail`. This document describes how to set up a working environment, the project's conventions, and how to land a change cleanly.

## Code of conduct

Participation is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it.

## Getting set up

You need:

- **Go 1.23 or later** — `go.mod` declares the minimum.
- **make** — used by the per-wave protocol and most local commands.
- **goimports**, **golangci-lint v2**, **govulncheck** — installed via `go install` from `go install golang.org/x/tools/cmd/goimports@latest`, `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`, and `golang.org/x/vuln/cmd/govulncheck@latest`. Make sure `~/go/bin` (or `%USERPROFILE%\go\bin` on Windows) is on your `PATH`.

Clone and verify:

```bash
git clone https://github.com/Real-Fruit-Snacks/topsail
cd topsail
make ci
```

`make ci` runs the same gauntlet that gates every commit — `go mod tidy` drift check, format check, `go vet`, `golangci-lint run`, `go test -race -cover`, `govulncheck`, `go build`, and a six-target cross-compile.

## Adding an applet

Each applet lives in its own package under `applets/<name>/` and self-registers in `init()`. The smallest applet skeleton:

```go
// applets/foo/foo.go
package foo

import "github.com/Real-Fruit-Snacks/topsail/internal/applet"

func init() {
    applet.Register(applet.Applet{
        Name:    "foo",
        Aliases: []string{"f"},
        Help:    "do the foo thing",
        Usage:   usage,
        Main:    Main,
    })
}

const usage = `Usage: foo [OPTION]... [FILE]...

Description.

Options:
  -x   short description.
`

// Main is the applet entry point. argv[0] is the invocation name.
func Main(argv []string) int {
    // implementation
    return 0
}
```

Then add one line to `cmd/topsail/main.go`:

```go
import _ "github.com/Real-Fruit-Snacks/topsail/applets/foo"
```

That's the entire wiring. The dispatcher will pick it up via `topsail foo …`, via multi-call symlinks, and via `--list` / `--help`.

### Tests for the applet

Drop `applets/foo/foo_test.go` next to it. Use `internal/ioutil` to capture stdio:

```go
package foo

import (
    "bytes"
    "testing"

    "github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func TestMain(t *testing.T) {
    out := &bytes.Buffer{}
    orig := ioutil.Stdout
    ioutil.Stdout = out
    t.Cleanup(func() { ioutil.Stdout = orig })

    rc := Main([]string{"foo"})
    if rc != 0 {
        t.Fatalf("rc = %d; want 0", rc)
    }
}
```

We aim for ≥85 % coverage on every applet package.

### Behavioral parity with mainsail

When porting an applet from `mainsail`, treat the Python implementation as the spec. Mirror flag names, error messages, and exit codes unless there is a documented reason to diverge. Document any divergence in [ARCHITECTURE.md](ARCHITECTURE.md) under "Divergences from mainsail."

## Coding style

- **Format:** `gofmt` + `goimports` with `-local github.com/Real-Fruit-Snacks/topsail`. CI fails on any drift.
- **Lint:** `golangci-lint run` must pass cleanly. The configuration enables a curated set; do not silence warnings — fix them.
- **Comments on exported identifiers:** required (revive enforces this).
- **Errors:** wrap with `fmt.Errorf("…: %w", err)` when adding context. Never discard with `_`.
- **No backward-compat shims for unreleased APIs.** Refactor freely.
- **Deps:** minimize. New direct dependencies need a comment in the PR explaining why we did not write the code ourselves.

## Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Examples:

```
feat: add wc applet
fix(grep): preserve trailing newline on -c output
docs(architecture): note RE2 divergence from POSIX BRE
chore(deps): bump goawk to v1.30.0
```

Subject line ≤ 72 characters, imperative mood. Use the body for the *why* — the diff already shows the *what*.

## Pull requests

- One logical change per PR. Don't bundle a refactor with a fix.
- Reference any related issues in the body.
- All CI checks must be green before merge.
- Squash-merge with a Conventional Commits subject is the preferred merge mode.

## Reporting bugs

Use the bug report issue template. Include:

- topsail version (`topsail --version`),
- OS and architecture,
- a minimal reproduction, and
- expected vs. observed behavior.

## Security issues

**Do not file public issues for security vulnerabilities.** See [SECURITY.md](SECURITY.md) for the private reporting process.
