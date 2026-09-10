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
- **Exit code is the OR of per-file results** — any single failure flips the overall exit non-zero, in every output mode. `-o json` still prints the full report on the way out, so a script reads the exit code and the report, not one or the other.
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

- **Refused while the app is installed.** The backend rejects the delete with `app (or its clone <name>) is still installing/running; please uninstall it first` rather than unpublishing a chart something is running from. Run `market uninstall <app>` first, then `market delete`.
- **`--version` does not narrow the delete.** The backend takes the app name and drops every version and every stored artifact; the version only names the request. There is no single-version delete today, so read this verb as "unpublish the app", and expect `market download mychart --version <any>` to stop working for all versions afterwards, not just the one named.
- **Deleting an app the bucket never held returns success.** The backend treats it as already done and answers HTTP 200; `deleted_rows` in the response is what distinguishes the two, and the CLI does not read it. Without `--version` the CLI's own version lookup fails first and you get a correct non-zero exit — so do **not** add `--version` to get past that failure, which converts it into a false success.

> The "delete the chart" and "uninstall the running app" are deliberately separate verbs, in that order: a chart can be uploaded without ever being installed, but it cannot be unpublished out from under a running app.

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

## Safety constraints

- **`delete` is destructive** — it unpublishes every version of the app from the bucket, so nothing can be installed or downloaded from it afterwards. It is refused while the app is installed, so it cannot strand a running deployment.
- **A published version's bytes are immutable.** `upload` requires a **strictly higher** version than the stored one. Re-uploading the stored version is refused with HTTP 409 `version <v> already exists for app <name> in source upload; bump the version to publish changes`; a **lower** version is refused with `version <v> does not supersede version <stored> already published for app <name> in source upload; upload a version higher than <stored>`. The two messages are worth telling apart, because the second one names a version you did not send. To ship a change, bump the version inside `Chart.yaml` and upload that.

  Most of what is confusing about versions here follows from that one rule. A same-version `upgrade` is legal — but see [which chart a version actually deploys](olares-market-lifecycle-add.md#which-chart-a-version-actually-deploys), which is not always the version you named. Recovering an `upgradeFailed` app with a *fixed* chart needs a new version, not a re-upload of the old one. And `delete` frees the version only by removing the entire app (see above), so it is not a way to republish one release.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `unsupported file format: expected .tgz or .tar.gz` | Wrong file type; refused locally before any request | Repackage with `helm package` |
| `failed to upload: HTTP 413 (Payload Too Large)` | Chart exceeds the server's upload size limit | Slim the chart's contents; ask the operator about the limit |
| `HTTP 400: Uploaded package could not be read as a chart archive` | The bytes are not a readable gzipped tar (truncated or partial transfer) | Repackage and re-upload; retrying the same bytes cannot help |
| `HTTP 400: Uploaded package contains no chart directory with an OlaresManifest.yaml` | The archive holds the chart's *contents* rather than the chart directory | Repackage from the parent directory (`helm package ./mychart`) |
| `HTTP 400: Invalid OlaresManifest.yaml: <field / rule>` | The manifest was found and rejected; the detail is the same verdict `olares-cli chart lint` prints | Fix the named field and repackage. Run `chart lint` first to see it without an upload |
| `HTTP 400: Failed to process uploaded package` | The catch-all: the failure was **not** in the uploaded bytes, so nothing about it is safe to report to the caller | Ask the operator for the market log line; do not repackage on a guess |
| `upload rejected: manifest supports [...] cluster provides [...]` | `spec.supportArch` has no intersection with the current cluster nodes | Check `olares-cli cluster node list`, fix `supportArch` and image platforms, then repackage; do not bump or retry the unchanged package |
| `upload blocked: cluster node discovery is unavailable` | Market cannot yet establish an authoritative node architecture set | Keep the package and version unchanged; retry after node discovery recovers |
| `app 'X' not found in source 'upload'` (delete) | The chart was never uploaded, or was uploaded to a different bucket | `market list -s upload` to confirm. Do **not** add `--version` — that skips the lookup and returns a false success |
| `app (or its clone X) is still installing/running; please uninstall it first` (delete) | The app is installed, so the chart cannot be unpublished | `market uninstall X --watch`, then `market delete X` |
| `unknown market source 'X': this cluster has ...` | Typo in `-s`, or a source this cluster is not subscribed to | Use one of the listed ids; `market list -a` enumerates them |
| Exit non-zero on directory upload despite some files succeeding | Partial failure | Inspect the per-file JSON report for which files failed |
