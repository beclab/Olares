# files cp / files mv

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files cp --help` and `olares-cli files mv --help`.

Copy / move one or more entries between locations. Same wire endpoint (`PATCH /api/paste/<node>/`), different `action`. Cross-volume (drive ↔ sync ↔ external) is supported.

## Safety constraints

- **Destructive (mutates the server) — confirm intent with the user.** `mv` even more so since the source is removed.
- **`<dst> MUST end with `/` (drop-into-directory mode).** Each `<src>`'s basename is appended; preserves the dir / file marker.
- **Renaming via `cp` / `mv` is not supported** — use `files rename` for in-place basename changes, or rename first and then `mv`.
- **Directory sources require `-r`** (Unix-style refusal otherwise).
- **`mv` source rejects protected names** ([quirk #4](../SKILL.md#4-the-system-managed-drivehome-directories-are-protected)): `mv drive/Home/Pictures/ ...` is refused because moving would unlink a dir that apps depend on. **`cp` (copy) is intentionally NOT gated** — duplicating bytes (e.g. `cp -r drive/Home/Pictures/ drive/Home/Pictures-Backup/`) preserves the original and is fine.
- **`external/<node>/` destinations are rejected** ([quirk #3](../SKILL.md#3-externalnode-is-a-virtual-volume-listing-layer-read-only)) — point at `external/<node>/<volume>/<sub>/`.

## Examples

```bash
# One file → directory.
olares-cli files cp drive/Home/notes.md drive/Home/Documents/

# Recursive directory copy.
olares-cli files cp -r drive/Home/Photos/ drive/Home/Backups/

# Multiple sources into a directory.
olares-cli files cp drive/Home/a.pdf drive/Home/b.pdf drive/Home/Archive/

# Cross-volume (drive → sync repo).
olares-cli files cp drive/Home/notes.md sync/<repo_id>/inbox/

# Move (mv replaces cp where source removal is intended).
olares-cli files mv drive/Home/notes.md drive/Home/Archive/
olares-cli files mv -r drive/Home/Photos/ drive/Home/Backups/
```

## Preflight existence check

Runs BEFORE any PATCH is sent:

- Each `<src>` MUST exist on the server, AND its trailing-slash form must match the actual file/dir kind.
- `<dst>` MUST exist as a directory on the server. **Create it first with `files mkdir -p` if needed** — `cp`/`mv` does NOT auto-create the destination (the auto-rename quirk #1 would land you in `<dst> (1)`).

A typo on either side aborts before the server's task queue sees it.

## Node selection (`--node`)

Each PATCH carries a `{node}` URL segment. Default cascade:

1. `--node` override (per invocation)
2. External / Cache `<extend>` (when the destination is `external/<node>/` or `cache/<node>/` — the GUI's `dst_node || src_node || default` cascade)
3. First entry from `/api/nodes/`

Use `--node` only when you have a specific multi-node deployment with a non-default node hosting the paste task.

## Agent notes

- **For renames, ALWAYS use `files rename`** — it's synchronous (no task queue) and works in place. Don't try to fake a rename with `cp`/`mv`.
- **`mv` is async** — the response is "task accepted", not "move completed". For most paths this is fast, but on huge directory trees consider running an `ls` afterward to confirm.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `<dst> must end with /` | Drop-into-dir mode requires trailing slash | Add `/` |
| `<dst> does not exist on the server` | Destination dir not pre-created | `files mkdir -p <dst>` first |
| `<src> is a directory; pass -r` | Directory source without recursion | Add `-r` |
| `refusing to mv drive/Home/<protected-name>` | Quirk #4 mv-source guard | Use `cp -r` (preserves the original) if the goal is duplication |
| `is the volume listing layer (read-only)` | `external/<node>/` destination | Add `<volume>` segment |
