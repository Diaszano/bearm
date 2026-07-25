# Bearm Removal Engine Implementation Plan

**Status:** Complete
**Completion date:** 2026-07-25
**Implementation range:** `ac35fabc68158fb381d56a274b9238e110190531..01931408c25b55e95ae4dd12061ce11388b348f8`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Bearm's hard safety policy, filesystem planner, interactive prompt engine, sequential executor, and compatibility-mode application flow.

**Architecture:** Planning is read-only and produces immutable targets. Safety checks run during planning and immediately before execution. The executor depends only on a trash backend, journal interface, terminal interface, clock, and ID generator, allowing complete unit testing without touching the real user trash.

**Tech Stack:** Go 1.24+, Go standard library, existing platform trash backends, table-driven tests.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Compatibility mode must be silent on success unless verbose output is requested.
- Diagnostics are written to stderr.
- Use `Lstat`; never follow the final symlink operand.
- `/`, `.`, `..`, trash roots, state roots, and config roots remain hard protected.
- A directory uses the no-traversal fast path unless interaction, verbose traversal, mount filtering, or protected-descendant inspection requires walking.
- Revalidate safety immediately before moving each target.
- A target failure sets a nonzero result but does not skip independent later operands.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
internal/
├── app/
│   ├── app.go
│   ├── app_test.go
│   ├── backend_darwin.go
│   └── backend_linux.go
├── domain/
│   ├── result.go
│   └── result_test.go
├── id/
│   ├── generator.go
│   └── generator_test.go
├── planner/
│   ├── planner.go
│   ├── planner_test.go
│   ├── preserve_root.go
│   ├── preserve_root_test.go
│   ├── walk.go
│   └── walk_test.go
├── removal/
│   ├── executor.go
│   ├── executor_test.go
│   ├── prompt.go
│   └── prompt_test.go
├── safety/
│   ├── policy.go
│   └── policy_test.go
└── testutil/
    ├── backend.go
    ├── journal.go
    └── terminal.go
```

## Dependency Order

```text
Task 1 IDs and result types
  ├── Task 2 hard safety policy
  ├── Task 3 filesystem planner
  ├── Task 4 prompt engine
  └── Task 5 removal executor
          └── Task 6 compatibility application integration
```

### Task 1: Add Stable IDs and Execution Result Types

**Files:**
- Create: `internal/id/generator.go`
- Create: `internal/id/generator_test.go`
- Create: `internal/domain/result.go`
- Create: `internal/domain/result_test.go`

**Interfaces:**
- Consumes: cryptographic random source.
- Produces:
  - `id.New() (string, error)`
  - `domain.ItemResult`
  - `domain.RemovalResult`
  - `domain.RemovalResult.ExitCode(profile) int`

- [x] **Step 1: Write failing ID tests**

Create `internal/id/generator_test.go`:

```go
package id_test

import (
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/id"
)

func TestNewReturnsFixedLengthLowercaseHex(t *testing.T) {
	t.Parallel()

	got, err := id.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("len(New()) = %d, want 32", len(got))
	}
	if got != strings.ToLower(got) {
		t.Fatalf("New() = %q, want lowercase", got)
	}
}

func TestNewReturnsUniqueValues(t *testing.T) {
	t.Parallel()

	first, err := id.New()
	if err != nil {
		t.Fatal(err)
	}
	second, err := id.New()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("duplicate IDs %q", first)
	}
}
```

- [x] **Step 2: Implement cryptographic IDs**

Create `internal/id/generator.go`:

```go
// Package id creates opaque Bearm operation and item identifiers.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a 128-bit lowercase hexadecimal identifier.
func New() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
```

- [x] **Step 3: Write failing result tests**

Create `internal/domain/result_test.go`:

```go
package domain_test

import (
	"errors"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestRemovalResultExitCodeSuccess(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{{Path: "file", Status: domain.ItemTrashed}},
	}
	if got := result.ExitCode(domain.ProfileGNU); got != 0 {
		t.Fatalf("ExitCode() = %d", got)
	}
}

func TestRemovalResultExitCodeFailure(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{{Path: "file", Status: domain.ItemFailed, Err: errors.New("failed")}},
	}
	if got := result.ExitCode(domain.ProfileGNU); got != 1 {
		t.Fatalf("GNU ExitCode() = %d", got)
	}
	if got := result.ExitCode(domain.ProfileBSD); got != 1 {
		t.Fatalf("BSD operational ExitCode() = %d", got)
	}
}

func TestRemovalResultHasFailures(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{
			{Path: "a", Status: domain.ItemTrashed},
			{Path: "b", Status: domain.ItemFailed, Err: errors.New("failed")},
		},
	}
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
}
```

- [x] **Step 4: Implement result types**

Create `internal/domain/result.go`:

```go
package domain

// ItemStatus identifies one compatibility removal outcome.
type ItemStatus string

const (
	// ItemTrashed indicates a successful move to trash.
	ItemTrashed ItemStatus = "trashed"
	// ItemSkipped indicates an intentionally ignored target.
	ItemSkipped ItemStatus = "skipped"
	// ItemDeclined indicates a user-declined interactive target.
	ItemDeclined ItemStatus = "declined"
	// ItemFailed indicates an operational failure.
	ItemFailed ItemStatus = "failed"
)

// ItemResult contains one target outcome.
type ItemResult struct {
	Path   string
	Status ItemStatus
	Record *TrashRecord
	Err    error
}

// RemovalResult contains the outcomes of one compatibility operation.
type RemovalResult struct {
	OperationID string
	Items       []ItemResult
}

// HasFailures reports whether any target failed.
func (r RemovalResult) HasFailures() bool {
	for _, item := range r.Items {
		if item.Status == ItemFailed {
			return true
		}
	}
	return false
}

// ExitCode returns the compatibility operational exit code.
func (r RemovalResult) ExitCode(_ CompatibilityProfile) int {
	if r.HasFailures() {
		return 1
	}
	return 0
}
```

- [x] **Step 5: Run tests**

Run:

```bash
gofmt -w internal/id internal/domain
go test ./internal/id ./internal/domain -v
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add internal/id internal/domain/result.go internal/domain/result_test.go
git commit -m "feat: add Bearm operation results"
```

### Task 2: Implement the Hard Safety Policy

**Files:**
- Create: `internal/safety/policy.go`
- Create: `internal/safety/policy_test.go`

**Interfaces:**
- Consumes: absolute lexical target paths and configured protected roots.
- Produces:
  - `safety.Policy`
  - `safety.NewPolicy(config) (*Policy, error)`
  - `(*Policy).Check(path string) error`

- [x] **Step 1: Write failing safety tests**

Create `internal/safety/policy_test.go`:

```go
package safety_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func TestPolicyRejectsRootAndDotSegments(t *testing.T) {
	t.Parallel()

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/", ".", "..", "/tmp/.", "/tmp/.."} {
		if err := policy.Check(path); err == nil {
			t.Errorf("Check(%q) error = nil, want non-nil", path)
		}
	}
}

func TestPolicyRejectsInternalRoots(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{
			filepath.Join(root, "trash"),
			filepath.Join(root, "state"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(root, "trash"),
		filepath.Join(root, "trash", "item"),
		filepath.Join(root, "state", "journal.jsonl"),
	} {
		if err := policy.Check(path); err == nil {
			t.Errorf("Check(%q) error = nil, want non-nil", path)
		}
	}
}

func TestPolicyEnforcesAllowedRoots(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	policy, err := safety.NewPolicy(safety.Config{AllowedRoots: []string{allowed}})
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Check(filepath.Join(allowed, "file")); err != nil {
		t.Fatalf("allowed Check() error = %v", err)
	}
	if err := policy.Check(filepath.Join(root, "outside")); err == nil {
		t.Fatal("outside Check() error = nil, want non-nil")
	}
}

func TestPolicyDoesNotResolveFinalSymlinkLexically(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policy, err := safety.NewPolicy(safety.Config{AllowedRoots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "link")
	if err := policy.Check(path); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}
```

- [x] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/safety -v
```

Expected:

```text
FAIL because package internal/safety does not exist.
```

- [x] **Step 3: Implement the safety policy**

Create `internal/safety/policy.go`:

```go
// Package safety enforces Bearm hard and configured path protections.
package safety

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Config contains immutable safety policy inputs.
type Config struct {
	HardProtectedRoots []string
	AllowedRoots       []string
}

// Policy validates lexical absolute target paths.
type Policy struct {
	hardProtected []string
	allowed       []string
}

// NewPolicy validates and creates a safety policy.
func NewPolicy(config Config) (*Policy, error) {
	policy := &Policy{}
	var err error

	policy.hardProtected, err = normalizeRoots(config.HardProtectedRoots)
	if err != nil {
		return nil, fmt.Errorf("hard protected roots: %w", err)
	}
	policy.allowed, err = normalizeRoots(config.AllowedRoots)
	if err != nil {
		return nil, fmt.Errorf("allowed roots: %w", err)
	}

	return policy, nil
}

// Check verifies that path is not hard protected and is inside configured scope.
func (p *Policy) Check(path string) error {
	if path == "." || path == ".." {
		return errors.New("dot and dot-dot may not be removed")
	}

	cleaned := filepath.Clean(path)
	if cleaned == string(filepath.Separator) {
		return errors.New("system root may not be removed")
	}
	if filepath.Base(path) == "." || filepath.Base(path) == ".." {
		return errors.New("dot and dot-dot may not be removed")
	}
	if !filepath.IsAbs(cleaned) {
		return errors.New("safety checks require an absolute path")
	}

	for _, root := range p.hardProtected {
		if isEqualOrDescendant(cleaned, root) {
			return fmt.Errorf("path is protected by Bearm: %s", root)
		}
	}

	if len(p.allowed) == 0 {
		return nil
	}
	for _, root := range p.allowed {
		if isEqualOrDescendant(cleaned, root) {
			return nil
		}
	}

	return errors.New("path is outside allowed roots")
}

func normalizeRoots(values []string) ([]string, error) {
	roots := make([]string, 0, len(values))
	for _, value := range values {
		if !filepath.IsAbs(value) {
			return nil, fmt.Errorf("%q is not absolute", value)
		}
		roots = append(roots, filepath.Clean(value))
	}
	return roots, nil
}

func isEqualOrDescendant(path, root string) bool {
	if path == root {
		return true
	}
	prefix := root
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(path, prefix)
}
```

- [x] **Step 4: Run safety tests**

Run:

```bash
gofmt -w internal/safety
go test ./internal/safety -v
```

Expected:

```text
PASS
```

- [x] **Step 5: Commit**

```bash
git add internal/safety
git commit -m "feat: enforce Bearm hard path protections"
```

### Task 3: Implement Filesystem Planning and the No-Traversal Fast Path

**Files:**
- Create: `internal/planner/planner.go`
- Create: `internal/planner/planner_test.go`
- Create: `internal/planner/walk.go`
- Create: `internal/planner/walk_test.go`

**Interfaces:**
- Consumes:
  - `domain.RemoveRequest`
  - `safety.Policy`
  - filesystem inspection functions.
- Produces:
  - `planner.Planner`
  - `planner.New(policy, idGenerator)`
  - `(*Planner).Plan(ctx, request) (domain.RemovalPlan, []domain.ItemResult)`

- [x] **Step 1: Write failing planner tests**

Create `internal/planner/planner_test.go`:

```go
package planner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
	"github.com/Diaszano/bearm/internal/safety"
)

func TestPlanClassifiesDanglingSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "missing"), link); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{link},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if len(plan.Targets) != 1 || plan.Targets[0].Kind != domain.TargetSymlink {
		t.Fatalf("Targets = %#v", plan.Targets)
	}
}

func TestPlanUsesFastPathForRecursiveDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(filepath.Join(directory, "pkg"), 0o700); err != nil {
		t.Fatal(err)
	}

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
		Options: domain.RemoveOptions{
			Recursive:   true,
			Interactive: domain.InteractiveNever,
		},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if plan.Targets[0].RequiresWalk {
		t.Fatal("RequiresWalk = true, want false")
	}
}

func TestPlanRequiresWalkForInteractiveDirectory(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
		Options: domain.RemoveOptions{
			Recursive:   true,
			Interactive: domain.InteractiveAlways,
		},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if !plan.Targets[0].RequiresWalk {
		t.Fatal("RequiresWalk = false, want true")
	}
}

func TestPlanIgnoresMissingOperandUnderForce(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{filepath.Join(t.TempDir(), "missing")},
		Options:  domain.RemoveOptions{Force: true},
	})
	if len(plan.Targets) != 0 || len(failures) != 0 {
		t.Fatalf("plan = %#v, failures = %#v", plan, failures)
	}
}

func TestPlanRejectsDirectoryWithoutRecursiveOrDirectoryOption(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	_, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
	})
	if len(failures) != 1 || failures[0].Status != domain.ItemFailed {
		t.Fatalf("failures = %#v", failures)
	}
}
```

- [x] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/planner -run TestPlan -v
```

Expected:

```text
FAIL because package internal/planner does not exist.
```

- [x] **Step 3: Implement the planner**

Create `internal/planner/planner.go`:

```go
// Package planner inspects removal operands and builds immutable execution plans.
package planner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/safety"
)

// IDGenerator creates operation IDs.
type IDGenerator func() (string, error)

// Planner builds removal plans without mutating the filesystem.
type Planner struct {
	policy      *safety.Policy
	idGenerator IDGenerator
}

// New creates a removal planner.
func New(policy *safety.Policy, idGenerator IDGenerator) *Planner {
	return &Planner{policy: policy, idGenerator: idGenerator}
}

// Plan validates and classifies all request operands.
func (p *Planner) Plan(
	ctx context.Context,
	request domain.RemoveRequest,
) (domain.RemovalPlan, []domain.ItemResult) {
	operationID, err := p.idGenerator()
	if err != nil {
		return domain.RemovalPlan{}, []domain.ItemResult{{
			Status: domain.ItemFailed,
			Err:    err,
		}}
	}

	plan := domain.RemovalPlan{
		ID:        operationID,
		CreatedAt: time.Now().UTC(),
		Request:   request,
	}
	failures := make([]domain.ItemResult, 0)
	seen := make(map[string]struct{})

	for _, operand := range request.Operands {
		if err := ctx.Err(); err != nil {
			failures = append(failures, domain.ItemResult{
				Path: operand, Status: domain.ItemFailed, Err: err,
			})
			break
		}

		absolute, err := filepath.Abs(operand)
		if err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}
		absolute = filepath.Clean(absolute)

		if _, exists := seen[absolute]; exists {
			continue
		}
		seen[absolute] = struct{}{}

		if err := p.policy.Check(absolute); err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		info, err := os.Lstat(absolute)
		if err != nil {
			if os.IsNotExist(err) && request.Options.Force {
				continue
			}
			failures = append(failures, failed(operand, err))
			continue
		}

		kind := classify(info)
		if kind == domain.TargetDir {
			if !request.Options.Recursive && !request.Options.Directory {
				failures = append(failures, failed(operand, errors.New("is a directory")))
				continue
			}
			if request.Options.Directory && !request.Options.Recursive {
				empty, err := directoryEmpty(absolute)
				if err != nil {
					failures = append(failures, failed(operand, err))
					continue
				}
				if !empty {
					failures = append(failures, failed(operand, errors.New("directory not empty")))
					continue
				}
			}
		}

		deviceID, err := platform.DeviceID(absolute)
		if err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		plan.Targets = append(plan.Targets, domain.PlannedTarget{
			InputPath:    operand,
			AbsolutePath: absolute,
			Kind:         kind,
			DeviceID:     deviceID,
			RequiresWalk: requiresWalk(kind, request.Options),
		})
	}

	return plan, failures
}

func classify(info os.FileInfo) domain.TargetKind {
	if info.Mode()&os.ModeSymlink != 0 {
		return domain.TargetSymlink
	}
	if info.IsDir() {
		return domain.TargetDir
	}
	if info.Mode().IsRegular() {
		return domain.TargetFile
	}
	return domain.TargetOther
}

func requiresWalk(kind domain.TargetKind, options domain.RemoveOptions) bool {
	if kind != domain.TargetDir {
		return false
	}
	return options.Interactive == domain.InteractiveAlways ||
		options.OneFileSystem
}

func directoryEmpty(path string) (bool, error) {
	directory, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer directory.Close()

	_, err = directory.Readdirnames(1)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		if errors.Is(err, os.ErrClosed) {
			return false, err
		}
		if err.Error() == "EOF" {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func failed(path string, err error) domain.ItemResult {
	return domain.ItemResult{Path: path, Status: domain.ItemFailed, Err: err}
}
```

- [x] **Step 4: Replace string-based EOF handling with `io.EOF`**

In `internal/planner/planner.go`, add:

```go
"io"
```

Replace the body of the EOF branch in `directoryEmpty` with:

```go
if errors.Is(err, io.EOF) {
	return true, nil
}
if err != nil {
	return false, err
}
return false, nil
```

The complete `directoryEmpty` function must be:

```go
func directoryEmpty(path string) (bool, error) {
	directory, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer directory.Close()

	_, err = directory.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
```

- [x] **Step 5: Add GNU `--preserve-root=all` tests**

Create `internal/planner/preserve_root_test.go`:

```go
package planner_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
)

func TestValidatePreserveRootAllRejectsDirectoryOnDifferentDeviceFromParent(t *testing.T) {
	t.Parallel()

	path := "/mnt/external"
	deviceID := func(current string) (uint64, error) {
		if current == path {
			return 20, nil
		}
		if current == filepath.Dir(path) {
			return 10, nil
		}
		return 10, nil
	}

	err := planner.ValidatePreserveRootAll(
		path,
		domain.TargetDir,
		domain.PreserveRootAll,
		deviceID,
	)
	if err == nil {
		t.Fatal("ValidatePreserveRootAll() error = nil, want non-nil")
	}
}

func TestValidatePreserveRootAllAllowsRegularFile(t *testing.T) {
	t.Parallel()

	err := planner.ValidatePreserveRootAll(
		"/mnt/external/file.txt",
		domain.TargetFile,
		domain.PreserveRootAll,
		func(string) (uint64, error) { return 20, nil },
	)
	if err != nil {
		t.Fatalf("ValidatePreserveRootAll() error = %v", err)
	}
}
```

- [x] **Step 6: Implement GNU preserve-root-all validation**

Create `internal/planner/preserve_root.go`:

```go
package planner

import (
	"errors"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/domain"
)

// DeviceIDFunc returns a filesystem device ID for a path.
type DeviceIDFunc func(string) (uint64, error)

// ValidatePreserveRootAll rejects directory operands whose parent is on another device.
func ValidatePreserveRootAll(
	path string,
	kind domain.TargetKind,
	mode domain.PreserveRootMode,
	deviceID DeviceIDFunc,
) error {
	if mode != domain.PreserveRootAll || kind != domain.TargetDir {
		return nil
	}

	targetDevice, err := deviceID(path)
	if err != nil {
		return err
	}
	parentDevice, err := deviceID(filepath.Dir(path))
	if err != nil {
		return err
	}
	if targetDevice != parentDevice {
		return errors.New("preserve-root=all protects this directory operand")
	}
	return nil
}
```

In `internal/planner/planner.go`, immediately after `kind := classify(info)`, add:

```go
if err := ValidatePreserveRootAll(
	absolute,
	kind,
	request.Options.PreserveRoot,
	platform.DeviceID,
); err != nil {
	failures = append(failures, failed(operand, err))
	continue
}
```

Run:

```bash
gofmt -w internal/planner
go test ./internal/planner -run TestValidatePreserveRootAll -v
```

Expected:

```text
PASS
```

- [x] **Step 7: Add deterministic traversal helper tests**

Create `internal/planner/walk_test.go`:

```go
package planner_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/planner"
)

func TestWalkDepthFirstDoesNotFollowSymlinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dir", "file"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "dir"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	got, err := planner.WalkDepthFirst(root, 0, false)
	if err != nil {
		t.Fatalf("WalkDepthFirst() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "dir", "file"),
		filepath.Join(root, "dir"),
		filepath.Join(root, "link"),
		root,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}
```

- [x] **Step 8: Implement deterministic depth-first traversal**

Create `internal/planner/walk.go`:

```go
package planner

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/Diaszano/bearm/internal/platform"
)

// WalkDepthFirst returns children before parents and never follows symlinks.
func WalkDepthFirst(root string, rootDevice uint64, oneFileSystem bool) ([]string, error) {
	paths := make([]string, 0)

	var visit func(string) error
	visit = func(path string) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			if oneFileSystem {
				deviceID, err := platform.DeviceID(path)
				if err != nil {
					return err
				}
				if rootDevice != 0 && deviceID != rootDevice {
					return nil
				}
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Name() < entries[j].Name()
			})
			for _, entry := range entries {
				if err := visit(filepath.Join(path, entry.Name())); err != nil {
					return err
				}
			}
		}

		paths = append(paths, path)
		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}
	return paths, nil
}
```

- [x] **Step 9: Run planner tests**

Run:

```bash
gofmt -w internal/planner
go test ./internal/planner -v
```

Expected:

```text
PASS
```

- [x] **Step 10: Commit**

```bash
git add internal/planner
git commit -m "feat: plan safe trash operations"
```

### Task 4: Implement Interactive Prompt Semantics

**Files:**
- Create: `internal/removal/prompt.go`
- Create: `internal/removal/prompt_test.go`

**Interfaces:**
- Consumes: input reader, prompt writer, request, plan.
- Produces:
  - `removal.Prompter`
  - `removal.NewPrompter(reader, writer)`
  - `(*Prompter).ConfirmOnce(plan) (bool, error)`
  - `(*Prompter).ConfirmTarget(path) (bool, error)`

- [x] **Step 1: Write failing prompt tests**

Create `internal/removal/prompt_test.go`:

```go
package removal_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
)

func TestConfirmOnceAcceptsAnswerStartingWithY(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("yes\n"), &output)

	accepted, err := prompter.ConfirmOnce(domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
		},
		Targets: []domain.PlannedTarget{{InputPath: "a"}, {InputPath: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !accepted {
		t.Fatal("accepted = false, want true")
	}
	if output.String() == "" {
		t.Fatal("prompt output is empty")
	}
}

func TestConfirmTargetDefaultsToNo(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("\n"), &output)

	accepted, err := prompter.ConfirmTarget("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if accepted {
		t.Fatal("accepted = true, want false")
	}
}

func TestNeedsOncePrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		plan domain.RemovalPlan
		want bool
	}{
		{
			name: "four targets",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				Targets: make([]domain.PlannedTarget, 4),
			},
			want: true,
		},
		{
			name: "recursive target",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{
						Interactive: domain.InteractiveOnce,
						Recursive:   true,
					},
				},
				Targets: []domain.PlannedTarget{{Kind: domain.TargetDir}},
			},
			want: true,
		},
		{
			name: "three files",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				Targets: make([]domain.PlannedTarget, 3),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := removal.NeedsOncePrompt(tt.plan); got != tt.want {
				t.Fatalf("NeedsOncePrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

- [x] **Step 2: Implement prompt behavior**

Create `internal/removal/prompt.go`:

```go
// Package removal executes validated Bearm removal plans.
package removal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
)

// Prompter handles compatibility confirmation input and output.
type Prompter struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewPrompter creates a confirmation prompter.
func NewPrompter(reader io.Reader, writer io.Writer) *Prompter {
	return &Prompter{reader: bufio.NewReader(reader), writer: writer}
}

// NeedsOncePrompt reports whether -I requires one confirmation.
func NeedsOncePrompt(plan domain.RemovalPlan) bool {
	if plan.Request.Options.Interactive != domain.InteractiveOnce {
		return false
	}
	if len(plan.Targets) > 3 {
		return true
	}
	return plan.Request.Options.Recursive
}

// ConfirmOnce asks for operation-level confirmation.
func (p *Prompter) ConfirmOnce(plan domain.RemovalPlan) (bool, error) {
	fmt.Fprint(p.writer, "rm: remove all arguments? ")
	return p.readYes()
}

// ConfirmTarget asks for one target confirmation.
func (p *Prompter) ConfirmTarget(path string) (bool, error) {
	fmt.Fprintf(p.writer, "rm: remove %s? ", path)
	return p.readYes()
}

func (p *Prompter) readYes() (bool, error) {
	answer, err := p.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer = strings.TrimSpace(answer)
	return len(answer) > 0 && (answer[0] == 'y' || answer[0] == 'Y'), nil
}
```

- [x] **Step 3: Run prompt tests**

Run:

```bash
gofmt -w internal/removal
go test ./internal/removal -run 'TestConfirm|TestNeedsOncePrompt' -v
```

Expected:

```text
PASS
```

- [x] **Step 4: Commit**

```bash
git add internal/removal/prompt.go internal/removal/prompt_test.go
git commit -m "feat: add rm-compatible prompts"
```

### Task 5: Execute Plans Through Backends and Journal Interfaces

**Files:**
- Create: `internal/removal/executor.go`
- Create: `internal/removal/executor_test.go`
- Create: `internal/testutil/backend.go`
- Create: `internal/testutil/journal.go`

**Interfaces:**
- Consumes:
  - `domain.RemovalPlan`
  - `domain.TrashBackend`
  - `removal.Journal`
  - `safety.Policy`
  - `removal.Prompter`.
- Produces:
  - `removal.Executor`
  - `removal.Journal`
  - `(*Executor).Execute(ctx, plan) domain.RemovalResult`

- [x] **Step 1: Create deterministic test doubles**

Create `internal/testutil/backend.go`:

```go
// Package testutil contains reusable Bearm test doubles.
package testutil

import (
	"context"
	"errors"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Backend is a configurable trash backend test double.
type Backend struct {
	Moved       []string
	FailForPath string
}

// Name returns the fake backend name.
func (b *Backend) Name() string {
	return "fake"
}

// Resolve returns a deterministic fake destination.
func (b *Backend) Resolve(_ context.Context, target domain.PlannedTarget) (domain.Destination, error) {
	if target.AbsolutePath == b.FailForPath {
		return domain.Destination{}, errors.New("resolve failed")
	}
	return domain.Destination{Root: "/trash"}, nil
}

// Move records a fake successful move.
func (b *Backend) Move(
	_ context.Context,
	target domain.PlannedTarget,
	_ domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if target.AbsolutePath == b.FailForPath {
		return domain.TrashRecord{}, errors.New("move failed")
	}
	b.Moved = append(b.Moved, target.AbsolutePath)
	return domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-" + target.InputPath,
		OperationID:   operationID,
		OriginalPath:  target.AbsolutePath,
		TrashedPath:   "/trash/" + target.InputPath,
		Backend:       b.Name(),
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
		Status:        "trashed",
	}, nil
}
```

Create `internal/testutil/journal.go`:

```go
package testutil

import (
	"context"

	"github.com/Diaszano/bearm/internal/domain"
)

// Journal is an in-memory journal test double.
type Journal struct {
	Records []domain.TrashRecord
	Err     error
}

// Append stores operation records.
func (j *Journal) Append(_ context.Context, records []domain.TrashRecord) error {
	j.Records = append(j.Records, records...)
	return j.Err
}
```

- [x] **Step 2: Write failing executor tests**

Create `internal/removal/executor_test.go`:

```go
package removal_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestExecuteMovesIndependentTargetsAndAppendsJournal(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Interactive: domain.InteractiveNever,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
			{InputPath: "b", AbsolutePath: "/work/b", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}
	if len(backend.Moved) != 2 || len(journal.Records) != 2 {
		t.Fatalf("moved = %#v, journal = %#v", backend.Moved, journal.Records)
	}
}

func TestExecuteContinuesAfterIndependentFailure(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{FailForPath: "/work/a"}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID:      "operation-1",
		Request: domain.RemoveRequest{Profile: domain.ProfileGNU},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
			{InputPath: "b", AbsolutePath: "/work/b", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
	if len(backend.Moved) != 1 || backend.Moved[0] != "/work/b" {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}

func TestExecuteDeclinesOncePromptWithoutMoves(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader("no\n"), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Interactive: domain.InteractiveOnce,
				Recursive:   true,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "dir", AbsolutePath: "/work/dir", Kind: domain.TargetDir},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}

func TestExecuteReportsJournalFailure(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{Err: errors.New("journal failed")}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID:      "operation-1",
		Request: domain.RemoveRequest{Profile: domain.ProfileGNU},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
}
```

- [x] **Step 3: Implement the executor**

Create `internal/removal/executor.go`:

```go
package removal

import (
	"context"
	"fmt"
	"io"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/safety"
)

// Journal appends durable trash records.
type Journal interface {
	Append(context.Context, []domain.TrashRecord) error
}

// Executor moves planned targets and records completed operations.
type Executor struct {
	backend  domain.TrashBackend
	journal  Journal
	policy   *safety.Policy
	prompter *Prompter
	verbose  io.Writer
}

// NewExecutor creates a removal executor.
func NewExecutor(
	backend domain.TrashBackend,
	journal Journal,
	policy *safety.Policy,
	prompter *Prompter,
	verbose io.Writer,
) *Executor {
	return &Executor{
		backend: backend, journal: journal, policy: policy, prompter: prompter, verbose: verbose,
	}
}

// Execute performs one immutable removal plan.
func (e *Executor) Execute(ctx context.Context, plan domain.RemovalPlan) domain.RemovalResult {
	result := domain.RemovalResult{OperationID: plan.ID}

	if NeedsOncePrompt(plan) {
		accepted, err := e.prompter.ConfirmOnce(plan)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{Status: domain.ItemFailed, Err: err})
			return result
		}
		if !accepted {
			for _, target := range plan.Targets {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemDeclined,
				})
			}
			return result
		}
	}

	records := make([]domain.TrashRecord, 0, len(plan.Targets))
	for _, target := range plan.Targets {
		if err := ctx.Err(); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			break
		}
		if err := e.policy.Check(target.AbsolutePath); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}

		if plan.Request.Options.Interactive == domain.InteractiveAlways {
			accepted, err := e.prompter.ConfirmTarget(target.InputPath)
			if err != nil {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemFailed, Err: err,
				})
				continue
			}
			if !accepted {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemDeclined,
				})
				continue
			}
		}

		destination, err := e.backend.Resolve(ctx, target)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}
		record, err := e.backend.Move(ctx, target, destination, plan.ID)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}

		records = append(records, record)
		recordCopy := record
		result.Items = append(result.Items, domain.ItemResult{
			Path: target.InputPath, Status: domain.ItemTrashed, Record: &recordCopy,
		})
		if plan.Request.Options.Verbose {
			fmt.Fprintln(e.verbose, target.InputPath)
		}
	}

	if len(records) > 0 {
		if err := e.journal.Append(ctx, records); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Status: domain.ItemFailed,
				Err:    fmt.Errorf("append operation journal: %w", err),
			})
		}
	}

	return result
}
```

- [x] **Step 4: Run executor tests**

Run:

```bash
gofmt -w internal/removal internal/testutil
go test ./internal/removal -v
```

Expected:

```text
PASS
```

- [x] **Step 5: Run race tests**

Run:

```bash
go test -race ./internal/removal ./internal/testutil
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add internal/removal internal/testutil
git commit -m "feat: execute safe removal plans"
```

### Task 6: Wire Compatibility Mode to Real Platform Backends

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`
- Create: `internal/app/backend_linux.go`
- Create: `internal/app/backend_darwin.go`
- Create: `internal/app/dependencies.go`

**Interfaces:**
- Consumes:
  - platform directories;
  - platform trash backend;
  - planner and executor;
  - journal interface.
- Produces:
  - `app.Dependencies`
  - `app.NewWithDependencies`
  - functional compatibility-mode removal.

- [x] **Step 1: Define dependency assembly contracts**

Create `internal/app/dependencies.go`:

```go
package app

import (
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
)

// Dependencies contains mutable infrastructure used by the application.
type Dependencies struct {
	Backend domain.TrashBackend
	Journal removal.Journal
	Policy  *safety.Policy
}
```

- [x] **Step 2: Add a compatibility removal application test**

Add to `internal/app/app_test.go`:

```go
func TestRunCompatibilityMovesPlannedTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{Backend: backend, Journal: journal, Policy: policy},
	)

	code := instance.Run(context.Background(), []string{"rm", source})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if len(backend.Moved) != 1 || backend.Moved[0] != source {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}
```

Add the required imports:

```go
"os"
"path/filepath"

"github.com/Diaszano/bearm/internal/safety"
"github.com/Diaszano/bearm/internal/testutil"
```

- [x] **Step 3: Refactor App constructors**

Update `internal/app/app.go` so `App` includes:

```go
dependencies Dependencies
```

Replace `New` with:

```go
// New creates an application without mutable removal infrastructure.
func New(stdin io.Reader, stdout, stderr io.Writer, info buildinfo.Info) *App {
	return NewWithDependencies(stdin, stdout, stderr, info, Dependencies{})
}

// NewWithDependencies creates an application with explicit infrastructure.
func NewWithDependencies(
	stdin io.Reader,
	stdout, stderr io.Writer,
	info buildinfo.Info,
	dependencies Dependencies,
) *App {
	return &App{
		stdin: stdin,
		out: stdout,
		err: stderr,
		info: info,
		dependencies: dependencies,
	}
}
```

- [x] **Step 4: Execute the parsed request**

In `runCompatibility`, replace the final `return 0` with:

```go
if len(request.Operands) == 0 {
	return 0
}
if a.dependencies.Backend == nil || a.dependencies.Journal == nil || a.dependencies.Policy == nil {
	fmt.Fprintln(a.err, "rm: removal infrastructure is not configured")
	return 1
}

instance := planner.New(a.dependencies.Policy, id.New)
plan, planningFailures := instance.Plan(context.Background(), request)
executor := removal.NewExecutor(
	a.dependencies.Backend,
	a.dependencies.Journal,
	a.dependencies.Policy,
	removal.NewPrompter(a.stdin, a.err),
	a.out,
)
result := executor.Execute(context.Background(), plan)
result.Items = append(planningFailures, result.Items...)

for _, item := range result.Items {
	if item.Status == domain.ItemFailed && item.Err != nil {
		fmt.Fprintf(a.err, "rm: %s: %v\n", item.Path, item.Err)
	}
}
return result.ExitCode(profile)
```

Change `runCompatibility` signature to accept context:

```go
func (a *App) runCompatibility(ctx context.Context, args []string) int
```

Pass `ctx` from `Run` and use it in `Plan` and `Execute`.

Add imports:

```go
"github.com/Diaszano/bearm/internal/id"
"github.com/Diaszano/bearm/internal/planner"
"github.com/Diaszano/bearm/internal/removal"
```

- [x] **Step 5: Add platform backend factories**

Create `internal/app/backend_linux.go`:

```go
//go:build linux

package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func newPlatformBackend(home string) domain.TrashBackend {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if !filepath.IsAbs(dataHome) {
		dataHome = filepath.Join(home, ".local", "share")
	}
	return linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(dataHome, "Trash"),
			UID:       os.Getuid(),
			PerMount:  true,
		},
		time.Now,
		id.New,
	)
}
```

Create `internal/app/backend_darwin.go`:

```go
//go:build darwin

package app

import (
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func newPlatformBackend(home string) domain.TrashBackend {
	return darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		id.New,
	)
}
```

Backend ID generation returns errors directly; a random-source failure aborts the move before the reservation is committed.

- [x] **Step 6: Run application tests**

Run:

```bash
gofmt -w internal/app
go test ./internal/app -v
go test -race ./internal/app
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/app
git commit -m "feat: execute compatibility removals"
```

## Plan Completion Verification

Run:

```bash
make verify
```

Then, using a temporary HOME and a test build:

```bash
tmp="$(mktemp -d)"
HOME="$tmp/home"
mkdir -p "$HOME"
mkdir "$tmp/work"
printf 'data' > "$tmp/work/file.txt"
go build -o "$tmp/bearm" ./cmd/bearm
HOME="$HOME" "$tmp/bearm" rm "$tmp/work/file.txt"
test ! -e "$tmp/work/file.txt"
```

Expected:

```text
All verification commands succeed.
The source file is absent from its original path and exists under the
platform trash selected for the temporary HOME.
```
