# GNU `rm` Compatibility

The GNU profile is selected automatically on Linux or explicitly with
`BEARM_COMPAT=gnu`.

## Supported Options

| Option | Behavior |
| --- | --- |
| `-f`, `--force` | Ignore missing operands and disable prior interaction |
| `-i`, `--interactive=always` | Prompt for each logical target |
| `-I`, `--interactive=once` | Prompt once for recursive or large requests |
| `--interactive=never` | Disable prompting |
| `-r`, `-R`, `--recursive` | Permit directory removal |
| `-d`, `--dir`, `--directory` | Permit empty directory removal |
| `-v`, `--verbose` | Print successful logical removals |
| `--one-file-system` | Keep traversal on the operand's filesystem |
| `--preserve-root` | Enable compatibility root preservation |
| `--preserve-root=all` | Protect directory operands on a different device from their parent |
| `--no-preserve-root` | Disable only GNU-level root preservation |
| `--help`, `--version` | Print compatibility help or Bearm build information |
| `--` | End option parsing |

Combined short options are supported. GNU options may occur after operands
until `--` ends option parsing.

## Option Order

Options are applied from left to right. A later `-f` or
`--interactive=never` disables prior interaction. A later `-i`, `-I`, or
`--interactive=...` clears force and establishes the new interaction mode.

```bash
bearm rm -i -f missing       # force wins
bearm rm -f -i file          # interactive always wins
bearm rm file -v             # -v is parsed as an option
```

## Diagnostics and Exit Status

The GNU profile uses GNU-style usage and path diagnostics and returns:

- `0` when all requested operations succeed or are ignored under force;
- `1` for usage and operational failures.

Successful compatibility operations write nothing unless `-v` is active.
Prompts and diagnostics are written to standard error. Locale selection checks
`LC_ALL`, then `LC_MESSAGES`, then `LANG`, and falls back to English.

## Unsupported Semantics

Unknown or incompatible options fail instead of being ignored. Bearm does not
implement BSD `-P`, BSD `-W`, or implementation-specific destructive
extensions outside the documented profile.

## Stricter Bearm Protections

`--no-preserve-root` does not disable Bearm's hard protection for `/`. Bearm
also refuses `.`, `..`, lexical equivalents, active trash roots, and its own
configuration, state, and data roots. A target already inside Bearm-managed
trash cannot be permanently deleted through compatibility mode.

## Differential Test Host

Linux CI runs the suite on GitHub-hosted `ubuntu-latest` with both Go 1.24 and
stable Go. Each run resolves the installed host `rm` and selects the GNU
profile; tests compare exit status, streams where stable, and normalized
source-tree outcomes. The runner image supplies the exact GNU coreutils
version, so compatibility is tied to the version reported by that CI run
rather than a hard-coded historical release.
