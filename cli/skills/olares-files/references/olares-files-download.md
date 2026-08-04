# files download

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files download --help`.

Download a file or directory tree from the per-user files-backend.

> **Not** the download-server task centre (`olares-cli knowledge download …`) — for URL/yt-dlp/aria2 *tasks*, use [`olares-knowledge`](../../olares-knowledge/SKILL.md).

## Safety constraints

- **Without `--resume` or `--overwrite`, the command refuses to clobber an existing local file** — confirm intent with the user before suggesting `--overwrite`.
- `--overwrite` writes to `<dst>.tmp` then renames, so the previous version stays intact until the new bytes land — safe to suggest after the user confirms.
- Directory mode mirrors the remote tree under the local destination; the remote root's own basename becomes the top-level directory (matches the LarePass folder-download UX).

## Examples

```bash
# One file into the current directory.
olares-cli files download drive/Home/Documents/report.pdf

# Same, but pick a different local name.
olares-cli files download drive/Home/Documents/report.pdf ./Q1.pdf

# Resume an interrupted download (server-driven Range; O_APPEND on the local file).
olares-cli files download drive/Home/Backups/big.tar ./big.tar --resume

# Recursively pull a folder, 4 files at a time (default).
olares-cli files download drive/Home/Documents/ ./out/ --parallel 4
```

## Agent notes

- **`Stat` always lists the parent directory** and finds the leaf in the items array — this is a workaround for [quirk #2](../SKILL.md#2-get-apiresourcesfile-no-trailing-slash-returns-http-500). You never need to suggest "just GET the file URL"; the CLI already handles it.
- **Single-file resume** uses server-driven `Range: bytes=<localSize>-` — there is no sidecar progress file. A Ctrl-C + re-run keeps making forward progress as long as the local file is preserved.
- **Directory downloads parallelize FILES, not chunks** — each file's bytes still stream sequentially. `--parallel N` bounds concurrent file fetches.
- Empty subdirectories are mirrored locally so the tree matches even when a directory has no files.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `<dst> exists; pass --overwrite or --resume` | Local target already on disk | Confirm with user, then `--overwrite` (replace) or `--resume` (continue) |
| `HTTP 500` from a raw resource URL | Quirk #2 — the bare file URL embeds bytes in JSON | Use this verb (which Stats via parent), not a manual `curl` |
| 401/403 | Token rotation or invalidation | See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) |
