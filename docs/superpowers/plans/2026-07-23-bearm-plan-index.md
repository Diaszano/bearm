# Bearm Implementation Plan Index

**Date:** 2026-07-23  
**Specification:** [`../specs/2026-07-23-bearm-design.md`](../specs/2026-07-23-bearm-design.md)

## Goal

Implement Bearm as a native Go replacement for Unix `rm` that preserves
observable GNU/BSD behavior where practical while moving targets to trash and
providing reliable recovery.

## Required Execution Skill

Use one of:

1. `superpowers:subagent-driven-development` — recommended, with a fresh
   implementation subagent and review gate for each task.
2. `superpowers:executing-plans` — execute plan tasks in dependency order with
   explicit checkpoints.

Create an isolated worktree with `superpowers:using-git-worktrees` before
implementation begins.

## Global Delivery Rules

- Execute plans in dependency order.
- Use TDD for every behavior task.
- Complete every verification command before committing.
- Produce one focused Conventional Commit per task.
- Keep code, identifiers, comments, technical documentation, branches, and
  commits in English.
- Do not add AI attribution trailers.
- Do not alias the user's real `rm` until the full compatibility and acceptance
  suites pass on that operating system.
- Preserve unrelated working-tree changes.
- Never weaken Bearm hard safety invariants to satisfy a compatibility test.

## Dependency DAG

```text
01 Foundation
    │
    ▼
02 CLI Compatibility
    │
    ├───────────────┐
    ▼               ▼
03 Linux Trash   04 macOS Trash
    └───────┬───────┘
            ▼
05 Removal Engine
            ▼
06 Journal and Recovery
            ▼
07 Configuration and Protection
            ▼
08 Compatibility, Hardening, and Release
```

## Plans

### 01. Foundation

[`2026-07-23-bearm-foundation.md`](2026-07-23-bearm-foundation.md)

Creates the Go module, repository rules, domain contracts, platform directory
resolution, build metadata, executable shell, CI baseline, and release
snapshot configuration.

**Exit gate:**

```bash
make verify
make build
./bin/bearm version
```

### 02. CLI Compatibility

[`2026-07-23-bearm-cli-compatibility.md`](2026-07-23-bearm-cli-compatibility.md)

Implements invocation detection, locale catalogs, ordered GNU/BSD option
parsing, native command parsing, application routing, and parser fuzz seeds.

**Exit gate:**

```bash
make verify
BEARM_COMPAT=gnu ./bin/bearm rm -f
```

### 03. Linux Trash Backend

[`2026-07-23-bearm-linux-trash.md`](2026-07-23-bearm-linux-trash.md)

Implements FreeDesktop metadata, device and mount resolution, per-mount trash,
atomic name reservation, rename transactions, and Linux concurrency tests.

**Exit gate:**

```bash
go test -race ./internal/trash/linux -count=10
```

### 04. macOS Trash Backend

[`2026-07-23-bearm-darwin-trash.md`](2026-07-23-bearm-darwin-trash.md)

Extracts shared collision reservation and implements home/per-volume
system-visible Trash routing, Bearm metadata, symlink behavior, and macOS
concurrency tests.

**Exit gate:**

```bash
go test -race ./internal/trash/darwin -count=10
```

### 05. Removal Engine

[`2026-07-23-bearm-removal-engine.md`](2026-07-23-bearm-removal-engine.md)

Implements IDs, result types, hard safety policy, filesystem planning,
`--preserve-root=all`, traversal decisions, prompts, execution, and real
compatibility-mode wiring.

**Exit gate:**

```bash
make verify
```

### 06. Journal and Recovery

[`2026-07-23-bearm-journal-recovery.md`](2026-07-23-bearm-journal-recovery.md)

Implements locked append-only JSONL history, lifecycle events, list, restore,
purge, doctor, and native command integration.

**Exit gate:**

```bash
make verify
go test -race ./internal/journal ./internal/restore
```

### 07. Configuration and Protection

[`2026-07-23-bearm-configuration-protection.md`](2026-07-23-bearm-configuration-protection.md)

Implements secure TOML loading, environment precedence, protected patterns,
protected-descendant traversal, custom trash, and config commands.

**Exit gate:**

```bash
make verify
go run ./cmd/bearm config check
```

### 08. Compatibility, Hardening, and Release

[`2026-07-23-bearm-compatibility-release.md`](2026-07-23-bearm-compatibility-release.md)

Adds profile-specific diagnostics, differential `/bin/rm` testing, edge and
security fixtures, fuzzing, benchmarks, complete documentation, Homebrew tap
publishing, SBOMs, signing, and final acceptance automation.

**Exit gate:**

```bash
./scripts/acceptance.sh
```

## Release Gate

Version `v1.0.0` must not be tagged until all acceptance criteria in the design
specification pass on both Ubuntu and macOS CI.

The release candidate must demonstrate:

```bash
bearm rm FILE
bearm list
bearm restore --last
bearm purge ITEM_ID --yes
bearm doctor
bearm config check
```

and a disposable shell alias:

```bash
alias rm='bearm rm'
```

The project must never install over `/bin/rm` automatically.
