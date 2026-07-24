# Security Model

Bearm reduces accidental deletion risk; it does not turn untrusted local
filesystem access into a safe operation. Run it with the same care and account
privileges as the host `rm`.

## Protected Assets

- source files and directory trees;
- symlink identity and destination integrity;
- trash items and collision metadata;
- configuration ownership and integrity;
- append-only operation history;
- explicit user intent for permanent purge.

## Threats and Controls

### Shell and Path Injection

Bearm executes no shell and passes paths directly to Go filesystem APIs.
Configuration is strict TOML and is never sourced or evaluated. Filenames may
contain whitespace, newlines, metacharacters, leading dashes, and non-ASCII
text; use `--` before ambiguous operands.

### Symlinks

Planning and safety checks use `Lstat`. A final symlink operand is moved as a
symlink, and its destination is not followed. Trash skeleton paths must be real
directories rather than symlinks.

### Hard and Configured Protection

Compatibility mode rejects `/`, `.`, `..`, lexical equivalents, resolved
Bearm trash, and Bearm configuration, state, and data roots. Allowed roots and
protected patterns can narrow scope further. Safety is checked during planning
and again immediately before execution.

### Collision and Concurrency

Metadata reservation uses exclusive file creation. Bearm tries the base name,
then numeric suffixes only on collision. It never overwrites an existing trash
item or metadata reservation. Concurrent operations serialize journal writes
through an advisory lock.

### Trash Directory Trust

Linux per-mount selection uses `<mount>/.Trash/<uid>` only when `.Trash` is a
real sticky directory; otherwise it uses `<mount>/.Trash-<uid>`. Custom trash
paths must be absolute, and their skeleton directories are validated.

### Configuration Trust

On Linux and macOS, the configuration file must be regular, owned by the
current user, and not writable by group or other users. Unknown fields,
relative protected roots, relative custom trash, and unknown enum values stop
startup before filesystem mutation.

### Permanent Deletion

Only native `bearm purge` permanently deletes managed trash items. It requires
interactive confirmation or `--yes`. Compatibility mode cannot turn a target
inside its resolved Bearm trash into another removal operation.

### Privacy

Bearm has no telemetry and performs no runtime network access. The journal
stores original and trash paths, operation and item IDs, timestamps, backend,
device ID, and status. It never stores file contents.

## Failure Boundaries

- Reservation failure leaves the source untouched.
- Rename failure rolls back reserved metadata.
- Cross-filesystem custom trash moves fail without a copy fallback.
- Independent target failures produce a nonzero result without falsely
  reporting failed targets as moved.
- An incomplete trailing journal line is reported by `bearm doctor`.

## Limitations

Bearm cannot protect against an account or process that can modify the binary,
its dependencies, its state directory, or its trash directories outside
Bearm. It does not securely erase content, preserve Finder **Put Back**
metadata, or provide backup guarantees. Trash is recovery convenience, not a
backup.

## Reporting a Vulnerability

Follow the private reporting process in the repository
[security policy](../SECURITY.md). Do not include sensitive real-world paths,
file contents, credentials, or exploit details in a public issue.
