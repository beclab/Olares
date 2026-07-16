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
- `--path` **must start with `drive/Home/` or `drive/Data/`** (e.g. `drive/Home/Pictures/`). The first segment is literally `drive`; the second is `Home` or `Data` (case-sensitive). A full API URL also works: `https://files.<user>.olares.cn/api/resources/drive/Home/Pictures/`. **Not** accepted: the browser address (`.../Files/Home/...`) or a bare `Home/...` without the `drive/` prefix — both fail with `unsupported file type`. Defaults to `drive/Home/Downloads/` (aligned with the wise UI). Pass `--path ""` to send an empty path (e.g. HuggingFace cache mode) so the server decides.
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

## list / info

```bash
olares-cli knowledge download list --app wise
olares-cli knowledge download list --status downloading --page 1 --page-size 20 -o json
olares-cli knowledge download info 42
```

Table columns: `ID`, `STATUS`, `PROVIDER`, `PERCENT`, `NAME`, `APP`, `UPDATED`. Footer shows `N of total` when the server returns `total`.

## pause / resume / cancel

```bash
olares-cli knowledge download pause 42
olares-cli knowledge download resume 42
olares-cli knowledge download cancel 42
```

No body. 409 means the task is in the yt-dlp mover phase — wait and retry.

## remove

```bash
olares-cli knowledge download remove 42
olares-cli knowledge download remove 42 --remove-file
```

`--remove-file` sets `remove_flag=true` (delete artefact on PVC). Default keeps the file and only drops the task row.
