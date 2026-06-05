---
name: olares-files
version: 4.1.0
description: "Olares Files via olares-cli files — ls, upload, download, edit, share, SMB mount, Seafile sync on drive/Home, drive/Data, cache, external, cloud. Use for Olares Files, drive, upload, download, share, SMB, LarePass Files."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# files (per-user files-backend)

**CRITICAL — before running any verb here, load the `olares-shared` skill first (profile selection, login, 401/403 recovery). Flag reference: `olares-cli files --help`.**

> **Source of truth for flags & wire shapes is always `olares-cli files <verb> --help`.** This file only carries what `--help` cannot give: the cross-cutting frontend-path concept, the trailing-slash convention, the five client-side hard constraints, and the verb index.

## When to use

- Olares Files, olares-cli files, LarePass Files, drive, Home, Data, sync, cache, upload, download, list, edit, rename, chown
- Share: internal cross-Olares-ID, public link (password / expiration), SMB / Samba, Connect to Server, Seafile sync repos
- Namespaces: `drive`, `cache`, `sync`, `external`, `awss3`, `dropbox`, `google`, `tencent`, `share`

## Core concept: the 3-segment frontend path

Every resource on the per-user files-backend is addressed by:

```
<fileType>/<extend>[/<subPath>]
```

| Segment | Meaning |
|---------|---------|
| `fileType` | Storage class (lowercase, case-sensitive): `drive`, `cache`, `sync`, `external`, `awss3`, `dropbox`, `google`, `tencent`, `share`, `internal` |
| `extend` | Volume / repo / account inside that class. **Case-sensitive.** Drive: only `Home` or `Data`. Cache / external: node name. Sync: seafile repo id. Cloud (`awss3`/`dropbox`/`google`/`tencent`): account key |
| `subPath` | Path inside `extend` (root if omitted). Leading `/` is implicit |

Examples: `drive/Home/`, `drive/Home/Documents/report.pdf`, `sync/<repo_id>/notes/`, `awss3/<account>/<bucket>/key.txt`.

> Drive's `extend` MUST be `Home` or `Data` exactly — `home` is rejected with `invalid drive type`.

### Per-verb namespace support

| Verb | Supported namespaces |
|------|----------------------|
| `ls` / `cat` / `download` / `rm` / `rename` | all of `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` |
| `edit` | `drive`, `sync`, `cache`, `external` only (cloud / tencent / share / internal refused) |
| `mkdir` | all of `drive`, `cache`, `sync`, `external`, `awss3`, `google`, `dropbox`, `tencent` |
| `cp` / `mv` | same as `mkdir` (PATCH `/api/paste/<node>/`) |
| `upload` | `drive/Home`, `drive/Data`, `sync/<repo_id>`, `cache/<node>`, `external/<node>/<volume>`, `awss3`, `google`, `dropbox` — **`tencent` rejected** (different upload protocol) |
| `chown` | `drive/Home`, `drive/Data`, `cache/<node>` only (cloud, sync, external all refused) |
| `share internal` | `drive`, `sync`, `external`, `cache` (cloud refused) |
| `share smb` | `drive`, `external`, `cache` (sync + cloud refused) |
| `share public` | `drive` only |
| `smb mount` / `unmount` / `history` | keyed by `<node>` + `<smb-url>`, not frontend paths |
| `repos` | operates on the Sync (Seafile) library catalog, not frontend paths |

## Trailing-slash convention (critical)

Whether a path ends with `/` is meaningful:

| Form | Meaning |
|------|---------|
| `drive/Home/Foo/` | Directory intent |
| `drive/Home/Foo` | File intent |

It shows up here:

- `files rm drive/Home/Foo/` requires `-r` — the trailing `/` declares "this is a directory".
- `files upload <local> drive/Home/Documents/` → upload INTO Documents; `files upload <local> drive/Home/Documents/2026-Q1.pdf` → upload AS that exact path.
- `files cp <src> <dst>/` and `files mv <src> <dst>/` — `<dst>` MUST end with `/` (drop-into-directory mode). Renaming via `cp`/`mv` is not supported; use `files rename` for in-place basename changes.
- `files cp -r drive/Home/old/` (trailing `/` on a source) requires `-r`.
- `files ls drive/Home/` lists the volume root; both `drive/Home` and `drive/Home/` are accepted but the slash is recommended.

## Client-side hard constraints (5 quirks — never work around)

These five rules are enforced client-side and reflect real backend / GUI invariants. Teach yourself AND the user to respect them — do not suggest curl / API workarounds.

### 1. POST `/api/resources/<dir>/` auto-renames on collision

Hitting the directory-create endpoint against an existing directory does NOT return 409 — it silently creates `<dir> (1)` instead. Therefore: `files upload` does NOT pre-create the destination directory; use `files mkdir [-p]` first if the parent doesn't exist yet.

### 2. GET `/api/resources/<file>` (no trailing slash) returns HTTP 500

The backend's single-file `List` handler tries to slurp file bytes into a JSON envelope and chokes on most files. Workaround baked into the CLI: `Stat` always lists the PARENT directory and finds the leaf in the items array. If the user reports `HTTP 500` on a direct file resource path, the answer is "use `files cat` / `files download`", never "retry the raw URL".

### 3. `external/<node>/` is a virtual volume-listing layer (read-only)

This level has no backing filesystem — it just enumerates attached volumes (`hdd1`, `usb1`, `smb-...`). Writes against it either fail server-side or trip quirk #1.

CLI client-side guards: `mkdir`, `cp` destination, `mv` destination, `upload`, AND `share` (all flavors) reject `external/<node>/` (and one level deeper for `mkdir`). Errors point at the corrected shape `external/<node>/<volume>/<sub>/`. Pure reads (`ls`, `cat`, `rm`, `rename`) DO work — that's how the user discovers what volumes are attached. Mount new volumes via LarePass, not via files-backend mkdir.

### 4. `drive/Home/{Pictures, Music, Movies, Downloads, Documents, Code, Cache, Data, Home, Ollama, Huggingface}` are system-managed

These eleven names under `drive/Home/` are LarePass bootstrap directories that user apps look up by exact name (e.g. the model-runtime app's `Ollama` cache, the LarePass UI's "Pictures" sidebar tile). The LarePass GUI greys out cut / copy / paste / delete / rename for them, and so does the CLI:

- `rename`, `rm`, and `mv source` REFUSE these names at the **first level under `drive/Home/` only**.
- `cp` (copy) is intentionally NOT gated — duplicating bytes (e.g. `cp -r drive/Home/Pictures/ drive/Home/Pictures-Backup/`) preserves the original and is fine.
- Content nested inside (`drive/Home/Pictures/Trip2024/`) is fully editable.
- Other namespaces (`drive/Data/Pictures`, `sync/<repo>/Pictures`, `external/...`) are unaffected.

Note LarePass casing: `Huggingface` is one word (not `HuggingFace`). Names are case-sensitive.

### 5. `cache/<node>/` is a node-picker for share-create only

`cache/<node>/` IS a real per-node directory on the wire, so `ls` / `cp` / `mkdir` / `upload` / `rm` / `rename` work fine. BUT the share-create flavors (`share internal` / `share public` / `share smb`) reject the bare node root because a share record on the node-picker layer points at no concrete dataset. Use `cache/<node>/<sub>/` for shares; `files ls cache/<node>/` for discovery.

## Authentication transport

Every files API call carries `X-Authorization: <access_token>` (NOT `Authorization: Bearer ...`). The transport auto-refreshes expired tokens transparently — reactive on 401/403 for replayable requests (every verb except `upload`), pro-active JWT-exp pre-flight for streaming `upload` chunks (because once an `*os.File` chunk is consumed it can't be replayed). Concurrent goroutines and processes serialize on a single `/api/refresh`.

**On `*ErrTokenInvalidated` / `*ErrNotLoggedIn`, do not retry — only `profile login` / `profile import` will help.** See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for the full recovery table.

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
| `share` | [references/olares-files-share.md](references/olares-files-share.md) | Three flavors (internal / public / smb); directory-only; per-flavor namespace allow-list; update verbs (`set-members` / `set-password` / `set-smb`) |
| `smb` | [references/olares-files-smb.md](references/olares-files-smb.md) | Mount → `external/<node>/<entry>/`; host-only address triggers share discovery; favorites history |
| `repos` | `olares-cli files repos --help` | List / create / rename / rm Seafile libraries; repo_id is the `<extend>` segment |

## Common errors (cross-verb)

| Error fragment | Meaning | Fix |
|---|---|---|
| `is the volume listing layer (read-only); point at a real volume, e.g. external/<node>/<volume>/<sub>/` | Quirk #3 — bare `external/<node>/` write attempt | Add the `<volume>` segment |
| `refusing to mkdir external/<node>/<X>/: depth-1 entries under external/<node>/ are mounted volumes` | Quirk #3 depth-1 — would create a phantom volume | Mount the volume via LarePass; target an existing one |
| `refusing to {rename\|delete\|mv source} drive/Home/<name>: this is a system-managed Home folder` | Quirk #4 — protected name | Pick a different name, or operate on a nested path |
| `refusing to share cache/<node>/: this is the node-picker layer (no concrete dataset to share)` | Quirk #5 — bare cache node-root share | Use `cache/<node>/<sub>/` |
| `file disappeared between stat and fetch` | Concurrent-delete race on `edit` | Re-pull the parent directory and decide |
| `tencent upload is not supported` (or similar) | Tencent's octet protocol is not implemented | Use the LarePass web app for tencent uploads |
| `<src> does not exist on the server` (from `cp`/`mv`/`rm`) | Preflight Stat failed | `files ls` the parent and confirm the path |
| `HTTP 500` from `/api/resources/<file>` | Quirk #2 — backend tried to embed file bytes | Use `files cat` / `files download` instead |

For auth-related errors (`server rejected the access token`, `refresh token for X became invalid`, …) see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).

## Safety contract

- **Write & delete verbs** (`rm`, `rename`, `cp`, `mv`, `chown --uid`, `share rm`, `repos rm`, `smb unmount`) — confirm intent with the user FIRST. Several verbs preflight against the server before any state change; do not bypass that by retry-on-404.
- **`rm -f` skips the y/N prompt but NOT the preflight existence check** — a missing path still aborts.
- **Never echo `access_token` / `refresh_token` to the terminal.** Use `--password-stdin` (where supported) for SMB passwords too.
- **Confirm destination paths** before any `upload --overwrite`, `cp` to an existing file, or any operation that could clobber bytes.
