---
name: olares-files
version: 4.3.2
description: "Olares Files via olares-cli files — browse known paths; upload or download file bytes; edit, share, mount SMB/NFS, compress/extract archives, and manage Seafile sync across Drive/cache/external/cloud. Use for Olares Files and LarePass Files operations, not URL/yt-dlp/torrent download tasks (olares-knowledge)."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# files (per-user files-backend)

**CRITICAL — before running any verb here, load the `olares-shared` skill first (profile selection, login, 401/403 recovery). Flag reference: `olares-cli files --help`.**

> **Platform model (read once):** the storage areas these paths address — the five userspace areas, their backends/durability, uid-1000 ownership, and the system-managed `drive/Home` directories — are defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md). This skill only adds the **addressing** view.

> **Source of truth for flags & wire shapes is always `olares-cli files <verb> --help`.** This file only carries what `--help` cannot give: the cross-cutting frontend-path concept, the trailing-slash convention, the five client-side hard constraints, and the verb index.

## When to use

- Olares Files, olares-cli files, LarePass Files, drive, Home, Data, Common, sync, cache, upload, download, list, edit, rename, chown
- Archives: compress / extract; inspect (`archive entries` / `archive cat`); manage the async queue (`task` cancel / pause / resume)
- Share: internal cross-Olares-ID, public link (password / expiration), SMB / Samba, Connect to Server, Seafile sync repos
- Mount external servers: SMB (`smb`) and NFS (`nfs`)
- Namespaces: `drive`, `cache`, `sync`, `external`, `awss3`, `dropbox`, `google`, `tencent`, `share`

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Finding files by name/content** (and what the index covers — filenames everywhere vs. full-text only in `/Documents/`) lives in [`olares-search`](../olares-search/SKILL.md); configure which directories get full-text indexing via `settings search dirs` in [`olares-settings`](../olares-settings/SKILL.md).

## Mental model

`files` is an **addressing layer** over the platform storage areas (which the [platform model](../olares-shared/references/olares-platform.md#userspace-storage-model) owns): you name a resource with a 3-segment frontend path and call a verb. Almost every surprise traces back to the path namespace, trailing-slash intent, the 5 client-side hard constraints, or the >= 1.12.6 version gate. Everything else is per-verb `--help`.

## Paths and namespace support

Every files resource uses the 3-segment frontend path `fileType/extend[/subPath]`. Before invoking a verb, confirm the namespace is supported and respect the trailing-slash directory convention. The full path grammar, per-verb namespace allow-list, and trailing-slash rules live in [references/olares-files-paths.md](references/olares-files-paths.md).

## Client-side hard constraints (5 quirks — never work around)

These five rules are enforced client-side and reflect real backend / GUI invariants. The rest of this skill refers to them by number.

### 1. POST `/api/resources/<dir>/` auto-renames on collision

Hitting the directory-create endpoint against an existing directory does NOT return 409 — it silently creates `<dir> (1)` instead. So `files upload` does NOT pre-create the destination directory; use `files mkdir [-p]` first if the parent doesn't exist yet.

### 2. GET `/api/resources/<file>` (no trailing slash) returns HTTP 500

The backend's single-file `List` handler tries to slurp file bytes into a JSON envelope and chokes on most files. The CLI works around it by `Stat`-ing the PARENT directory and finding the leaf in the items array. If the user reports `HTTP 500` on a direct file resource path, the answer is `files cat` / `files download`, never "retry the raw URL".

### 3. `external/<node>/` is a virtual volume-listing layer (read-only)

This level has no backing filesystem — it just enumerates attached volumes (`hdd1`, `usb1`, `smb-...`). `mkdir`, `cp`/`mv` destination, `upload`, AND `share` (all flavors) reject `external/<node>/` (and one level deeper for `mkdir`); errors point at the corrected shape `external/<node>/<volume>/<sub>/`. Pure reads (`ls`, `cat`, `rm`, `rename`) work — that is how the user discovers attached volumes. Mount new volumes via LarePass.

### 4. The system-managed `drive/Home` directories are protected

The eleven LarePass bootstrap directories under `drive/Home/` (canonical names in the platform model → System-managed Home directories) are guarded: `rename`, `rm`, and `mv source` REFUSE these names at the **first level under `drive/Home/` only**. `cp` is NOT gated (duplicating bytes preserves the original). Nested content (`drive/Home/Pictures/Trip2024/`) is fully editable, and other namespaces are unaffected.

### 5. `cache/<node>/` is a node-picker for share-create only

`cache/<node>/` IS a real per-node directory on the wire, so `ls` / `cp` / `mkdir` / `upload` / `rm` / `rename` work. BUT the share-create flavors (`share internal` / `share public` / `share smb`) reject the bare node root because a share record there points at no concrete dataset. Use `cache/<node>/<sub>/` for shares.

## Async task queue (compress / extract)

Unlike every other verb, `compress` and `extract` are **asynchronous**: the POST returns immediately with a `task_id`, and the byte-writing runs on the server's per-node task queue. Mental model:

- Without `--wait`, the command prints the `task_id` and exits — the task is still running server-side.
- With `--wait`, the command polls and prints progress until the task reaches a terminal status.
- **Ctrl-C / context cancel only stops the local poll; the server-side task keeps running.** To actually stop it, use `files task cancel <task_id> --node <node>`.
- Tasks are **per-node** — the `<node>` segment must match the node the task was queued on (`compress` / `extract` print it in their "queued ... task" line). `files task pause` / `resume` / `cancel` manage a queued task; `cancel --all` drops every task on the node.

## Version gate (Olares >= 1.12.6)

The archive surface (`compress` / `extract` / `archive`), the `nfs` verbs, and the `drive/Common` namespace were all introduced in **Olares 1.12.6**. The CLI fails closed on an older (or undetectable) backend with an actionable upgrade message rather than an opaque server 404/500.

- Before reaching for these, check the backend version: the `VERSION` column of `olares-cli profile list`, or live `olares-cli settings me version` (see [olares-shared](../olares-shared/SKILL.md) and [platform.md → version model](../olares-shared/references/olares-platform.md#olares-version--semver-model)).
- Comparison is on `major.minor.patch` only, so a daily build like `1.12.6-20260603` still counts as `>= 1.12.6`.
- If detection fails, confirm the active profile is logged in and run `olares-cli profile list --refresh-version`.

## Authentication transport

Files uses `X-Authorization: <access_token>` and auto-refreshes expired tokens where the request can be safely replayed. On `*ErrTokenInvalidated` / `*ErrNotLoggedIn`, do not retry; use the recovery table in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).

## Verb index

For flags, examples, and wire shapes, **always start with `olares-cli files <verb> --help`**. The references below add only what `--help` cannot give — agent-facing safety constraints, multi-step orchestrations, and common-error → fix maps.

| Verb | `--help` first, then... | Notes |
|---|---|---|
| `ls` | [references/olares-files-ls.md](references/olares-files-ls.md) | Drive vs. cloud envelope shapes; `--json` semantics |
| `cat` | `olares-cli files cat --help` | Trivial GET to stdout; binary-safe |
| `download` | [references/olares-files-download.md](references/olares-files-download.md) | `--resume` / `--overwrite` semantics; directory parallel fetch |
| `upload` | [references/olares-files-upload.md](references/olares-files-upload.md) | Two-stage cloud upload (stage 1 chunks → stage 2 server-side transfer task); `--parallel` semantics; tencent rejection |
| `edit` | [references/olares-files-edit.md](references/olares-files-edit.md) | Editor cascade; three-tier size cap; text-only guard; concurrent-delete detection; cloud writeback gap |
| `mkdir` | [references/olares-files-mkdir.md](references/olares-files-mkdir.md) | `-p` skips existing prefixes; auto-rename quirk on the leaf; `external/<node>/<X>/` depth-1 guard |
| `rm` | [references/olares-files-rm.md](references/olares-files-rm.md) | Preflight existence check before prompt; trailing-slash signals dir; protected-names list |
| `rename` | [references/olares-files-rename.md](references/olares-files-rename.md) | In-place only (synchronous PATCH); protected-names list; bare basename only |
| `cp` / `mv` | [references/olares-files-cp-mv.md](references/olares-files-cp-mv.md) | Drop-into-dir semantics; `mv` source rejects protected names; preflight Stat of every src + dst dir |
| `chown` | [references/olares-files-chown.md](references/olares-files-chown.md) | UID 0 / 1000 conventions; namespace allow-list; volume-root refusal |
| `compress` | [references/olares-files-compress.md](references/olares-files-compress.md) | Async (`task_id` + `--wait`); format set; single-file-compressor limits; password / split-volume zip+7z only; `--level` / `--conflict`. Needs >= 1.12.6 |
| `extract` | [references/olares-files-extract.md](references/olares-files-extract.md) | Async; `<dst-dir>/` must end with `/`; interactive password retry; `--conflict`. Needs >= 1.12.6 |
| `archive` | [references/olares-files-archive.md](references/olares-files-archive.md) | Read-only inspect: `entries` (`--json` / `--max-entries`), `cat` (`-o`); bzip2 / xz not previewable. Needs >= 1.12.6 |
| `task` | [references/olares-files-task.md](references/olares-files-task.md) | Control the per-node queue: `cancel [--all]` / `pause` / `resume`; `pause_able` precheck; `--force` |
| `share` | [references/olares-files-share.md](references/olares-files-share.md) | Three flavors (internal / public / smb); directory-only; per-flavor namespace allow-list; update verbs (`set-members` / `set-password` / `set-smb`) |
| `smb` | [references/olares-files-smb.md](references/olares-files-smb.md) | Mount → `external/<node>/<entry>/`; host-only address triggers share discovery; favorites history |
| `nfs` | [references/olares-files-nfs.md](references/olares-files-nfs.md) | Mount → `external/<node>/<entry>/`; bare host triggers export discovery; no credentials; shares the smb favorites book. Needs >= 1.12.6 |
| `repos` | `olares-cli files repos --help` | List / create / rename / rm Seafile libraries; repo_id is the `<extend>` segment |

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Quirk #3 — bare `external/<node>/` write attempt | Add the `<volume>` segment |
| `refusing to mkdir external/<node>/<X>/: depth-1 entries under external/<node>/ are mounted volumes` | Quirk #3 depth-1 — would create a phantom volume | Mount the volume via LarePass; target an existing one |
| `refusing to {rename\|delete\|mv source} drive/Home/<name>: this is a system-managed Home folder` | Quirk #4 — protected name | Pick a different name, or operate on a nested path |
| `refusing to share cache/<node>/: this is the node-picker layer (no concrete dataset to share)` | Quirk #5 — bare cache node-root share | Use `cache/<node>/<sub>/` |
| `file disappeared between stat and fetch` | Concurrent-delete race on `edit` | Re-pull the parent directory and decide |
| `tencent upload is not supported` (or similar) | Tencent's octet protocol is not implemented | Use the LarePass web app for tencent uploads |
| `<src> does not exist on the server` (from `cp`/`mv`/`rm`) | Preflight Stat failed | `files ls` the parent and confirm the path |
| `HTTP 500` from `/api/resources/<file>` | Quirk #2 — backend tried to embed file bytes | Use `files cat` / `files download` instead |
| Backend version could not be determined | Profile cache is missing/stale or `/api/olares-info` is unreachable | Confirm `profile login`, then run `olares-cli profile list --refresh-version` |
| `require Olares >= 1.12.6` with a detected older version | Backend predates `compress`/`extract`/`archive`/`nfs` or `drive/Common` | Upgrade Olares |
| `files compress` does not support the "sync"/"<cloud>" namespace | Archive allow-list (drive/cache/external only) | Stage into `drive/Home` first, or use the LarePass web app for cloud |
| `archive requires a password` / `archive password is incorrect` | Encrypted zip / 7z | Supply it via `--password-stdin` (or answer the interactive prompt) |
| `previewing "bzip2"/"xz" archives is not supported` | Raw single-stream compressor, no entry table | `files extract` it instead of `archive entries` / `archive cat` |
| `task ... is already <status>` / `not controllable: ... pause_able=false` | Task is terminal or non-interruptible | Nothing to do, or pass `--force` to send anyway |

For auth-related errors (`server rejected the access token`, `refresh token for X became invalid`, …) see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).

## Safety contract

- **Write & delete verbs** (`rm`, `rename`, `cp`, `mv`, `chown --uid`, `compress`, `extract`, `share rm`, `repos rm`, `smb unmount`, `nfs unmount`, `nfs history rm`, `task cancel`) — confirm intent with the user FIRST. Several verbs preflight against the server before any state change; do not bypass that by retry-on-404.
- **`rm -f` skips the y/N prompt but NOT the preflight existence check** — a missing path still aborts.
- **`task cancel --all` drops EVERY task on the node** (including ones started elsewhere) — confirm explicitly; it refuses in a non-TTY context without `--force`.
- **Never echo `access_token` / `refresh_token` to the terminal.** Use `--password-stdin` (where supported) for SMB and archive passwords too.
- **Confirm destination paths** before any `upload --overwrite`, `cp` to an existing file, `compress`/`extract --conflict overwrite`, or any operation that could clobber bytes.
