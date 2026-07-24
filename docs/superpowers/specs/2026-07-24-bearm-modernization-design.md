# Bearm Modernization Design

## Status

Approved on 2026-07-24. This document defines the modernization work before
implementation begins.

## Goal

Move Bearm to a supported Go 1.26 baseline, refresh its maintained Go
dependencies and developer tooling, replace the unmaintained gitignore matcher,
and make formatting and CI tool versions reproducible without weakening safety
or platform coverage.

## Scope

The work covers:

- Go source-build support and GitHub Actions test matrix;
- direct Go module dependencies;
- the protected-path gitignore matcher;
- `golangci-lint` v2 and all project formatting;
- reproducible versions for linting, vulnerability scanning, release tooling,
  and GitHub Actions dependencies;
- Dependabot configuration for Go modules and GitHub Actions;
- documentation and contributor commands that name the affected versions.

It does not change Bearm's command-line interface, trash format, journal
schema, safety policy, supported operating systems, or release publication
process.

## Compatibility Baseline

Bearm's minimum supported Go version becomes Go 1.26. The repository will use
`go 1.26.0` in `go.mod` and declare the current approved patch toolchain
(`go1.26.5`) with the `toolchain` directive. Project documentation must state
that Go 1.26.5 or newer is required to build from source.

The test matrix retains two distinct goals:

- an exact `1.26.5` job verifies the documented minimum patched toolchain;
- a `stable` job detects compatibility regressions in the current Go release.

When a future Go release becomes stable, it remains in the `stable` job until a
separate approved compatibility-floor decision changes the documented minimum.

## Dependency Policy

Direct runtime modules after the update are:

| Module | Approved version | Decision |
| --- | --- | --- |
| `github.com/pelletier/go-toml/v2` | `v2.4.3` | Retain; it is current. |
| `golang.org/x/sys` | `v0.47.0` | Upgrade from `v0.35.0`. |
| `github.com/idelchi/go-gitignore` | `v0.0.3` | Adopt in place of the unmaintained matcher. |

`github.com/sabhiram/go-gitignore` must be absent from `go.mod`, `go.sum`, and
production source once the migration is complete. No `replace` directive,
vendored source, or compatibility wrapper that exposes either third-party
matcher type is permitted.

## Protected-Path Matcher Migration

`internal/safety.PatternMatcher` remains Bearm's only public-internal boundary
for gitignore-style pattern evaluation. The replacement library is used only in
`internal/safety/patterns.go`; callers continue to use
`CompilePatterns([]string) (*PatternMatcher, error)` and
`(*PatternMatcher).Matches(string) bool`.

The adapter continues to accept only absolute positive patterns and absolute
negations. It converts those patterns to root-relative slash-separated input
for the matcher, and it rejects a relative input path at `Matches` time.

Safety behavior is conservative. The migration must preserve the existing
meaning of positive matching and negation for every supported pattern. If a
candidate library behavior is ambiguous, differs from the established tests,
or cannot model a safety policy, the affected path is treated as protected and
the dependency update is not merged until an explicit test and policy decision
resolve it.

The test suite must cover anchored patterns, `**`, directory-only patterns,
negation, comments, escaped leading `!` and `#`, escaped trailing spaces,
literal metacharacters, root paths, symlinks, and relative-path rejection.
Where Git semantics apply, a table-driven differential test compares matcher
results with `git check-ignore` in an isolated temporary repository. Bearm
must never run Git or any subprocess at runtime; the subprocess is test-only.

## Tooling and Formatting

`golangci-lint` is upgraded from v1 to v2 and becomes the only formatting
engine invoked by project commands and CI. The v2 configuration uses the
dedicated `formatters` section and enables exactly these formatters:

- `gci`;
- `gofumpt` with `module-path: github.com/Diaszano/bearm` and
  `extra-rules: true`;
- `golines` with `max-len: 120`.

No other formatter is enabled. The `gci` configuration separates standard
library, third-party, and `github.com/Diaszano/bearm` imports. Format checking
uses `golangci-lint fmt --diff`; explicit formatting uses
`golangci-lint fmt`. `gofmt` is not run as a standalone project formatter.

The repository pins the lint and vulnerability scanner versions using Go's
tool dependency support and invokes them with `go tool`. `Makefile`, local
acceptance checks, CI, and release verification use those same commands. The
selected `golangci-lint` v2 release and `govulncheck` release are fixed after
their migration and validation pass; no project command installs either with
`@latest`.

The v1 configuration is migrated with `golangci-lint migrate`, then reduced to
the existing intended analysis set plus valid v2 names and settings. A format
idempotence test runs `golangci-lint fmt`, then requires
`golangci-lint fmt --diff` to emit no diff.

## CI, Release, and Dependency Automation

Every job uses the Go version appropriate to its purpose: the compatibility
matrix uses `1.26.5` and `stable`; release jobs use `1.26.5` so release output
is reproducible; security scanning uses the pinned Go tool dependency.

GitHub Actions remain versioned references. The implementation replaces
unbounded tool references such as GoReleaser's `latest` and `~> v2` selection
with exact, validated releases. A release-tool upgrade is accepted only after
the snapshot job and release configuration validation pass. Existing action
major references are reviewed against their upstream releases, not changed
merely for churn.

`.github/dependabot.yml` schedules weekly updates for `gomod` and
`github-actions` at the repository root. Dependabot changes run the complete
quality workflow and remain subject to review; automatic merging is out of
scope.

## Validation and Acceptance Criteria

The modernization change is accepted only when all of the following are true:

1. `go.mod` requires Go 1.26.0 and the `go1.26.5` toolchain; docs state Go
   1.26.5 as the source-build floor.
2. `go mod tidy` and `go mod verify` succeed, and the deprecated matcher is
   absent from runtime source and module manifests.
3. The complete protected-pattern table and the isolated Git differential
   tests pass on Linux and macOS.
4. `go tool golangci-lint config verify`, `go tool golangci-lint run ./...`,
   and `go tool golangci-lint fmt --diff` succeed. The active formatter list
   contains exactly `gci`, `gofumpt`, and `golines`.
5. `go tool govulncheck ./...`, `go vet ./...`, `go test ./...`, and
   `go test -race ./...` succeed.
6. Linux and macOS CI succeed for `1.26.5` and `stable`; platform backend
   race tests and fuzz smoke checks remain enabled.
7. A clean snapshot release validates `.goreleaser.yaml`, builds all existing
   platform artifacts, and produces the expected SBOMs and checksums.
8. Dependabot configuration is syntactically valid and covers both `gomod` and
   `github-actions`.

## Rollback

The implementation is a single focused pull request with no release tag. A
failed matcher differential, safety test, release snapshot, or platform CI job
blocks merge. Revert the complete modernization commit series to restore the
previous Go baseline, matcher, configuration, and workflows. User journal and
trash data require no migration and remain untouched.
