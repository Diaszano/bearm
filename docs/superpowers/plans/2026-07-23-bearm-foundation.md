# Bearm Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a buildable, testable Bearm Go repository with stable domain contracts, platform directory resolution, version reporting, and dependency assembly.

**Architecture:** The executable remains a thin composition root. Domain types contain no infrastructure dependencies. Platform directory resolution is isolated behind build-tag-free standard-library code so later configuration, trash, and journal packages can depend on stable paths.

**Tech Stack:** Go 1.24+, Go standard library, `github.com/pelletier/go-toml/v2`, Make, GitHub Actions, GoReleaser.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Native Bearm user-facing text defaults to `pt-BR`; compatibility text is profile and locale aware.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- No shell execution from application code.
- No telemetry or network access.
- Use TDD for behavior.
- Run formatting, tests, race tests, and vet before completing the plan.
- Use one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Preserve a single-responsibility package boundary; do not create a generic `utils` package.
- Go version floor is `1.24.0`.

---

## File Structure

```text
bearm/
├── cmd/bearm/main.go
├── internal/app/app.go
├── internal/app/app_test.go
├── internal/buildinfo/buildinfo.go
├── internal/buildinfo/buildinfo_test.go
├── internal/domain/removal.go
├── internal/domain/removal_test.go
├── internal/domain/trash.go
├── internal/platform/dirs.go
├── internal/platform/dirs_test.go
├── .github/workflows/ci.yml
├── .gitignore
├── .golangci.yml
├── .goreleaser.yaml
├── LICENSE
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

## Dependency Order

```text
Task 1 repository/toolchain
  ├── Task 2 domain contracts
  ├── Task 3 platform directories
  └── Task 4 build information
          └── Task 5 application entry point and CI
```

### Task 1: Initialize the Repository and Toolchain

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `Makefile`
- Create: `.golangci.yml`
- Create: `LICENSE`
- Create: `README.md`

**Interfaces:**
- Consumes: none.
- Produces: Go module `github.com/Diaszano/bearm`, standard verification commands, and repository conventions used by every later task.

- [ ] **Step 1: Create the Go module**

Create `go.mod`:

```go
module github.com/Diaszano/bearm

go 1.24.0

require github.com/pelletier/go-toml/v2 v2.2.3
```

- [ ] **Step 2: Create repository ignore rules**

Create `.gitignore`:

```gitignore
/bin/
/dist/
/coverage.out
coverage.html
*.prof
.DS_Store
.idea/
.vscode/
```

- [ ] **Step 3: Create deterministic development commands**

Create `Makefile`:

```makefile
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
```

- [ ] **Step 4: Add lint configuration**

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - asciicheck
    - bodyclose
    - errcheck
    - errorlint
    - exhaustive
    - gocritic
    - govet
    - ineffassign
    - misspell
    - nilerr
    - nolintlint
    - revive
    - staticcheck
    - unconvert
    - unused

issues:
  exclude-use-default: false
```

- [ ] **Step 5: Add the MIT license**

Create `LICENSE` with the standard MIT license text and copyright:

```text
MIT License

Copyright (c) 2026 Lucas Santos

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 6: Add the initial README**

Create `README.md`:

````markdown
# Bearm

Bearm is a safe, native replacement for Unix `rm`. Compatibility-mode removal
moves targets to operating-system trash locations instead of permanently
deleting them.

## Status

The project is under active development. Do not alias `rm` to Bearm until the
compatibility and platform test suites pass for your operating system.

## Development

```bash
make verify
make build
./bin/bearm version
```

## Project Rules

- Code and technical documentation are written in English.
- Commits follow Conventional Commits.
- Branches use conventional English names such as `feat/linux-trash-backend`.
- Bearm has no telemetry and performs no network requests.
````

- [ ] **Step 7: Resolve dependencies**

Run:

```bash
go mod tidy
```

Expected:

```text
go.sum is created and go.mod remains valid.
```

- [ ] **Step 8: Verify repository metadata**

Run:

```bash
go mod verify
```

Expected:

```text
all modules verified
```

- [ ] **Step 9: Commit**

```bash
git add go.mod go.sum .gitignore Makefile .golangci.yml LICENSE README.md
git commit -m "chore: initialize Bearm Go module"
```

### Task 2: Define Stable Domain Contracts

**Files:**
- Create: `internal/domain/removal.go`
- Create: `internal/domain/removal_test.go`
- Create: `internal/domain/trash.go`

**Interfaces:**
- Consumes: Go standard library.
- Produces:
  - `domain.CompatibilityProfile`
  - `domain.InteractiveMode`
  - `domain.PreserveRootMode`
  - `domain.RemoveOptions`
  - `domain.RemoveRequest`
  - `domain.PlannedTarget`
  - `domain.RemovalPlan`
  - `domain.TrashRecord`
  - `domain.TrashBackend`

- [ ] **Step 1: Write failing validation tests**

Create `internal/domain/removal_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestRemoveRequestValidateAcceptsGNURequest(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{"file.txt"},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRemoveRequestValidateRejectsUnknownProfile(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile:  domain.CompatibilityProfile("unknown"),
		Operands: []string{"file.txt"},
	}

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestRemoveRequestValidateRejectsMissingOperandsWithoutForce(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{Profile: domain.ProfileGNU}

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestRemoveRequestValidateAcceptsMissingOperandsWithForce(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile: domain.ProfileGNU,
		Options: domain.RemoveOptions{Force: true},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
```

- [ ] **Step 2: Run the tests and verify failure**

Run:

```bash
go test ./internal/domain -run TestRemoveRequest -v
```

Expected:

```text
FAIL because package internal/domain or RemoveRequest does not exist.
```

- [ ] **Step 3: Implement removal domain types**

Create `internal/domain/removal.go`:

```go
// Package domain contains Bearm's infrastructure-independent domain types.
package domain

import (
	"errors"
	"time"
)

// CompatibilityProfile selects the rm behavior Bearm emulates.
type CompatibilityProfile string

const (
	// ProfileGNU selects GNU coreutils-style parsing and diagnostics.
	ProfileGNU CompatibilityProfile = "gnu"
	// ProfileBSD selects BSD/macOS-style parsing and diagnostics.
	ProfileBSD CompatibilityProfile = "bsd"
	// ProfilePOSIX selects the portable POSIX subset.
	ProfilePOSIX CompatibilityProfile = "posix"
)

// InteractiveMode controls removal confirmation behavior.
type InteractiveMode string

const (
	// InteractiveDefault lets the compatibility profile choose its default.
	InteractiveDefault InteractiveMode = "default"
	// InteractiveNever disables confirmation.
	InteractiveNever InteractiveMode = "never"
	// InteractiveOnce prompts once for risky multi-target operations.
	InteractiveOnce InteractiveMode = "once"
	// InteractiveAlways prompts for every logical removal.
	InteractiveAlways InteractiveMode = "always"
)

// PreserveRootMode controls root and mount-root protection.
type PreserveRootMode string

const (
	// PreserveRootDefault protects the system root.
	PreserveRootDefault PreserveRootMode = "default"
	// PreserveRootNone disables compatibility-level root preservation.
	PreserveRootNone PreserveRootMode = "none"
	// PreserveRootAll protects command-line directories on distinct devices.
	PreserveRootAll PreserveRootMode = "all"
)

// RemoveOptions contains parsed compatibility-mode removal options.
type RemoveOptions struct {
	Force         bool
	Recursive     bool
	Directory     bool
	Verbose       bool
	Interactive   InteractiveMode
	PreserveRoot  PreserveRootMode
	OneFileSystem bool
}

// RemoveRequest is a compatibility-mode removal request.
type RemoveRequest struct {
	Profile  CompatibilityProfile
	Operands []string
	Options  RemoveOptions
}

// Validate validates profile-independent request invariants.
func (r RemoveRequest) Validate() error {
	switch r.Profile {
	case ProfileGNU, ProfileBSD, ProfilePOSIX:
	default:
		return errors.New("unsupported compatibility profile")
	}

	if len(r.Operands) == 0 && !r.Options.Force {
		return errors.New("missing operand")
	}

	return nil
}

// TargetKind classifies a planned filesystem target.
type TargetKind string

const (
	// TargetFile identifies a regular file.
	TargetFile TargetKind = "file"
	// TargetDir identifies a directory that is not a symlink.
	TargetDir TargetKind = "directory"
	// TargetSymlink identifies a symbolic link.
	TargetSymlink TargetKind = "symlink"
	// TargetOther identifies another filesystem object.
	TargetOther TargetKind = "other"
)

// PlannedTarget contains immutable target data captured during planning.
type PlannedTarget struct {
	InputPath    string
	AbsolutePath string
	Kind         TargetKind
	DeviceID     uint64
	RequiresWalk bool
}

// RemovalPlan is the validated set of targets for one operation.
type RemovalPlan struct {
	ID        string
	CreatedAt time.Time
	Request   RemoveRequest
	Targets   []PlannedTarget
}
```

- [ ] **Step 4: Define trash records and backend interface**

Create `internal/domain/trash.go`:

```go
package domain

import (
	"context"
	"time"
)

// Destination is a reserved trash destination.
type Destination struct {
	Root       string
	FilesDir   string
	InfoDir    string
	TargetPath string
	InfoPath   string
}

// TrashRecord describes one completed trash move.
type TrashRecord struct {
	SchemaVersion int       `json:"schema_version"`
	ItemID        string    `json:"item_id"`
	OperationID   string    `json:"operation_id"`
	OriginalPath  string    `json:"original_path"`
	TrashedPath   string    `json:"trashed_path"`
	Backend       string    `json:"backend"`
	DeviceID      uint64    `json:"device_id"`
	DeletedAt     time.Time `json:"deleted_at"`
	Status        string    `json:"status"`
}

// TrashBackend moves planned targets into a platform trash.
type TrashBackend interface {
	Name() string
	Resolve(context.Context, PlannedTarget) (Destination, error)
	Move(context.Context, PlannedTarget, Destination, string) (TrashRecord, error)
}
```

- [ ] **Step 5: Run domain tests**

Run:

```bash
go test ./internal/domain -v
```

Expected:

```text
PASS
```

- [ ] **Step 6: Format and vet**

Run:

```bash
gofmt -w internal/domain
go vet ./internal/domain
```

Expected:

```text
No output and exit code 0.
```

- [ ] **Step 7: Commit**

```bash
git add internal/domain
git commit -m "feat: define Bearm domain contracts"
```

### Task 3: Resolve Platform Directories

**Files:**
- Create: `internal/platform/dirs.go`
- Create: `internal/platform/dirs_test.go`

**Interfaces:**
- Consumes: `runtime.GOOS`, process environment, and `os.UserHomeDir`.
- Produces:
  - `platform.Dirs`
  - `platform.ResolveDirs(getenv, home, goos)`

- [ ] **Step 1: Write failing directory-resolution tests**

Create `internal/platform/dirs_test.go`:

```go
package platform_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/platform"
)

func TestResolveDirsUsesXDGOnLinux(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"XDG_CONFIG_HOME": "/tmp/config",
		"XDG_STATE_HOME":  "/tmp/state",
		"XDG_DATA_HOME":   "/tmp/data",
	}
	getenv := func(key string) string { return env[key] }

	got, err := platform.ResolveDirs(getenv, "/home/dias", "linux")
	if err != nil {
		t.Fatalf("ResolveDirs() error = %v", err)
	}

	if got.ConfigRoot != "/tmp/config/bearm" {
		t.Fatalf("ConfigRoot = %q", got.ConfigRoot)
	}
	if got.StateRoot != "/tmp/state/bearm" {
		t.Fatalf("StateRoot = %q", got.StateRoot)
	}
	if got.DataRoot != "/tmp/data/bearm" {
		t.Fatalf("DataRoot = %q", got.DataRoot)
	}
}

func TestResolveDirsUsesMacApplicationSupport(t *testing.T) {
	t.Parallel()

	got, err := platform.ResolveDirs(func(string) string { return "" }, "/Users/dias", "darwin")
	if err != nil {
		t.Fatalf("ResolveDirs() error = %v", err)
	}

	want := "/Users/dias/Library/Application Support/Bearm"
	if got.ConfigRoot != want {
		t.Fatalf("ConfigRoot = %q, want %q", got.ConfigRoot, want)
	}
	if got.StateRoot != want+"/state" {
		t.Fatalf("StateRoot = %q", got.StateRoot)
	}
}

func TestResolveDirsRejectsRelativeOverride(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_STATE_HOME": "relative/path"}
	_, err := platform.ResolveDirs(func(key string) string { return env[key] }, "/home/dias", "linux")
	if err == nil {
		t.Fatal("ResolveDirs() error = nil, want non-nil")
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/platform -run TestResolveDirs -v
```

Expected:

```text
FAIL because ResolveDirs does not exist.
```

- [ ] **Step 3: Implement directory resolution**

Create `internal/platform/dirs.go`:

```go
// Package platform contains operating-system-specific path and filesystem helpers.
package platform

import (
	"errors"
	"path/filepath"
)

// Dirs contains Bearm configuration, state, and data roots.
type Dirs struct {
	ConfigRoot string
	StateRoot  string
	DataRoot   string
}

// ResolveDirs resolves Bearm directories without reading global process state directly.
func ResolveDirs(getenv func(string) string, home, goos string) (Dirs, error) {
	if home == "" || !filepath.IsAbs(home) {
		return Dirs{}, errors.New("home directory must be absolute")
	}

	base := Dirs{}
	switch goos {
	case "darwin":
		support := filepath.Join(home, "Library", "Application Support", "Bearm")
		base = Dirs{
			ConfigRoot: support,
			StateRoot:  filepath.Join(support, "state"),
			DataRoot:   filepath.Join(support, "data"),
		}
	default:
		configHome := absoluteOrFallback(getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config"))
		stateHome := absoluteOrFallback(getenv("XDG_STATE_HOME"), filepath.Join(home, ".local", "state"))
		dataHome := absoluteOrFallback(getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share"))
		base = Dirs{
			ConfigRoot: filepath.Join(configHome, "bearm"),
			StateRoot:  filepath.Join(stateHome, "bearm"),
			DataRoot:   filepath.Join(dataHome, "bearm"),
		}
	}

	var err error
	if base.ConfigRoot, err = applyAbsoluteOverride(getenv("BEARM_CONFIG_HOME"), base.ConfigRoot); err != nil {
		return Dirs{}, err
	}
	if base.StateRoot, err = applyAbsoluteOverride(getenv("BEARM_STATE_HOME"), base.StateRoot); err != nil {
		return Dirs{}, err
	}
	if base.DataRoot, err = applyAbsoluteOverride(getenv("BEARM_DATA_HOME"), base.DataRoot); err != nil {
		return Dirs{}, err
	}

	return base, nil
}

func absoluteOrFallback(value, fallback string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return fallback
}

func applyAbsoluteOverride(value, fallback string) (string, error) {
	if value == "" {
		return fallback, nil
	}
	if !filepath.IsAbs(value) {
		return "", errors.New("Bearm directory overrides must be absolute")
	}
	return filepath.Clean(value), nil
}
```

- [ ] **Step 4: Run and format tests**

Run:

```bash
gofmt -w internal/platform
go test ./internal/platform -v
```

Expected:

```text
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/platform
git commit -m "feat: resolve Bearm platform directories"
```

### Task 4: Add Build Information

**Files:**
- Create: `internal/buildinfo/buildinfo.go`
- Create: `internal/buildinfo/buildinfo_test.go`

**Interfaces:**
- Consumes: linker variables `Version`, `Commit`, and `Date`.
- Produces:
  - `buildinfo.Info`
  - `buildinfo.Current()`
  - `buildinfo.Info.String()`

- [ ] **Step 1: Write failing build-info tests**

Create `internal/buildinfo/buildinfo_test.go`:

```go
package buildinfo_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/buildinfo"
)

func TestInfoString(t *testing.T) {
	t.Parallel()

	info := buildinfo.Info{
		Version: "1.2.3",
		Commit:  "abcdef0",
		Date:    "2026-07-23T12:00:00Z",
	}

	want := "Bearm 1.2.3 (abcdef0, 2026-07-23T12:00:00Z)"
	if got := info.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestCurrentUsesDevelopmentDefaults(t *testing.T) {
	t.Parallel()

	info := buildinfo.Current()
	if info.Version == "" || info.Commit == "" || info.Date == "" {
		t.Fatalf("Current() = %#v, expected non-empty fields", info)
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/buildinfo -v
```

Expected:

```text
FAIL because package internal/buildinfo does not exist.
```

- [ ] **Step 3: Implement build information**

Create `internal/buildinfo/buildinfo.go`:

```go
// Package buildinfo exposes immutable release metadata.
package buildinfo

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Info contains Bearm release metadata.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Current returns the metadata injected at build time.
func Current() Info {
	return Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

// String returns a stable human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("Bearm %s (%s, %s)", i.Version, i.Commit, i.Date)
}
```

- [ ] **Step 4: Run tests**

Run:

```bash
gofmt -w internal/buildinfo
go test ./internal/buildinfo -v
```

Expected:

```text
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/buildinfo
git commit -m "feat: expose Bearm build information"
```

### Task 5: Add the Thin Application Entry Point and Baseline CI

**Files:**
- Create: `internal/app/app.go`
- Create: `internal/app/app_test.go`
- Create: `cmd/bearm/main.go`
- Create: `.github/workflows/ci.yml`
- Create: `.goreleaser.yaml`
- Modify: `Makefile`

**Interfaces:**
- Consumes:
  - `buildinfo.Current() buildinfo.Info`
  - `platform.ResolveDirs(func(string) string, string, string) (platform.Dirs, error)`
- Produces:
  - `app.App`
  - `app.New(io.Reader, io.Writer, io.Writer, buildinfo.Info) *App`
  - `(*app.App).Run(context.Context, []string) int`
  - working `bearm version`

- [ ] **Step 1: Write failing application tests**

Create `internal/app/app_test.go`:

```go
package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Info{
		Version: "1.0.0",
		Commit:  "abcdef0",
		Date:    "2026-07-23T12:00:00Z",
	})

	code := instance.Run(context.Background(), []string{"bearm", "version"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "Bearm 1.0.0 (abcdef0, 2026-07-23T12:00:00Z)\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"bearm", "unknown"})
	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "comando desconhecido") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/app -v
```

Expected:

```text
FAIL because app.New and App.Run do not exist.
```

- [ ] **Step 3: Implement the application shell**

Create `internal/app/app.go`:

```go
// Package app coordinates Bearm use cases.
package app

import (
	"context"
	"fmt"
	"io"

	"github.com/Diaszano/bearm/internal/buildinfo"
)

// App is the Bearm application shell.
type App struct {
	stdin io.Reader
	out   io.Writer
	err   io.Writer
	info  buildinfo.Info
}

// New creates an application with explicit input and output streams.
func New(stdin io.Reader, stdout, stderr io.Writer, info buildinfo.Info) *App {
	return &App{
		stdin: stdin,
		out:   stdout,
		err:   stderr,
		info:  info,
	}
}

// Run executes one Bearm invocation and returns a process exit code.
func (a *App) Run(_ context.Context, argv []string) int {
	if len(argv) == 2 && argv[1] == "version" {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}

	fmt.Fprintln(a.err, "bearm: comando desconhecido")
	return 2
}
```

- [ ] **Step 4: Add the executable composition root**

Create `cmd/bearm/main.go`:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	instance := app.New(os.Stdin, os.Stdout, os.Stderr, buildinfo.Current())
	os.Exit(instance.Run(ctx, os.Args))
}
```

- [ ] **Step 5: Run the application tests and binary**

Run:

```bash
gofmt -w cmd internal/app
go test ./internal/app -v
go run ./cmd/bearm version
```

Expected final output:

```text
Bearm dev (none, unknown)
```

- [ ] **Step 6: Add CI**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: ["**"]
  pull_request:

permissions:
  contents: read

jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest]
        go: ["1.24.x", "stable"]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ matrix.go }}
          cache: true
      - run: go mod verify
      - run: test -z "$(gofmt -l .)"
      - run: go vet ./...
      - run: go test ./...
      - run: go test -race ./...

  snapshot:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: latest
          args: release --snapshot --clean
```

- [ ] **Step 7: Add GoReleaser configuration**

Create `.goreleaser.yaml`:

```yaml
version: 2

project_name: bearm

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: bearm
    main: ./cmd/bearm
    binary: bearm
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    ldflags:
      - >-
        -s -w
        -X github.com/Diaszano/bearm/internal/buildinfo.version={{.Version}}
        -X github.com/Diaszano/bearm/internal/buildinfo.commit={{.Commit}}
        -X github.com/Diaszano/bearm/internal/buildinfo.date={{.Date}}

archives:
  - formats: [tar.gz]
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}

checksum:
  name_template: checksums.txt

changelog:
  use: github-native
```

- [ ] **Step 8: Add a snapshot target**

Append to `Makefile`:

```makefile
.PHONY: snapshot

snapshot:
	goreleaser release --snapshot --clean
```

- [ ] **Step 9: Run complete verification**

Run:

```bash
make verify
make build
./bin/bearm version
```

Expected:

```text
All verification commands exit 0.
The final command prints Bearm development build information.
```

- [ ] **Step 10: Commit**

```bash
git add cmd internal/app .github/workflows/ci.yml .goreleaser.yaml Makefile
git commit -m "feat: add Bearm application entry point"
```

## Plan Completion Verification

Run:

```bash
make verify
make build
./bin/bearm version
git status --short
```

Expected:

```text
All commands succeed.
The binary prints Bearm version information.
The working tree is clean.
```
