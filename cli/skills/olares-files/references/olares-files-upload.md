# files upload

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files upload --help`.

Resumable chunked upload of a local file or directory tree into the per-user files-backend. The most complex verb in the file tree — auto-resume and parallel files. A cloud destination adds a second transfer stage on top of all of this: [uploading to a connected cloud drive](olares-files-upload-cloud.md).

## Supported destinations

| Frontend path | Notes |
|---|---|
| `drive/Home/<sub>` | Olares Home volume. Upload `<node>` defaults to the first entry from `/api/nodes/`; `olares-cli cluster node list` names them |
| `drive/Data/<sub>` | Olares Data volume. Same default node |
| `drive/Common/<sub>` | App common data area (`appCommon`). Needs Olares >= 1.12.6 |
| `sync/<repo_id>/<sub>` | Seafile library. Chunk POST hits `/seafhttp/upload-aj/<token>`; the form's `parent_dir` is the **in-repo** path (no `sync/<repo>/` prefix) |
| `cache/<node>/<sub>` | Node-local cache. `<node>` IS the upload node — CLI skips `/api/nodes/` |
| `external/<node>/<volume>/<sub>` | Attached external storage. Same `<node>` short-circuit as cache |
| `awss3` / `google` / `dropbox` | Two-stage; `tencent` is rejected — [uploading to a connected cloud drive](olares-files-upload-cloud.md) |

## Safety constraints

- **Writes new bytes on the server.** The requested upload authorises the named destination, but a path without trailing `/` creates that exact path and may clobber an existing file; ask when overwrite intent is not explicit.
- **The destination directory MUST already exist on the server.** `upload` does NOT pre-create it because [directory creation may auto-rename on collision](../SKILL.md#backend-quirks-that-change-decisions). Always `files mkdir -p <dest-dir>` first if the parent is new.
- **`external/<node>/` (no `<volume>`) is rejected** — see [backend quirks](../SKILL.md#backend-quirks-that-change-decisions).

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

## Concurrency

- `--parallel N` (default 2): per-FILE concurrency. **Per-file chunks remain sequential by design** — the resume probe assumes one in-flight chunk per file.
- `--chunk-size <bytes>` (default 8 MiB): align with the server's expected size. Rarely needs tuning.
- `--max-retries` (default 3): per-chunk transient retry budget. The CLI's auto-refresh handles 401/403 separately (no extra retry needed).

## Token refresh on streaming chunks

`upload` uses the **pro-active** refresh path: before each chunk, the CLI decodes the access_token's JWT exp; if within 60s of expiry, it refreshes BEFORE sending. This avoids the impossible scenario where a streaming `*os.File` body is consumed by the first send and then can't be replayed on a 401. The general refresh behaviour is in [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md), under the auth model.

## Agent notes

- **Pattern: `mkdir -p` + `upload`** for new destinations. Always.
- For huge dir trees, **start with `--parallel 1` to validate the path** and a small subset; then bump to 4-8 for throughput.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `is the volume listing layer (read-only)` | `external/<node>/` destination | Add `<volume>`: `external/<node>/<volume>/<sub>/` |
| `Documents (1)` materialized instead of `Documents` | Destination dir not pre-created; auto-rename (quirk #1) | Delete the dup; `files mkdir -p` next time |
| 401/403 mid-upload | Pre-flight refresh failed (streaming body can't be replayed) | Apply the [auth-readiness gate](../../olares-shared/SKILL.md#auth-readiness-gate): `ErrTokenInvalidated` / `ErrNotLoggedIn` → `profile login`; a transient refresh failure → just re-run `upload` |
