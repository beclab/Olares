# files rm

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files rm --help`.

Delete one or more files / directories. Batch-aware: multiple targets that share a parent collapse into one DELETE request.

## Safety constraints

- **Destructive verb — confirm intent with the user before invocation.** The CLI prompts y/N by default; `-f` skips the prompt.
- **`-f` does NOT bypass the preflight existence check** — a missing path still aborts (safer-than-Unix `rm -f`). This means a typo cannot half-delete a batch.
- **Trailing slash signals directory intent.** Without `-r`, `rm drive/Home/Foo/` errors with "<Foo> is a folder, pass -r". Without trailing slash AND target is actually a dir, errors with the same CTA.
- **Protected names** ([quirk #4](../SKILL.md#4-the-system-managed-drivehome-directories-are-protected)) are refused at the first level under `drive/Home/` only; deeper paths (`drive/Home/Pictures/Trip2024/`) are fully deletable.

## Examples

```bash
olares-cli files rm drive/Home/Documents/old.pdf
olares-cli files rm -r drive/Home/Backups/2024
olares-cli files rm -r drive/Home/Backups/2024/
olares-cli files rm -rf drive/Home/junk drive/Home/scratch/

# Batch: two siblings + one cross-parent — collapses into 2 DELETE calls.
olares-cli files rm drive/Home/a.pdf drive/Home/b.pdf sync/<repo>/old.md
```

## Preflight existence check

Runs BEFORE the confirmation prompt. Aborts (with no "will delete N entries" line printed) if:

- A target path doesn't exist on the server (typo / stale path)
- The user typed `<target>/` or passed `--recursive`, but the entry is actually a FILE
- The user typed `<target>` (no slash) without `--recursive`, but the entry is actually a DIRECTORY

Volume roots are rejected upstream by the planner; the preflight only sees real entries.

## Agent notes

- **Always `files ls` the parent before suggesting `rm`** so the user sees what's actually there. Confirms the basename and gives them a chance to abort.
- **Mixing files and folders in one `-r` invocation is unusual.** If a target list has both, split into two `rm` calls (one with `-r`, one without).
- The CLI sorts requests by `fileType + extend + parent` — output ordering is stable and useful in scripts.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `<path> is a folder; pass -r` | Trailing `/` or actual-dir without `-r` | Add `-r` |
| `<path> is a file; remove the trailing slash` | Wrong intent | Drop the `/` |
| `refusing to delete drive/Home/<protected-name>` | Quirk #4 | Pick a different target, or operate on a nested path |
| `<path> does not exist on the server` | Stale / typo | `files ls` the parent first |
| 401/403 | Token rotation | See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) |
