---
name: olares-search
version: 1.5.0
description: "Olares search via olares-cli search — the Desktop global search over Drive files, Sync (Seafile) libraries, Google Drive, Dropbox, Wise/Knowledge, and installed applications, with paging and JSON output. On Olares 1.12.7+ one `search drive` (alias `search files`) covers all four file sources. Use for Olares search, full-text search, find a file by content, Text Search, global search, search apps, search google drive, search dropbox, search sync, search wise."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# search (Desktop global search)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Use `olares-cli search <subcommand> --help` for syntax.

## When to use

- Find a file or knowledge item by name/content across Drive, Sync, Google Drive, Dropbox, or Wise. One `search drive` covers all four file sources on Olares 1.12.7+.
- Find an installed application by visible title.

> **Mental model:** `search drive` / `search sync` / `search knowledge` answer *"which file CONTAINS this text"* by querying the pre-built per-user index. `search app` answers *"which installed app matches this name"*. It is not lifecycle inventory (`market list --mine` / `market status`) or resource ranking (`dashboard applications`). To LIST or READ a known path, use [`olares-files`](../olares-files/SKILL.md) (`files ls` / `files cat` / `files download`).

## Verb index

| Subcommand | Purpose | Key decision |
|---|---|---|
| `drive` (`files`) | One search across Drive, Sync, Google Drive, and Dropbox | choose filename-only or aggregate based on intent |
| `sync` | Sync (Seafile) libraries only | a narrowed view of what `drive` already covers |
| `knowledge` (`wise`) | Wise/Knowledge content search | requires Olares 1.12.7+; aggregate only |
| `app` | Visible installed-app title search | not lifecycle inventory |

There is no separate `gdrive` or `dropbox` verb: on Olares 1.12.7+ `drive` searches `files_v2`, `google_drive`, `dropbox`, and `seafile` in one asynchronous federated request, mirroring the Desktop dialog's single "Files" entry. `sync` restricts that same channel to `seafile`. Olares 1.12.6 and older have no federated channel: there `drive` covers local Drive files only and `sync` falls back to `/api/search/sync`, so on those versions both commands are needed to cover what one covers on 1.12.7+.

Cloud result locations such as `google/<account>/...` and `dropbox/<account>/...` are valid files paths for `files ls`, `files cat`, or `files download`. A Sync location reads `/sync/<repo_id>/...` and is annotated with the library's display name — `files ls sync/<repo_id>/` takes the id, not the name.

Google Drive and Dropbox results require Olares 1.12.7+ and a bound integration: an unbound account comes back as a failed source rather than an error, so the job still completes and the other sources' results are printed.

### Indexing (drive / sync / cloud / knowledge)

Indexing is asynchronous. A miss does not prove absence. For a known path, verify with `files ls`; inspect extraction failures and index progress with `settings search status`. Rebuild only after confirming coverage and obtaining approval.

Google Drive and Dropbox use provider-side crawling after account binding. Check whether an account is actually bound with `olares-cli settings integration accounts list-by-type google|dropbox`: an empty list is what distinguishes "no integration" from "indexed but no match", and a cloud search cannot tell you which one you hit. Binding itself happens in LarePass → Settings → Integration, not from the CLI. Wise owns Knowledge indexing.

## Index coverage (drive only)

- Filename indexing is broad, subject to exclude patterns.
- Full-content indexing defaults to Drive `/Documents/`; other directories require `settings search dirs add`.
- Supported formats can still fail extraction because they are encrypted, corrupt, or oversized. Inspect `settings search status -o json`.
- Exclude-pattern management remains in the Settings UI; do not invent a CLI verb for it.

## Safety and escalation

- Search is read-only. `settings search dirs add/rm` and `rebuild` are separate configuration mutations and need confirmation.
- Do not treat `no results` as proof that a file or app does not exist. Distinguish indexing delay, coverage, extraction failure, integration absence, visibility, and a genuine miss.
- Stop if the user asks to bind an integration or rebuild an index without authorizing that separate action; route to [`olares-settings`](../olares-settings/SKILL.md).
