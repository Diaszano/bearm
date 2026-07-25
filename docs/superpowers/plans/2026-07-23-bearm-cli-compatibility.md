# Bearm CLI Compatibility Implementation Plan

**Status:** Complete
**Completion date:** 2026-07-25
**Implementation range:** `ac35fabc68158fb381d56a274b9238e110190531..01931408c25b55e95ae4dd12061ce11388b348f8`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement deterministic invocation detection, ordered GNU/BSD `rm` parsing, typed diagnostics, locale selection, and native command parsing without touching the filesystem.

**Architecture:** `internal/cli` translates raw `argv` into typed requests. Parsing is pure and order-sensitive. `internal/i18n` owns all user-facing strings, while `internal/app` only selects parsers and renders typed parse errors.

**Tech Stack:** Go 1.24+, Go standard library, table-driven tests, fuzz tests.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Native Bearm user-facing text defaults to `pt-BR`; compatibility text is profile and locale aware.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- No shell execution from application code.
- Parsing must be pure and perform no filesystem access.
- Preserve command-line option order.
- Never silently ignore unsupported options.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
internal/
├── app/
│   ├── app.go
│   └── app_test.go
├── cli/
│   ├── invocation.go
│   ├── invocation_test.go
│   ├── compatibility.go
│   ├── compatibility_test.go
│   ├── native.go
│   ├── native_test.go
│   ├── errors.go
│   └── fuzz_test.go
├── domain/
│   └── removal.go
└── i18n/
    ├── catalog.go
    ├── catalog_test.go
    ├── locale.go
    └── locale_test.go
```

## Dependency Order

```text
Task 1 invocation resolver
  ├── Task 2 locale and catalogs
  ├── Task 3 ordered compatibility parser
  └── Task 4 native parser
          └── Task 5 application integration and fuzzing
```

### Task 1: Detect Native and Compatibility Invocation

**Files:**
- Create: `internal/cli/invocation.go`
- Create: `internal/cli/invocation_test.go`

**Interfaces:**
- Consumes: raw `argv`.
- Produces:
  - `cli.InvocationMode`
  - `cli.Invocation`
  - `cli.ResolveInvocation(argv []string) (Invocation, error)`

- [x] **Step 1: Write failing invocation tests**

Create `internal/cli/invocation_test.go`:

```go
package cli_test

import (
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
)

func TestResolveInvocationUsesRMModeFromArgvZero(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"/usr/local/bin/rm", "-rf", "build"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeCompatibility {
		t.Fatalf("Mode = %q", got.Mode)
	}
	want := []string{"-rf", "build"}
	if !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("Args = %#v, want %#v", got.Args, want)
	}
}

func TestResolveInvocationUsesRMSubcommand(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"bearm", "rm", "-f", "missing"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeCompatibility {
		t.Fatalf("Mode = %q", got.Mode)
	}
	want := []string{"-f", "missing"}
	if !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("Args = %#v, want %#v", got.Args, want)
	}
}

func TestResolveInvocationUsesNativeMode(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"bearm", "list", "--limit", "10"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeNative {
		t.Fatalf("Mode = %q", got.Mode)
	}
}

func TestResolveInvocationRejectsEmptyArgv(t *testing.T) {
	t.Parallel()

	if _, err := cli.ResolveInvocation(nil); err == nil {
		t.Fatal("ResolveInvocation() error = nil, want non-nil")
	}
}
```

- [x] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/cli -run TestResolveInvocation -v
```

Expected:

```text
FAIL because ResolveInvocation does not exist.
```

- [x] **Step 3: Implement invocation detection**

Create `internal/cli/invocation.go`:

```go
// Package cli parses Bearm native and rm-compatible command lines.
package cli

import (
	"errors"
	"path/filepath"
)

// InvocationMode identifies the parser selected for an invocation.
type InvocationMode string

const (
	// ModeCompatibility selects rm-compatible parsing.
	ModeCompatibility InvocationMode = "compatibility"
	// ModeNative selects Bearm native subcommands.
	ModeNative InvocationMode = "native"
)

// Invocation contains the mode and arguments after executable/subcommand removal.
type Invocation struct {
	Program string
	Mode    InvocationMode
	Args    []string
}

// ResolveInvocation determines whether argv requests compatibility or native mode.
func ResolveInvocation(argv []string) (Invocation, error) {
	if len(argv) == 0 {
		return Invocation{}, errors.New("argv must include the executable name")
	}

	program := filepath.Base(argv[0])
	if program == "rm" {
		return Invocation{
			Program: program,
			Mode:    ModeCompatibility,
			Args:    append([]string(nil), argv[1:]...),
		}, nil
	}

	if len(argv) > 1 && argv[1] == "rm" {
		return Invocation{
			Program: "rm",
			Mode:    ModeCompatibility,
			Args:    append([]string(nil), argv[2:]...),
		}, nil
	}

	return Invocation{
		Program: program,
		Mode:    ModeNative,
		Args:    append([]string(nil), argv[1:]...),
	}, nil
}
```

- [x] **Step 4: Run tests**

Run:

```bash
gofmt -w internal/cli
go test ./internal/cli -run TestResolveInvocation -v
```

Expected:

```text
PASS
```

- [x] **Step 5: Commit**

```bash
git add internal/cli/invocation.go internal/cli/invocation_test.go
git commit -m "feat: detect Bearm invocation mode"
```

### Task 2: Add Locale Selection and Typed Message Catalogs

**Files:**
- Create: `internal/i18n/locale.go`
- Create: `internal/i18n/locale_test.go`
- Create: `internal/i18n/catalog.go`
- Create: `internal/i18n/catalog_test.go`

**Interfaces:**
- Consumes: `LC_ALL`, `LC_MESSAGES`, `LANG`, `BEARM_LANG`.
- Produces:
  - `i18n.Language`
  - `i18n.ResolveNativeLanguage(getenv)`
  - `i18n.ResolveCompatibilityLanguage(getenv)`
  - `i18n.Catalog`
  - `i18n.NewCatalog(language)`

- [x] **Step 1: Write failing locale tests**

Create `internal/i18n/locale_test.go`:

```go
package i18n_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/i18n"
)

func TestResolveNativeLanguageDefaultsToBrazilianPortuguese(t *testing.T) {
	t.Parallel()

	got := i18n.ResolveNativeLanguage(func(string) string { return "" })
	if got != i18n.LanguagePTBR {
		t.Fatalf("language = %q", got)
	}
}

func TestResolveNativeLanguageHonorsBearmLanguage(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_LANG": "en"}
	got := i18n.ResolveNativeLanguage(func(key string) string { return env[key] })
	if got != i18n.LanguageEN {
		t.Fatalf("language = %q", got)
	}
}

func TestResolveCompatibilityLanguageUsesLocalePrecedence(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"LANG":        "en_US.UTF-8",
		"LC_MESSAGES": "pt_BR.UTF-8",
		"LC_ALL":      "",
	}
	got := i18n.ResolveCompatibilityLanguage(func(key string) string { return env[key] })
	if got != i18n.LanguagePTBR {
		t.Fatalf("language = %q", got)
	}
}
```

- [x] **Step 2: Write failing catalog tests**

Create `internal/i18n/catalog_test.go`:

```go
package i18n_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/i18n"
)

func TestCatalogPortugueseUnknownCommand(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(i18n.LanguagePTBR)
	if got := catalog.Text(i18n.MessageUnknownCommand); got != "comando desconhecido" {
		t.Fatalf("Text() = %q", got)
	}
}

func TestCatalogEnglishMissingOperand(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(i18n.LanguageEN)
	if got := catalog.Text(i18n.MessageMissingOperand); got != "missing operand" {
		t.Fatalf("Text() = %q", got)
	}
}
```

- [x] **Step 3: Run tests and verify failure**

Run:

```bash
go test ./internal/i18n -v
```

Expected:

```text
FAIL because package internal/i18n does not exist.
```

- [x] **Step 4: Implement locale selection**

Create `internal/i18n/locale.go`:

```go
// Package i18n owns Bearm message catalogs and locale selection.
package i18n

import "strings"

// Language identifies a supported Bearm language.
type Language string

const (
	// LanguagePTBR selects Brazilian Portuguese.
	LanguagePTBR Language = "pt-BR"
	// LanguageEN selects English.
	LanguageEN Language = "en"
)

// ResolveNativeLanguage resolves the language for Bearm-native commands.
func ResolveNativeLanguage(getenv func(string) string) Language {
	if value := normalizeLanguage(getenv("BEARM_LANG")); value != "" {
		return value
	}
	return LanguagePTBR
}

// ResolveCompatibilityLanguage resolves the locale for compatibility diagnostics.
func ResolveCompatibilityLanguage(getenv func(string) string) Language {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := normalizeLanguage(getenv(key)); value != "" {
			return value
		}
	}
	return LanguageEN
}

func normalizeLanguage(value string) Language {
	normalized := strings.ToLower(strings.ReplaceAll(value, "_", "-"))
	switch {
	case strings.HasPrefix(normalized, "pt-br"), normalized == "pt":
		return LanguagePTBR
	case strings.HasPrefix(normalized, "en"):
		return LanguageEN
	default:
		return ""
	}
}
```

- [x] **Step 5: Implement typed catalogs**

Create `internal/i18n/catalog.go`:

```go
package i18n

// Message identifies one translated message.
type Message string

const (
	// MessageUnknownCommand reports an unsupported native command.
	MessageUnknownCommand Message = "unknown_command"
	// MessageMissingOperand reports an empty compatibility operand list.
	MessageMissingOperand Message = "missing_operand"
	// MessageIllegalOption reports an unsupported short option.
	MessageIllegalOption Message = "illegal_option"
	// MessageUnrecognizedOption reports an unsupported long option.
	MessageUnrecognizedOption Message = "unrecognized_option"
	// MessageInvalidInteractive reports an invalid --interactive value.
	MessageInvalidInteractive Message = "invalid_interactive"
	// MessageNativeUsage is the native Bearm usage heading.
	MessageNativeUsage Message = "native_usage"
)

// Catalog renders typed messages.
type Catalog struct {
	messages map[Message]string
}

// NewCatalog creates a complete catalog for language.
func NewCatalog(language Language) Catalog {
	if language == LanguagePTBR {
		return Catalog{messages: map[Message]string{
			MessageUnknownCommand:      "comando desconhecido",
			MessageMissingOperand:       "operando ausente",
			MessageIllegalOption:        "opção ilegal",
			MessageUnrecognizedOption:   "opção não reconhecida",
			MessageInvalidInteractive:   "valor inválido para --interactive",
			MessageNativeUsage:          "Uso: bearm <comando> [opções]",
		}}
	}

	return Catalog{messages: map[Message]string{
		MessageUnknownCommand:      "unknown command",
		MessageMissingOperand:       "missing operand",
		MessageIllegalOption:        "illegal option",
		MessageUnrecognizedOption:   "unrecognized option",
		MessageInvalidInteractive:   "invalid value for --interactive",
		MessageNativeUsage:          "Usage: bearm <command> [options]",
	}}
}

// Text returns the translated text for key.
func (c Catalog) Text(key Message) string {
	return c.messages[key]
}
```

- [x] **Step 6: Run tests**

Run:

```bash
gofmt -w internal/i18n
go test ./internal/i18n -v
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/i18n
git commit -m "feat: add Bearm message catalogs"
```

### Task 3: Implement Ordered GNU and BSD Compatibility Parsing

**Files:**
- Create: `internal/cli/errors.go`
- Create: `internal/cli/compatibility.go`
- Create: `internal/cli/compatibility_test.go`
- Modify: `internal/domain/removal.go`

**Interfaces:**
- Consumes:
  - `domain.CompatibilityProfile`
  - raw compatibility arguments.
- Produces:
  - `cli.UsageError`
  - `cli.ParseCompatibility(args, profile) (domain.RemoveRequest, error)`
  - `domain.RemoveOptions.ShowHelp`
  - `domain.RemoveOptions.ShowVersion`

- [x] **Step 1: Extend domain options for compatibility-only output**

Modify `internal/domain/removal.go` so `RemoveOptions` is:

```go
// RemoveOptions contains parsed compatibility-mode removal options.
type RemoveOptions struct {
	Force         bool
	Recursive     bool
	Directory     bool
	Verbose       bool
	Interactive   InteractiveMode
	PreserveRoot  PreserveRootMode
	OneFileSystem bool
	ShowHelp      bool
	ShowVersion   bool
}
```

- [x] **Step 2: Write ordered-parser tests**

Create `internal/cli/compatibility_test.go`:

```go
package cli_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func TestParseGNUCombinedOptions(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-rfv", "build"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Recursive || !got.Options.Force || !got.Options.Verbose {
		t.Fatalf("Options = %#v", got.Options)
	}
	if !reflect.DeepEqual(got.Operands, []string{"build"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseGNUOptionOrderForceThenInteractive(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-f", "-i", "file"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if got.Options.Force {
		t.Fatal("Force = true, want false after later -i")
	}
	if got.Options.Interactive != domain.InteractiveAlways {
		t.Fatalf("Interactive = %q", got.Options.Interactive)
	}
}

func TestParseGNUOptionOrderInteractiveThenForce(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-i", "-f", "file"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Force {
		t.Fatal("Force = false, want true after later -f")
	}
	if got.Options.Interactive != domain.InteractiveNever {
		t.Fatalf("Interactive = %q", got.Options.Interactive)
	}
}

func TestParseGNUAcceptsOptionsAfterOperand(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"build", "-rf"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Recursive || !got.Options.Force {
		t.Fatalf("Options = %#v", got.Options)
	}
	if !reflect.DeepEqual(got.Operands, []string{"build"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseBSDStopsOptionsAtFirstOperand(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"build", "-rf"}, domain.ProfileBSD)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if got.Options.Recursive || got.Options.Force {
		t.Fatalf("Options = %#v", got.Options)
	}
	want := []string{"build", "-rf"}
	if !reflect.DeepEqual(got.Operands, want) {
		t.Fatalf("Operands = %#v, want %#v", got.Operands, want)
	}
}

func TestParseCompatibilityHonorsDoubleDash(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"--", "-rf"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !reflect.DeepEqual(got.Operands, []string{"-rf"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseGNUInteractiveValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arg  string
		want domain.InteractiveMode
	}{
		{"--interactive=never", domain.InteractiveNever},
		{"--interactive=once", domain.InteractiveOnce},
		{"--interactive=always", domain.InteractiveAlways},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.arg, func(t *testing.T) {
			t.Parallel()

			got, err := cli.ParseCompatibility([]string{tt.arg, "file"}, domain.ProfileGNU)
			if err != nil {
				t.Fatalf("ParseCompatibility() error = %v", err)
			}
			if got.Options.Interactive != tt.want {
				t.Fatalf("Interactive = %q, want %q", got.Options.Interactive, tt.want)
			}
		})
	}
}

func TestParseCompatibilityRejectsUnsupportedOption(t *testing.T) {
	t.Parallel()

	_, err := cli.ParseCompatibility([]string{"-P", "file"}, domain.ProfileBSD)
	var usageErr *cli.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("error = %T %v, want *cli.UsageError", err, err)
	}
	if usageErr.Option != "-P" {
		t.Fatalf("Option = %q", usageErr.Option)
	}
}
```

- [x] **Step 3: Run tests and verify failure**

Run:

```bash
go test ./internal/cli -run TestParse -v
```

Expected:

```text
FAIL because ParseCompatibility and UsageError do not exist.
```

- [x] **Step 4: Implement typed usage errors**

Create `internal/cli/errors.go`:

```go
package cli

import "fmt"

// UsageError reports invalid command-line syntax without rendering text.
type UsageError struct {
	Kind   string
	Option string
	Value  string
}

func (e *UsageError) Error() string {
	switch e.Kind {
	case "invalid-interactive":
		return fmt.Sprintf("invalid interactive value %q", e.Value)
	case "missing-value":
		return fmt.Sprintf("missing value for %q", e.Option)
	default:
		return fmt.Sprintf("unsupported option %q", e.Option)
	}
}
```

- [x] **Step 5: Implement ordered compatibility parsing**

Create `internal/cli/compatibility.go`:

```go
package cli

import (
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
)

// ParseCompatibility parses GNU, BSD, or POSIX rm-compatible arguments.
func ParseCompatibility(args []string, profile domain.CompatibilityProfile) (domain.RemoveRequest, error) {
	state := compatibilityState{
		profile: profile,
		request: domain.RemoveRequest{
			Profile: profile,
			Options: domain.RemoveOptions{
				Interactive:  domain.InteractiveDefault,
				PreserveRoot: domain.PreserveRootDefault,
			},
		},
		parseOptions: true,
	}

	for _, arg := range args {
		if arg == "--" && state.parseOptions {
			state.parseOptions = false
			continue
		}

		if state.parseOptions && isOption(arg) {
			if err := state.applyOption(arg); err != nil {
				return domain.RemoveRequest{}, err
			}
			continue
		}

		state.request.Operands = append(state.request.Operands, arg)
		if profile == domain.ProfileBSD || profile == domain.ProfilePOSIX {
			state.parseOptions = false
		}
	}

	return state.request, nil
}

type compatibilityState struct {
	profile      domain.CompatibilityProfile
	request      domain.RemoveRequest
	parseOptions bool
}

func isOption(arg string) bool {
	return len(arg) > 1 && arg[0] == '-'
}

func (s *compatibilityState) applyOption(arg string) error {
	if strings.HasPrefix(arg, "--") {
		return s.applyLongOption(arg)
	}

	for _, option := range arg[1:] {
		if err := s.applyShortOption("-" + string(option)); err != nil {
			return err
		}
	}

	return nil
}

func (s *compatibilityState) applyShortOption(option string) error {
	switch option {
	case "-f":
		s.request.Options.Force = true
		s.request.Options.Interactive = domain.InteractiveNever
	case "-i":
		s.request.Options.Force = false
		s.request.Options.Interactive = domain.InteractiveAlways
	case "-I":
		s.request.Options.Force = false
		s.request.Options.Interactive = domain.InteractiveOnce
	case "-r", "-R":
		s.request.Options.Recursive = true
	case "-d":
		s.request.Options.Directory = true
	case "-v":
		s.request.Options.Verbose = true
	default:
		return &UsageError{Kind: "unsupported-option", Option: option}
	}

	return nil
}

func (s *compatibilityState) applyLongOption(option string) error {
	switch option {
	case "--force":
		return s.applyShortOption("-f")
	case "--interactive":
		return s.applyInteractive("always")
	case "--recursive", "--Recursive":
		s.request.Options.Recursive = true
	case "--dir", "--directory":
		s.request.Options.Directory = true
	case "--verbose":
		s.request.Options.Verbose = true
	case "--one-file-system":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.OneFileSystem = true
	case "--preserve-root":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootDefault
	case "--preserve-root=all":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootAll
	case "--no-preserve-root":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootNone
	case "--help":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.ShowHelp = true
	case "--version":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.ShowVersion = true
	default:
		if strings.HasPrefix(option, "--interactive=") {
			return s.applyInteractive(strings.TrimPrefix(option, "--interactive="))
		}
		return &UsageError{Kind: "unsupported-option", Option: option}
	}

	return nil
}

func (s *compatibilityState) applyInteractive(value string) error {
	s.request.Options.Force = false

	switch value {
	case "never":
		s.request.Options.Interactive = domain.InteractiveNever
	case "once":
		s.request.Options.Interactive = domain.InteractiveOnce
	case "always":
		s.request.Options.Interactive = domain.InteractiveAlways
	default:
		return &UsageError{Kind: "invalid-interactive", Option: "--interactive", Value: value}
	}

	return nil
}
```

- [x] **Step 6: Run parser tests**

Run:

```bash
gofmt -w internal/cli internal/domain
go test ./internal/cli -run TestParse -v
```

Expected:

```text
PASS
```

- [x] **Step 7: Add parser validation tests for no operands**

Append to `internal/cli/compatibility_test.go`:

```go
func TestParsedRequestValidationHandlesForceWithoutOperands(t *testing.T) {
	t.Parallel()

	request, err := cli.ParseCompatibility([]string{"-f"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestParsedRequestValidationRejectsMissingOperand(t *testing.T) {
	t.Parallel()

	request, err := cli.ParseCompatibility(nil, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}
	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
```

Run:

```bash
go test ./internal/cli ./internal/domain -v
```

Expected:

```text
PASS
```

- [x] **Step 8: Commit**

```bash
git add internal/cli internal/domain/removal.go
git commit -m "feat: parse ordered rm compatibility options"
```

### Task 4: Parse Native Bearm Commands

**Files:**
- Create: `internal/cli/native.go`
- Create: `internal/cli/native_test.go`

**Interfaces:**
- Consumes: native arguments after executable removal.
- Produces:
  - `cli.NativeCommand`
  - `cli.NativeRequest`
  - `cli.ParseNative(args []string) (NativeRequest, error)`

- [x] **Step 1: Write failing native-parser tests**

Create `internal/cli/native_test.go`:

```go
package cli_test

import (
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
)

func TestParseNativeList(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"list", "--limit", "25", "--json"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if got.Command != cli.CommandList || got.Limit != 25 || !got.JSON {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativeRestoreLast(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"restore", "--last"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if got.Command != cli.CommandRestore || !got.Last {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativePurgeItems(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"purge", "--yes", "item-1", "item-2"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if !got.Yes || !reflect.DeepEqual(got.ItemIDs, []string{"item-1", "item-2"}) {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativeRejectsConflictingRestoreSelectors(t *testing.T) {
	t.Parallel()

	_, err := cli.ParseNative([]string{"restore", "--last", "--operation", "op-1"})
	if err == nil {
		t.Fatal("ParseNative() error = nil, want non-nil")
	}
}
```

- [x] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/cli -run TestParseNative -v
```

Expected:

```text
FAIL because ParseNative does not exist.
```

- [x] **Step 3: Implement native parsing**

Create `internal/cli/native.go`:

```go
package cli

import (
	"errors"
	"strconv"
)

// NativeCommand identifies a Bearm-native command.
type NativeCommand string

const (
	CommandVersion NativeCommand = "version"
	CommandList    NativeCommand = "list"
	CommandRestore NativeCommand = "restore"
	CommandPurge   NativeCommand = "purge"
	CommandDoctor  NativeCommand = "doctor"
	CommandConfig  NativeCommand = "config"
)

// NativeRequest contains parsed native command arguments.
type NativeRequest struct {
	Command   NativeCommand
	ItemIDs   []string
	Operation string
	Last      bool
	Limit     int
	JSON      bool
	Yes       bool
	ConfigOp  string
}

// ParseNative parses Bearm-native command arguments.
func ParseNative(args []string) (NativeRequest, error) {
	if len(args) == 0 {
		return NativeRequest{}, errors.New("missing native command")
	}

	request := NativeRequest{Command: NativeCommand(args[0]), Limit: 50}
	switch request.Command {
	case CommandVersion:
		if len(args) != 1 {
			return NativeRequest{}, errors.New("version accepts no arguments")
		}
		return request, nil
	case CommandList:
		return parseNativeList(request, args[1:])
	case CommandRestore:
		return parseNativeSelection(request, args[1:], false)
	case CommandPurge:
		return parseNativeSelection(request, args[1:], true)
	case CommandDoctor:
		return parseNativeDoctor(request, args[1:])
	case CommandConfig:
		if len(args) != 2 || (args[1] != "path" && args[1] != "check") {
			return NativeRequest{}, errors.New("config requires path or check")
		}
		request.ConfigOp = args[1]
		return request, nil
	default:
		return NativeRequest{}, errors.New("unknown native command")
	}
}

func parseNativeList(request NativeRequest, args []string) (NativeRequest, error) {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--json":
			request.JSON = true
		case "--operation":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--operation requires a value")
			}
			request.Operation = args[index]
		case "--limit":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--limit requires a value")
			}
			value, err := strconv.Atoi(args[index])
			if err != nil || value < 1 {
				return NativeRequest{}, errors.New("--limit must be a positive integer")
			}
			request.Limit = value
		default:
			return NativeRequest{}, errors.New("unsupported list option")
		}
	}

	return request, nil
}

func parseNativeSelection(request NativeRequest, args []string, allowYes bool) (NativeRequest, error) {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--last":
			request.Last = true
		case "--operation":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--operation requires a value")
			}
			request.Operation = args[index]
		case "--yes":
			if !allowYes {
				return NativeRequest{}, errors.New("--yes is valid only for purge")
			}
			request.Yes = true
		default:
			if len(args[index]) > 0 && args[index][0] == '-' {
				return NativeRequest{}, errors.New("unsupported selection option")
			}
			request.ItemIDs = append(request.ItemIDs, args[index])
		}
	}

	selectors := 0
	if request.Last {
		selectors++
	}
	if request.Operation != "" {
		selectors++
	}
	if len(request.ItemIDs) > 0 {
		selectors++
	}
	if selectors != 1 {
		return NativeRequest{}, errors.New("select exactly one of item IDs, --operation, or --last")
	}

	return request, nil
}

func parseNativeDoctor(request NativeRequest, args []string) (NativeRequest, error) {
	if len(args) == 0 {
		return request, nil
	}
	if len(args) == 1 && args[0] == "--json" {
		request.JSON = true
		return request, nil
	}
	return NativeRequest{}, errors.New("doctor accepts only --json")
}
```

- [x] **Step 4: Run native-parser tests**

Run:

```bash
gofmt -w internal/cli
go test ./internal/cli -run TestParseNative -v
```

Expected:

```text
PASS
```

- [x] **Step 5: Commit**

```bash
git add internal/cli/native.go internal/cli/native_test.go
git commit -m "feat: parse Bearm native commands"
```

### Task 5: Integrate Parsers and Add Fuzz Coverage

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`
- Create: `internal/cli/fuzz_test.go`

**Interfaces:**
- Consumes:
  - `cli.ResolveInvocation`
  - `cli.ParseCompatibility`
  - `cli.ParseNative`
  - i18n catalogs.
- Produces:
  - parser-only application routing;
  - stable help/version handling;
  - parser fuzz targets.

- [x] **Step 1: Replace the application test with parser-routing coverage**

Add these tests to `internal/app/app_test.go`:

```go
func TestRunRMForceWithoutOperandsReturnsSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"rm", "-f"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
}

func TestRunRMMissingOperandReturnsGNUFailure(t *testing.T) {
	t.Parallel()

	t.Setenv("BEARM_COMPAT", "gnu")
	t.Setenv("LC_ALL", "C")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"rm"})
	if code != 1 {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "missing operand") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunNativeUnknownCommandUsesPortuguese(t *testing.T) {
	t.Parallel()

	t.Setenv("BEARM_LANG", "pt-BR")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"bearm", "unknown"})
	if code != 2 {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "comando desconhecido") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
```

- [x] **Step 2: Run tests and verify at least one failure**

Run:

```bash
go test ./internal/app -v
```

Expected:

```text
FAIL because App.Run does not route through the new parsers.
```

- [x] **Step 3: Integrate pure parsing into App.Run**

Replace `internal/app/app.go` with:

```go
// Package app coordinates Bearm use cases.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
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
	return &App{stdin: stdin, out: stdout, err: stderr, info: info}
}

// Run executes one Bearm invocation and returns a process exit code.
func (a *App) Run(_ context.Context, argv []string) int {
	invocation, err := cli.ResolveInvocation(argv)
	if err != nil {
		fmt.Fprintln(a.err, err)
		return 2
	}

	if invocation.Mode == cli.ModeCompatibility {
		return a.runCompatibility(invocation.Args)
	}

	return a.runNative(invocation.Args)
}

func (a *App) runCompatibility(args []string) int {
	profile := resolveProfile(os.Getenv, runtime.GOOS)
	request, err := cli.ParseCompatibility(args, profile)
	if err != nil {
		fmt.Fprintf(a.err, "rm: %v\n", err)
		return usageCode(profile)
	}

	if request.Options.ShowVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}
	if request.Options.ShowHelp {
		fmt.Fprintln(a.out, "Usage: rm [OPTION]... [FILE]...")
		return 0
	}

	if err := request.Validate(); err != nil {
		catalog := i18n.NewCatalog(i18n.ResolveCompatibilityLanguage(os.Getenv))
		fmt.Fprintf(a.err, "rm: %s\n", catalog.Text(i18n.MessageMissingOperand))
		return usageCode(profile)
	}

	return 0
}

func (a *App) runNative(args []string) int {
	catalog := i18n.NewCatalog(i18n.ResolveNativeLanguage(os.Getenv))
	request, err := cli.ParseNative(args)
	if err != nil {
		if strings.Contains(err.Error(), "unknown") || strings.Contains(err.Error(), "missing native") {
			fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
		} else {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
		}
		return 2
	}

	if request.Command == cli.CommandVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}

	fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
	return 2
}

func resolveProfile(getenv func(string) string, goos string) domain.CompatibilityProfile {
	switch strings.ToLower(getenv("BEARM_COMPAT")) {
	case "gnu":
		return domain.ProfileGNU
	case "bsd":
		return domain.ProfileBSD
	case "posix":
		return domain.ProfilePOSIX
	}

	if goos == "darwin" {
		return domain.ProfileBSD
	}
	return domain.ProfileGNU
}

func usageCode(profile domain.CompatibilityProfile) int {
	if profile == domain.ProfileBSD {
		return 64
	}
	return 1
}
```

- [x] **Step 4: Run application tests**

Run:

```bash
gofmt -w internal/app
go test ./internal/app -v
```

Expected:

```text
PASS
```

- [x] **Step 5: Add parser fuzz tests**

Create `internal/cli/fuzz_test.go`:

```go
package cli_test

import (
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func FuzzParseCompatibility(f *testing.F) {
	for _, seed := range []string{
		"-rf build",
		"-if file",
		"--interactive=once a b c d",
		"-- -rf",
		"file -rf",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		args := strings.Fields(input)
		for _, profile := range []domain.CompatibilityProfile{
			domain.ProfileGNU,
			domain.ProfileBSD,
			domain.ProfilePOSIX,
		} {
			request, err := cli.ParseCompatibility(args, profile)
			if err == nil && request.Profile != profile {
				t.Fatalf("Profile = %q, want %q", request.Profile, profile)
			}
		}
	})
}

func FuzzParseNative(f *testing.F) {
	for _, seed := range []string{
		"list --limit 10",
		"restore --last",
		"purge --yes item-1",
		"doctor --json",
		"config check",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = cli.ParseNative(strings.Fields(input))
	})
}
```

- [x] **Step 6: Run fuzz seeds and race tests**

Run:

```bash
go test ./internal/cli ./internal/app
go test -race ./internal/cli ./internal/app
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/app internal/cli/fuzz_test.go
git commit -m "feat: route Bearm command parsing"
```

## Plan Completion Verification

Run:

```bash
make verify
go run ./cmd/bearm version
BEARM_COMPAT=gnu go run ./cmd/bearm rm -f
BEARM_COMPAT=gnu go run ./cmd/bearm rm
```

Expected:

```text
The version command exits 0.
The forced empty compatibility command exits 0 silently.
The final command exits 1 and reports a missing operand on stderr.
```
