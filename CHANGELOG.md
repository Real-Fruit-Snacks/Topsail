# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-04-26

### Added

- **Linux packages**: each release now ships `.deb`, `.rpm`, and `.apk` archives alongside the tarballs, generated via goreleaser's `nfpms` block. Packages install the binary to `/usr/bin/topsail`, man pages under `/usr/share/man/man1/`, and shell completions to the canonical bash / zsh / fish locations.
- **Shell completions** for bash, zsh, and fish bundled into every release archive (under `completions/`) and into the Linux packages.
- **Man pages** for `topsail(1)` and one per applet, generated from each applet's Usage string. Bundled into archives (under `man/`) and Linux packages (under `/usr/share/man/man1/`).
- New `cmd/gendocs` tool emits the completions and man pages; runs as a goreleaser `before:` hook so artifacts stay in sync with the registry.
- New `internal/applets` package centralises the blank-import list so `cmd/topsail` and `cmd/gendocs` share a single source of truth for the registered applets.

### Changed

- **Asset names dropped the version**: archives are now `topsail_<os>_<arch>.tar.gz` (and `.zip` on Windows) rather than `topsail_<version>_<os>_<arch>.tar.gz`. The `https://github.com/.../releases/latest/download/topsail_linux_amd64.tar.gz` URL now resolves correctly regardless of the latest tag.
- README and Pages quick-start `curl` snippets now use proper case + arch translation. `uname -s` is lowercased and `uname -m` is mapped (`x86_64` → `amd64`, `aarch64` → `arm64`).

### Fixed

- README quick-start `curl` URL was 404-ing because `uname -s` returns `Linux`/`Darwin` (capitalized) and `uname -m` returns `x86_64`/`aarch64`, neither of which matched the goreleaser asset naming.

## [1.0.0] - 2026-04-25

The first stable release. Ships the full **74-applet roster** across six platforms with cosign-signed checksums and per-archive SBOMs.

### Added

#### Foundation (wave 1)

`cat`, `echo`, `true`, `false`, `yes`, `printf`, `pwd`, `basename`, `dirname`, `mkdir`, `rmdir`, `touch`, `mv`, `cp`, `rm` — the smallest correct set of POSIX utilities needed for shell scripts to do useful work.

#### POSIX text utilities (wave 2)

`head`, `tail`, `wc`, `tee`, `tac`, `rev`, `tr`, `cut`, `sort`, `uniq`, `seq`, `sleep`, `expr`, `test` (with `[` alias).

#### Heavy text and filesystem (wave 3)

`grep` (with `egrep`/`fgrep` aliases), `sed`, `awk` (with `gawk`/`nawk` aliases, embedding [benhoyt/goawk](https://github.com/benhoyt/goawk)), `find`, `ls`, `stat`, `file`, `du`, `df`, `chmod`, `chown`, `which`, `xargs`, `date`.

#### Archives and hashing (wave 4)

`tar`, `gzip` / `gunzip`, `zip` / `unzip`, `sha256sum`, `sha512sum`, `md5sum`, `base64`. tar and zip both refuse path-traversal entries during extract.

#### Network and JSON (wave 5)

`curl` (with `wget` alias), `jq` (embedding [itchyny/gojq](https://github.com/itchyny/gojq)), `host` (with `nslookup` alias), `ping`. ping uses TCP probes rather than ICMP — pure-Go raw-socket ICMP requires elevated privileges.

#### Coreutils gap-fillers (wave 6)

`env`, `whoami`, `id`, `hostname`, `uname`, `ln`, `readlink`, `nl`, `paste`, `fold`, `split`, `factor`, `shuf`, `comm`, `join`, `sum`, `column`, `tsort`.

#### Project infrastructure

- Applet contract (`Name`, `Aliases`, `Help`, `Usage`, `Main`) with a thread-safe registry that panics on duplicate names or alias collisions.
- Multi-call dispatcher with `argv[0]` basename match and `.exe`-suffix stripping on Windows.
- Mockable `internal/ioutil.{Stdin,Stdout,Stderr}` for stdio-capture tests.
- Cross-OS `internal/platform` for terminal detection and user/group lookup.
- `internal/testutil` stdio-capture helpers used by every applet test.
- CI matrix: Linux / macOS / Windows × Go {1.25, 1.26}, six-target cross-compile gate.
- Lint stack: golangci-lint v2 (errcheck, gocritic, gosec, govet, ineffassign, misspell, revive, staticcheck, unparam, unused).
- Security stack: govulncheck on every push and weekly cron, CodeQL static analysis.
- Release stack: goreleaser with cross-platform builds, checksums, SBOMs (syft), cosign keyless signing.
- Cross-compilation `Makefile` targets covering linux/{amd64,arm64}, darwin/{amd64,arm64}, windows/{amd64,arm64}.
- Architecture documentation, contributor guide, security policy, code of conduct (Contributor Covenant 2.1).

### Documented divergences from mainsail / POSIX

Captured in detail in [`ARCHITECTURE.md`](ARCHITECTURE.md#documented-divergences-from-mainsail):

- `grep` uses RE2 (Go `regexp`), not POSIX BRE/ERE.
- `sed` ships only the `s/.../.../` substitute command this release.
- `awk` runs with `system()` disabled (`goawk`'s `NoExec = true`).
- `ping` is a TCP probe, not ICMP.
- `chown` on Windows is a stub (`-` for uid/gid, exit 1).
- `basename` / `dirname` use the OS-agnostic `path` package, treating `/` as the separator on every platform.
- `tail -f` follow mode and symbolic file modes (`u+rwx`) are queued for follow-up.

[1.1.0]: https://github.com/Real-Fruit-Snacks/topsail/releases/tag/v1.1.0
[1.0.0]: https://github.com/Real-Fruit-Snacks/topsail/releases/tag/v1.0.0
