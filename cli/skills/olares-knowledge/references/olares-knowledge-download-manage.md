# knowledge download: managing tasks that exist

> **Flags:** `olares-cli knowledge download list|info|pause|resume|cancel|remove --help`. Creating a task and waiting for it is [create and follow](olares-knowledge-download-create.md).

## list / info

```bash
olares-cli knowledge download list --app wise
olares-cli knowledge download list --status downloading --page 1 --page-size 20 -o json
olares-cli knowledge download list --all
olares-cli knowledge download list --all-apps
olares-cli knowledge download info 42
```

- `--all` pages through `/api/download/list` until every matching row is collected (distinct from `sync --all`, which drains the sync cursor).
- `--all-apps` lists across every app and cannot be combined with an explicit `--app` (the default `--app larepass` is omitted when `--all-apps` is set).
- `--status` is validated locally against the server enum; illegal values fail before any request.

Table columns: `ID`, `STATUS`, `PROVIDER`, `PERCENT`, `NAME`, `SOURCE`, `APP`, `UPDATED`. `SOURCE` is the task URL (magnet / http / …). Footer shows `N of total` when the server returns `total`.

`NAME` is `file_name` when the server has one, else the torrent metadata name, the magnet's `dn=` / info-hash, or a URL basename that actually looks like a file. It shows `-` rather than echoing the URL, so a `-` means "no name written yet", not an error — the URL is in `SOURCE` either way.

## pause / resume / cancel

```bash
olares-cli knowledge download pause 42
olares-cli knowledge download pause 42 43 44
olares-cli knowledge download resume 42
olares-cli knowledge download cancel 42
```

One id uses the single-task route. Two or more ids use
`PUT /api/download/batch/{pause,resume,cancel}` (max 500). Table prints
succeeded/failed counts; any failure exits non-zero.

Single-task HTTP semantics during the yt-dlp mover phase
(`waiting_to_move` / `moving`):

- **resume / cancel / remove** → **409** — wait for the move, then retry
  (`info <id>`).
- **pause** → **400** (status not pausable) — not a 409.

Batch routes stay HTTP 200; per-id failures land in `failed[]`.

## remove

```bash
olares-cli knowledge download remove 42
olares-cli knowledge download remove 42 --remove-file
olares-cli knowledge download remove 42 43 44 --remove-file
```

`--remove-file` sets `remove_flag=true` (delete artefact on PVC). Without it the downloaded file is kept. Multiple ids use `DELETE /api/download/batch/remove`.

`remove` retires the task rather than deleting it: the row stays in `list` with status `removed`, which is terminal. So a `list` that still shows the task is not a failed remove, and the id is not freed for reuse — check the status, not the presence of the row.
