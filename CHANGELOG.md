# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project scaffold: applet contract, registry, dispatcher, CLI harness.
- Wave 1: 15 foundation applets (cat, echo, true, false, yes, printf, pwd, basename, dirname, mkdir, rmdir, touch, mv, cp, rm) plus `internal/testutil` for stdio capture.
- CI matrix across Linux / macOS / Windows × Go {1.25, 1.26}.
- Lint stack: `golangci-lint` (govet, staticcheck, errcheck, revive, gosec, gocritic, misspell, unparam, ineffassign, unused).
- Security stack: `govulncheck` in CI plus weekly cron, CodeQL static analysis.
- Release stack: `goreleaser` with cross-platform builds, checksums, SBOM, and cosign keyless signing.
- Cross-compilation `Makefile` target covering linux/{amd64,arm64}, windows/{amd64,arm64}, darwin/{amd64,arm64}.
- Architecture documentation, contributor guide, security policy, code of conduct.

### Changed

- Minimum Go version raised from 1.23 to 1.25 to pick up the os-package fix for [GO-2026-4602](https://pkg.go.dev/vuln/GO-2026-4602) ("FileInfo can escape from a Root in os").

[Unreleased]: https://github.com/Real-Fruit-Snacks/topsail/compare/HEAD...HEAD
