---
name: olares-download
version: 1.0.0
description: "Olares download task centre via olares-cli download — create / list / pause / resume / cancel / remove tasks, inspect URLs, and set yt-dlp quality prefs. Use for download task, pause download, yt-dlp, download-server, wise download."
compatibility: Requires olares-cli on PATH, active Olares profile, Olares >= 1.12.7
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# download (task centre)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli download <verb> --help`.**

> **Source of truth for flags is always `olares-cli download <verb> --help`.** This file only carries scope, edge path, version gate, and the error → fix matrix.

## When to use

- Manage download-server tasks: create a URL download, list / inspect progress, pause / resume / cancel / remove.
- Probe a URL for provider + available yt-dlp qualities before create.
- Read or set per-app default yt-dlp quality (`prefs`).
- Keywords: download task, pause download, yt-dlp, aria2, huggingface download, wise download, download-server.

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Not** `files download` (pull a Drive/Sync file to the local machine) — that lives in [`olares-files`](../olares-files/SKILL.md). **Not** `os download` (installer packages).

## Edge path & auth

```text
olares-cli download
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
`download` requires Olares >= 1.12.7 (settings /download edge + download provider), but this backend is …
```

Escape hatch when version detection fails: `--olares-version 1.12.7` (same flag as other gated trees).

## Verb index

| Family | Verbs | Details |
|---|---|---|
| lifecycle | `create` / `list` / `info` / `pause` / `resume` / `cancel` / `remove` | [references/olares-download-lifecycle.md](references/olares-download-lifecycle.md) |
| probe + prefs | `inspect` / `prefs get` / `prefs set` | [references/olares-download-inspect.md](references/olares-download-inspect.md) |

Universal: `-o table|json`. Identity/cluster from the active profile only (`profile use` / `profile login`).

## Error → fix

| Symptom | Fix |
|---|---|
| `requires Olares >= 1.12.7` | Upgrade, or pass `--olares-version` only when you know the edge is present |
| `server rejected the access token` / 401 / 403 | `olares-cli profile login` (see olares-shared Auth-readiness gate) |
| `task not found` on pause/info/remove | Wrong id, or task owned by another user (ownership is header-only) |
| create / prefs `ytdlp_quality must be one of…` | Use `best`, `2160p`, `1080p`, `720p`, `480p`, `360p`, or `audio` |
| inspect shows `Error:` / empty qualities | Advisory only — create may still work; check provider / yt-dlp install |
| HTTP 409 on resume/cancel/remove | Task is mid-move (`waiting_to_move` / `moving`); retry after move finishes |
