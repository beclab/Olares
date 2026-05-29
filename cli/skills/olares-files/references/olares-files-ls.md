# files ls

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first for the profile model, the 3-segment frontend path, and the 5 client-side quirks.
> **Flags & wire shape:** `olares-cli files ls --help` (single source of truth).

List a directory on the per-user files-backend. Uniform across all 10 namespaces.

## Examples

```bash
olares-cli files ls drive/Home/
olares-cli files ls drive/Home/Documents
olares-cli files ls sync/<repo_id>/
olares-cli files ls awss3/<account>/<bucket>
olares-cli files ls cache/<node>/
olares-cli files ls external/<node>/           # virtual volume-listing layer (see SKILL.md quirk #3)
olares-cli files ls drive/Home/Documents --json  # raw envelope, pretty-printed
```

## Output shape

Default table: `MODE  SIZE  TYPE  MODIFIED  NAME`. Directories sort before files; directory names get a trailing `/`. Empty directories print `(empty)`.

`--json` prints the raw JSON envelope, useful for scripting.

## Envelope shapes (transparent to the user, matters when reading `--json`)

| Namespace | Children field | Per-item size | `mode` / `modified` |
|---|---|---|---|
| `drive` / `sync` / `cache` / `external` / `share` | `items` | `size` (number) | numeric `mode`, RFC3339 `modified` |
| `awss3` / `google` / `dropbox` / `tencent` | `data` | `fileSize` | empty strings; the table renders `d---------` / `----------` and `-` in MODE / MODIFIED |

The cloud envelope ALSO omits the parent-level `numDirs` / `numFiles` / `modified` summary; the table header falls back to counting items so it stays informative.

## Agent notes

- `ls` is the canonical discovery verb. Use it before any write to confirm parent existence and the exact basename casing.
- `ls external/<node>/` is the right way to discover attached volumes (`hdd1`, `usb1`, `smb-...`) before targeting `external/<node>/<volume>/<sub>/`.
- `ls cache/<node>/` is the right way to discover what's under a node before sharing — share-create rejects bare `cache/<node>/`.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `invalid drive type` | `drive/home/...` instead of `drive/Home/...` | Use exact casing: `Home` or `Data` |
| Empty `items` / `data` array on a known-non-empty dir | Wrong identity — the active profile can't see this scope | `olares-cli profile list` and switch with `profile use` |
| 401/403 | Token rotation or invalidation | See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) |
