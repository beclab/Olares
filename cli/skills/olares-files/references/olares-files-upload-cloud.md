# files upload to a connected cloud drive

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md), the parent [`../SKILL.md`](../SKILL.md), and [files upload](olares-files-upload.md) — everything there about resume, concurrency and destination directories holds here too. This file is only what a cloud destination adds.

```bash
olares-cli files upload backup.tar awss3/<account>/<bucket>/Backups/
olares-cli files upload doc.pdf google/<account>/Documents/
olares-cli files upload notes.md dropbox/<account>/Notes/
```

**Tencent (`tencent/<account>/...`) is rejected up-front.** TencentDataAPI in v2 uses an octet-only `/drive/direct_upload_file/<task_id>` protocol that the CLI's chunk pipeline does not implement. Use the LarePass web app for those.

## The upload finishes in two stages

For `awss3` / `google` / `dropbox`, chunks first land in **Olares-internal staging** (stage 1, identical to Drive). Then the server queues an "Olares-staging → cloud bucket" transfer task. The final chunk's response body contains the `taskId`:

```json
[{"taskId":"<task-id>"}]
```

The CLI polls `GET /api/task/<node>/?task_id=<id>` every 2s until the status hits:

- `completed` → success
- `failed` → error (server-supplied `failed_reason` surfaces verbatim)
- `canceled` / `cancelled` → surfaced as an error

The per-file errgroup slot stays held during stage 2 so `--parallel N` remains honest (you won't accidentally have N+M files in flight when M are mid-stage-2).

That second stage can take meaningfully longer than the first, so a long delay after the last chunk POST is the transfer running, not a hang.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `tencent upload is not supported` | Protocol divergence | Use the LarePass web app for tencent uploads |
| Stage-2 `failed` with `failed_reason` | Cloud-side rejection (account scopes, bucket policy, quota) | Read `failed_reason`, fix the cloud-side configuration; re-run `upload` (resumes from 0 since stage-1 already completed) |
