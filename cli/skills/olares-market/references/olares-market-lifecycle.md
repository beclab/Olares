# market lifecycle verbs (install / upgrade / uninstall / clone / stop / resume / cancel)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) (especially "App lifecycle / state machine", "OpType vs State", and "`--watch` semantics") first. **Flags & examples:** `olares-cli market <verb> --help` for each verb.

The mutating verb family. Every verb here returns an `OperationResult` JSON shape on `-o json`:

```json
{
  "app": "firefox",
  "operation": "install",
  "status": "accepted",       // "accepted" (no --watch) | "success" | "failed" (--watch verdict)
  "message": "",
  "source": "market.olares",  // omitempty
  "version": "1.2.3",         // omitempty
  "state": "running",         // omitempty; latest observed row state
  "finalState": "running",    // omitempty; set only by --watch once terminal
  "finalOpType": "",          // omitempty; set only by --watch once terminal
  "targetApp": "firefoxe992"  // omitempty; only set for `clone` (the new instance name)
}
```

> Field keys are exactly `app` / `operation` / `status` / `targetApp` / `finalState` / `finalOpType` (not `name` / `op` / `accepted` / `watched` / `cloneTarget`). Scripts parse the watch verdict from `.status` (`"success"`/`"failed"`) and the landing state from `.finalState`.

## Source-aware vs source-implicit verbs

| Verb | `-s / --source` | Why |
|---|---|---|
| `install`, `upgrade`, `clone` | accepts; defaults to auto-selected source | The chart can live in different sources |
| `uninstall`, `stop`, `resume` | **NOT exposed** | Acts on whichever per-user state row matches the app name, regardless of source |
| `cancel` | exposes `--source`, but only as a 1.12.6 fallback | Source is read from the state row; pass `--source` only when the row is gone (or `/market/state` is unreadable) and the 1.12.6 cancel body still needs one |

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
3. **Newer chart available** — `targetVersion > installedVersion` (semver compare). **Exception for `-s upload`:** `targetVersion == installedVersion` is allowed — re-uploading the same version overwrites the stored chart, and app-service permits a same-version upgrade (it gates on `>= deployed`). This is the sanctioned way to re-apply an edited upload chart or recover an `upgradeFailed` upload app **without** bumping the version. A true downgrade (`target < installed`) is still rejected for every source.
4. **Catalog row not withdrawn** — `app_simple_info.app_labels` must not contain `suspend` or `remove` (the only two labels `isAppSuspended` checks; mirrors the SPA hiding the Upgrade button). On a transient catalog-probe error this gate soft-fails (warns, lets the upgrade proceed)

### Where an upgrade lands

Two outcomes settle on `stopped` rather than `running`, and `reason` is what tells them apart (see [olares-market-watch.md](olares-market-watch.md#--watch-interaction-with-each-verb)). **Upgrading an already-`stopped` app** re-renders the chart at `replicas=0` and returns to `stopped` — a normal success with nothing to launch. **A cancelled upgrade** also settles at `stopped`, but carries `reason=upgradeCancelByUser` (or `upgradeCancelBySystem` when the backend TTL fired), stays on its **previous** version, and is reported by `--watch` as failure.

## `uninstall`

```bash
olares-cli market uninstall firefox                    # implicit source
olares-cli market uninstall firefox --cascade=true --watch  # tear down shared sub-charts too (C/S v2 multi-chart)
```

### `--cascade` (C/S v2 multi-chart apps)

The JSON payload field is `all`. Behavior depends on the backend version:

- **Olares 1.12.6+ (current):** a CS/shared app (detected from `simpleInfo`: `apiVersion=='v2' || shared`) is **always cascaded** — the backend forces `all=true` and the SPA disables the checkbox. `--cascade=false` is overridden (stderr prints `--cascade force-enabled ...`). Non-CS apps keep your value (default false).
- **Olares 1.12.5:** `--cascade NOT passed` is **auto-decided** — single-user cluster AND v2 multi-chart bundle (`isCSV2`) defaults to `--cascade=true`, else false; an explicit value wins. A short reason is printed on stderr when the auto-decision flips to true.
- Probe errors (user count / app info / simpleInfo) soft-fail to the user's value; the backend has the final say either way.

> **1.12.6 caveat — cascade-cleanup after the row is gone:** once a prior uninstall has cleared the per-user row, 1.12.6's uninstall body has no source to send, so the CLI reports an idempotent `nothing to uninstall`. `market uninstall` does **not** expose `--source`, so re-running it to tear down leftover shared sub-charts is not reachable from the CLI — clean those up from the Market SPA.

### `--delete-data`

The JSON payload field is `deleteData`. It gates **only** the app's private `drive/Data/<app>`; `cache/<node>/<app>` is cleared either way and `drive/Home` is never touched. The full per-area rule and the reason `Home` is exempt live in the platform **Userspace storage model** (loaded via this skill's prerequisite).

> **"Persistent data" is narrower than it sounds.** An app whose manifest declares only `permission.userData` (paths under `Home`) owns no app-private data, so `--delete-data` finds nothing to remove and its files survive the uninstall. That is by design, not a backend bug — clean them up with `olares-cli files rm --recursive <path>`.

On a **cascading uninstall of a v2 multi-chart app** the flag reaches only the user's own client chart. The shared sub-charts are torn down with their cache cleared but their `Data/<subChart>` left in place — `v2`'s uninstall path never consults `deleteData`. Remove those by hand if you need the storage back.

### Uninstalling an in-flight app (auto-orchestrated)

app-service only accepts `uninstall` from a settled state (`running` / `stopped` / a terminal `*Failed`, including `installFailed`); while an operation is in flight it accepts only `cancel`. `market uninstall` handles this for you so **`uninstall` always means "fully remove"** regardless of state:

- If the app is **in-flight** (`pending` / `downloading` / `installing` / `initializing` / `upgrading` / `applyingEnv` / `resuming`), the CLI **cancels first**, then follows the teardown-vs-stop split under `cancel` below: a cancel that tore the partial install down finishes the job, while one that only stopped the app is followed by the **real uninstall**.
- The cancel step always blocks (it must, to decide the next step) even without `--watch`.
- `installFailed` no longer needs this dance — `uninstall` is accepted directly.

## `clone`

```bash
olares-cli market clone firefox --title "Work Browser"
olares-cli market clone firefox --title "Work Browser" --entrance-title web=WorkWeb
olares-cli market clone comfyui --title "ComfyUI Dev" --compute-mode nvidia --watch  # pin GPU mode (1.12.6+)
```

- **Clonable apps** are either multi-instance apps (`allowMultipleInstall: true`) **or** template apps (`templateOnly: true`). A template app has no installable body — instances are created from it via clone — and on 1.12.6+ the CLI sends `templateClone:true` for it automatically. Pre-flight check the source app's `market get <app> -o json` if unsure.
- `--title` is REQUIRED — it feeds the cloned app's desktop shortcut title, and is also the default entrance title. On a multi-entrance app, `--entrance-title NAME=TITLE` (repeatable) overrides individual entrances.
- `--compute-mode <type>` (**Olares 1.12.6+ only**) works exactly like on `install`: apps runnable on more than one accelerator (`cpu`, `nvidia`, ...) require a choice, so when it is omitted the backend returns HTTP 422 / `type=computeModeSelect` and the CLI either **prompts interactively** (TTY) or **fails listing the installable modes** (non-interactive: `-q`, `-o json`, or a pipe) so you re-run with the flag. On **1.12.5 the clone path is unchanged** and `--compute-mode` is rejected.
- **The backend mints a per-instance app name** (e.g. `firefoxe992`). The CLI surfaces it as `targetApp` in the JSON output so scripted callers can chain follow-ups (`jq -r '.targetApp'`). **`--watch` tracks the new clone name, not the source app.**

## `stop` / `resume`

```bash
olares-cli market stop firefox                         # suspend
olares-cli market stop firefox --cascade=true          # C/S v2: shared sub-charts too
olares-cli market stop firefox --watch                 # block until `stopped`

olares-cli market resume firefox                       # un-suspend
olares-cli market resume firefox --watch               # block until `running`
olares-cli market resume comfyui --compute-binding node-1:gpu-0        # pin a device (1.12.6+)
olares-cli market resume comfyui --compute-binding node-1:gpu-0:512Mi  # MemorySlice: 512 Mi (bare number = Gi)
olares-cli market resume vllm --compute-binding node-1:gpu-0:8 --compute-binding node-1:gpu-1:8  # once per card
```

- Source is implicit on both.
- `--cascade` on `stop` follows the same rules as `uninstall` — including the 1.12.6 force-on for CS/shared apps (`--cascade=false` cannot disable it there).
- **`resume` is idempotent**: against an already-`running` row, returns immediately with success (`{state=running, opType=""}`), instead of hanging until `--watch-timeout` fires.
- `--compute-binding <node>:<device>[:<mem>]` (repeatable; **Olares 1.12.6+ only**) pins the accelerator device(s) a GPU app resumes onto; the optional `mem` is a `MemorySlice` allocation — a bare number is Gi, or add a `Gi`/`Mi` suffix (e.g. `8`, `8Gi`, `512Mi`), mirroring the SPA's two-unit VRAM input. `<node>` / `<device>` are the NODE / DEVICE-ID from `olares-cli settings compute list`. When a binding is required and the flag is omitted, the backend returns HTTP 422 / `type=computeBindingRequired` (or `computeBindingUnavailable` when a prior choice no longer fits) and the CLI **prompts** the operable devices (TTY — a multi-card scope accepts a comma-separated list like `1,2`, and each `MemorySlice` card then prompts for its allocation) or **fails listing them** (non-interactive: piped / `-q` / `-o json`) so you re-run with the flag. An explicit binding the backend rejects is reported with the reason rather than retried. **`stop` takes no compute flags** — the backend releases the allocation automatically. On **1.12.5 the resume path is unchanged** and `--compute-binding` is rejected.
- **Multi-GPU apps**: pass `--compute-binding` once per card. How many cards and which nodes are allowed is the app's decision, enforced server-side and reported as the binding `scope`: `scope=card` takes exactly one binding (more is `multi-card-not-supported`), `scope=single-node-cards` takes several on the **same** node (spanning is `multi-node-not-supported`), and `scope=cross-node-cards` may span nodes (`node-2:gpu-0` form).
- Multi-card VRAM is checked against the **combined** VRAM of the selected cards, so a shortfall reads `aggregate-vram-insufficient` rather than the single-card `device-vram-insufficient`.
- **Rejection reasons mirror the SPA**: the failure text is the same wording `SelectComputeBindingDialog` shows for that backend `validation.code` — e.g. `aggregate-vram-insufficient` / `device-vram-insufficient` / `device-memory-insufficient`, and `node-pressure` additionally lists the pressured `Memory` / `CPU` / `Disk` dimensions as `Total / Used / Needed`. Structural codes the dialog can't produce (e.g. `gpu-type-mismatch`, `exclusive-already-bound`, `multi-card-not-supported`) surface the raw code.

## `cancel`

```bash
olares-cli market cancel firefox                       # cancel current op
olares-cli market cancel firefox --watch               # block until row stops moving
```

- Source is normally implicit (read from the per-user state row). On **1.12.6+** the cancel body requires a source; if the row is gone (or `/market/state` is unreadable) the CLI reports an idempotent `nothing to cancel` — pass `--source <id>` to still send the request. On 1.12.5 the body needs no source, so a failed state read never blocks cancel.
- **Cancelling a `resuming` or an `upgrading` app requires Olares >= 1.12.7.** Both reuse this same `DELETE /apps/{name}/install`, and both arrived on that line: the SPA shipped the resume-cancel UX in 1.12.7, and Market's cancel state whitelist gained its `upgrading` entry there. The CLI rejects it up front on an older backend; when the version is undetectable, confirm the active profile is logged in and run `olares-cli profile list --refresh-version`. Every other in-flight state (`pending` / `downloading` / `installing` / `initializing` / `applyingEnv`) is unaffected and cancels on any backend.
- **`API error (HTTP 404): App not found or current state does not allow operation` usually means the operation already finished**, not that the app or the backend is wrong — a cancel racing a watch often lands after the install it was meant to stop. Market spells "no such app" and "nothing left to cancel" the same way; the CLI adds the app's last known state and points at `market status <app>` when it knows the app exists. Confirm where the row settled before treating it as a failure.
- A cancelled resume settles at `stopped` (it never reaches a `resumingCanceled` state — that transition does not exist); a rejected cancel request lands at `resumingCancelFailed`. A cancelled upgrade likewise settles at `stopped`, on the **previous** version, with `reason=upgradeCancelByUser`; a rejected one lands at `upgradingCancelFailed`.
- **The widest watcher in the tree**: any "row stopped moving" state counts as success, including `*Canceled`, `*Failed` (the underlying op died, cancel "won by default"), and stable resting states `running` / `stopped` / `uninstalled` (cancel raced and lost, OR rollback landed). Failure is surfaced ONLY for `*CancelFailed` — the cancel request itself was rejected.
- The terminal row carries the **underlying op** (install / upgrade / ...) as its `opType`, not `cancel`. `matchOpType` is OFF — no race-tracking gate applies.
- **Teardown vs stop**: cancel of the `pending` / `downloading` / `installing` flow **tears the partial install down (namespace deleted)** — functionally equivalent to uninstall. Cancel of `initializing` / `upgrading` / `applyingEnv` / `resuming` only **stops** the app (lands in `stopped`); the app is still installed. `market uninstall` relies on this split when auto-orchestrating (see `uninstall` above).
