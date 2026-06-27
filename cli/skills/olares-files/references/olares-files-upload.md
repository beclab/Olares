# files upload

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files upload --help`.

Resumable chunked upload of a local file or directory tree into the per-user files-backend. The most complex verb in the file tree — auto-resume, parallel files, two-stage cloud pipeline.

## Supported destinations

| Frontend path | Notes |
|---|---|
| `drive/Home/<sub>` | Olares Home volume. Upload `<node>` defaults to the first entry from `/api/nodes/` |
| `drive/Data/<sub>` | Olares Data volume. Same default node |
| `drive/Common/<sub>` | App common data area (`appCommon`). Needs Olares >= 1.12.6 |
| `sync/<repo_id>/<sub>` | Seafile library. Chunk POST hits `/seafhttp/upload-aj/<token>`; the form's `parent_dir` is the **in-repo** path (no `sync/<repo>/` prefix) |
| `cache/<node>/<sub>` | Node-local cache. `<node>` IS the upload node — CLI skips `/api/nodes/` |
| `external/<node>/<volume>/<sub>` | Attached external storage. Same `<node>` short-circuit as cache |
| `awss3/<account>/<bucket>/<sub>` | S3-compatible. **Two-stage** (see below) |
| `google/<account>/<sub>` | Google Drive. Two-stage |
| `dropbox/<account>/<sub>` | Dropbox. Two-stage |

**Tencent (`tencent/<account>/...`) is rejected up-front.** TencentDataAPI in v2 uses an octet-only `/drive/direct_upload_file/<task_id>` protocol that the CLI's chunk pipeline does not implement.

## Safety constraints

- **Destructive (writes new bytes on the server) — confirm intent with the user.** Especially when the user typed a path without trailing `/`: `upload <local> drive/Home/Documents/2026.pdf` would CREATE that exact path, possibly clobbering an existing `2026.pdf`.
- **The destination directory MUST already exist on the server.** `upload` does NOT pre-create it (because [quirk #1](../SKILL.md#1-post-apiresourcesdir-auto-renames-on-collision) would auto-rename to `Documents (1)`). Always `files mkdir -p <dest-dir>` first if the parent is new.
- **`external/<node>/` (no `<volume>`) is rejected** — virtual layer ([quirk #3](../SKILL.md#3-externalnode-is-a-virtual-volume-listing-layer-read-only)).

## Examples

```bash
# Upload one file into an existing directory.
olares-cli files upload report.pdf drive/Home/Documents/

# Upload AND rename on the server (no trailing slash → exact target path).
olares-cli files upload report.pdf drive/Home/Documents/2026-Q1.pdf

# Upload a whole directory tree (preserves the source folder name under <dest>).
olares-cli files upload ./photos drive/Home/Backups/

# Two files in flight concurrently (chunks within each file stay sequential).
olares-cli files upload ./photos drive/Home/Backups/ --parallel 2

# Upload into a Sync (Seafile) library.
olares-cli files upload notes.md sync/<repo_id>/Notes/

# Upload into node-local cache or external storage.
olares-cli files upload report.csv cache/<node>/<app>/
olares-cli files upload movie.mp4 external/<node>/hdd1/Movies/

# Upload into a connected cloud drive (two-stage).
olares-cli files upload backup.tar awss3/<account>/<bucket>/Backups/
olares-cli files upload doc.pdf google/<account>/Documents/
olares-cli files upload notes.md dropbox/<account>/Notes/
```

## Wire protocol (Drive v2 / Resumable.js-compatible)

```
1. GET /upload/upload-link/<node>/...            → upload session
2. GET /upload/file-uploaded-bytes/<node>/...    → server-driven resume offset
3. POST chunks (8 MiB default) with
        Content-Range: bytes <start>-<end>/<total>
   until the file is complete.
```

There is **no local sidecar progress file**. The resume probe asks the server "how many bytes have you received?" on each run. A Ctrl-C + re-run picks up where the server stopped accepting bytes.

## Cloud drives: stage-2 transfer task

For `awss3` / `google` / `dropbox`, chunks first land in **Olares-internal staging** (stage 1, identical to Drive). Then the server queues an "Olares-staging → cloud bucket" transfer task. The final chunk's response body contains the `taskId`:

```json
[{"taskId":"<task-id>"}]
```

The CLI polls `GET /api/task/<node>/?task_id=<id>` every 2s until the status hits:

- `completed` → success
- `failed` → error (server-supplied `failed_reason` surfaces verbatim)
- `canceled` / `cancelled` → surfaced as an error

The per-file errgroup slot stays held during stage 2 so `--parallel N` remains honest (you won't accidentally have N+M files in flight when M are mid-stage-2).

## Concurrency

- `--parallel N` (default 2): per-FILE concurrency. **Per-file chunks remain sequential by design** — the resume probe assumes one in-flight chunk per file.
- `--chunk-size <bytes>` (default 8 MiB): align with the server's expected size. Rarely needs tuning.
- `--max-retries` (default 3): per-chunk transient retry budget. The CLI's auto-refresh handles 401/403 separately (no extra retry needed).

## Token refresh on streaming chunks

`upload` uses the **pro-active** refresh path: before each chunk, the CLI decodes the access_token's JWT exp; if within 60s of expiry, it refreshes BEFORE sending. This avoids the impossible scenario where a streaming `*os.File` body is consumed by the first send and then can't be replayed on a 401. See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md#automatic-token-refresh).

## Agent notes

- **Pattern: `mkdir -p` + `upload`** for new destinations. Always.
- For huge dir trees, **start with `--parallel 1` to validate the path** and a small subset; then bump to 4-8 for throughput.
- The cloud stage-2 task can take meaningfully longer than stage-1. Don't interpret a long delay after the last chunk POST as a hang — `--parallel N` is still honest because the slot is held.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `tencent upload is not supported` | Protocol divergence | Use LarePass web app for tencent uploads |
| `is the volume listing layer (read-only)` | `external/<node>/` destination | Add `<volume>`: `external/<node>/<volume>/<sub>/` |
| `Documents (1)` materialized instead of `Documents` | Destination dir not pre-created; auto-rename (quirk #1) | Delete the dup; `files mkdir -p` next time |
| Stage-2 `failed` with `failed_reason` | Cloud-side rejection (account scopes, bucket policy, quota) | Read `failed_reason`, fix the cloud-side configuration; re-run `upload` (resumes from 0 since stage-1 already completed) |
| 401/403 mid-upload | Pre-flight refresh failed | See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) — only `profile login` will help |
