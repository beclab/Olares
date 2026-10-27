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

## Re-running does not retry stage 2

The `taskId` only ever rides on the **last chunk's** response, and the resume probe is what decides whether a chunk gets sent. Once stage 1 has completed, that probe reports the whole file, so a re-run sends nothing, never sees a `taskId`, and prints `done:` — a success line for a file that is still sitting in Olares-internal staging and is not in the bucket. The web app has the same hole.

So a cloud upload is only finished if its output carried a `cloud transfer queued (task=…)` line followed by `cloud transfer completed`. A run that goes straight from `✓` to `done:` on an `awss3` / `google` / `dropbox` destination did not run stage 2 at all.

To actually retry, give the pipeline something it has not staged: upload to a **different remote name or directory**, which makes the probe report zero and runs both stages. Then confirm with `olares-cli files ls <cloud-dir> -o json` that the object is really there — the staging copy is invisible from the cloud namespace, so this listing is the only honest check.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `tencent upload is not supported` | Protocol divergence | Use the LarePass web app for tencent uploads |
| Stage-2 `failed` with `failed_reason` | Cloud-side rejection (account scopes, bucket policy, quota) | Read `failed_reason` and fix the cloud-side configuration, then re-upload under a different name — see above, the same name is a no-op that reports success |
