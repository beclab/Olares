---
name: olares-files
version: 1.0.0
description: "olares-cli files command tree: list (ls), upload, download, cat, and rm against the per-user files-backend (drive/Home, drive/Data, sync, cache, external, awss3, dropbox, google, tencent, share). Covers the 3-segment frontend path schema (<fileType>/<extend>/<subPath>), resumable chunked upload (Drive v2 protocol), Range-based resumable download, recursive directory transfer with errgroup parallelism, batch DELETE wire shape, and two server-side quirks the user MUST know about (POST mkdir auto-renames existing dirs to 'Foo (1)'; GET single-file resource returns HTTP 500). Use whenever the user mentions files / drive / Home / Data / sync / cache, uploading or downloading files, listing a remote directory, deleting remote files, cat-ting a remote file, resumable transfers, /api/resources, /api/raw, frontend path, or sees errors like 'Documents (1)' appearing on the server."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli files --help"
---

# files (Drive v2 + per-user files-backend)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for the profile selection, login, and HTTP 401/403 recovery rules that every command here depends on.**

## Core concept: the 3-segment frontend path

Every resource on the per-user files-backend is addressed by a 3-segment "frontend path" (see [`cli/cmd/ctl/files/path.go`](cli/cmd/ctl/files/path.go)):

```
<fileType>/<extend>[/<subPath>]
```

| Segment | Meaning |
|---------|---------|
| `fileType` | Storage class (lowercase, case-sensitive). One of: `drive`, `cache`, `sync`, `external`, `awss3`, `dropbox`, `google`, `tencent`, `share`, `internal` |
| `extend` | Volume / repo / account inside that class. **Case-sensitive.** Drive: only `Home` or `Data`. Cache / external: node name. Sync: seafile repo id. Cloud (`awss3`/`dropbox`/`google`/`tencent`): account key |
| `subPath` | Path inside `extend` (root if omitted). The leading `/` is implicit |

Examples:

```bash
drive/Home/                            # Home volume root
drive/Home/Documents/report.pdf        # a file under Home/Documents
drive/Data/Backups/                    # Data volume, Backups subfolder
sync/<repo_id>/notes/                  # seafile sync repo
cache/<node>/                          # node-local cache
awss3/<account>/<bucket>/key.txt       # S3-compatible cloud drive
```

> The first segment is normalized to lowercase by the backend; the CLI accepts only the canonical lowercase form on input. Drive's `extend` MUST be `Home` or `Data` exactly — `home` will be rejected with `invalid drive type`.

## Trailing-slash convention (critical)

Whether a path ends with `/` is meaningful and changes command behavior:

| Path form | Meaning |
|-----------|---------|
| `drive/Home/Foo/` | Directory intent |
| `drive/Home/Foo` | File intent |

This shows up in three places:

- `files rm drive/Home/Foo/` requires `-r` (the trailing `/` declares "this is a directory")
- `files upload <local> drive/Home/Documents/` means "upload INTO Documents/"; `files upload <local> drive/Home/Documents/2026-Q1.pdf` means "upload AS that exact path (rename on the way in)"
- `files ls drive/Home/` lists the volume root; the parser tolerates both `drive/Home` and `drive/Home/` for ls but the trailing slash is recommended for clarity

## Server-side quirks (critical, do not work around)

These are real backend behaviors that have already cost us debugging time. Teach yourself and the user to respect them; **do not** suggest "workarounds" that bypass the CLI's existing handling.

### 1. POST `/api/resources/<dir>/` auto-renames existing directories

Hitting the directory-create endpoint against an existing directory does **not** return 409. The server creates a sibling named `<dir> (1)` instead. See the docstring on [`cli/internal/files/upload/api.go`](cli/internal/files/upload/api.go)'s `Mkdir` for the precise wording.

Consequence baked into the CLI: `files upload` does **not** pre-create the destination directory. It relies on the chunk POST to implicitly materialize parents. **The destination directory MUST already exist on the server** — if you need a fresh directory, create it through the LarePass web app first (a future `files mkdir` verb may cover this).

User-visible symptom of getting this wrong (older CLI versions): an extra `Documents (1)` directory appears on the server even though the upload "succeeded".

### 2. GET `/api/resources/<file>` (no trailing slash) returns HTTP 500

The backend's single-file `List` handler hard-codes `Content: true` (`files/pkg/drivers/posix/posix/posix.go` `getFiles`) and tries to slurp the file's bytes into the response envelope. For json / binary / large files, this just 500s.

Consequence baked into the CLI: `Stat` always lists the **parent** directory and looks up the leaf in its items array (see [`cli/internal/files/download/stat.go`](cli/internal/files/download/stat.go)). This matches what the LarePass web app does — it never probes a single-file resource directly. Both `download` and `cat` use this code path.

If the user reports `HTTP 500` against `/api/resources/.../<filename>` with no trailing slash, do NOT suggest "just retry". The right answer is: use the CLI command (`files cat` / `files download`), or list the parent and look at items.

## Authentication transport

Every files API call carries `X-Authorization: <access_token>` as a header (NOT the standard `Authorization: Bearer ...`). The Factory injects this automatically; see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go). Do not try to call the backend via `curl` with a Bearer token — that header shape is not what the per-user files-backend expects and the request will fail.

## Command cheatsheet (5 verbs)

### `files ls <path> [--json]`

List a remote directory. See [`cli/cmd/ctl/files/ls.go`](cli/cmd/ctl/files/ls.go).

```bash
olares-cli files ls drive/Home/
olares-cli files ls drive/Home/Documents
olares-cli files ls sync/<repo_id>/
olares-cli files ls drive/Home/Documents --json   # raw envelope, pretty-printed
```

Default output: a one-line header (`<path>  (N dirs, M files, modified ...)`) followed by a 5-column table `MODE  SIZE  TYPE  MODIFIED  NAME`. Directories sort before files; directory names get a trailing `/`. Empty directories print `(empty)`.

`--json` prints the raw JSON envelope from the backend, useful for scripting.

### `files upload <local-path> <remote-path>`

Resumable chunked upload to drive/Home/<...>. See [`cli/cmd/ctl/files/upload.go`](cli/cmd/ctl/files/upload.go) and [`cli/internal/files/upload/`](cli/internal/files/upload/).

```bash
# Upload one file into an existing directory.
olares-cli files upload report.pdf drive/Home/Documents/

# Upload AND rename on the server.
olares-cli files upload report.pdf drive/Home/Documents/2026-Q1.pdf

# Upload a whole directory tree.
olares-cli files upload ./photos drive/Home/Backups/

# Two files in flight at a time, chunks remain sequential per file.
olares-cli files upload ./photos drive/Home/Backups/ --parallel 2
```

Wire protocol (Drive v2 / Resumable.js-compatible):

1. `GET /upload/upload-link/<node>/...` → upload session
2. `GET /upload/file-uploaded-bytes/<node>/...` → server-driven resume offset (no local progress file)
3. `POST` chunks (8 MiB default) with `Content-Range: bytes <start>-<end>/<total>` until done

Constraints / flags:

- **Destination MUST be under `drive/Home`** (`drive/Data` is read-only on the wire); the CLI rejects anything else with `upload destination must be under drive/Home`.
- **Destination directory MUST already exist** — see "POST auto-renames" above.
- A trailing `/` on `<remote-path>` means "into this directory"; without one, `<remote-path>` is treated as the full target path (rename on the way in).
- `--parallel N` (default 2): per-file concurrency. **Per-file chunks remain sequential** by design — the resume probe assumes one in-flight chunk per file.
- `--chunk-size <bytes>` (default 8 MiB): align with the server's expected size; rarely needs tuning.
- `--max-retries N`: per-chunk retry budget on transient failures.
- `--node <name>`: override the upload node; default is the first node from `/api/nodes/`.

Resume is automatic and server-driven: re-running the same command after a Ctrl-C / network drop just re-asks the server how many bytes it already has, floors to a chunk boundary, and continues.

### `files download <remote-path> [<local-path>]`

Download a single file or a whole directory tree. See [`cli/cmd/ctl/files/download.go`](cli/cmd/ctl/files/download.go) and [`cli/internal/files/download/`](cli/internal/files/download/).

```bash
# Single file into the current directory (./<basename>).
olares-cli files download drive/Home/Documents/report.pdf

# Same, but pick a different local name.
olares-cli files download drive/Home/Documents/report.pdf ./Q1.pdf

# Resume an interrupted download via Range:.
olares-cli files download drive/Home/Backups/big.tar ./big.tar --resume

# Recursively pull a directory; 4 files at a time.
olares-cli files download drive/Home/Documents/ ./out/ --parallel 4
```

Local destination resolution (single-file mode):

| `<local-path>` | Result |
|----------------|--------|
| omitted | `./<basename(remote)>` |
| existing directory | `<local-path>/<basename(remote)>` (mirrors `cp`) |
| any other path (incl. trailing `/` if not yet existing) | treated as the full target file path |

Flags:

- `--resume`: send `Range: bytes=<localSize>-` and append (server-native, no sidecar progress file).
- `--overwrite`: replace an existing local file via `<dst>.tmp` + atomic rename. The previous version stays intact until the new one lands.
- `--resume` and `--overwrite` are **mutually exclusive** — pick one.
- `--parallel N` (default 4): only meaningful in directory mode (errgroup-bounded concurrency).
- `--max-retries N`: per-file transient-failure budget (5xx triggers retry; 4xx fails fast).

Directory mode (trailing `/` on `<remote-path>`):

- The remote tree is walked recursively via `/api/resources/.../`.
- The remote root's basename becomes the top-level directory under `<local-path>` (matches the LarePass folder-download UX). Empty subdirectories are mirrored locally.
- Single `Stat` lookup at the start to confirm the path is actually a directory; then `BuildPlan` materializes the file list before any byte is written.

### `files cat <remote-file>`

Stream a single file's bytes to stdout. See [`cli/cmd/ctl/files/cat.go`](cli/cmd/ctl/files/cat.go).

```bash
olares-cli files cat drive/Home/Documents/notes.md
olares-cli files cat drive/Home/Logs/today.log | tail -n 50
olares-cli files cat drive/Home/Photos/banner.png > banner.png  # binary-safe
```

Wire shape: `GET /api/raw/<encPath>?inline=true` (the same endpoint LarePass uses for text previews; `inline=true` only affects `Content-Disposition`, the body is identical).

- Binary-safe: bytes are copied verbatim, no sniffing or transformation. Pipe into `less` / `head -c` / `hexdump` as needed.
- Pre-flight `Stat` (parent listing) refuses directories early with a clear error, instead of letting the server return its terser 400. Use `files download` for directories.

### `files rm [-r] [-f] <remote-path>...`

Delete one or more remote files / directories. See [`cli/cmd/ctl/files/rm.go`](cli/cmd/ctl/files/rm.go) and [`cli/internal/files/rm/`](cli/internal/files/rm/).

```bash
# Delete one file.
olares-cli files rm drive/Home/Documents/old.pdf

# Recursively remove a directory.
olares-cli files rm -r drive/Home/Backups/2024/

# Multiple targets, no prompt (scripts).
olares-cli files rm -rf drive/Home/junk drive/Home/scratch/
```

Wire shape (one batch DELETE per parent dir):

```
DELETE /api/resources/<encParentDir>/   body: {"dirents": ["<name1>", "<name2>", ...]}
```

Targets sharing a parent collapse into a single request (matches the LarePass web app's `batchDeleteFileItems`). Targets across different parents send one request each, sorted by `fileType + extend + parent` for stable output.

Flags / rules:

- `-r` / `-R` / `--recursive`: required for directories. A trailing `/` on a target IS a directory-intent signal and triggers the same check (so `files rm drive/Home/Foo/` errors without `-r` even if `Foo` is technically empty).
- `-f` / `--force`: skip the y/N prompt. **In a non-TTY context (CI, piped stdin) the command refuses without `--force`** rather than guessing.
- Without `-f`: prints "will delete N entries in M batches" with the full list, then prompts `[y/N]`.
- Removing the root of a volume (`drive/Home/`, `sync/<repo>/`, ...) is rejected by the planner: `refusing to delete the root of <fileType>/<extend>`.

Aliases: `olares-cli files remove ...`, `olares-cli files delete ...` are the same command.

## Common errors → fixes

| Error message (excerpt) | Likely cause | Fix |
|-------------------------|--------------|-----|
| `server rejected the access token (HTTP 401)` / `(HTTP 403)` | Token expired / revoked | Follow olares-shared's recovery: `olares-cli profile login --olares-id <id>` |
| `HTTP 404 ... not found on the server` | Path typo or wrong case (`Home` vs `home`, `Data` vs `data`) | `files ls` the parent directory to confirm spelling |
| `invalid drive type: <x>` | Drive's `extend` isn't `Home` or `Data` | Use exactly `Home` or `Data` |
| `upload destination must be under drive/Home` | Tried to upload to `drive/Data/...` or another fileType | Move the target under `drive/Home/...` |
| `Documents (1)` (or similar) appearing on the server after upload | Older CLI version triggered the POST-mkdir auto-rename quirk | Upgrade to a CLI version that has the pre-mkdir removal fix |
| `--overwrite and --resume are mutually exclusive` | Passed both download flags | Pick one |
| `refusing to delete without --force in a non-interactive context (no TTY)` | `files rm` from a script with no TTY | Add `-f` after **explicitly listing the targets to the user first** |
| `refusing to delete the root of <fileType>/<extend>` | Tried `files rm -r drive/Home/` (or another volume root) | The CLI does not support volume-root deletion; remove children individually |
| `cat ... is a directory` | `files cat` on a path that resolves to a directory | Use `files download <path>/` instead |
| `HTTP 500` against `/api/resources/.../<filename>` (no trailing slash) | Hit the backend's single-file List quirk directly (e.g. via curl) | Don't bypass the CLI; the CLI uses parent-listing Stat for a reason |

## Typical workflow

```bash
# 1. Explore
olares-cli files ls drive/Home/
olares-cli files ls drive/Home/Documents/

# 2. Push a local tree up (target dir must already exist).
olares-cli files upload ~/local-dir drive/Home/Documents/

# 3. Pull a tree down with parallelism + resume.
olares-cli files download drive/Home/Documents/ ./out --parallel 4 --resume

# 4. Quick peek at a file.
olares-cli files cat drive/Home/Notes/today.md | tail -n 50

# 5. Clean up after confirming with `ls` first.
olares-cli files ls drive/Home/Old/
olares-cli files rm -r drive/Home/Old/
```

When operating across multiple Olares instances, prefix each command with `--profile <olaresId>` (see `olares-shared` for the global flag) instead of flipping the persistent current pointer.

## Security rules

- **Always preview destructive operations.** Before passing `-f` to `rm` in a script, list the exact paths to the user and get explicit confirmation. The interactive `[y/N]` prompt is a safety net, not a substitute for thoughtful intent.
- **Local files are never overwritten implicitly.** `files download` refuses to clobber unless `--overwrite` (atomic via `.tmp`+rename) or `--resume` (append) is passed. Never recommend `--overwrite` without checking with the user.
- **Do not echo `<access_token>` to the terminal.** The token lives in the OS keychain for a reason; pulling it out into a shell variable for `curl` defeats that. Use the CLI commands.
- **`files upload` does NOT delete the local source** — it's a copy, not a move. If a user wants delete-after-upload semantics, they have to do it explicitly and after verifying the upload succeeded.
