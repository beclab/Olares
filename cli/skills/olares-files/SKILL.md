---
name: olares-files
version: 1.18.0
description: "olares-cli files command tree: list (ls), upload, download, cat, edit (open in $EDITOR — TTY-required interactive verb that GETs /api/raw with a LimitReader-bounded read so an unreliable Stat.Size can't trigger an unbounded download, spawns $VISUAL/$EDITOR/vi on a temp file whose basename matches the remote so syntax highlighting works, and PUTs /api/resources when bytes changed; UPDATE-ONLY by design — there is no --create flag (the backend's PUT /api/resources/<path> handler is wired as 'replace bytes of existing file' and returns HTTP 500: file not exists for any path Stat doesn't see; create-then-edit is a two-verb shape: `files upload` to seed the file, then `files edit` to update it); supports drive/sync/cache/external ONLY — cloud drives awss3/google/dropbox/tencent are refused because the writeback PUT against /api/resources/<cloud-path> has no per-driver wire-shape signoff (only awss3's v2 utils exports a save helper; google/dropbox/tencent have no GUI plumbing to validate against — note the fetch leg via the unified /api/raw/ endpoint IS fine on cloud now), recovery is the proven download → edit-locally → upload round-trip; default 1 MiB --max-size cap fires in three places (pre-fetch via Stat, during-fetch via LimitReader, post-edit via len); two-layer text-only guard — extension deny-list (jpg/png/pdf/zip/exe/...) PLUS post-fetch NUL-byte sniff in the first 8 KiB — refuses non-text formats up front; uses HTTPClientWithoutTimeout (matching cat / download) so larger allowed edits and slow links don't trip the 30s ceiling; concurrent-delete race detection rejects Fetch-404-after-Stat-success rather than silently recreating the file; with --allow-binary / --max-size 0 / --create / --keep-temp / --content-type / --editor escape hatches), mkdir (-p / md), rm, cp, mv, rename (rn), chown (POSIX owner uid get/set, recursive), share (internal / public / smb), smb (mount / unmount external SMB shares + per-node history book — LarePass \"Connect to Server\" CLI counterpart), and Sync-repo CRUD (repos list / get / create / rename / rm) against the per-user files-backend (drive/Home, drive/Data, sync, cache, external, awss3, dropbox, google, tencent, share). Covers the 3-segment frontend path schema (<fileType>/<extend>/<subPath>), resumable chunked upload (Drive v2 protocol), Range-based resumable download, recursive directory transfer with errgroup parallelism, batch DELETE wire shape, server-side copy/move via PATCH /api/paste/<node>/ (async task_id queue, cross-volume supported), synchronous in-place rename via PATCH /api/resources/.../?destination=..., directory creation via POST /api/resources/.../<sub>/ (uniform across all namespaces; -p mode does parent-listing existence checks to side-step the auto-rename quirk), folder-share creation across the three flavors (Internal cross-user, Public link with password+expiration, SMB Samba mount) via POST /api/share/share_path/<...>/, share management (list / get / rm) plus per-flavor update verbs (set-password rolls a Public-link password via PUT /api/share/share_password/, set-members REPLACES an Internal share's member list via PUT /api/share/share_path/share_members/, set-smb REPLACES an SMB share's account list or flips public-SMB mode via POST /api/share/smb_share_member/) plus SMB-account roster (smb-users list/create), Sync (Seafile) library catalog management via /api/repos/ (GET list with type=mine|share_to_me|shared, POST create with ?repoName=, PATCH rename with ?destination=&repoId=, DELETE with ?repoId=), and four server-side / GUI quirks the user MUST know about (POST mkdir auto-renames existing dirs to 'Foo (1)'; GET single-file resource returns HTTP 500; external/<node>/ is a virtual volume-listing layer with no underlying filesystem so writes there fail / auto-rename — mkdir / cp / mv dst / upload all reject it client-side and the user must point at external/<node>/<volume>/<sub>; the system-managed first-level children directly under drive/Home/ — Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface — refuse rename / rm / mv-source client-side to mirror the LarePass GUI's `disableMenuItem` policy and protect bootstrap dirs that user apps assume exist; share-create on external/<node>/ AND cache/<node>/ is rejected client-side because the LarePass app's /Files/External and /Files/Cache root views render node / volume pickers via formatAppDataNode, not concrete datasets — the user must point at a real volume `external/<node>/<volume>/<sub>` or per-node directory `cache/<node>/<sub>`; share-create namespace gates further restrict each flavor (each row mirrors the LarePass GUI's per-driver share-menu condition): Internal allows drive / sync / external / cache; SMB allows drive / external / cache (sync excluded — Seafile-via-SMB has no working server-side path, the GUI excludes it too); Public is locked to drive ONLY (mirrors the GUI's per-driver Share-to-Public condition `event.type === DriveType.Drive || DriveType.Data`); ALL three flavors uniformly refuse the cloud namespaces awss3 / google / dropbox / tencent because the share endpoints don't grant cross-cloud-account access — recovery is `files download` followed by re-uploading to drive/Home or drive/Data, then sharing that. For sync namespace shares, every share-create command pre-resolves the repo's display name via /api/repos/ before posting the share record, so the share's `name` and the CLI's path output read like `sync/<repo-name>/<sub>  (repo <repo-id>)` instead of leaking the bare repo UUID; list output trusts the server-supplied SyncRepoName (no extra lookups), get / set-* update output falls back to a one-shot repos.Get when the server didn't echo it. POSIX file ownership is exposed via `files chown <path> [--uid <int>] [--recursive]` against /api/permission/<fileType>/<extend><subPath>/ — without --uid it does a GET and prints `uid=<int> (Root|User)`; with --uid it does a PUT with ?uid=<int>[&recursive=1] and an empty `{}` body, byte-for-byte matching the LarePass app's Permission tab in the file properties dialog (operationStore.getPermission / setPermission); namespace allow-list mirrors the GUI's `permissionInDriveType` exactly (drive/Home, drive/Data, cache/<node>/...) and rejects sync (use `files repos` for ACLs), external (GUI hides the Permission tab), and every cloud account (object stores have no POSIX uid concept) with per-namespace recovery hints; volume roots are refused to keep the blast radius bounded. External SMB shares are managed via `files smb mount|unmount|history` — `files smb mount <smb-url>` POSTs `/api/mount/<node>/?external_type=smb` with `{smbPath, user, password}` and surfaces the server's two-branch reply (code 200 → mounted at external/<node>/<entry>/; code 300 → host-only address, server returns the discovered share-paths so the user can re-run with one of them); `files smb unmount <name>` POSTs `/api/unmount/external/<node>/<name>/?external_type=smb`; `files smb history list/add/rm` plumbs the per-node \"Favorite Servers\" book through GET/PUT/DELETE `/api/smb_history/<node>/` (list shows URL + username + has-password; add stores URL with optional saved credentials so a future mount can reuse them; rm batches by URL). Password handling supports interactive (TTY only, no echo via golang.org/x/term), `--password` (echoed in shell history — convenience only), and `--password-stdin` (script-friendly, reads first stdin line) modes; `--node` defaults to the first /api/nodes/ entry, same as `files cp`. Use whenever the user mentions files / drive / Home / Data / sync / cache, uploading or downloading files, listing a remote directory, creating a directory, deleting remote files, cat-ting a remote file, editing a remote file in place via $EDITOR / vi / nano / nvim / VSCode (`files edit`), the editor cascade ($VISUAL / $EDITOR / fallback), max-size / 1 MiB cap on edit, --allow-binary / --create / --keep-temp / --max-size flags on edit, copying or moving (renaming) remote files / directories, in-place renaming, getting/setting POSIX file ownership (uid 0 root / uid 1000 user), the LarePass file properties Permission tab, sharing a folder with other users, public links with password / expiration, SMB / Samba network shares (incoming Samba export AND outgoing SMB-server mount), connecting to / mounting / unmounting external SMB servers, the LarePass \"Connect to Server\" dialog, SMB favorite servers / history, listing / creating / renaming / deleting Sync (Seafile) libraries, repo_id discovery, /api/resources, /api/raw, /api/paste, /api/share, /api/permission, /api/repos, /api/mount, /api/unmount, /api/smb_history, frontend path, or sees errors like 'Documents (1)' appearing on the server, 'volume listing layer (read-only)' on external paths, or 'system-managed Home folder reserved by LarePass' on rename / rm / mv attempts under drive/Home/."
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
| `subPath` | Path inside `extend` (root if omitted). The leading `/` is implicit. **For `external` ONLY**, `subPath` MUST contain at least one segment (the `<volume>`) for any write — `external/<node>/` is the virtual volume-listing layer (see "Server-side quirks" below). |

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

### Per-verb namespace support

All verbs except `files repos` consume frontend paths. The reachable namespaces per verb (verified against the CLI source, not just the docs) are:

| Verb | Supported namespaces | Notes |
|------|----------------------|-------|
| `ls` / `cat` / `download` / `rm` / `rename` | **all** of: `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` | These verbs hit the generic backend endpoints (`/api/resources/...`, `/api/raw/...`, `PATCH /api/resources/.../?destination=...`) with FrontendPath passed straight through. **Cloud-drive divergences:** `ls` decodes the cloud envelope's `data` array (instead of `items`) plus empty-string `mode`/`modified` fields. `cat` is fully uniform now — it always hits `GET /api/raw/<fileType>/<extend><subPath>?inline=true`, including for `awss3`/`google`/`dropbox`/`tencent` (the server-side cloud-bridge fetch is dispatched internally; older CLI builds routed those through `/drive/download_sync_stream` but that proxy is retired). See the per-verb sections for the exact wire shapes. **`rm` and `rename` reject system-managed Home children client-side** — `drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` refuse rename/delete to mirror the LarePass GUI's `disableMenuItem` policy (Server-side quirks #4); deeper paths under those dirs (e.g. `drive/Home/Pictures/Trip2024/`) remain freely editable. |
| `edit` | `drive`, `sync`, `cache`, `external` ONLY (cloud / tencent / share / internal all refused) | `GET /api/raw/<encPath>` for the pre-edit fetch, `PUT /api/resources/<encPath>` (`Content-Type: text/plain` by default) for the post-edit writeback. Mirrors the LarePass web app's per-driver `saveFile` / `updateFile` / `put` helpers in [`apps/.../api/files/v2/{drive,sync,cache,external}/utils.ts`](apps/packages/app/src/api/files/v2/drive/utils.ts) — every supported namespace funnels into the same PUT against `/api/resources/...` with the new bytes in the body. Wholesale replace, no diff/patch wire. **Cloud drives (`awss3` / `google` / `dropbox` / `tencent`) are NOT supported.** The FETCH leg is fine now — the unified `/api/raw/<fileType>/<extend><subPath>?inline=true` endpoint serves raw bytes uniformly across drive / sync / cache / external / cloud (see the [`files cat` wire-shape note](#files-cat-remote-file)) — but the WRITEBACK leg is still unverified per cloud driver: only `awss3/utils.ts` exports a `put()` helper in the LarePass GUI; `google/utils.ts`, `dropbox/utils.ts`, and the tencent driver have no save plumbing at all, so PUT-ing against `/api/resources/<cloud-path>` would hit an endpoint nobody has exercised end-to-end. The planner emits a targeted error for cloud namespaces that names the writeback gap and points at the proven recovery: `files download <cloud-path> <local>` → edit `<local>` → `files upload <local> <cloud-path>`. `share` / `internal` are also refused as cross-user / read-only views. **Three-tier size cap** (default 1 MiB, `--max-size <bytes>`, `--max-size 0` disables): (1) pre-fetch — `Stat.Size > cap` refuses the GET; (2) during-fetch — the GET body is read through `io.LimitReader(_, cap+1)` and surfaces `*edit.TooLargeError` if the server delivers more than the cap, defending against a `Stat.Size==0` listing followed by a multi-MB body; (3) post-edit — `len(newBytes) > cap` refuses the PUT, with the temp file retained for `files upload` recovery. **Text-only guard** (default-on, `--allow-binary` to disable): an extension deny-list (jpg/png/gif/heic/pdf/docx/mp4/mp3/zip/tar.gz/exe/so/sqlite/ttf/...) plus a post-Fetch NUL-byte sniff over the first 8 KiB (git/diff(1)/grep(1) heuristic). Pure-text formats with binary-looking neighbors (.svg / .html / .xml / .csv / .yaml / .ts) pass. **`--create` forces a PUT even when the editor exits without changes** — the verb's contract is "materialise this file"; a silent no-op would defeat that, so `:q!` over an empty buffer creates an empty file on the server. Re-edits of existing files use the cheaper bytes-equal short-circuit (no PUT). **Concurrent-delete race detection**: if Stat said the file existed but the subsequent Fetch returns 404, the verb refuses with `file disappeared between stat and fetch` instead of falling through to --create-empty-buffer (which would silently recreate a file someone else just deleted). **Volume roots, directory paths (trailing `/`), and `.` / `..` segments are rejected client-side** — `edit` is a per-FILE verb. Unlike `cat`, edit ALSO requires an interactive TTY (it spawns $EDITOR foreground); CI / pipe / heredoc invocations get a clean refusal with a `download` + `upload` recovery hint. **HTTP client**: `edit` uses `HTTPClientWithoutTimeout` (the same one `cat` and `download` use), NOT the 30s-capped `HTTPClient` — with `--max-size` widened or a slow link, the Fetch + PutBytes round-trip can legitimately exceed 30s. |
| `share internal` | `drive`, `sync`, `external`, `cache` (cloud refused) | `POST /api/share/share_path/<fileType>/<extend><subPath>/` with `share_type:"internal"`. **Cloud namespaces (awss3 / google / dropbox / tencent) are rejected client-side** — the share endpoints don't grant cross-cloud-account access (recovery: `files download` then re-upload to drive, then share that). **`external/<node>/` and `cache/<node>/` roots are rejected** because those layers are virtual node / volume pickers (Server-side quirks #3 and #5); deeper paths (e.g. `external/<node>/<volume>/`, `cache/<node>/<sub>/`) work fine. |
| `share smb` | `drive`, `external`, `cache` — sync AND cloud refused | Same wire shape as `share internal`, with `share_type:"smb"`. Allow-list matches the LarePass GUI's per-driver Share-to-SMB condition exactly (`DriveType.Drive\|Data\|External\|Cache`). **`sync` is rejected** — Seafile libraries have their own mount story, not Samba, so an SMB share record over a `sync/<repo>/` path has no working server-side mount path (the GUI excludes sync for the same reason); the CLI's error suggests `files share internal` as the only remaining flavor that accepts sync. **Cloud namespaces are rejected** with the same uniform cloud-rejection message as the other flavors. **`external/<node>/` and `cache/<node>/` roots are rejected** (Server-side quirks #3 and #5); deeper paths work. |
| `share public` | `drive` only — every other namespace refused | Tightest of the three flavors. Mirrors the LarePass GUI's per-driver Share-to-Public condition (`event.type === DriveType.Drive \|\| event.type === DriveType.Data`, both under the `drive` fileType). **All other namespaces are rejected at every depth**: sync / external / cache get a "Public only supports the {drive} namespace" message pointing at the other flavors that DO accept that fileType (sync → `share internal` only; external / cache → `share internal` or `share smb`); cloud namespaces get the same uniform cloud-rejection message as the other flavors. The Public namespace gate fires BEFORE the volume-listing / node-picker root checks, so `share public external/<node>/` surfaces the broader "Public only supports drive" error rather than the narrower volume-listing-layer one. |
| `mkdir` | **all** of: `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` | `POST /api/resources/<fileType>/<extend><subPath>/` (trailing '/' is the "this is a directory" marker). Uniform across every namespace — the LarePass web app's per-driver `createDir` helpers all funnel through the same endpoint. **Auto-rename on collision** (POST against an existing dir creates `Foo (1)` instead of returning 409); `-p` mode side-steps this for parents by listing each prefix's parent and skipping when the basename is already there as a directory. **`external/<node>/` is rejected client-side** — that layer is the virtual volume listing (see Server-side quirks #3); mkdir there has nowhere to land and would trip the auto-rename quirk. |
| `cp` / `mv` | **all** of the above | Goes through `PATCH /api/paste/<node>/`. Node selection cascades `--node > External/Cache extend > /api/nodes/`; cloud drives use the `/api/nodes/` default since their `<extend>` is an account, not a node. **Destinations at `external/<node>/` are rejected client-side** — same volume-listing-layer rule as `mkdir` (Server-side quirks #3); point at `external/<node>/<volume>/<sub>/` instead. **`mv` source-side rejects system-managed Home children** — `mv drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` is refused client-side because moving would unlink bootstrap dirs that user apps depend on (Server-side quirks #4); `cp` (copy) is intentionally NOT gated since it preserves the source — a `Pictures-Backup` clone via `cp -r` is fine. |
| `upload` | `drive/Home`, `drive/Data`, `sync/<repo_id>`, `cache/<node>`, `external/<node>/<volume>`, `awss3/<account>`, `google/<account>`, `dropbox/<account>` | **`tencent` is rejected** because `TencentDataAPI` in v2 uses an octet-only `/drive/direct_upload_file/<task_id>` protocol the CLI's chunk pipeline does not implement. `share`/`internal` are also rejected (they're read-only views into other namespaces). **`external/<node>/` (no `<volume>`) is rejected client-side** — that layer is the virtual volume listing (see Server-side quirks #3); the `<volume>` segment is required for upload to land on a real filesystem. Cloud drives (`awss3`/`google`/`dropbox`) are uploaded in **two stages**: stage 1 chunks bytes to Olares-internal staging via the regular Drive multipart pipeline, then stage 2 polls the server-side "Olares-staging → cloud bucket" transfer task to completion (taskId returned in the FINAL chunk's response body, polled via `/api/task/<node>/?task_id=<id>`). |
| `chown` | `drive/Home`, `drive/Data`, `cache/<node>` only | `GET /api/permission/<fileType>/<extend><subPath>/` for read, `PUT /api/permission/<fileType>/<extend><subPath>/?uid=<int>[&recursive=1]` (body `{}`) for write. Allow-list mirrors the LarePass file-properties dialog's `permissionInDriveType` array exactly (DriveType.Drive + Data + Cache). **`sync` is rejected client-side** because Seafile permissions live on the library itself (use `files repos`), **`external` is rejected** because the LarePass GUI hides the Permission tab for external mounts, and **all four cloud namespaces are rejected** because object stores have no POSIX uid concept. Volume roots (`drive/Home/`, `drive/Data/`, `cache/<node>/`) are refused — the blast radius if a typo set the uid on the entire volume is too high; pick a one-level-deeper path with `-r` if you need to fan out. |
| `smb mount` / `unmount` / `history` | n/a — the SMB-mount surface is keyed by `<node>` and `<smb-url>`, not by frontend paths. **The mount RESULT lives at `external/<node>/<entry>/` and is consumed by every other `files` verb the same way any external mount is.** See `files smb` below. |
| `repos` | n/a — this verb operates on the Sync (Seafile) library catalog (`/api/repos/...`), not on frontend paths |

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

### 3. `external/<node>/` is a virtual volume-listing layer (read-only on the wire)

Unlike `cache/<node>/` (which is a real per-node directory), `external/<node>/` does NOT have an underlying filesystem. The web app's [`ExternalDataAPI.fetchDrive`](apps/packages/app/src/api/files/v2/external/data.ts) short-circuits this layer through `formatAppDataNode` and synthesizes children from the attached-volume list (`hdd1` / `usb1` / `smb-...` mount points) — the backend at `/api/resources/external/<node>/` returns the same list. There is nowhere to write at that level: a POST mkdir, a PATCH paste destination, or a chunked upload landing on `external/<node>/` either fails server-side or trips the [auto-rename quirk](#1-post-apiresourcesdir-auto-renames-existing-directories) against a non-existent target.

Consequence baked into the CLI: `mkdir`, `cp` / `mv` destination, `upload`, **AND `share` (create)** all fail fast client-side via [`FrontendPath.IsExternalNodeRoot`](cli/cmd/ctl/files/path.go) when the target is `external/<node>/` (i.e. SubPath is just `/`). The error message points at the corrected shape (`external/<node>/<volume>/<sub>/`) so the next invocation works without server-round-trip trial-and-error. Pure reads (`ls`, `cat`, `rm`, `rename`) DO work at this layer — that's how the user discovers what volumes are attached. Share-CREATE is rejected because a share record on the volume-listing layer points at no concrete dataset (recipients would land on an empty mount-point list); share `list` / `get` / `rm` are share-id-driven and unaffected.

**The same guard extends one level deeper for `mkdir`** — depth-1 entries under `external/<node>/` ARE the mounted volumes (USB-0, SMB-..., per-disk mount-points), so creating a NEW depth-1 entry would either land as a phantom volume with no backing filesystem OR collide with an existing mount and trip the auto-rename quirk into `Foo (1)`. `mkdir.Plan` ([cli/internal/files/mkdir/mkdir.go](cli/internal/files/mkdir/mkdir.go)) refuses any `external/<node>/<single-segment>/` target client-side, and `runMkdirP` ([cli/cmd/ctl/files/mkdir.go](cli/cmd/ctl/files/mkdir.go)) refuses to auto-create a missing depth-1 intermediate in `-p` mode. Users must mount the volume via LarePass first (or target an already-mounted volume's sub-path); the depth-1 layer is GUI-managed, not files-backend-managed. `upload`, `cp`, and `mv` still allow depth-1 destinations (the corresponding gates haven't been tightened yet — they'd hit the same phantom-or-collision outcome on the server and inherit the existing IsExternalNodeRoot bare-root rejection only).

User-visible signal of the client-side guard: errors phrased `external/<node>/ is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` (writes against the bare root), `refusing to mkdir external/<node>/<X>/: depth-1 entries under external/<node>/ are mounted volumes (managed via LarePass, not via files-backend mkdir); point at a sub-path inside a real volume, e.g. external/<node>/<volume>/<X>/` (mkdir against a depth-1 target), or `refusing to share external/<node>/: this is the volume listing layer (read-only); ...` (share). Do NOT suggest creating depth-1 entries as a workaround — the constraint reflects the wire reality, and the CLI now refuses these client-side rather than letting the backend silently produce phantoms.

### 4. `drive/Home/{Pictures,Music,Movies,Downloads,Documents,Code,Cache,Data,Home,Ollama,Huggingface}` are system-managed (no rename / rm / mv-source)

Unlike the other quirks here (which document **server** behavior), this one is a **GUI-aligned client policy** that the CLI enforces to keep scripts from producing states the LarePass web app cannot reach.

The web app's [`disableMenuItem` array in `apps/packages/app/src/stores/operation.ts`](apps/packages/app/src/stores/operation.ts) — gated by `path === '/Files/Home/'` in `isDisableMenuItem` — greys out cut / copy / paste / delete / rename for these eleven names whenever the user is sitting at `/Files/Home/`. The names are LarePass's bootstrap directories under the Home volume:

| Name | Typical role |
|------|--------------|
| `Documents`, `Pictures`, `Movies`, `Downloads`, `Music` | Standard user content folders, surfaced as "shortcut" tiles in the LarePass sidebar |
| `Code` | Project workspace — referenced by the Code app integration |
| `Cache`, `Data` | App-data scratch space — system-bootstrapped by user apps |
| `Ollama`, `Huggingface` | LLM model caches — created by the model-runtime apps and consumed by them by exact name |
| `Home` | Defensive entry mirrored from the GUI array (guards against historical nested `Home/Home/` shapes) |

Note the LarePass-quirk casing: `Huggingface` is one word (not `HuggingFace`), and the names are case-sensitive across the GUI and CLI.

The backend itself does not refuse these renames / deletes — it does not encode the policy. So a CLI without this guard would happily POST `DELETE /api/resources/drive/Home/Pictures/` for a user, removing a directory that user apps look up by exact name (e.g. the model-runtime app's `Ollama` cache, or the LarePass UI's "Pictures" sidebar tile that points at this exact path). The result would be a state the GUI couldn't restore (re-creating a same-named dir doesn't republish the sidebar entry) and apps quietly breaking their fixed-name lookups.

Consequence baked into the CLI: three verbs reject these names client-side via [`FrontendPath.IsProtectedDriveHomeChild`](cli/cmd/ctl/files/path.go) (and the duplicated `protectedDriveHomeChildren` maps in `cli/internal/files/{rename,rm,cp}` that mirror it):

- **`rename`** refuses to rename `drive/Home/<protected>` to anything.
- **`rm`** refuses to delete `drive/Home/<protected>` (with or without `-r`).
- **`mv`** refuses these names AS THE SOURCE — moving would unlink the dir from `drive/Home/`. **`cp` (copy) is intentionally NOT gated** — duplicating the bytes (e.g. `cp -r drive/Home/Pictures/ drive/Home/Pictures-Backup/`) preserves the original and is a perfectly reasonable workflow even if the GUI happens to disable the menu item.

Match scope is **exact first level only**. User content nested inside (e.g. `drive/Home/Pictures/Trip2024/`, `drive/Home/Documents/notes.md`) is fully editable through every verb — the same scope the GUI uses by gating per-row on the user being at `/Files/Home/` rather than disabling the entire subtree. Other namespaces and volumes (`drive/Data/Pictures`, `sync/<repo>/Pictures`, `external/<node>/<vol>/Pictures`) are also unaffected — the policy is `drive/Home/` only.

User-visible signal of the client-side guard: errors phrased `refusing to {rename|delete|mv source} drive/Home/<name>: this is a system-managed Home folder reserved by Files; the Files GUI also disables {rename|delete|move} for {<list>} under drive/Home/.` Do NOT suggest workarounds (e.g. "use the API directly with curl") — the names are load-bearing for user apps, and bypassing the guard would be a footgun, not a feature.

### 5. `cache/<node>/` is a node-picker layer (no share-create)

This one is much narrower than #3 — `cache/<node>/` IS a real per-node directory on the wire, so `ls` / `cat` / `cp` / `mv` / `mkdir` / `upload` / `rm` / `rename` all work fine against it. The constraint is share-create-only.

The LarePass web app's [`CacheDataAPI.fetchCache`](apps/packages/app/src/api/files/v2/cache/data.ts) short-circuits the root URL `/Cache/` (note: not `/Cache/<node>/`) via `formatAppDataNode` and synthesizes children from `filesStore.nodes` — so the user sitting at `/Files/Cache/` is picking a NODE, not browsing a directory. Once a node is picked, navigation drops into `/Files/Cache/<node>/<sub>/` and the wire goes back to the regular `/api/resources/cache/<node>/...` directory listing.

Sharing a node selector does not produce a useful share record: recipients would arrive at a path that resolves to "the cache root of node X" with no concrete dataset behind it, and the LarePass UI's per-row context menu only fires on rows that map to actual files / directories — so a "share this node" affordance doesn't exist in the GUI either.

Consequence baked into the CLI: `share internal` / `share public` / `share smb` reject `cache/<node>/` (SubPath is just `/`) client-side via [`FrontendPath.IsCacheNodeRoot`](cli/cmd/ctl/files/path.go), with an error pointing at the corrected `cache/<node>/<sub>/` shape and at `files ls cache/<node>/` for discovery. **Other verbs against `cache/<node>/` are unaffected** — `ls`, `cp`, `mkdir`, etc. work normally because the per-user files-backend's `/api/resources/cache/<node>/` IS a real per-node filesystem.

User-visible signal: errors phrased `refusing to share cache/<node>/: this is the node-picker layer (no concrete dataset to share); point at a directory inside the node, e.g. cache/<node>/<sub>/`. Note the wording (`node-picker layer`) is intentionally different from external's `volume listing layer` — it tells the reader the underlying reasons differ (cache subpaths ARE shareable; external volume roots are too) so they don't infer a wider rejection than the policy enforces.

## Authentication transport

Every files API call carries `X-Authorization: <access_token>` as a header (NOT the standard `Authorization: Bearer ...`). The Factory's `refreshingTransport` injects this automatically; see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go). Do not try to call the backend via `curl` with a Bearer token — that header shape is not what the per-user files-backend expects and the request will fail.

The transport **auto-refreshes expired tokens transparently** through two paths (both detailed in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) "Automatic token refresh"):

| Verb(s) | Body shape | Refresh path |
|---------|------------|--------------|
| `ls`, `cat`, `edit`, `download`, `rm`, `cp`, `mv`, `rename`, `share` (all subcommands), `repos` (all subcommands) | No body or `*bytes.Reader`/`*bytes.Buffer` (replayable) | **Reactive** — send with current token; on 401/403 call `/api/refresh` and retry once with the new token. (`edit`'s PUT body is a `*bytes.Reader` — replayable for one-shot retry.) |
| `upload` (chunk POST) | `*os.File` slice (non-replayable streaming body) | **Pro-active** — decode the JWT's `exp` before each chunk; if within 60s of expiry, refresh BEFORE handing the body to the transport. |

The pro-active path on `upload` exists because once a `*os.File` chunk is consumed by the first send, we can't replay it on a 401 — the resume probe would re-pull from the server-known offset on the next run, but the in-flight chunk would already have failed the user's command. Pre-flight rotation collapses that into a silent rotate-and-continue, even when `--parallel N>1` has multiple chunks racing the same expiry window (the `Refresher`'s in-process mutex + cross-process flock guarantee a single `/api/refresh` hit per stale token).

Stat / Range probes inside `download` and `cat` use the reactive path normally — they're cheap GETs with no body.

When the refresh leg itself fails (`/api/refresh` rejects the refresh_token), the typed `*credential.ErrTokenInvalidated` propagates through `reformatHTTPErr` / `reformatRmHTTPErr` so the user sees the canonical "run profile login" CTA directly, without a `Get "https://...":` URL prefix. Recovery rules live in `olares-shared`.

## Command cheatsheet (14 top-level verbs)

### `files ls <path> [--json]`

List a remote directory. See [`cli/cmd/ctl/files/ls.go`](cli/cmd/ctl/files/ls.go).

```bash
olares-cli files ls drive/Home/
olares-cli files ls drive/Home/Documents
olares-cli files ls sync/<repo_id>/
olares-cli files ls drive/Home/Documents --json   # raw envelope, pretty-printed

# Cloud drives — same command, different envelope on the wire.
olares-cli files ls awss3/<account>/
olares-cli files ls google/<account>/Documents/
olares-cli files ls dropbox/<account>/
```

Default output: a one-line header (`<path>  (N dirs, M files, modified ...)`) followed by a 5-column table `MODE  SIZE  TYPE  MODIFIED  NAME`. Directories sort before files; directory names get a trailing `/`. Empty directories print `(empty)`.

`--json` prints the raw JSON envelope from the backend, useful for scripting.

**Two envelope shapes are accepted by the decoder, transparently to the user:**

| Namespace | Children field | Per-item size field | Per-item `mode` / `modified` |
| --------- | -------------- | ------------------- | ---------------------------- |
| `drive` / `sync` / `cache` / `external` / `share` | `items` | `size` (number) | numeric `mode`, RFC3339 `modified` |
| `awss3` / `google` / `dropbox` / `tencent` | `data` | `fileSize` (number; `size` is also populated on most server versions) | empty strings (`"mode":""`, `"modified":""`); the SIZE column is rendered from `fileSize` and the MODE / MODIFIED columns fall back to `d---------` / `----------` and `-` respectively |

The cloud-drive envelope ALSO omits the parent-level `numDirs` / `numFiles` / `modified` summary; the renderer's "fall back to counting items" path picks up the per-row counts so the header line stays informative. See `listingItem.UnmarshalJSON` and `listingResponse.UnmarshalJSON` in [`cli/cmd/ctl/files/ls.go`](cli/cmd/ctl/files/ls.go) for the flex-decode logic.

### `files upload <local-path> <remote-path>`

Resumable chunked upload to one of the per-user files-backend namespaces (drive, sync, cache, external) or a connected cloud drive (awss3, google, dropbox). See [`cli/cmd/ctl/files/upload.go`](cli/cmd/ctl/files/upload.go) and [`cli/internal/files/upload/`](cli/internal/files/upload/).

```bash
# Upload one file into an existing directory.
olares-cli files upload report.pdf drive/Home/Documents/

# Upload AND rename on the server.
olares-cli files upload report.pdf drive/Home/Documents/2026-Q1.pdf

# Upload a whole directory tree.
olares-cli files upload ./photos drive/Home/Backups/

# Two files in flight at a time, chunks remain sequential per file.
olares-cli files upload ./photos drive/Home/Backups/ --parallel 2

# Upload into the Data volume.
olares-cli files upload bigtar drive/Data/Backups/

# Upload into a Sync (Seafile) library.
olares-cli files upload notes.md sync/<repo_id>/Notes/

# Upload into a node-local cache (the path's <node> IS the upload node).
olares-cli files upload report.csv cache/<node>/<app>/

# Upload into attached external storage.
olares-cli files upload movie.mp4 external/<node>/hdd1/Movies/

# Upload into a connected cloud drive (S3, Google Drive, Dropbox).
olares-cli files upload backup.tar awss3/<account>/<bucket>/Backups/
olares-cli files upload doc.pdf google/<account>/Documents/
olares-cli files upload notes.md dropbox/<account>/Notes/
```

Supported destinations:

| Frontend path | Notes |
| ------------- | ----- |
| `drive/Home/<sub>` | Olares Home volume. Default upload `<node>` from `/api/nodes/`. |
| `drive/Data/<sub>` | Olares Data volume. Default upload `<node>` from `/api/nodes/`. |
| `sync/<repo_id>/<sub>` | Seafile library. Chunk POST hits `/seafhttp/upload-aj/<token>` and Seafile reads `parent_dir` as a path **inside** the repo, so the CLI sends `/<sub>/` (not `/sync/<repo_id>/<sub>/`) on the chunk form even though the API queries still use the API form. |
| `cache/<node>/<sub>` | Node-local cache. Path's `<node>` IS the upload node — the CLI **skips the `/api/nodes/` round-trip** and uses `<extend>` directly, mirroring `files cp`. |
| `external/<node>/<volume>/<sub>` | Attached external storage. Same pathNode short-circuit as `cache`. |
| `awss3/<account>/<bucket>/<sub>` | S3-compatible cloud drive. **Two-stage** upload: stage 1 chunks bytes to Olares-internal staging (multipart POST, identical to Drive); stage 2 is a server-side "Olares-staging → cloud bucket" transfer task that the backend queues and the CLI polls to completion via `/api/task/<node>/?task_id=<id>`. The stage-2 `taskId` is returned in the FINAL chunk's response body (`[{"taskId":"..."}]`) — same contract resumejs.ts onFileUploadSuccess L591-606 consumes via `Taskmanager.addTask`. |
| `google/<account>/<sub>` | Google Drive. Same two-stage pattern as awss3 (stage-1 multipart POST + stage-2 server-side transfer). |
| `dropbox/<account>/<sub>` | Dropbox. Same two-stage pattern as awss3. |

`tencent/<account>/<sub>` is the **lone unsupported namespace**: in v2, `TencentDataAPI` overrides `getFileServerUploadLink` to POST `/drive/create_direct_upload_task` and stream chunks via `/drive/direct_upload_file/<task_id>` as **octet payloads** (not multipart). The CLI's chunk pipeline doesn't speak that protocol yet, so `files upload tencent/...` fails fast with a self-describing error pointing at the protocol divergence rather than a generic "must be under <list>" message.

Wire protocol (Drive v2 / Resumable.js-compatible):

1. `GET /upload/upload-link/<node>/...` → upload session
2. `GET /upload/file-uploaded-bytes/<node>/...` → server-driven resume offset (no local progress file)
3. `POST` chunks (8 MiB default) with `Content-Range: bytes <start>-<end>/<total>` until done
4. **Cloud drives only (`awss3` / `google` / `dropbox`):** parse the FINAL chunk's response body for `[{"taskId":"<id>"}]`, then poll `GET /api/task/<node>/?task_id=<id>` every 2s until the status hits `completed` (success), `failed` (error, server-supplied `failed_reason` is surfaced), or `canceled`/`cancelled` (also surfaced as an error). The CLI keeps the per-file errgroup slot held during stage 2 so `--parallel N` stays honest. See `Client.WaitCloudTask` in `cli/internal/files/upload/api.go`.

Constraints / flags:

- **Destination MUST be one of the supported namespaces above**; tencent and unknown fileTypes fail fast.
- **`external/<node>/` (no `<volume>`) is rejected** with `... is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/`. The volume-listing layer has no underlying filesystem to land bytes on (see Server-side quirks #3).
- **Destination directory MUST already exist** — see "POST auto-renames" above.
- A trailing `/` on `<remote-path>` means "into this directory"; without one, `<remote-path>` is treated as the full target path (rename on the way in).
- `--parallel N` (default 2): per-file concurrency. **Per-file chunks remain sequential** by design — the resume probe assumes one in-flight chunk per file.
- `--chunk-size <bytes>` (default 8 MiB): align with the server's expected size; rarely needs tuning.
- `--max-retries N`: per-chunk retry budget on transient failures.
- `--node <name>`: override the upload node. Cascade: `--node` > path's `<extend>` for cache/external > first node from `/api/nodes/` for everything else (drive/sync/awss3/google/dropbox).

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

# Cloud drives: bytes come from a different endpoint, but the
# command-line ergonomics are identical.
olares-cli files cat awss3/<account>/photos/img.png > img.png
olares-cli files cat google/<account>/Documents/notes.md
olares-cli files cat dropbox/<account>/Notes/idea.md
```

Wire shape (uniform across every supported namespace — `cat` no longer per-dispatches on the FrontendPath's first segment):

| Namespace | Endpoint | Notes |
| --------- | -------- | ----- |
| `drive` / `sync` / `cache` / `external` / `share` / `awss3` / `google` / `dropbox` / `tencent` | `GET /api/raw/<fileType>/<extend><encSubPath>?inline=true` | Same endpoint the web app uses for previews and the same one `files download` uses for bytes; `inline=true` only affects `Content-Disposition` (the body is identical). Cloud drives go through the same path now — the server-side cloud-bridge fetch is dispatched internally, so the CLI no longer needs to route awss3 / google / dropbox / tencent through `/drive/download_sync_stream` the way it once did. (Earlier CLI builds keyed cloud cat off `/drive/download_sync_stream?drive=&cloud_file_path=&name=` per the LarePass `generateDownloadUrl` helpers in [`apps/.../v2/{awss3,google,dropbox}/utils.ts`](apps/packages/app/src/api/files/v2/awss3/utils.ts); that proxy is retired and the unified route is the supported wire path.) Range support depends on the underlying namespace — `download` / `cat` falls back to a full GET when the server ignores a Range header on cloud-backed handlers. |

- Binary-safe: bytes are copied verbatim, no sniffing or transformation. Pipe into `less` / `head -c` / `hexdump` as needed.
- Pre-flight `Stat` (parent listing) refuses directories early with a clear error, instead of letting the server return its terser 400. Use `files download` for directories. Stat works uniformly across all namespaces — the parent-listing decoder accepts both the Drive `items` envelope and the cloud-drive `data` envelope (see `files ls` below).

### `files edit <remote-path>`

Edit a single remote file in place via `$EDITOR`. See [`cli/cmd/ctl/files/edit.go`](cli/cmd/ctl/files/edit.go) and [`cli/internal/files/edit/`](cli/internal/files/edit/).

```bash
# Vanilla edit — pulls current bytes, opens $EDITOR, PUTs back if changed.
olares-cli files edit drive/Home/Documents/notes.md

# Override the editor for one invocation. Same precedence as `git commit`:
#   --editor flag  >  $VISUAL  >  $EDITOR  >  vi (POSIX) / notepad (Windows)
olares-cli files edit drive/Home/.config/app.yaml --editor nano
EDITOR='code --wait' olares-cli files edit drive/Home/Notes/draft.md

# Create a brand-new file: two-verb shape. `files edit` is UPDATE-only —
# the backend's PUT /api/resources/<path> handler returns
# `HTTP 500: file ... not exists` for missing paths, so the CLI no
# longer exposes a `--create` flag. `files upload` seeds the file
# (any source: a local file, `-` reading from stdin, etc.), then
# `files edit` updates it in $EDITOR.
echo "" | olares-cli files upload - drive/Home/scratch/new.txt
olares-cli files edit drive/Home/scratch/new.txt

# Override the size cap (default 1 MiB; fires pre-fetch, during-fetch, and post-edit).
olares-cli files edit drive/Home/Logs/today.log --max-size 5242880    # 5 MiB
olares-cli files edit drive/Home/Logs/today.log --max-size 0          # disable cap entirely

# Supported namespaces — same command shape.
olares-cli files edit sync/<repo_id>/Notes/2026-04.md
olares-cli files edit cache/<node>/build/config.toml
olares-cli files edit external/<node>/usb1/config.json

# Cloud drives are REFUSED at the planner. Recovery is the proven round-trip:
olares-cli files download awss3/<account>/<bucket>/config.json /tmp/cfg.json
$EDITOR /tmp/cfg.json
olares-cli files upload /tmp/cfg.json awss3/<account>/<bucket>/config.json
```

Wire shape (uniform across the four supported namespaces — same endpoint pair the LarePass web app's per-driver `saveFile` / `updateFile` / `put` helpers in [`apps/packages/app/src/api/files/v2/{drive,sync,cache,external}/utils.ts`](apps/packages/app/src/api/files/v2/drive/utils.ts) emit):

```
GET  /api/raw/<encPath>           → pull current bytes (404 is a hard error pointing at the
                                    `files upload` recovery path — `edit` is UPDATE-only,
                                    no `--create` flag; create-then-edit is a two-verb shape);
                                    bounded by io.LimitReader(_, --max-size + 1) so a server
                                    that ignores the cap or a Stat.Size==0 listing followed
                                    by a multi-MB body cannot trigger an unbounded download
PUT  /api/resources/<encPath>     Content-Type: text/plain
                                  <body: full new contents>
```

The PUT replaces the file's contents wholesale; there is no patch / diff API on the wire. Body is a replayable `*bytes.Reader`, so the auth transport's reactive refresh path covers `edit` the same way it covers `cat` (see [Authentication transport](#authentication-transport)). `edit` uses `HTTPClientWithoutTimeout` (the same client `cat` and `download` use); the 30s-capped `HTTPClient` would fail mid-edit on widened `--max-size` values or slow links — `ctx` cancellation still propagates through Ctrl-C as usual.

**Editor cascade.** Same precedence `git commit` and `crontab -e` use — argument-bearing editor strings (`'code --wait'`, `'emacs -nw'`) survive verbatim:

```
--editor flag  →  $VISUAL  →  $EDITOR  →  vi (POSIX) / notepad (Windows)
```

The editor binary is resolved against `PATH` up front, so a typo / missing binary fails BEFORE the temp file is created (friendlier than seeing the editor command fail with an opaque exec error after the user already typed). Quoting / shell expansion is intentionally NOT supported; users with truly exotic editor commands wrap them in a script and point `--editor` at that.

**Temp-file lifecycle.** A fresh `$TMPDIR/olares-files-edit-XXXX/` directory is created per invocation; the temp file's BASENAME matches the remote basename so editor-side syntax detection picks the right mode (`.md` / `.json` / `.yaml` / `.go` / ...). On clean exit / no-change exit the directory is `os.RemoveAll`'d. Three paths retain the temp file regardless of `--keep-temp` so the user can salvage typed work:

| Failure | Why retained |
| ------- | ------------ |
| Editor exits non-zero | Diagnostic — the user may have typed real content before the editor crashed. |
| Read-back from temp file fails | I/O error after the editor closed; same recovery rationale. |
| Upload fails (`PutBytes` returns non-2xx) | Bytes are NOT on the server; the user can re-PUT via `files upload <tmp> <remote>`. |
| Post-edit body exceeds `--max-size` | Same recovery rationale — the user can split the work or `files upload` past the cap. |

The retain message points at the exact temp-file path so recovery is one `cp` / `files upload` away. Pass `--keep-temp` to retain it on the no-change path too (useful for "let me diff what I almost typed" workflows).

**No-change detection.** The cobra layer compares pre- and post-edit bytes byte-for-byte (`bytes.Equal`) and skips the PUT entirely on equality. A re-edit on an existing file with `:q` / `:q!` therefore costs one GET + one editor spawn with zero network writeback; the server's modified-time stays untouched when the user just looked at the file.

**UPDATE-only verb (no `--create` flag).** `edit` only operates on files that already exist on the server. The backend's `PUT /api/resources/<path>` handler is wired as "replace the bytes of an existing file" — it returns `HTTP 500: {"code":1,"message":"file ... not exists"}` for any path Stat doesn't already see. A previous draft of this verb shipped a `--create` flag that PUT against the missing path directly; it never worked end-to-end (the wire path was incorrect) and silently produced "saved!" → 500 round-trips. The flag has since been removed. The CLI matches the LarePass GUI here: in the web app the Edit affordance only lights up on existing files; creating a new file is the upload flow (chunked POST → bucket-fetch leg, not the per-resource PUT). To create-then-edit, use two verbs:

```
echo "" | olares-cli files upload - drive/Home/scratch/new.md   # seed an empty / templated file via the chunked-upload pipeline
olares-cli files edit drive/Home/scratch/new.md                 # PUT now has an existing target
```

This shape also gives you more control over the seeding content (a Markdown frontmatter template, a YAML skeleton, etc.) than `--create`-empty-buffer ever did.

**Size cap (default 1 MiB).** `edit` is meant for text editing — configs, notes, short logs — and a hard cap protects users from accidentally streaming a multi-MB binary through their editor (vim's "swap file" warning helps but doesn't stop it). The cap fires in THREE places:

1. **Pre-fetch (Stat-driven):** Stat is the parent-listing strategy used by [`cli/internal/files/download/stat.go`](cli/internal/files/download/stat.go) (same one `cat` uses to side-step Server-side quirk #2). When Stat reports the body is larger than `--max-size`, the GET against `/api/raw/...` is skipped — no wasted multi-MB transfer just to refuse at the client.
2. **During fetch (LimitReader-driven):** `edit.Client.Fetch` wraps the response body in `io.LimitReader(_, maxBytes+1)` and returns `*edit.TooLargeError` if the read overflows. This is the safety net for the case where Stat omits / misreports `Size` (some namespaces / server versions populate `Size: 0` even for non-empty files) — without it, the pre-fetch check would silently pass and `io.ReadAll` would slurp the entire body before the post-fetch length check. The LimitReader keeps the buffer bounded at `maxBytes + 1` regardless of what the server delivers.
3. **Post-edit (length-driven):** before the PUT, we check `len(newBytes)` against the cap and refuse with a CTA pointing at `--max-size 0` and the temp-file recovery path (`files upload <tmp> <remote>`).

Override with `--max-size <bytes>`; `--max-size 0` disables ALL three layers entirely. Both directions share one knob — there's no "different cap on read vs write" mode (would be a foot-gun: GET succeeds but PUT refuses, leaving the user with edits they can't save).

**Concurrent-delete race detection.** A file that Stat said existed but Fetch reports 404 means another client / device deleted it between the two calls. The verb refuses with `file disappeared between stat and fetch (HTTP 404 on the GET); a concurrent delete is most likely — re-run to confirm, and use `files upload <local> <path>` if you want to recreate the file`. Recreating a file someone else just deleted is almost never what the caller asked for; the recovery path is the same two-verb shape as create-then-edit (`files upload` to seed, then `files edit` to update). Concurrent UPDATES still follow last-writer-wins (no ETag / If-Match support on the wire), same as the LarePass GUI.

**Text-only policy.** Two layers, both default-on, both bypassable via `--allow-binary`:

1. **Layer 1 — extension deny-list.** Fires BEFORE Stat / Fetch (so we don't waste a round-trip to refuse an obvious binary). Set is the LarePass GUI's preview classifier (image / pdf / video / audio / blob) PLUS archive / executable / DB / font extensions every editor would corrupt the moment $EDITOR touches them. Single-extension entries (`.jpg`, `.pdf`, `.mp4`, `.zip`, `.exe`, `.so`, `.sqlite3`, `.ttf`, ...) live in a `binaryExtensions` map; compound suffixes (`.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar.zst`, `.tar.lz`) are matched by `HasSuffix` because `filepath.Ext` only returns the trailing component. Matching is case-insensitive (`FOO.JPG` and `Foo.jpg` are equally refused). Pure-text formats with binary-looking neighbors are intentionally NOT in the deny-list — `.svg` / `.html` / `.xml` / `.csv` / `.yaml` / `.ts` (TypeScript) / extensionless `Dockerfile` / `Makefile` / `.env` / `README` all pass and rely on Layer 2 to backstop them.
2. **Layer 2 — NUL-byte content sniff.** Fires AFTER Fetch returns, before the temp file is written. Reads the first 8 KiB (`binarySniffLen = 8 << 10`) and refuses if a NUL (`0x00`) byte is present — same heuristic git, diff(1), and grep(1) use to detect binaries (`buffer_is_binary` in `git/diff.c`). Real text never carries a NUL; every binary container we care about (PNG / JPEG / PDF / ELF / Mach-O / ZIP / Office .docx) hits one within its first kilobyte. Empty buffers (e.g. a zero-byte file freshly seeded via `files upload`) trivially pass — there's nothing binary about an empty file. Trade-off: a NUL beyond the 8 KiB window is a false-negative by design (cheap O(1) sniff > exhaustive scan); UTF-16-encoded text with embedded NULs is a false-positive (rare in modern Unix-y workflows; recovery is `--allow-binary`).

`--allow-binary` disables BOTH layers in one knob. Don't suggest it as a routine workaround; it's the escape hatch for "I know what I'm doing" cases (UTF-16 with embedded NULs, hand-auditing a small ELF, ...). The size cap from `--max-size` still applies independently.

**Refusals (client-side, before any wire call):**

- **Non-TTY environments are rejected.** `edit` spawns $EDITOR as a foreground child process inheriting the parent's stdin/stdout/stderr — without an interactive stdin the editor would either hang waiting for input or write garbage to a non-TTY. Mirrors `rm`'s non-TTY refusal pattern. CI / pipe / heredoc users get a clean error with the suggested `files download <remote> <local>` + edit-locally + `files upload <local> <remote>` round-trip as the script-friendly alternative.
- **Volume roots are rejected** with a CTA that includes a sample file path (`drive/Home/notes.md`, `sync/<repo>/notes.md`, etc.) — `edit` is a per-FILE verb and pointing it at `drive/Home/` is almost certainly a typo.
- **Directory paths (trailing `/`) are rejected** with a "use `files ls <path>` to list it" hint — same per-FILE rationale.
- **`.` / `..` segments ANYWHERE in `<remote-path>` are rejected** — same path-traversal blacklist `mkdir` / `rename` enforce, applied here BEFORE `path.Clean` silently collapses them. Without this guard `edit drive/Home/foo/./bar` would land bytes on `drive/Home/foo/bar` (a different file than the user typed); with it, the offending segment is named in the error.
- **Cloud drives (`awss3` / `google` / `dropbox` / `tencent`) are REFUSED at the planner.** The FETCH leg is fine now — the unified `/api/raw/<fileType>/<extend><subPath>?inline=true` endpoint serves raw bytes on cloud namespaces (see the [`files cat` wire-shape note](#files-cat-remote-file)) — but the WRITEBACK leg is the remaining gap: only `awss3/utils.ts` has a `put()` helper at all in the GUI; `google/utils.ts`, `dropbox/utils.ts`, and the tencent driver have no save-related plumbing, so PUT-ing against `/api/resources/<cloud-path>` for those drivers would hit an endpoint nobody has exercised end-to-end. Until the PUT shape is wire-verified per cloud driver, the planner returns a targeted error pointing at the proven recovery: `files download <cloud-path> <local>` → edit `<local>` → `files upload <local> <cloud-path>`. Both legs of the recovery now go through the unified `/api/raw/` (download) and namespace-aware multi-stage upload pipelines, so the round-trip is byte-exact.
- **`share` / `internal` are also refused** as cross-user / read-only views in the LarePass UX with no documented save surface.

**Common error patterns:**

| Symptom | Cause | Fix |
| ------- | ----- | --- |
| `refusing to spawn an editor without a TTY (no interactive stdin)` | Running in CI / pipe / heredoc | Use `files download` + edit locally + `files upload`. |
| `refusing to edit <path>: extension looks like a non-text format` | Extension is on the deny-list (jpg / pdf / mp4 / zip / so / ...) | If you really meant it, pass `--allow-binary`. Otherwise this is the safety net working — `edit` is for text. |
| `refusing to edit <path>: content looks binary (NUL byte in the first 8192 bytes)` | Layer-2 sniff caught a binary file with an innocuous extension | Same recovery — `--allow-binary` to opt out, or `files download` for the binary. |
| `edit <path>: remote size ... exceeds --max-size ...` | Pre-edit Stat-driven cap fired | Override with `--max-size <bytes>` or `--max-size 0`; or switch to `files cat <path> | head` if you only need to peek. |
| `edit <path>: post-edit size ... exceeds --max-size ...; temp file retained at ...` | Editor produced a buffer larger than the cap | Re-run with a wider `--max-size` and the same temp file via `files upload <tmp> <remote>`. |
| `edit <path>: not found on the server (HTTP 404); `files edit` only updates existing files. To create a new file, use `files upload <local> <path>` first ...` | Typo in path OR genuinely new file. `edit` is UPDATE-only — no `--create` flag (it was removed; see the "UPDATE-only verb" note above for the wire-shape rationale). | Read the CTA — either fix the typo, or use the two-verb shape: `files upload` to seed, then re-run `files edit`. |
| `edit <path>: file disappeared between stat and fetch (HTTP 404 on the GET); a concurrent delete is most likely` | Another client / device deleted the file in the millisecond window between Stat and Fetch | Re-run to confirm. If the file is genuinely gone and you want it back, use `files upload <local> <path>` to recreate it, then re-run `files edit`. |
| `edit <path>: cloud-drive namespace "<awss3\|google\|dropbox\|tencent>" is not supported end-to-end` | Tried to edit a cloud-drive path | Use the documented round-trip: `files download <path> <local>` → edit locally → `files upload <local> <path>`. The error itself includes the recovery snippet. |
| `edit <path>: fetched body exceeds --max-size <cap> (Stat reported 0; either the listing's size field is unreliable or the file was concurrently appended between stat and fetch)` | Server's listing put `Size: 0` but the actual body is larger; the `LimitReader` inside Fetch caught it before the buffer grew unbounded | Widen `--max-size <bytes>` if the file is legitimately bigger; pass `--max-size 0` to disable the cap; or use `files cat <path> | head -c <bytes>` to peek at the head if you only need a sample. |
| `editor "<bin>" not found in PATH` | Bad `--editor` flag / `$EDITOR` value | Fix the env or `--editor` arg; the cascade fallback (`vi` / `notepad`) is exercised when neither is set. |
| `editor "<bin>" exited non-zero; temp file retained at <path>` | Editor crashed (vim panic, VSCode launch failure, ...) | Recover from the temp file; never trust the remote in that state. |
| `<path> is a directory: edit only works on files (use \`files ls <path>\` to list it)` | Pointed at a directory | Pick a file path inside it. |

### `files mkdir [-p] <remote-path>` (alias: `md`)

Create a remote directory. See [`cli/cmd/ctl/files/mkdir.go`](cli/cmd/ctl/files/mkdir.go) and [`cli/internal/files/mkdir/`](cli/internal/files/mkdir/).

```bash
# Create one directory whose parent already exists.
olares-cli files mkdir drive/Home/Documents/Backups

# `-p` creates missing intermediate directories; safe to re-run.
olares-cli files mkdir -p drive/Home/A/B/C/

# Works uniformly across all 3-segment namespaces.
olares-cli files mkdir -p sync/<repo_id>/notes/2026/Q2
olares-cli files mkdir -p cache/<node>/scratch
olares-cli files mkdir -p external/<node>/usb1/Backups
olares-cli files mkdir -p awss3/<account>/Backups/2026
olares-cli files mkdir -p google/<account>/Drafts
olares-cli files mkdir -p dropbox/<account>/Notes
```

Wire shape (uniform across drive / sync / cache / external / awss3 / google / dropbox / tencent — same endpoint the LarePass web app's per-driver `createDir` helpers emit):

```
POST /api/resources/<fileType>/<extend><subPath>/
```

The trailing '/' on the URL is what the backend uses to disambiguate "create directory" from "create empty file" (`isDir ? '/' : ''` in [`v2/common/utils.ts`](apps/packages/app/src/api/files/v2/common/utils.ts)). Body is empty.

**Important — the auto-rename quirk.** On the current files-backend, POST against an existing directory does NOT return 409; it silently creates `Foo (1)` next to the original. That has two practical consequences:

| Mode | What we do | What you should do |
| ---- | ---------- | ------------------ |
| no `-p` | Single POST against the leaf path. We can't detect collisions cheaply (there's no 409 to react to), so the call may have created `Foo (1)` instead of `Foo`. | Run `olares-cli files ls <parent>/` afterwards if the dir might have already existed. |
| `-p` | List each prefix's parent first; skip prefixes that already exist as directories; POST only the missing ones. One extra GET per existing prefix, in exchange for correctness. | Just trust it — `-p` is safe to re-run on an already-fully-created tree (it'll skip everything and exit clean). |

Refusals:

- The volume root (`drive/Home/`, `sync/<repo>/`, etc.) is rejected — those always exist; trying to "create" them would just trigger the auto-rename on the extend folder, which is never what you want.
- **`external/<node>/` is rejected with a customized message** pointing at `external/<node>/<volume>/<sub>/` — the volume-listing layer has no underlying filesystem (see Server-side quirks #3). The generic "pick a subdirectory name (e.g. external/<node>/NewFolder)" hint other namespaces emit would be misleading here, since `NewFolder` would still land at the volume-list level.
- **`.` / `..` segments ANYWHERE in `<remote-path>` are rejected** — path-traversal blacklist. This is a pre-`path.Clean` check on the RAW user input, so `mkdir drive/Home/.`, `mkdir drive/Home/foo/./bar`, `mkdir drive/Home/foo/../bar`, and `mkdir drive/Home/../../etc` all surface a targeted `path segment "." (or "..") is a reserved name (...path-traversal blacklist)` error instead of silently collapsing to the parent volume root or rewriting under the user's feet. Empty segments (`drive/Home//foo`) collapse normally per POSIX (the cobra layer's `path.Clean` keeps that behavior); only the literal `.` / `..` segments trip the blacklist.
- In `-p` mode, a prefix that exists as a NON-directory (e.g. you asked for `mkdir -p Foo/Bar` but `Foo` is a file) errors out instead of letting the auto-rename quirk silently create a `Foo (1)/` sibling.

### `files rm [-r] [-f] <remote-path>...`

Delete one or more remote files / directories. See [`cli/cmd/ctl/files/rm.go`](cli/cmd/ctl/files/rm.go) and [`cli/internal/files/rm/`](cli/internal/files/rm/).

```bash
# Delete one file (no -r → FILE form, wire dirent `/old.pdf`).
olares-cli files rm drive/Home/Documents/old.pdf

# Recursively remove a directory — both forms produce the same
# wire request (`/2024/` dirent); pick whichever reads naturally.
olares-cli files rm -r drive/Home/Backups/2024
olares-cli files rm -r drive/Home/Backups/2024/

# Multiple folders, no prompt (scripts).
olares-cli files rm -rf drive/Home/junk drive/Home/scratch/
```

Wire shape (one batch DELETE per parent dir):

```
DELETE /api/resources/<encParentDir>/   body: {"dirents": ["<name1>", "<name2>", ...]}
```

Targets sharing a parent collapse into a single request (matches the LarePass web app's `batchDeleteFileItems`). Targets across different parents send one request each, sorted by `fileType + extend + parent` for stable output.

Flags / rules:

- `-r` / `-R` / `--recursive`: required for directories. A trailing `/` on a target IS a directory-intent signal and triggers the same check (so `files rm drive/Home/Foo/` errors without `-r` even if `Foo` is technically empty). **`-r` also FORCES every target in the same invocation to be sent as a directory dirent (`/<name>/` on the wire) regardless of whether the path string ended in `/`** — that's the Unix `rm -r foo` semantic and the only one that reliably routes the request through the server's POSIX directory-removal path. Practical consequence: `files rm` (no `-r`) is the FILE form (wire dirent `/<name>`), `files rm -r` is the FOLDER form (wire dirent `/<name>/`); don't mix files and folders in a single `-r` call. If you have a file to delete alongside a directory tree, run two `files rm` invocations — one with `-r` for the folder, one without for the file.
- `-f` / `--force`: skip the y/N prompt. **In a non-TTY context (CI, piped stdin) the command refuses without `--force`** rather than guessing.
- Without `-f`: prints "will delete N entries in M batches" with the full list, then prompts `[y/N]`.
- Removing the root of a volume (`drive/Home/`, `sync/<repo>/`, ...) is rejected by the planner: `refusing to delete the root of <fileType>/<extend>`.
- **System-managed Home children are rejected**: `files rm [-r] drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` errors with `refusing to delete drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...`. The names match the LarePass app's `disableMenuItem` array (Server-side quirks #4); their casing is fixed (`Huggingface` is one word). Children INSIDE these dirs (e.g. `drive/Home/Pictures/Trip2024/`) are user content and remain freely deletable. To clear out a protected folder, delete its contents (`files rm -r drive/Home/Pictures/<entry>`); the folder itself stays.

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
- **`external/<node>/` destinations are rejected**: `cp src external/<node>/` errors with `... destination external/<node>/ is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/`. The volume-listing layer has no underlying filesystem (see Server-side quirks #3). Volume roots like `external/<node>/hdd1/` ARE valid destinations — the strict rule is only on the bare-node form.
- **`src == dst` is NOT rejected client-side** — the LarePass web app doesn't enforce this either. For a **file** target (`cp drive/Home/foo.pdf drive/Home/foo.pdf`) the backend auto-renames into `foo (1).pdf` (same POST-mkdir auto-rename quirk users already work with from `mkdir`); for a **directory** target the cycle check below catches it. `mv foo foo` on a file is a server-side no-op.
- **Cycle detection**: copying `drive/Home/a/` into `drive/Home/a/sub/` (or `drive/Home/a/` into itself, since dir-onto-same-dir would create an infinitely-recursing tree) errors with `destination ... is inside source ... (would create a cycle)`.
- **`mv` source-side rejects system-managed Home children**: `mv drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface} ...` errors with `refusing to mv source drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...`. Moving would unlink bootstrap dirs that user apps depend on (Server-side quirks #4). **`cp` (copy) is intentionally NOT gated** — `cp -r drive/Home/Pictures/ drive/Home/Pictures-Backup/` is fine because the source is preserved. Children inside these dirs (e.g. `drive/Home/Pictures/Trip2024/`) are user content and remain freely movable.

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

> **System-managed Home children refuse `mv` as the source** (see `cp` section above). `mv drive/Home/Pictures ...` and the other ten LarePass-protected names will fail client-side with a self-describing error. If the goal is to relocate the *contents*, mv the children (`mv drive/Home/Pictures/* drive/Home/Backups/`) — the protected folder will stay where it is. If the goal is a renamed clone, use `cp -r` instead, then optionally clear the originals via `rm` (which the policy also blocks for the protected folder itself, but allows on its children).

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
- `<remote-path>` MUST NOT contain `.` or `..` segments anywhere (path-traversal blacklist, pre-`path.Clean` check on the RAW input). `rename drive/Home/foo/../bar new` would otherwise silently collapse to `rename drive/Home/bar new` — different entry than the user typed — so the CLI surfaces a targeted `path segment "." (or "..") is a reserved name (...path-traversal blacklist)` error instead.
- The source MUST NOT be the volume root (`drive/Home/`, `sync/<repo>/`, ...).
- `<new-name>` MUST differ from the source's current basename — same-name rename is a no-op the server would silently accept; we reject it client-side so a typo doesn't go unnoticed.
- **System-managed Home children are rejected**: `files rename drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface} ...` errors with `refusing to rename drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...`. These names mirror the LarePass app's `disableMenuItem` array exactly (case-sensitive — `Huggingface` is one word) and protect bootstrap dirs that user apps look up by name (Server-side quirks #4). Children INSIDE these dirs (e.g. `drive/Home/Pictures/Album1/`) are user content and remain freely renamable.

If the server replies HTTP 409, that's typically a basename collision (a sibling under the same parent already has `<new-name>`). The CLI surfaces this as `... server reported a conflict (HTTP 409); ...`. Pick a different name or `rm` the existing sibling first.

> **Use `rename` for in-place basename changes; use `mv` for moves between directories or volumes.** Picking the right verb keeps the wire shape simple and makes the user's intent legible in shell history.

### `files chown <remote-path> [--uid <int>] [--recursive | -r]`

Read or write the POSIX owner UID of a single file or directory. CLI counterpart of the LarePass app's "Permission" tab in the file properties dialog ([`apps/packages/app/src/components/files/prompts/InfoDialog.vue`](apps/packages/app/src/components/files/prompts/InfoDialog.vue), `permissionInDriveType` + `permissionOption` + `onSubmit`). The dialog exposes exactly two presets — `Root`=0 and `User`=1000 — plus a "Recursive" toggle; the CLI accepts any non-negative integer for flexibility but those two are what the GUI plumbs the user through.

See [`cli/cmd/ctl/files/chown.go`](cli/cmd/ctl/files/chown.go) and [`cli/internal/files/permission/`](cli/internal/files/permission/).

#### Two modes (no subcommands — flag presence picks the verb)

| Invocation | Wire | Behavior |
|------------|------|----------|
| `files chown <path>` | `GET /api/permission/<fileType>/<extend><subPath>/` | Print the current uid as `drive/Home/foo  uid=1000 (User)`. The `(Root)` / `(User)` annotation is added only for the LarePass preset values; other UIDs render bare. |
| `files chown <path> --uid <int>` | `PUT /api/permission/<fileType>/<extend><subPath>/?uid=<int>` body `{}` | Replace the uid on the named entry only. |
| `files chown <path> --uid <int> -r` | `PUT /api/permission/<fileType>/<extend><subPath>/?uid=<int>&recursive=1` body `{}` | Replace the uid on the entry AND every descendant. Mirrors the LarePass GUI's "Recursive" checkbox; the GUI sends literal `recursive=1` (not `true`), so the CLI does too. |

The body is unconditionally `{}` (an explicit empty JSON object); the actual parameters travel in the query string. `setPermission` in [`apps/packages/app/src/stores/operation.ts`](apps/packages/app/src/stores/operation.ts) does the same byte-for-byte.

#### Namespace gate (allow-list of two)

| `<fileType>` | Accepted? | Why |
|--------------|-----------|-----|
| `drive` (Home, Data) | ✓ | LarePass's `permissionInDriveType` lists `DriveType.Drive` + `DriveType.Data`, both of which wire to `fileType="drive"`. |
| `cache` (per-node) | ✓ | LarePass lists `DriveType.Cache`. |
| `sync` | ✗ — error suggests `files repos` | Seafile permissions live on the library itself (per-user library ACLs); POSIX uid is not the right concept. |
| `external` | ✗ — error mentions the GUI hides the Permission tab | The LarePass file-properties dialog hides the Permission tab when `currentFile.driveType` is External, so the wire surface there is not part of this contract. |
| `awss3` / `dropbox` / `google` / `tencent` | ✗ — error reads "object stores have no POSIX uid concept" | Cloud accounts are object stores; ownership is meaningless at this layer. |

#### Volume-root refusal

Chowning the root of a namespace is refused client-side regardless of the gate above:

```
$ olares-cli files chown drive/Home/
Error: refusing to chown the root of drive/Home; pick a child path (use -r to fan out across the volume)

$ olares-cli files chown cache/<node>/
Error: refusing to chown the root of cache/<node>; pick a child path (use -r to fan out across the volume)
```

This is purely a CLI-layer guard (the wire would accept it). The blast radius if a typo set every entry under `drive/Home/` to uid 0 is severe enough to be worth one extra keystroke; recover by pointing at a subdirectory and using `-r`.

#### Other refusals

- `--recursive` without `--uid` → `--recursive only applies when setting a uid; pass --uid <int> to use it, or drop --recursive to GET the current uid`. The GUI's recursive toggle only makes sense alongside Submit; surfacing the inconsistency early prevents a script that meant to PUT but forgot `--uid` from silently no-op'ing as a GET.
- `--uid` without an integer → `--uid "<input>" is not an integer (LarePass GUI uses 0 for Root and 1000 for User)`.
- `--uid -1` (negative) → `--uid must be non-negative (got -1)`. The wire casts to uint server-side, so -1 would silently become a huge UID; reject it client-side.

#### Examples

```bash
# Inspect the current owner of a file.
olares-cli files chown drive/Home/Documents/foo.pdf
# → drive/Home/Documents/foo.pdf  uid=1000 (User)

# Hand a single file to root.
olares-cli files chown drive/Home/Documents/foo.pdf --uid 0
# → set uid: drive/Home/Documents/foo.pdf → uid=0 (Root)
#   ✓ drive/Home/Documents/foo.pdf  uid=0 (Root)

# Hand an entire subtree back to the default user (recursive).
olares-cli files chown drive/Home/Pictures/Trip2024/ --uid 1000 -r
# → set uid (recursive): drive/Home/Pictures/Trip2024/ → uid=1000 (User)
#   ✓ drive/Home/Pictures/Trip2024/  uid=1000 (User) (recursive)

# Cache namespace works the same way.
olares-cli files chown cache/<node>/scratch/build/ --uid 1000 -r

# drive/Data is also supported (separate volume from drive/Home).
olares-cli files chown drive/Data/builds/out.bin --uid 0
```

`olares-cli files chmod` is **not** an alias — POSIX `chmod` changes mode bits, which the per-user files-backend doesn't expose at this layer. If you reach for `chmod`, swap it for `chown` and the wire shape will line up with the GUI.

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

#### Sync repo name resolution (auto-applied to every share verb)

When a share path is in the `sync` namespace, the `<extend>` segment is the Seafile library's UUID (e.g. `sync/b7ffab7f-3ceb-4e36-aeb7-74d958ad0a7a/`). UUIDs are noisy in CLI output and meaningless to share recipients, so every share verb that touches sync paths swaps the bare UUID for the repo's display name where possible.

Mechanics:

- **Create flavors** (`share internal` / `share public` / `share smb`) call [`lookupSyncRepoName`](cli/cmd/ctl/files/share.go) (one-shot `/api/repos/` list, internally `Get` walks `mine` → `share_to_me` → `shared`) BEFORE posting the share record. The resolved name becomes the share's wire-level `name` field (instead of the legacy "last segment / extend" derivation, which would yield the UUID for repo-root shares), and the CLI's "created share" output renders the path as `sync/<repo-name>/<sub>/  (repo <repo-id>)` so the user sees both labels at a glance.
- **Single-record reads** (`share get`) and **the three update verbs** (`set-password` / `set-members` / `set-smb`) use [`resolveShareDisplayPath`](cli/cmd/ctl/files/share.go), which trusts the server's `SyncRepoName` echo when present and falls back to a one-shot `lookupSyncRepoName` only when the field is empty. Same display format.
- **List output** (`share list`) uses [`formatSharePathLine`](cli/cmd/ctl/files/share.go) with `override=""`, i.e. it consumes the server-supplied `SyncRepoName` only and **never** falls back to a per-row `repos.Get` — that would be an N+1 explosion on large lists.

Failure semantics:

- `lookupSyncRepoName` is fire-and-forget on the error path: any error / missing repo / unreachable `/api/repos/` returns `""`, and the caller falls back to displaying the UUID. The share-create / share-display call NEVER fails because of a repo lookup hiccup.
- For repo-root shares with no resolved name, the legacy "extend as name" behavior wins — the share record's `name` becomes the UUID. Less pretty but still functional.

Display format detail:

- Resolved: `sync/Project Alpha/Reports/Q1/  (repo b7ffab7f-3ceb-4e36-aeb7-74d958ad0a7a)`
- Unresolved: `sync/b7ffab7f-3ceb-4e36-aeb7-74d958ad0a7a/Reports/Q1/`

The parenthesised `(repo <id>)` is preserved on resolved sync paths so users can still cross-reference `repos rename` / `repos rm` / `repos get` against the same UUID. Non-sync namespaces (`drive` / `external` / `cache` / cloud) render unchanged — the resolution helpers no-op.

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

Namespace gate (Public-only — distinct from the all-flavors gates further down):

Public is locked to the `drive` namespace. Every other fileType is refused, at every depth (root, volume root, deeper subpath). This mirrors the LarePass GUI's per-driver Share-to-Public condition in `apps/packages/app/src/stores/operation.ts`:

```js
condition: (event) =>
    event.type === DriveType.Drive ||
    event.type === DriveType.Data
```

Both `DriveType.Drive` (drive/Home) and `DriveType.Data` (drive/Data) live under the same `drive` fileType on the wire, so the CLI's allow-list is `{drive}` — the path's `<extend>` (`Home` vs `Data`) is not gated client-side, and the server is the authoritative gate for whether a Drive subpath is shareable.

Refusals by namespace (with the recovery hint the CLI surfaces in the error):

- **`sync/<repo>/<...>`** → `files share public` only supports drive; use `files share internal` for the same path (SMB also refuses sync — Seafile-via-SMB has no working server-side mount), or copy the data into `drive/Home` / `drive/Data` first.
- **`external/<node>/<...>`** → same Public-only message. External volumes are mounted under `external/<node>/<volume>/` and a public link is awkward there in practice (volumes detach, mount points move between reboots, the link recipient can't tell which volume the link points at); the recovery is the same — `share internal` / `share smb` for in-place sharing, or copy into Drive first.
- **`cache/<node>/<...>`** → same Public-only message. Per-node cache holds transient app data unsuitable for long-lived public links.
- **`awss3/<...>` / `google/<...>` / `dropbox/<...>` / `tencent/<...>`** → cloud rejection (see the next section); the message is the uniform "no cloud sharing" one, NOT the per-flavor Public-only one. Recovery is `files download` + re-upload to Drive + share that.

The Public namespace gate fires BEFORE the volume-listing / node-picker root checks below, so `share public external/<node>/` surfaces the broader "Public only supports drive" error rather than the narrower volume-listing-layer one — same final answer (the share isn't created), but the more useful explanation wins. Internal / SMB still go through their own narrower gate (next sub-section) plus the volume-listing / node-picker root checks.

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

Namespace gate (SMB-only — distinct from the all-flavors cloud gate further down):

SMB allows only `drive`, `external`, `cache`. The allow-list mirrors the LarePass GUI's per-driver Share-to-SMB condition:

```js
condition: (event) =>
    event.type === DriveType.Drive ||
    event.type === DriveType.Data ||
    event.type === DriveType.External ||
    event.type === DriveType.Cache
```

`DriveType.Drive` (drive/Home) and `DriveType.Data` (drive/Data) both live under the `drive` fileType on the wire, so the CLI's allow-list collapses to `{drive, external, cache}`.

- **`sync/<repo>/<...>` is rejected**. Seafile libraries have their own mount story (the web app's `/Files/Sync/` view uses Seafile's HTTP API, not Samba); an SMB share record pointing at a sync path would have no working server-side path to mount. Use `files share internal` for sync paths instead — Internal cross-user access is the supported flow for Seafile content. (Public is also out for sync — see the Public-only gate above.)
- **`external/<node>/<...>` and `cache/<node>/<...>`** are allowed at every depth (subject to the volume-listing / node-picker root rejections below — those still fire for `external/<node>/` and `cache/<node>/` exactly).
- **Cloud namespaces** are rejected by the all-flavors cloud gate covered in the next section.

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

#### Cloud rejection (applies to all three create flavors)

Every share-create flavor (`share internal` / `share public` / `share smb`) refuses every cloud namespace — `awss3`, `google`, `dropbox`, `tencent` — at every depth. Implemented as a uniform branch in `validateShareNamespace` ([`cli/cmd/ctl/files/share.go`](cli/cmd/ctl/files/share.go)).

Why it's uniform across flavors: the share endpoints record a `<fileType>/<extend><subPath>` reference into the share table, but the cloud namespaces' `<extend>` is a per-Olares-user account credential, not a node mount. Recipients of an internal / SMB / Public share have no credential to read the cloud-side path, so the share record would resolve to "404 / 403 / no access" for everyone except the original owner — making the share record meaningless. The CLI fast-fails up front rather than letting the user create a guaranteed-broken share.

Error template (the `<flavor>` slot is `internal` / `public` / `smb`):

```
refusing to create a <flavor> share for <path>: cloud namespaces
(awss3 / google / dropbox / tencent) do not support sharing through
`files share` — the share endpoints don't grant cross-cloud-account
access, and the resulting share record would point at a path that no
other Olares user has the credential to read. If you need to share
cloud-backed data, download it first (`files download`) and re-upload
it to drive/Home or drive/Data, then share that.
```

The recovery step is the same regardless of flavor — copy the bytes through Drive first, then share the Drive path. There is no in-place fallback; the cloud rejection is a hard policy.

#### Path refusals (apply to share internal / share smb only — Public's namespace gate already handled these)

The conversion from user path to share Target lives in `frontendPathToShareTarget` ([`cli/cmd/ctl/files/share.go`](cli/cmd/ctl/files/share.go)) and is shared by all three flavors. After the per-flavor namespace gate (above) and the all-flavors cloud rejection, two more checks fire for the remaining cases:

- **`external/<node>/` is rejected**: `refusing to share external/<node>/: this is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/`. The volume-listing layer has no real filesystem (Server-side quirks #3); pass a volume root or any subpath under it (e.g. `external/<node>/hdd1/` or `external/<node>/hdd1/Backups/`). Public never reaches this branch (the namespace gate already refused all of `external/*`); only `share internal` / `share smb` actually surface this message.
- **`cache/<node>/` is rejected**: `refusing to share cache/<node>/: this is the node-picker layer (no concrete dataset to share); point at a directory inside the node, e.g. cache/<node>/<sub>/`. The LarePass app's `/Files/Cache/` view is a node selector, not a directory listing (Server-side quirks #5). Subpaths under the node — including `cache/<node>/<app>/` — work fine. Public never reaches this branch either.
- **Volume roots ARE allowed** for `drive/Home/`, `drive/Data/`, and `sync/<repo>/` (subject to each flavor's own namespace allow-list — `share public sync/<repo>/` is still refused by the Public-only gate). Sharing `drive/Home/` itself is unusual but legitimate; the server is the authoritative gate for whether the resulting share is meaningful — the CLI doesn't second-guess.

To recover from either rejection, use `files ls` on the rejected path to find the navigable children:

```bash
olares-cli files ls external/<node>/   # → hdd1 / usb1 / smb-... mount points
olares-cli files ls cache/<node>/      # → per-node directory listing
```

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

#### Per-flavor update verbs (`set-password` / `set-members` / `set-smb`)

Three verbs for editing an existing share without re-creating it. Each one targets a single flavor and rejects mismatched share types client-side via [`requireShareType`](cli/cmd/ctl/files/share.go), pointing the user at the correct verb in the error message — `requireShareType` runs after a `Query` (single GET), so the type check happens before the mutation hits the wire and the server never sees the wrong-flavor request.

Common mismatch error template (anchored by tests in [`share_update_test.go`](cli/cmd/ctl/files/share_update_test.go)):

```
refusing to <verb> share <id>: the share is <friendly-type> (wire type "<wire-type>"),
not <expected-friendly-type>; use the matching update verb instead — `share set-password`
for public shares, `share set-members` for internal shares, `share set-smb` for SMB shares
```

Note: the friendly-name translation for Public is load-bearing — the wire `share_type` discriminator for Public is the historically confusing `"external"` string, so the error always says "public" rather than "external" to keep users from hunting for a non-existent `share external` command.

##### `files share set-password <share-id> [--password STR]`

Roll the access password of a Public-link share. Wire shape:

```
PUT /api/share/share_password/
    body: {path_id: "<share-id>", password: "<new-password>"}
```

Mirrors LarePass's "Reset Password" modal at `apps/.../components/files/share/ShareResetPassword.vue` (calls `share.resetPassword(path_id, password)`). The share id stays stable across the reset — recipients keep the same `/sharable-link/<id>/` URL but need the new password the next time they open it.

Password rules — same as `share public` create:

- `--password` is optional; when omitted the CLI generates an 8-byte URL-safe random password (`crypto/rand` → `base64.RawURLEncoding`, ≈11 chars) and prints it ONCE. The server does not echo passwords back on subsequent `share get` / `share list`, so capture the first-print output.
- Minimum length is 6 characters (LarePass's `passwordLimitRule`).

Type gate: only Public shares accept this verb. `set-password` against an Internal or SMB share fails with the standard share-type-mismatch error.

```bash
# Auto-generated random password.
olares-cli files share set-password <share-id>

# Explicit password.
olares-cli files share set-password <share-id> --password "n3w-pw-1"
```

##### `files share set-members <share-id> (--users name:perm,... | --clear)`

REPLACE the entire member list of an Internal share. Wire shape:

```
PUT /api/share/share_path/share_members/
    body: {path_id: "<share-id>",
           share_members: [{share_member: "<name>", permission: <int>}, ...]}
```

Mirrors LarePass's "Edit Permissions" modal at `apps/.../components/files/share/Internal/internal.ts` (calls `share.updateInternalShareMembers` when `internalShareId` already exists, around L137-147 there).

**REPLACES, not appends** — every existing member NOT listed in `--users` is dropped. The endpoint has no append / single-member-PATCH variant on the wire, and the LarePass app uses the same `PUT` for additive flows (it carries every existing member through plus the new one). For an additive update from the CLI, list every existing member plus the new one in `--users`.

Flag rules:

- `--users` and `--clear` are mutually exclusive; exactly one MUST be passed (the no-flag form is rejected to guard against shell-quoting accidents wiping a member list).
- `--users` format mirrors `share internal --users`: `name1:perm1,name2:perm2,name3` — perm defaults to `view` when omitted.
- `--users` accepts permissions: `view` / `upload` / `edit` / `admin` (or `1..4`). `parseShareMembers` is shared with `share internal` so the parsing rules are identical between create and update.
- `--clear` drops every member; the share record stays but becomes private to its owner until the next `set-members`.
- `--users " ,, "` (all-whitespace, no usable entries) is rejected with a hint to use `--clear` for intentional empties.

Type gate: only Internal shares accept this verb. SMB shares use a different account model (SMB-account IDs, not Olares user names) and have their own `set-smb` verb.

```bash
# Replace with two users.
olares-cli files share set-members <share-id> --users alice:edit,bob:view

# Promote bob from view to admin (carry alice through unchanged).
olares-cli files share set-members <share-id> --users alice:edit,bob:admin

# Drop every member.
olares-cli files share set-members <share-id> --clear
```

##### `files share set-smb <share-id> (--public | --users smb-id:perm,...) [--read-only]`

REPLACE the SMB-account list of an SMB share, or flip it to public-SMB mode. Wire shape:

```
POST /api/share/smb_share_member/
    body: {path_id: "<share-id>",
           users: [{id: "<smb-account-id>", permission: <int>}, ...],
           public_smb: <bool>}
```

Mirrors LarePass's SMB edit-mode at `apps/.../components/files/share/SMB/smb.ts` L79-104 (calls `share.updateSMBShareMember(path_id, users, public_smb)`).

Flag rules — same shape as `share smb` create, so muscle memory carries over:

- `--public` and `--users` are mutually exclusive; exactly one MUST be passed.
- `--public` flips the share to public-SMB; the CLI sends `users: []` (typed empty slice, NOT `null`) so the wire shape matches LarePass exactly.
- `--users` entries are SMB-account IDs (NOT Olares user names — list them via `share smb-users list`); per-entry perm is `view` / `edit` (or `1` / `3`) — `upload` and `admin` are rejected because the SMB endpoint doesn't grant those.
- `--read-only` forces every entry's perm to `view`, regardless of each entry's `:perm` annotation. Public-SMB mode is independent of `--read-only`.
- Shared parser with `share smb` create (`parseSMBUsers`) — single source of truth for the format, no drift between create and update.

Type gate: only SMB shares accept this verb. Public / Internal shares fail with the share-type-mismatch error.

```bash
# Switch to public-SMB.
olares-cli files share set-smb <share-id> --public

# Replace member list with two SMB users.
olares-cli files share set-smb <share-id> --users smb-uid-1:edit,smb-uid-2:view

# Demote everyone to read-only.
olares-cli files share set-smb <share-id> --users smb-uid-1,smb-uid-2 --read-only
```

### `files smb <subcommand>`

CLI counterpart of the LarePass web app's "Connect to Server" modal — mount external SMB shares into the per-user files-backend's `external/<node>/...` namespace, and manage the per-node "Favorite Servers" history book. See [`cli/cmd/ctl/files/smb.go`](cli/cmd/ctl/files/smb.go) and [`cli/internal/files/smbmount/smbmount.go`](cli/internal/files/smbmount/smbmount.go); the GUI references are [`apps/.../components/files/smb/ConnectServerStep1.vue`](apps/packages/app/src/components/files/smb/ConnectServerStep1.vue) (favorites + URL entry), [`ConnectServerStep3.vue`](apps/packages/app/src/components/files/smb/ConnectServerStep3.vue) (credential entry), [`apps/.../stores/files.ts`](apps/packages/app/src/stores/files.ts) `mountSmbInExternal`, and [`apps/.../stores/operation.ts`](apps/packages/app/src/stores/operation.ts) `unmount`.

> **Why this is its own verb tree, not part of `files share`.** `files share smb` creates an OUTGOING Samba share (the user's data exposed over SMB to other clients). `files smb mount` is the INCOMING direction — the per-user files-backend mounts a remote SMB server into `external/<node>/<entry>/` so the user can browse it as if it were local. They share the "SMB" word and nothing else.

#### Wire shape (three endpoints, five methods)

| Verb | HTTP method + URL | Body | Source |
|------|-------------------|------|--------|
| `mount` | `POST /api/mount/[<node>/]?external_type=smb` | `{smbPath, user, password}` | `mountSmbInExternal` (stores/files.ts L1263-L1302) |
| `unmount` | `POST /api/unmount/external/<node>/<name>/?external_type=smb` | `{}` | `unmount` (stores/operation.ts L844-L887) |
| `history list` | `GET /api/smb_history/<node>/` | — | `getFavoriteList` (ConnectServerStep1.vue L113-L122) |
| `history add` (upsert) | `PUT /api/smb_history/<node>/` | `[{url, username?, password?}]` | `saveFavorite` (ConnectServerStep1.vue L124-L134) |
| `history rm` | `DELETE /api/smb_history/<node>/` | `[{url}]` | `removeFavorite` (ConnectServerStep1.vue L136-L140) |

`{<node>}` is conditionally present in the mount URL — when the per-user files-backend has no clustered nodes, the LarePass app drops it entirely (`/api/mount/?external_type=smb`); the CLI replicates that branch byte-for-byte. **For `unmount` and the three `history` verbs the `<node>` segment is mandatory** (the URL shape is fixed).

#### Mount-reply envelope: code 200 vs code 300

The mount endpoint has TWO valid success-shape replies:

| Code | Meaning | CLI behavior |
|------|---------|--------------|
| `200` | Mounted successfully; the share is now visible at `external/<node>/<entry>/` | Print `✓ mounted; the share is now visible at external/<node>/<entry>/` plus a hint to confirm via `files ls external/<node>/`; exit 0 |
| `300` | The supplied `smbPath` was a host-only address (e.g. `//host`); `data` is `[{path:"//host/share[/sub]"}, ...]` — the list of shares the server discovered | Print the candidate paths (one per line, OR JSON when `--json`); exit non-zero with `mount returned a multi-share list (code 300); re-run with one of the paths above` so a script can detect the case |

Any other code is surfaced as `server rejected (code N): <message>` — a 200/HTTP envelope-error, NOT a `*HTTPError`. This matches the LarePass GUI flow: code 300 pops up `ConnectServerPath.vue` (a chooser dialog) and the user re-runs; code-other shows a toast and bails.

#### `files smb mount <smb-url> [-u <user>] [-p <password> | --password-stdin] [--no-history] [--node <node>] [--json]`

`<smb-url>` MUST start with `//` (the SMB convention). Four credential modes, in strict priority order — the first match wins, later modes only kick in when their predecessors are absent:

| Mode | Trigger | Behavior |
|------|---------|----------|
| Explicit literal | `-p <password>` | Used as-is. Convenient for one-offs but **echoed in shell history** — never use this in scripts. |
| Stdin pipe | `--password-stdin` | Reads the first stdin line, strips trailing CR/LF. Mutually exclusive with `-p`. The script-friendly form: `printf '%s' "$SMB_PASSWORD" \| olares-cli files smb mount ... --password-stdin`. Empty stdin is rejected. |
| Saved favorite | `<smb-url>` matches an entry from `files smb history list` | Autofills missing flags from the favorite — `username` fills `-u` when not passed, and `password` fills the prompt when `-p` / `--password-stdin` are both absent. **Cross-account safety:** if `-u <user>` is passed AND it disagrees with the favorite's username, the CLI emits a one-line `note: history has saved credentials for user "..." but -u "..." was passed; using flags as-is` and falls through — one account's saved password is never lent to a different account. Mirrors LarePass's "Connect to Server" autofill (`ConnectServerStep1.vue` populates the same fields when a row is clicked). Disable with `--no-history`. |
| Interactive | none of the above | **Two-step prompt.** When `-u` is ALSO missing (and history didn't fill it in) the CLI first prints `SMB username (empty for anonymous): ` — echoed, since the username isn't sensitive — and reads one line; pressing Enter without typing is the explicit "anonymous mount" gesture. Then `SMB password: ` is read via `golang.org/x/term` (no echo). Requires a real TTY at the password step — failing the TTY check produces `stdin is not a terminal — pipe a password with --password-stdin or pass --password explicitly`. **Why the user prompt was added:** previously a flag-less `mount //host/share` against a URL not in history would silently default to `(user=(anonymous))` and prompt only for a password, then almost always reject with `server rejected (code 500): Incorrect username or password` because the share actually wanted a real account. The two-step prompt makes the account choice visible. Anonymous shares (empty user + empty password) are accepted at the wire layer; if the server rejects them, the error surfaces verbatim. |

When autofill fires, mount prints a `· using saved credentials from SMB history (user=<name>)` line right before the `mount: <url> @ <node>` progress line, so the user can tell at a glance which path the credentials came from. A history endpoint failure (network blip / 401 on `/api/smb_history/`) is a **soft failure**: a single `note: SMB history unavailable (...); proceeding without autofill` is emitted and the command falls through to the flag- / prompt-driven path — the mount itself is never blocked by a flaky favorites lookup.

`-u` is OPTIONAL — anonymous shares (where the server allows un-authenticated `guest` access) work without it; the progress line then reads `(user=(anonymous))`. Most modern SMB targets need a real username, so the CLI does NOT auto-supply one (unless a matching saved favorite provides it via the autofill flow above).

`--no-history` opts out of the autofill flow entirely; the saved favorite is ignored even when it matches. Use this when the saved credentials are known stale (e.g. the user just rotated the password and wants to type the new one without first `smb history rm` + `smb history add`).

`--node` is OPTIONAL — without it the CLI calls `/api/nodes/` and uses `nodes[0].Name` (same default cascade `files cp` and `files upload` use). Pass `--node <name>` only when the auto-detected first node isn't the one you want.

```bash
# Mount a specific share (server already knows which one).
olares-cli files smb mount //host.local/Public -u alice -p s3cret

# CI-friendly: pipe the password from a secret store.
printf '%s' "$SMB_PASSWORD" | \
    olares-cli files smb mount //host.local/Public -u alice --password-stdin

# Interactive (TTY-only, no echo) with -u already supplied.
olares-cli files smb mount //host.local/Public -u alice
# → SMB password: ********

# Interactive — no flags at all, URL not in history. Both the
# username and the password are prompted, in that order.
olares-cli files smb mount //host.local/Public
# → SMB username (empty for anonymous): alice
# → SMB password: ********

# Interactive — anonymous share (just hit Enter at the username prompt).
olares-cli files smb mount //host.local/Public
# → SMB username (empty for anonymous):
# → SMB password: ********  (or empty if the share allows it)

# Saved-favorite autofill — no flags needed when the URL was
# previously saved with credentials via `smb history add`.
olares-cli files smb history add //host.local/Public -u alice -p s3cret
olares-cli files smb mount //host.local/Public
# →  · using saved credentials from SMB history (user=alice)
# →  mount: //host.local/Public @ <node> (user=alice)
# →  ✓ mounted ...

# Force a fresh prompt even though a favorite is saved (e.g. after
# rotating the password).
olares-cli files smb mount //host.local/Public --no-history -u alice
# → SMB password: ******** (typed new value)

# Server-side share discovery — pass the host alone, get a list,
# re-run with the chosen share path.
olares-cli files smb mount //host.local
# → server returned 3 candidate share path(s) — pick one and re-run mount:
#     //host.local/Public
#     //host.local/Movies
#     //host.local/Backups
# (exit code 1)

olares-cli files smb mount //host.local/Public

# JSON form of the code-300 list (script consumption).
olares-cli files smb mount //host.local --json
# → {"code":300,"paths":["//host.local/Public","//host.local/Movies",...]}
```

After a successful mount, every other `files` verb works against the resulting `external/<node>/<entry>/` path the same way it does for any other namespace — `files ls external/<node>/` shows the new entry, `files cp` works against subpaths, and so on.

#### `files smb unmount <name> [--node <node>]`

`<name>` is the entry name AS IT APPEARS in `files ls external/<node>/` — typically something like `smb-host-share`. **The CLI rejects names that contain `/` or `\` client-side** because the URL shape is `/api/unmount/external/<node>/<name>/...` — passing a multi-segment path as `<name>` would silently 404. Discover the right name first:

```bash
olares-cli files ls external/main/
# → smb-host-share/    drwxr-xr-x  ...
#   hdd1/              drwxr-xr-x  ...

olares-cli files smb unmount smb-host-share --node main
# → unmount: external/main/smb-host-share
#   ✓ unmounted external/main/smb-host-share
```

`unmount` is **not interactive** — there's no `[y/N]` prompt — because losing a mount is recoverable (just re-run `mount`); preserve the audit trail in the user's terminal output and that's it.

`--node` defaults to the first `/api/nodes/` entry, same as `mount`.

#### `files smb history <subcommand>`

Per-node SMB favorites — the same "Favorite Servers" list the LarePass app keeps in its "Connect to Server" dialog. Each entry carries a `url` plus optional saved credentials (`username` + `password`); a history entry with credentials makes a future `mount` one keystroke faster (the user / app picks the URL, credentials autofill). Three subverbs:

##### `files smb history list [--node <node>] [--json]`

Default columns: `URL  USERNAME  HAS-PASSWORD`. `HAS-PASSWORD` is `yes` / `no` — the saved password itself is **never printed** in default mode; pass `--json` if you need the raw fields (e.g. to migrate a favorites list across instances). Each `--json` line is one entry, including `url`, `username`, `password`, `timestamp`.

```bash
olares-cli files smb history list
# URL                       USERNAME      HAS-PASSWORD
# //host.local/Public       alice         yes
# //backup.local/Archive    (anonymous)   no

# Full record for migration / scripting.
olares-cli files smb history list --json
```

The wire reply tolerates both shapes the GUI has used historically — bare array (`[{url:..., ...}, ...]`) AND envelope (`{code, data:[...]}`) — so the CLI keeps working across server-side refactors.

##### `files smb history add <smb-url> [-u <user>] [-p <password> | --password-stdin] [--node <node>]`

Add or update (upsert) an entry by URL. The wire is a PUT — the server merges by `url`, replacing existing rows in place. Three flag combinations:

| Combination | Resulting entry |
|-------------|-----------------|
| no `-u` / no `-p` | URL-only favorite. Subsequent `mount` calls prompt for credentials. |
| `-u` only | URL + username, no saved password. `mount` will still prompt for the password. |
| `-u` + (`-p` or `--password-stdin`) | URL + saved credentials. `mount` can pull both halves from the favorites list (LarePass auto-uses them). |

**`-p` / `--password-stdin` without `-u` is rejected client-side** with `--password / --password-stdin requires --user; SMB auth needs both halves` — a saved password without a username is unusable for SMB auth and almost always indicates a typo.

```bash
# URL-only favorite.
olares-cli files smb history add //host.local/Public

# With a saved username (mount still prompts for password).
olares-cli files smb history add //host.local/Public -u alice

# Full saved credentials (one-step mount later).
olares-cli files smb history add //host.local/Public -u alice -p s3cret
# → ✓ saved favorite //host.local/Public on node main (user=alice, password saved)

# CI-friendly secret pipe.
printf '%s' "$SMB_PASSWORD" | \
    olares-cli files smb history add //host.local/Public -u alice --password-stdin
```

##### `files smb history rm <smb-url>... [--node <node>]`

Remove one or more entries by URL. Multiple URLs are batched into a single DELETE request. Each URL must start with `//` (same shape rule as `mount` and `add`); whitespace-only / empty entries are skipped (the all-empty case is rejected with `no SMB urls given` so a typo can't silently no-op).

```bash
olares-cli files smb history rm //host.local/Public
olares-cli files smb history rm //a/Public //b/Movies
# → ✓ removed favorite //a/Public on node main
#   ✓ removed favorite //b/Movies on node main
```

The DELETE is unconditional — the favorites list is small and easily re-built, so there's no `[y/N]` prompt. Pair with `history list` first if you want a safety net.

#### Operational notes

- **Saved passwords are stored server-side as plaintext** (the LarePass GUI does the same — there's no client-side encryption). Treat the favorites list the same way you'd treat any other credential blob: don't expose it via `--json` to a logger / chat / public artifact, and prefer URL-only or username-only entries when the convenience is small.
- **Unmount BEFORE deleting an SMB favorite if the favorite was the only record of the connection** — once the entry is out of the history list, the user has to manually re-type the URL / credentials to remount.
- **Mount entries persist across CLI invocations and reboots** until explicitly unmounted (or until the underlying SMB server goes unreachable for long enough that the per-user files-backend GCs the entry — out of scope for the CLI).
- **`<entry>` (the name visible in `external/<node>/`) is server-chosen** — the CLI does NOT let you pick it. If you mount the same SMB share twice the entry name will collide; unmount the old one first.

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

> **Reserved names.** Empty / `.` / `..` are rejected client-side with `repos Create: repo name "..." is a path-traversal segment, not a real name` — same blacklist `files mkdir` and `files rename` enforce. Pick a non-reserved label; everything else (spaces, unicode, parens, etc.) is fair game.

`--json` prints the full repo record (mirroring `repos list`'s row shape) so jq pipelines can extract any field, not just `repo_id`.

#### `files repos rename <repo_id> <new_name>`

```bash
olares-cli files repos rename abc-123 "Project Alpha (archived)"
# Output (when the old name is fetchable):
#   renamed repo abc-123: "Project Alpha" -> "Project Alpha (archived)"
```

The repo's UUID is stable across renames — already-cached `sync/<repo_id>/...` frontend paths keep working. The CLI does a best-effort `Get` first to fetch the old name for the audit line; if that lookup fails, the rename still proceeds and the output simplifies to `renamed repo <id> -> "<new>"`.

> **Reserved names.** Same `.` / `..` blacklist `repos create` enforces — the rename path can't be used as an end-around to land a reserved name, error wording is `repos Rename: new name "..." is a path-traversal segment, not a real name`.

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
| `upload destination must be under drive/Home, drive/Data, sync/<repo_id>, cache/<node>, external/<node>/<volume>, awss3/<account>, google/<account>, or dropbox/<account>` | Tried to upload to an unsupported `<fileType>/<extend>` (e.g. `share/...`, `internal/...`) | Move the target under one of the supported namespaces |
| `upload to tencent COS is not supported by this verb: it uses the octet /drive/direct_upload_file/<task_id> protocol ...` | Tried `files upload <local> tencent/<account>/...` | Tencent COS uses a different upload protocol (octet streams) that the CLI's chunk pipeline does not implement; use the LarePass web app for tencent uploads, or upload to a different namespace |
| `cloud transfer task <id> failed: <reason>` | Stage-2 (Olares-staging → cloud bucket) failed; chunks landed on Olares but the cloud-bridge worker reported an error | Stage-2 failures usually come from the cloud side: bad credentials, bucket quota / permissions, target path collision in the cloud, or transient cloud-API outages. Inspect the surfaced `<reason>`; verify the connected cloud account is still authorized in LarePass; retry the upload (chunked stage-1 will resume; stage-2 starts fresh once the last chunk re-arms it) |
| `cloud transfer task <id> was cancelled server-side` | Someone (or another component) cancelled the stage-2 task while the CLI was waiting | Re-run the same `files upload` — stage 1 will resume from where it left off and a fresh stage-2 task gets queued |
| `query cloud task <id> on node <node>: ...` | Stage-2 polling endpoint flapped (typically a redeploy or transient 5xx on the task service) | The error preserves the underlying HTTP status; if it's transient, retry; if it's persistent, the task service may be unhealthy — check Olares cluster status |
| `Documents (1)` (or similar) appearing on the server after upload | Older CLI version triggered the POST-mkdir auto-rename quirk | Upgrade to a CLI version that has the pre-mkdir removal fix |
| `Foo (1)` appearing after `files mkdir Foo` | The leaf already existed; the backend silently auto-renamed instead of returning 409 | This is the documented auto-rename behavior. Use `olares-cli files ls <parent>/` to confirm; remove the `(1)` suffix manually if needed. Use `files mkdir -p` next time — `-p` mode lists each prefix's parent first and skips existing ones, side-stepping the quirk. |
| `mkdir -p ... already exists but is NOT a directory` | A prefix in your `-p` path is a file (e.g. `mkdir -p Foo/Bar/` but `Foo` is a file) | Pick a different path; the auto-rename quirk would otherwise create a `Foo (1)/` sibling silently. |
| `mkdir <path>: parent directory does not exist (HTTP 404); pass -p to create missing intermediates` | Tried `files mkdir A/B/C` where `A/B` doesn't exist | Add `-p` to create the missing intermediates in one shot. |
| `refusing to mkdir the root of <fileType>/<extend>` | Tried `files mkdir drive/Home/` (or another volume root) | Pick a subdirectory name (e.g. `drive/Home/NewFolder`). |
| `refusing to mkdir at external/<node>/: this is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Tried `files mkdir external/<node>/` or `files mkdir external/<node>/NewVolume/` (creating an entry at the volume-list level) | `external/<node>/` is the virtual volume listing — there is no real filesystem there (see Server-side quirks #3). Point `mkdir` at a directory inside an existing volume, e.g. `external/<node>/hdd1/NewFolder/`. Use `files ls external/<node>/` first to discover the attached volumes. |
| `cp/mv: destination external/<node>/ is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Tried `cp src external/<node>/` (drop-into-dir at the volume-list level) | Same root cause as the mkdir variant. Use `files ls external/<node>/` to find the right volume, then re-run with `external/<node>/<volume>/...` as the destination. Volume roots like `external/<node>/hdd1/` are valid destinations. |
| `upload destination external/<node>/ is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Tried `files upload <local> external/<node>/` (no `<volume>` segment) | Same root cause. Add the `<volume>` segment so the upload targets a real attached volume, e.g. `external/<node>/hdd1/Backups/`. |
| `--overwrite and --resume are mutually exclusive` | Passed both download flags | Pick one |
| `refusing to delete without --force in a non-interactive context (no TTY)` | `files rm` from a script with no TTY | Add `-f` after **explicitly listing the targets to the user first** |
| `refusing to delete the root of <fileType>/<extend>` | Tried `files rm -r drive/Home/` (or another volume root) | The CLI does not support volume-root deletion; remove children individually |
| `cat ... is a directory` | `files cat` on a path that resolves to a directory | Use `files download <path>/` instead |
| `files cat awss3/<account>/...` returns `HTTP 502 nginx` (or an opaque cloud-bridge error) | Older CLI version routed cloud-drive `cat` through the now-retired `/drive/download_sync_stream` proxy | Upgrade to a CLI version that uses the unified `GET /api/raw/<fileType>/<extend><subPath>?inline=true` endpoint for every namespace including cloud (this branch); the server-side cloud-bridge fetch is dispatched internally. |
| `files ls awss3/<account>/` shows `(empty)` despite the bucket having files | Older CLI version only decoded the `items` envelope; cloud-drive listings live under `data` | Upgrade to a CLI version that decodes both envelope shapes (this branch). |
| `HTTP 500` against `/api/resources/.../<filename>` (no trailing slash) | Hit the backend's single-file List quirk directly (e.g. via curl) | Don't bypass the CLI; the CLI uses parent-listing Stat for a reason |
| `... is a directory: pass -r/-R to copy/move it recursively` | `cp` / `mv` on a path with trailing `/` (or any directory) without `-r` | Add `-r` after confirming the user wants the whole tree |
| `target ... must end with '/' when more than one source is given` | `cp src1 src2 dst` with `dst` not ending in `/` | Add a trailing `/` to `<dst>` (drop-into-dir) or split into separate single-source `cp` calls |
| `destination ... is inside source ... (would create a cycle)` | `cp -r drive/Home/a/ drive/Home/a/sub/` (or `cp -r drive/Home/a/ drive/Home/a/`, i.e. dir → same dir) | Pick a destination outside the source tree. Note: `cp foo foo` on a **file** is no longer rejected client-side — the backend auto-renames into `foo (1)` (same POST-mkdir quirk); `mv foo foo` is a server-side no-op. |
| `refusing to copy/move the root of <fileType>/<extend>` | `cp drive/Home/ ...` | Volume-root copy/move is unsupported; specify a child path |
| `paste <src> → <dst>: <message>` (HTTP 200, code -1) | Server-side rejection — typically a literal backslash in the path | Fix the path; don't retry verbatim |
| `cannot resolve {node} URL segment` | Neither side has an External/Cache hint and `/api/nodes/` returned no usable default | Pass `--node <name>` explicitly |
| `queued N copy/move task(s): ...` but the file isn't visible yet | Task is queued; backend hasn't processed it | Wait briefly and `files ls` the destination; the CLI does not currently poll task completion |
| `<new-name> contains '/'` / `... contains '\\'` | `rename` was given a path-like new name | `rename` only changes the basename; use `mv` for cross-directory moves |
| `<new-name> is "."` / `is ".."` / `is empty` | `rename` was given a sentinel basename | Pick a real basename |
| `mkdir: path segment "." (or "..") is a reserved name (... path-traversal blacklist)` | `files mkdir` was given a path containing a `.` / `..` segment (`drive/Home/.`, `drive/Home/foo/./bar`, `drive/Home/foo/../bar`, ...) | Pick a name that does not match `.` / `..` in any segment — the check runs on RAW input before `path.Clean`, so silently-collapsing forms like `foo/../bar` no longer rewrite to `bar` |
| `rename: path segment "." (or "..") is a reserved name (... path-traversal blacklist)` | `files rename` was given a `<remote-path>` containing a `.` / `..` segment | Same fix as the mkdir case — type the literal source path; `path.Clean`-style traversals are not accepted on the rename source either |
| `<new-name> equals the current basename` | `rename` was a no-op | If the user intended a cross-directory move, use `mv` |
| `refusing to rename the root of <fileType>/<extend>` | `rename drive/Home/ ...` | Volume roots can't be renamed; pick a child path |
| `refusing to rename drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...` | Tried `files rename drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` (the LarePass app's `disableMenuItem` set) | These names are LarePass-bootstrapped Home folders that user apps look up by exact name (Server-side quirks #4) — the Files GUI also greys out rename for them. Rename a child instead (e.g. `drive/Home/Pictures/<album>`), or copy the contents into a new sibling (`cp -r drive/Home/Pictures/ drive/Home/PicturesArchive/`). |
| `refusing to delete drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...` | Tried `files rm [-r] drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` | Same root cause as the rename variant (Server-side quirks #4). Children INSIDE the protected folder are freely deletable — `files rm -r drive/Home/Pictures/<entry>` works, the protected folder itself stays. |
| `refusing to mv source drive/Home/<name>: this is a system-managed Home folder reserved by Files; ...` | Tried `files mv drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface} ...` | Moving would unlink bootstrap dirs that user apps depend on (Server-side quirks #4). For a renamed clone, use `cp` instead (`cp -r drive/Home/Pictures/ drive/Home/Pictures-Backup/`); for relocating contents, mv the children (`mv drive/Home/Pictures/* drive/Home/Backups/`) — the protected folder will remain in place. |
| `... server reported a conflict (HTTP 409)` (rename) | A sibling with `<new-name>` already exists under the same parent | Pick a different name or `rm` the existing sibling first |
| `refusing to share external/<node>/: this is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Tried `files share internal\|public\|smb external/<node>/` (the volume-listing layer has no real filesystem behind it) | Same root cause as the mkdir / cp / upload variants of this rejection (Server-side quirks #3). Run `files ls external/<node>/` to discover attached volumes, then re-run the share with `external/<node>/<volume>/...` as the path. Volume roots like `external/<node>/hdd1/` are valid share targets. |
| `refusing to create a public share for <path>: 'files share public' only supports the {drive} namespace(s) (matches the LarePass GUI's per-driver gating). Use 'files share internal' or 'files share smb' for that namespace, or copy the data into drive/Home or drive/Data first.` | Tried `files share public <path>` where the path is in `sync` / `external` / `cache` (Public is locked to `drive` only — mirrors the LarePass GUI's per-driver Share-to-Public condition `event.type === DriveType.Drive || event.type === DriveType.Data`). The gate fires at every depth (root, volume root, deeper subpath). | If the same recipients are Olares users on this node, switch to `files share internal <path>` (cross-user access). For external / cache, `files share smb <path>` also works (network mount). For sync, only `files share internal` works — SMB also refuses sync. If a public link is genuinely required, copy the data into `drive/Home/` or `drive/Data/` first (`files cp -r <path> drive/Home/<sub>/`) and run `files share public` against the Drive copy. |
| `refusing to create a smb share for <path>: 'files share smb' only supports the {cache, drive, external} namespace(s) (matches the LarePass GUI's per-driver gating). Use 'files share internal' for that namespace, or copy the data into drive/Home or drive/Data first.` | Tried `files share smb sync/<repo>/<...>`. SMB rejects the entire `sync` namespace because Seafile libraries don't have a Samba-compatible mount path on the server (the LarePass GUI excludes sync from its Share-to-SMB condition for the same reason). | Use `files share internal sync/<repo>/<...>` to share the Seafile library with other Olares users — that's the only flavor that accepts sync paths. If a network mount is genuinely required, copy / sync the data into `drive/Home/<sub>/` first and `files share smb` against the Drive copy. |
| `refusing to create a <flavor> share for <path>: cloud namespaces (awss3 / google / dropbox / tencent) do not support sharing through 'files share' — the share endpoints don't grant cross-cloud-account access ...` (`<flavor>` ∈ `internal` / `public` / `smb`) | Tried `files share internal\|public\|smb` against an `awss3/`, `google/`, `dropbox/`, or `tencent/` path. Every share-create flavor refuses cloud namespaces uniformly because the resulting share record would point at a per-account credential the recipient has no way to use. | Cloud-backed data has no in-place share path. Recovery: `files download awss3/<account>/<path>` (or the equivalent for the cloud you're using), then `files upload <local> drive/Home/<sub>/`, then `files share <flavor> drive/Home/<sub>/`. The Drive copy is what's actually shareable. |
| `refusing to share cache/<node>/: this is the node-picker layer (no concrete dataset to share); point at a directory inside the node, e.g. cache/<node>/<sub>/` | Tried `files share internal\|public\|smb cache/<node>/` (the LarePass app renders /Files/Cache/ as a node picker, not a directory) | Server-side quirks #5: cache subpaths ARE shareable (`cache/<node>/<app>/...`), but `cache/<node>/` itself is a node selector. Run `files ls cache/<node>/` to see the navigable subdirectories, then re-run with `cache/<node>/<sub>/`. Other verbs (`ls`, `cp`, `upload`, ...) DO work at `cache/<node>/` — the rejection is share-create-only. |
| `Public shares require an expiration` | `share public` without `--expire-days` or `--expire-time` | Pass exactly one expiration flag (Public-link shares need a TTL) |
| `--expire-days and --expire-time are mutually exclusive` | Both flags passed together | Pick one |
| `--password must be at least 6 characters` | `share public --password` shorter than 6 chars | Use a longer password (or omit `--password` to auto-generate one) |
| `--public and --users are mutually exclusive` | `share smb --public --users ...` | Pick one recipient model — Public-SMB OR a specific account list |
| `share smb requires either --public OR --users` | Neither flag set on `share smb` | Add `--public` or `--users id:perm,...` |
| `entry "...": SMB shares accept only view or edit` | `share smb --users id:upload` or `:admin` | SMB perm is binary; use `view` or `edit` |
| `share <id> created, but adding members failed: ...` | `share internal --users` post-create call failed | The share record IS on the server (id is in the message); fix the underlying error and re-run `share internal` with the same `--users`, or call member-add directly through the LarePass app |
| `share ... not found on the server` (`share get`) | Share id is wrong / already removed | List shares to confirm: `files share list` |
| `share ... server reported a conflict (HTTP 409)` (share) | Resource is already shared, or the share id is in use | Use `share list` to find the existing share id; `share rm` it first if you intended to recreate |
| `refusing to <verb> share <id>: the share is <X> (wire type "..."), not <Y>; use the matching update verb instead — share set-password for public shares, share set-members for internal shares, share set-smb for SMB shares` | Ran `share set-password` / `set-members` / `set-smb` against the wrong flavor — e.g. `set-password` on an Internal share (no recipient password); `set-members` on an SMB share (different account model); `set-smb` on a Public share (no SMB user list) | The error already lists every update verb's flavor; switch to the one that matches your share. Run `files share get <id>` if you need to re-confirm what type the share is. Note: Public's wire type is `"external"`, but the CLI / error always says "public" — that's the right verb name. |
| `share set-members requires --users <list> OR --clear` | `files share set-members <id>` with neither flag | Either pass `--users alice:view,bob:edit` (REPLACE the member list) or `--clear` (drop every member). The no-flag form is rejected so a shell-quoting accident can't silently wipe members. |
| `--users contained no usable entries; use --clear to empty the member list intentionally` | `share set-members --users " ,, "` (whitespace-only / commas-only) | If you meant to clear, use `--clear`; if you meant to replace, list at least one entry. |
| `share set-smb requires either --public OR --users` | `files share set-smb <id>` with neither flag | Same recipient-model rule as `share smb` create. |
| `namespace "sync" is not supported by 'files chown'; Seafile permissions live on the library itself — use 'olares-cli files repos' for sync ACLs` | Tried `files chown sync/<repo_id>/<...>` | Sync permissions are managed at the library level (per-user library ACLs in Seahub), not via POSIX uid. Use `olares-cli files repos` (list/get/rename/rm) for the surface that ships today; there is no `files chown` equivalent for sync. |
| `namespace "external" is not supported by 'files chown'; the LarePass GUI hides the Permission tab for external mounts. Allowed: cache, drive` | Tried `files chown external/<node>/<volume>/<...>` | External mounts are supplied by the host (USB / SMB / etc.) and the LarePass file-properties dialog hides the Permission tab for them; the wire surface there is not part of the `chown` contract. If the underlying mount supports POSIX permissions, change them on the host instead. |
| `namespace "<cloud>" is a cloud account; object stores have no POSIX uid concept. Allowed: cache, drive` (`<cloud>` ∈ awss3 / google / dropbox / tencent) | Tried `files chown awss3/<account>/<...>` (or another cloud) | Cloud accounts wrap object stores (S3, Drive, Dropbox, COS); ownership / mode are not part of the object-store data model. There is no recovery path on the cloud namespace itself; if you need POSIX permissions on the data, copy it into `drive/Home/<sub>/` or `drive/Data/<sub>/` (`files cp -r awss3/<account>/<path> drive/Home/<sub>/`) and chown the Drive copy. |
| `refusing to chown the root of <fileType>/<extend>; pick a child path (use -r to fan out across the volume)` | Tried `files chown drive/Home/`, `drive/Data/`, or `cache/<node>/` (`<extend>`-only paths) | Chowning a volume root is a CLI-layer guard — the wire would accept it but the blast radius is severe. Pick a one-level-deeper path (e.g. `drive/Home/Pictures/`) and add `-r` if you want to fan out across that subtree. |
| `--recursive only applies when setting a uid; pass --uid <int> to use it, or drop --recursive to GET the current uid` | Ran `files chown <path> -r` without `--uid` | Without `--uid` the verb is a GET; `-r` is meaningless. Either add `--uid <int>` to make it a recursive PUT, or drop `-r` to inspect the current uid. |
| `--uid "..." is not an integer (LarePass GUI uses 0 for Root and 1000 for User)` | `files chown <path> --uid User` (or any non-integer) | LarePass plumbs human-readable labels into integer wire values; pass the integer directly (`--uid 0` for Root, `--uid 1000` for User). Other UIDs (custom POSIX users on the node) are accepted unchanged. |
| `--uid must be non-negative (got -1)` | `files chown <path> --uid -1` | Negative values are rejected client-side because the server casts to uint and `-1` would silently become a huge UID. Use `0` for Root or `1000` for User. |
| `chown set drive/Home/...: not found on the server (HTTP 404)` | Tried `files chown <path> --uid <int>` against a path that doesn't exist | Confirm the path with `files ls <parent>/`; the chown endpoint is a thin wrapper over the same backend that serves `/api/resources/...`, so a 404 here usually means the file is gone. The same wording is used for the GET path (`chown get ...`). |
| `smb url "..." must start with `//`` | `files smb mount` / `history add` / `history rm` was passed an SMB URL without the leading `//` (e.g. `host.local/Public`) | SMB URLs MUST start with two slashes — fix the path and re-run (`//host.local/Public`). |
| `--password and --password-stdin are mutually exclusive` | Passed both `-p` and `--password-stdin` to `files smb mount` / `history add` | Pick one — `-p <password>` for one-offs (echoed in shell history), `--password-stdin` for scripts (preferred). |
| `--password / --password-stdin requires --user; SMB auth needs both halves` | `files smb history add <url> -p s3cret` (saved password without username) | Add `-u <user>` — a password without a username is unusable for SMB auth, so the CLI assumes it's a typo. If you genuinely want a URL-only favorite, drop `-p` / `--password-stdin`. |
| `--password-stdin: password is empty` | `printf '' \| olares-cli files smb mount ... --password-stdin` (or the pipe died before sending anything) | Make sure the upstream command actually emits the password. For anonymous shares, omit `--password-stdin` and `-p` altogether — interactive mode accepts empty input. |
| `stdin is not a terminal — pipe a password with --password-stdin or pass --password explicitly` | `files smb mount` invoked from a script / CI / heredoc without `-p` or `--password-stdin` | The interactive password prompt requires a TTY. Either pipe the secret with `--password-stdin` or pass `-p` (acceptable when the surrounding tooling already redacts the command line). |
| `mount returned a multi-share list (code 300); re-run with one of the paths above` | `files smb mount //host.local` (host-only address — server replied with the list of discovered shares) | Pick one of the printed paths and re-run mount. Use `--json` if you're scripting around this — it puts the path list under `paths` for jq pipelines. |
| `server rejected (code N): <message>` (smb mount) | The server returned 200/HTTP but a non-200/300 envelope code (typically `code 401` for bad credentials, `code 500` for server-side mount failure) | The wire `<message>` is the server's diagnostic — surface it to the user verbatim. For `code 401` the most common cause is wrong username / password; for `code 500` the SMB target itself is unreachable, the server-side `cifs-utils` mount call failed (verify the share is reachable from the Olares node), or the server is mapping a real auth failure into a 500 — the user-reported `Incorrect username or password` text under `code 500` is exactly this auth-rejection shape. **Triage:** check the progress line — `mount: //host/share @ <node> (user=...)`. If `user=(anonymous)` and the share wants a real account, the user dropped through the interactive flow without typing a username; re-run and type the account at the `SMB username (empty for anonymous):` prompt, or pass `-u <user>` explicitly. **If `smb history list` shows a saved password for the URL but the mount still rejects**, the favorite is stale — rotate it via `smb history rm <url>` + `smb history add <url> -u <user> -p <new-pw>`, or pass `--no-history -u <user> -p <new-pw>` once and re-save afterwards. |
| `note: SMB history unavailable (...); proceeding without autofill` | `files smb mount` tried to fetch `/api/smb_history/<node>/` to autofill missing -u / -p but the call failed (network, 401, 5xx, ...) | **Soft failure — the mount itself is not blocked.** Falls through to the flag- / prompt-driven credential path. If autofill is critical (e.g. the user expected the saved password to be reused), verify the favorite exists via `files smb history list` and that the profile is still authenticated (`profile login`). |
| `note: SMB history has saved credentials for user "<a>" but -u "<b>" was passed; using flags as-is` | The favorite stored credentials for one account but the user asked for a different account via `-u`. | Informational only — the CLI deliberately does not lend account A's password to account B. Either drop `-u <b>` (to use the saved account), pass the matching `-p` / `--password-stdin` for `<b>`, or run a fresh `smb history add //url -u <b> -p <pw>` to save the new pairing. |
| `entry name "..." must not contain '/' or '\\'` | `files smb unmount external/main/smb-host-share` (passed a 3-segment frontend path instead of the bare entry name) | The wire URL shape is `/api/unmount/external/<node>/<name>/`; pass only the `<name>` segment. Discover it with `files ls external/<node>/` first, then re-run `files smb unmount <name> --node <node>`. |
| `could not resolve a node for ... ; pass --node <name> explicitly` | `files smb mount` / `unmount` / `history` was called and the auto-detected `--node` came back empty | Pass `--node <name>` explicitly. Use `olares-cli files cp --help` (the same `--node` cascade lives there) or check the LarePass app's node picker for the right name. |
| `files-backend returned no Drive nodes; cannot resolve default {node}` | The server's `/api/nodes/` returned an empty list AND `--node` was not passed | Pass `--node <name>` explicitly — every SMB verb needs a concrete `<node>` segment in its URL. |
| `unknown repos type "..."` (`files repos list --type ...`) | Misspelled `--type` value | Use `mine`, `share-to-me`, `shared`, or `all` |
| `repos Create: empty repo name` | `files repos create ""` (or whitespace-only) | Pass a non-empty `<name>` argument |
| `repos Create: repo name "." (or "..") is a path-traversal segment, not a real name` | `files repos create .` / `files repos create ..` | Pick a non-reserved name — `.` and `..` are blocked uniformly across `files mkdir`, `files rename`, `files repos create`, and `files repos rename` |
| `repos Create: server accepted the request but did not return a repo_id` | Rare Seahub path that 200s but elides the response payload | Run `files repos list` to discover the new repo; the create likely succeeded |
| `repos Create: server rejected (code N): ...` | Server-side validation failed (duplicate name, encryption-only deployment, etc.) | Surface the message verbatim — it's the Seahub-side reason |
| `repos Rename: new name "." (or "..") is a path-traversal segment, not a real name` | `files repos rename <id> .` / `files repos rename <id> ..` | Same `.`/`..` blacklist as `repos create`; sidestepping create with a benign-then-rename workflow is also blocked |
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

# 4b. Edit a file in place via $EDITOR (TTY-only; no PUT if unchanged).
#     Default 1 MiB cap, text-only guard (extension deny-list + NUL sniff).
#     UPDATE-only: missing path → use `files upload` first, then re-edit.
olares-cli files edit drive/Home/.config/app.yaml
olares-cli files edit drive/Home/Logs/today.log --max-size 0    # disable size cap

# Create-then-edit (two-verb shape; replaces the old --create flag):
echo "" | olares-cli files upload - drive/Home/scratch/new.md
olares-cli files edit drive/Home/scratch/new.md

# 5. Reorganize on the server side (no local round-trip).
olares-cli files cp drive/Home/Documents/2026-Q1.pdf drive/Home/Archive/
olares-cli files mv drive/Home/Inbox/draft.md drive/Home/Notes/2026-04-draft.md
olares-cli files mv -r drive/Home/scratch/ drive/Home/Old/scratch/

# 5b. In-place basename change (synchronous, no task queue).
olares-cli files rename drive/Home/Documents/notes.md 2026-Q1-notes.md

# 5c. Inspect / change the POSIX owner uid (LarePass file-properties Permission tab).
olares-cli files chown drive/Home/Documents/foo.pdf                       # GET — print current uid
olares-cli files chown drive/Home/Documents/foo.pdf --uid 1000            # PUT — hand to default User
olares-cli files chown drive/Home/Pictures/Trip2024/ --uid 0 -r           # PUT recursively to Root
olares-cli files chown cache/<node>/scratch/build/ --uid 1000 -r          # cache namespace works the same way

# 6. Clean up after confirming with `ls` first.
olares-cli files ls drive/Home/Old/
olares-cli files rm -r drive/Home/Old/

# 7. Sharing workflows (orthogonal to the upload / move / cleanup loop).
olares-cli files share internal drive/Home/Reports/Q1.pdf \
    --users alice:edit,bob:view
olares-cli files share public drive/Home/Photos/ --expire-days 7
olares-cli files share smb drive/Home/Movies/ --public

# Edit existing shares without re-creating them.
olares-cli files share set-password <public-share-id>                 # roll the password
olares-cli files share set-members <internal-share-id> \
    --users alice:edit,bob:admin                                       # REPLACE member list
olares-cli files share set-smb <smb-share-id> --public                 # flip to public-SMB

# Inspect / clean up.
olares-cli files share list --shared-by-me
olares-cli files share rm <share-id>

# 8. Sync (Seafile) library lifecycle (orthogonal to drive/Home).
olares-cli files repos list                                  # discover repo_ids
REPO_ID=$(olares-cli files repos create "Project Alpha" --json | jq -r .repo_id)
olares-cli files upload ~/local-tree sync/$REPO_ID/Backups/  # use it like any other path
olares-cli files repos rename $REPO_ID "Project Alpha (archived)"
olares-cli files repos rm $REPO_ID --yes

# 9. Mount an external SMB server, browse it as a regular path, unmount.
#    CLI counterpart of the LarePass "Connect to Server" dialog.
olares-cli files smb mount //host.local/Public -u alice -p s3cret      # ad-hoc
printf '%s' "$SMB_PASSWORD" | \
    olares-cli files smb mount //host.local/Public -u alice --password-stdin   # CI

# Once mounted, every other files verb works against external/<node>/<entry>/.
olares-cli files ls external/main/                  # discover the new entry name
olares-cli files cp ./report.pdf external/main/smb-host-share/Reports/
olares-cli files download external/main/smb-host-share/Movies/ ./local --parallel 4

# Save credentials as a per-node favorite for next time.
olares-cli files smb history add //host.local/Public -u alice -p s3cret
olares-cli files smb history list

# When done.
olares-cli files smb unmount smb-host-share --node main
```

When operating across multiple Olares instances, run `olares-cli profile use <name>` to switch identity before issuing commands. There is no per-invocation `--profile` override — agents/scripts must commit to one profile up-front rather than silently hopping identities mid-pipeline. See `olares-shared` for the full rationale.

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
- **`files smb mount` / `history add` SMB passwords MUST go through `--password-stdin` (or the interactive prompt) in any non-throwaway context.** The `-p <password>` form is documented for one-off ad-hoc invocations only — it leaves the secret in shell history (`bash`, `zsh`, the IDE terminal scrollback, CI command-line logs, ...). For scripts: `printf '%s' "$SMB_PASSWORD" | olares-cli files smb mount --password-stdin ...`. For agents / chat: never inline an SMB password in a suggested command — use the env-var-pipe shape every time. The same rule applies to `files smb history add -p ...`: that wire-PUT stores the password server-side **as plaintext** in the favorites book, which is fine for the LarePass GUI's local-network use case but means a `history list --json` dump is a credential blob — treat it that way (don't pipe it to a logger or an LLM context).
- **`files smb history` is per-NODE, not per-user.** A favorite added against `--node main` is invisible from `--node backup` and vice versa. If a user reports "my saved server isn't appearing", confirm they're on the same node the entry was added under (the default `--node` cascade picks the first `/api/nodes/` entry, which can swap order across deployments).
