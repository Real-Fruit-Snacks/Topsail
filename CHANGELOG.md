# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project scaffold: applet contract, registry, dispatcher, CLI harness.
- CI matrix across Linux / macOS / Windows × Go {1.23, 1.24}.
- Lint stack: `golangci-lint` (govet, staticcheck, errcheck, revive, gosec, gocritic, misspell, unparam, ineffassign, unused).
- Security stack: `govulncheck` in CI plus weekly cron, CodeQL static analysis.
- Release stack: `goreleaser` with cross-platform builds, checksums, SBOM, and cosign keyless signing.
- Cross-compilation `Makefile` target covering linux/{amd64,arm64}, windows/{amd64,arm64}, darwin/{amd64,arm64}.
- Architecture documentation, contributor guide, security policy, code of conduct.

[Unreleased]: https://github.com/Real-Fruit-Snacks/topsail/compare/HEAD...HEAD
