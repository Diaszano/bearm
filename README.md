<div align="center">

# 🐻 Bearm

**Remove safely, restore confidently.**

A native, high-performance, crash-resilient replacement for Unix `rm` written in Go.<br>
Moves files, directories, and symlinks to your operating system's native trash instead of permanently obliterating them.

[![Go Version](https://img.shields.io/github/go-mod/go-version/Diaszano/bearm?style=flat-square&logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Platform Support](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-informational?style=flat-square)]()
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-informational?style=flat-square)]()
[![Pure Go](https://img.shields.io/badge/dependencies-zero%20external-brightgreen?style=flat-square)]()
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](CONTRIBUTING.md)

---

**Languages:** 🇺🇸 **English** • [🇧🇷 Português](README.pt-BR.md)

</div>

---

## 📖 Table of Contents

- [Why Bearm?](#-why-bearm)
- [Feature Comparison](#-feature-comparison)
- [Core Features](#-core-features)
- [Supported Platforms](#-supported-platforms)
- [Installation](#-installation)
  - [Method 1: via `go install` (Recommended)](#method-1-via-go-install-recommended)
  - [Method 2: Build from Source](#method-2-build-from-source)
- [Safe First Run (Sandbox Test)](#-safe-first-run-sandbox-test)
- [Daily Usage](#-daily-usage)
  - [1. Using Bearm Directly](#1-using-bearm-directly)
  - [2. Using Bearm as your `rm` Alias](#2-using-bearm-as-your-rm-alias)
  - [3. Bypassing the Alias](#3-bypassing-the-alias)
- [Recovery & Native Subcommands](#-recovery--native-subcommands)
  - [Listing Trashed Files (`list`)](#listing-trashed-files-list)
  - [Restoring Files (`restore`)](#restoring-files-restore)
  - [Permanent Deletion (`purge`)](#permanent-deletion-purge)
  - [Diagnosing Journal Health (`doctor`)](#diagnosing-journal-health-doctor)
  - [Configuration Validation (`config`)](#configuration-validation-config)
- [Safety & Guardrails](#-safety--guardrails)
- [Configuration](#-configuration)
- [Documentation & Deep Dives](#-documentation--deep-dives)
- [Contributing](#-contributing)
- [License](#-license)

---

## 💡 Why Bearm?

Every developer and system administrator has experienced the terror of an accidental `rm -rf`. Standard Unix `rm(1)` unlinks filesystem inodes immediately: once confirmed, the data is gone forever without an easy undo button.

Alternative trash utilities often introduce slow Python runtimes, non-standard CLI flags, unhandled cross-device failures, or incomplete flag compatibility.

**Bearm solves this cleanly:**
1. **Drop-in `rm` flag compatibility**: Keeps your muscle memory intact with full GNU and BSD flag support (`-r`, `-f`, `-i`, `-v`, `-d`, etc.).
2. **Native OS Trash**: Integrates with FreeDesktop Trash specifications on Linux and system-visible Trash directories on macOS.
3. **Atomic O(1) directory moves**: Moves entire folders in a single filesystem `rename` operation without walking millions of child files.
4. **Hardened safety rules**: Strictly refuses to delete `/`, `.`, `..`, your trash folders, or Bearm configuration files.
5. **Instant restore**: Accidental deletion? Run `bearm restore --last` to instantly bring your files back.
6. **Zero runtime overhead**: Pure Go static binary, no background daemons, zero telemetry, and zero network calls.

---

## 📊 Feature Comparison

| Feature | Standard `/bin/rm` | `trash-cli` (Python) | `rmtrash` (Shell) | **Bearm** (Go) |
| :--- | :---: | :---: | :---: | :---: |
| **Recovery / Undo** | ❌ No (Permanent) | ✅ Yes | ✅ Yes | 🛡️ **Yes (`restore --last`)** |
| **`rm` CLI Flag Parity** | ✅ Native | ❌ Incompatible flags | ⚠️ Partial | 🎯 **Full GNU & BSD parity** |
| **Directory Fast-Path** | N/A | ❌ Recursive walk | ❌ Recursive walk | ⚡ **Atomic O(1) rename** |
| **Hard Root Protection** | ⚠️ `--preserve-root` | ⚠️ Inconsistent | ❌ Minimal | 🔒 **Strict (`/`, `..`, trash, config)** |
| **Platform Trash Standard** | ❌ None | ⚠️ Linux only | ⚠️ macOS only | 🍏🐧 **Linux + macOS native** |
| **Audit Journal** | ❌ None | ⚠️ Basic metadata | ❌ None | 📜 **Append-only JSONL log** |
| **Runtime Footprint** | Low (C) | High (Python) | High (Shell subshells) | ⚡ **Single static binary** |
| **Telemetry & Network** | None | None | None | 🚫 **Zero telemetry & offline** |

---

## ✨ Core Features

- **Compatibility Profiles**: Automatic profile selection based on OS (GNU profile on Linux, BSD profile on macOS, with POSIX option).
- **Atomic Same-Filesystem Moves**: Fast-path O(1) directory moves without walking tree operands when entry-level inspection is not required.
- **Native OS Trash Integration**: FreeDesktop Trash on Linux (`~/.local/share/Trash` and per-mount `.Trash/<uid>`) and system-visible Trash on macOS (`~/.Trash`).
- **Collision-Safe Metadata**: Unique IDs and atomic metadata reservation prevent overwriting files with colliding names.
- **Hard Protection Safeguards**: Immune to `/`, `.`, `..`, lexical equivalents, and Bearm-managed data/state roots.
- **Append-Only Journal**: Every lifecycle action (`trashed`, `restored`, `purged`) is written with nanosecond timestamps to a durable JSONL journal with crash-recovery validation.
- **Strict Declarative Configuration**: Fast TOML validation without shell evaluation or code injection vectors.
- **Privacy & Security First**: Zero telemetry, zero network access, and zero subshell execution at runtime.

---

## 💻 Supported Platforms

| Operating System | Architecture | Go Requirement | Status |
| :--- | :--- | :--- | :--- |
| **Linux** | `amd64` (x86_64) | Go 1.24+ | Fully Supported (FreeDesktop Trash) |
| **Linux** | `arm64` (AArch64) | Go 1.24+ | Fully Supported (FreeDesktop Trash) |
| **macOS** | `amd64` (Intel) | Go 1.24+ | Fully Supported (Native Trash) |
| **macOS** | `arm64` (Apple Silicon) | Go 1.24+ | Fully Supported (Native Trash) |

> [!NOTE]
> Windows is not supported. Finder **Put Back** metadata is not preserved on macOS; use `bearm restore` as the authoritative recovery mechanism.

---

## 📦 Installation

Bearm is distributed directly via Go tooling and source builds. No third-party package managers are needed.

### Method 1: via `go install` (Recommended)

If you already have **Go 1.24+** installed, install the latest release with:

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

Make sure your Go binary path (usually `~/go/bin` or `$(go env GOPATH)/bin`) is added to your `$PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Verify the installation:

```bash
bearm version
```

---

### Method 2: Build from Source

Build directly on your machine with full compiler optimization:

```bash
# 1. Clone repository
git clone https://github.com/Diaszano/bearm.git
cd bearm

# 2. Build the binary
make build
# Alternatively, without make:
# go build -trimpath -o bin/bearm ./cmd/bearm

# 3. Install binary to your user bin
mkdir -p "$HOME/.local/bin"
install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
```

Ensure `$HOME/.local/bin` is in your `$PATH`.

For in-depth setup and environment configuration, see the [Installation Guide](docs/installation.md).

---

## 🧪 Safe First Run (Sandbox Test)

Before trusting Bearm with real files, run this isolated sandbox test to observe removal and recovery in action:

```bash
# 1. Create a temporary sandbox directory
trial="$(mktemp -d)"
mkdir -p "$trial/home" "$trial/work"
echo "Important document content" > "$trial/work/sample.txt"

# 2. Run bearm rm in an isolated environment
HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm rm "$trial/work/sample.txt"

# Verify file was moved away
ls -la "$trial/work/sample.txt" 2>/dev/null || echo "File safely trashed!"

# 3. Undo the operation instantly
HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm restore --last

# 4. Confirm restoration
cat "$trial/work/sample.txt"

# 5. Clean up sandbox
rm -rf "$trial"
```

---

## 🚀 Daily Usage

### 1. Using Bearm Directly

Invoke `bearm rm` using the exact same arguments you would pass to standard `rm`:

```bash
# Remove single or multiple files
bearm rm file.txt notes.md

# Remove directories recursively
bearm rm -r build/
bearm rm -rf temp_cache/

# Verbose output (prints each moved item)
bearm rm -v document.pdf

# Safe handling of filenames starting with a dash
bearm rm -- -filename-with-dash.txt
```

### 2. Using Bearm as your `rm` Alias

To seamlessly safeguard your terminal against accidental deletions, alias `rm` in your shell profile:

#### Bash (`~/.bashrc` or `~/.bash_profile`):
```bash
alias rm='bearm rm'
```

#### Zsh (`~/.zshrc`):
```zsh
alias rm='bearm rm'
```

#### Fish (`~/.config/fish/config.fish`):
```fish
alias rm='bearm rm'
```

> [!IMPORTANT]
> **Never overwrite or delete `/bin/rm`**. Always use a shell alias. System scripts, package managers, and automated cron jobs rely on `/bin/rm`.

### 3. Bypassing the Alias

When you explicitly need the raw destructive unlinking behavior:

```bash
\rm temp_file.tmp       # Prefix with backslash
command rm temp_file.tmp # Use the shell builtin
/bin/rm temp_file.tmp    # Invoke absolute binary path
```

---

## ♻️ Recovery & Native Subcommands

Bearm includes a full suite of native subcommands to manage your trashed files and inspect history.

### Listing Trashed Files (`list`)

List items currently in your trash that were removed by Bearm:

```bash
bearm list
bearm list --limit 20
bearm list --json
```

Output format:
```text
ITEM_ID           DELETION_DATE               ORIGINAL_PATH
itm_01j7abc123    2026-09-14T21:30:00-03:00   /home/user/project/main.go
```

### Restoring Files (`restore`)

Recover trashed files back to their exact original locations:

```bash
# Undo the most recent removal operation
bearm restore --last

# Restore a specific item by its Item ID
bearm restore itm_01j7abc123

# Restore all files trashed in a specific operation
bearm restore --operation op_01j7abc999
```

> [!TIP]
> If a file already exists at the destination, Bearm's default policy (`fail`) prevents accidental overwrites. You can configure `restore.collision_policy = "rename"` in your `config.toml` to automatically restore colliding files with a `.restored.N` suffix.

### Permanent Deletion (`purge`)

`bearm purge` is the **only** command in Bearm that permanently deletes files from disk:

```bash
# Interactive prompt before permanent deletion
bearm purge itm_01j7abc123

# Bypass prompt with --yes
bearm purge itm_01j7abc123 --yes

# Purge the last trashed operation permanently
bearm purge --last --yes
```

### Diagnosing Journal Health (`doctor`)

Check the integrity of your append-only journal and verify that recorded items match physical trash files:

```bash
bearm doctor
bearm doctor --json
```

### Configuration Validation (`config`)

Locate and validate your configuration file:

```bash
bearm config path   # Prints the path to config.toml
bearm config check  # Validates syntax and safety rules
```

---

## 🛡️ Safety & Guardrails

Bearm is engineered with defensive guarantees at every layer:

- **Strict Root & Lexical Boundaries**: Refuses `/`, `.`, `..`, lexical equivalents, and paths inside Bearm's internal state, configuration, or active trash directories.
- **Non-Dereferencing Symlink Handling**: A symlink operand is moved *as a symlink*; its target file or directory is never followed or touched.
- **Atomic Reservations**: Trash names are reserved exclusively. If a move fails mid-flight, metadata reservations roll back and source files remain intact.
- **No Shell Execution**: Bearm invokes standard Go filesystem syscalls directly. Filenames containing spaces, quotes, newlines, or command injection payloads are parsed safely without subshells.
- **Crash Consistency**: The append-only journal records transactions using exclusive locks and `fsync`. Power loss or sudden termination will not corrupt state.

For complete details on threat models and trust boundaries, see the [Security Model](docs/security.md).

---

## ⚙️ Configuration

Bearm works out of the box with zero configuration. You can customize behavior using a declarative TOML file:

- **Linux**: `$XDG_CONFIG_HOME/bearm/config.toml` (or `~/.config/bearm/config.toml`)
- **macOS**: `~/Library/Application Support/Bearm/config.toml`

### Example `config.toml`:

```toml
language = "pt-BR"            # "pt-BR" or "en" for native command messages
compatibility_profile = "auto" # "auto", "gnu", "bsd", or "posix"

[trash]
per_mount = true              # Linux: route to same-filesystem trash when possible
custom_path = ""              # Optional custom absolute trash directory

[safety]
preserve_root = true          # Enforce root preservation
allowed_roots = []            # Optional list of absolute directories allowed for removal
protected_patterns = [        # Glob patterns that Bearm will strictly refuse to delete
  "**/.git/**",
  "**/production.env"
]
inspect_descendants = false   # Walk recursive folders to check protected patterns

[restore]
collision_policy = "fail"     # "fail", "rename", or "overwrite"
```

### Key Environment Variables:

| Variable | Description |
| :--- | :--- |
| `BEARM_COMPAT` | Override compatibility mode profile (`gnu`, `bsd`, `posix`) |
| `BEARM_LANG` | Override native command output language (`pt-BR`, `en`) |
| `BEARM_CONFIG_HOME`| Absolute path overriding configuration directory |
| `BEARM_STATE_HOME` | Absolute path overriding journal & state directory |
| `BEARM_DATA_HOME`  | Absolute path overriding data directory |
| `BEARM_TRASH`      | Absolute path overriding custom trash directory |

For a complete description of all options, consult the [Configuration Guide](docs/configuration.md).

---

## 📚 Documentation & Deep Dives

| Guide | Description |
| :--- | :--- |
| [Installation & Setup](docs/installation.md) | In-depth setup, PATH configuration, and shell integration |
| [CLI Reference](docs/cli.md) | Complete list of commands, flags, arguments, and exit codes |
| [Configuration Reference](docs/configuration.md) | Schema specification, safety scopes, and environment variables |
| [Recovery & Journal](docs/recovery.md) | Lifecycle states, journal format, restore mechanisms, and disaster recovery |
| [GNU Compatibility](docs/compatibility/gnu.md) | Detailed parity analysis with GNU coreutils `rm` |
| [BSD Compatibility](docs/compatibility/bsd.md) | Detailed parity analysis with BSD/macOS `rm` |
| [Architecture Overview](docs/architecture/overview.md) | System components, data flow, and no-traversal fast path |
| [Security Model](docs/security.md) | Threat model, trust boundaries, and safety invariants |

---

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on code style, running differential compatibility tests, and PR procedures.

To run the full test suite locally:

```bash
make verify
```

---

## 📄 License

Bearm is licensed under the [MIT License](LICENSE).
