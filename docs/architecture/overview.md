# Architecture Overview

Bearm is a single Go binary with a thin composition root in `cmd/bearm`.
Packages under `internal` separate parsing, policy, filesystem planning,
platform trash operations, history, and recovery.

## Dependency Direction

- `internal/domain` contains infrastructure-independent request, plan, result,
  trash, and journal types.
- `internal/cli` translates arguments into domain requests and renders typed
  results and compatibility diagnostics.
- `internal/config`, `internal/i18n`, and `internal/platform` provide typed
  configuration, catalogs, and operating-system paths.
- `internal/safety` and `internal/planner` inspect requests and filesystem
  state without writing to trash.
- `internal/removal` executes a completed plan through domain interfaces.
- `internal/trash/linux`, `internal/trash/darwin`, and
  `internal/trash/custom` implement platform destinations.
- `internal/journal` persists lifecycle events; it does not make safety
  decisions.
- `internal/restore` implements list-adjacent recovery, purge, and doctor use
  cases without parsing command lines.
- `internal/app` assembles these parts and coordinates use cases.

Domain and policy packages do not import platform backends. Trash packages do
not parse arguments or write user-facing output.

## Removal Flow

```text
argv
  |
  v
resolve invocation -> parse -> validate -> plan -> prompt -> execute -> journal
                          |          |         |          |
                          |          |         |          +-- render result
                          |          |         +-- user may decline
                          |          +-- safety and filesystem inspection
                          +-- typed GNU, BSD, POSIX, or native request
```

Parsing is profile-aware and option order is retained in parser state.
Planning uses `Lstat`, applies hard and configured policy, removes duplicate
coverage, and creates immutable planned targets. The executor rechecks policy
immediately before each move.

## No-Traversal Fast Path

An ordinary recursive directory removal can move the directory with one
`rename` when per-entry prompting, protected-descendant inspection, and
mount-boundary filtering are inactive. The planner records the directory as a
single target and does not visit descendants. This keeps large same-filesystem
directory operations effectively constant-time with respect to descendant
count.

Traversal is explicit when interaction or policy requires entry-level
decisions. Walks are deterministic, use `Lstat`, and do not follow symlinks.

## Platform Backend Boundaries

Linux routes targets to the FreeDesktop home trash or a safe per-mount trash.
It creates `files/` and `info/` and writes `.trashinfo` metadata.

macOS routes home-volume targets to `$HOME/.Trash` and other-volume targets to
`<mount>/.Trashes/<uid>`. Bearm metadata lives in a private `.bearm-info`
directory inside that trash root.

Custom trash configuration uses the shared reservation mechanism but accepts
only atomic `rename`; an `EXDEV` failure never falls back to recursive copy.

## Atomic Reservation and Failure Boundaries

The shared trash package reserves metadata with exclusive creation, trying the
base name and then numeric suffixes only on collision. A failed reservation
leaves the source untouched. A failed rename rolls back the reservation. A
successful rename commits the reservation and produces a `TrashRecord`.

The journal serializes append operations with an advisory lock and writes one
complete JSON event per line. An incomplete trailing line is ignored when
reading active state and reported by `bearm doctor`.

## Restore and Purge Lifecycle

History is append-only:

```text
trashed event -> active item -> restored event
                         \----> purged event
```

Restore renames the trash item back to its recorded path and appends a
`restored` event. Purge permanently removes only explicitly selected active
items after confirmation and appends a `purged` event. Historical lines are
never rewritten to represent a state transition.
