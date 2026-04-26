# topsail

[![Go Version](https://img.shields.io/badge/go-1.23%2B-00ADD8.svg)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/ci.yml/badge.svg)](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Real-Fruit-Snacks/topsail.svg)](https://pkg.go.dev/github.com/Real-Fruit-Snacks/topsail)

**Single-file BusyBox-like multi-call binary, in Go.**

`topsail` is the Go sibling of [mainsail](https://github.com/Real-Fruit-Snacks/mainsail). Same applet contract, same dispatch UX, same `--list` / `--help` formatting — but native Go, with the static-binary and cross-compile story Python cannot offer.

> **Status:** under active construction. Wave 0 (project foundation) lands first; the 73-applet roster fills in over six implementation waves. See the [CHANGELOG](CHANGELOG.md) for progress.

## Why a Go port?

- **Fully-static Linux binary.** `CGO_ENABLED=0 go build` produces one binary that runs unchanged on glibc, musl/Alpine, distroless, and scratch — what a Python interpreter structurally cannot do.
- **One-machine cross-compile.** `GOOS=… GOARCH=… go build` for all six platforms, no per-OS CI gymnastics.
- **Microsecond startup.** No interpreter spin-up; pipelines feel native.
- **One artifact story.** Replaces mainsail's binary + zipapp split with a single self-contained executable.

## Quick start

```bash
# Build from source
git clone https://github.com/Real-Fruit-Snacks/topsail
cd topsail
make build
./topsail --list

# Or via go install
go install github.com/Real-Fruit-Snacks/topsail/cmd/topsail@latest
```

## Multi-call dispatch

```bash
topsail ls -la                    # explicit dispatch
ln -s topsail dir && ./dir        # symlink dispatch (Unix)
copy topsail.exe dir.exe & dir    # copy dispatch (Windows)
```

The dispatcher checks `argv[0]`'s basename against the applet registry. Unknown invocations fall through to `topsail <applet> [args]` wrapper mode.

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for the registry mechanism, applet contract, dispatch flow, and documented divergences from mainsail (e.g. RE2 regex semantics).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Adding an applet is one new package under `applets/<name>/` plus one blank import in `cmd/topsail/main.go`.

## License

[MIT](LICENSE).
