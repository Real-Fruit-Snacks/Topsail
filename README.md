<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-light.svg">
  <img alt="Topsail" src="https://raw.githubusercontent.com/Real-Fruit-Snacks/topsail/main/docs/assets/logo-dark.svg" width="100%">
</picture>

> [!IMPORTANT]
> **A BusyBox-style multi-call binary in Go** — 84 Unix utilities, one ~3 MB statically-linked executable, native on Linux, macOS, and Windows. Sister project to [`mainsail`](https://github.com/Real-Fruit-Snacks/mainsail) (Python) with the same applet roster, same flag conventions, same exit codes.

> *A topsail rides above the mainsail — captures wind the lower sail can't reach. Felt fitting for a Go port that runs in the static-binary niche where Python can't go: distroless, scratch, Alpine.*

---

## §1 / Premise

[`mainsail`](https://github.com/Real-Fruit-Snacks/mainsail) is the Python reference — easy to embed, easy to extend, but it always ships an interpreter or a zipapp that needs one. Topsail covers the cases mainsail can't reach: **fully-static Linux binaries** that run on Alpine (musl), distroless, and scratch base images; **microsecond startup** with no interpreter cold-start; **one-machine cross-compile** to all six platforms in a single CI job; **one artifact** per platform instead of binary + zipapp.

Same applet roster, same flag conventions, same exit codes. **Cosign-signed releases with SBOMs**, keyless OIDC via GitHub Actions. The CI suite runs **710 unit tests** plus an integration harness that builds and exercises every applet on Linux, macOS, and Windows on Go 1.25 and 1.26.

---

## §2 / Specs

| KEY      | VALUE                                                                        |
|----------|------------------------------------------------------------------------------|
| BINARY   | One **~3 MB statically-linked executable** · `CGO_ENABLED=0` · runs on scratch |
| APPLETS  | **84 POSIX utilities** + `jq` (gojq) + `awk` (goawk) + `curl` + `host` + `ping` |
| BUILDS   | **6 platform/arch combos** + `.deb` / `.rpm` / `.apk` per release            |
| RELEASES | **Cosign Sigstore bundle** + syft SBOM per archive · keyless OIDC            |
| TESTS    | **710 unit tests** + integration harness · Linux / macOS / Windows · Go 1.25 + 1.26 |
| STACK    | Go **1.25+** · `goawk` · `gojq` · zero CGO · golangci-lint v2 + govulncheck   |

Per-applet status in the [`--list`](https://real-fruit-snacks.github.io/topsail/) output. Architecture in §5 below.

---

## §3 / Quickstart

```bash
# From a release — no Go toolchain required
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -fL -o topsail.tar.gz \
  "https://github.com/Real-Fruit-Snacks/topsail/releases/latest/download/topsail_${OS}_${ARCH}.tar.gz"
tar xzf topsail.tar.gz
./topsail --list
```

```bash
# Linux distro packages — installs binary, man pages, shell completions
sudo dpkg -i topsail_<arch>.deb                       # Debian / Ubuntu
sudo rpm -i topsail_<arch>.rpm                        # Fedora / RHEL / openSUSE
sudo apk add --allow-untrusted topsail_<arch>.apk     # Alpine
```

```bash
# From source — Go 1.25+
git clone https://github.com/Real-Fruit-Snacks/topsail && cd topsail
make build
./topsail --list

# Or via go install
go install github.com/Real-Fruit-Snacks/topsail/cmd/topsail@latest
```

```bash
# Multi-call dispatch — symlink any applet name
topsail ls -la                              # explicit dispatch
ln -s topsail ls && ./ls -la                # symlink dispatch (Unix)
copy topsail.exe ls.exe & ls -la            :: copy dispatch (Windows)
```

---

## §4 / Reference

```
APPLET CATEGORIES                                       # 84 total

  FOUNDATION    cat echo true false yes printf pwd basename dirname
                mkdir rmdir touch mv cp rm
  POSIX TEXT    head tail wc tee tac rev tr cut sort uniq seq sleep
                expr test ([) expand unexpand
  HEAVY TEXT    grep (egrep/fgrep) sed awk (gawk/nawk) find ls stat
                file du df chmod chown which xargs date realpath
                truncate unlink mktemp
  ARCHIVES      tar gzip gunzip zip unzip
  HASHING       sha256sum sha512sum md5sum base64 xxd
  NETWORK       curl (wget) jq host (nslookup) ping
  PROCESS       time timeout nproc
  GAP-FILLERS   env whoami id hostname uname ln readlink nl paste fold
                split factor shuf comm join sum column tsort

DISPATCH

  topsail <applet> [args]                  # subcommand form
  ln -s topsail <applet>                   # multi-call: argv[0] basename
  copy topsail.exe <applet>.exe            # Windows: .exe stripped before lookup

RELEASE BINARIES                                        # 6 per release tag

  Linux x86_64                             topsail_linux_amd64.tar.gz
  Linux ARM64                              topsail_linux_arm64.tar.gz
  macOS Intel                              topsail_darwin_amd64.tar.gz
  macOS Apple Silicon                      topsail_darwin_arm64.tar.gz
  Windows x86_64                           topsail_windows_amd64.zip
  Windows ARM64                            topsail_windows_arm64.zip
                                           + .deb / .rpm / .apk per arch

NOTABLE FLAG SUPPORT
  find          predicate tree · -exec · -prune · -and/-or · parens
                size/time predicates · -delete · -print0
  sed           s/// d p q · addresses · ranges · negation
                multi-command -e · BRE/ERE
  awk           BEGIN/END · regex + expression patterns · arrays
                length/substr/split/sub/gsub/match (goawk, NoExec=true)
  jq            pipes · comma · alternatives · constructors · slices
                iterators · -r raw · -c compact · -s slurp (gojq)
  sort          -k FIELD[,FIELD][OPTS] repeatable · -t SEP
                numeric/reverse/unique/case-folding
  tail          -f follow · poll-based growth · truncation handling
                multi-file headers · --sleep-interval=N
  chmod / mkdir -m   octal + POSIX symbolic (u+x, go-w, a=rwx,u+s)
  xxd           canonical + plain hex dump · -r revert · round-trips
  tar           create / extract / list · gzip filter · path-traversal refuse
  truncate      -s [+|-|<|>]SIZE · K/M/G/T (1024) + KB/MB/GB/TB (1000)

VERIFICATION
  cosign verify-blob \                     # keyless OIDC, Rekor inclusion proof
    --bundle checksums.txt.sigstore.json \
    --certificate-identity-regexp 'https://github.com/Real-Fruit-Snacks/topsail/.*' \
    --certificate-oidc-issuer    https://token.actions.githubusercontent.com \
    checksums.txt
  sha256sum -c checksums.txt --ignore-missing

DEVELOPMENT
  make build                               Build the binary
  make ci                                  Full local mirror of CI
  make test / test-race / fuzz             Test suite + race detector + fuzz
  make lint                                golangci-lint v2
  make vuln                                govulncheck
  make cross                               Cross-compile to all six targets
```

---

## §5 / Architecture

```
cmd/topsail/main.go        process entry → cli.Run → exit code
   │
   ▼
internal/cli               argv[0] basename match · multi-call vs wrapper mode
   │
   ▼
internal/applet            registry: Register · Get · All · sync.RWMutex
   │
   ▼
applets/<name>/            one package per applet · init() registers · Main entry
```

**Four-layer flow:** `cmd/topsail/main.go` imports `internal/applets` for side-effect registration, then calls `cli.Run(os.Args)`. Same path whether invoked directly, via `topsail <applet>`, or through a symlink. `internal/cli` matches `argv[0]`'s basename against the registry; `.exe` is stripped on Windows. Unknown invocations fall through to wrapper mode. `internal/applet` exposes `Register`, `Get`, `All` guarded by a `sync.RWMutex` — duplicate names panic at startup so collisions surface immediately. 74 packages under `applets/`, each implementing `Name`, `Aliases`, `Help`, `Usage`, `Main(argv) -> int`. Stdio reads/writes go through `internal/ioutil` package-level vars so tests swap in `bytes.Buffer`s directly.

**Documented divergences:** `grep` uses RE2 (not POSIX BRE/ERE); `awk` runs with `system()` disabled (NoExec=true) — pipelines via `|` blocked; `ping` is TCP not ICMP (raw sockets need privileges); `chown` on Windows is a stub; `sed` lacks `y///`, hold space, `-i` (queued); `sort` lacks character offsets in `-k`. Full list in [`ARCHITECTURE.md`](ARCHITECTURE.md).

---

[License: MIT](LICENSE) · [Architecture](ARCHITECTURE.md) · [Contributing](CONTRIBUTING.md) · [Changelog](CHANGELOG.md) · Part of [Real-Fruit-Snacks](https://github.com/Real-Fruit-Snacks) — building offensive security tools, one wave at a time. Sibling: [mainsail](https://github.com/Real-Fruit-Snacks/mainsail) (Python) · [jib](https://github.com/Real-Fruit-Snacks/jib) (Rust) · [staysail](https://github.com/Real-Fruit-Snacks/Staysail) (Zig) · [moonraker](https://github.com/Real-Fruit-Snacks/Moonraker) (Lua) · [rill](https://github.com/Real-Fruit-Snacks/rill) (NASM).
