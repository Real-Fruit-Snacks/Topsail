<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-light.svg">
  <img alt="topsail" src="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-dark.svg" width="620">
</picture>

![Go](https://img.shields.io/badge/language-Go-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)
![Arch](https://img.shields.io/badge/arch-x86__64%20%7C%20ARM64-blue)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Tests](https://img.shields.io/badge/tests-710%20passing-brightgreen.svg)

A BusyBox-style multi-call binary in Go — **84 Unix utilities**, one ~3 MB statically-linked executable, native on Linux, macOS, and Windows.

[Download Latest](https://github.com/Real-Fruit-Snacks/topsail/releases/latest)
&nbsp;·&nbsp;
[GitHub Pages](https://real-fruit-snacks.github.io/topsail/)
&nbsp;·&nbsp;
[Changelog](CHANGELOG.md)
&nbsp;·&nbsp;
[Sibling: mainsail (Python)](https://github.com/Real-Fruit-Snacks/mainsail)

</div>

---

## Quick start

**From a release** — no Go required:

```bash
# Linux / macOS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -fL -o topsail.tar.gz \
  "https://github.com/Real-Fruit-Snacks/topsail/releases/latest/download/topsail_${OS}_${ARCH}.tar.gz"
tar xzf topsail.tar.gz
./topsail --list
```

**From a Linux distro package** — `.deb`, `.rpm`, and `.apk` ship in every release:

```bash
sudo dpkg -i topsail_<arch>.deb                       # Debian / Ubuntu
sudo rpm -i topsail_<arch>.rpm                        # Fedora / RHEL / openSUSE
sudo apk add --allow-untrusted topsail_<arch>.apk     # Alpine
```

Each package installs the binary to `/usr/bin/topsail`, man pages under `/usr/share/man/man1/`, and shell completions for bash / zsh / fish in their canonical locations.

**From source** — Go 1.25+:

```bash
git clone https://github.com/Real-Fruit-Snacks/topsail
cd topsail
make build
./topsail --list
```

**Or via `go install`** — straight from the module path:

```bash
go install github.com/Real-Fruit-Snacks/topsail/cmd/topsail@latest
```

**Multi-call dispatch:**

```bash
topsail ls -la                    # explicit dispatch
ln -s topsail ls && ./ls -la      # symlink dispatch (Unix)
copy topsail.exe ls.exe & ls -la  # copy dispatch (Windows)
```

The dispatcher checks `argv[0]`'s basename against the applet registry. Unknown invocations fall through to `topsail <applet> [args]` wrapper mode. `.exe` suffixes are stripped on Windows before lookup.

---

## Pre-built binaries

Each release ships six platform/architecture combinations, plus a checksums file and a cosign Sigstore bundle:

| Platform | Architecture | Archive                              |
| -------- | ------------ | ------------------------------------ |
| Linux    | x86_64       | `topsail_linux_amd64.tar.gz`         |
| Linux    | ARM64        | `topsail_linux_arm64.tar.gz`         |
| macOS    | Intel        | `topsail_darwin_amd64.tar.gz`        |
| macOS    | Apple Silicon| `topsail_darwin_arm64.tar.gz`        |
| Windows  | x86_64       | `topsail_windows_amd64.zip`          |
| Windows  | ARM64        | `topsail_windows_arm64.zip`          |

Linux binaries are built with `CGO_ENABLED=0` and are fully static — they run unchanged on Alpine (musl), distroless, and scratch base images.

### Verifying signatures

Each release includes a `checksums.txt` plus a cosign Sigstore bundle (`checksums.txt.sigstore.json`) that combines the signature, certificate, and Rekor inclusion proof. Verify with cosign v3+:

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/Real-Fruit-Snacks/topsail/.*' \
  --certificate-oidc-issuer    https://token.actions.githubusercontent.com \
  checksums.txt
```

Then verify your archive:

```bash
sha256sum -c checksums.txt --ignore-missing
```

---

## Features

### One static binary, eighty-four utilities

Every common POSIX tool you'd reach for in a shell pipeline — plus `jq` for JSON, `curl` for HTTP, `host` for DNS, `ping` for reachability, archives (`tar`/`gzip`/`zip`), hashing (`sha256sum`/`sha512sum`/`md5sum`), and the BusyBox parity gap-fillers. Dispatch via `topsail <applet>` or symlink/copy to call the applet directly.

```bash
topsail ls -la                              # GNU-style flags
topsail cat file.txt | topsail grep -C 2 pattern
topsail find . -name '*.go' -size +1k -mtime -7
topsail seq 100 | topsail sort -rn | topsail head -5
```

### Fully static Linux binaries

`CGO_ENABLED=0 go build` produces a binary that runs unchanged on glibc, musl/Alpine, distroless, and scratch base images. No interpreter, no shared libs, no runtime dependencies.

```dockerfile
FROM gcr.io/distroless/static
COPY topsail /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/topsail"]
```

### Real applets, not stubs

Each applet implements the common POSIX flags and edge cases.

- `find` — full predicate tree with `-exec`, `-prune`, `-and`/`-or`, parens, size/time predicates, `-delete`, `-print0`
- `sed` — `s/.../.../[gi]`, `d` (delete), `p` (print), `q` (quit), addresses (`N`, `$`, `/re/`), ranges (`addr1,addr2`), negation (`addr!cmd`), multi-command scripts (`;` or newline), repeatable `-e EXPR`
- `awk` — embeds [`benhoyt/goawk`](https://github.com/benhoyt/goawk) with `system()` disabled: BEGIN/END, regex and expression patterns, range patterns, full control flow, associative arrays, `length`/`substr`/`index`/`split`/`sub`/`gsub`/`match`/`toupper`/`tolower`/`sprintf`
- `jq` — embeds [`itchyny/gojq`](https://github.com/itchyny/gojq): pipes, comma, alternatives, comparison/arithmetic, object & array constructors, slices and iterators, `if`/`then`/`elif`/`else`/`end`, full built-in library, raw output (`-r`), compact (`-c`), slurp (`-s`)
- `sort` — `-k FIELD[,FIELD][OPTS]` repeatable keys with per-key option suffixes (`-k 2nr`), `-t SEP` custom separator, numeric/reverse/unique/case-folding
- `tail` — follow mode (`-f`) with poll-based file growth detection, file-truncation handling, multi-file with headers, configurable poll cadence (`--sleep-interval=N`)
- `chmod` / `mkdir -m` — both octal and POSIX symbolic modes (`u+x`, `go-w`, `a=rwx,u+s`, `g=u`, `+X`) via the shared `internal/filemode` parser
- `xxd` — canonical and plain hex dump, plus `-r` revert; `-c COLS`, `-g GROUP`, `-u` uppercase. Round-trips its own canonical output.
- `tar` — create/extract/list with gzip filter; refuses path-traversal entries during extract
- `truncate` — `-s [+|-|<|>]SIZE` with K/M/G/T (1024) and KB/MB/GB/TB (1000) suffixes; auto-creates files unless `-c`

```bash
topsail find . -name '*.tmp' -delete
topsail sed -e '/^#/d' -e 's/foo/bar/g' config.txt
topsail awk -F, '{s+=$3} END{print s/NR}' data.csv
topsail jq '.servers[] | select(.region == "us") | .name' inventory.json
topsail sort -t : -k 3,3 -n /etc/passwd
topsail xxd -r -p <<< 48656c6c6f0a    # → "Hello"
```

### Pipeline-grade I/O

Binary-safe through `cat`/`tee`/`gzip`. CRLF survives Windows text-mode round-trips. `tail -f` follows files and detects truncation. `xargs` accepts `-print0`/`-0` to handle paths with whitespace and backslashes.

```bash
topsail find . -type f -print0 | topsail xargs -0 topsail sha256sum
topsail tail -f /var/log/app.log
topsail gzip -c data.bin | topsail gunzip > data.bin.copy
```

### Cosign-signed releases with SBOMs

Each release ships a `checksums.txt.sigstore.json` Sigstore bundle (signature + certificate + Rekor inclusion proof in one file) plus a syft-generated SBOM per archive. Keyless OIDC via GitHub Actions — no key management.

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/Real-Fruit-Snacks/topsail/.*' \
  --certificate-oidc-issuer    https://token.actions.githubusercontent.com \
  checksums.txt
```

### Cross-platform integrity

Same SHA-256 of `"abc"` (`ba7816bf…015ad`) on every supported platform. `tar` archives are interchangeable. The CI suite runs **710 unit tests** plus an integration harness that builds the binary and exercises real applets through subprocess invocation — across Linux, macOS, and Windows on Go 1.25 and 1.26.

---

## Supported applets

| Category               | Applets                                                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Foundation**         | `cat` `echo` `true` `false` `yes` `printf` `pwd` `basename` `dirname` `mkdir` `rmdir` `touch` `mv` `cp` `rm`              |
| **POSIX text**         | `head` `tail` `wc` `tee` `tac` `rev` `tr` `cut` `sort` `uniq` `seq` `sleep` `expr` `test` (also `[`) `expand` `unexpand`  |
| **Heavy text / FS**    | `grep` (also `egrep`/`fgrep`) `sed` `awk` (also `gawk`/`nawk`) `find` `ls` `stat` `file` `du` `df` `chmod` `chown` `which` `xargs` `date` `realpath` `truncate` `unlink` `mktemp` |
| **Archives & hashing / encoding** | `tar` `gzip` `gunzip` `zip` `unzip` `sha256sum` `sha512sum` `md5sum` `base64` `xxd`                            |
| **Network & JSON**     | `curl` (also `wget`) `jq` `host` (also `nslookup`) `ping`                                                                  |
| **Process & system**   | `time` `timeout` `nproc`                                                                                                  |
| **Coreutils gap-fillers** | `env` `whoami` `id` `hostname` `uname` `ln` `readlink` `nl` `paste` `fold` `split` `factor` `shuf` `comm` `join` `sum` `column` `tsort` |

84 registered names across 74 packages. `awk` and `jq` embed [`benhoyt/goawk`](https://github.com/benhoyt/goawk) and [`itchyny/gojq`](https://github.com/itchyny/gojq) respectively; everything else is built on the Go standard library.

Run `topsail --list` for the full set with one-line descriptions, or `topsail <applet> --help` for per-applet usage and flags.

<details>
<summary><strong>Documented divergences from POSIX / mainsail</strong></summary>

Captured in detail in [`ARCHITECTURE.md`](ARCHITECTURE.md#documented-divergences-from-mainsail):

- **`grep`** uses RE2 (Go `regexp`), not POSIX BRE/ERE. Most patterns work; back-references and some POSIX classes do not.
- **`awk`** runs with `system()` disabled (`goawk`'s `NoExec = true`) — pipelines via `|` and `getline` from external commands are blocked.
- **`ping`** is a TCP probe rather than ICMP. Pure-Go raw-socket ICMP requires elevated privileges; TCP works unprivileged on every platform.
- **`chown`** on Windows is a stub (`-` for uid/gid, exits 1) — Windows ACLs don't map cleanly to Unix uid/gid pairs.
- **`basename`** / **`dirname`** use the OS-agnostic `path` package, treating `/` as the separator on every platform.
- **`sed`** does not support `y/.../.../` transliteration, label/branch (`b`/`t`/`:`), hold space (`h`/`H`/`g`/`G`), or in-place edit (`-i`) — those are queued for a future wave.
- **`sort`** does not support character offsets in `-k` (e.g. `-k 1.3,1.5`).

Every divergence is intentional and documented; none change the exit codes for the supported flag set.

</details>

---

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

**Four-layer flow:**

1. **Entry** — `cmd/topsail/main.go` imports `internal/applets` for side-effect registration, then calls `cli.Run(os.Args)`. Same path whether invoked directly, via `topsail <applet>`, or through a symlink.
2. **Dispatch** — `internal/cli` matches `argv[0]`'s basename against the registry (multi-call mode); `.exe` suffix is stripped on Windows. Unknown invocations fall through to wrapper mode (`topsail <applet> [args]`). `--help` is intercepted long-form only — `-h` stays free for applet flags like `df -h`.
3. **Registry** — `internal/applet` exposes `Register`, `Get`, and `All`, guarded by a `sync.RWMutex`. Duplicate names or alias collisions panic at startup so every conflict surfaces immediately.
4. **Applets** — 74 packages under `applets/`, each implementing `Name`, `Aliases`, `Help`, `Usage`, and `Main(argv) -> int`. Stdio reads/writes go through `internal/ioutil` package-level vars so tests swap in `bytes.Buffer`s directly.

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for the full applet contract, dispatch flow, cross-platform shims, build identification (`-ldflags`), and the divergence list.

---

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
$EDITOR internal/applets/all.go    # add one blank import line
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
make fuzz      # go test -fuzz=. -fuzztime=30s on every Fuzz target
make vuln      # govulncheck
make cross     # cross-compile to all six release targets
make ci        # everything above, in CI order
```

---

## Why a Go port of mainsail?

[mainsail](https://github.com/Real-Fruit-Snacks/mainsail) is great when you want a single Python file you can `chmod +x` and run on any box that has CPython. topsail covers the cases mainsail can't:

- **Fully-static Linux binary.** No interpreter, no shared libs. Works on Alpine (musl), distroless, scratch — none of which can host a Python zipapp.
- **One-machine cross-compile.** `GOOS=… GOARCH=… go build` for all six platforms in one job; mainsail needs per-OS CI runners for native binaries.
- **Microsecond startup.** No interpreter cold-start. Pipelines that spawn 1000 instances feel native.
- **One artifact story.** Replaces mainsail's binary + zipapp split with a single self-contained executable.

The two share the same applet roster, the same flag conventions, and the same exit codes. Pick mainsail when you want to embed your own Python in a custom build; pick topsail when you want one binary that runs everywhere.

---

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the contributor guide. Adding an applet is one new package under `applets/<name>/` plus one blank import in `internal/applets/all.go`.

Reporting a security vulnerability? Use [GitHub Security Advisories](https://github.com/Real-Fruit-Snacks/topsail/security/advisories/new) — see [`SECURITY.md`](SECURITY.md).

This project follows the [Contributor Covenant 2.1](CODE_OF_CONDUCT.md) Code of Conduct.

---

## License

[MIT](LICENSE).

`topsail` is the Go sibling in the same family as [mainsail](https://github.com/Real-Fruit-Snacks/mainsail). Same applet contract, same dispatch UX, different runtime story. Part of the Real-Fruit-Snacks water-themed toolkit.
