# Recovery

Bearm records every completed trash move as an append-only lifecycle event.
Use Bearm's native commands to inspect and recover those items.

## List Active Items

```bash
bearm list
bearm list --limit 20
bearm list --json
```

Text output contains the item ID, deletion time, and original path. JSON output
contains the complete active `TrashRecord`, including its operation ID and
backend.

## Restore

Restore one or more item IDs:

```bash
bearm restore ITEM_ID
bearm restore ITEM_ID_1 ITEM_ID_2
```

Restore every active item from one operation:

```bash
bearm restore --operation OPERATION_ID
```

Restore the most recent active operation:

```bash
bearm restore --last
```

The default `fail` collision policy leaves both the existing original path and
the trash item untouched. Set `restore.collision_policy = "rename"` to recover
to a unique `.restored.N` sibling. The current native command rejects
unconfirmed overwrite policy.

Restoring appends a `restored` event; historical `trashed` events are not
rewritten.

## Purge

Purge permanently deletes selected active trash items:

```bash
bearm purge ITEM_ID
bearm purge ITEM_ID --yes
bearm purge --operation OPERATION_ID
bearm purge --last
```

Without `--yes`, Bearm prompts on standard error. Declining leaves every item
unchanged. Purge is available only in native mode; compatibility-mode removal
cannot permanently delete a Bearm-managed trash item.

Purge appends a `purged` event after permanent deletion. Treat `--yes` as an
irreversible operation.

## Diagnose History

```bash
bearm doctor
bearm doctor --json
```

Doctor is non-mutating. It reports:

- an incomplete trailing journal line;
- an active record whose trash item is missing.

An incomplete final JSONL line is ignored when reconstructing active state but
is retained for diagnosis.

## Journal Locations

- Linux: `$XDG_STATE_HOME/bearm/journal.jsonl`, falling back to
  `$HOME/.local/state/bearm/journal.jsonl`.
- macOS:
  `$HOME/Library/Application Support/Bearm/state/journal.jsonl`.

`BEARM_STATE_HOME` replaces the platform state root and must be absolute.

Journal writes use an exclusive lock, one JSON object per line, and `fsync`.
Records contain paths and lifecycle metadata but never file contents.

## macOS Finder

Bearm uses system-visible Trash directories but does not guarantee Finder
**Put Back** metadata. Finder may show the item without knowing its original
location. `bearm restore` and the Bearm journal are authoritative.

## Manual Incident Procedure

If `bearm doctor` reports a missing item or incomplete journal:

1. Stop running restore or purge commands against the affected item.
2. Copy the journal and relevant trash metadata to a safe diagnostic
   directory.
3. Record `bearm version`, the operating system, and the doctor output.
4. Inspect paths without editing or truncating the original journal.
5. Report the issue through the private security process when data exposure or
   unsafe deletion is possible.
