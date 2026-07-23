SHELL := /bin/sh

BINARY := bin/bearm
PACKAGES := ./...

.PHONY: all build test test-race vet fmt-check lint verify clean

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

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt is required for:"; gofmt -l .; exit 1; }

lint:
	golangci-lint run ./...

verify: fmt-check vet test test-race

clean:
	rm -rf bin dist coverage.out coverage.html
