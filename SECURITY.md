# Security policy

## Supported versions

`topsail` is a single-binary multi-call utility distributed via tagged releases. Security fixes are issued against the **latest tagged release** on `main`. Older versions are not patched in place; upgrading to the current release is the supported remediation path.

| Version       | Supported          |
| ------------- | ------------------ |
| `main` / latest tag | :white_check_mark: |
| Older tags    | :x:                |

## Reporting a vulnerability

**Please do not file a public GitHub Issue for security problems.** Public disclosure of an unpatched vulnerability puts users at risk.

The supported channel is GitHub Security Advisories — a private, end-to-end-encrypted form that opens a confidential thread between you and the maintainers:

> <https://github.com/Real-Fruit-Snacks/topsail/security/advisories/new>

Include in your report, to the extent you can:

- The affected version (`topsail --version`).
- The OS and architecture you ran on.
- A minimal reproduction (a command line, an input file, an environment).
- The observed unsafe behavior — what's exposed, corrupted, or escalated.
- Any suggested fix or mitigation.

If you cannot use GitHub Security Advisories, you may instead open a Discussion in the [project's Discussions](https://github.com/Real-Fruit-Snacks/topsail/discussions) titled "Please contact me about a security issue" without details, and a maintainer will follow up to set up a private channel.

## Response targets

| Stage                      | Target            |
| -------------------------- | ----------------- |
| Acknowledgement of report  | within 7 days     |
| Triage decision            | within 14 days    |
| Fix landed for accepted bug | within 60 days   |
| Public advisory + release  | coordinated with reporter |

These are targets, not guarantees. Complex or supply-chain bugs may take longer; we'll keep you informed in the advisory thread.

## Scope

In scope:

- The `topsail` binary itself, on any of its six release platforms.
- Build-and-distribution artifacts: cosign signatures, SBOMs, checksums in release attachments.
- The applet dispatcher (registry, argv[0] basename matching, multi-call symlink/copy resolution).
- Vulnerabilities introduced by topsail's own code or by how topsail wires its third-party dependencies.

Out of scope:

- CVEs disclosed against an upstream dependency (`golang.org/x/term`, `benhoyt/goawk`, `itchyny/gojq`, etc.) **before** topsail has had a reasonable window to bump the dependency. Please report those upstream first; we will pick up patched versions on the next release cycle.
- Issues that require a privileged attacker to already have write access to the binary on disk, the user's PATH, or the source tree being built — those are operating-system-level concerns.

## Hardening notes

A few design choices worth being explicit about:

- **No `system()` from awk.** The `awk` applet runs the embedded interpreter with `interp.Config.NoExec = true`, which disables the `system()` function and pipe-into-shell forms. Awk programs cannot fork arbitrary commands.
- **Path-traversal refusal.** `tar -x` and `unzip` reject archive entries whose cleaned path is `..` or starts with `../`. Tar bombs with absolute or escaping paths terminate extraction.
- **Pinned tooling in CI.** `goimports` and `govulncheck` are pinned to specific versions in `.github/workflows/`. Random upstream releases that bump their go directive cannot break our matrix without an explicit PR.
- **Reproducible builds.** Releases are built with `-trimpath -ldflags="-s -w"` and signed via cosign keyless OIDC; `goreleaser` records SBOMs alongside each archive. The binary's version/commit/date fields come from `-ldflags -X` and are visible in `topsail --version`.

## Acknowledgements

Reporters who follow this process will be credited (with their permission) in the published advisory and the `CHANGELOG.md` entry for the release that contains the fix.
