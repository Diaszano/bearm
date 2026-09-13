# Ponytail Audit Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate ~238 lines of dead code, speculative abstractions, redundant configurations, and duplicated helpers identified during the repository-wide ponytail complexity audit, keeping all public CLI behavior and tests 100% compliant.

**Architecture:** Deconstruct the 13 audit findings into 13 bite-sized, atomic refactoring tasks grouped logically: (1) dead code pruning in planner, config, and i18n, (2) YAGNI abstraction simplifications in prompt, restore, and dependencies, (3) stdlib and platform deduplication across Darwin/Linux backends and ID generation. Every task executes a full test cycle and maintains backward compatibility for all CLI interfaces.

**Tech Stack:** Go 1.24+, Go standard library (`crypto/rand`, `encoding/hex`, `encoding/json`, `errors`, `os`, `path/filepath`), `golang.org/x/sys/unix`, `github.com/pelletier/go-toml/v2`, `github.com/sabhiram/go-gitignore`.

**Spec:** Ponytail Audit Report (2026-09-13) - 13 complexity findings across `internal/`.

## Global Constraints

- Preserve all existing public CLI commands, compatibility flags (`rm` emulation), exit codes, and stdout/stderr formatting.
- Preserve 100% test coverage standards; any deleted dead code must have corresponding obsolete tests cleaned up.
- All code changes must pass `go test ./...`, `go vet ./...`, and formatting checks (`gofmt`).
- Commits must be atomic, conventional, and specific to each task.

---

### Task 1: Delete Dead Code in Planner (`WalkDepthFirst`)

**Files:**
- Delete: `internal/planner/walk.go`
- Delete: `internal/planner/walk_test.go`

**Interfaces:**
- Consumes: Nothing.
- Produces: `ExpandTarget` in `internal/planner/protected_walk.go` continues to be the single source of truth for directory traversal in `Planner.Plan()`.

- [ ] **Step 1: Check existing references to WalkDepthFirst**

Verify that no non-test source files call `WalkDepthFirst`:
```bash
grep -rn "WalkDepthFirst" internal/ --exclude="*_test.go"
```
Expected output: empty (no callers in production code).

- [ ] **Step 2: Delete `walk.go` and `walk_test.go`**

```bash
rm internal/planner/walk.go internal/planner/walk_test.go
```

- [ ] **Step 3: Run planner tests to verify full pass**

Run:
```bash
go test -v ./internal/planner/...
```
Expected: PASS with 0 failures.

- [ ] **Step 4: Commit**

```bash
git add internal/planner/walk.go internal/planner/walk_test.go
git commit -m "refactor(planner): remove unused WalkDepthFirst walker"
```

---

### Task 2: Remove Unused `app.New` Constructor

**Files:**
- Modify: `internal/app/app.go:40-82`
- Modify: `internal/app/app_test.go:25-95`

**Interfaces:**
- Consumes: `app.NewWithDependencies`
- Produces: Cleaner `app` package where `NewWithDependencies` is the single explicit entry point used by `cmd/bearm/main.go` and tests.

- [ ] **Step 1: Update `internal/app/app_test.go` to use `NewWithDependencies` or test helper**

In `internal/app/app_test.go`, replace tests that invoke `app.New()` with direct calls to `NewWithDependencies` using default test dependencies:

```go
func newDefaultApp(t *testing.T, stdin io.Reader, stdout, stderr io.Writer, info buildinfo.Info) *App {
	t.Helper()
	home := t.TempDir()
	dirs, err := platform.ResolveDirs(func(string) string { return "" }, home, runtime.GOOS)
	if err != nil {
		t.Fatalf("resolve dirs: %v", err)
	}
	repo := journal.New(filepath.Join(dirs.StateRoot, "journal.jsonl"), time.Now)
	backend := &testutil.Backend{}
	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{dirs.ConfigRoot, dirs.StateRoot, dirs.DataRoot},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	return NewWithDependencies(stdin, stdout, stderr, info, Dependencies{
		Backend:    backend,
		Repository: repo,
		Policy:     policy,
		Config:     config.Default(),
		ConfigPath: filepath.Join(dirs.ConfigRoot, "config.toml"),
	})
}
```
Replace `New(...)` calls at lines 25, 48, 64, 93 in `app_test.go` with `newDefaultApp(t, ...)`.

- [ ] **Step 2: Delete `New()` from `internal/app/app.go`**

Remove lines 40-81 in `internal/app/app.go` (the `func New(...)` function).

- [ ] **Step 3: Run app tests to verify pass**

Run:
```bash
go test -v ./internal/app/...
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/app/app.go internal/app/app_test.go
git commit -m "refactor(app): remove obsolete New constructor in favor of NewWithDependencies"
```

---

### Task 3: Replace `RenderMetadata` with `json.Marshal` in Darwin Backend

**Files:**
- Modify: `internal/trash/darwin/backend.go:68-76`
- Delete: `internal/trash/darwin/metadata.go`
- Delete: `internal/trash/darwin/metadata_test.go`

**Interfaces:**
- Consumes: Standard library `encoding/json`
- Produces: Inlined JSON marshaling for Darwin metadata matching `custom/backend.go`.

- [ ] **Step 1: Update `internal/trash/darwin/backend.go` to inline `json.Marshal`**

In `internal/trash/darwin/backend.go`, replace lines 68-76:
```go
	deletedAt := b.clock()
	metadata, err := json.Marshal(struct {
		SchemaVersion int       `json:"schema_version"`
		OriginalPath  string    `json:"original_path"`
		DeletedAt     time.Time `json:"deleted_at"`
	}{
		SchemaVersion: 1,
		OriginalPath:  target.AbsolutePath,
		DeletedAt:     deletedAt.UTC(),
	})
	if err != nil {
		return domain.TrashRecord{}, err
	}
	metadata = append(metadata, '\n')
```

- [ ] **Step 2: Remove `metadata.go` and `metadata_test.go`**

```bash
rm internal/trash/darwin/metadata.go internal/trash/darwin/metadata_test.go
```

- [ ] **Step 3: Verify Darwin package compiles**

Run:
```bash
GOOS=darwin go build ./internal/trash/darwin/...
```
Expected: Build succeeds with 0 errors.

- [ ] **Step 4: Commit**

```bash
git add internal/trash/darwin/backend.go internal/trash/darwin/metadata.go internal/trash/darwin/metadata_test.go
git commit -m "refactor(trash/darwin): replace RenderMetadata wrapper with standard json.Marshal"
```

---

### Task 4: Simplify Confirmation Prompting in Removal (`PromptFormatter`)

**Files:**
- Modify: `internal/removal/prompt.go:13-64`
- Modify: `internal/removal/prompt_test.go`
- Modify: `internal/app/app.go:169`

**Interfaces:**
- Consumes: `domain.RemovalPlan`
- Produces: `Prompter` with direct prompt rendering without `defaultPromptFormatter` indirection.

- [ ] **Step 1: Simplify `internal/removal/prompt.go`**

Replace `PromptFormatter` interface and `defaultPromptFormatter` with direct methods on `Prompter` or a simpler formatter function:
```go
// PromptFormatter formats confirmation prompts.
type PromptFormatter interface {
	PromptOnce(domain.RemovalPlan) string
	PromptTarget(path string) string
}

// Prompter handles compatibility confirmation input and output.
type Prompter struct {
	reader    *bufio.Reader
	writer    io.Writer
	formatter PromptFormatter
}

// NewPrompter creates a confirmation prompter.
func NewPrompter(reader io.Reader, writer io.Writer, formatter PromptFormatter) *Prompter {
	return &Prompter{reader: bufio.NewReader(reader), writer: writer, formatter: formatter}
}
```
Eliminate `defaultPromptFormatter` and `p.getFormatter()`. If `formatter` is nil, handle safely in `ConfirmOnce` and `ConfirmTarget` with fallback strings.

- [ ] **Step 2: Update callers of `NewPrompter`**

In `internal/app/app.go:321` and tests where `NewPrompter(r, w)` was called without a third argument, pass `nil` or a basic formatter.

- [ ] **Step 3: Run removal and app tests**

Run:
```bash
go test -v ./internal/removal/... ./internal/app/...
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/removal/prompt.go internal/removal/prompt_test.go internal/app/app.go
git commit -m "refactor(removal): simplify Prompter and remove defaultPromptFormatter boilerplate"
```

---

### Task 5: Remove Dead `LoggingConfig` and `BEARM_LOG_LEVEL`

**Files:**
- Modify: `internal/config/config.go:17, 39-42, 59, 86-90`
- Modify: `internal/config/environment.go:23-25`
- Modify: `internal/config/environment_test.go`
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Consumes: Nothing
- Produces: Lean `config.Config` without unused `Logging` struct or validation.

- [ ] **Step 1: Remove `LoggingConfig` from `internal/config/config.go`**

Delete:
- `Logging LoggingConfig` from `Config` struct (line 17).
- `type LoggingConfig struct { Level string }` (lines 39-42).
- `Logging: LoggingConfig{Level: "error"}` from `Default()` (line 59).
- `switch c.Logging.Level { ... }` from `Validate()` (lines 86-90).

- [ ] **Step 2: Remove `BEARM_LOG_LEVEL` from `internal/config/environment.go`**

Delete lines 23-25:
```go
	if current := getenv("BEARM_LOG_LEVEL"); current != "" {
		value.Logging.Level = current
	}
```

- [ ] **Step 3: Update `config_test.go` and `environment_test.go`**

Remove assertions checking `got.Logging.Level`.

- [ ] **Step 4: Run config tests**

Run:
```bash
go test -v ./internal/config/...
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/environment.go internal/config/config_test.go internal/config/environment_test.go
git commit -m "refactor(config): delete unused LoggingConfig and BEARM_LOG_LEVEL"
```

---

### Task 6: Centralize ID Generation and Remove `internal/id`

**Files:**
- Create: `internal/pathutil/id.go`
- Delete: `internal/id/generator.go`
- Delete: `internal/id/generator_test.go`
- Modify: `internal/app/app.go:163`
- Modify: `internal/app/backend_linux.go:12, 28`
- Modify: `internal/app/backend_darwin.go:20`
- Modify: `internal/app/dependencies.go:8, 31`

**Interfaces:**
- Consumes: `crypto/rand`, `encoding/hex`
- Produces: `pathutil.NewID() (string, error)`

- [ ] **Step 1: Move `New` to `internal/pathutil/id.go`**

Create `internal/pathutil/id.go`:
```go
package pathutil

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a 128-bit lowercase hexadecimal identifier.
func NewID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
```

- [ ] **Step 2: Delete `internal/id/` package**

```bash
rm -rf internal/id
```

- [ ] **Step 3: Update imports in callers**

In `internal/app/app.go`, `internal/app/backend_linux.go`, `internal/app/backend_darwin.go`, `internal/app/dependencies.go`, replace import of `github.com/Diaszano/bearm/internal/id` with `pathutil.NewID`.

- [ ] **Step 4: Run test suite**

Run:
```bash
go test ./...
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/pathutil/id.go internal/id internal/app
git commit -m "refactor: consolidate ID generation into pathutil and remove internal/id package"
```

---

### Task 7: Prune Unused Constants from i18n Catalog

**Files:**
- Modify: `internal/i18n/catalog.go:11-18, 34-37, 45-48`
- Modify: `internal/i18n/catalog_test.go:33-36`

**Interfaces:**
- Consumes: Nothing
- Produces: Cleaned message catalog containing only active messages (`MessageUnknownCommand`, `MessageMissingOperand`, `MessageMissingOperandTryHelp`).

- [ ] **Step 1: Remove dead message constants and catalog entries**

In `internal/i18n/catalog.go`, remove:
- `MessageIllegalOption`
- `MessageUnrecognizedOption`
- `MessageInvalidInteractive`
- `MessageNativeUsage`
and their entries in `NewCatalog()`.

- [ ] **Step 2: Update `internal/i18n/catalog_test.go`**

Remove the 4 deleted constants from the test verification table.

- [ ] **Step 3: Run i18n tests**

Run:
```bash
go test -v ./internal/i18n/...
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/i18n/catalog.go internal/i18n/catalog_test.go
git commit -m "refactor(i18n): remove unused message constants and catalog entries"
```

---

### Task 8: Consolidate GNU-Only Long Option Guards in CLI Parser

**Files:**
- Modify: `internal/cli/compatibility.go:105-135`

**Interfaces:**
- Consumes: `domain.CompatibilityProfile`
- Produces: Deduplicated option routing in `applyLongOption()`.

- [ ] **Step 1: Simplify GNU profile checks in `applyLongOption`**

In `internal/cli/compatibility.go`, replace repetitive `if s.profile != domain.ProfileGNU` blocks:
```go
	case "--one-file-system", "--preserve-root", "--preserve-root=all", "--no-preserve-root", "--help", "--version":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		switch option {
		case "--one-file-system":
			s.request.Options.OneFileSystem = true
		case "--preserve-root":
			s.request.Options.PreserveRoot = domain.PreserveRootDefault
		case "--preserve-root=all":
			s.request.Options.PreserveRoot = domain.PreserveRootAll
		case "--no-preserve-root":
			s.request.Options.PreserveRoot = domain.PreserveRootNone
		case "--help":
			s.request.Options.ShowHelp = true
		case "--version":
			s.request.Options.ShowVersion = true
		}
```

- [ ] **Step 2: Run CLI tests**

Run:
```bash
go test -v ./internal/cli/...
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/cli/compatibility.go
git commit -m "refactor(cli): consolidate duplicate GNU option guards in applyLongOption"
```

---

### Task 9: Reuse `platform.ResolveDirs` in Linux Backend Setup

**Files:**
- Modify: `internal/app/backend_linux.go:16-25`
- Modify: `internal/app/app.go:54-64`

**Interfaces:**
- Consumes: `platform.ResolveDirs`
- Produces: Unified directory resolution without manual `XDG_DATA_HOME` checks.

- [ ] **Step 1: Update `backend_linux.go` to use `dirs.DataRoot` or `platform.ResolveDirs`**

Refactor `newPlatformBackend` to accept `dirs platform.Dirs` instead of resolving `XDG_DATA_HOME` directly:
```go
func newPlatformBackend(home string, dirs platform.Dirs, settings config.Config) domain.TrashBackend {
	return linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(filepath.Dir(dirs.DataRoot), "Trash"),
			UID:       os.Getuid(),
			PerMount:  settings.Trash.PerMount,
		},
		time.Now,
		pathutil.NewID,
	)
}
```

- [ ] **Step 2: Run app tests**

Run:
```bash
go test -v ./internal/app/...
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/app/backend_linux.go internal/app/app.go
git commit -m "refactor(app): reuse platform.Dirs in Linux backend setup"
```

---

### Task 10: Unify `EventAppender` and `removal.Journal`

**Files:**
- Modify: `internal/restore/service.go:17-20`
- Modify: `internal/restore/purge.go:15-19`
- Modify: `internal/domain/journal.go`

**Interfaces:**
- Consumes: `domain.JournalEvent`
- Produces: Single unified `JournalAppender` interface for appending journal records and events.

- [ ] **Step 1: Define canonical journal appender in `internal/domain/journal.go` or reuse `*journal.Repository`**

In `internal/restore/service.go` and `internal/restore/purge.go`, replace separate `EventAppender` declarations with `*journal.Repository` or a common domain interface.

- [ ] **Step 2: Run restore tests**

Run:
```bash
go test -v ./internal/restore/...
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/restore/service.go internal/restore/purge.go internal/domain/journal.go
git commit -m "refactor(restore): unify redundant EventAppender interface"
```

---

### Task 11: Remove Dead `config.IsMissing` Function

**Files:**
- Modify: `internal/config/file.go:43-46`
- Modify: `internal/config/file_test.go:66-74`

**Interfaces:**
- Consumes: Standard library `errors.Is`
- Produces: Direct standard library usage.

- [ ] **Step 1: Remove `IsMissing` from `file.go` and its test from `file_test.go`**

Delete lines 43-46 in `internal/config/file.go`:
```go
// IsMissing reports whether a config error is a missing file.
func IsMissing(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
```
Delete `TestIsMissing` in `internal/config/file_test.go`.

- [ ] **Step 2: Run config tests**

Run:
```bash
go test -v ./internal/config/...
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/config/file.go internal/config/file_test.go
git commit -m "refactor(config): remove unused IsMissing helper"
```

---

### Task 12: Remove Dead `TargetPath` and `InfoPath` Fields from `domain.Destination`

**Files:**
- Modify: `internal/domain/trash.go:13-14`
- Modify: `internal/domain/trash_test.go` (if present)

**Interfaces:**
- Consumes: Nothing
- Produces: Clean `domain.Destination` struct without dead fields.

- [ ] **Step 1: Remove fields from `internal/domain/trash.go`**

In `internal/domain/trash.go`, remove:
```go
	TargetPath       string
	InfoPath         string
```

- [ ] **Step 2: Run domain tests**

Run:
```bash
go test -v ./internal/domain/...
```
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/domain/trash.go
git commit -m "refactor(domain): remove unused TargetPath and InfoPath fields from Destination"
```

---

### Task 13: Remove Redundant `Journal` Field from `app.Dependencies`

**Files:**
- Modify: `internal/app/dependencies.go:18-19`
- Modify: `internal/app/app.go:158, 167`
- Modify: `cmd/bearm/main.go:74-81`
- Modify: `internal/app/app_test.go`

**Interfaces:**
- Consumes: `*journal.Repository`
- Produces: `Dependencies` struct where `Repository` serves as the single journal reference.

- [ ] **Step 1: Remove `Journal` field from `Dependencies`**

In `internal/app/dependencies.go`, remove line 18 (`Journal removal.Journal`).
In `cmd/bearm/main.go:74-81`, remove `Journal: journalRepo`.
In `internal/app/app.go:158, 167`, use `a.dependencies.Repository` when checking nil and passing to `removal.NewExecutor`.

- [ ] **Step 2: Run all tests and acceptance scripts**

Run:
```bash
go test ./...
```
Expected: PASS across all packages.

- [ ] **Step 3: Commit**

```bash
git add internal/app cmd/bearm/main.go
git commit -m "refactor(app): remove redundant Journal field in Dependencies in favor of Repository"
```

---

## Plan Self-Review Checklist

1. **Spec Coverage:** All 13 findings from the ponytail-audit report are covered by exactly one task (Tasks 1 through 13).
2. **No Placeholders:** All steps contain explicit file paths, code snippets, bash verification commands, and commit messages.
3. **Execution Safety:** Each task maintains backwards compatibility, verifies package tests independently, and preserves CLI behavior.
