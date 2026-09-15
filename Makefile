SHELL := /bin/sh

BINARY := bin/bearm
PACKAGES := ./...

.PHONY: all build test test-race vet fmt fmt-check lint lint-fix verify clean snapshot

all: verify build

build:
	mkdir -p bin
	go build -trimpath -o $(BINARY) ./cmd/bearm

test:
	go test $(PACKAGES)

test-race:
	go test -race $(PACKAGES)

vet:
	go vet $(PACKAGES)

fmt:
	golangci-lint fmt $(PACKAGES)

fmt-check:
	golangci-lint fmt --diff $(PACKAGES)

lint:
	golangci-lint run $(PACKAGES)

lint-fix:
	golangci-lint run --fix $(PACKAGES)

verify: fmt-check vet lint test test-race

clean:
	rm -rf bin dist coverage.out coverage.html

snapshot:
	goreleaser release --snapshot --clean
