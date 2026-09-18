# market lifecycle: putting an app on the machine (install / upgrade / clone)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md), especially [App lifecycle / state machine](../SKILL.md#app-lifecycle--state-machine), first. **Flags & examples:** `olares-cli market <verb> --help` for each verb. Taking an app off or pausing it is [the other half](olares-market-lifecycle-remove.md).

These three verbs all put a chart somewhere it was not, so all three take `-s / --source` (defaulting to an auto-selected source) and all three take `--compute-mode` on an app that can run on more than one accelerator. Each prints one `OperationResult` document on `-o json`; `olares-cli market <verb> --help` lists its fields.

## `install`

```bash
olares-cli market install firefox                      # auto-selected source, latest version
olares-cli market install firefox --version 1.2.3      # pin version (strict semver)
olares-cli market install firefox -s upload            # install a locally-uploaded chart
olares-cli market install gitea --env GITEA_TOKEN=...  # required envs
olares-cli market install comfyui --compute-mode nvidia  # pin GPU mode (1.12.6+)
olares-cli market install firefox --watch              # block until terminal (add -o json for scripts)
```

- `--version` defaults to the latest catalog version. Strict semver validated client-side before send.
- `--env KEY=VALUE` (repeatable) for required env vars. The server returns HTTP 422 / `type=appenv`, which the CLI renders as `environment variable requirements not met` followed by a `Missing required values: KEY1, KEY2` line and the `market get <app>` command that lists the declared envs. Match on the `Missing required values:` line, not on the older one-line `missing required env var(s): ...` phrasing.
- **To install a locally-uploaded chart, pass `-s upload`** (the bucket `market upload` writes to).
- `--compute-mode <type>` (**Olares 1.12.6+ only**) pins the accelerator mode (`cpu`, `nvidia`, ...). Apps that can run on more than one mode require a choice: when `--compute-mode` is omitted the backend returns HTTP 422 / `type=computeModeSelect`, and the CLI either **prompts interactively** (TTY) or **fails listing the installable modes** (non-interactive: `-q`, `-o json`, or a pipe) so you re-run with the flag. On **1.12.5 the install path is unchanged** and `--compute-mode` is rejected.

## `upgrade`

```bash
olares-cli market upgrade firefox                      # latest catalog
olares-cli market upgrade firefox --version 1.5.0 --watch
```

### Pre-flight gates (run BEFORE the PUT request)

Mirrors the SPA's `canUpgrade()`. Bails locally with a self-contained error (formatted via `failOp`, so `-o json` carries it in `.message` and `-q` still surfaces the exit code):

1. **Row exists** — state row found via `Name` or `RawName` (clones included)
2. **State is upgradable** — `running` / `stopped` / `stopFailed` / `upgradeFailed` / `applyEnvFailed`
3. **Newer chart available** — `targetVersion > installedVersion` (semver compare). **Exception for `-s upload`:** `targetVersion == installedVersion` is allowed, because app-service gates on `>= deployed`. Use it to retry an upgrade that failed for a transient reason — but read [which chart a version actually deploys](#which-chart-a-version-actually-deploys) before relying on which bytes land. It cannot deploy an edited chart either way: a published version's bytes are immutable and re-uploading one is refused (see [the immutability rule](olares-market-chart-publish.md#safety-constraints)), so recovering an `upgradeFailed` app with a *fixed* chart means bumping the version. A true downgrade (`target < installed`) is still rejected for every source.
4. **Catalog row not withdrawn** — `app_simple_info.app_labels` must not contain `suspend` or `remove` (the only two labels `isAppSuspended` checks; mirrors the SPA hiding the Upgrade button). On a transient catalog-probe error this gate soft-fails (warns, lets the upgrade proceed)

### Which chart a version actually deploys

Known app-service defect, present through 1.12.7. **`upgrade` deploys the newest chart in the source, not the version the request named.** The install path pins the version when it fetches the chart; the upgrade path omits it, so the fetch resolves the index's latest and unpacks it over the version-less chart cache. The response and the state row both report the version you asked for, which is what makes this hard to catch.

In practice the upload bucket makes this hard to hit, and that is worth knowing before trying to work around it: **it catalogs one version per app**, and `upload` requires a strictly higher version than the one already there. So "the newest chart in the source" and "the version you named" are normally the same chart, and the defect only shows up where a source really does index several versions.

What follows from the single-version bucket, rather than from the defect:

- A same-version `upgrade` (gate 3's exception) re-applies the only version the bucket has, which is the one you named. That makes it a usable retry right after an upload.
- There is no pair of stored versions in the bucket to move between. An earlier version is gone from the catalog once a higher one is uploaded, so `install --version <older>` answers 404 rather than reinstalling it. Recovering an older build means repackaging it under a higher version number.

A cancelled upgrade compounds it: cancel does no helm rollback, so the release keeps the new chart's templates merged with the old values. After A1 is fixed those templates are at least the version that was requested.

### Where an upgrade lands

Two outcomes settle on `stopped` rather than `running`. **Upgrading an already-`stopped` app** re-renders the chart at `replicas=0` and returns to `stopped` — a normal success with nothing to launch. **A cancelled upgrade** also settles at `stopped`, and `--watch` reports it as failure. What separates them is `status.reason` matching `upgradeCancelByUser` or `upgradeCancelBySystem` — not whether `reason` is set, which it always is. *Non-obvious terminal behaviors* in the shared **application state machine** has the reasoning, and why the row's version field cannot discriminate either.

That version field is the one `market status` and `market list --mine` both report, and it is **the version last requested, not the one running**. After a failed or cancelled upgrade the row names the target that did not land, so a script that reads it as "what is deployed" is wrong exactly when it matters. The deployed version is on the workload, in the `applications.app.bytetrade.io/version` annotation:

```bash
olares-cli cluster workload get <app> -n user-space-<user> --kind deployment -o json
```

`--kind` is required (`deployment` | `statefulset` | `daemonset`); there is no `cluster deployment` verb. This is the only way to see what is actually deployed rather than what was last asked for, so it is worth confirming the annotation and the container image agree before concluding a version landed.

## `clone`

```bash
olares-cli market clone firefox --title "Work Browser"
olares-cli market clone firefox --title "Work Browser" --entrance-title web=WorkWeb
olares-cli market clone comfyui --title "ComfyUI Dev" --compute-mode nvidia --watch  # pin GPU mode (1.12.6+)
```

- **Clonable apps** are either multi-instance apps (`allowMultipleInstall: true`) **or** template apps (`templateOnly: true`). A template app has no installable body — instances are created from it via clone — and on 1.12.6+ the CLI sends `templateClone:true` for it automatically. Pre-flight check the source app's `market get <app> -o json` if unsure.
- `--title` is REQUIRED — it feeds the cloned app's desktop shortcut title, and is also the default entrance title. On a multi-entrance app, `--entrance-title NAME=TITLE` (repeatable) overrides individual entrances.
- `--compute-mode <type>` (**Olares 1.12.6+ only**) works exactly like on `install`: apps runnable on more than one accelerator (`cpu`, `nvidia`, ...) require a choice, so when it is omitted the backend returns HTTP 422 / `type=computeModeSelect` and the CLI either **prompts interactively** (TTY) or **fails listing the installable modes** (non-interactive: `-q`, `-o json`, or a pipe) so you re-run with the flag. On **1.12.5 the clone path is unchanged** and `--compute-mode` is rejected.
- **The backend mints a per-instance app name** (e.g. `firefoxe992`). The CLI surfaces it as `targetApp` in the JSON output so scripted callers can chain follow-ups (`jq -r '.targetApp'`), and it is `.targetApp` rather than `.app` for this verb alone. **`--watch` tracks the new clone name, not the source app.**
