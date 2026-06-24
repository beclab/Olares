# market lifecycle (install / upgrade / uninstall / clone / stop / resume / cancel)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) (especially "App lifecycle / state machine", "OpType vs State", and "`--watch` semantics") first.
> **Flags & examples:** `olares-cli market <verb> --help` for each verb.

The mutating verb family. Every verb here returns an `OperationResult` JSON shape on `-o json`:

```json
{
  "name": "firefox",
  "op": "install",
  "accepted": true,
  "finalState": "running",
  "finalOpType": "",
  "message": "",
  "watched": true,
  "cloneTarget": ""          // only set for `clone`
}
```

## Source-aware vs source-implicit verbs

| Verb | `-s / --source` | Why |
|---|---|---|
| `install`, `upgrade`, `clone` | accepts; defaults to auto-selected source | The chart can live in different sources |
| `uninstall`, `stop`, `resume`, `cancel` | **NOT exposed** | Acts on whichever per-user state row matches the app name, regardless of source |

## `install`

```bash
olares-cli market install firefox                      # auto-selected source, latest version
olares-cli market install firefox --version 1.2.3      # pin version (strict semver)
olares-cli market install firefox -s upload            # install a locally-uploaded chart
olares-cli market install gitea --env GITEA_TOKEN=...  # required envs
olares-cli market install comfyui --compute-mode nvidia  # pin GPU mode (1.12.6+)
olares-cli market install firefox --watch              # block until terminal
olares-cli market install firefox --watch -o json
```

- `--version` defaults to the latest catalog version. Strict semver validated client-side before send.
- `--env KEY=VALUE` (repeatable) for required env vars. Missing required envs surface as `missing required env var(s): KEY1, KEY2 ...` (server returns HTTP 422 / `type=appenv`).
- **To install a locally-uploaded chart, pass `-s upload`** (the bucket `market upload` writes to).
- `--compute-mode <type>` (**Olares 1.12.6+ only**) pins the accelerator mode (`cpu`, `nvidia`, ...). Apps that can run on more than one mode require a choice: when `--compute-mode` is omitted the backend returns HTTP 422 / `type=computeModeSelect`, and the CLI either **prompts interactively** (TTY) or **fails listing the installable modes** (non-interactive: `-q`, `-o json`, or a pipe) so you re-run with the flag. On **1.12.5 the install path is unchanged** and `--compute-mode` is rejected.

## `upgrade`

```bash
olares-cli market upgrade firefox                      # latest catalog
olares-cli market upgrade firefox --version 1.5.0
olares-cli market upgrade firefox --watch
```

### Pre-flight gates (run BEFORE the PUT request)

Mirrors the SPA's `canUpgrade()`. Bails locally with a self-contained error (formatted via `failOp`, so `-o json` carries it in `.message` and `-q` still surfaces the exit code):

1. **Row exists** — state row found via `Name` or `RawName` (clones included)
2. **State is upgradable** — `running` / `stopped` / `stopFailed` / `upgradeFailed` / `applyEnvFailed`
3. **Newer chart available** — `targetVersion > installedVersion` (semver compare). **Exception for `-s upload`:** `targetVersion == installedVersion` is allowed — re-uploading the same version overwrites the stored chart, and app-service permits a same-version upgrade (it gates on `>= deployed`). This is the sanctioned way to re-apply an edited upload chart or recover an `upgradeFailed` upload app **without** bumping the version. A true downgrade (`target < installed`) is still rejected for every source.
4. **App labels don't forbid upgrade** — `disabled-upgrade`, `suspend`, `remove` labels

## `uninstall`

```bash
olares-cli market uninstall firefox                    # implicit source
olares-cli market uninstall firefox --cascade=true     # tear down shared sub-charts (C/S v2 multi-chart)
olares-cli market uninstall firefox --watch
```

### `--cascade` auto-decision (C/S v2 multi-chart apps)

The JSON payload field is `all`. Default behavior mirrors the SPA's `csAppUninstall()` dialog:

- `--cascade NOT passed`: **auto-decided**. When the cluster has a single user AND the target app is a v2 multi-chart bundle (`isCSV2`), default to `--cascade=true`; otherwise `--cascade=false`.
- A short reason is printed on **stderr** when the auto-decision flips the default to true.
- Probe errors (user count / app info) soft-fail to `--cascade=false`; the backend has the final say either way.

### Uninstalling an in-flight app (auto-orchestrated)

app-service only accepts `uninstall` from a settled state (`running` / `stopped` / a terminal `*Failed`, including `installFailed`); while an operation is in flight it accepts only `cancel`. `market uninstall` handles this for you so **`uninstall` always means "fully remove"** regardless of state:

- If the app is **in-flight** (`pending` / `downloading` / `installing` / `initializing` / `upgrading` / `applyingEnv` / `resuming`), the CLI **cancels first**, then:
  - the `pending` / `downloading` / `installing` flow cancels into a `*Canceled` state that **tears the partial install down (namespace deleted)** — equivalent to uninstalled, so the command finishes there;
  - `initializing` / `upgrading` / `applyingEnv` / `resuming` cancel only **stops** the app (lands in `stopped`), so the CLI then issues the **real uninstall** to finish removing it.
- The cancel step always blocks (it must, to decide the next step) even without `--watch`.
- `installFailed` no longer needs this dance — `uninstall` is accepted directly.

## `clone`

```bash
olares-cli market clone firefox --title "Work Browser"
olares-cli market clone firefox --title "Work Browser" --entrance-title web=WorkWeb
olares-cli market clone firefox --title "Work Browser" --watch
```

- **Only apps that advertise `cloneable: true`** in `market get <app> -o json` support this. Pre-flight check the source app first if unsure.
- `--title` is REQUIRED — feeds the cloned app's desktop shortcut title.
- For apps with multiple entrances: `--entrance-title NAME=TITLE` (repeatable) overrides per-entrance titles. For single-entrance apps, the entrance title defaults to `--title`.
- **The backend mints a per-instance app name** (e.g. `firefoxe992`). The CLI surfaces it as `cloneTarget` in the JSON output so scripted callers can chain follow-ups. **`--watch` tracks the new clone name, not the source app.**

## `stop` / `resume`

```bash
olares-cli market stop firefox                         # suspend
olares-cli market stop firefox --cascade=true          # C/S v2: shared sub-charts too
olares-cli market stop firefox --watch                 # block until `stopped`

olares-cli market resume firefox                       # un-suspend
olares-cli market resume firefox --watch               # block until `running`
olares-cli market resume comfyui --compute-binding node-1:gpu-0        # pin a device (1.12.6+)
olares-cli market resume comfyui --compute-binding node-1:gpu-0:8      # MemorySlice: 8 Gi
olares-cli market resume comfyui --compute-binding node-1:gpu-0:512Mi  # MemorySlice: 512 Mi

# Multi-GPU apps: repeat --compute-binding once per card.
olares-cli market resume vllm --compute-binding node-1:gpu-0 --compute-binding node-1:gpu-1                # two cards on one node
olares-cli market resume vllm --compute-binding node-1:gpu-0:8 --compute-binding node-1:gpu-1:8            # two MemorySlice cards, 8 Gi each
olares-cli market resume vllm --compute-binding node-1:gpu-0 --compute-binding node-2:gpu-0                # cross-node (multi-node apps only)
```

- Source is implicit on both.
- `--cascade` on `stop` follows the same auto-decision rules as `uninstall`.
- **`resume` is idempotent**: against an already-`running` row, returns immediately with success (`{state=running, opType=""}`), instead of hanging until `--watch-timeout` fires.
- `--compute-binding <node>:<device>[:<mem>]` (repeatable; **Olares 1.12.6+ only**) pins the accelerator device(s) a GPU app resumes onto; the optional `mem` is a `MemorySlice` allocation — a bare number is Gi, or add a `Gi`/`Mi` suffix (e.g. `8`, `8Gi`, `512Mi`), mirroring the SPA's two-unit VRAM input. `<node>` / `<device>` are the NODE / DEVICE-ID from `olares-cli settings compute list`. When a binding is required and the flag is omitted, the backend returns HTTP 422 / `type=computeBindingRequired` (or `computeBindingUnavailable` when a prior choice no longer fits) and the CLI **prompts interactively** (TTY) or **fails listing the available devices** (non-interactive). An explicit `--compute-binding` the backend rejects is reported with the reason rather than retried. **`stop` takes no compute flags** — the backend releases the device allocation automatically. On **1.12.5 the resume path is unchanged** and `--compute-binding` is rejected.
- **Multi-GPU apps**: pass `--compute-binding` once per card. How many cards / which nodes are allowed is decided by the app and enforced server-side (the backend reports the binding `scope`):
  - **single card** (`scope=card`): exactly one binding. Passing more is rejected with `multi-card-not-supported`.
  - **single-node multi-card** (`scope=single-node-cards`): several cards, but all on the **same** node. Spanning nodes is rejected with `multi-node-not-supported`.
  - **cross-node multi-card** (`scope=cross-node-cards`): cards may span nodes.
  - For multi-card VRAM the backend checks the **combined** VRAM of the selected cards; a shortfall is reported as `aggregate-vram-insufficient` (vs. `device-vram-insufficient` for a single card).
- **Interactive selection** (TTY, no `--compute-binding`): the CLI lists the operable devices and prompts. For a multi-card scope it accepts a **comma-separated** list (e.g. `1,2`); for a single-card scope it takes one choice. Each selected `MemorySlice` card then prompts for its allocation (Gi by default, or a `Gi`/`Mi` suffix). Non-interactive sessions (piped/`--quiet`/`-o json`) never prompt — they fail listing the available devices so you can re-run with the flag.
- **Rejection reasons mirror the SPA**: the failure text is the same wording `SelectComputeBindingDialog` shows for that backend `validation.code` — e.g. `aggregate-vram-insufficient` / `device-vram-insufficient` / `device-memory-insufficient`, and `node-pressure` additionally lists the pressured `Memory` / `CPU` / `Disk` dimensions as `Total / Used / Needed`. Structural codes the dialog can't produce (e.g. `gpu-type-mismatch`, `exclusive-already-bound`, `multi-card-not-supported`) surface the raw code.

## `cancel`

```bash
olares-cli market cancel firefox                       # cancel current op
olares-cli market cancel firefox --watch               # block until row stops moving
```

- Source is implicit.
- **The widest watcher in the tree**: any "row stopped moving" state counts as success, including `*Canceled`, `*Failed` (the underlying op died, cancel "won by default"), and stable resting states `running` / `stopped` / `uninstalled` (cancel raced and lost, OR rollback landed).
- Failure is ONLY surfaced for `*CancelFailed` (the cancel request itself was rejected).
- The terminal row carries the **underlying op** (install / upgrade / ...) as its `opType`, not `cancel`. `matchOpType` is OFF — no race-tracking gate applies.
- **Teardown vs stop**: cancel of the `pending` / `downloading` / `installing` flow **tears the partial install down (namespace deleted)** — functionally equivalent to uninstall. Cancel of `initializing` / `upgrading` / `applyingEnv` / `resuming` only **stops** the app (lands in `stopped`); the app is still installed. `market uninstall` relies on this split when auto-orchestrating (see `uninstall` above).

## `--watch` interaction with each verb

| Verb | Terminal-success buckets | Idempotent shortcut |
|---|---|---|
| `install` | `running` | App already installed → returns immediately |
| `upgrade` | `running` (matchOpType=upgrade) | — (handled by pre-flight) |
| `uninstall` | `uninstalled`, row disappears (or `*Canceled` when an in-flight app was auto-canceled) | App already uninstalled → returns immediately; in-flight apps are canceled first, then uninstalled if still present |
| `clone` | `running` on the new clone name | — |
| `stop` | `stopped` | Already stopped → returns immediately |
| `resume` | `running` | Already running → returns immediately |
| `cancel` | Any "stopped moving" state | — |

## Agent best practices

- **For "install X and tell me when it's running"** → `market install X --watch -o json`, then parse `.finalState`.
- **For "upgrade X if there's a newer version"** → `market get X -o json` to check version, then `market upgrade X --watch`. The pre-flight will short-circuit if there's no newer chart.
- **For "re-apply an upload chart I just re-uploaded"** → `market upgrade X -s upload --version <same-version> --watch`. The same-version upgrade is allowed for `-s upload` (gate 3 exception) and is the right verb once the app already exists (`running` / `upgradeFailed` / ...) — `install` would be rejected by app-service in those states.
- **For "stop everything for this user"** → `market list --mine -o json | jq -r '.[].name'` + a shell loop calling `market stop`. The cluster doesn't expose a bulk-stop verb.
- **For "install a custom chart"** → `market upload ./mychart.tgz` (always lands in source `upload`), then `market install <name> -s upload`.
- **For ambiguous source rows on uninstall/stop/resume**: the verb already resolves automatically. Don't pass `-s` even when the SPA shows it under multiple sources.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `missing required env var(s): KEY1, KEY2 ...` (install) | App declares required envs | Re-run with `--env KEY=VALUE` per missing var |
| `app 'X' is not in an upgradable state (current: Y)` | Pre-flight gate 2 | Wait for terminal state, or run `cancel` first |
| `target version '1.2.3' is already installed — nothing to do` | Pre-flight gate 3 | Nothing to upgrade. **Does NOT fire for `-s upload`** — same-version upgrade is allowed there to re-apply an overwritten chart |
| `app labels forbid upgrade (suspend / remove / disabled-upgrade)` | Pre-flight gate 4 | App is marked non-upgradable in the catalog; contact app maintainer |
| `app 'X' is not cloneable` | `clone` against non-cloneable app | Check `market get X -o json` for `cloneable` |
| `--title is required` | `clone` without `--title` | Add `--title "..."` |
| Watcher hangs near `*Failed` | Backend op failed | `market status <app>` to inspect; `market cancel <app>` if applicable |
| Stderr hint about `--cascade` auto-decision | C/S v2 single-user cluster | Informational; override with `--cascade=false` if needed |
