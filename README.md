<p align="center">
  <img src="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-dark.svg" alt="topsail" width="620">
</p>

# topsail

[![Go Version](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-blue.svg)](#pre-built-binaries)
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-lightgrey.svg)](#pre-built-binaries)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/ci.yml/badge.svg)](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/codeql.yml/badge.svg)](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/codeql.yml)
[![govulncheck](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/Real-Fruit-Snacks/topsail/actions/workflows/govulncheck.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Real-Fruit-Snacks/topsail.svg)](https://pkg.go.dev/github.com/Real-Fruit-Snacks/topsail)

**A BusyBox-style multi-call binary in Go — 74 Unix utilities, one ~3 MB statically-linked executable, native on Linux, macOS, and Windows.**

`topsail` is the Go sibling of [mainsail](https://github.com/Real-Fruit-Snacks/mainsail). Same applet contract, same dispatch UX, same `--list` / `--help` formatting — but compiled to a single static binary that runs unchanged on glibc, musl/Alpine, distroless, and scratch.

## Quick start

### Pre-built binary

Download the appropriate archive from the [latest release](https://github.com/Real-Fruit-Snacks/topsail/releases/latest), verify it (see [verifying signatures](#verifying-signatures)), then drop it on your `PATH`:

```bash
# Linux / macOS
curl -L -o topsail.tar.gz https://github.com/Real-Fruit-Snacks/topsail/releases/latest/download/topsail_$(uname -s)_$(uname -m).tar.gz
tar xzf topsail.tar.gz
./topsail --list
```

### Build from source

```bash
git clone https://github.com/Real-Fruit-Snacks/topsail
cd topsail
make build
./topsail --list
```

Or, if you just want the binary somewhere on your `GOPATH`:

```bash
go install github.com/Real-Fruit-Snacks/topsail/cmd/topsail@latest
```

### Multi-call dispatch

```bash
topsail ls -la                    # explicit dispatch
ln -s topsail ls && ./ls -la      # symlink dispatch (Unix)
copy topsail.exe ls.exe & ls -la  # copy dispatch (Windows)
```

The dispatcher checks `argv[0]`'s basename against the applet registry. Unknown invocations fall through to `topsail <applet> [args]` wrapper mode. `.exe` suffixes are stripped on Windows before lookup.

## Pre-built binaries

Each release ships six platform/architecture combinations, plus a checksums file and cosign keyless signatures:

| Platform | Architecture | Archive                              |
| -------- | ------------ | ------------------------------------ |
| Linux    | x86_64       | `topsail_<version>_Linux_amd64.tar.gz`   |
| Linux    | ARM64        | `topsail_<version>_Linux_arm64.tar.gz`   |
| macOS    | Intel        | `topsail_<version>_Darwin_amd64.tar.gz`  |
| macOS    | Apple Silicon | `topsail_<version>_Darwin_arm64.tar.gz` |
| Windows  | x86_64       | `topsail_<version>_Windows_amd64.zip`    |
| Windows  | ARM64        | `topsail_<version>_Windows_arm64.zip`    |

Linux binaries are built with `CGO_ENABLED=0` and are fully static — they run unchanged on Alpine (musl), distroless, and scratch base images.

### Verifying signatures

Each release includes a `checksums.txt`, a cosign signature (`checksums.txt.sig`), and a certificate (`checksums.txt.pem`). Verify with cosign:

```bash
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature   checksums.txt.sig \
  --certificate-identity-regexp 'https://github.com/Real-Fruit-Snacks/topsail/.*' \
  --certificate-oidc-issuer    https://token.actions.githubusercontent.com \
  checksums.txt
```

Then verify your archive:

```bash
sha256sum -c checksums.txt --ignore-missing
```

## Features

- **One static Linux binary.** `CGO_ENABLED=0 go build` — runs unchanged on glibc, musl/Alpine, distroless, scratch.
- **Six-platform release.** Linux / macOS / Windows × amd64 / arm64. Cross-compiled from a single CI job.
- **Microsecond startup.** No interpreter spin-up; pipelines feel native.
- **Multi-call dispatch.** Symlink `topsail` to any applet name, or copy the binary on Windows. Falls through to `topsail <applet> [args]` for unknown names.
- **Self-registering applets.** Each applet is one Go package under `applets/<name>/` plus one blank import in `cmd/topsail/main.go`. Registry is `sync.RWMutex`-protected and panics loudly on duplicate names.
- **Mockable I/O.** `internal/ioutil` exposes `Stdin/Stdout/Stderr` as package vars; tests swap in `bytes.Buffer`s without subprocess plumbing.
- **Cosign-signed releases with SBOMs.** Each release ships keyless cosign signatures for `checksums.txt` and a syft-generated SBOM per archive.
- **govulncheck on every push and weekly cron.** CodeQL static analysis on top of `golangci-lint` v2 (errcheck, gocritic, gosec, govet, revive, staticcheck, unparam, unused, ineffassign, misspell).

## Supported applets

| Category               | Applets                                                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Foundation**         | `cat` `echo` `true` `false` `yes` `printf` `pwd` `basename` `dirname` `mkdir` `rmdir` `touch` `mv` `cp` `rm`              |
| **POSIX text**         | `head` `tail` `wc` `tee` `tac` `rev` `tr` `cut` `sort` `uniq` `seq` `sleep` `expr` `test` (also `[`)                       |
| **Heavy text / FS**    | `grep` (also `egrep`/`fgrep`) `sed` `awk` (also `gawk`/`nawk`) `find` `ls` `stat` `file` `du` `df` `chmod` `chown` `which` `xargs` `date` |
| **Archives & hashing** | `tar` `gzip` `gunzip` `zip` `unzip` `sha256sum` `sha512sum` `md5sum` `base64`                                             |
| **Network & JSON**     | `curl` (also `wget`) `jq` `host` (also `nslookup`) `ping`                                                                  |
| **Coreutils gap-fillers** | `env` `whoami` `id` `hostname` `uname` `ln` `readlink` `nl` `paste` `fold` `split` `factor` `shuf` `comm` `join` `sum` `column` `tsort` |

74 registered names across 64 packages. `awk` and `jq` embed [`benhoyt/goawk`](https://github.com/benhoyt/goawk) and [`itchyny/gojq`](https://github.com/itchyny/gojq) respectively; everything else is built on the Go standard library.

Documented divergences from POSIX and from mainsail are listed in [`ARCHITECTURE.md`](ARCHITECTURE.md#documented-divergences-from-mainsail).

## Architecture

`topsail` boils down to four moving parts:

```
cmd/topsail/main.go        process entry; imports applet packages, calls cli.Run
   │
   ▼
internal/cli                argv[0] basename match; multi-call vs wrapper mode
   │
   ▼
internal/applet             registry: Register, Get, All; thread-safe; panics on dup
   │
   ▼
applets/<name>/             one package per applet; init() registers; Main is the entry
```

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for the full applet contract, dispatch flow, cross-platform shims, build identification (-ldflags), and the documented divergence list.

## Development

### Setting up

```bash
git clone https://github.com/Real-Fruit-Snacks/topsail
cd topsail
make ci    # full local mirror of CI
```

You need:

- **Go 1.25 or later** (the `go.mod` floor; bumped from 1.23 to pick up the [GO-2026-4602](https://pkg.go.dev/vuln/GO-2026-4602) fix)
- **make** (any flavor)
- **goimports**, **golangci-lint v2**, **govulncheck** — installed via `go install` (see [`CONTRIBUTING.md`](CONTRIBUTING.md))

### Adding an applet

```bash
mkdir -p applets/myapplet
$EDITOR applets/myapplet/myapplet.go applets/myapplet/myapplet_test.go
$EDITOR cmd/topsail/main.go    # add one blank import
make ci
```

Full step-by-step in [`CONTRIBUTING.md`](CONTRIBUTING.md). The smallest correct applet is ~25 lines.

### Running the gauntlet

```bash
make fmt       # goimports -w
make vet       # go vet
make lint      # golangci-lint run
make test      # go test -cover
make test-race # go test -race -cover (requires CGO + a C toolchain)
make vuln      # govulncheck
make cross     # cross-compile to all six release targets
make ci        # everything above, in CI order
```

## Why a Go port of mainsail?

mainsail is great when you want a single Python file you can `chmod +x` and run on any box that has CPython. topsail covers the cases mainsail can't:

- **Fully-static Linux binary.** No interpreter, no shared libs. Works on Alpine (musl), distroless, scratch — none of which can host a Python zipapp.
- **One-machine cross-compile.** `GOOS=… GOARCH=… go build` for all six platforms in one job; mainsail needs per-OS CI runners for native binaries.
- **Microsecond startup.** No interpreter cold-start. Pipelines that spawn 1000 instances feel native.
- **One artifact story.** Replaces mainsail's binary + zipapp split with a single self-contained executable.

The two share the same applet roster, the same flag conventions, and the same exit codes. Pick mainsail when you want to embed your own Python in a custom build; pick topsail when you want one binary that runs everywhere.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the contributor guide. Adding an applet is one new package under `applets/<name>/` plus one blank import in `cmd/topsail/main.go`.

Reporting a security vulnerability? Use [GitHub Security Advisories](https://github.com/Real-Fruit-Snacks/topsail/security/advisories/new) — see [`SECURITY.md`](SECURITY.md).

This project follows the [Contributor Covenant 2.1](CODE_OF_CONDUCT.md) Code of Conduct.

## License

[MIT](LICENSE).

## About

`topsail` is the Go sibling in the same family as [mainsail](https://github.com/Real-Fruit-Snacks/mainsail). Same applet contract, same dispatch UX, different runtime story.
