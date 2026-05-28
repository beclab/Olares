# files chown

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files chown --help`.

Get or set the POSIX owner UID of a file / directory. CLI counterpart of the LarePass web app's "Permission" tab.

## Modes

| Form | What it does |
|---|---|
| `files chown <path>` | GET — print the current uid |
| `files chown <path> --uid <int>` | PUT — replace the uid |
| `files chown <path> --uid <int> -r` | PUT — recurse into children |

## UID conventions (LarePass presets)

| UID | Meaning |
|---|---|
| `0` | Root (system; only set this if you know why) |
| `1000` | User (the default LarePass user; matches the GUI's "User" preset) |

Any integer is accepted, but these are the values the GUI surfaces.

## Supported namespaces (allow-list)

`drive/Home/<sub>`, `drive/Data/<sub>`, `cache/<node>/<sub>` only.

Refused namespaces:

| Namespace | Why |
|---|---|
| `sync/<repo_id>/...` | Seafile permissions live on the library itself — use `files repos` |
| `external/<node>/<volume>/...` | LarePass GUI hides the Permission tab for external mounts |
| `awss3` / `dropbox` / `google` / `tencent` | Object stores have no POSIX uid concept |

## Safety constraints

- **Destructive when `--uid` is provided — confirm intent with the user.** UID changes affect every app that reads the directory.
- **Volume roots are refused** (`drive/Home/`, `drive/Data/`, `cache/<node>/`) — chowning an entire namespace root has too much blast radius. Pick a one-level-deeper path with `-r` if you need to fan out.
- **`-r` recurses** — every descendant gets the new uid. Confirm directory contents with `files ls` first.

## Examples

```bash
# Inspect.
olares-cli files chown drive/Home/Documents/foo.pdf

# Hand a file to root.
olares-cli files chown drive/Home/Documents/foo.pdf --uid 0

# Hand an entire directory tree to the default user.
olares-cli files chown drive/Home/Pictures/Trip2024/ --uid 1000 -r

# Cache namespace.
olares-cli files chown cache/<node>/scratch/build/ --uid 1000 -r
```

## Agent notes

- The GET form is cheap — use it before any PUT to show the user the current uid and confirm the change.
- Inside `-r`, partial-failure behavior is per-server — on an error the server may have already changed some descendants. **Do NOT retry blindly**; re-run the GET form on a few sampled paths to see what landed.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `chown is not supported for this namespace` | sync / external / cloud target | Use `files repos` (sync) or LarePass GUI (external) |
| `refusing to chown a volume root` | `drive/Home/` / `drive/Data/` / `cache/<node>/` | Pick a sub-path |
| 403 from server | Server-side ACL rejection | Confirm via `files ls -ld` (when available) or LarePass that the active user has permission |
