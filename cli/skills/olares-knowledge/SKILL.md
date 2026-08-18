---
name: olares-knowledge
version: 1.1.0
description: "Olares Knowledge via olares-cli knowledge — manage download-server URL, yt-dlp, aria2, torrent, HuggingFace, and Wise download tasks: create/list/inspect/pause/resume/cancel/remove, prefs, sync, and file probes. Requires Olares 1.12.7+. Not for installer download or copying a Drive file with files download."
compatibility: Requires olares-cli on PATH, active Olares profile, Olares >= 1.12.7
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# knowledge

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Use `olares-cli knowledge download <verb> --help` for syntax.

## When to use

- Create and manage URL, yt-dlp, aria2, torrent, HuggingFace, or Wise download-server tasks.
- Inspect providers/qualities, synchronize task changes, manage torrent state, or probe/remove downloaded resources.
- Read or change download preferences and download-server settings.

> **Not** top-level `download` (installer packages: `download component` / `wizard` / `check`). **Not** `files download` (pull a Drive/Sync file) — that lives in [`olares-files`](../olares-files/SKILL.md).

## Version gate

All verbs require Olares 1.12.7+ because the Settings download edge and provider are both required. If the version cannot be established, follow the shared profile/auth gate before deciding that an upgrade is needed.

## Verb index

| Family | Verbs | Read when triggered |
|---|---|---|
| lifecycle | `create`, `list`, `info`, `wait`, `pause`, `resume`, `cancel`, `remove` | [task lifecycle and state decisions](references/olares-knowledge-download-lifecycle.md) |
| probe + prefs | `inspect`, `prefs get`, `prefs set` | [provider/quality inspection](references/olares-knowledge-download-inspect.md) |
| sync | `unfinished`, `sync` | [cursor and drain semantics](references/olares-knowledge-download-sync.md) |
| torrent | `torrent inspect`, `stats`, `peers`, `files`, `seed stop/resume`; torrent create | [torrent selection and seeding](references/olares-knowledge-download-torrent.md) |
| file tools | `file exists`, `file remove` | [URL vs resource-path decisions](references/olares-knowledge-download-files.md) |
| settings | `settings get`, `settings set` | [global download-server settings](references/olares-knowledge-download-settings.md) |

## Task and asynchronous semantics

- Create returns a server-side task; command success does not mean bytes have finished downloading or moving. Use `wait` / `create --wait` when scripts need a true terminal status (`waiting_to_move` / `moving` are still in progress, and an `error` row with `will_auto_retry` is still the server's to resolve).
- Re-submitting the same URL always creates a **new** task (no identity dedup). Landing-name collisions are resolved with a `(n)` suffix; they do not reuse or block an existing row. Create sends `Idempotency-Key` only to collapse transport retries of one attempt — not URL dedup.
- Pause only applies while `waiting` or `downloading` (otherwise 400). Resume, cancel, and remove return **409** while the task is in the yt-dlp mover phase (`waiting_to_move` / `moving`) — wait and retry; do not treat pause the same way.
- Task ownership follows the active profile. Do not infer another user's task from an id or try alternate identities.
- `inspect` is advisory: provider/quality probing may fail while a create still works. Report that uncertainty instead of declaring the URL undownloadable.
- Sync cursors describe change observation, not task completion. Preserve the cursor when continuing an incremental sync.

## Safety and escalation

- Confirm create, cancel, remove, seed stop/resume, file remove, preference writes, and global setting writes.
- Before create, confirm destination/app, provider intent, torrent file selection, and whether an existing equivalent task should be reused. Use `list` / `info` to check; the CLI does **not** detect or block duplicates.
- `file remove` takes a download-server resource path, not an arbitrary local filesystem path.
- A URL that fails for a missing login is not a dead end: cookies live in [`olares-settings`](../olares-settings/SKILL.md) under `settings integration cookie`. See [provider/quality inspection](references/olares-knowledge-download-inspect.md) for the signals and the hand-off.
- Stop on ambiguous URL/resource path, task owner, duplicate-task intent, torrent selection, or any credential request the cookie hand-off does not cover.
