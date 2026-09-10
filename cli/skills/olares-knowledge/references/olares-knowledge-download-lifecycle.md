# knowledge download lifecycle

> **Flags:** `olares-cli knowledge download create|list|info|wait|pause|resume|cancel|remove --help`.

## create

```bash
olares-cli knowledge download create 'https://example.com/video' --app wise
olares-cli knowledge download create 'https://…' --path drive/Home/Downloads/ --quality 1080p
olares-cli knowledge download create 'https://…' --format-id 'bv*+ba/b' -o json
```

- `--quality` → `extra.ytdlp_quality`; `--format-id` → `extra.format_id`.
- `--extra` is a JSON object of string values merged into `extra`. `--quality` / `--format-id` are applied after and override matching keys.
- `--path` is normalized locally to match download-server `CreateFileParam`: bare `drive/Home/...` / `drive/Data/...`, `/api/resources/drive/...`, or a full Files API URL (`https://files.<user>.olares.<tld>/api/resources/drive/Home/...`). `Home` / `Data` are case-sensitive. **Not** accepted: browser Files UI URLs (`.../Files/Home/...`), bare `Home/...`, or other file types (cache/external/…). Defaults to `drive/Home/Downloads/`. Pass `--path ""` for HuggingFace cache mode so the server decides.
- Re-creating the same URL always inserts a **new** task row (no 409 duplicate). Check `list` / `info` before creating another copy if reuse was intended.
- Each create sends a fresh `Idempotency-Key`. Transport retries of the **same** CLI attempt reuse that key (server returns the same task). A second user invoke gets a new key and still inserts a new row.
- Success table line: `Created task <id> status=… provider=… name=…`. Use `-o json` for the full task row.

### Naming: leave `--name` off

`create` probes the URL itself and sends the inspect **title** as `file_name` — the same prefill LarePass does in its New Task dialog. That is why the task reads as the video title from the moment it is created instead of sitting nameless until the provider writes the real filename back minutes later.

- **Do not invent a `--name` from the URL.** A routing segment such as YouTube's `/watch` or Bilibili's `/video/BV…` is not a filename, and a name passed on create is pinned to the row for good — the inspect title never gets a chance to land.
- Pass `--name` only when the user asked for a specific filename.
- A failed probe (505, cookie-gated, channel URL past the server deadline) just omits the field; the provider still writes the real name back. Create is not blocked, so do not retry or report failure over it.
- `inspect` is still worth running on its own when you need the qualities or want to show the user the title first — it is not needed to get the name right.

The probe costs a round trip, and a yt-dlp inspect regularly takes several seconds against a sub-second ordinary call, so create is correspondingly slower than it used to be:

- `--inspect-timeout <duration>` caps it (default 15s, chosen to clear a normal yt-dlp probe while still cutting off a pathological one). On expiry the field is omitted and create carries on.
- `--no-inspect` skips it entirely — use when create latency matters more than having a name straight away. Cannot be combined with `--inspect-timeout`.

Magnet links, `--torrent` uploads and HuggingFace repos **reject** `--name`:

| Source | `--name` | Where the name comes from |
|---|---|---|
| yt-dlp / plain http | optional, usually omit | inspect title, else the provider |
| magnet / `--torrent` | rejected | torrent metadata (aria2 skips its `out` option) |
| HuggingFace | rejected | repo id, or the shared-cache layout |

The flag fails locally with the provider named. This mirrors LarePass, which disables its rename field for the same three: a custom name there would only be pinned to the task row while the file on disk keeps the original one.

### HuggingFace (`--path` behaviour)

For HuggingFace URLs the destination is chosen by `extra._hf_dest`, **not** by `--path`:

- **local** (backend default when `_hf_dest` is unset): lands under `<path>/<repoID>/`. `--path` applies; the repo id is the folder name (create-time `(n)` de-dup still applies).
- **cache**: shared `HF_HOME` (Files UI shows `/Common/huggingface/`). `--path` is **ignored** — the `huggingface_hub` cache layout (`models--org--repo`) is fixed. Send `--path ""` to match wise.

Set HF options through `--extra` (flat string keys map 1:1 to `hf` CLI flags; `_hf_dest` is the only internal key):

```bash
# cache mode (what the wise UI defaults to)
olares-cli knowledge download create 'https://huggingface.co/org/repo' \
  --extra '{"_hf_dest":"cache"}' --path ""

# local mode with token / revision / include filter
olares-cli knowledge download create 'https://huggingface.co/org/repo' \
  --path drive/Home/Downloads/ \
  --extra '{"_hf_dest":"local","token":"hf_xxx","revision":"v1.0","include":"*.safetensors"}'
```

Recognised HF `--extra` keys: `_hf_dest` (`cache`|`local`), `token`, `revision`, `include`, `exclude`, `max-workers`, `repo-type`. Note wise defaults HF to **cache**; this CLI defaults to **local** unless you pass `_hf_dest`.

## Following a task: `info` polling, not `--wait`

`create` returns as soon as the row exists — the download itself is the server's job and can run for many minutes. **Default path: take the id off create and poll `info <id>` yourself**, every few seconds, so the conversation stays responsive and a slow download does not eat a whole turn.

```bash
olares-cli knowledge download create 'https://www.youtube.com/watch?v=…' -o json   # → id
olares-cli knowledge download info 42                                              # → status / percent / file_name
```

Terminal states are the same however you poll, and the **status is the only input**. **Success:** `completed`, `seeding`. **Failure:** `error`, `cancelled`, `removed`. **Still running:** everything else, including `waiting_to_move` / `moving` — bytes are still being relocated, so neither counts as success.

`will_auto_retry` does not enter that judgement. It is a flag the server derives for the UI (an `error` row whose `err_category` is non-terminal and whose retry budget is not spent), and it means "a background sweep may pick this up again" — not "this row is still running". Report the failure. If the user wants to know whether the retry lands, poll `info` again rather than holding the task open.

`--wait` is the blocking alternative, for scripts that want one call rather than a poll loop:

```bash
olares-cli knowledge download wait 42
olares-cli knowledge download wait 42 --timeout 10m
olares-cli knowledge download create 'https://…' --wait --timeout 30m
```

`wait <id>` / `create --wait` poll `info` every 2s until a terminal status. `--timeout` defaults to 15m (same as the market / users watch commands); on expiry the exit is non-zero and the current status is printed — `create --wait` still prints the created row, so the id survives. Transient poll errors are retried until 5 of them arrive in a row, then wait aborts; Ctrl-C reports as cancelled by user, not as a timeout. Polling only — no WebSocket watch.

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
