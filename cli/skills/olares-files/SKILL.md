---
name: olares-files
version: 4.4.0
description: "Olares Files via olares-cli files — browse known paths; upload or download file bytes; edit, share, mount SMB/NFS, compress/extract archives, and manage Seafile sync across Drive/cache/external/cloud. Use for Olares Files and LarePass Files operations, not URL/yt-dlp/torrent download tasks (olares-knowledge)."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# files (per-user files-backend)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Load the shared [platform model](../olares-shared/references/olares-platform.md) when the task depends on userspace backends, durability, uid 1000, or system-managed Home directories. Use `olares-cli files <verb> --help` for syntax.

## When to use

- Address, read, write, move, delete, share, or transfer known Olares file paths.
- Compress/extract or manage their asynchronous task queue.
- Mount SMB/NFS servers or manage Seafile repositories.

> **Finding files by name/content** (and what the index covers — filenames everywhere vs. full-text only in `/Documents/`) lives in [`olares-search`](../olares-search/SKILL.md); configure which directories get full-text indexing via `settings search dirs` in [`olares-settings`](../olares-settings/SKILL.md).

## Paths and namespace support

Every resource uses `fileType/extend[/subPath]`. Before invoking a verb, confirm its namespace support and whether the target is a file or directory. Load [path grammar and namespace rules](references/olares-files-paths.md) whenever constructing or interpreting a path.

## Backend quirks that change decisions

- **Directory creation may auto-rename on collision** instead of returning conflict. Upload does not pre-create destination directories. If exact naming matters, inspect the parent and create it deliberately.
- **Direct raw GET of a file resource may return HTTP 500.** Use `files cat` or `files download`; do not retry the raw resource URL.
- `external/<node>/` is a virtual volume picker, not a write destination. Writes need `external/<node>/<volume>/...`.
- Eleven [reserved names](../olares-shared/references/olares-platform.md#system-managed-home-directories) are system-managed at the first level under `drive/Home/`. `rename`, `rm`, and `mv` as a source are refused before any request goes out; `cp` and everything nested stay editable. The refusal enumerates all eleven and is a platform invariant, not a permission error to retry — act on a child instead. The names are matched exactly, so a name that merely looks similar is not protected.
- `cache/<node>/` is valid storage but not a concrete dataset for share creation; share `cache/<node>/<sub>/`.

## Async task queue (compress / extract)

- `compress` and `extract` enqueue server-side work and return a task id unless waiting.
- Ctrl-C stops only the local poll. Cancel the server task explicitly if that is the user's intent.
- Tasks are per-node; retain the node printed when the task was queued.
- `task cancel --all` affects every task on the node, including work started elsewhere.

## Version gate (Olares >= 1.12.6)

Archives, NFS, and `drive/Common` require Olares 1.12.6+. Treat daily builds by their `major.minor.patch` base. Follow the shared auth/version gate when the backend version cannot be established.

## Verb index

| Verb | Read when triggered |
|---|---|
| `ls` | [listing and cloud shapes](references/olares-files-ls.md) |
| `cat` | `files cat --help` |
| `download` | [resume, overwrite, directory downloads](references/olares-files-download.md) |
| `upload` | [collision decisions and cloud transfer](references/olares-files-upload.md) |
| `edit` | [text/size guards and writeback](references/olares-files-edit.md) |
| `mkdir` | [parents, auto-rename, external depth](references/olares-files-mkdir.md) |
| `rm` | [existence, directory intent, protected paths](references/olares-files-rm.md) |
| `rename` | [in-place rename and protected paths](references/olares-files-rename.md) |
| `cp`, `mv` | [destination and overwrite semantics](references/olares-files-cp-mv.md) |
| `chown` | [UID and namespace decisions](references/olares-files-chown.md) |
| `compress` | [formats, conflicts, passwords, async task](references/olares-files-compress.md) |
| `extract` | [destination, conflicts, passwords, async task](references/olares-files-extract.md) |
| `archive` | [read-only archive inspection](references/olares-files-archive.md) |
| `task` | [per-node cancel/pause/resume](references/olares-files-task.md) |
| `share` | [internal/public/SMB sharing](references/olares-files-share.md) |
| `smb` | [discovery, mount, history](references/olares-files-smb.md) |
| `nfs` | [export discovery and mount](references/olares-files-nfs.md) |
| `repos` | `files repos --help` |

## Safety contract

- Treat the user's requested file operation and named paths as task-scope authorisation. Ask again only when a target is ambiguous, bytes may be overwritten without explicit intent, deletion expands beyond the named target, or the action leaves that scope.
- For upload, first decide whether collision should overwrite, fail, or create a distinct name; backend auto-renaming is not an acceptable implicit decision.
- Do not retry a missing-path preflight by weakening safety flags.
- Confirm `task cancel --all` separately because it affects unrelated work on that node.
- Never expose tokens or passwords. Use stdin-based secret input where available.
- Stop on ambiguous frontend paths, node/volume identity, overwrite intent, or concurrent deletion; re-list the parent and ask the user.
