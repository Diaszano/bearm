# Bearm

Bearm is a native safe replacement for Unix `rm`. It moves files, symlinks,
and directories to operating-system trash locations instead of permanently
deleting them.

> Bearm is pre-1.0 software. Test it with disposable data before using an
> `rm` alias.

## Features

- GNU/Linux and BSD/macOS compatibility profiles.
- Atomic same-filesystem moves without walking ordinary recursive directory
  operands.
- FreeDesktop Trash support on Linux and system-visible Trash support on
  macOS.
- Collision-safe metadata and destination reservation.
- Hard protection for `/`, `.`, `..`, and Bearm-managed directories.
- Append-only history with list, restore, purge, and doctor commands.
- Strict declarative TOML configuration with no shell evaluation.
- No telemetry, network access, or subprocess execution at runtime.

## Supported Platforms

Bearm supports Linux and macOS on `amd64` and `arm64`. The Go version floor for
building from source is Go 1.24.

Windows is not supported. Finder **Put Back** metadata is not guaranteed on
macOS; `bearm restore` is authoritative.

## Installation

Release archives are published for supported platforms. Download the archive
for your operating system and architecture, verify it against
`checksums.txt`, and place `bearm` on your `PATH`.

After the public Homebrew tap is available:

```bash
brew install --cask Diaszano/tap/bearm
```

To build from source:

```bash
git clone https://github.com/Diaszano/bearm.git
cd bearm
make build
install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
```

Bearm never replaces `/bin/rm` automatically.

## Safe First Run

Use a disposable directory and an isolated Bearm state area:

```bash
trial="$(mktemp -d)"
mkdir -p "$trial/home" "$trial/work"
printf 'disposable\n' > "$trial/work/example.txt"

HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm rm "$trial/work/example.txt"

HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm restore --last
```

Confirm that `example.txt` is restored before trying Bearm with valuable data.

## Using Bearm Directly

Compatibility mode begins with the `rm` subcommand:

```bash
bearm rm FILE
bearm rm -rf DIRECTORY
bearm rm -- -filename
```

Successful compatibility operations are silent unless `-v` is used.
Diagnostics and prompts are written to standard error.

## Using Bearm as `rm`

Start with a reversible shell alias:

```bash
alias rm='bearm rm'
```

Put it in the appropriate shell startup file only after testing Bearm on your
operating system. Never replace, rename, or overwrite `/bin/rm`.

Bypass the alias when the host utility is explicitly required:

```bash
command rm
/bin/rm
```

An executable or symlink named `rm` that points to `bearm` also selects
compatibility mode, but an alias is easier to inspect and reverse.

## Restore and Purge

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

Restore fails by default when the original path already exists. Purge is the
only Bearm command that permanently deletes a trashed item and requires a
prompt or `--yes`. See [Recovery](docs/recovery.md).

## Configuration

Show the active configuration path and validate its contents:

```bash
bearm config path
bearm config check
```

Bearm works without a configuration file. Environment values override TOML,
which overrides platform defaults. See
[Configuration](docs/configuration.md) for the complete schema and directory
locations.

## Compatibility

Linux selects the GNU profile and macOS selects the BSD profile by default.
Override profile selection for testing:

```bash
BEARM_COMPAT=gnu bearm rm FILE
BEARM_COMPAT=bsd bearm rm FILE
BEARM_COMPAT=posix bearm rm FILE
```

Bearm compares observable behavior against the host `rm` in isolated
differential tests. Moving to trash necessarily differs from unlinking, and
Bearm's hard safety protections remain stricter. See the
[GNU](docs/compatibility/gnu.md) and [BSD](docs/compatibility/bsd.md)
compatibility references.

## Safety Guarantees

- Bearm uses direct filesystem APIs and never evaluates paths through a shell.
- Final symlink operands are moved as symlinks; their destinations are not
  followed.
- `/`, `.`, `..`, and Bearm configuration, state, data, and trash roots cannot
  be removed in compatibility mode.
- Trash names and metadata are reserved atomically and never overwrite an
  existing trash item.
- A cross-filesystem custom trash move fails without copying or deleting the
  source.
- Configuration must be owned by the current user and must not be writable by
  group or other users.

See [Security](docs/security.md) for the threat model and trust boundaries.

## Known Limitations

- Windows is not supported.
- Finder **Put Back** metadata is not preserved.
- Cross-filesystem copying into a custom centralized trash is unsupported.
- BSD secure overwrite (`-P`) and undelete (`-W`) semantics are unsupported.
- Diagnostics can vary across historical `rm` and locale implementations.
- Bearm does not provide a GUI, daemon, telemetry, or trash synchronization.

## Development

```bash
make verify
make build
./bin/bearm version
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for workflow and compatibility-fixture
requirements. Architecture details are in
[docs/architecture/overview.md](docs/architecture/overview.md).

## License

Bearm is released under the [MIT License](LICENSE).
