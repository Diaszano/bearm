# BSD/macOS `rm` Compatibility

The BSD profile is selected automatically on macOS or explicitly with
`BEARM_COMPAT=bsd`.

## Supported Options

| Option | Behavior |
| --- | --- |
| `-f`, `--force` | Ignore missing operands and disable prior interaction |
| `-i`, `--interactive=always` | Prompt for each logical target |
| `-I`, `--interactive=once` | Prompt once for recursive or large requests |
| `-r`, `-R`, `--recursive` | Permit directory removal |
| `-d`, `--dir`, `--directory` | Permit empty directory removal |
| `-v`, `--verbose` | Print successful logical removals |
| `--` | End option parsing |

Combined short options are supported.

## Option Order

BSD and POSIX profile parsing stops at the first operand. Later values that
begin with `-` are treated as operands unless `--` was used before the first
operand. Before the first operand, options are applied left to right: later
force and interaction options win.

```bash
bearm rm -f file -v          # -v is an operand
bearm rm -i -f missing       # force wins
bearm rm -- -filename        # explicit option terminator
```

## Diagnostics and Exit Status

The BSD profile uses BSD-style usage and path diagnostics. Usage failures
return `64` (`EX_USAGE`) where the selected profile uses it; operational
failures return `1`, and success returns `0`.

Successful compatibility operations write nothing unless `-v` is active.
Prompts and diagnostics are written to standard error. Locale selection checks
`LC_ALL`, then `LC_MESSAGES`, then `LANG`, and falls back to English.

## Unsupported Semantics

Bearm rejects BSD `-P` secure-overwrite and `-W` undelete semantics because
they conflict with move-to-trash behavior. Unsupported profile-specific
options fail with a usage diagnostic.

## Stricter Bearm Protections

Bearm always refuses `/`, `.`, `..`, lexical equivalents, active trash roots,
and its own configuration, state, and data roots. It moves a symlink operand
as a symlink and never removes its destination.

Finder **Put Back** metadata is not preserved. Bearm's journal and
`bearm restore` command are authoritative.

## Differential Test Host

macOS CI runs the suite on GitHub-hosted `macos-latest` with both Go 1.24 and
stable Go. Each run resolves the installed host `/bin/rm` and selects the BSD
profile; tests compare exit status, stable streams, and normalized source-tree
outcomes. The runner image supplies the exact BSD `rm` version, so
compatibility is tied to the host image recorded by that CI run.
