#!/bin/sh
set -eu

test -z "$(gofmt -l .)"
go mod verify
go vet ./...
go test ./...
go test -race ./...
go test ./test/compatibility -v
go test ./test/integration -v

if command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint run ./...
fi

if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./...
fi

if command -v goreleaser >/dev/null 2>&1; then
  goreleaser check
  goreleaser_skips="sign"
  if ! command -v syft >/dev/null 2>&1; then
    goreleaser_skips="${goreleaser_skips},sbom"
  fi
  goreleaser release --snapshot --clean --skip="${goreleaser_skips}"
fi
