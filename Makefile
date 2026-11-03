SHELL := /bin/sh

BINARY := bin/bearm
PACKAGES := ./...

.PHONY: all build test test-race vet fmt fmt-check lint lint-fix verify clean snapshot hooks check-commits

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
	go tool -modfile=tools/go.mod golangci-lint fmt $(PACKAGES)

fmt-check:
	go tool -modfile=tools/go.mod golangci-lint fmt --diff $(PACKAGES)

lint:
	go tool -modfile=tools/go.mod golangci-lint run $(PACKAGES)

lint-fix:
	go tool -modfile=tools/go.mod golangci-lint run --fix $(PACKAGES)

verify: fmt-check vet lint test test-race

clean:
	rm -rf bin dist coverage.out coverage.html

snapshot:
	@skips=; \
		if ! command -v cosign >/dev/null 2>&1; then skips=sign; fi; \
		if ! command -v syft >/dev/null 2>&1; then \
			if [ -n "$$skips" ]; then skips="$$skips,sbom"; else skips=sbom; fi; \
		fi; \
		if [ -n "$$skips" ]; then skips="--skip=$$skips"; fi; \
		go tool -modfile=tools/go.mod goreleaser release --snapshot --clean $$skips

hooks:
	@if [ -x .githooks/commit-msg ]; then \
		git config core.hooksPath .githooks; \
		echo "Configured git hooks via .githooks"; \
	else \
		mkdir -p ~/.config/bearm/hooks; \
		printf '#!/bin/sh\nTOPLEVEL=$$(git rev-parse --show-toplevel 2>/dev/null || pwd)\nexec sh "$$TOPLEVEL/.githooks/commit-msg" "$$@"\n' > ~/.config/bearm/hooks/commit-msg; \
		chmod +x ~/.config/bearm/hooks/commit-msg; \
		git config core.hooksPath ~/.config/bearm/hooks; \
		echo "Configured git hooks via ~/.config/bearm/hooks (compatibility trampoline for filesystem without direct execute permissions)"; \
	fi

check-commits:
	@sh ./scripts/verify-commits.sh origin/main HEAD
