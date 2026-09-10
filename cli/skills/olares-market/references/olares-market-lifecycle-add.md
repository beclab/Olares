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
- `--env KEY=VALUE` (repeatable) for required env vars. Missing required envs surface as `missing required env var(s): KEY1, KEY2 ...` (server returns HTTP 422 / `type=appenv`).
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
3. **Newer chart available** — `targetVersion > installedVersion` (semver compare). **Exception for `-s upload`:** `targetVersion == installedVersion` is allowed, because app-service gates on `>= deployed`. It re-applies the **stored** chart at that version, so use it to retry an upgrade that failed for a transient reason. It cannot deploy an edited chart: a published version's bytes are immutable and re-uploading one is refused (see [the immutability rule](olares-market-chart-publish.md#safety-constraints)), so recovering an `upgradeFailed` app with a *fixed* chart means bumping the version. A true downgrade (`target < installed`) is still rejected for every source.
4. **Catalog row not withdrawn** — `app_simple_info.app_labels` must not contain `suspend` or `remove` (the only two labels `isAppSuspended` checks; mirrors the SPA hiding the Upgrade button). On a transient catalog-probe error this gate soft-fails (warns, lets the upgrade proceed)

### Where an upgrade lands

Two outcomes settle on `stopped` rather than `running`. **Upgrading an already-`stopped` app** re-renders the chart at `replicas=0` and returns to `stopped` — a normal success with nothing to launch. **A cancelled upgrade** also settles at `stopped`, and `--watch` reports it as failure. What separates them is `status.reason` matching `upgradeCancelByUser` or `upgradeCancelBySystem` — not whether `reason` is set, which it always is. *Non-obvious terminal behaviors* in the shared **application state machine** has the reasoning, and why the row's version field cannot discriminate either.

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
