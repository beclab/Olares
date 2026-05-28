# files mkdir

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files mkdir --help`.

Create a directory on the per-user files-backend. Uniform across all namespaces (`POST /api/resources/<path>/`).

## Critical caveat: auto-rename quirk

`POST /api/resources/<dir>/` against an existing directory does NOT return 409 — the server silently creates `<dir> (1)` instead (see [quirk #1](../SKILL.md#1-post-apiresourcesdir-auto-renames-on-collision)).

- `-p` mode side-steps this for parents (it lists each prefix's parent and skips when the basename already exists).
- For the LEAF, the CLI prints a hint after the call so the user can `files ls` and confirm.

## Examples

```bash
olares-cli files mkdir drive/Home/Documents/Backups
olares-cli files mkdir -p drive/Home/A/B/C/
olares-cli files mkdir -p sync/<repo_id>/notes/2026/Q2
olares-cli files mkdir -p awss3/<account>/Backups/2026
olares-cli files mkdir -p google/<account>/Drafts
```

## Refusals (client-side)

- **Volume roots** (`drive/Home/`, `drive/Data/`, `sync/<repo_id>/`, etc.) — they always exist, so the call would be a no-op or trip the auto-rename quirk on the extend folder.
- **`.` or `..` segments ANYWHERE in the path** — path-traversal blacklist on raw input (before normalization), so `drive/Home/foo/../bar` errors out instead of being rewritten to `drive/Home/bar`.
- **`external/<node>/` (quirk #3 bare root)** — virtual layer with no backing filesystem.
- **`external/<node>/<single-segment>/` (depth-1 under external)** — depth-1 entries ARE the mounted volumes (USB-0, SMB-..., per-disk mount-points). Creating a new depth-1 entry would land as a phantom volume or collide with an existing mount. **Mount new volumes via LarePass first.** `-p` mode also refuses to auto-create a missing depth-1 intermediate.

## Agent notes

- **Always `-p` when chaining mkdir → upload.** `upload` does NOT pre-create directories (because of quirk #1), so the cheapest pattern is `mkdir -p <dest-dir> && upload <local> <dest-dir>`.
- After a non-`-p` mkdir, if the user wonders why they see `Foo (1)`, the answer is "you ran mkdir on an existing dir". The fix: `files rm -r drive/Home/Foo (1)/` (the bytes inside `Foo` are untouched).

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `Foo (1)` appeared instead of `Foo` | Auto-rename quirk on existing leaf | Delete the dup; in future check with `files ls` first or use `-p` |
| `refusing to mkdir external/<node>/<X>/: depth-1 entries under external/<node>/ are mounted volumes` | Quirk #3 depth-1 guard | Mount the volume via LarePass; target an existing volume's sub-path |
| `is the volume listing layer (read-only)` | Quirk #3 bare-root | Add the `<volume>` segment |
| `path contains . or .. segments` | Path-traversal blacklist | Use absolute / canonical paths only |
