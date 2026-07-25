# Bearm Design Specification

**Status:** Approved for planning  
**Date:** 2026-07-23  
**Project:** Bearm  
**Module:** `github.com/Diaszano/bearm`  
**Primary binary:** `bearm`  
**License:** MIT

## 1. Executive Summary

Bearm is a native Go command-line tool that acts as a safe replacement for Unix `rm`. In compatibility mode, it accepts the expected GNU/Linux or BSD/macOS `rm` arguments, preserves option precedence, diagnostics, prompts, and exit-code behavior as closely as practical, but moves targets to an operating-system trash location instead of deleting them permanently.

Bearm also provides explicit management commands for listing, restoring, inspecting, and permanently purging items that Bearm moved to trash.

The project prioritizes:

1. Safety over irreversible behavior.
2. Observable compatibility with the host `rm`.
3. Constant-time same-filesystem directory moves whenever traversal is unnecessary.
4. Atomic metadata and collision handling.
5. A single native binary with no Node.js, npm, Git, GNU utilities, or shell-script runtime dependency.
6. Clear platform boundaries for Linux and macOS.
7. Test-driven implementation and differential testing against the host `rm`.

## 2. Product Identity

### 2.1 Name

- Product name: **Bearm**
- Executable: `bearm`
- Compatibility invocation:
  - `bearm rm [OPTIONS] TARGET...`
  - a symlink or wrapper named `rm` pointing to `bearm`
- Native management invocation:
  - `bearm list`
  - `bearm restore`
  - `bearm purge`
  - `bearm doctor`
  - `bearm config`

### 2.2 Tagline

> Remove safely. Restore confidently.

### 2.3 Language Rules

- Source code, identifiers, comments, technical documentation, branches, and commits are written in English.
- Conventional Commits are mandatory.
- Conventional branch names are mandatory.
- No AI attribution trailers are added to commits.
- Native Bearm commands default to Brazilian Portuguese (`pt-BR`) and may be switched to English.
- `rm` compatibility-mode diagnostics and prompts follow the selected compatibility profile and active locale catalog instead of forcing Bearm-native wording.

## 3. Problem Statement

The standard Unix `rm` command permanently unlinks files and recursively deletes directory trees. A single incorrect path, wildcard, variable expansion, or `-rf` invocation can cause irreversible data loss.

Existing shell-based safe-rm implementations prove the value of moving items to trash, but a large shell script has structural limits:

- argument parsing and platform behavior become tightly coupled;
- path quoting and unusual filenames are difficult to guarantee;
- each operation may spawn multiple external processes;
- concurrent trash writes require careful shell-level coordination;
- configuration loaded with `source` can execute arbitrary shell code;
- Linux and macOS behavior cannot be isolated cleanly;
- exact diagnostic and exit-code compatibility becomes difficult to test and maintain.

Bearm solves these problems with a modular native implementation.

## 4. Goals

### 4.1 Functional Goals

Bearm must:

1. Move files, symlinks, and directories to trash instead of permanently deleting them.
2. Support host-aware GNU/Linux and BSD/macOS `rm` compatibility profiles.
3. Support combined short options and the `--` option terminator.
4. Preserve correct option precedence, including order-sensitive `-f` and interactive options.
5. Refuse unsafe root, dot, dot-dot, scope, and protected-path operations.
6. Treat symlinks as targets themselves and never delete the symlink destination.
7. Use same-filesystem trash locations by default so ordinary moves use atomic rename operations.
8. Create collision-safe trash names without overwriting existing items or metadata.
9. Record Bearm operations in a durable journal.
10. Restore one item, a complete operation, or the most recent operation.
11. Permanently purge only through explicit Bearm-native commands.
12. Support configuration through declarative TOML without executing shell code.
13. Run on Linux and macOS as a single Go binary per target platform.
14. Provide complete automated unit, integration, fuzz, race, and compatibility tests.
15. Provide installation artifacts for GitHub Releases and Homebrew.

### 4.2 Non-Functional Goals

Bearm must:

- avoid directory traversal for ordinary non-interactive directory moves;
- perform same-filesystem moves through atomic rename where supported;
- use atomic exclusive creation for trash metadata and journal records;
- never overwrite an existing trash item;
- remain silent on successful compatibility-mode operations unless verbose output is requested;
- print diagnostics to stderr;
- keep normal startup memory below 32 MiB in release builds;
- add no telemetry or network access;
- remain usable without a configuration file;
- document every public Go API using standard Go documentation comments;
- pass `go test -race ./...`, `go vet ./...`, and formatting checks.

## 5. Non-Goals

Version 1 does not:

1. Reimplement secure multi-pass overwrite options such as BSD `rm -P`.
2. Reimplement undelete semantics such as BSD `rm -W`.
3. Guarantee byte-for-byte compatibility with every historical `rm` implementation.
4. Preserve Finder's native **Put Back** metadata on macOS. Bearm's own restore command is authoritative.
5. Copy arbitrary directory trees across filesystems into a custom centralized trash. Default per-filesystem trash routing avoids this condition.
6. Provide Windows support.
7. Run as a daemon.
8. synchronize trash or history across machines.
9. expose a graphical interface.
10. follow symlinks recursively.
11. permanently delete targets through ordinary compatibility-mode `rm` execution by default.

## 6. Compatibility Profiles

### 6.1 Automatic Selection

The default profile is selected from the host:

- Linux: `gnu`
- macOS: `bsd`

The profile can be explicitly selected with Bearm-native configuration or environment settings:

```bash
BEARM_COMPAT=gnu bearm rm ...
BEARM_COMPAT=bsd bearm rm ...
BEARM_COMPAT=posix bearm rm ...
```

When invoked through a binary name of `rm`, Bearm enters compatibility mode automatically.

### 6.2 Version 1 Options

#### Shared Options

| Option | Meaning |
|---|---|
| `-f`, `--force` | Ignore nonexistent targets and suppress interactive behavior according to profile precedence |
| `-i`, `--interactive=always` | Prompt before each target or visited entry |
| `-I`, `--interactive=once` | Prompt once for recursive removal or more than three operands |
| `-r`, `-R`, `--recursive` | Permit directory removal |
| `-d`, `--dir`, `--directory` | Permit removal of empty directories without recursive mode |
| `-v`, `--verbose` | Print each logical removal |
| `--` | End option parsing |

#### GNU Profile Options

| Option | Meaning |
|---|---|
| `--interactive=never` | Disable prompting |
| `--one-file-system` | Do not descend into mounted filesystems during traversal-required operations |
| `--preserve-root` | Protect `/`; enabled by default |
| `--preserve-root=all` | Protect any command-line directory operand located on a different device from its parent |
| `--no-preserve-root` | Disable compatibility root protection, while Bearm hard safety rules still protect `/` unless explicit native purge mode is used |
| `--help` | Print GNU-profile compatibility help |
| `--version` | Print Bearm version and compatibility information |

#### Explicitly Rejected Options

Unsupported options must fail with the profile-appropriate diagnostic and nonzero exit code. Bearm must never silently ignore a destructive or semantically incompatible option.

Examples include:

- BSD `-P`
- BSD `-W`
- BSD or implementation-specific `-x` where the selected profile does not define equivalent behavior

### 6.3 Option Precedence

Options are parsed in command-line order.

For GNU behavior:

- a later `-f` or `--interactive=never` disables prior interactive settings;
- a later `-i`, `-I`, or `--interactive=...` overrides prior force/interactivity behavior as GNU `rm` does;
- options may appear after operands unless `--` has terminated parsing.

For BSD behavior:

- option parsing stops at the first operand unless `--` is used;
- profile-specific option precedence is represented explicitly in parser state and verified against `/bin/rm`.

### 6.4 Operands

Bearm must correctly handle:

- relative and absolute paths;
- filenames beginning with `-`;
- whitespace, newlines, tabs, shell metacharacters, and non-ASCII characters;
- dangling symlinks;
- trailing slashes;
- duplicate operands;
- operands referring to the same inode through different path spellings;
- nonexistent operands;
- files without parent-directory write permission;
- read-only targets when parent permissions still permit rename.

## 7. Invocation Modes

### 7.1 Compatibility Mode

Compatibility mode is selected when:

- `argv[0]` basename is `rm`; or
- the first native argument is `rm`.

Compatibility mode:

- accepts only the documented compatibility options;
- performs no permanent deletion by default;
- uses profile-specific diagnostics, prompts, and exit codes;
- does not expose native subcommands;
- is silent on success unless `-v` is active.

### 7.2 Native Mode

Native mode is selected when the executable is invoked as `bearm` without the `rm` subcommand.

Version 1 native commands:

```text
bearm list [--operation ID] [--limit N] [--json]
bearm restore ITEM_ID...
bearm restore --operation ID
bearm restore --last
bearm purge ITEM_ID... [--yes]
bearm purge --operation ID [--yes]
bearm doctor [--json]
bearm config path
bearm config check
bearm version
```

Native commands use `pt-BR` by default and support `BEARM_LANG=en`.

## 8. Architecture

### 8.1 High-Level Flow

```text
argv
  │
  ▼
Invocation Resolver
  │
  ├── Compatibility Parser ──► Compatibility Request
  │
  └── Native Parser ─────────► Native Request
                               │
                               ▼
                         Validation Layer
                               │
                               ▼
                         Operation Planner
                               │
                  ┌────────────┴────────────┐
                  ▼                         ▼
             Prompt Engine             Safety Policy
                  └────────────┬────────────┘
                               ▼
                         Trash Executor
                               │
                  ┌────────────┴────────────┐
                  ▼                         ▼
             Platform Backend         Operation Journal
                               │
                               ▼
                          Result Renderer
```

### 8.2 Package Boundaries

```text
cmd/bearm
    Process entry point and dependency assembly only.

internal/app
    Use cases for compatibility removal and native commands.

internal/cli
    Invocation detection, GNU/BSD parsing, native parsing, and rendering.

internal/domain
    Immutable request, plan, target, operation, result, and error types.

internal/planner
    Filesystem inspection, traversal decisions, and operation planning.

internal/safety
    Root protection, scope enforcement, protected patterns, and symlink rules.

internal/trash
    Backend interface and shared collision/metadata behavior.

internal/trash/linux
    FreeDesktop home and per-mount trash implementation.

internal/trash/darwin
    Home and per-volume system-trash implementation.

internal/journal
    Durable JSONL operation journal and locking.

internal/restore
    Restore and purge use cases.

internal/config
    TOML loading, defaults, environment overrides, and validation.

internal/i18n
    Native command catalogs and compatibility diagnostic catalogs.

internal/platform
    Build-tagged filesystem and mount helpers.

internal/testutil
    Test fixtures, command runner, fake terminal, and filesystem helpers.
```

### 8.3 Dependency Direction

- `domain` imports only the Go standard library.
- `app` depends on interfaces from `domain`, `planner`, `trash`, `journal`, and `restore`.
- platform packages implement interfaces and must not be imported by domain packages.
- CLI packages translate arguments into domain requests and domain results into output.
- trash backends do not parse arguments or print output.
- filesystem traversal does not write to trash.
- journal code does not decide safety policy.
- native management commands do not bypass the same restore and purge policies used by tests.

## 9. Core Domain Model

```go
type CompatibilityProfile string

const (
    ProfileGNU   CompatibilityProfile = "gnu"
    ProfileBSD   CompatibilityProfile = "bsd"
    ProfilePOSIX CompatibilityProfile = "posix"
)

type InteractiveMode string

const (
    InteractiveDefault InteractiveMode = "default"
    InteractiveNever   InteractiveMode = "never"
    InteractiveOnce    InteractiveMode = "once"
    InteractiveAlways  InteractiveMode = "always"
)

type PreserveRootMode string

const (
    PreserveRootDefault PreserveRootMode = "default"
    PreserveRootNone    PreserveRootMode = "none"
    PreserveRootAll     PreserveRootMode = "all"
)

type RemoveOptions struct {
    Force         bool
    Recursive     bool
    Directory     bool
    Verbose       bool
    Interactive   InteractiveMode
    PreserveRoot  PreserveRootMode
    OneFileSystem bool
}

type RemoveRequest struct {
    Profile  CompatibilityProfile
    Operands []string
    Options  RemoveOptions
}

type TargetKind string

const (
    TargetFile    TargetKind = "file"
    TargetDir     TargetKind = "directory"
    TargetSymlink TargetKind = "symlink"
    TargetOther   TargetKind = "other"
)

type PlannedTarget struct {
    InputPath     string
    AbsolutePath  string
    Kind          TargetKind
    DeviceID      uint64
    RequiresWalk  bool
}

type RemovalPlan struct {
    ID      string
    Request RemoveRequest
    Targets []PlannedTarget
}

type TrashRecord struct {
    ItemID       string
    OperationID  string
    OriginalPath string
    TrashedPath  string
    Backend      string
    DeviceID     uint64
    DeletedAt    time.Time
}
```

## 10. Removal Planning

### 10.1 No-Traversal Fast Path

A directory is moved as a single target without traversal when all are true:

- recursive mode permits the directory;
- per-entry interaction is not active;
- verbose output does not require per-descendant messages for the selected profile;
- `--one-file-system` does not require mount-boundary filtering;
- no protected descendant rule requires inspection.

This path performs one metadata inspection and one rename operation.

### 10.2 Traversal Path

Traversal is required when:

- `InteractiveAlways` must prompt for entries;
- compatibility verbose behavior requires descendant ordering;
- `--one-file-system` must exclude mounted descendants;
- configured descendant protection is enabled.

Traversal must:

- use `Lstat`, never follow symlinks;
- preserve deterministic depth-first order matching the selected profile test fixtures;
- stop at mount boundaries when requested;
- detect changes between planning and execution;
- return partial-failure results without reporting unperformed targets as successful.

### 10.3 Duplicate Operands

The planner normalizes identity using absolute lexical paths plus `Lstat` file identity. A target already covered by an earlier recursive target is not moved twice. Compatibility diagnostics for duplicates are verified by profile tests.

## 11. Safety Model

### 11.1 Hard Safety Rules

The following cannot be removed through compatibility mode:

- `/`;
- `.` and `..`;
- lexical equivalents ending in `/.` or `/..`;
- the active trash root itself;
- the Bearm state directory;
- the Bearm configuration directory.

`--no-preserve-root` affects compatibility behavior but does not disable Bearm's hard protection of `/`.

### 11.2 Scope

Configuration may restrict compatibility removal to one or more roots:

```toml
[safety]
allowed_roots = ["/Users/dias", "/Volumes/Work"]
```

A target must be equal to or descend from an allowed root after lexical absolute normalization. Bearm does not resolve the final symlink target when checking a symlink operand.

### 11.3 Protected Patterns

Protected patterns are declarative and use gitignore-compatible syntax:

```toml
[safety]
protected_patterns = [
  "/Users/dias/Documents/important/**",
  "/Users/dias/Projects/**/.git",
]
inspect_descendants = false
```

With `inspect_descendants = false`, only explicit operands are checked, preserving the no-traversal fast path.

With `inspect_descendants = true`, recursive targets are traversed and protected descendants are retained. This mode is slower and is explicit.

### 11.4 Already-Trashed Targets

Default behavior:

- compatibility mode refuses to permanently delete items already inside a Bearm-managed trash;
- native `bearm purge` performs permanent deletion after confirmation;

## 12. Trash Backends

### 12.1 Shared Backend Interface

```go
type Backend interface {
    Name() string
    Resolve(ctx context.Context, target PlannedTarget) (Destination, error)
    Move(ctx context.Context, target PlannedTarget, destination Destination) (TrashRecord, error)
}

type Destination struct {
    Root       string
    FilesDir   string
    InfoDir    string
    TargetPath string
    InfoPath   string
}
```

### 12.2 Linux

Linux follows the FreeDesktop Trash specification.

Home trash:

```text
$XDG_DATA_HOME/Trash
```

Fallback:

```text
$HOME/.local/share/Trash
```

Per-mount routing is enabled by default:

1. use `<mount>/.Trash/<uid>` only when `.Trash` is a real directory with sticky bit;
2. otherwise use `<mount>/.Trash-<uid>`;
3. ensure `files/` and `info/` exist with safe permissions;
4. store a relative `Path` in `.trashinfo` for per-mount trash;
5. store an absolute `Path` for home trash.

Metadata reservation:

- reserve `<name>.trashinfo` through `O_CREATE|O_EXCL`;
- retry deterministic suffixes: `name`, `name.1`, `name.2`, and so on;
- write UTF-8 percent-encoded path and local deletion timestamp;
- `fsync` metadata before moving;
- remove reserved metadata if the move fails.

### 12.3 macOS

Bearm uses system-visible trash directories without relying on Finder scripting.

Home-volume trash:

```text
$HOME/.Trash
```

Other volumes:

```text
<mount>/.Trashes/<uid>
```

Bearm:

- resolves the mount root by walking parent directories until the device ID changes;
- creates or validates the user trash directory;
- reserves collision-safe names;
- moves targets with `rename`;
- writes Bearm metadata and journal records for reliable CLI restore;
- does not promise Finder **Put Back** metadata in version 1.

### 12.4 Custom Trash

A configured custom trash is supported only when the target can be atomically renamed into it.

If rename returns `EXDEV`, Bearm returns a clear error and does not copy or partially delete the source.

## 13. Atomicity and Failure Handling

For each target:

1. inspect and validate the target;
2. resolve a same-filesystem destination;
3. reserve metadata and destination identity;
4. rename the target;
5. append and sync the journal record;
6. finalize metadata.

Failure rules:

- if reservation fails, the source remains untouched;
- if rename fails, reserved metadata is removed;
- if journal append fails after rename, a recovery record is written beside the trash item and the command exits nonzero;
- startup and `doctor` reconcile recoverable orphan metadata;
- one target failure does not prevent independent later operands unless the failure invalidates the entire plan;
- exit code is nonzero when any requested operand fails.

Signals:

- before the first rename, interruption exits without changes;
- during execution, the active atomic rename completes, journal recovery is written, and no new target begins.

## 14. Journal and Recovery

### 14.1 Storage

Linux:

```text
$XDG_STATE_HOME/bearm/journal.jsonl
```

Fallback:

```text
$HOME/.local/state/bearm/journal.jsonl
```

macOS:

```text
$HOME/Library/Application Support/Bearm/state/journal.jsonl
```

Overrides:

```text
BEARM_STATE_HOME
BEARM_CONFIG_HOME
BEARM_DATA_HOME
```

### 14.2 Record Format

Each line is one JSON record:

```json
{
  "schema_version": 1,
  "item_id": "01J...",
  "operation_id": "01J...",
  "original_path": "/Users/dias/project/file.txt",
  "trashed_path": "/Users/dias/.Trash/file.txt",
  "backend": "darwin-system-trash",
  "device_id": 16777234,
  "deleted_at": "2026-07-23T12:30:00Z",
  "status": "trashed"
}
```

Journal writes use:

- an exclusive process lock;
- append-only writes;
- one complete JSON object per line;
- `fsync` after each completed operation;
- schema versioning;
- no secret or file-content capture.

### 14.3 Restore

Restore defaults to collision failure.

Supported collision policies:

```text
fail
rename
overwrite
```

`overwrite` requires `--yes` and permanently purges the conflicting destination through the native purge policy before restore.

A restored item appends a new journal event rather than mutating historical lines.

## 15. Configuration

### 15.1 Locations

Linux:

```text
$XDG_CONFIG_HOME/bearm/config.toml
```

Fallback:

```text
$HOME/.config/bearm/config.toml
```

macOS:

```text
$HOME/Library/Application Support/Bearm/config.toml
```

### 15.2 Example

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

### 15.3 Precedence

Highest to lowest:

1. explicit command-line Bearm-native flags;
2. environment variables;
3. TOML configuration;
4. platform defaults.

Compatibility-mode flags override only the corresponding removal semantics and do not disable hard safety invariants.

### 15.4 Validation

Invalid configuration causes:

- `bearm config check` to return nonzero with all detected errors;
- compatibility mode to print one concise error and perform no filesystem changes;
- no silent fallback for invalid paths, unknown enum values, or insecure directory ownership.

## 16. Internationalization

- Native catalogs: `pt-BR`, `en`.
- Compatibility catalogs: `gnu/en`, `gnu/pt-BR`, `bsd/en`, `bsd/pt-BR`.
- Catalog keys are typed constants.
- No package formats messages ad hoc outside `internal/i18n`.
- Tests use forced locale values for deterministic output.
- Native mode defaults to `pt-BR`.
- Compatibility mode uses `LC_ALL`, then `LC_MESSAGES`, then `LANG`, and falls back to English.

## 17. Exit Codes

Compatibility mode follows the selected profile where practical:

- `0`: all requested operations succeeded or were ignored under force semantics;
- `1`: GNU-profile operational or usage failure;
- `64`: BSD-profile usage failure where BSD `rm` uses `EX_USAGE`;
- other platform-specific errors are normalized only when compatibility tests prove the host behavior.

Native mode:

- `0`: success;
- `1`: operational failure;
- `2`: invalid native command usage;
- `3`: invalid configuration;
- `4`: partial success;
- `130`: interrupted by SIGINT.

## 18. Security Threat Model

### 18.1 Threats

Bearm must defend against:

- shell injection through paths or configuration;
- symlink target deletion;
- destination collision overwrite;
- malicious trash-directory symlinks;
- world-writable trash directories without sticky-bit protections;
- configuration files owned by another user;
- time-of-check/time-of-use target replacement;
- concurrent Bearm processes selecting the same destination;
- journal truncation or partial lines;
- path traversal in `.trashinfo`;
- accidental permanent purge;
- environment variables pointing to unsafe relative directories.

### 18.2 Controls

- no shell execution;
- subprocesses are avoided in version 1;
- all filesystem operations receive direct argument values;
- `Lstat` is used for operands;
- trash skeleton directories are validated as real directories;
- `O_EXCL` reservations prevent concurrent name reuse;
- custom directories must be absolute;
- protected path checks occur during planning and immediately before execution;
- purge requires native command context and explicit confirmation;
- journal parsing ignores incomplete trailing lines but reports them through `doctor`;
- logs redact no file paths by default because paths are required diagnostics, but never include file contents.

## 19. Performance Requirements

1. Moving a non-interactive directory on the same filesystem performs no recursive walk.
2. Name collision selection must not scan the entire trash directory for every ordinary move; it attempts the base name and increments only on collision.
3. CLI parsing is allocation-conscious and does not invoke subprocesses.
4. Journal writes are batched per operation and synced once after all item records for that operation.
5. Benchmarks cover:
   - parser throughput;
   - no-traversal planning;
   - collision reservation;
   - journal append;
   - 10,000-operand command parsing.
6. Release builds use `-trimpath`.
7. Startup and simple forced missing-target behavior must complete without initializing backends or journal storage.

## 20. Observability

- No telemetry.
- Default log level: `error`.
- Debug mode:
  - `BEARM_LOG=debug`;
  - native `--debug`.
- Debug output is written to stderr.
- Optional JSON debug logs are available only in native mode.
- Every failure includes:
  - operation or item ID when available;
  - path;
  - phase: parse, validate, plan, reserve, move, journal, restore, or purge;
  - wrapped operating-system error.

## 21. Repository Structure

```text
bearm/
├── cmd/
│   └── bearm/
│       └── main.go
├── internal/
│   ├── app/
│   ├── cli/
│   ├── config/
│   ├── domain/
│   ├── i18n/
│   ├── journal/
│   ├── planner/
│   ├── platform/
│   ├── restore/
│   ├── safety/
│   ├── testutil/
│   └── trash/
│       ├── darwin/
│       └── linux/
├── docs/
│   ├── architecture/
│   ├── compatibility/
│   └── superpowers/
│       ├── plans/
│       └── specs/
├── test/
│   ├── compatibility/
│   ├── integration/
│   └── fixtures/
├── .github/
│   └── workflows/
├── .golangci.yml
├── .goreleaser.yaml
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## 22. Testing Strategy

### 22.1 Unit Tests

Unit tests cover:

- ordered argument parsing;
- GNU and BSD option-order behavior;
- path normalization;
- hard safety rules;
- scope matching;
- protected patterns;
- traversal decisions;
- collision naming;
- percent encoding;
- configuration precedence;
- journal parsing and recovery;
- restore collision policies;
- i18n lookup.

### 22.2 Integration Tests

Integration tests use temporary filesystems and subprocess execution to verify:

- files, directories, and symlinks are moved;
- directory fast path does not walk descendants;
- metadata is created correctly;
- failed moves leave sources intact;
- duplicate names never overwrite;
- concurrent processes receive unique names;
- restore recreates the original path;
- purge is impossible without confirmation;
- invalid configuration causes no changes.

### 22.3 Differential Compatibility Tests

The same fixture is executed with:

```text
/bin/rm
test-built bearm
```

The harness compares:

- exit code;
- stdout;
- stderr;
- prompts;
- logical files selected;
- traversal order where applicable.

Because Bearm moves rather than unlinks, filesystem postconditions are normalized into logical states:

- removed from original location;
- untouched;
- failed;
- prompted and declined.

### 22.4 Fuzzing

Fuzz targets:

- compatibility argument parser;
- path percent encoding and decoding;
- journal JSONL parser;
- protected-pattern matcher;
- config decoder.

### 22.5 CI Matrix

Required:

- Ubuntu;
- macOS;
- supported Go version floor;
- current stable Go;
- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- formatting check;
- differential profile tests;
- release snapshot build.

## 23. Delivery Phases

1. Foundation and domain contracts.
2. Compatibility parser and diagnostics.
3. Linux FreeDesktop trash backend.
4. macOS system-visible trash backend.
5. Safety planner, executor, and journal.
6. Native list, restore, purge, doctor, and configuration commands.
7. Differential compatibility suite, hardening, packaging, and documentation.

Each phase has a separate executable Superpowers implementation plan and one Conventional Commit per task.

## 24. Acceptance Criteria

Bearm version 1 is accepted when:

1. `bearm rm -rf directory` moves the directory to same-filesystem trash without traversing descendants when no traversal feature is active.
2. `rm`-named invocation automatically selects compatibility mode.
3. Linux and macOS choose the documented trash location.
4. duplicate trash names never overwrite existing files or metadata.
5. dangling symlinks are moved as symlinks and their targets remain untouched.
6. `/`, `.`, `..`, trash roots, state roots, and config roots are protected.
7. GNU and BSD argument-order fixtures pass against their host `rm`.
8. invalid options are rejected instead of ignored.
9. force semantics for nonexistent operands match the selected profile.
10. interactive prompts and exit codes pass compatibility fixtures.
11. every successful move has a recoverable journal record.
12. `bearm restore --last` restores the latest complete operation.
13. `bearm purge` requires confirmation or `--yes`.
14. invalid configuration performs no filesystem mutation.
15. concurrent collision tests pass under `go test -race`.
16. `go test -race ./...`, `go vet ./...`, formatting checks, and release snapshot builds pass on Linux and macOS.
17. the repository documents installation, aliasing, recovery, limitations, and uninstall procedures.

## 25. Risks and Mitigations

### Risk: Exact `rm` diagnostics vary by operating system and locale

Mitigation: maintain host-specific golden fixtures captured in CI and keep compatibility catalogs separate from native text.

### Risk: Planning and execution races

Mitigation: revalidate immediately before rename, use exclusive metadata reservation, and report changed targets rather than acting on stale assumptions.

### Risk: Custom trash is on another filesystem

Mitigation: reject `EXDEV` without modifying the source; document per-mount trash as the default.

### Risk: Finder Put Back is unavailable

Mitigation: Bearm journal and restore commands are authoritative; document the limitation.

### Risk: Protected-descendant scanning harms performance

Mitigation: disable descendant inspection by default and make the performance trade-off explicit.

### Risk: Scope grows into a complete file manager

Mitigation: version 1 includes only removal compatibility, list, restore, purge, doctor, configuration, and release tooling.

## 26. Conventional Development Rules

### Branches

Examples:

```text
feat/domain-foundation
feat/gnu-cli-parser
feat/linux-trash-backend
feat/darwin-trash-backend
feat/operation-journal
feat/restore-command
test/differential-rm-suite
chore/release-pipeline
```

### Commits

Examples:

```text
chore: initialize Bearm Go module
feat: add ordered GNU rm option parsing
feat: add FreeDesktop trash backend
feat: add macOS per-volume trash routing
feat: add atomic operation journal
feat: restore the latest trash operation
test: add GNU rm differential fixtures
docs: document installation and recovery
```

One task produces one focused Conventional Commit. Commits must not include AI attribution trailers.
