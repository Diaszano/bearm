# Installation & Setup

This guide covers installing, building, and configuring **Bearm** on supported Unix platforms.

---

## Supported Platforms & Prerequisites

- **Operating Systems**: Linux (kernel 3.x+), macOS (11+ Big Sur, Monterey, Ventura, Sonoma, Sequoia).
- **Architectures**: `amd64` (x86_64), `arm64` (Apple Silicon & 64-bit ARM Linux).
- **Go Version**: **Go 1.24.0 or newer** (required to compile from source or install via `go install`).

> [!NOTE]
> Windows is not supported. Bearm relies on Unix filesystem semantics, FreeDesktop Trash specifications (Linux), and Apple macOS Trash directories.

---

## Installation Methods

Bearm is distributed exclusively via **`go install`** and **direct source compilation**. No third-party package managers (such as Homebrew or APT) are required or maintained.

### Method 1: Install via `go install` (Recommended)

If you have Go 1.24+ installed on your machine, install the latest version with a single command:

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

#### Ensure Go Binaries Are in Your `PATH`

By default, `go install` places binaries into `$(go env GOPATH)/bin` (usually `$HOME/go/bin`). Ensure this directory is in your `$PATH`:

- **For Bash** (`~/.bashrc` or `~/.bash_profile`):
  ```bash
  export PATH="$(go env GOPATH)/bin:$PATH"
  ```

- **For Zsh** (`~/.zshrc`):
  ```zsh
  export PATH="$(go env GOPATH)/bin:$PATH"
  ```

- **For Fish** (`~/.config/fish/config.fish`):
  ```fish
  fish_add_path (go env GOPATH)/bin
  ```

Reload your shell or run `source ~/.bashrc` (or `source ~/.zshrc`), then verify:

```bash
bearm version
```

---

### Method 2: Build from Source

Building locally gives you full control over compilation flags and reproducible builds.

1. **Clone the repository**:
   ```bash
   git clone https://github.com/Diaszano/bearm.git
   cd bearm
   ```

2. **Build the binary**:

   Using `make`:
   ```bash
   make build
   ```

   Or directly using the standard `go` tool:
   ```bash
   go build -trimpath -o bin/bearm ./cmd/bearm
   ```

3. **Install the binary onto your `PATH`**:

   For user-local installation (recommended, no `sudo` needed):
   ```bash
   mkdir -p "$HOME/.local/bin"
   install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
   ```
   *(Ensure `$HOME/.local/bin` is in your `$PATH`)*.

   For system-wide installation:
   ```bash
   sudo install -m 0755 bin/bearm /usr/local/bin/bearm
   ```

4. **Verify the installation**:
   ```bash
   bearm version
   ```

---

## Verifying the Installation

Run `bearm version` to inspect the build metadata:

```bash
bearm version
```

Expected output format:
```text
bearm dev (commit: none, built at: unknown, go1.24.0, linux/amd64)
```

You can also check the configuration paths recognized by Bearm:

```bash
bearm config path
bearm config check
```

---

## Shell Integration: Aliasing `rm`

To use Bearm transparently as a safe alternative to `rm`, set up a shell alias.

> [!CAUTION]
> **Never** replace, overwrite, or delete `/bin/rm` on your system. System scripts and package managers may depend on exact unlinking behavior. Always use a reversible user-level shell alias.

### Bash (`~/.bashrc` or `~/.bash_profile`)

Append the following line:

```bash
alias rm='bearm rm'
```

Apply changes:
```bash
source ~/.bashrc
```

### Zsh (`~/.zshrc`)

Append the following line:

```zsh
alias rm='bearm rm'
```

Apply changes:
```zsh
source ~/.zshrc
```

### Fish (`~/.config/fish/config.fish`)

Append the following line:

```fish
alias rm='bearm rm'
```

---

## Bypassing the Alias

When you explicitly need the system's destructive unlinking utility `/bin/rm`:

1. Prefix with a backslash to bypass alias expansion:
   ```bash
   \rm unwanted-file.tmp
   ```
2. Use the `command` builtin:
   ```bash
   command rm unwanted-file.tmp
   ```
3. Use the absolute path:
   ```bash
   /bin/rm unwanted-file.tmp
   ```

---

## Upgrading Bearm

To upgrade to the newest release when installed via `go install`:

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

To upgrade from source:

```bash
cd bearm
git pull origin main
make build
install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
```

---

## Uninstallation

To remove Bearm completely:

1. Remove the binary:
   ```bash
   rm -f "$(which bearm)"
   # or: rm -f "$HOME/go/bin/bearm" "$HOME/.local/bin/bearm"
   ```

2. Remove the shell alias from your `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`.

3. *(Optional)* Remove configuration and state files:
   - **Linux**:
     ```bash
     rm -rf "${XDG_CONFIG_HOME:-$HOME/.config}/bearm"
     rm -rf "${XDG_STATE_HOME:-$HOME/.local/state}/bearm"
     rm -rf "${XDG_DATA_HOME:-$HOME/.local/share}/bearm"
     ```
   - **macOS**:
     ```bash
     rm -rf "$HOME/Library/Application Support/Bearm"
     ```
