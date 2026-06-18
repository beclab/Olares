---
name: olares-search
version: 1.1.0
description: "Olares full-content search via olares-cli search — the Desktop global search (\"Text Search\") over the per-user index, across Drive files and Sync (Seafile) libraries, with paging and JSON output. Use for Olares search, full-text search, find a file by content, Text Search, global search."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# search (Desktop global search)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli search --help`.**

> **Source of truth for flags is always `olares-cli search --help`.** This file only carries what `--help` cannot give: what the index covers, how `search` differs from `files`, the session model, and the error → fix matrix.

## When to use

- Find content by keyword across the user's data: "where is the file that mentions X", "search my drive for invoice", full-text / Text Search, Desktop global search.
- Data sources (`--app`): `drive` (user Drive files, default) and `sync` (Seafile / Sync libraries).
- Keywords: Olares search, full-text search, find by content, Text Search, global search, search drive, search sync.

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** `search` answers *"which file CONTAINS this text"* by querying the pre-built per-user index. To LIST or READ a known path, use [`olares-files`](../olares-files/SKILL.md) (`files ls` / `files cat` / `files download`).

## What it is

`olares-cli search <keyword>` is the CLI mirror of the Olares Desktop global search dialog ("Text Search", reached via the Dock search icon / Shift+Space). It runs full-content retrieval against the per-user **search3** index — the same index the desktop SPA queries — so it returns hits by file *content*, not just filename.

- It is a **leaf command** (no subcommands): `olares-cli search <keyword>` runs the query directly. Multiple words are joined into one keyword (quote multi-word phrases to be explicit).
- Results show a title, a location (resource URI / path / repo), and a content snippet with the match. `--output json` emits the raw upstream rows verbatim (nothing the index returns is dropped).

## Search model

- `--app drive` (default) searches Drive via a **session**: the CLI bootstraps `/api/search/init`, pages deeper with `/api/search/more` (same `reqid`), and best-effort cleans up with `/api/search/cancel`. `--app sync` uses the simpler `/api/search/sync` endpoint.
- `--type aggregate` (default) is full-content search; `--type file_name` matches on filename only. **`--type` applies to `--app drive` only** — it is ignored for `--app sync`.
- `--limit` (default 20) caps results; `--offset` pages. For Drive, `--offset > 0` triggers the `/more` page after the init bootstrap.
- Indexing is asynchronous on the server: a just-created/just-uploaded file may not be searchable until the index catches up. A miss is "not indexed yet", not necessarily "absent" — confirm a known path with `files ls`, check progress with `olares-cli settings search status`, and force a full reindex with `olares-cli settings search rebuild` (both in the `olares-settings` skill).

## Index coverage (what is searchable)

Two distinct indexes back this search, and they cover different things — knowing which one a query hits explains most "why didn't it find X" cases.

- **Filenames — indexed broadly by default.** Effectively every Drive file's *name* is indexed, so `--type file_name` hits almost anywhere. The only opt-outs are paths matching an **exclude pattern** (a per-user regex list: built-in non-removable entries plus user-added ones).
- **Full-text (content) — only `/Documents/` by default.** A `--type aggregate` query matches file *contents* only where content indexing is enabled, which by default is just the Drive `/Documents/` directory. Everywhere else you get filename matches but no in-content hits. Supported full-text formats: pdf, doc/docx, csv, rtf, txt/md/json/xml.
- **To make another directory full-text searchable**, add it to the full-content index: `olares-cli settings search dirs add <path>` (inspect the current set with `olares-cli settings search dirs list`). This lives in the `olares-settings` skill (`settings search`).
- **To see what is being blocked**, the exclude patterns are currently viewable/editable only in the Olares Settings SPA → Search → File search (the CLI's `settings search excludes` subcommand is not wired up); built-in patterns cannot be removed.
- **Full-text extraction can fail per file** (encrypted / corrupt / oversized), so a supported-format file in an indexed directory may still have no in-content hits. `olares-cli settings search status` surfaces a `Failures` count (use `-o json` for the per-file detail) — check it first when content search misses a file you expect.

## Examples

```bash
# Default: full-content search of Drive.
olares-cli search report

# Multi-word phrase, explicit app.
olares-cli search "design doc" --app drive

# Filename-only match, larger page.
olares-cli search invoice --app drive --type file_name --limit 50

# Search Sync (Seafile) libraries, paged, JSON for scripting.
olares-cli search notes --app sync --offset 20 -o json
```

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `a non-empty search keyword is required` | No keyword given | Pass `<keyword>` (quote phrases) |
| `unsupported --app "<x>" (allowed: drive, sync)` | Unknown data source | Use `drive` or `sync` |
| `unsupported --type "<x>" (allowed: aggregate, file_name)` | Unknown search mode | Use `aggregate` or `file_name` (drive only) |
| `--limit must be a positive integer` / `--offset must not be negative` | Bad paging values | Use `--limit > 0`, `--offset >= 0` |
| `no results` but the file exists | Index hasn't caught up, or content not indexed | Check `settings search status` for progress/failures; verify the path with `files ls`. If it's a content miss, confirm the dir is in `settings search dirs`, then `settings search rebuild` |
| `server rejected the access token` / token invalidated | Auth | See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) recovery table |
