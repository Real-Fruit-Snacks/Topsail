SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

MODULE  := github.com/Real-Fruit-Snacks/topsail
PKG     := ./...
BIN     := topsail
DIST    := dist

PLATFORMS := \
	linux/amd64   \
	linux/arm64   \
	darwin/amd64  \
	darwin/arm64  \
	windows/amd64 \
	windows/arm64

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(MODULE)/internal/cli.Version=$(VERSION) \
	-X $(MODULE)/internal/cli.Commit=$(COMMIT) \
	-X $(MODULE)/internal/cli.Date=$(DATE)

GO            ?= go
GOIMPORTS     ?= goimports
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK   ?= govulncheck

.PHONY: help
help: ## Show this help
	@awk 'BEGIN{FS=":.*?## "} /^[a-zA-Z_-]+:.*?## /{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary for the host platform
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/topsail

.PHONY: install
install: ## Install the binary into $$GOPATH/bin
	CGO_ENABLED=0 $(GO) install -trimpath -ldflags '$(LDFLAGS)' ./cmd/topsail

.PHONY: test
test: ## Run tests with coverage
	$(GO) test -cover $(PKG)

.PHONY: test-race
test-race: ## Run tests with -race and coverage
	$(GO) test -race -cover $(PKG)

.PHONY: cover
cover: ## Generate coverage profile and HTML report
	$(GO) test -race -coverprofile=coverage.txt -covermode=atomic $(PKG)
	$(GO) tool cover -html=coverage.txt -o coverage.html

.PHONY: bench
bench: ## Run benchmarks
	$(GO) test -run=^$$ -bench=. -benchmem $(PKG)

.PHONY: fuzz
fuzz: ## Run fuzz tests for 30s each
	@for pkg in $$($(GO) list $(PKG)); do \
		for fn in $$($(GO) test -list '^Fuzz' $$pkg 2>/dev/null | grep '^Fuzz' || true); do \
			echo ">> fuzzing $$pkg::$$fn"; \
			$(GO) test -run=^$$ -fuzz=$$fn -fuzztime=30s $$pkg || exit 1; \
		done; \
	done

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run

.PHONY: fmt
fmt: ## Format Go sources with goimports
	$(GOIMPORTS) -w -local $(MODULE) .

.PHONY: fmt-check
fmt-check: ## Verify formatting (CI gate)
	@diff=$$($(GOIMPORTS) -l -local $(MODULE) .); \
	if [ -n "$$diff" ]; then \
		echo "Files need formatting (run 'make fmt'):"; \
		echo "$$diff"; \
		exit 1; \
	fi

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)

.PHONY: tidy
tidy: ## Run go mod tidy
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Verify go.mod / go.sum are tidy (CI gate)
	@cp go.mod go.mod.bak; \
	[ -f go.sum ] && cp go.sum go.sum.bak || true; \
	$(GO) mod tidy; \
	drift=0; \
	diff -q go.mod go.mod.bak >/dev/null 2>&1 || drift=1; \
	if [ -f go.sum.bak ]; then \
		diff -q go.sum go.sum.bak >/dev/null 2>&1 || drift=1; \
	fi; \
	mv go.mod.bak go.mod; \
	[ -f go.sum.bak ] && mv go.sum.bak go.sum || true; \
	if [ $$drift -ne 0 ]; then \
		echo "go.mod / go.sum drift; run 'make tidy'"; \
		exit 1; \
	fi

.PHONY: vuln
vuln: ## Run govulncheck
	$(GOVULNCHECK) $(PKG)

.PHONY: cross
cross: ## Cross-compile for all release platforms
	@mkdir -p $(DIST)
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		out=$(DIST)/$(BIN)-$$os-$$arch; \
		[ "$$os" = "windows" ] && out=$$out.exe; \
		echo ">> $$os/$$arch -> $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $$out ./cmd/topsail; \
	done

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BIN) $(BIN).exe coverage.txt coverage.html
	rm -rf $(DIST) build

.PHONY: ci
ci: tidy-check fmt-check vet lint test-race vuln build cross ## Full CI gauntlet (matches .github/workflows/ci.yml)
