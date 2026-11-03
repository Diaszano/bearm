#!/bin/sh
set -eu

test -z "$(gofmt -l .)"
go mod verify
go vet ./...
go test ./...
go test -race ./...
go test ./test/compatibility -v
go test ./test/integration -v

go tool -modfile=tools/go.mod golangci-lint run ./...
go tool -modfile=tools/go.mod govulncheck ./...
go tool -modfile=tools/go.mod goreleaser check
goreleaser_skips="sign"
if ! command -v syft >/dev/null 2>&1; then
  goreleaser_skips="${goreleaser_skips},sbom"
fi
go tool -modfile=tools/go.mod goreleaser release --snapshot --clean --skip="${goreleaser_skips}"
