# Bearm Compatibility, Hardening, and Release Implementation Plan

**Status:** Complete
**Completion date:** 2026-07-25
**Implementation range:** `ac35fabc68158fb381d56a274b9238e110190531..01931408c25b55e95ae4dd12061ce11388b348f8`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify observable `rm` compatibility on Linux and macOS, harden edge cases through fuzzing and benchmarks, and publish production-ready documentation and release artifacts.

**Architecture:** A differential harness executes the host `/bin/rm` and a test-built Bearm against isolated fixture trees, then compares exit codes, streams, prompts, and normalized source-tree outcomes. Host-specific compatibility renderers own exact diagnostics. Release automation runs the complete acceptance matrix before creating signed checksums and archives.

**Tech Stack:** Go 1.24+, Go test subprocesses, golden fixtures, fuzzing, benchmarks, GitHub Actions, GoReleaser, Markdown, roff man page.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Compatibility behavior is tested against the host `/bin/rm`.
- Bearm-native output defaults to `pt-BR`.
- Compatibility diagnostics are written to stderr and success remains silent unless verbose mode is active.
- Tests operate only inside isolated temporary directories.
- Differential tests compare normalized source-tree outcomes because Bearm trashes while `/bin/rm` unlinks.
- Never weaken hard protection of `/`, `.`, `..`, Bearm trash roots, state roots, or config roots to satisfy a compatibility fixture.
- Release artifacts support Linux and macOS on amd64 and arm64.
- Release builds use `-trimpath` and have no telemetry.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
bearm/
├── internal/
│   ├── cli/
│   │   ├── render_compatibility.go
│   │   └── render_compatibility_test.go
│   └── planner/
│       └── fuzz_test.go
├── test/
│   ├── compatibility/
│   │   ├── cases.go
│   │   ├── harness_test.go
│   │   ├── normalize.go
│   │   ├── testmain_test.go
│   │   └── testdata/
│   │       ├── bsd/
│   │       └── gnu/
│   ├── integration/
│   │   ├── acceptance_test.go
│   │   ├── signal_test.go
│   │   └── security_test.go
│   └── fixtures/
├── docs/
│   ├── architecture/overview.md
│   ├── compatibility/bsd.md
│   ├── compatibility/gnu.md
│   ├── configuration.md
│   ├── recovery.md
│   └── security.md
├── man/bearm.1
├── scripts/acceptance.sh
├── .github/workflows/
│   ├── ci.yml
│   ├── release.yml
│   └── security.yml
├── .goreleaser.yaml
├── CHANGELOG.md
├── CONTRIBUTING.md
├── README.md
└── SECURITY.md
```

## Dependency Order

```text
Task 1 typed compatibility rendering
  └── Task 2 differential harness
          ├── Task 3 edge-case and security fixtures
          ├── Task 4 fuzzing and performance budgets
          ├── Task 5 documentation and packaging
          └── Task 6 release and final acceptance
```

### Task 1: Render Profile-Specific Compatibility Diagnostics

**Files:**
- Create: `internal/cli/render_compatibility.go`
- Create: `internal/cli/render_compatibility_test.go`
- Modify: `internal/app/app.go`

**Interfaces:**
- Consumes:
  - `domain.CompatibilityProfile`
  - `i18n.Language`
  - typed parser, planner, and executor errors.
- Produces:
  - `cli.CompatibilityRenderer`
  - `cli.NewCompatibilityRenderer(profile, language, program)`
  - `Usage() string`
  - `MissingOperand() string`
  - `UnsupportedOption(error) string`
  - `PathError(path, error) string`
  - `PromptOnce() string`
  - `PromptTarget(path) string`.

- [x] **Step 1: Write failing renderer tests**

Create `internal/cli/render_compatibility_test.go`:

```go
package cli_test

import (
	"errors"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
)

func TestGNURendererMissingOperand(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	if got := renderer.MissingOperand(); got != "rm: missing operand\n" {
		t.Fatalf("MissingOperand() = %q", got)
	}
}

func TestBSDRendererUsage(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")
	want := "usage: rm [-f | -i] [-dIRrv] file ...\n"
	if got := renderer.Usage(); got != want {
		t.Fatalf("Usage() = %q, want %q", got, want)
	}
}

func TestGNURendererUnrecognizedLongOption(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	got := renderer.UnsupportedOption(&cli.UsageError{
		Kind: "unsupported-option", Option: "--unknown",
	})
	want := "rm: unrecognized option '--unknown'\n"
	if got != want {
		t.Fatalf("UnsupportedOption() = %q, want %q", got, want)
	}
}

func TestBSDRendererIllegalShortOption(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")
	got := renderer.UnsupportedOption(&cli.UsageError{
		Kind: "unsupported-option", Option: "-P",
	})
	want := "rm: illegal option -- P\n"
	if got != want {
		t.Fatalf("UnsupportedOption() = %q, want %q", got, want)
	}
}

func TestRendererPathError(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	got := renderer.PathError("missing", errors.New("No such file or directory"))
	if got != "rm: cannot remove 'missing': No such file or directory\n" {
		t.Fatalf("PathError() = %q", got)
	}
}
```

- [x] **Step 2: Implement compatibility rendering**

Create `internal/cli/render_compatibility.go`:

```go
package cli

import (
	"fmt"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
)

// CompatibilityRenderer renders profile-specific rm diagnostics.
type CompatibilityRenderer struct {
	profile  domain.CompatibilityProfile
	language i18n.Language
	program  string
}

// NewCompatibilityRenderer creates a deterministic compatibility renderer.
func NewCompatibilityRenderer(
	profile domain.CompatibilityProfile,
	language i18n.Language,
	program string,
) CompatibilityRenderer {
	return CompatibilityRenderer{
		profile: profile,
		language: language,
		program: program,
	}
}

// Usage renders the selected profile usage text.
func (r CompatibilityRenderer) Usage() string {
	if r.profile == domain.ProfileBSD {
		return fmt.Sprintf("usage: %s [-f | -i] [-dIRrv] file ...\n", r.program)
	}
	return fmt.Sprintf("Usage: %s [OPTION]... [FILE]...\n", r.program)
}

// MissingOperand renders a missing operand diagnostic.
func (r CompatibilityRenderer) MissingOperand() string {
	if r.language == i18n.LanguagePTBR {
		return fmt.Sprintf("%s: operando ausente\n", r.program)
	}
	return fmt.Sprintf("%s: missing operand\n", r.program)
}

// UnsupportedOption renders a typed parser usage error.
func (r CompatibilityRenderer) UnsupportedOption(err *UsageError) string {
	if r.profile == domain.ProfileBSD && strings.HasPrefix(err.Option, "-") && !strings.HasPrefix(err.Option, "--") {
		return fmt.Sprintf("%s: illegal option -- %s\n", r.program, strings.TrimPrefix(err.Option, "-"))
	}
	if strings.HasPrefix(err.Option, "--") {
		return fmt.Sprintf("%s: unrecognized option '%s'\n", r.program, err.Option)
	}
	return fmt.Sprintf("%s: invalid option -- '%s'\n", r.program, strings.TrimPrefix(err.Option, "-"))
}

// PathError renders an operational path error.
func (r CompatibilityRenderer) PathError(path string, err error) string {
	if r.profile == domain.ProfileBSD {
		return fmt.Sprintf("%s: %s: %v\n", r.program, path, err)
	}
	return fmt.Sprintf("%s: cannot remove '%s': %v\n", r.program, path, err)
}

// PromptOnce renders the once-only confirmation prompt.
func (r CompatibilityRenderer) PromptOnce() string {
	return fmt.Sprintf("%s: remove all arguments? ", r.program)
}

// PromptTarget renders a per-target confirmation prompt.
func (r CompatibilityRenderer) PromptTarget(path string) string {
	return fmt.Sprintf("%s: remove '%s'? ", r.program, path)
}
```

- [x] **Step 3: Run renderer tests**

Run:

```bash
gofmt -w internal/cli
go test ./internal/cli -run Test.*Renderer -v
```

Expected:

```text
PASS
```

- [x] **Step 4: Replace ad hoc compatibility rendering in the application**

In `internal/app/app.go`, create the renderer immediately after resolving profile:

```go
language := i18n.ResolveCompatibilityLanguage(os.Getenv)
renderer := cli.NewCompatibilityRenderer(profile, language, "rm")
```

Replace parser error rendering with:

```go
var usageErr *cli.UsageError
if errors.As(err, &usageErr) {
	fmt.Fprint(a.err, renderer.UnsupportedOption(usageErr))
	fmt.Fprint(a.err, renderer.Usage())
	return usageCode(profile)
}
fmt.Fprintf(a.err, "rm: %v\n", err)
return usageCode(profile)
```

Replace missing operand rendering with:

```go
fmt.Fprint(a.err, renderer.MissingOperand())
if profile == domain.ProfileBSD {
	fmt.Fprint(a.err, renderer.Usage())
}
return usageCode(profile)
```

Replace failed-item output with:

```go
fmt.Fprint(a.err, renderer.PathError(item.Path, item.Err))
```

Inject renderer prompt strings into `removal.Prompter` by extending `NewPrompter` to accept a small prompt-text interface or explicit strings. Update prompt tests to preserve the exact defaults.

- [x] **Step 5: Run app and CLI tests**

Run:

```bash
gofmt -w internal/app internal/removal
go test ./internal/cli ./internal/removal ./internal/app -v
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add internal/cli internal/app internal/removal
git commit -m "feat: render rm-compatible diagnostics"
```

### Task 2: Build the Differential Host `rm` Harness

**Files:**
- Create: `test/compatibility/cases.go`
- Create: `test/compatibility/normalize.go`
- Create: `test/compatibility/testmain_test.go`
- Create: `test/compatibility/harness_test.go`

**Interfaces:**
- Consumes:
  - `/bin/rm`;
  - test-built Bearm;
  - isolated fixture setup and stdin.
- Produces:
  - `compatibility.Case`
  - normalized process outcomes;
  - host-specific differential tests.

- [x] **Step 1: Define compatibility cases**

Create `test/compatibility/cases.go`:

```go
package compatibility

import (
	"os"
	"path/filepath"
)

// Case is one host rm differential scenario.
type Case struct {
	Name       string
	Args       []string
	Stdin      string
	Setup      func(root string) error
	CompareOut bool
	CompareErr bool
}

// CoreCases returns safe isolated compatibility scenarios.
func CoreCases() []Case {
	return []Case{
		{
			Name:       "missing operand",
			CompareErr: true,
		},
		{
			Name:       "force missing operand",
			Args:       []string{"-f"},
			CompareOut: true,
			CompareErr: true,
		},
		{
			Name:       "remove regular file",
			Args:       []string{"file.txt"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "remove missing file",
			Args:       []string{"missing.txt"},
			CompareOut: true,
			CompareErr: true,
		},
		{
			Name:       "recursive directory",
			Args:       []string{"-rf", "directory"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "directory", "sub"), 0o700); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(root, "directory", "sub", "file"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "directory without recursive",
			Args:       []string{"directory"},
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.Mkdir(filepath.Join(root, "directory"), 0o700)
			},
		},
		{
			Name:       "double dash filename",
			Args:       []string{"--", "-rf"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "-rf"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "force then interactive declined",
			Args:       []string{"-f", "-i", "file.txt"},
			Stdin:      "n\n",
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "interactive then force",
			Args:       []string{"-i", "-f", "file.txt"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
	}
}
```

- [x] **Step 2: Implement source-tree normalization**

Create `test/compatibility/normalize.go`:

```go
package compatibility

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// TreeEntry is one normalized fixture path after execution.
type TreeEntry struct {
	Path string
	Mode fs.FileMode
	Link string
}

// NormalizeTree returns the logical source-tree state and ignores Bearm internals.
func NormalizeTree(root string) ([]TreeEntry, error) {
	entries := make([]TreeEntry, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == ".bearm-trash" || relative == ".bearm-state" ||
			hasPrefix(relative, ".bearm-trash") || hasPrefix(relative, ".bearm-state") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		item := TreeEntry{Path: filepath.ToSlash(relative), Mode: info.Mode()}
		if info.Mode()&os.ModeSymlink != 0 {
			item.Link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		entries = append(entries, item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func hasPrefix(path, root string) bool {
	return path != root && len(path) > len(root) && path[:len(root)] == root &&
		(path[len(root)] == '/' || path[len(root)] == filepath.Separator)
}
```

- [x] **Step 3: Build Bearm once for subprocess tests**

Create `test/compatibility/testmain_test.go`:

```go
package compatibility

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var bearmBinary string

func TestMain(m *testing.M) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	buildDir, err := os.MkdirTemp("", "bearm-compat-build-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(buildDir)

	bearmBinary = filepath.Join(buildDir, "bearm")
	command := exec.Command("go", "build", "-trimpath", "-o", bearmBinary, "./cmd/bearm")
	command.Dir = root
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}
```

- [x] **Step 4: Write the differential harness**

Create `test/compatibility/harness_test.go`:

```go
package compatibility

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type processResult struct {
	Code   int
	Stdout string
	Stderr string
	Tree   []TreeEntry
}

func TestCoreCompatibility(t *testing.T) {
	for _, testCase := range CoreCases() {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			rmResult := runCase(t, "/bin/rm", false, testCase)
			bearmResult := runCase(t, bearmBinary, true, testCase)

			if rmResult.Code != bearmResult.Code {
				t.Fatalf("exit code: rm=%d bearm=%d\nrm stderr=%q\nbearm stderr=%q",
					rmResult.Code, bearmResult.Code, rmResult.Stderr, bearmResult.Stderr)
			}
			if testCase.CompareOut && rmResult.Stdout != bearmResult.Stdout {
				t.Fatalf("stdout:\nrm=%q\nbearm=%q", rmResult.Stdout, bearmResult.Stdout)
			}
			if testCase.CompareErr && normalizeProgram(rmResult.Stderr) != normalizeProgram(bearmResult.Stderr) {
				t.Fatalf("stderr:\nrm=%q\nbearm=%q", rmResult.Stderr, bearmResult.Stderr)
			}
			if !reflect.DeepEqual(rmResult.Tree, bearmResult.Tree) {
				t.Fatalf("tree:\nrm=%#v\nbearm=%#v", rmResult.Tree, bearmResult.Tree)
			}
		})
	}
}

func runCase(t *testing.T, binary string, bearm bool, testCase Case) processResult {
	t.Helper()

	root := t.TempDir()
	if testCase.Setup != nil {
		if err := testCase.Setup(root); err != nil {
			t.Fatal(err)
		}
	}

	args := append([]string(nil), testCase.Args...)
	if bearm {
		args = append([]string{"rm"}, args...)
	}

	command := exec.Command(binary, args...)
	command.Dir = root
	command.Stdin = strings.NewReader(testCase.Stdin)
	command.Env = append(os.Environ(),
		"LC_ALL=C",
		"LANG=C",
		"BEARM_LANG=en",
		"BEARM_COMPAT="+hostProfile(),
		"BEARM_CONFIG_HOME="+filepath.Join(root, ".bearm-config"),
		"BEARM_STATE_HOME="+filepath.Join(root, ".bearm-state"),
		"BEARM_TRASH="+filepath.Join(root, ".bearm-trash"),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	code := 0
	if err := command.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %s: %v", binary, err)
		}
		code = exitErr.ExitCode()
	}

	tree, err := NormalizeTree(root)
	if err != nil {
		t.Fatal(err)
	}
	return processResult{
		Code: code, Stdout: stdout.String(), Stderr: stderr.String(), Tree: tree,
	}
}

func hostProfile() string {
	if runtime.GOOS == "darwin" {
		return "bsd"
	}
	return "gnu"
}

func normalizeProgram(value string) string {
	return strings.ReplaceAll(value, bearmBinary, "rm")
}
```

- [x] **Step 5: Run the first differential suite**

Run:

```bash
go test ./test/compatibility -v
```

Expected:

```text
Any mismatch fails with exit code, stream, or normalized tree details.
Fix compatibility rendering and semantics until all core cases pass.
```

- [x] **Step 6: Commit**

```bash
git add test/compatibility
git commit -m "test: add rm differential harness"
```

### Task 3: Add Edge-Case, Security, and Signal Fixtures

**Files:**
- Modify: `test/compatibility/cases.go`
- Create: `test/integration/security_test.go`
- Create: `test/integration/signal_test.go`
- Create: `test/integration/acceptance_test.go`

**Interfaces:**
- Consumes: test-built Bearm and isolated temporary trees.
- Produces: regression coverage for symlinks, unusual filenames, permission failures, path protection, concurrent operations, and interruption.

- [x] **Step 1: Add unusual filename and symlink differential cases**

Append to `CoreCases()`:

```go
{
	Name:       "filename with spaces and unicode",
	Args:       []string{"ação com espaço.txt"},
	CompareOut: true,
	CompareErr: true,
	Setup: func(root string) error {
		return os.WriteFile(filepath.Join(root, "ação com espaço.txt"), []byte("data"), 0o600)
	},
},
{
	Name:       "dangling symlink",
	Args:       []string{"dangling"},
	CompareOut: true,
	CompareErr: true,
	Setup: func(root string) error {
		return os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "dangling"))
	},
},
{
	Name:       "trailing slash on file",
	Args:       []string{"file.txt/"},
	CompareOut: true,
	CompareErr: false,
	Setup: func(root string) error {
		return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
	},
},
{
	Name:       "empty directory with d",
	Args:       []string{"-d", "empty"},
	CompareOut: true,
	CompareErr: false,
	Setup: func(root string) error {
		return os.Mkdir(filepath.Join(root, "empty"), 0o700)
	},
},
{
	Name:       "interactive once four operands declined",
	Args:       []string{"-I", "a", "b", "c", "d"},
	Stdin:      "n\n",
	CompareOut: true,
	CompareErr: false,
	Setup: func(root string) error {
		for _, name := range []string{"a", "b", "c", "d"} {
			if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o600); err != nil {
				return err
			}
		}
		return nil
	},
},
```

- [x] **Step 2: Add hard-protection integration tests**

Create `test/integration/security_test.go`:

```go
package integration_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestCompatibilityCannotRemoveHardProtectedRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "state")
	if err := os.MkdirAll(protected, 0o700); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{protected},
	})
	if err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: &testutil.Backend{},
			Journal: &testutil.Journal{},
			Policy: policy,
			Config: config.Default(),
		},
	)

	code := instance.Run(context.Background(), []string{"rm", "-rf", protected})
	if code == 0 {
		t.Fatalf("Run() code = 0, stderr = %q", stderr.String())
	}
	if _, err := os.Stat(protected); err != nil {
		t.Fatalf("protected root changed: %v", err)
	}
}

func TestCompatibilityCannotRemoveSystemRootWithNoPreserveRoot(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: &testutil.Backend{},
			Journal: &testutil.Journal{},
			Policy: policy,
			Config: config.Default(),
		},
	)

	code := instance.Run(context.Background(), []string{"rm", "-rf", "--no-preserve-root", "/"})
	if code == 0 {
		t.Fatal("Run() code = 0, want nonzero")
	}
}
```

- [x] **Step 3: Add subprocess cancellation coverage**

Create `test/integration/signal_test.go`:

```go
package integration_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestCancelledContextPreventsNewMoves(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: backend,
			Journal: &testutil.Journal{},
			Policy: policy,
			Config: config.Default(),
		},
	)

	code := instance.Run(ctx, []string{"rm", "file"})
	if code == 0 {
		t.Fatal("Run() code = 0, want nonzero")
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("ctx.Err() = %v", ctx.Err())
	}
}

var _ = domain.ProfileGNU
```

Remove the final `var _` and unused domain import after the test compiles; it is not part of the committed result.

- [x] **Step 4: Add an end-to-end acceptance test**

Create `test/integration/acceptance_test.go`:

```go
package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemoveListRestoreRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess acceptance test")
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	binary := filepath.Join(root, "bearm")
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/bearm")
	build.Dir = repositoryRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build error = %v\n%s", err, output)
	}

	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work")
	configRoot := filepath.Join(root, "config")
	stateRoot := filepath.Join(root, "state")
	trashRoot := filepath.Join(root, "trash")
	for _, directory := range []string{home, work, configRoot, stateRoot, trashRoot} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	source := filepath.Join(work, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(),
		"HOME="+home,
		"BEARM_CONFIG_HOME="+configRoot,
		"BEARM_STATE_HOME="+stateRoot,
		"BEARM_TRASH="+trashRoot,
		"BEARM_COMPAT="+map[bool]string{true: "bsd", false: "gnu"}[runtime.GOOS == "darwin"],
	)

	run := func(args ...string) string {
		command := exec.Command(binary, args...)
		command.Env = env
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%s error = %v\n%s", strings.Join(args, " "), err, output)
		}
		return string(output)
	}

	run("rm", source)
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	if output := run("list", "--json"); !strings.Contains(output, source) {
		t.Fatalf("list output = %q", output)
	}
	run("restore", "--last")
	if got, err := os.ReadFile(source); err != nil || string(got) != "data" {
		t.Fatalf("restored data = %q, error = %v", got, err)
	}
}
```

- [x] **Step 5: Run edge and security tests**

Run:

```bash
go test ./test/compatibility -v
go test ./test/integration -v
go test -race ./test/integration
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add test
git commit -m "test: cover Bearm safety edge cases"
```

### Task 4: Add Fuzzing and Performance Budgets

**Files:**
- Create: `internal/planner/fuzz_test.go`
- Create: `internal/cli/benchmark_test.go`
- Create: `internal/journal/benchmark_test.go`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: parser, path planner, journal repository.
- Produces: fuzz targets and stable benchmark reporting.

- [x] **Step 1: Add planner fuzzing**

Create `internal/planner/fuzz_test.go`:

```go
package planner_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func FuzzSafetyPolicyCheck(f *testing.F) {
	for _, seed := range []string{
		"/",
		"/tmp/file",
		"/tmp/../etc",
		"/Users/dias/project/link",
		"/home/dias/a b",
	} {
		f.Add(seed)
	}

	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{"/tmp/bearm-state"},
	})
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if !filepath.IsAbs(input) {
			return
		}
		_ = policy.Check(input)
	})
}
```

- [x] **Step 2: Add parser benchmarks**

Create `internal/cli/benchmark_test.go`:

```go
package cli_test

import (
	"fmt"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func BenchmarkParseCompatibility(b *testing.B) {
	args := []string{"-rfv", "--one-file-system", "directory", "file.txt"}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := cli.ParseCompatibility(args, domain.ProfileGNU); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTenThousandOperands(b *testing.B) {
	args := make([]string, 10001)
	args[0] = "-f"
	for index := 1; index < len(args); index++ {
		args[index] = fmt.Sprintf("file-%d", index)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := cli.ParseCompatibility(args, domain.ProfileGNU); err != nil {
			b.Fatal(err)
		}
	}
}
```

- [x] **Step 3: Add journal benchmarks**

Create `internal/journal/benchmark_test.go`:

```go
package journal_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
)

func BenchmarkAppendOperation(b *testing.B) {
	path := filepath.Join(b.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	records := make([]domain.TrashRecord, 100)
	for index := range records {
		records[index] = domain.TrashRecord{
			SchemaVersion: 1,
			ItemID:        fmt.Sprintf("item-%d", index),
			OperationID:   "operation",
			OriginalPath:  fmt.Sprintf("/work/%d", index),
			TrashedPath:   fmt.Sprintf("/trash/%d", index),
			Backend:       "benchmark",
			DeletedAt:     time.Now(),
			Status:        "trashed",
		}
	}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := repository.Append(context.Background(), records); err != nil {
			b.Fatal(err)
		}
	}
}
```

- [x] **Step 4: Add a CI fuzz smoke job**

Append to `.github/workflows/ci.yml`:

```yaml
  fuzz-smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: go test ./internal/cli -fuzz=FuzzParseCompatibility -fuzztime=20s
      - run: go test ./internal/cli -fuzz=FuzzParseNative -fuzztime=20s
      - run: go test ./internal/planner -fuzz=FuzzSafetyPolicyCheck -fuzztime=20s

  benchmarks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: go test ./internal/cli ./internal/journal -run '^$' -bench . -benchmem
```

- [x] **Step 5: Run local fuzz seeds and benchmarks**

Run:

```bash
go test ./internal/cli ./internal/planner
go test ./internal/cli ./internal/journal -run '^$' -bench . -benchmem
```

Expected:

```text
All tests pass and benchmark output reports allocations and time.
```

- [x] **Step 6: Commit**

```bash
git add internal/cli/benchmark_test.go internal/journal/benchmark_test.go internal/planner/fuzz_test.go .github/workflows/ci.yml
git commit -m "test: add Bearm fuzz and performance coverage"
```

### Task 5: Write Production Documentation and Packaging

**Files:**
- Modify: `README.md`
- Create: `docs/architecture/overview.md`
- Create: `docs/compatibility/gnu.md`
- Create: `docs/compatibility/bsd.md`
- Create: `docs/configuration.md`
- Create: `docs/recovery.md`
- Create: `docs/security.md`
- Create: `man/bearm.1`
- Create: `CHANGELOG.md`
- Create: `CONTRIBUTING.md`
- Create: `SECURITY.md`

**Interfaces:**
- Consumes: final commands, config schema, limitations, and release artifact naming.
- Produces: install, alias, recovery, security, contribution, and man-page documentation.

- [x] **Step 1: Replace README with final user documentation**

`README.md` must contain these exact top-level sections:

```markdown
# Bearm

Bearm is a native safe replacement for Unix `rm`. It moves files, symlinks,
and directories to operating-system trash locations instead of permanently
deleting them.

## Features
## Supported Platforms
## Installation
## Safe First Run
## Using Bearm Directly
## Using Bearm as `rm`
## Restore and Purge
## Configuration
## Compatibility
## Safety Guarantees
## Known Limitations
## Development
## License
```

The alias section must recommend a reversible shell alias first:

```bash
alias rm='bearm rm'
```

It must explicitly warn users not to replace `/bin/rm`, and it must document how to bypass the alias:

```bash
command rm
/bin/rm
```

- [x] **Step 2: Document architecture and compatibility**

Create `docs/architecture/overview.md` with:

- package dependency direction;
- parse → validate → plan → prompt → execute → journal flow;
- no-traversal fast path;
- platform backend boundaries;
- atomic metadata reservation;
- restore and purge lifecycle events.

Create `docs/compatibility/gnu.md` and `docs/compatibility/bsd.md` with:

- supported options;
- option-order behavior;
- diagnostics and exit-code differences;
- intentionally unsupported options;
- hard Bearm protections that remain stricter than host `rm`;
- host versions used by CI.

- [x] **Step 3: Document configuration and recovery**

Create `docs/configuration.md` containing the exact version 1 TOML schema:

```toml
language = "pt-BR"
compatibility_profile = "auto"

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
```

Create `docs/recovery.md` documenting:

```bash
bearm list
bearm list --json
bearm restore ITEM_ID
bearm restore --operation OPERATION_ID
bearm restore --last
bearm purge ITEM_ID
bearm purge ITEM_ID --yes
bearm doctor
```

Explain that Finder **Put Back** is not guaranteed and Bearm restore is authoritative.

- [x] **Step 4: Document security and contribution rules**

Create `docs/security.md` and `SECURITY.md` covering:

- threat model;
- no shell execution;
- symlink rules;
- config ownership/permissions;
- collision reservation;
- no telemetry;
- private vulnerability reporting process.

Create `CONTRIBUTING.md` with:

- Go 1.24+;
- `make verify`;
- TDD expectation;
- Conventional Commits;
- conventional English branch names;
- one focused commit per task;
- no AI attribution trailers;
- compatibility fixture requirements for behavior changes.

- [x] **Step 5: Add the man page**

Create `man/bearm.1` with sections:

```text
NAME
SYNOPSIS
DESCRIPTION
COMPATIBILITY MODE
NATIVE COMMANDS
OPTIONS
CONFIGURATION
ENVIRONMENT
FILES
EXIT STATUS
SAFETY
LIMITATIONS
EXAMPLES
SEE ALSO
```

Verify:

```bash
man -l man/bearm.1
```

Expected:

```text
The man page renders without roff errors.
```

- [x] **Step 6: Configure automatic Homebrew tap publishing**

Extend `.goreleaser.yaml` with:

```yaml
homebrew_casks:
  - name: bearm
    repository:
      owner: Diaszano
      name: homebrew-tap
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/Diaszano/bearm"
    description: "Safe native replacement for Unix rm with restore support"
    license: "MIT"
    binaries:
      - bearm
    manpages:
      - man/bearm.1
```

Document the following installation note in `README.md`: Bearm never replaces
`/bin/rm` automatically; users may invoke `bearm rm FILE` directly or create the
reversible shell alias `alias rm='bearm rm'` themselves.

Create the `Diaszano/homebrew-tap` repository before the first public release and
add repository secret `HOMEBREW_TAP_GITHUB_TOKEN` with permission to update that
tap. The release workflow must fail clearly when this secret is unavailable; it
must not publish an incomplete formula.

- [x] **Step 7: Add changelog**

Create `CHANGELOG.md`:

```markdown
# Changelog

All notable changes to Bearm are documented in this file.

The format is based on Keep a Changelog, and the project follows Semantic
Versioning.

## [Unreleased]

### Added

- Native GNU/BSD-aware `rm` compatibility mode.
- Linux FreeDesktop and macOS system-visible Trash backends.
- Atomic collision-safe metadata reservation.
- Append-only operation journal.
- List, restore, purge, doctor, and config commands.
- Secure TOML configuration and protected path patterns.
```

- [x] **Step 8: Run documentation link and command checks**

Run:

```bash
python3 - <<'PY'
from pathlib import Path

blocked = ["T" + "ODO", "T" + "BD", "example" + ".com", "<" + "owner" + ">"]
paths = [
    Path("README.md"),
    Path("docs"),
    Path("man"),
    Path("CONTRIBUTING.md"),
    Path("SECURITY.md"),
    Path("CHANGELOG.md"),
]
matches = []
for path in paths:
    files = [path] if path.is_file() else list(path.rglob("*"))
    for file in files:
        if not file.is_file():
            continue
        text = file.read_text(encoding="utf-8")
        for marker in blocked:
            if marker in text:
                matches.append(f"{file}: contains blocked marker {marker!r}")
if matches:
    raise SystemExit("\n".join(matches))
PY
go run ./cmd/bearm version
go run ./cmd/bearm config path
```

Expected:

```text
The marker scan prints no errors.
Both commands succeed.
```

- [x] **Step 9: Commit**

```bash
git add README.md docs man .goreleaser.yaml CHANGELOG.md CONTRIBUTING.md SECURITY.md
git commit -m "docs: document Bearm usage and recovery"
```

### Task 6: Add Release, Security, and Final Acceptance Automation

**Files:**
- Create: `scripts/acceptance.sh`
- Create: `.github/workflows/release.yml`
- Create: `.github/workflows/security.yml`
- Modify: `.goreleaser.yaml`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: complete repository.
- Produces:
  - one local acceptance command;
  - tag-triggered release workflow;
  - dependency and vulnerability scanning;
  - archives, checksums, SBOMs, and release notes.

- [x] **Step 1: Create the final acceptance script**

Create `scripts/acceptance.sh`:

```sh
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
  goreleaser release --snapshot --clean
fi
```

Run:

```bash
chmod +x scripts/acceptance.sh
./scripts/acceptance.sh
```

Expected:

```text
All installed verification tools exit 0.
```

- [x] **Step 2: Add release metadata and SBOM generation**

Extend `.goreleaser.yaml`:

```yaml
sboms:
  - artifacts: archive

release:
  prerelease: auto

source:
  enabled: true

snapshot:
  version_template: "{{ incpatch .Version }}-next"

signs:
  - cmd: cosign
    artifacts: checksum
    signature: "${artifact}.sigstore.json"
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - --yes
```

Keep signing in release CI only; snapshot builds may skip signing when `cosign` identity is unavailable.

- [x] **Step 3: Add tag-triggered release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write
  id-token: write
  attestations: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - name: Verify
        run: |
          go mod verify
          test -z "$(gofmt -l .)"
          go vet ./...
          go test ./...
          go test -race ./...
          govulncheck ./...
      - uses: sigstore/cosign-installer@v4
      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

- [x] **Step 4: Add security workflow**

Create `.github/workflows/security.yml`:

```yaml
name: Security

on:
  pull_request:
  push:
    branches: [main]
  schedule:
    - cron: "17 8 * * 1"

permissions:
  contents: read
  security-events: write

jobs:
  govulncheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: govulncheck ./...

  codeql:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      packages: read
      contents: read
    steps:
      - uses: actions/checkout@v6
      - uses: github/codeql-action/init@v4
        with:
          languages: go
          queries: security-extended
      - uses: github/codeql-action/autobuild@v4
      - uses: github/codeql-action/analyze@v4
```

- [x] **Step 5: Require acceptance jobs in CI**

Ensure `.github/workflows/ci.yml` includes:

```yaml
  acceptance:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: ./scripts/acceptance.sh
```

Do not run GoReleaser in this job unless installed; the script already checks availability.

- [x] **Step 6: Run final local acceptance**

Run:

```bash
./scripts/acceptance.sh
git status --short
```

Expected:

```text
All checks pass.
The working tree is clean after committing generated source changes.
No release archive is committed.
```

- [x] **Step 7: Commit**

```bash
git add scripts/acceptance.sh .github/workflows .goreleaser.yaml
git commit -m "chore: add Bearm release pipeline"
```

## Plan Completion Verification

Run on both Linux and macOS:

```bash
./scripts/acceptance.sh
```

Create and test an isolated alias in a disposable shell:

```bash
tmp="$(mktemp -d)"
go build -trimpath -o "$tmp/bearm" ./cmd/bearm
HOME="$tmp/home" BEARM_STATE_HOME="$tmp/state" BEARM_TRASH="$tmp/trash" \
  sh -c '
    mkdir -p "$HOME" "'"$tmp"'/work"
    printf data > "'"$tmp"'/work/file.txt"
    alias rm="'"$tmp"'/bearm rm"
    rm "'"$tmp"'/work/file.txt"
    test ! -e "'"$tmp"'/work/file.txt"
    "'"$tmp"'/bearm" restore --last
    test -f "'"$tmp"'/work/file.txt"
  '
```

Expected:

```text
The acceptance suite passes on Linux and macOS.
The disposable alias removes to trash and restores the file.
Release snapshot archives exist only under dist/.
```
