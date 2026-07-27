# files rename

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files rename --help`.

Rename a remote entry in place — same parent directory, new basename. Synchronous PATCH (no `task_id` polling).

## When to use this vs. `mv`

| Want to ... | Use |
|---|---|
| Change basename, same parent | `rename` (synchronous, no node) |
| Move to a different directory or volume | `mv` (async via paste queue) |

## Safety constraints

- **Destructive (mutates the server) — confirm intent with the user.**
- **`<new-name>` is a BARE basename** — no `/` or `\`. Empty, `.`, `..` are rejected.
- **Protected names** ([quirk #4](../SKILL.md#4-the-system-managed-drivehome-directories-are-protected)) refuse to be renamed at the first level under `drive/Home/`; deeper paths are fine.
- **Volume roots** (`drive/Home/`, `sync/<repo>/`, ...) are refused.
- **`.` or `..` segments ANYWHERE in `<remote-path>`** are rejected (path-traversal blacklist on raw input).

## Examples

```bash
# Rename a file.
olares-cli files rename drive/Home/Documents/foo.pdf foo-final.pdf

# Rename a directory (trailing slash on source signals dir).
olares-cli files rename drive/Home/Documents/old/ new-name

# Sync repo.
olares-cli files rename sync/<repo_id>/notes/draft.md final.md
```

## Agent notes

- Trailing slash on `<remote-path>` is preserved on the wire so the backend routes through its directory handler — keep it when the source is a dir.
- `rename` cannot be combined with a `mv`-style drop-into-directory; if the user wants both (move + rename), do `rename` first, then `mv` the result.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `refusing to rename drive/Home/<protected-name>` | Quirk #4 | Pick a different target |
| `new name must be a bare basename` | `<new-name>` contains `/` | Drop the slash; use `mv` if you actually want to move |
| `new name is empty` / `cannot rename to . or ..` | Invalid basename | Provide a real name |
| `path contains . or .. segments` | Path-traversal blacklist | Use a clean path |
