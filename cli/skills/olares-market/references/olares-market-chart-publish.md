# market upload / delete (publishing a chart to Local Sources)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli market upload --help`, `olares-cli market delete --help`. Pulling a stored chart back out is [market download](olares-market-chart-download.md).

Push a helm chart package into the SPA's "Local Sources → Upload" bucket, and take it back off.

## Hard-coded source

`upload` and `delete` **hard-code the target source to `upload`** — the same bucket the SPA's "Local Sources → Upload" tab writes to. **`-s / --source` is intentionally NOT exposed on those two.**

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

### What is finished when upload returns

The response covers rendering and the catalog write — once `upload` succeeds the chart is installable. **Image analysis is the one part that continues in the background**, filling in each image's size and platform on the catalog entry.

Treat that field as decoration, not as a gate. An image the registry cannot answer for leaves the placeholder sitting in `analyzing` past the backend's own refine deadline, and a failed refine pushes no notification, so polling for a terminal image status can wait forever on a chart that is already perfectly usable. If you need to block on something after upload, block on the install.

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
olares-cli market delete mychart                       # remove the app and every uploaded version
olares-cli market delete mychart --version 1.0.0       # same result: all versions go
olares-cli market delete mychart -o json
olares-cli market delete mychart -q
```

- **Does NOT uninstall the app if it is running.** Use `market uninstall <app>` first, then `market delete` to also remove the chart from local sources.
- **`--version` does not narrow the delete.** The backend takes the app name and drops every version and every stored artifact; the version only names the request. There is no single-version delete today, so read this verb as "unpublish the app", and expect `market download mychart --version <any>` to stop working for all versions afterwards, not just the one named.

> The "delete the chart" and "uninstall the running app" are deliberately separate verbs. A chart can be uploaded without ever being installed; an installed app can keep running after the source chart is deleted from the bucket.

## Agent workflows

```bash
# Full custom-chart cycle.
olares-cli market upload ./mychart-1.0.0.tgz                       # land it in source=upload
olares-cli market install mychart -s upload --version 1.0.0 --watch
# ... use the app ...
olares-cli market uninstall mychart --watch                        # tear down the deployment
olares-cli market delete mychart                                   # unpublish the chart (all versions)
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
- **A published version's bytes are immutable.** `upload` requires a **strictly higher** version than the stored one; re-uploading a version that already exists is refused with HTTP 409 `version <v> already exists for app <name> in source upload; bump the version to publish changes`, and a lower version is refused too. To ship a change, bump the version inside `Chart.yaml` and upload that.

  Most of what is confusing about versions here follows from that one rule. A same-version `upgrade` is legal, but since the stored bytes cannot have changed it re-applies the *same* chart — it is a retry, not a way to deploy an edit. Recovering an `upgradeFailed` app with a *fixed* chart therefore needs a new version, not a re-upload of the old one. And `delete` frees the version only by removing the entire app (see above), so it is not a way to republish one release.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `unsupported file extension: must be .tgz or .tar.gz` | Wrong file type | Repackage with `helm package` |
| `failed to upload: HTTP 413 (Payload Too Large)` | Chart exceeds the server's upload size limit | Slim the chart's contents; ask the operator about the limit |
| `upload rejected: manifest supports [...] cluster provides [...]` | `spec.supportArch` has no intersection with the current cluster nodes | Check `olares-cli cluster node list`, fix `supportArch` and image platforms, then repackage; do not bump or retry the unchanged package |
| `upload blocked: cluster node discovery is unavailable` | Market cannot yet establish an authoritative node architecture set | Keep the package and version unchanged; retry after node discovery recovers |
| `chart not found in source 'upload'` (delete) | The chart was never uploaded, or was uploaded to a different bucket | `market list -s upload` to confirm |
| `delete` removed the chart but the app keeps running | `delete` only removes the chart from the `upload` bucket; it never uninstalls | Expected — run `market uninstall X` separately to stop/remove the app |
| Exit non-zero on directory upload despite some files succeeding | Partial failure | Inspect the per-file JSON report for which files failed |
