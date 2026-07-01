# market upload / delete (local chart management)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli market upload --help`, `olares-cli market delete --help`.

Manage helm chart packages in the SPA's "Local Sources → Upload" bucket.

## Hard-coded source (`upload`)

Both verbs **hard-code the target source to `upload`** — the same bucket the SPA's "Local Sources → Upload" tab writes to. **`-s / --source` is intentionally NOT exposed.**

The history: an earlier revision exposed `-s` on these verbs, and users pushed charts to `cli` or `studio` that were then invisible to the SPA's Local Sources tab (different buckets, same backend). Pinning the source eliminates that footgun. To install/delete from a different bucket, use `market install -s <id>` separately.

## `upload`

```bash
olares-cli market upload ./mychart.tgz                 # single file
olares-cli market upload ./charts/                     # every chart in a directory (no recursion)
olares-cli market upload ./charts/ -o json             # structured per-file report
olares-cli market upload ./mychart.tgz -q              # exit code only
```

- Takes **exactly one path argument** — a single `.tgz` / `.tar.gz` file, or one directory. To upload several charts at once, point it at a directory; it does not accept multiple file arguments.
- Directory mode uploads every `.tgz` / `.tar.gz` directly under the directory. **Subdirectories are NOT recursed.**
- **Per-file results are summarized at the end.** `-o json` emits a structured report with one entry per file (`status` / `message`).
- **Exit code is the OR of per-file results** — any single failure flips the overall exit non-zero.
- Multipart upload through a dedicated `uploadClient` with no timeout (chart pushes can be slow over large WAN links). The same `refreshingTransport` is shared with the JSON client, so a token refresh on one is immediately visible on the other.

### After upload

Match the chart with an install / delete:

```bash
# Install (must pass -s upload):
olares-cli market install mychart -s upload --version 1.0.0

# Browse what's in the upload bucket:
olares-cli market list -s upload

# Detail of one uploaded chart:
olares-cli market get mychart -s upload -o json
```

## `delete`

```bash
olares-cli market delete mychart                       # every uploaded version
olares-cli market delete mychart --version 1.0.0       # one version
olares-cli market delete mychart -o json
olares-cli market delete mychart -q
```

- **Does NOT uninstall the app if it is running.** Use `market uninstall <app>` first, then `market delete` to also remove the chart from local sources.
- `--version` omitted → every uploaded version of the chart in the `upload` bucket is removed.

> The "delete the chart" and "uninstall the running app" are deliberately separate verbs. A chart can be uploaded without ever being installed; an installed app can keep running after the source chart is deleted from the bucket.

## Agent workflows

```bash
# Full custom-chart cycle.
olares-cli market upload ./mychart-1.0.0.tgz                       # land it in source=upload
olares-cli market install mychart -s upload --version 1.0.0 --watch
# ... use the app ...
olares-cli market uninstall mychart --watch                        # tear down the deployment
olares-cli market delete mychart --version 1.0.0                   # remove the chart from local sources
```

```bash
# Bulk-upload every chart in a release directory.
olares-cli market upload ./dist/ -o json | jq '.[] | select(.status != "success")'
# JSON exit code is the OR of all per-file results; the jq filter surfaces the failures
```

```bash
# Spring cleaning: remove every version of a chart from the upload bucket.
olares-cli market delete mychart                                   # all versions
olares-cli market list -s upload                                   # confirm
```

## Safety constraints

- **`delete` is destructive** — it removes the chart from the bucket. If the app is still running, the deployment continues to work but you can no longer reinstall from the local bucket.
- **`upload` overwrites by `(name, version)`** — uploading `mychart-1.0.0.tgz` twice replaces the previous bytes. The uploaded version must be **>= the stored** version (equal overwrites; a *lower* version is rejected). To bump, change the version inside `Chart.yaml` and re-upload.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `unsupported file extension: must be .tgz or .tar.gz` | Wrong file type | Repackage with `helm package` |
| `failed to upload: HTTP 413 (Payload Too Large)` | Chart exceeds the server's upload size limit | Slim the chart's contents; ask the operator about the limit |
| `chart not found in source 'upload'` (delete) | The chart was never uploaded, or was uploaded to a different bucket | `market list -s upload` to confirm |
| `delete` removed the chart but the app keeps running | `delete` only removes the chart from the `upload` bucket; it never uninstalls | Expected — run `market uninstall X` separately to stop/remove the app |
| Exit non-zero on directory upload despite some files succeeding | Partial failure | Inspect the per-file JSON report for which files failed |
