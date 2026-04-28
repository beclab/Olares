---
name: olares-files
version: 1.4.0
description: "olares-cli files command tree: list (ls), upload, download, cat, rm, cp, mv, rename (rn), share (internal / public / smb), and Sync-repo CRUD (repos list / get / create / rename / rm) against the per-user files-backend (drive/Home, drive/Data, sync, cache, external, awss3, dropbox, google, tencent, share). Covers the 3-segment frontend path schema (<fileType>/<extend>/<subPath>), resumable chunked upload (Drive v2 protocol), Range-based resumable download, recursive directory transfer with errgroup parallelism, batch DELETE wire shape, server-side copy/move via PATCH /api/paste/<node>/ (async task_id queue, cross-volume supported), synchronous in-place rename via PATCH /api/resources/.../?destination=..., folder-share creation across the three flavors (Internal cross-user, Public link with password+expiration, SMB Samba mount) via POST /api/share/share_path/<...>/, share management (list / get / rm) plus SMB-account roster (smb-users list/create), Sync (Seafile) library catalog management via /api/repos/ (GET list with type=mine|share_to_me|shared, POST create with ?repoName=, PATCH rename with ?destination=&repoId=, DELETE with ?repoId=), and two server-side quirks the user MUST know about (POST mkdir auto-renames existing dirs to 'Foo (1)'; GET single-file resource returns HTTP 500). Use whenever the user mentions files / drive / Home / Data / sync / cache, uploading or downloading files, listing a remote directory, deleting remote files, cat-ting a remote file, copying or moving (renaming) remote files / directories, in-place renaming, sharing a folder with other users, public links with password / expiration, SMB / Samba network shares, listing / creating / renaming / deleting Sync (Seafile) libraries, repo_id discovery, /api/resources, /api/raw, /api/paste, /api/share, /api/repos, frontend path, or sees errors like 'Documents (1)' appearing on the server."
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

This shows up in five places:

- `files rm drive/Home/Foo/` requires `-r` (the trailing `/` declares "this is a directory")
- `files upload <local> drive/Home/Documents/` means "upload INTO Documents/"; `files upload <local> drive/Home/Documents/2026-Q1.pdf` means "upload AS that exact path (rename on the way in)"
- `files cp <src> <dst>/` drops `<src>` into the directory by basename; `files cp <src> <dst>` (no trailing slash, single source only) treats `<dst>` as the **full target path** (rename / exact-target mode). Same for `files mv`.
- `files cp -r drive/Home/old/` (trailing `/` on a source) requires `-r` — Unix-style refusal to operate on directories without recursion. Same for `files mv`.
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

Every files API call carries `X-Authorization: <access_token>` as a header (NOT the standard `Authorization: Bearer ...`). The Factory's `refreshingTransport` injects this automatically; see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go). Do not try to call the backend via `curl` with a Bearer token — that header shape is not what the per-user files-backend expects and the request will fail.

The transport **auto-refreshes expired tokens transparently** through two paths (both detailed in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) "Automatic token refresh"):

| Verb(s) | Body shape | Refresh path |
|---------|------------|--------------|
| `ls`, `cat`, `download`, `rm`, `cp`, `mv`, `rename`, `share` (all subcommands), `repos` (all subcommands) | No body or `*bytes.Reader`/`*bytes.Buffer` (replayable) | **Reactive** — send with current token; on 401/403 call `/api/refresh` and retry once with the new token. |
| `upload` (chunk POST) | `*os.File` slice (non-replayable streaming body) | **Pro-active** — decode the JWT's `exp` before each chunk; if within 60s of expiry, refresh BEFORE handing the body to the transport. |

The pro-active path on `upload` exists because once a `*os.File` chunk is consumed by the first send, we can't replay it on a 401 — the resume probe would re-pull from the server-known offset on the next run, but the in-flight chunk would already have failed the user's command. Pre-flight rotation collapses that into a silent rotate-and-continue, even when `--parallel N>1` has multiple chunks racing the same expiry window (the `Refresher`'s in-process mutex + cross-process flock guarantee a single `/api/refresh` hit per stale token).

Stat / Range probes inside `download` and `cat` use the reactive path normally — they're cheap GETs with no body.

When the refresh leg itself fails (`/api/refresh` rejects the refresh_token), the typed `*credential.ErrTokenInvalidated` propagates through `reformatHTTPErr` / `reformatRmHTTPErr` so the user sees the canonical "run profile login" CTA directly, without a `Get "https://...":` URL prefix. Recovery rules live in `olares-shared`.

## Command cheatsheet (10 top-level verbs)

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

### `files cp [-r] <src>... <dst>`

Server-side copy across remote paths via the per-user files-backend's paste endpoint. See [`cli/cmd/ctl/files/cp.go`](cli/cmd/ctl/files/cp.go) and [`cli/internal/files/cp/`](cli/internal/files/cp/).

```bash
# Copy one file into a directory.
olares-cli files cp drive/Home/notes.md drive/Home/Documents/

# Copy with a new name on the destination side.
olares-cli files cp drive/Home/notes.md drive/Home/notes-2026.md

# Recursive directory copy.
olares-cli files cp -r drive/Home/Photos/ drive/Home/Backups/

# Multiple sources into a directory.
olares-cli files cp drive/Home/a.pdf drive/Home/b.pdf drive/Home/Archive/

# Cross-volume copy (drive → sync repo).
olares-cli files cp drive/Home/notes.md sync/<repo_id>/inbox/
```

Wire shape (one PATCH per source — the endpoint does **not** batch like `rm`):

```
PATCH /api/paste/<node>/   body: {"action": "copy", "source": "/<fileType>/<extend><sub>", "destination": "/<fileType>/<extend><sub>"}
```

Source / destination are plain UTF-8 paths (not percent-encoded) — the LarePass web app `decodeURIComponent`s before serializing, and the CLI builds them directly in that shape. Cross-volume / cross-storage-class is fully supported because the endpoint takes raw string paths and the backend dispatches by `<fileType>`.

Destination semantics (Unix-style):

| Form of `<dst>` | Meaning |
|-----------------|---------|
| ends with `/` | Drop-into-directory mode. Each `<src>`'s basename is appended; the dir / file marker on the source is preserved on the destination. |
| no trailing `/` (single source only) | Rename / exact-target mode. `<dst>` is used verbatim as the full target path. |
| no trailing `/` with **multi-source** | Rejected: `target ... must end with '/' when more than one source is given`. |

Recursion + safety rules:

- **`-r` / `-R` / `--recursive` is required for directory sources.** A trailing `/` on a `<src>` IS a directory-intent signal — without `-r` you get `... is a directory: pass -r/-R to copy it recursively`.
- **Volume-root sources are rejected**: `cp drive/Home/ ...` errors with `refusing to copy the root of drive/Home`. Same rule that protects `rm`.
- **`src == dst` is rejected** — almost always a typo.
- **Cycle detection**: copying `drive/Home/a/` into `drive/Home/a/sub/` errors with `destination ... is inside source ... (would create a cycle)`.

Async / task semantics — important:

- The PATCH returns a `task_id` after the server enqueues the copy task. **The actual byte movement happens asynchronously on the files-backend's task queue**; the CLI does NOT block until completion. The summary line prints `queued N copy task(s): <id>, <id>, ...`.
- For multi-source `cp src1 src2 src3 dst/`, the CLI sends N PATCH requests **serially** (no parallelism) so a per-call failure aborts the rest cleanly — paste tasks have no transactional rollback, and serial execution lets the user see exactly which call failed and re-run from there.
- There is currently **no built-in completion polling**. If the user needs "fail / succeed in foreground", they have to monitor server-side task progress through the LarePass web app for now.

Node selection (`--node`):

- Each PATCH carries a `<node>` URL segment. Default is the first entry from `/api/nodes/` (same default `files upload` uses) — the CLI fetches `/api/nodes/` once per invocation, but skips the round-trip when both source and destination already supply a node hint.
- For `external` and `cache` fileTypes the path's `<extend>` IS the node name, and the CLI follows the LarePass web app's `dst_node || src_node || default` cascade automatically. Drive-↔-Drive copies and `<sync>` use the default.
- `--node <name>` forces a specific node for **every** PATCH in the batch and overrides the cascade. Useful when copying across multi-master clusters where you need to pin the operation to a particular node.

Server-side rejection signal (`code: -1`):

- The endpoint returns HTTP 200 with `{"code": -1, "message": "..."}` for malformed paths (most commonly a literal backslash anywhere in `<src>` or `<dst>` — same failure mode the LarePass web app surfaces as the `files.backslash_upload` notification). The CLI surfaces this as `paste <src> → <dst>: <message>`. **Do not retry blindly on this** — the path itself needs to be fixed.

### `files mv [-r] <src>... <dst>`

Same wire endpoint as `files cp` (`PATCH /api/paste/<node>/`) but with `action: "move"` instead of `"copy"` — the server moves the source instead of duplicating it. Every flag, rule, and failure mode from `files cp` applies verbatim; the only difference is the verb.

```bash
# Rename a file in place.
olares-cli files mv drive/Home/notes.md drive/Home/notes-2026.md

# Move several files into a directory.
olares-cli files mv drive/Home/a.pdf drive/Home/b.pdf drive/Home/Archive/

# Recursive directory move.
olares-cli files mv -r drive/Home/Photos/ drive/Home/Backups/
```

`mv` is a separate cobra command (not a `cp --move` alias) so the help text stays honest about what each verb does and so `olares-cli files mv` reads the way users expect.

> **`mv` is a single-step destructive operation from the user's POV.** Even though it's just a flag flip on the wire, treat it the way you would a `mv` on local disk: confirm with the user before running it on directories, and prefer `cp` then `rm` when you want a "verify then delete" workflow.

### `files rename <remote-path> <new-name>` (alias `files rn`)

Synchronous in-place rename of a single file or directory. Different wire endpoint and semantics from `mv` — pick `rename` whenever you literally just want to change the basename. See [`cli/cmd/ctl/files/rename.go`](cli/cmd/ctl/files/rename.go) and [`cli/internal/files/rename/`](cli/internal/files/rename/).

```bash
# Rename a file in place (notes.md → 2026-Q1-notes.md, same parent dir).
olares-cli files rename drive/Home/Documents/notes.md 2026-Q1-notes.md

# Rename a directory; the trailing '/' on the source confirms directory intent.
olares-cli files rename drive/Home/Photos/ Photos-old/

# Short alias.
olares-cli files rn drive/Home/draft.txt final.txt
```

Wire shape:

```
PATCH /api/resources/<fileType>/<extend><subPath>[/]?destination=<newName>
```

Key differences vs. `mv`:

| Aspect | `mv` | `rename` |
|--------|------|----------|
| Wire endpoint | `PATCH /api/paste/<node>/` | `PATCH /api/resources/.../?destination=...` |
| Body | `{action:"move", source, destination}` | None (new name is a query param) |
| Async / sync | Async (returns `task_id`, processed by task queue) | **Synchronous** (returns immediately, change is visible right away) |
| Cross-directory | Yes (move within or across volumes) | No — basename change only, same parent dir |
| Multi-source / batch | Yes (one PATCH per source) | No — exactly one path + one new name per call |
| `<node>` URL segment | Yes (per-node routing) | No |

Validation rules (rejected client-side, before any HTTP call):

- `<new-name>` MUST NOT contain `/` or `\` — those are not basename characters; use `mv` for cross-directory moves.
- `<new-name>` MUST NOT be empty, `.`, or `..`.
- The source MUST NOT be the volume root (`drive/Home/`, `sync/<repo>/`, ...).
- `<new-name>` MUST differ from the source's current basename — same-name rename is a no-op the server would silently accept; we reject it client-side so a typo doesn't go unnoticed.

If the server replies HTTP 409, that's typically a basename collision (a sibling under the same parent already has `<new-name>`). The CLI surfaces this as `... server reported a conflict (HTTP 409); ...`. Pick a different name or `rm` the existing sibling first.

> **Use `rename` for in-place basename changes; use `mv` for moves between directories or volumes.** Picking the right verb keeps the wire shape simple and makes the user's intent legible in shell history.

### `files share <subcommand>`

Create / list / remove shares for files-backend resources. Three creation flavors (Internal cross-user, Public link, SMB Samba) plus management verbs (`list` / `get` / `rm`) plus an SMB-account roster (`smb-users`). All flavors converge on the same wire endpoint and disambiguate via the `share_type` field in the JSON body.

See [`cli/cmd/ctl/files/share.go`](cli/cmd/ctl/files/share.go), [`cli/cmd/ctl/files/share_create.go`](cli/cmd/ctl/files/share_create.go), and [`cli/internal/files/share/`](cli/internal/files/share/).

#### Three share flavors

| Flavor | Audience | Auth model | Lifetime | Output the user needs |
|--------|----------|------------|----------|-----------------------|
| `internal` | Other Olares users on the same node | Olares user identity | Persistent until `share rm` | Share id, member list |
| `public` | Anyone who has the link + password | Per-share password (auto-gen 8-char default) | Required: `--expire-days N` OR `--expire-time RFC3339` | Share id, password (printed once, NOT echoed back later), link template `<share-host>/sharable-link/<id>/` |
| `smb` | LAN clients via SMB protocol | Per-share `smb_user` / `smb_password` issued by the server, OR public-SMB ("anyone on the local network") | Persistent until `share rm` | UNC `smb_link`, `smb_user`, `smb_password` |

#### Common wire shape

```
POST /api/share/share_path/<fileType>/<extend><subPath>/
body: {name, share_type, permission, password, expire_in?, expire_time?,
       users?, public_smb?, upload_size_limit?}
```

Permission integers (`SharePermission` in the LarePass app):

| Value | Label | Meaning |
|-------|-------|---------|
| 0 | none | filter sentinel only — not a sensible create-time value |
| 1 | view | read-only |
| 2 | upload | upload-only (Public-link "drop-box" mode) |
| 3 | edit | read + write (default Public-link recipient perm; default SMB read-write) |
| 4 | admin | full control (default Internal-share owner perm) |

The CLI accepts canonical labels OR the numeric form: `--permission edit`, `--permission 3`, and `view` / `read` / `ro` / `read-only` are all aliases for `1`. See `share.ParsePermission` in [`cli/internal/files/share/share.go`](cli/internal/files/share/share.go) for the full alias list.

#### `files share internal <remote-path> [--users name:perm,...] [--permission admin]`

```bash
# Bare share record, no members yet.
olares-cli files share internal drive/Home/Backups/

# Share a single file with two members.
olares-cli files share internal drive/Home/Reports/Q1.pdf \
    --users alice:edit,bob:view
```

Two-call sequence on the wire:

```
POST /api/share/share_path/<...>/      → creates the share record (returns id)
POST /api/share/share_member/          → adds the listed users (only when --users is given)
```

If the second call fails, the share record is already on the server — the CLI surfaces the share id in the error message so the user can recover by calling member-add directly (or re-running `share internal` with the same `--users`).

`--permission` controls the OWNER's permission on the share record; default `admin` matches the LarePass web app and the only other sensible value is `edit`. Per-user permission lives in `--users`'s `name:perm` syntax (default `view` if omitted).

#### `files share public <remote-path> [--password] [--expire-days N | --expire-time RFC3339] [--upload-only] [--upload-size-limit 100M]`

```bash
# 7-day expiration, auto-generated 8-char URL-safe password.
olares-cli files share public drive/Home/Photos/ --expire-days 7

# Explicit password, 30 days, 100 MiB upload cap.
olares-cli files share public drive/Home/Photos/ \
    --password "s3cret-pw-1" --expire-days 30 \
    --upload-size-limit 100M

# Upload-only "drop box" with explicit expiration time.
olares-cli files share public drive/Home/Inbox/ --upload-only \
    --password drop --expire-time 2026-12-31T23:59:00Z
```

Required flags:

- **An expiration is mandatory.** Pass exactly one of `--expire-days N` or `--expire-time RFC3339`. The web app forces this choice; the CLI mirrors it client-side. Public links without an expiration are not supported by the backend.
- `--password` is technically optional — when omitted, the CLI generates an 8-byte URL-safe random password (`crypto/rand` → `base64.RawURLEncoding`, ≈11 chars) and prints it ONCE. **The server does not echo the password back on subsequent reads** (`share get` / `share list` show only the share record, not the cleartext password), so capture it from the create-output the first time.
- Minimum password length is 6 chars (matches the web app's `passwordLimitRule`).

Optional flags:

- `--upload-only` flips recipient permission from `edit` (read+write, default) to `upload` (drop-only — recipients can POST files but can't list / download). Useful for inbox-style intake links.
- `--upload-size-limit` accepts `100M` / `1G` / `500K` / `512` (bytes). Suffixes are 1024-based (KiB/MiB/GiB/TiB) to match the web app's `fileLimitSize()`. Omitted / zero means no cap.

The CLI prints the share id, password, expiration, and a **link template** of the form `<share-host>/sharable-link/<id>/`. The full link is built by the LarePass app from the user's hostname (`shareBaseUrl()` in `apps/.../stores/files.ts`); the CLI deliberately does NOT guess this prefix because the right value depends on subdomain layout the CLI can't observe — copy the id and let the LarePass app or the share-recipient's browser resolve the prefix.

#### `files share smb <remote-path> [--public | --users id:perm,...] [--read-only]`

```bash
# Public-SMB: anyone on the local network can mount.
olares-cli files share smb drive/Home/Movies/ --public

# Per-user SMB: must list SMB-account IDs (NOT Olares user names).
olares-cli files share smb drive/Home/Backups/ \
    --users smb-uid-1:edit,smb-uid-2:view

# Read-only override for every member.
olares-cli files share smb drive/Home/Reports/ \
    --users smb-uid-1 --read-only
```

Mutually exclusive recipient model:

| Mode | Flag | Body shape |
|------|------|-----------|
| Public-SMB | `--public` | `public_smb: true`, no `users` array |
| Per-user | `--users id:perm,...` | `public_smb: false`, `users: [{id, permission}, ...]` |

Exactly one of these flags MUST be passed — the CLI rejects an empty SMB share at parse time because a share visible to no one is rarely intentional.

Permission shape:

- SMB shares accept only `view` (1) or `edit` (3) per recipient. The CLI rejects `upload` / `admin` in `--users` parsing with a clean message.
- `--read-only` overrides every recipient's perm to `view` regardless of `--users`'s `:perm` annotations.
- Public-SMB mode forces `permission: edit` at the share-record level (matches the web app).

The IDs in `--users` are SMB-account IDs, NOT Olares user names. Discover available IDs via:

```bash
olares-cli files share smb-users list
```

Create new SMB accounts via:

```bash
olares-cli files share smb-users create <name> <password>
```

The output of `share smb` includes the `smb_link` (UNC path), `smb_user`, and `smb_password` — these are what a Finder / Explorer / `mount.cifs` client needs to mount the share.

#### Management verbs

```bash
# List all shares (default: both shared-by-me and shared-to-me).
olares-cli files share list

# Filter by direction / type / owner.
olares-cli files share list --shared-by-me
olares-cli files share list --shared-by-me=false  # only "shared with me"
olares-cli files share list --type smb,external
olares-cli files share list --owner alice,bob

# Inspect one share by id (returns shareable details for SMB shares).
olares-cli files share get <share-id>

# Remove one or more shares (single batched DELETE on the wire).
olares-cli files share rm <share-id> [<share-id>...]
```

`list` columns: `ID  TYPE  NAME  OWNER  PATH  PERMISSION  EXPIRE`. Output is sorted by `(type, id)` for stable repeated invocations — the server doesn't guarantee any ordering.

`get` renders the full share record key:value-style; SMB shares' `smb_link` / `smb_user` / `smb_password` print on their own lines so they're easy to copy.

`rm` accepts multiple IDs and joins them into a single `DELETE /api/share/share_path/?path_ids=<comma-joined>` call (atomic from the CLI's perspective).

### `files repos <subcommand>`

CRUD verbs for the per-user files-backend's catalog of **Sync (Seafile) libraries**. A "repo" (the Seafile name; LarePass UI calls them "libraries") is the unit of storage that backs every `sync/<repo_id>/<sub>` frontend path. Each repo has a stable UUID (`repo_id`) that becomes the `<extend>` segment, and a mutable display name. The CLI exposes the same five operations LarePass surfaces in its left-nav, plus `--json` for scripting. See [`cli/cmd/ctl/files/repos.go`](cli/cmd/ctl/files/repos.go) and [`cli/internal/files/repos/`](cli/internal/files/repos/).

> **Why this is its own verb.** `sync/<repo_id>/...` is the only fileType whose `<extend>` segment is a server-assigned UUID rather than a user-typed name. Without `files repos`, the `repo_id` had to be copied out of the LarePass web app every time. `files repos list` keeps the discovery loop in-CLI; the rest of the verbs let the CLI also create / rename / tear down libraries without leaving the terminal.

#### Wire shape (one endpoint, four methods)

| Verb | HTTP method + URL | Source |
|------|-------------------|--------|
| `list` (mine) | `GET /api/repos/` (no `type` param) | mirrors `fetchMineRepo` in `apps/.../api/files/v2/sync/utils.ts` |
| `list --type share-to-me` | `GET /api/repos/?type=share_to_me` | `fetchtosharedRepo` |
| `list --type shared` | `GET /api/repos/?type=shared` | `fetchsharedRepo` |
| `create <name>` | `POST /api/repos/?repoName=<name>` (no body) | `createLibrary` |
| `rename <id> <name>` | `PATCH /api/repos/?destination=<name>&repoId=<id>` (no body) | `renameRepo` |
| `rm <id>` | `DELETE /api/repos/?repoId=<id>` | `deleteRepo` |

All write verbs return the standard `{code, message, ...}` envelope; a non-zero `code` is promoted to a Go error and surfaced verbatim, matching the LarePass response interceptor at [`apps/packages/app/src/api/files/fetch.ts`](apps/packages/app/src/api/files/fetch.ts) L118-133.

#### `files repos list [--type mine|share-to-me|shared|all] [--json]`

```bash
# Default: repos you own (matches LarePass's "My Libraries" group).
olares-cli files repos list

# Repos others have shared with you, or repos you've shared out.
olares-cli files repos list --type share-to-me
olares-cli files repos list --type shared

# Concatenated fan-out across all three flavors (for one-shot
# discovery in fresh terminals). Adds a TYPE column.
olares-cli files repos list --type all

# Raw JSON for jq pipelines.
olares-cli files repos list --json
```

Default columns: `REPO_ID  NAME  PERMISSION  OWNER  SIZE  MODIFIED  ENC`. Rows are sorted (type group → name → id) so repeated invocations diff cleanly.

`PERMISSION` is `rw` / `r` / `-`. For shared variants the permission lives in the `share_permission` field on the wire — the CLI normalizes both into one column. `OWNER` falls back to the share counterparty when the row was reached via `share-to-me` / `shared`. `ENC` flags client-side encrypted libraries (not unlock-able from the CLI — see below).

#### `files repos get <repo_id>`

```bash
olares-cli files repos get <repo-id>
```

Fans out across the three flavors (`mine` → `share-to-me` → `shared`) and returns the first match. Exits non-zero with `repo not found` if the id isn't in any list — useful for scripts that branch on absence without parsing list output. Output is key:value-style with `Repo ID`, `Name`, `Owner`, `Permission`, `Encrypted`, `Size`, `Last modified`, plus a trailing `use with: olares-cli files ls sync/<repo_id>/` hint.

#### `files repos create <name> [--json]`

```bash
# Provision a new (unencrypted) library.
olares-cli files repos create "Project Alpha"
# Output:
#   created repo: Project Alpha (id: <uuid>)
#   use with: olares-cli files ls sync/<uuid>/

# Capture the id for scripting.
REPO_ID=$(olares-cli files repos create "Project Alpha" --json | jq -r .repo_id)
olares-cli files ls sync/$REPO_ID/
```

> **Encryption is NOT exposed.** The per-user files-backend's `createLibrary` endpoint accepts no password / encryption flags, and the LarePass UI has no equivalent option either. If the user needs an encrypted library they must create it from the LarePass app or directly via Seahub; once it exists the CLI can list it (`ENC=yes`), but every `upload` / `download` / `cat` / `ls` against it will fail until the user unlocks the repo via the web app.

`--json` prints the full repo record (mirroring `repos list`'s row shape) so jq pipelines can extract any field, not just `repo_id`.

#### `files repos rename <repo_id> <new_name>`

```bash
olares-cli files repos rename abc-123 "Project Alpha (archived)"
# Output (when the old name is fetchable):
#   renamed repo abc-123: "Project Alpha" -> "Project Alpha (archived)"
```

The repo's UUID is stable across renames — already-cached `sync/<repo_id>/...` frontend paths keep working. The CLI does a best-effort `Get` first to fetch the old name for the audit line; if that lookup fails, the rename still proceeds and the output simplifies to `renamed repo <id> -> "<new>"`.

#### `files repos rm <repo_id>... [--yes|-y] [--force|-f]`

```bash
# In a real terminal: lists targets (id + name) and asks y/N first.
olares-cli files repos rm abc-123

# Opt out of the prompt (required when stdin is not a TTY — scripts, CI, pipes).
olares-cli files repos rm abc-123 --yes
olares-cli files repos rm abc-123 -y
olares-cli files repos rm abc-123 -f
olares-cli files repos rm abc-123 def-456 ghi-789 -y
```

The confirmation model matches `files rm` / `cp` / `mv` safety: **default is interactive** on a TTY (`proceed with repo deletion? [y/N]:`); in a **non-interactive** context the command **refuses** until you add `-y` / `--yes` (or `--force` / `-f` as a `files rm -f`-style alias). Both `-y` and `-f` bind the same bool — either alone skips the prompt.

Destructive: removes the repo and all of its contents. The Seafile deployment may keep the data in a server-side trash window, but the CLI does **not** expose a restore verb — recovery requires the LarePass app or direct Seahub access.

Multiple ids are deleted in turn. Per-id failures are printed as `failed: <id> (<reason>)`, the loop continues on, and the command exits non-zero if any deletion failed (with all per-id errors joined via `errors.Join`).

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
| `... is a directory: pass -r/-R to copy/move it recursively` | `cp` / `mv` on a path with trailing `/` (or any directory) without `-r` | Add `-r` after confirming the user wants the whole tree |
| `target ... must end with '/' when more than one source is given` | `cp src1 src2 dst` with `dst` not ending in `/` | Add a trailing `/` to `<dst>` (drop-into-dir) or split into separate single-source `cp` calls |
| `source and destination are the same` | `cp foo foo` (typo) | Pick a real destination |
| `destination ... is inside source ... (would create a cycle)` | `cp -r drive/Home/a/ drive/Home/a/sub/` | Pick a destination outside the source tree |
| `refusing to copy/move the root of <fileType>/<extend>` | `cp drive/Home/ ...` | Volume-root copy/move is unsupported; specify a child path |
| `paste <src> → <dst>: <message>` (HTTP 200, code -1) | Server-side rejection — typically a literal backslash in the path | Fix the path; don't retry verbatim |
| `cannot resolve {node} URL segment` | Neither side has an External/Cache hint and `/api/nodes/` returned no usable default | Pass `--node <name>` explicitly |
| `queued N copy/move task(s): ...` but the file isn't visible yet | Task is queued; backend hasn't processed it | Wait briefly and `files ls` the destination; the CLI does not currently poll task completion |
| `<new-name> contains '/'` / `... contains '\\'` | `rename` was given a path-like new name | `rename` only changes the basename; use `mv` for cross-directory moves |
| `<new-name> is "."` / `is ".."` / `is empty` | `rename` was given a sentinel basename | Pick a real basename |
| `<new-name> equals the current basename` | `rename` was a no-op | If the user intended a cross-directory move, use `mv` |
| `refusing to rename the root of <fileType>/<extend>` | `rename drive/Home/ ...` | Volume roots can't be renamed; pick a child path |
| `... server reported a conflict (HTTP 409)` (rename) | A sibling with `<new-name>` already exists under the same parent | Pick a different name or `rm` the existing sibling first |
| `Public shares require an expiration` | `share public` without `--expire-days` or `--expire-time` | Pass exactly one expiration flag (Public-link shares need a TTL) |
| `--expire-days and --expire-time are mutually exclusive` | Both flags passed together | Pick one |
| `--password must be at least 6 characters` | `share public --password` shorter than 6 chars | Use a longer password (or omit `--password` to auto-generate one) |
| `--public and --users are mutually exclusive` | `share smb --public --users ...` | Pick one recipient model — Public-SMB OR a specific account list |
| `share smb requires either --public OR --users` | Neither flag set on `share smb` | Add `--public` or `--users id:perm,...` |
| `entry "...": SMB shares accept only view or edit` | `share smb --users id:upload` or `:admin` | SMB perm is binary; use `view` or `edit` |
| `share <id> created, but adding members failed: ...` | `share internal --users` post-create call failed | The share record IS on the server (id is in the message); fix the underlying error and re-run `share internal` with the same `--users`, or call member-add directly through the LarePass app |
| `share ... not found on the server` (`share get`) | Share id is wrong / already removed | List shares to confirm: `files share list` |
| `share ... server reported a conflict (HTTP 409)` (share) | Resource is already shared, or the share id is in use | Use `share list` to find the existing share id; `share rm` it first if you intended to recreate |
| `unknown repos type "..."` (`files repos list --type ...`) | Misspelled `--type` value | Use `mine`, `share-to-me`, `shared`, or `all` |
| `repos Create: empty repo name` | `files repos create ""` (or whitespace-only) | Pass a non-empty `<name>` argument |
| `repos Create: server accepted the request but did not return a repo_id` | Rare Seahub path that 200s but elides the response payload | Run `files repos list` to discover the new repo; the create likely succeeded |
| `repos Create: server rejected (code N): ...` | Server-side validation failed (duplicate name, encryption-only deployment, etc.) | Surface the message verbatim — it's the Seahub-side reason |
| `repos Rename ...: server rejected (code N): ...` | Permission denied / target name conflict on rename | Read the embedded message; if it's a name conflict, pick a different `<new_name>` |
| `repo <id>: not found in any of mine / share-to-me / shared` (`repos get`) | Wrong / removed repo id, or it's an encrypted repo the caller has no access to | Run `files repos list --type all` to confirm |
| `repos rm: refusing to delete without -y / --yes` (no TTY) | Running in CI / pipe / heredoc without a confirmation opt-out | Add `-y`, `--yes`, `-f`, or `--force` (or run from a real TTY to get the y/N prompt) |
| `repos rm: N of M failed: ...` | One or more ids hit per-id errors mid-batch | Each `failed: <id> (<reason>)` line above identifies which one; fix or retry that subset |

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

# 5. Reorganize on the server side (no local round-trip).
olares-cli files cp drive/Home/Documents/2026-Q1.pdf drive/Home/Archive/
olares-cli files mv drive/Home/Inbox/draft.md drive/Home/Notes/2026-04-draft.md
olares-cli files mv -r drive/Home/scratch/ drive/Home/Old/scratch/

# 5b. In-place basename change (synchronous, no task queue).
olares-cli files rename drive/Home/Documents/notes.md 2026-Q1-notes.md

# 6. Clean up after confirming with `ls` first.
olares-cli files ls drive/Home/Old/
olares-cli files rm -r drive/Home/Old/

# 7. Sharing workflows (orthogonal to the upload / move / cleanup loop).
olares-cli files share internal drive/Home/Reports/Q1.pdf \
    --users alice:edit,bob:view
olares-cli files share public drive/Home/Photos/ --expire-days 7
olares-cli files share smb drive/Home/Movies/ --public

# Inspect / clean up.
olares-cli files share list --shared-by-me
olares-cli files share rm <share-id>

# 8. Sync (Seafile) library lifecycle (orthogonal to drive/Home).
olares-cli files repos list                                  # discover repo_ids
REPO_ID=$(olares-cli files repos create "Project Alpha" --json | jq -r .repo_id)
olares-cli files upload ~/local-tree sync/$REPO_ID/Backups/  # use it like any other path
olares-cli files repos rename $REPO_ID "Project Alpha (archived)"
olares-cli files repos rm $REPO_ID --yes
```

When operating across multiple Olares instances, prefix each command with `--profile <olaresId>` (see `olares-shared` for the global flag) instead of flipping the persistent current pointer.

## Security rules

- **Always preview destructive operations.** Before passing `-f` to `rm` in a script, list the exact paths to the user and get explicit confirmation. The interactive `[y/N]` prompt is a safety net, not a substitute for thoughtful intent.
- **`mv` is destructive too.** It has no `[y/N]` prompt (the wire endpoint is async / fire-and-forget), so always show the user the exact `<src> → <dst>` plan before running it on directories or in scripts. When in doubt, prefer `cp` + `ls` + `rm` so the user can verify the new copy before deleting the original.
- **Local files are never overwritten implicitly.** `files download` refuses to clobber unless `--overwrite` (atomic via `.tmp`+rename) or `--resume` (append) is passed. Never recommend `--overwrite` without checking with the user.
- **Server-side overwrite on `cp` / `mv` is the backend's call, not the CLI's.** The PATCH /api/paste endpoint does not expose an "overwrite vs auto-rename" flag the way the LarePass web app's modal dialog does; the server picks its own collision behavior, which historically matches the POST-mkdir auto-rename quirk (creates `<name> (1)` instead of overwriting). Treat `cp dst/` and `mv dst/` as **non-idempotent** and confirm the destination is empty / non-conflicting before running.
- **Do not echo `<access_token>` to the terminal.** The token lives in the OS keychain for a reason; pulling it out into a shell variable for `curl` defeats that. Use the CLI commands.
- **`files upload` does NOT delete the local source** — it's a copy, not a move. If a user wants delete-after-upload semantics, they have to do it explicitly and after verifying the upload succeeded.
- **`files share public` passwords are surfaced ONCE.** The CLI prints the (auto- or user-supplied) password in the create-output; the server does NOT echo it back on `share get` / `share list`. When you script `share public`, capture the create-output before piping anywhere it might be lost — the only recovery path is `share rm` + recreate, which gives the recipient a new id.
- **SMB-share secrets (`smb_user`, `smb_password`) are returned by the create call too** — same capture-once discipline as Public-link passwords. They are queryable later via `share get <id>`, but treat the create-time output as the canonical record.
- **`share rm` removes the share record only, not the underlying resource.** A user who expects "stop sharing AND delete this folder" needs `share rm <id>` followed by `files rm -r <path>`. Confirm intent before chaining the two.
- **`share public --password ""` is rejected** (the CLI minimum is 6 chars). Don't try to bypass it by passing a single character — the wire-side check exists, you'll get an opaque server rejection a few seconds later.
- **`files repos rm` is irreversible from the CLI.** There is no `restore` verb; the Seafile-side trash window (if the deployment has one) is reachable only through the LarePass app or direct Seahub access. On a TTY the command lists id + name and asks y/N; in scripts, require `-y` after explicit user intent. Prefer `repos rename "<old> (archived)"` for "soft delete" workflows.
- **`files repos create` cannot make encrypted libraries.** Don't suggest a password flag — the CLI will refuse, and the per-user files-backend has no endpoint for it. Encrypted libraries must be created from the LarePass app; the CLI can only enumerate / rename / delete them after the fact.
