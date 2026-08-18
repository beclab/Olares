# knowledge download lifecycle

> **Flags:** `olares-cli knowledge download create|list|info|pause|resume|cancel|remove --help`.

## create

```bash
olares-cli knowledge download create 'https://example.com/video' --app wise
olares-cli knowledge download create 'https://…' --path drive/Home/Downloads/ --name clip.mp4 --quality 1080p
olares-cli knowledge download create 'https://…' --format-id 'bv*+ba/b' -o json
```

- `--quality` → `extra.ytdlp_quality`; `--format-id` → `extra.format_id`.
- `--extra` is a JSON object of string values merged into `extra`. `--quality` / `--format-id` are applied after and override matching keys.
- `--path` is normalized locally to match download-server `CreateFileParam`: bare `drive/Home/...` / `drive/Data/...`, `/api/resources/drive/...`, or a full Files API URL (`https://files.<user>.olares.<tld>/api/resources/drive/Home/...`). `Home` / `Data` are case-sensitive. **Not** accepted: browser Files UI URLs (`.../Files/Home/...`), bare `Home/...`, or other file types (cache/external/…). Defaults to `drive/Home/Downloads/`. Pass `--path ""` for HuggingFace cache mode so the server decides.
- Re-creating the same URL always inserts a **new** task row (no 409 duplicate). Check `list` / `info` before creating another copy if reuse was intended.
- Each create sends a fresh `Idempotency-Key`. Transport retries of the **same** CLI attempt reuse that key (server returns the same task). A second user invoke gets a new key and still inserts a new row.
- `--wait [--timeout <duration>]` polls like `wait <id>` after a successful create (mover phases are not success). The created row is printed even when the wait times out or ends in failure, so a script never loses the id.
- Success table line: `Created task <id> status=… provider=… name=…`. Use `-o json` for the full task row.

### HuggingFace (`--path` / `--name` behaviour)

For HuggingFace URLs the destination is chosen by `extra._hf_dest`, **not** by `--path` / `--name`:

- **local** (backend default when `_hf_dest` is unset): lands under `<path>/<repoID>/`. `--path` applies; `--name` is unnecessary because the repo id is the folder name (create-time `(n)` de-dup still applies).
- **cache**: shared `HF_HOME` (Files UI shows `/Common/huggingface/`). `--path` and `--name` are **ignored** — the `huggingface_hub` cache layout (`models--org--repo`) is fixed. Send `--path ""` to match wise.

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

## list / info / wait

```bash
olares-cli knowledge download list --app wise
olares-cli knowledge download list --status downloading --page 1 --page-size 20 -o json
olares-cli knowledge download list --all
olares-cli knowledge download list --all-apps
olares-cli knowledge download info 42
olares-cli knowledge download wait 42
olares-cli knowledge download wait 42 --timeout 10m
olares-cli knowledge download create 'https://…' --wait --timeout 30m
```

- `--all` pages through `/api/download/list` until every matching row is collected (distinct from `sync --all`, which drains the sync cursor).
- `--all-apps` lists across every app and cannot be combined with an explicit `--app` (default `--app wise` is omitted when `--all-apps` is set).
- `--status` is validated locally against the server enum; illegal values fail before any request.
- `wait <id>` / `create --wait` poll `info` every 2s until a terminal status. **Success:** `completed`, `seeding`. **Failure:** `error` (only when `will_auto_retry` is false), `cancelled`, `removed`. **`waiting_to_move` / `moving` are not success** (still relocating bytes), and an `error` the server will retry on its own keeps polling instead of failing the command. `--timeout` defaults to 15m (same as the market / users watch commands); on expiry the exit is non-zero and the current status is printed — `create --wait` still prints the created row first, so the id survives. Transient poll errors are retried until 5 of them arrive in a row, then wait aborts; Ctrl-C reports as cancelled by user, not as a timeout. Polling only — no WebSocket watch.

Table columns: `ID`, `STATUS`, `PROVIDER`, `PERCENT`, `NAME`, `SOURCE`, `APP`, `UPDATED`. `SOURCE` is the task URL (magnet / http / …). Footer shows `N of total` when the server returns `total`.

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
