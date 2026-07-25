# Bearm Configuration and Protection Implementation Plan

**Status:** Complete
**Completion date:** 2026-07-25
**Implementation range:** `ac35fabc68158fb381d56a274b9238e110190531..01931408c25b55e95ae4dd12061ce11388b348f8`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement secure TOML configuration, environment precedence, gitignore-compatible protected patterns, protected-descendant traversal, custom trash support, and native configuration commands.

**Architecture:** Configuration is decoded into typed values before application dependencies are assembled. Safety patterns are compiled once and injected into the planner. Descendant inspection is explicit and converts a top-level recursive target into child-first planned targets while retaining protected nodes and their ancestor directories.

**Tech Stack:** Go 1.24+, `github.com/pelletier/go-toml/v2`, `github.com/sabhiram/go-gitignore`, Go standard library.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Native Bearm user-facing text defaults to `pt-BR`.
- Configuration is declarative TOML and is never executed.
- Unknown TOML fields are rejected.
- Environment overrides configuration; configuration overrides platform defaults.
- Config and directory overrides must be absolute.
- A configuration file must be a regular non-symlink file owned by the current user and not group/world writable.
- Protected patterns use gitignore-compatible syntax.
- Protected-descendant inspection is disabled by default.
- Enabling protected-descendant inspection may traverse recursive targets and must never move a protected descendant through an ancestor rename.
- A custom trash performs only same-filesystem rename; `EXDEV` leaves the source untouched.
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
│   ├── backend_linux.go
│   └── dependencies.go
├── config/
│   ├── config.go
│   ├── config_test.go
│   ├── environment.go
│   ├── environment_test.go
│   ├── file.go
│   ├── file_test.go
│   ├── ownership_unix.go
│   └── ownership_unix_test.go
├── planner/
│   ├── planner.go
│   ├── planner_test.go
│   ├── protected_walk.go
│   └── protected_walk_test.go
├── safety/
│   ├── patterns.go
│   ├── patterns_test.go
│   ├── policy.go
│   └── policy_test.go
└── trash/
    └── custom/
        ├── backend.go
        └── backend_test.go
```

## Dependency Order

```text
Task 1 typed config and defaults
  └── Task 2 secure file loading and environment precedence
          ├── Task 3 protected pattern policy
          ├── Task 4 protected-descendant planning
          ├── Task 5 custom trash backend
          └── Task 6 application/config command integration
```

### Task 1: Define Typed Configuration and Exact Defaults

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Interfaces:**
- Consumes: no process state.
- Produces:
  - `config.Config`
  - `config.Default() Config`
  - `config.Config.Validate() error`
  - exact TOML field names.

- [x] **Step 1: Write failing default and validation tests**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestDefaultConfiguration(t *testing.T) {
	t.Parallel()

	got := config.Default()
	if got.Language != "pt-BR" {
		t.Fatalf("Language = %q", got.Language)
	}
	if got.CompatibilityProfile != "auto" {
		t.Fatalf("CompatibilityProfile = %q", got.CompatibilityProfile)
	}
	if !got.Trash.PerMount {
		t.Fatal("Trash.PerMount = false, want true")
	}
	if !got.Safety.PreserveRoot {
		t.Fatal("Safety.PreserveRoot = false, want true")
	}
	if got.Safety.InspectDescendants {
		t.Fatal("Safety.InspectDescendants = true, want false")
	}
	if got.Restore.CollisionPolicy != "fail" {
		t.Fatalf("CollisionPolicy = %q", got.Restore.CollisionPolicy)
	}
}

func TestValidateRejectsUnknownLanguage(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Language = "es"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestValidateRejectsRelativeCustomTrash(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Trash.CustomPath = "relative/trash"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestValidateRejectsInvalidCollisionPolicy(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Restore.CollisionPolicy = "replace"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
```

- [x] **Step 2: Implement typed configuration**

Create `internal/config/config.go`:

```go
// Package config loads and validates Bearm configuration.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Config is the complete Bearm configuration.
type Config struct {
	Language             string        `toml:"language"`
	CompatibilityProfile string        `toml:"compatibility_profile"`
	Trash                TrashConfig   `toml:"trash"`
	Safety               SafetyConfig  `toml:"safety"`
	Restore              RestoreConfig `toml:"restore"`
	Logging              LoggingConfig `toml:"logging"`
}

// TrashConfig controls platform trash selection.
type TrashConfig struct {
	PerMount   bool   `toml:"per_mount"`
	CustomPath string `toml:"custom_path"`
}

// SafetyConfig controls root, scope, and pattern protection.
type SafetyConfig struct {
	PreserveRoot       bool     `toml:"preserve_root"`
	AllowedRoots       []string `toml:"allowed_roots"`
	ProtectedPatterns  []string `toml:"protected_patterns"`
	InspectDescendants bool     `toml:"inspect_descendants"`
}

// RestoreConfig controls restore collisions.
type RestoreConfig struct {
	CollisionPolicy string `toml:"collision_policy"`
}

// LoggingConfig controls local diagnostic verbosity.
type LoggingConfig struct {
	Level string `toml:"level"`
}

// Default returns exact version 1 platform-independent defaults.
func Default() Config {
	return Config{
		Language:             "pt-BR",
		CompatibilityProfile: "auto",
		Trash: TrashConfig{
			PerMount:                       true,
		},
		Safety: SafetyConfig{
			PreserveRoot:       true,
			AllowedRoots:       []string{},
			ProtectedPatterns:  []string{},
			InspectDescendants: false,
		},
		Restore: RestoreConfig{CollisionPolicy: "fail"},
		Logging: LoggingConfig{Level: "error"},
	}
}

// Validate validates all configuration values.
func (c Config) Validate() error {
	if c.Language != "pt-BR" && c.Language != "en" {
		return fmt.Errorf("unsupported language %q", c.Language)
	}
	switch c.CompatibilityProfile {
	case "auto", "gnu", "bsd", "posix":
	default:
		return fmt.Errorf("unsupported compatibility profile %q", c.CompatibilityProfile)
	}
	if c.Trash.CustomPath != "" && !filepath.IsAbs(c.Trash.CustomPath) {
		return errors.New("trash.custom_path must be absolute")
	}
	for _, root := range c.Safety.AllowedRoots {
		if !filepath.IsAbs(root) {
			return fmt.Errorf("safety.allowed_roots entry must be absolute: %q", root)
		}
	}
	switch c.Restore.CollisionPolicy {
	case "fail", "rename", "overwrite":
	default:
		return fmt.Errorf("unsupported restore collision policy %q", c.Restore.CollisionPolicy)
	}
	switch c.Logging.Level {
	case "error", "debug":
	default:
		return fmt.Errorf("unsupported logging level %q", c.Logging.Level)
	}
	return nil
}
```

- [x] **Step 3: Run config tests**

Run:

```bash
gofmt -w internal/config
go test ./internal/config -run 'TestDefault|TestValidate' -v
```

Expected:

```text
PASS
```

- [x] **Step 4: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: define Bearm configuration"
```

### Task 2: Load TOML Securely and Apply Environment Precedence

**Files:**
- Create: `internal/config/file.go`
- Create: `internal/config/file_test.go`
- Create: `internal/config/environment.go`
- Create: `internal/config/environment_test.go`
- Create: `internal/config/ownership_unix.go`
- Create: `internal/config/ownership_unix_test.go`

**Interfaces:**
- Consumes: config path, environment function, current UID.
- Produces:
  - `config.Load(path, getenv) (Config, error)`
  - `config.CheckSecureFile(path, uid) error`
  - environment keys:
    - `BEARM_LANG`
    - `BEARM_COMPAT`
    - `BEARM_TRASH`
    - `BEARM_TRASH_PER_MOUNT`
    - `BEARM_LOG`.

- [x] **Step 1: Write failing TOML loader tests**

Create `internal/config/file_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestLoadUsesDefaultsWhenFileDoesNotExist(t *testing.T) {
	t.Parallel()

	got, err := config.Load(filepath.Join(t.TempDir(), "missing.toml"), func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != config.Default() {
		t.Fatalf("Load() = %#v, want defaults %#v", got, config.Default())
	}
}

func TestLoadMergesTOMLIntoDefaults(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
language = "en"

[trash]
per_mount = false

[restore]
collision_policy = "rename"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := config.Load(path, func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Language != "en" || got.Trash.PerMount || got.Restore.CollisionPolicy != "rename" {
		t.Fatalf("Load() = %#v", got)
	}
	if !got.Safety.PreserveRoot {
		t.Fatal("default PreserveRoot was not retained")
	}
}

func TestLoadRejectsUnknownTOMLField(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("unknown = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path, func(string) string { return "" }); err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}
```

Because `Config` contains slices and cannot be compared directly, replace the first test equality assertion with:

```go
want := config.Default()
if !reflect.DeepEqual(got, want) {
	t.Fatalf("Load() = %#v, want %#v", got, want)
}
```

Add import:

```go
"reflect"
```

- [x] **Step 2: Write environment precedence tests**

Create `internal/config/environment_test.go`:

```go
package config_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestEnvironmentOverridesFileValues(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"BEARM_LANG":                  "en",
		"BEARM_COMPAT":                "gnu",
		"BEARM_TRASH":                 "/tmp/bearm-trash",
		"BEARM_TRASH_PER_MOUNT":       "false",
		"BEARM_LOG":                   "debug",
	}
	value := config.Default()
	got, err := config.ApplyEnvironment(value, func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("ApplyEnvironment() error = %v", err)
	}

	if got.Language != "en" ||
		got.CompatibilityProfile != "gnu" ||
		got.Trash.CustomPath != "/tmp/bearm-trash" ||
		got.Trash.PerMount ||
		got.Logging.Level != "debug" {
		t.Fatalf("ApplyEnvironment() = %#v", got)
	}
}

func TestEnvironmentRejectsInvalidBoolean(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_TRASH_PER_MOUNT": "sometimes"}
	if _, err := config.ApplyEnvironment(config.Default(), func(key string) string { return env[key] }); err == nil {
		t.Fatal("ApplyEnvironment() error = nil, want non-nil")
	}
}
```

- [x] **Step 3: Implement environment overrides**

Create `internal/config/environment.go`:

```go
package config

import (
	"fmt"
	"strconv"
)

// ApplyEnvironment applies exact Bearm environment overrides.
func ApplyEnvironment(value Config, getenv func(string) string) (Config, error) {
	if current := getenv("BEARM_LANG"); current != "" {
		value.Language = current
	}
	if current := getenv("BEARM_COMPAT"); current != "" {
		value.CompatibilityProfile = current
	}
	if current := getenv("BEARM_TRASH"); current != "" {
		value.Trash.CustomPath = current
	}
	if current := getenv("BEARM_LOG"); current != "" {
		value.Logging.Level = current
	}

	var err error
	if current := getenv("BEARM_TRASH_PER_MOUNT"); current != "" {
		value.Trash.PerMount, err = strconv.ParseBool(current)
		if err != nil {
			return Config{}, fmt.Errorf("BEARM_TRASH_PER_MOUNT: %w", err)
		}
	}

	if err := value.Validate(); err != nil {
		return Config{}, err
	}
	return value, nil
}
```

- [x] **Step 4: Implement secure Unix ownership checks**

Create `internal/config/ownership_unix.go`:

```go
//go:build linux || darwin

package config

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// CheckSecureFile verifies that a config file is regular, non-symlink, user-owned, and not writable by group or others.
func CheckSecureFile(path string, uid int) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("configuration path must be a regular non-symlink file")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("configuration file permissions %o are insecure", info.Mode().Perm())
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("configuration ownership is unavailable")
	}
	if int(stat.Uid) != uid {
		return fmt.Errorf("configuration file owner %d does not match current user %d", stat.Uid, uid)
	}
	return nil
}
```

Create `internal/config/ownership_unix_test.go`:

```go
//go:build linux || darwin

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestCheckSecureFileAcceptsPrivateRegularFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("language = \"pt-BR\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := config.CheckSecureFile(path, os.Getuid()); err != nil {
		t.Fatalf("CheckSecureFile() error = %v", err)
	}
}

func TestCheckSecureFileRejectsGroupWritableFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(""), 0o620); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o620); err != nil {
		t.Fatal(err)
	}
	if err := config.CheckSecureFile(path, os.Getuid()); err == nil {
		t.Fatal("CheckSecureFile() error = nil, want non-nil")
	}
}
```

- [x] **Step 5: Implement strict TOML loading**

Create `internal/config/file.go`:

```go
package config

import (
	"bytes"
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Load loads defaults, strict TOML, environment overrides, and validation.
func Load(path string, getenv func(string) string) (Config, error) {
	value := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ApplyEnvironment(value, getenv)
		}
		return Config{}, err
	}

	if err := CheckSecureFile(path, os.Getuid()); err != nil {
		return Config{}, err
	}

	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return Config{}, err
	}

	value, err = ApplyEnvironment(value, getenv)
	if err != nil {
		return Config{}, err
	}
	if err := value.Validate(); err != nil {
		return Config{}, err
	}
	return value, nil
}

// IsMissing reports whether a config error is a missing file.
func IsMissing(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
```

- [x] **Step 6: Run config tests**

Run:

```bash
gofmt -w internal/config
go test ./internal/config -v
go test -race ./internal/config
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/config
git commit -m "feat: load Bearm configuration securely"
```

### Task 3: Compile and Enforce Protected Patterns

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Create: `internal/safety/patterns.go`
- Create: `internal/safety/patterns_test.go`
- Modify: `internal/safety/policy.go`
- Modify: `internal/safety/policy_test.go`

**Interfaces:**
- Consumes: gitignore-compatible absolute patterns.
- Produces:
  - `safety.PatternMatcher`
  - `safety.CompilePatterns(lines) (*PatternMatcher, error)`
  - `(*PatternMatcher).Matches(absolutePath) bool`
  - pattern-aware `Policy.Check`.

- [x] **Step 1: Add the pattern dependency**

Run:

```bash
go get github.com/sabhiram/go-gitignore@v0.0.0-20210923224102-525f6e181f06
go mod tidy
```

Expected:

```text
go.mod and go.sum include go-gitignore.
```

- [x] **Step 2: Write failing pattern tests**

Create `internal/safety/patterns_test.go`:

```go
package safety_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func TestPatternMatcherMatchesAbsoluteProtectedPath(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{
		"/Users/dias/Documents/important/**",
		"/Users/dias/Projects/**/.git",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !matcher.Matches("/Users/dias/Documents/important/report.pdf") {
		t.Fatal("important report was not matched")
	}
	if !matcher.Matches("/Users/dias/Projects/bearm/.git") {
		t.Fatal(".git path was not matched")
	}
	if matcher.Matches("/Users/dias/Downloads/file.zip") {
		t.Fatal("unprotected path was matched")
	}
}

func TestPolicyRejectsProtectedPattern(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{"/work/**/.git"})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := safety.NewPolicy(safety.Config{Patterns: matcher})
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Check("/work/bearm/.git"); err == nil {
		t.Fatal("Check() error = nil, want protected-pattern failure")
	}
}
```

- [x] **Step 3: Implement the matcher**

Create `internal/safety/patterns.go`:

```go
package safety

import (
	"errors"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// PatternMatcher evaluates gitignore-compatible patterns against absolute paths.
type PatternMatcher struct {
	compiled *ignore.GitIgnore
}

// CompilePatterns compiles absolute gitignore-compatible patterns.
func CompilePatterns(lines []string) (*PatternMatcher, error) {
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "/") && !strings.HasPrefix(line, "!") {
			return nil, errors.New("protected patterns must be absolute or negated absolute patterns")
		}
		normalized = append(normalized, strings.TrimPrefix(line, "/"))
	}
	return &PatternMatcher{compiled: ignore.CompileIgnoreLines(normalized...)}, nil
}

// Matches reports whether absolutePath is protected.
func (m *PatternMatcher) Matches(absolutePath string) bool {
	if m == nil || m.compiled == nil || !filepath.IsAbs(absolutePath) {
		return false
	}
	relativeToRoot := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(absolutePath)), "/")
	return m.compiled.MatchesPath(relativeToRoot)
}
```

- [x] **Step 4: Extend the safety policy**

Add to `safety.Config`:

```go
Patterns           *PatternMatcher
InspectDescendants bool
```

Add to `Policy`:

```go
patterns           *PatternMatcher
inspectDescendants bool
```

In `NewPolicy`, assign both fields.

Before returning success from `Policy.Check`, add:

```go
if p.patterns != nil && p.patterns.Matches(cleaned) {
	return errors.New("path is protected by configured pattern")
}
```

Add:

```go
// InspectDescendants reports whether recursive planning must inspect protected descendants.
func (p *Policy) InspectDescendants() bool {
	return p.inspectDescendants
}
```

- [x] **Step 5: Run pattern and policy tests**

Run:

```bash
gofmt -w internal/safety
go test ./internal/safety -v
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add go.mod go.sum internal/safety
git commit -m "feat: protect configured path patterns"
```

### Task 4: Expand Traversal Without Moving Protected Descendants Through Ancestors

**Files:**
- Create: `internal/planner/protected_walk.go`
- Create: `internal/planner/protected_walk_test.go`
- Modify: `internal/planner/planner.go`
- Modify: `internal/planner/planner_test.go`

**Interfaces:**
- Consumes:
  - recursive top-level target;
  - safety policy;
  - interaction, verbose, and one-filesystem options.
- Produces:
  - `planner.ExpandTarget`
  - child-first planned targets;
  - blocked protected paths and ancestor directories.

- [x] **Step 1: Write failing protected-walk tests**

Create `internal/planner/protected_walk_test.go`:

```go
package planner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
	"github.com/Diaszano/bearm/internal/safety"
)

func TestExpandTargetRetainsProtectedDescendantAndAncestors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "keep", "important.txt")
	movable := filepath.Join(root, "remove", "temporary.txt")
	for _, path := range []string{protected, movable} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	matcher, err := safety.CompilePatterns([]string{filepath.ToSlash(protected)})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := safety.NewPolicy(safety.Config{
		Patterns:           matcher,
		InspectDescendants: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	targets, skipped, err := planner.ExpandTarget(
		domain.PlannedTarget{
			InputPath: root, AbsolutePath: root, Kind: domain.TargetDir,
		},
		domain.RemoveOptions{Recursive: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}

	paths := map[string]bool{}
	for _, target := range targets {
		paths[target.AbsolutePath] = true
	}
	if paths[protected] || paths[filepath.Dir(protected)] || paths[root] {
		t.Fatalf("protected path or ancestor planned: %#v", paths)
	}
	if !paths[movable] || !paths[filepath.Dir(movable)] {
		t.Fatalf("movable subtree missing: %#v", paths)
	}
	if len(skipped) != 1 || skipped[0].Path != protected {
		t.Fatalf("skipped = %#v", skipped)
	}
}

func TestExpandTargetVerbosePlansEveryEntry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	file := filepath.Join(root, "sub", "file.txt")
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, _ := safety.NewPolicy(safety.Config{})
	targets, _, err := planner.ExpandTarget(
		domain.PlannedTarget{InputPath: root, AbsolutePath: root, Kind: domain.TargetDir},
		domain.RemoveOptions{Recursive: true, Verbose: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 {
		t.Fatalf("targets = %#v, want file, subdirectory, root", targets)
	}
}
```

- [x] **Step 2: Implement protected traversal expansion**

Create `internal/planner/protected_walk.go`:

```go
package planner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/safety"
)

// ExpandTarget expands a traversal-required directory into child-first targets.
func ExpandTarget(
	root domain.PlannedTarget,
	options domain.RemoveOptions,
	policy *safety.Policy,
) ([]domain.PlannedTarget, []domain.ItemResult, error) {
	type entry struct {
		path     string
		kind     domain.TargetKind
		deviceID uint64
	}

	entries := make([]entry, 0)
	protected := make(map[string]struct{})
	skipped := make([]domain.ItemResult, 0)

	var visit func(string) error
	visit = func(path string) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		deviceID, err := platform.DeviceID(path)
		if err != nil {
			return err
		}
		if options.OneFileSystem && root.DeviceID != 0 && deviceID != root.DeviceID {
			return nil
		}

		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			children, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			sort.Slice(children, func(i, j int) bool {
				return children[i].Name() < children[j].Name()
			})
			for _, child := range children {
				if err := visit(filepath.Join(path, child.Name())); err != nil {
					return err
				}
			}
		}

		kind := classify(info)
		if err := policy.Check(path); err != nil {
			protected[path] = struct{}{}
			skipped = append(skipped, domain.ItemResult{
				Path: path, Status: domain.ItemSkipped, Err: err,
			})
		}
		entries = append(entries, entry{path: path, kind: kind, deviceID: deviceID})
		return nil
	}

	if err := visit(root.AbsolutePath); err != nil {
		return nil, nil, err
	}

	blocked := make(map[string]struct{})
	for path := range protected {
		current := path
		for isEqualOrDescendant(current, root.AbsolutePath) {
			blocked[current] = struct{}{}
			if current == root.AbsolutePath {
				break
			}
			current = filepath.Dir(current)
		}
	}

	targets := make([]domain.PlannedTarget, 0, len(entries))
	for _, current := range entries {
		if _, isBlocked := blocked[current.path]; isBlocked {
			continue
		}
		targets = append(targets, domain.PlannedTarget{
			InputPath:    current.path,
			AbsolutePath: current.path,
			Kind:         current.kind,
			DeviceID:     current.deviceID,
			RequiresWalk: false,
		})
	}
	return targets, skipped, nil
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

- [x] **Step 3: Mark verbose and protected inspection as traversal-required**

Change `requiresWalk` in `internal/planner/planner.go` to accept the policy:

```go
func requiresWalk(
	kind domain.TargetKind,
	options domain.RemoveOptions,
	policy *safety.Policy,
) bool {
	if kind != domain.TargetDir {
		return false
	}
	return options.Interactive == domain.InteractiveAlways ||
		options.Verbose ||
		options.OneFileSystem ||
		policy.InspectDescendants()
}
```

Update its caller:

```go
RequiresWalk: requiresWalk(kind, request.Options, p.policy),
```

- [x] **Step 4: Expand traversal-required targets during planning**

After constructing a `PlannedTarget`, replace direct append with:

```go
planned := domain.PlannedTarget{
	InputPath:    operand,
	AbsolutePath: absolute,
	Kind:         kind,
	DeviceID:     deviceID,
	RequiresWalk: requiresWalk(kind, request.Options, p.policy),
}
if planned.RequiresWalk {
	expanded, skipped, err := ExpandTarget(planned, request.Options, p.policy)
	if err != nil {
		failures = append(failures, failed(operand, err))
		continue
	}
	plan.Targets = append(plan.Targets, expanded...)
	failures = append(failures, skipped...)
	continue
}
plan.Targets = append(plan.Targets, planned)
```

Change skipped protected descendants from failures to non-failing results when the application combines planning outcomes. `RemovalResult.HasFailures` already ignores `ItemSkipped`.

- [x] **Step 5: Run planner and removal tests**

Run:

```bash
gofmt -w internal/planner
go test ./internal/planner ./internal/removal -v
```

Expected:

```text
PASS
```

- [x] **Step 6: Add an integration test proving protected content remains**

Append to `internal/planner/protected_walk_test.go`:

```go
func TestProtectedPlanNeverContainsAncestorThatWouldMoveProtectedContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "keep", "file.txt")
	if err := os.MkdirAll(filepath.Dir(protected), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(protected, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	matcher, _ := safety.CompilePatterns([]string{filepath.ToSlash(protected)})
	policy, _ := safety.NewPolicy(safety.Config{Patterns: matcher, InspectDescendants: true})
	targets, _, err := planner.ExpandTarget(
		domain.PlannedTarget{AbsolutePath: root, Kind: domain.TargetDir},
		domain.RemoveOptions{Recursive: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range targets {
		relative, err := filepath.Rel(target.AbsolutePath, protected)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("planned ancestor %q would move protected %q", target.AbsolutePath, protected)
		}
	}
}
```

Add import:

```go
"strings"
```

Run:

```bash
go test ./internal/planner -run TestProtected -v
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/planner
git commit -m "feat: retain protected recursive descendants"
```

### Task 5: Add a Same-Filesystem Custom Trash Backend

**Files:**
- Create: `internal/trash/custom/backend.go`
- Create: `internal/trash/custom/backend_test.go`

**Interfaces:**
- Consumes: absolute custom root, clock, ID generator.
- Produces:
  - `custom.Backend`
  - implementation of `domain.TrashBackend`.

- [x] **Step 1: Write failing custom-backend tests**

Create `internal/trash/custom/backend_test.go`:

```go
package custom_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	customtrash "github.com/Diaszano/bearm/internal/trash/custom"
)

func TestBackendMovesFileIntoCustomTrash(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "work", "file.txt")
	trashRoot := filepath.Join(root, "custom-trash")
	if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend, err := customtrash.NewBackend(
		trashRoot,
		func() time.Time { return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC) },
		func() (string, error) { return "item-1", nil },
	)
	if err != nil {
		t.Fatal(err)
	}

	target := domain.PlannedTarget{
		InputPath: source, AbsolutePath: source, Kind: domain.TargetFile,
	}
	destination, err := backend.Resolve(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	record, err := backend.Move(context.Background(), target, destination, "operation-1")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(record.TrashedPath); err != nil {
		t.Fatalf("trashed path missing: %v", err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
}

func TestNewBackendRejectsRelativeRoot(t *testing.T) {
	t.Parallel()

	if _, err := customtrash.NewBackend("relative", time.Now, func() (string, error) { return "id", nil }); err == nil {
		t.Fatal("NewBackend() error = nil, want non-nil")
	}
}
```

- [x] **Step 2: Implement custom backend**

Create `internal/trash/custom/backend.go`:

```go
// Package custom implements a same-filesystem user-selected trash.
package custom

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

// Backend moves targets into a configured trash root.
type Backend struct {
	root        string
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a custom trash backend.
func NewBackend(root string, clock func() time.Time, idGenerator func() (string, error)) (*Backend, error) {
	if !filepath.IsAbs(root) {
		return nil, errors.New("custom trash root must be absolute")
	}
	return &Backend{root: filepath.Clean(root), clock: clock, idGenerator: idGenerator}, nil
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "custom-trash"
}

// Resolve creates and validates the custom trash skeleton.
func (b *Backend) Resolve(_ context.Context, _ domain.PlannedTarget) (domain.Destination, error) {
	filesDir := filepath.Join(b.root, "files")
	infoDir := filepath.Join(b.root, "info")
	for _, directory := range []string{b.root, filesDir, infoDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return domain.Destination{}, err
		}
		info, err := os.Lstat(directory)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return domain.Destination{}, errors.New("custom trash contains an unsafe directory")
		}
	}
	return domain.Destination{
		Root: b.root, FilesDir: filesDir, InfoDir: infoDir,
	}, nil
}

// Move renames one target into the custom trash.
func (b *Backend) Move(
	ctx context.Context,
	target domain.PlannedTarget,
	destination domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrashRecord{}, err
	}

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

	reservation, err := trash.ReserveName(
		destination.FilesDir,
		destination.InfoDir,
		filepath.Base(target.AbsolutePath),
		".json",
		metadata,
	)
	if err != nil {
		return domain.TrashRecord{}, err
	}
	defer func() { _ = reservation.Rollback() }()

	if err := os.Rename(target.AbsolutePath, reservation.TargetPath); err != nil {
		return domain.TrashRecord{}, err
	}

	itemID, err := b.idGenerator()
	if err != nil {
		return domain.TrashRecord{}, err
	}

	reservation.Commit()
	return domain.TrashRecord{
		SchemaVersion: 1,
		ItemID: itemID, OperationID: operationID,
		OriginalPath: target.AbsolutePath, TrashedPath: reservation.TargetPath,
		Backend: b.Name(), DeviceID: target.DeviceID,
		DeletedAt: deletedAt.UTC(), Status: "trashed",
	}, nil
}
```

- [x] **Step 3: Run custom backend tests**

Run:

```bash
gofmt -w internal/trash/custom
go test ./internal/trash/custom -v
```

Expected:

```text
PASS
```

- [x] **Step 4: Commit**

```bash
git add internal/trash/custom
git commit -m "feat: add custom trash backend"
```

### Task 6: Assemble Configuration, Safety, Backends, and Native Config Commands

**Files:**
- Modify: `cmd/bearm/main.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`
- Modify: `internal/app/backend_linux.go`
- Modify: `internal/app/backend_darwin.go`
- Modify: `internal/app/dependencies.go`

**Interfaces:**
- Consumes:
  - `config.Config`;
  - platform directories;
  - compiled patterns;
  - platform/custom backend selection.
- Produces:
  - `bearm config path`;
  - `bearm config check`;
  - config-driven language, profile, trash, safety, and restore policy.

- [x] **Step 1: Extend application dependencies with config path and values**

Add to `internal/app/dependencies.go`:

```go
Config     config.Config
ConfigPath string
```

Add import:

```go
"github.com/Diaszano/bearm/internal/config"
```

- [x] **Step 2: Implement native config commands**

In `internal/app/app.go`, handle `cli.CommandConfig` before requiring the repository:

```go
if request.Command == cli.CommandConfig {
	switch request.ConfigOp {
	case "path":
		fmt.Fprintln(a.out, a.dependencies.ConfigPath)
		return 0
	case "check":
		if err := a.dependencies.Config.Validate(); err != nil {
			fmt.Fprintf(a.err, "bearm: configuração inválida: %v\n", err)
			return 3
		}
		fmt.Fprintln(a.out, "Configuração válida.")
		return 0
	default:
		fmt.Fprintln(a.err, "bearm: operação de configuração inválida")
		return 2
	}
}
```

- [x] **Step 3: Use configured compatibility profile**

Change profile resolution so explicit config is used before host auto-detection:

```go
func resolveConfiguredProfile(
	configured string,
	getenv func(string) string,
	goos string,
) domain.CompatibilityProfile {
	if configured != "" && configured != "auto" {
		switch configured {
		case "gnu":
			return domain.ProfileGNU
		case "bsd":
			return domain.ProfileBSD
		case "posix":
			return domain.ProfilePOSIX
		}
	}
	return resolveProfile(getenv, goos)
}
```

Use:

```go
profile := resolveConfiguredProfile(
	a.dependencies.Config.CompatibilityProfile,
	os.Getenv,
	runtime.GOOS,
)
```

- [x] **Step 4: Assemble configuration in `main`**

In `cmd/bearm/main.go`, after resolving platform directories:

```go
configPath := filepath.Join(dirs.ConfigRoot, "config.toml")
settings, err := config.Load(configPath, os.Getenv)
if err != nil {
	fmt.Fprintf(os.Stderr, "bearm: configuração inválida: %v\n", err)
	os.Exit(3)
}

patterns, err := safety.CompilePatterns(settings.Safety.ProtectedPatterns)
if err != nil {
	fmt.Fprintf(os.Stderr, "bearm: padrões de proteção inválidos: %v\n", err)
	os.Exit(3)
}

policy, err := safety.NewPolicy(safety.Config{
	HardProtectedRoots: []string{dirs.ConfigRoot, dirs.StateRoot, dirs.DataRoot},
	AllowedRoots:       settings.Safety.AllowedRoots,
	Patterns:           patterns,
	InspectDescendants: settings.Safety.InspectDescendants,
})
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(3)
}
```

Replace backend assembly with:

```go
backend, err := app.NewConfiguredBackend(home, settings)
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(3)
}
```

Set dependencies:

```go
dependencies := app.Dependencies{
	Backend:    backend,
	Journal:    repository,
	Repository: repository,
	Policy:     policy,
	Config:     settings,
	ConfigPath: configPath,
}
```

Add import:

```go
"github.com/Diaszano/bearm/internal/config"
```

- [x] **Step 5: Implement configured backend factory**

Add a cross-platform function to `internal/app/dependencies.go`:

```go
// NewConfiguredBackend creates a platform or custom trash backend.
func NewConfiguredBackend(home string, settings config.Config) (domain.TrashBackend, error) {
	if settings.Trash.CustomPath != "" {
		return customtrash.NewBackend(
			settings.Trash.CustomPath,
			time.Now,
			id.New,
		)
	}
	return newPlatformBackend(home, settings), nil
}

```

Add imports:

```go
"time"

"github.com/Diaszano/bearm/internal/id"
customtrash "github.com/Diaszano/bearm/internal/trash/custom"
```

Change platform factory signatures:

```go
func newPlatformBackend(home string, settings config.Config) domain.TrashBackend
```

On Linux, use:

```go
PerMount: settings.Trash.PerMount,
```

On macOS, accept `settings` and explicitly ignore `PerMount` because system-visible per-volume routing is always enabled:

```go
_ = settings
```

Delete `NewPlatformBackend` because `NewConfiguredBackend` replaces it.

- [x] **Step 6: Use configured restore collision policy**

In `runRestore`, map the configured value:

```go
policy := restore.CollisionPolicy(a.dependencies.Config.Restore.CollisionPolicy)
if policy == restore.CollisionOverwrite {
	fmt.Fprintln(a.err, "bearm: overwrite exige confirmação explícita e não é usado por restore padrão")
	return 2
}
service := restore.NewService(a.dependencies.Repository, time.Now)
results := service.Restore(ctx, records, policy)
```

- [x] **Step 7: Add config command integration tests**

Add to `internal/app/app_test.go`:

```go
func TestRunConfigPath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Config:     config.Default(),
			ConfigPath: "/home/dias/.config/bearm/config.toml",
		},
	)

	code := instance.Run(context.Background(), []string{"bearm", "config", "path"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "/home/dias/.config/bearm/config.toml" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunConfigCheck(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Config:     config.Default(),
			ConfigPath: "/tmp/config.toml",
		},
	)

	code := instance.Run(context.Background(), []string{"bearm", "config", "check"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Configuração válida") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
```

Add import:

```go
"github.com/Diaszano/bearm/internal/config"
```

- [x] **Step 8: Run complete verification**

Run:

```bash
gofmt -w cmd/bearm internal/app
go test ./internal/config ./internal/safety ./internal/planner ./internal/trash/custom ./internal/app
go test -race ./...
go vet ./...
```

Expected:

```text
PASS
```

- [x] **Step 9: Commit**

```bash
git add cmd/bearm internal/app
git commit -m "feat: apply Bearm configuration"
```

## Plan Completion Verification

Run:

```bash
make verify
```

Create an isolated config:

```bash
tmp="$(mktemp -d)"
export HOME="$tmp/home"
export BEARM_CONFIG_HOME="$tmp/config"
export BEARM_STATE_HOME="$tmp/state"
mkdir -p "$HOME" "$BEARM_CONFIG_HOME"
cat > "$BEARM_CONFIG_HOME/config.toml" <<'EOF'
language = "pt-BR"
compatibility_profile = "gnu"

[trash]
per_mount = true
custom_path = ""

[safety]
preserve_root = true
allowed_roots = []
protected_patterns = []
inspect_descendants = false

[restore]
collision_policy = "fail"

[logging]
level = "error"
EOF
chmod 600 "$BEARM_CONFIG_HOME/config.toml"

go run ./cmd/bearm config path
go run ./cmd/bearm config check
```

Expected:

```text
The first command prints the exact config path.
The second command prints "Configuração válida." and exits 0.
Invalid permissions, unknown fields, relative directories, and invalid enums
exit with native configuration code 3 before any filesystem mutation.
```
