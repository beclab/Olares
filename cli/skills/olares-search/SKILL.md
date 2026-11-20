---
name: olares-search
version: 0.0.0-cli.0
description: "Olares search via olares-cli search — Files RAG retrieval plus Desktop global search over Drive, Sync, connected clouds, Wise/Knowledge, installed applications, and Market. Use `search rag` when an answer needs relevant indexed file evidence and citations; use `search drive` for the existing global filename/full-content result list."
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
- Retrieve relevant chunks from the authenticated user's configured Files RAG directories, including source URI, locator, scores, and optional highlights.
- Find an installed application by visible title.
- Find an app in the Market catalog, installed or not, by name, title, or description.

> **Mental model:** `search rag` answers *"which indexed file passages are relevant evidence for this question?"* and returns citation data for an Agent or UI. `search drive` / `search sync` / `search knowledge` are the existing global-search result lists. `search app` answers *"which installed app matches this name"*. `search market` searches the catalog. To LIST or READ a known path, use [`olares-files`](../olares-files/SKILL.md).

## Fast paths

| Task | Read | First command |
|---|---|---|
| Answer a question from configured Files RAG evidence | this file | `olares-cli search rag "What is the refund policy?" --highlights -o json` |
| Find a file by name or content | this file | `olares-cli search drive "quarterly report" -o json` |
| Find an installed app by its title | this file | `olares-cli search app music -o json` |
| Find an app to install from the Market | this file, then [`olares-market`](../olares-market/SKILL.md) to install it | `olares-cli search market "photo editor" -o json` |
| Find something in Wise | this file | `olares-cli search knowledge "retrieval" -o json` |
| Tell "not indexed" from "not there" | this file, then [`olares-files`](../olares-files/SKILL.md) for the known path | `olares-cli settings search status -o json` |

## Verb index

| Subcommand | Purpose | Read when triggered |
|---|---|---|
| `rag` | Retrieve relevant chunks from configured Files RAG roots | [Files RAG retrieval](#files-rag-retrieval) — scope, citations, and miss semantics |
| `drive` (`files`) | One search across Drive, Sync, Google Drive, and Dropbox | [index coverage](#index-coverage-drive-only) — what the index reaches decides whether a miss means anything |
| `sync` | Sync (Seafile) libraries only | [indexing](#indexing-drive--sync--cloud--knowledge) — a narrowed view of what `drive` already covers |
| `knowledge` (`wise`) | Wise/Knowledge content search | [indexing](#indexing-drive--sync--cloud--knowledge) — Wise owns this index; needs Olares 1.12.7+ |
| `app` | Visible installed-app title search | nothing further; it is a title match, not lifecycle inventory |
| `market` | Market catalog search over app name, title, and description (`--type aggregate\|file_name`) | nothing further; needs Olares 1.12.7+ and a search3 build that serves the `market` source. `meta.installable: false` means this machine's hardware cannot run the app, `null` means Market has no verdict yet |

There is no separate `gdrive` or `dropbox` verb: on Olares 1.12.7+ `drive` searches `files_v2`, `google_drive`, `dropbox`, and `seafile` in one asynchronous federated request, mirroring the Desktop dialog's single "Files" entry. `sync` restricts that same channel to `seafile`. Olares 1.12.6 and older have no federated channel: there `drive` covers local Drive files only and `sync` falls back to `/api/search/sync`, so on those versions both commands are needed to cover what one covers on 1.12.7+.

### Files RAG retrieval

Use RAG when the task needs evidence passages rather than only a file list:

```bash
olares-cli search rag "Which component owns retry backoff?" \
  --mode hybrid \
  --highlights \
  -o json
```

The command calls user-service, which derives the current user and authorized directory roots. Never invent or pass an owner, Search3 service-protocol URI, or another user's path.
`--root` accepts a frontend path such as `drive/Home/Documents/Project/` and may only narrow the server-authorized set.

JSON output preserves `resource_uri`, `chunk_key`, typed `locator`, scores, provenance, content/source highlights, degraded channels, and trace id.
Keep these fields when citing a result; do not reduce a citation to untraceable generated prose.

An empty hit list means no relevant chunk was returned from the current indexed and authorized scope; it does not prove that the file does not exist.
Check RAG status/directories through the settings skill or verify a known path with `olares-cli files ls`.

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
- `search rag` is also read-only. Do not add an index directory, retry a job, pause a task, or rebuild merely because retrieval returned no hits.
- Do not treat `no results` as proof that a file or app does not exist. Distinguish indexing delay, coverage, extraction failure, integration absence, visibility, and a genuine miss.
- Stop if the user asks to bind an integration or rebuild an index without authorizing that separate action; route to [`olares-settings`](../olares-settings/SKILL.md).
