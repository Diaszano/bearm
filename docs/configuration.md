# Configuration

Bearm works without a configuration file. Defaults are loaded first, then
strict TOML, then environment overrides. An invalid configuration stops the
process before any filesystem mutation.

## Configuration Path

Show the active path:

```bash
bearm config path
```

Default locations:

- Linux: `$XDG_CONFIG_HOME/bearm/config.toml`, falling back to
  `$HOME/.config/bearm/config.toml`.
- macOS: `$HOME/Library/Application Support/Bearm/config.toml`.

`BEARM_CONFIG_HOME` replaces the platform configuration root and must be an
absolute path.

## Version 1 Schema

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

Unknown fields are rejected.

### Top-Level Values

- `language`: `pt-BR` or `en` for Bearm-native commands.
- `compatibility_profile`: `auto`, `gnu`, `bsd`, or `posix`.

### Trash

- `per_mount`: routes Linux targets to same-filesystem trash when possible.
- `custom_path`: an absolute custom trash root. Empty selects the platform
  backend. Cross-filesystem moves fail without copying or deleting the source.

macOS always resolves the system-visible home or per-volume Trash location;
`per_mount` applies to the Linux backend.

### Safety

- `preserve_root`: enables the configured compatibility root policy. Bearm's
  hard `/` protection remains active independently.
- `allowed_roots`: absolute roots that bound compatibility removals. An empty
  list allows any path not rejected by another safety rule.
- `protected_patterns`: absolute, gitignore-compatible path patterns.
- `inspect_descendants`: walks recursive operands to retain protected
  descendants. The default `false` preserves the directory fast path and
  checks only explicit operands.

Example:

```toml
[safety]
allowed_roots = ["/home/alice/work"]
protected_patterns = [
  "/home/alice/work/important/**",
  "/home/alice/work/**/.git",
]
inspect_descendants = true
```

Negated protected patterns must also be absolute, such as
`!/home/alice/work/important/cache/**`.

### Restore

`collision_policy` accepts:

- `fail`: reject an existing original path;
- `rename`: restore to a sibling such as `file.restored.1`;
- `overwrite`: reserved for explicitly confirmed overwrite flows and rejected
  by the current default restore command.

### Logging

`level` accepts `error` or `debug`. Bearm writes diagnostics locally to
standard error and never sends telemetry.

## Environment Overrides

| Variable | Effect |
| --- | --- |
| `BEARM_LANG` | Native language: `pt-BR` or `en` |
| `BEARM_COMPAT` | Compatibility profile |
| `BEARM_TRASH` | Absolute custom trash root |
| `BEARM_TRASH_PER_MOUNT` | Boolean Linux per-mount routing |
| `BEARM_LOG` | `error` or `debug` |
| `BEARM_CONFIG_HOME` | Absolute configuration root |
| `BEARM_STATE_HOME` | Absolute journal/state root |
| `BEARM_DATA_HOME` | Absolute Bearm data root |
| `XDG_CONFIG_HOME` | Linux configuration base |
| `XDG_STATE_HOME` | Linux state base |
| `XDG_DATA_HOME` | Linux data and home-trash base |
| `LC_ALL`, `LC_MESSAGES`, `LANG` | Compatibility locale selection |

Relative Bearm directory overrides are rejected. Relative XDG values are
ignored in favor of platform fallbacks.

## Security Checks

On Linux and macOS, the configuration path must:

- be a regular file, not a symlink;
- be owned by the current user;
- not be writable by group or other users.

Recommended creation:

```bash
install -d -m 0700 "$(dirname "$(bearm config path)")"
install -m 0600 /dev/null "$(bearm config path)"
bearm config check
```

`bearm config check` returns `0` for valid configuration and `3` for invalid
configuration.
