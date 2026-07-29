.PHONY: build install test test-integration lint fmt clean snapshot release

# Version from the VERSION file (used by local/snapshot builds).
VERSION ?= $(shell cat VERSION)

# LDFLAGS mirror the goreleaser release build: stripped binary with the
# version injected into internal/cli.version. goreleaser itself adds
# -s -w; we keep them here so local builds behave the same way.
LDFLAGS := -s -w -X github.com/gleicon/devskills/internal/cli.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o ./devskills .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test -race ./...

test-integration:
	go test -tags integration ./internal/acceptance/

lint:
	golangci-lint run

fmt:
	gofmt -w .

snapshot:
	goreleaser build --snapshot --clean

release:
	goreleaser release --clean

clean:
	rm -f ./devskills
	rm -rf ./dist
