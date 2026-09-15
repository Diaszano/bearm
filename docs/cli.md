# CLI Command & Flag Reference

Bearm operates in two distinct execution modes:

1. **Compatibility Mode**: Activated via `bearm rm ...` (or when invoked through an executable/alias named `rm`). Mimics standard Unix `rm(1)` flags and behaviors while safely relocating files to the trash.
2. **Native Mode**: Activated via subcommands like `bearm list`, `bearm restore`, `bearm purge`, `bearm doctor`, and `bearm config`.

---

## Compatibility Mode (`rm`)

```bash
bearm rm [OPTIONS]... [FILE]...
```

### Supported Options by Profile

| Option | Long Flag | Profiles | Description |
| :--- | :--- | :---: | :--- |
| `-f` | `--force` | GNU, BSD, POSIX | Ignore nonexistent files and never prompt for confirmation. |
| `-i` | `--interactive=always` | GNU, BSD, POSIX | Prompt before removing every target item. |
| `-I` | `--interactive=once` | GNU, BSD | Prompt once when removing more than three files or recursively. |
| `-r`, `-R` | `--recursive` | GNU, BSD, POSIX | Remove directories and their contents recursively. |
| `-d` | `--dir`, `--directory` | GNU, BSD | Remove empty directories. |
| `-v` | `--verbose` | GNU, BSD | Explain what is being moved to trash. |
| | `--one-file-system` | GNU | Do not traverse into directories on a different filesystem mount. |
| | `--preserve-root` | GNU | Do not remove `/` (default and enforced by Bearm unconditionally). |
| | `--no-preserve-root` | GNU | Disables GNU-level root check (Bearm hard safety still rejects `/`). |
| | `--help` | GNU | Display compatibility help text. |
| | `--version` | GNU | Display Bearm version and build metadata. |
| `--` | | GNU, BSD, POSIX | Treat all subsequent arguments as file operands. |

### Profile Differences & Option Ordering

- **GNU Profile** (Default on Linux, or `BEARM_COMPAT=gnu`):
  - Supports flags anywhere in the argument list until `--` is encountered.
  - Options are applied strictly left-to-right. A subsequent `-f` overrides earlier `-i`.
- **BSD Profile** (Default on macOS, or `BEARM_COMPAT=bsd`):
  - Stops option parsing at the first non-option operand (POSIX-compliant).
  - Subsequent tokens starting with `-` are treated as file names unless `--` preceded them.
- **POSIX Profile** (`BEARM_COMPAT=posix`):
  - Follows strict POSIX.1 utility conventions.

### Safety Guarantees During `rm`

- **Protected Targets**: Rejects `/`, `.`, `..`, lexical root paths, active trash folders, and Bearm configuration/state files.
- **Symlink Protection**: Moving a symlink moves only the symlink itself; the link target is never dereferenced or deleted.
- **Atomic Directory Moves**: When recursive flags are used and no interactive prompts are active, Bearm moves the entire directory in a single O(1) atomic filesystem `rename`.

---

## Native Subcommands

### 1. `bearm list`

Lists active items currently residing in the trash that were removed by Bearm.

```bash
bearm list [--limit <N>] [--operation <OP_ID>] [--json]
```

- `--limit <N>`: Maximum number of records to return (default: `50`).
- `--operation <OP_ID>`: Filter results to items from a specific batch removal operation.
- `--json`: Output machine-readable JSON array of records.

#### Example Output:
```text
itm_01j7abc123   2026-09-14T21:30:00-03:00   /home/alice/project/old_module.go
itm_01j7abc124   2026-09-14T21:30:00-03:00   /home/alice/project/temp.log
```

---

### 2. `bearm restore`

Restores files or directories from trash back to their original locations.

```bash
# Restore specific item(s) by Item ID
bearm restore ITEM_ID [ITEM_ID...]

# Restore all items belonging to a batch operation ID
bearm restore --operation OPERATION_ID

# Undo the most recent removal operation
bearm restore --last
```

#### Collision Handling:
- If the original destination path already exists on disk:
  - Under the default policy (`fail`), Bearm stops and preserves both files.
  - If configured to `rename` (`restore.collision_policy = "rename"`), the file is restored with a `.restored.N` suffix.
  - Overwrite policy is rejected unless confirmed explicitly.

---

### 3. `bearm purge`

Permanently deletes items from the trash. This is the **only** Bearm command that unlinks files permanently.

```bash
# Purge specific item(s) (prompts for confirmation unless --yes is passed)
bearm purge ITEM_ID [ITEM_ID...] [--yes]

# Purge all items from an operation
bearm purge --operation OPERATION_ID [--yes]

# Purge the last trashed operation
bearm purge --last [--yes]
```

- `--yes`: Bypass interactive confirmation prompt.

> [!WARNING]
> Purging is permanent and irreversible. Once purged, files cannot be recovered with Bearm.

---

### 4. `bearm doctor`

Audits and verifies the health and consistency of Bearm's append-only journal and trash backends.

```bash
bearm doctor [--json]
```

Checks performed:
- Detects orphaned trash items missing journal records.
- Detects journal entries pointing to missing files in trash.
- Detects and flags incomplete trailing journal lines (e.g. from power loss).

---

### 5. `bearm config`

Inspects and validates the current Bearm configuration.

```bash
# Show active configuration file path
bearm config path

# Validate configuration file syntax and safety boundaries
bearm config check
```

---

### 6. `bearm version`

Prints the binary version, commit hash, build timestamp, Go runtime, and platform architecture.

```bash
bearm version
```

---

## Exit Codes

| Code | Profile / Context | Meaning |
| :---: | :---: | :--- |
| `0` | All | Success (all items moved or restored, or missing files ignored under `-f`). |
| `1` | All | Operational failure (item not found, permission denied, collision, disk error). |
| `2` | Native / App | CLI argument parsing or unknown subcommand error. |
| `3` | App / Config | Fatal configuration, directory resolution, or protected pattern validation error. |
| `64` | BSD Profile | Usage error (`EX_USAGE` for unsupported or invalid flags). |

---

## Environment Variables

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `BEARM_COMPAT` | `auto` | Force compatibility profile: `auto`, `gnu`, `bsd`, `posix`. |
| `BEARM_LANG` | `pt-BR` | Language for native command messages: `pt-BR` or `en`. |
| `BEARM_CONFIG_HOME` | Linux: `~/.config/bearm`<br>macOS: `~/Library/Application Support/Bearm` | Override configuration directory (must be an absolute path). |
| `BEARM_STATE_HOME` | Linux: `~/.local/state/bearm`<br>macOS: `.../Bearm/state` | Override journal & state directory (must be an absolute path). |
| `BEARM_DATA_HOME` | Linux: `~/.local/share/bearm`<br>macOS: `.../Bearm/data` | Override data directory (must be an absolute path). |
| `BEARM_TRASH` | Platform default | Custom absolute path for the trash root. |
| `BEARM_TRASH_PER_MOUNT` | `true` | When `true`, uses same-filesystem trash on Linux mounts (`.Trash/<uid>` / `.Trash-<uid>`). |
