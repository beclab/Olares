---
name: olares-search
version: 1.2.0
description: "Olares search via olares-cli search — the Desktop global search over Drive files, Sync (Seafile) libraries, and installed applications, with paging and JSON output. Use for Olares search, full-text search, find a file by content, Text Search, global search, search apps."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# search (Desktop global search)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli search <subcommand> --help`.**

> **Source of truth for flags is always `olares-cli search <subcommand> --help`.** This file only carries what `--help` cannot give: what the index covers, how `search` differs from `files`, the session model, and the error → fix matrix.

## When to use

- Find content by keyword across the user's data: "where is the file that mentions X", "search my drive for invoice", full-text / Text Search, Desktop global search.
- Find installed applications by title: "open wise", "search for firefox app".
- Subcommands: `drive` (user Drive files), `sync` (Seafile / Sync libraries), `app` (installed applications).
- Keywords: Olares search, full-text search, find by content, Text Search, global search, search drive, search sync, search app.

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** `search drive` / `search sync` answers *"which file CONTAINS this text"* by querying the pre-built per-user index. `search app` answers *"which installed app matches this name"*. It is not lifecycle inventory (`market list --mine` / `market status`) or resource ranking (`dashboard applications`). To LIST or READ a known path, use [`olares-files`](../olares-files/SKILL.md) (`files ls` / `files cat` / `files download`).

## What it is

`olares-cli search` is the CLI mirror of the Olares Desktop global search dialog (reached via the Dock search icon / Shift+Space). It is a **command group** with three subcommands:

| Subcommand | Backend | Purpose |
|---|---|---|
| `search drive <keyword>` | search3 index (`/api/search/init` + `/more`) | Full-content / filename search of Drive files |
| `search sync <keyword>` | `/api/search/sync` | Search Seafile/Sync libraries |
| `search app <keyword>` | `/server/myApps` (local filter) | Search installed applications by entrance title |

Multiple words are joined into one keyword (quote multi-word phrases to be explicit).

- `drive` / `sync` results show a title, a location (resource URI / path / repo), and a content snippet with the match. `--output json` emits the raw upstream rows verbatim.
- `app` results show title, app id, entrance name, state, and URL. `--output json` emits structured app rows.

## Search model

### drive

- Session-based: bootstraps `/api/search/init`, pages deeper with `/api/search/more` (same `reqid`), best-effort cleanup via `/api/search/cancel`.
- `--type aggregate` (default) is full-content search; `--type file_name` matches on filename only.
- `--limit` (default 20) caps results; `--offset` pages. `/init` always returns the first 20 hits and ignores offset/limit, so the CLI honors them itself: when the requested window fits inside those 20 hits (or the search returned fewer), it is sliced client-side with no extra call; only a window past the first 20 triggers `/more` (limit clamped to the backend's 1-100). A single search resolves at most ~50 hits server-side, so `--limit` is effectively capped around 50, and an `--offset` past the end prints "no results" rather than erroring.

### sync

- Single endpoint: `POST /api/search/sync {query}` (the proxy ignores any paging fields).
- **No `--type` flag** (sync does not support search mode selection). `--limit` / `--offset` are applied **client-side** by the CLI (the backend returns the full result set and does not paginate), so they still work but every call fetches everything first.

### app

- Fetches installed apps from `/server/myApps`, expands visible entrances, filters by case-insensitive substring match on the entrance title (falling back to the app title when an entrance declares none).
- Returns all matches in one response (**no `--limit` / `--offset`**).
- Skips uninstalled/failed apps and invisible entrances (same rules as the Desktop SPA). For a `running` app the row shows the per-entrance state.

### Indexing (drive / sync only)

Indexing is asynchronous on the server: a just-created/just-uploaded file may not be searchable until the index catches up. A miss is "not indexed yet", not necessarily "absent" — confirm a known path with `files ls`, check progress with `olares-cli settings search status`, and force a full reindex with `olares-cli settings search rebuild` (both in the [`olares-settings`](../olares-settings/SKILL.md) skill).

## Index coverage (drive only)

Two distinct indexes back Drive search — knowing which one a query hits explains most "why didn't it find X" cases.

- **Filenames — indexed broadly by default.** Effectively every Drive file's *name* is indexed, so `--type file_name` hits almost anywhere. The only opt-outs are paths matching an **exclude pattern** (a per-user regex list: built-in non-removable entries plus user-added ones).
- **Full-text (content) — only `/Documents/` by default.** A `--type aggregate` query matches file *contents* only where content indexing is enabled, which by default is just the Drive `/Documents/` directory. Everywhere else you get filename matches but no in-content hits. Supported full-text formats: pdf, Word (doc/docx), Excel (xls/xlsx), csv, rtf, plus a wide range of plain-text and source-code extensions (txt/md/json/xml/yaml and many more).
- **To make another directory full-text searchable**, add it to the full-content index: `olares-cli settings search dirs add <path>` (inspect the current set with `olares-cli settings search dirs list`). This lives in the [`olares-settings`](../olares-settings/SKILL.md) skill (`settings search`).
- **To see what is being blocked**, the exclude patterns are viewable/editable only in the Olares Settings SPA → Search → File search (the CLI's `settings search excludes` subcommand is not wired up); built-in patterns cannot be removed.
- **Full-text extraction can fail per file** (encrypted / corrupt / oversized), so a supported-format file in an indexed directory may still have no in-content hits. `olares-cli settings search status` surfaces a `Failures` count (use `-o json` for the per-file detail) — check it first when content search misses a file you expect.

## Examples

```bash
# Full-content search of Drive.
olares-cli search drive report

# Multi-word phrase, filename-only match.
olares-cli search drive "design doc" --type file_name --limit 50

# Search Sync (Seafile) libraries, paged, JSON for scripting.
olares-cli search sync notes --offset 20 -o json

# Search installed applications.
olares-cli search app wise
olares-cli search app firefox -o json
```

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `a non-empty search keyword is required` | No keyword given | Pass `<keyword>` (quote phrases) |
| `unsupported --type "<x>" (allowed: aggregate, file_name)` | Unknown search mode on drive | Use `aggregate` or `file_name` (drive only) |
| `--limit must be a positive integer` / `--offset must not be negative` | Bad paging values | Use `--limit > 0`, `--offset >= 0` |
| `no results` but the file exists | Index hasn't caught up, or content not indexed | Check `settings search status` for progress/failures; verify the path with `files ls`. If it's a content miss, confirm the dir is in `settings search dirs`, then `settings search rebuild` |
| `no results` for app search | No installed app matches the keyword | Try a shorter keyword; confirm the app is installed and visible |
| `server rejected the access token` / token invalidated | Auth | See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) recovery table |
