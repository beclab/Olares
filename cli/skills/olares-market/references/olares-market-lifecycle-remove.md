# market lifecycle: taking an app off or pausing it (uninstall / stop / resume / cancel)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md), especially [App lifecycle / state machine](../SKILL.md#app-lifecycle--state-machine), first. **Flags & examples:** `olares-cli market <verb> --help` for each verb. Installing, upgrading and cloning are [the other half](olares-market-lifecycle-add.md).

These verbs act on an app that is already here, so none of them takes `-s / --source`: the target is whichever per-user state row matches the name, regardless of which source it came from. `cancel` is the one exception, and only as a fallback — it reads the source from the state row, and accepts `--source` only when the row is gone (or `/market/state` is unreadable) and the 1.12.6 cancel body still needs one.

Each verb prints one `OperationResult` document on `-o json`; `olares-cli market <verb> --help` lists its fields. The three below are coupled: `uninstall` orchestrates `cancel`, and `stop --cascade` follows `uninstall`'s rules.

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

app-service only accepts `uninstall` from a settled state (`running` / `stopped` / `installFailed` / `upgradeFailed` and the other post-install `*Failed` states); while an operation is in flight it accepts only `cancel`. `market uninstall` handles the in-flight case for you:

- If the app is **in-flight** (`pending` / `downloading` / `installing` / `initializing` / `upgrading` / `applyingEnv` / `resuming`), the CLI **cancels first**, then follows the teardown-vs-stop split under `cancel` below: a cancel that tore the partial install down finishes the job, while one that only stopped the app is followed by the **real uninstall**.
- The cancel step always blocks (it must, to decide the next step) even without `--watch`.
- `installFailed` no longer needs this dance — `uninstall` is accepted directly.

**There are settled states `uninstall` cannot act on at all.**
`downloadFailed`, `uninstalled` and the `pendingCanceled` /
`downloadingCanceled` / `installingCanceled` trio accept only a fresh
`install`, so an uninstall there is answered with a 404. Nothing was ever
deployed; the row is a record of the operation that failed.

`downloadFailed` is the one to expect, and the reason this is easy to get
wrong: **a bad image reference lands in `downloadFailed`, not
`installFailed`**, because the pull fails before the install begins. Most
prose about broken apps says `installFailed`, which does accept an
uninstall.

The row stays in `market status` either way. What clears it depends on where
the chart came from:

```bash
olares-cli market delete <app>     # upload-source apps: removes the chart, and the row with it
olares-cli market install <app>    # catalog apps: retry; a success replaces the row
```

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

- `--cascade` on `stop` follows the same rules as `uninstall` above — including the 1.12.6 force-on for CS/shared apps (`--cascade=false` cannot disable it there).
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

- On **1.12.6+** the cancel body requires a source; if the row is gone (or `/market/state` is unreadable) the CLI reports an idempotent `nothing to cancel` — pass `--source <id>` to still send the request. On 1.12.5 the body needs no source, so a failed state read never blocks cancel.
- **Cancelling a `resuming` or an `upgrading` app requires Olares >= 1.12.7.** Both reuse this same `DELETE /apps/{name}/install`, and both arrived on that line: the SPA shipped the resume-cancel UX in 1.12.7, and Market's cancel state whitelist gained its `upgrading` entry there. The CLI rejects it up front on an older backend; when the version is undetectable, confirm the active profile is logged in and run `olares-cli profile list --refresh-version`. Every other in-flight state (`pending` / `downloading` / `installing` / `initializing` / `applyingEnv`) is unaffected and cancels on any backend.
- **`API error (HTTP 404): App not found or current state does not allow operation` usually means the operation already finished**, not that the app or the backend is wrong — a cancel racing a watch often lands after the install it was meant to stop. Market spells "no such app" and "nothing left to cancel" the same way; the CLI adds the app's last known state and points at `market status <app>` when it knows the app exists. Confirm where the row settled before treating it as a failure.
- A cancelled resume settles at `stopped` (it never reaches a `resumingCanceled` state — that transition does not exist); a rejected cancel request lands at `resumingCancelFailed`. A cancelled upgrade likewise settles at `stopped`, on the **previous** version, with `reason=upgradeCancelByUser`; a rejected one lands at `upgradingCancelFailed`.
- **The widest watcher in the tree**: any "row stopped moving" state counts as success, including `*Canceled`, `*Failed` (the underlying op died, cancel "won by default"), and stable resting states `running` / `stopped` / `uninstalled` (cancel raced and lost, OR rollback landed). Failure is surfaced ONLY for `*CancelFailed` — the cancel request itself was rejected.
- The terminal row's `opType` is **`cancel`**, not the op that was cancelled — observed on app-service 0.6.49 for both a cancelled `upgrade` and a cancelled `download`. Either way, do not gate on it: `matchOpType` is OFF for this watcher, so no race-tracking gate applies, and the state is what tells you where the app landed.
- **Teardown vs stop**: cancel of the `pending` / `downloading` / `installing` flow **tears the partial install down (namespace deleted)** — functionally equivalent to uninstall. Cancel of `initializing` / `upgrading` / `applyingEnv` / `resuming` only **stops** the app (lands in `stopped`); the app is still installed. `market uninstall` relies on this split when auto-orchestrating (see `uninstall` above).
