SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help build install test test-integration bench verify fmt tidy lint tools snapshot clean

.DEFAULT_GOAL := help

# Version from the VERSION file (used by local/snapshot builds).
VERSION ?= $(shell cat VERSION)

# LDFLAGS mirror the goreleaser release build: stripped binary with the
# version injected into internal/cli.version. goreleaser itself adds
# -s -w; we keep them here so local builds behave the same way.
LDFLAGS := -s -w -X github.com/gleicon/devskills/internal/cli.version=$(VERSION)

# Pinned on purpose, and must match release.yml: a new golangci-lint release
# ships new linters that fire under the same .golangci.yml, turning the gate
# red with zero code change. `tools` reinstalls (including downgrade)
# whenever the binary on disk doesn't match exactly.
GOLANGCI_LINT_VERSION ?= v2.11.4

# Invoke the pinned tool by absolute path so PATH ordering (brew, asdf,
# system) can't shadow what `tools` just installed.
GOBIN_DIR := $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN_DIR)/golangci-lint

help: ## Show this help.
	@echo ""
	@echo "devskills — common tasks (see README.md § Development)"
	@awk ' \
	  BEGIN { FS = ":[^#]*## " } \
	  /^# === / { section = substr($$0, 7); printed = 0; next } \
	  /^[a-zA-Z][a-zA-Z0-9_-]*:.*## / { \
	    if (section != "" && !printed) { printf "\n  \033[1m%s\033[0m\n", section; printed = 1 } \
	    printf "    \033[36m%-18s\033[0m %s\n", $$1, $$2; \
	  } \
	' $(MAKEFILE_LIST)
	@echo ""

# === Build

build: ## Build the ./devskills binary.
	go build -ldflags "$(LDFLAGS)" -o ./devskills .

install: ## Install the current tree to GOPATH/bin/devskills.
	go install -ldflags "$(LDFLAGS)" .

# === Test

test: ## Unit tests with the race detector.
	go test -race ./...

test-integration: ## Acceptance tests (tag-gated; a plain `go test ./...` skips them).
	go test -tags integration ./internal/acceptance/

# Always benches the working tree, never a stale installed binary.
bench: build ## Benchmark a skill change. Usage: make bench SKILL=ds-deslop [ARGS="--format pr-md"]
	./devskills bench $(SKILL) $(ARGS)

# === Gate

# Same steps as the release gate (.github/workflows/release.yml), which runs
# only on pushes to main — run this before trusting a green PR. fmt and tidy
# fix in place here; the gate diff-checks them instead.
verify: fmt tidy lint test test-integration ## Run the release gate locally: fmt, tidy, lint, unit + acceptance tests.

fmt: ## Format with gofmt (what the gate checks — no goimports).
	gofmt -w .

tidy: ## Tidy go.mod / go.sum.
	go mod tidy

lint: tools ## Lint with the pinned golangci-lint.
	$(GOLANGCI_LINT) run

tools: ## Install golangci-lint at the pinned version; reinstall on drift.
	@want="$(GOLANGCI_LINT_VERSION)"; want_bare="$${want#v}"; \
	have_bare=""; \
	if [ -x "$(GOLANGCI_LINT)" ]; then \
		have_bare="$$('$(GOLANGCI_LINT)' --version 2>/dev/null | awk '/has version/ {print $$4; exit}')"; \
	fi; \
	if [ "$$have_bare" != "$$want_bare" ]; then \
		echo "installing golangci-lint $$want (have: $${have_bare:-none})"; \
		GOBIN='$(GOBIN_DIR)' go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$$want; \
	fi

# === Release

snapshot: ## goreleaser cross-build dry-run (no publish).
	goreleaser build --snapshot --clean

clean: ## Remove ./devskills and ./dist.
	rm -f ./devskills
	rm -rf ./dist
