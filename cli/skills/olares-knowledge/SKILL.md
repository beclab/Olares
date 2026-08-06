---
name: olares-knowledge
version: 1.0.1
description: "Olares Knowledge via olares-cli knowledge — manage download-server URL, yt-dlp, aria2, torrent, HuggingFace, and Wise download tasks: create/list/inspect/pause/resume/cancel/remove, prefs, sync, and file probes. Requires Olares 1.12.7+. Not for installer download or copying a Drive file with files download."
compatibility: Requires olares-cli on PATH, active Olares profile, Olares >= 1.12.7
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# knowledge

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli knowledge download <verb> --help`.**

> **Source of truth for flags is always `olares-cli knowledge download <verb> --help`.** This file only carries scope, edge path, version gate, and the error → fix matrix.

## When to use

- Manage download-server tasks via `knowledge download`: create a URL download, list / inspect progress, pause / resume / cancel / remove.
- Probe a URL for provider + available yt-dlp qualities before create.
- Read or set per-app default yt-dlp quality (`prefs`).
- Keywords: knowledge download, download task, pause download, yt-dlp, aria2, huggingface download, wise download, download-server.

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Not** top-level `download` (installer packages: `download component` / `wizard` / `check`). **Not** `files download` (pull a Drive/Sync file) — that lives in [`olares-files`](../olares-files/SKILL.md).

## Edge path & auth

```text
olares-cli knowledge download
  → SettingsURL + "/download" + /api/...
  → settings nginx → user-service DownloadController
  → download provider → download-server
```

- Auth: profile access token (`X-Authorization`). The gateway injects `X-Bfl-User`; the CLI must not set it.
- Response envelope is download-server's `{code, data|list|total|message}` (success `code` 200 or 0), not BFL's market envelope.
- Default `--app` is `wise`.

## Version gate

Requires **Olares >= 1.12.7** (settings `/download` edge + download provider). Below that, every verb fails closed before any HTTP call:

```text
`knowledge download` requires Olares >= 1.12.7 (settings /download edge + download provider), but this backend is …
```

If version detection fails, confirm the active profile is logged in and run `olares-cli profile list --refresh-version`. If the detected version is below 1.12.7, upgrade Olares.

## Verb index

| Family | Verbs | Details |
|---|---|---|
| lifecycle | `create` / `list` / `info` / `pause` / `resume` / `cancel` / `remove` | [references/olares-knowledge-download-lifecycle.md](references/olares-knowledge-download-lifecycle.md) |
| probe + prefs | `inspect` / `prefs get` / `prefs set` | [references/olares-knowledge-download-inspect.md](references/olares-knowledge-download-inspect.md) |
| sync | `unfinished` / `sync` (`--since`/`--since-id` cursor, `--all` drain) | [references/olares-knowledge-download-sync.md](references/olares-knowledge-download-sync.md) |
| torrent | `torrent inspect` / `stats` / `peers` / `files` / `seed stop\|resume` (+ `create --torrent` / `--select-files` / magnet) | [references/olares-knowledge-download-torrent.md](references/olares-knowledge-download-torrent.md) |
| file tools | `file exists` (URL pre-check) / `file check` / `file remove` (resource path) | [references/olares-knowledge-download-files.md](references/olares-knowledge-download-files.md) |
| settings | `settings get` / `settings set` (download-server global config) | [references/olares-knowledge-download-settings.md](references/olares-knowledge-download-settings.md) |

> `cookies` is hidden: production integration-provider is read-only (`retrieve` only); `list` / `set` / `delete` / `health` 404 on a real cluster. See [references/olares-knowledge-download-cookies.md](references/olares-knowledge-download-cookies.md).

Universal: `-o table|json`. Identity/cluster from the active profile only (`profile use` / `profile login`).

## Error → fix

| Symptom | Fix |
|---|---|
| Backend version could not be determined | Confirm `profile login`, then run `olares-cli profile list --refresh-version` |
| `requires Olares >= 1.12.7`, with a detected older version | Upgrade Olares; the Settings `/download` edge is not available on older releases |
| `server rejected the access token` / 401 / 403 | `olares-cli profile login` (see olares-shared Auth-readiness gate) |
| `task not found` on pause/info/remove | Wrong id, or task owned by another user (ownership is header-only) |
| create / prefs `ytdlp_quality must be one of…` | Use `best`, `2160p`, `1080p`, `720p`, `480p`, `360p`, or `audio` |
| inspect shows `Error:` / empty qualities | Advisory only — create may still work; check provider / yt-dlp install |
| HTTP 409 on resume/cancel/remove | Task is mid-move (`waiting_to_move` / `moving`); retry after move finishes |
