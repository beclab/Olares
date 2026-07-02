---
name: olares-market
version: 4.5.0
description: "Olares Market via olares-cli market — install, upgrade, uninstall, clone, stop, resume, restart apps; catalog, status, chart upload, --watch. Use for Olares app store, my apps, 我的应用, install app, restart app, upload chart."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# market (App-store v2)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli market --help`.**

> **Source of truth for flags is always `olares-cli market <verb> --help`.** This file only carries what `--help` cannot give: source resolution, the CLI's OpType-vs-State view of the lifecycle (the state machine itself lives in the appstate reference), OpType-vs-State race safety, the verb index, the `-s`/`-a` matrix, and the "what apps do I have" routing.

> **Platform model:** app namespaces (`<app>-<owner>` vs the admin-only `<app>-shared`) and which system-middleware apps an admin installs are defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md#app-namespace--networking-model).

## When to use

- Olares Market, olares-cli market, Olares app store, install / upgrade / uninstall / clone / stop / resume / restart / cancel an app
- `my Olares apps`, `我的应用`, `market list --mine`, `is <app> installed yet`, chart upload / delete, `--watch` until terminal state
- Catalog: `list`, `get`, `categories`; runtime: `status`

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** `market` is **lifecycle and inventory** at the app-store level (install / upgrade / chart push). For runtime K8s objects, settings, or metrics, route to a sibling.

## Verb families

| Family | Verbs | Mutating? |
|---|---|---|
| **catalog** | `list`, `get`, `categories` | no |
| **runtime** | `status` | no |
| **lifecycle** | `install`, `upgrade`, `uninstall`, `clone`, `stop`, `resume`, `restart`, `cancel` | yes |
| **charts** | `upload`, `delete` | yes |

For verb-specific behavior, **always start with `olares-cli market <verb> --help`**. Then drill into a reference if listed:

| Family | Reference |
|---|---|
| catalog + runtime | [references/olares-market-list.md](references/olares-market-list.md) (`list` / `--mine` / `categories` / `get` / `status`) |
| lifecycle | [references/olares-market-lifecycle.md](references/olares-market-lifecycle.md) (`install` / `upgrade` / `uninstall` / `clone` / `stop` / `resume` / `cancel`). `restart` (POST `/apps/restart`, version-agnostic body `{app_name, source}`) shares the `resume` watch bucket (`running`) and, like `stop`/`resume`, exposes no `-s`; it also backs the auto-restart in `settings network overlay app enable/disable` |
| `--watch` / stuck / errors | [references/olares-market-watch.md](references/olares-market-watch.md) (per-verb watch buckets, foreground windows, stuck-state handling, common errors) |
| charts | [references/olares-market-charts.md](references/olares-market-charts.md) (`upload` / `delete`) |

## Source resolution (cross-cutting)

The market backend serves multiple "sources" of charts. The CLI resolves which one to talk to from `-s / --source`, falling back to a default that depends on the verb:

| Source id | What it is | Used by |
|---|---|---|
| `market.olares` | Public catalog (read-only browse) | default for `list`, `get`, `categories`, `install`, `upgrade`, `clone`, `status` |
| `upload` | SPA "Local Sources → Upload" bucket | **hard-coded for `upload` / `delete`** — `-s` is intentionally NOT exposed on those two verbs |
| `cli` | Legacy CLI-upload bucket | read-only (`list`, `status`) |
| `studio` | Devbox / Studio bucket | read-only (`list`, `status`) |

- When `-s` is omitted, every verb that accepts it prints `Using source: <id>` to stderr so the agent can confirm which backend was hit.
- `-a / --all-sources` bypasses single-source resolution and spans every source the user has — read-only verbs only.
- **Unknown source ids silently produce an empty result** with a `no apps in source 'X'` stderr hint. Run `market list -a` to enumerate the configured sources.

### `-s` / `-a` matrix

| Flag | Read-only browse | Lifecycle (mutating) | Chart management |
|---|---|---|---|
| `-s / --source` | `list`, `categories`, `status`, `get` | `install`, `upgrade`, `clone` | — (hard-coded `upload`) |
| `-a / --all-sources` | `list`, `categories`, `status` | — | — |

> **`-s` is NOT on `uninstall` / `stop` / `resume` / `restart`:** they act on whichever per-user state row matches the app name, regardless of source.
> **`cancel` now exposes `--source`** — but only as a fallback for the Olares 1.12.6+ edge case where the per-user state row is gone (or `/market/state` is unreadable) and the 1.12.6 cancel body still needs a source. In the normal case the source is read from the state row; do not pass it.

## App lifecycle / state machine

> **Backend facts (states, legal transitions, per-state allowed operations, fail TTLs, single-download serialization, `running` semantics, progress) live once in [`../olares-shared/references/olares-platform-appstate.md`](../olares-shared/references/olares-platform-appstate.md).** This section is only the CLI's view of them. Don't re-derive the state machine here.

The backend tracks two orthogonal axes per app: **`State`** (where the row currently is) and **`OpType`** (which mutation is in flight). The CLI groups the full enum into four buckets:

| Bucket | Examples | Meaning |
|---|---|---|
| **Progressing** | `pending`, `downloading`, `installing`, `initializing`, `applyingEnv`, `upgrading`, `uninstalling`, `stopping`, `resuming`, `*Canceling` | Backend is actively working — keep polling |
| **Terminal success** | `running`, `stopped`, `uninstalled` | Mutation finished cleanly |
| **Terminal failure** | `downloadFailed`, `installFailed`, `applyEnvFailed`, `upgradeFailed`, `uninstallFailed`, `stopFailed`, `resumeFailed` | Mutation finished with a hard error |
| **Canceled / cancel-failed** | `*Canceled`, `*CancelFailed` | A `cancel` request landed (or itself failed) |

(Examples are illustrative, not exhaustive — the full state enum per bucket is in the [appstate reference](../olares-shared/references/olares-platform-appstate.md#lifecycle-state-machine).) Each lifecycle verb maps to its own subset of terminal-success buckets — see the lifecycle reference.

## OpType vs State (race-safety)

The same `State` can mean different things depending on which mutation is in flight. Concrete example: an `upgrade` issued against an app already in `running` will return `state=running, opType=running` for one or two ticks before the backend flips to `state=upgrading, opType=upgrade`. A naive watcher would declare success at tick zero.

**The CLI's mutating-verb watchers refuse to accept any "success" classification until either:**

1. the row's `OpType` matches the op the CLI just issued, **or**
2. the row disappears entirely (only legal for `uninstall` / `status`).

`cancel` and `status` deliberately set `matchOpType=false` because they are op-agnostic by design — `cancel` declares success on any "row stopped moving" state.

## `--watch` semantics (lifecycle verbs)

- **Polling, not streaming.** Tick on `--watch-interval` (default tuned to the backend's progress cadence).
- `--watch-timeout D` caps total wall-clock time.
- One-shot (no `--watch`) returns as soon as the backend ACKs the mutation request — the row may still be `progressing` for minutes.
- With `--watch`, the CLI blocks until the row reaches a terminal bucket (success OR failure) matching the OpType safety rules above.
- **Idempotent no-ops — `stop` / `resume` only**: `stop` on an already-`stopped` row and `resume` on an already-`running` row return immediately with success (the watcher recognizes the backend's no-op `opType=""` instead of hanging). `install` / `upgrade` / `clone` have **no** such shortcut — they require the matching `OpType` before declaring success. `restart` reuses the `resume` watch bucket (terminal `running`), so its `--watch` inherits the same idempotent-`running` shortcut.
- `--watch-timeout` / `--watch-interval` are **no-ops without `--watch`** (silently ignored, not rejected). There is no `--watch-iterations` flag on market verbs.

### Agent watch discipline (don't block on a long watch)

`--watch` defaults to a 15-minute timeout, and progressing states have very long backend TTLs (`downloading` is **30 days** — see the appstate reference). A foreground `--watch` can therefore block far longer than an agent should sit idle. Discipline:

1. **Use a short foreground window, not the 15m default.** Pass a small `--watch-timeout` sized to the verb (see [references/olares-market-watch.md](references/olares-market-watch.md#per-op-foreground-watch-windows)): ~30s for `stop`/`cancel`/`resume`/`uninstall`, ~1m for the `install` deploy phase / `upgrade` / `clone`.
2. **Timeout is NOT failure.** A `--watch` that times out only means "not terminal yet". Don't report failure — switch to polling `market status <app> --watch --watch-interval 5s` (or fire-and-forget + periodic `market status <app>`), or hand off to diagnosis.
3. **Judge by STATE transitions, never the PROGRESS number** (it is unreliable — appstate reference). A long `downloading` is judged by image-pull progress, not by waiting for terminal.
4. **When a short window expires with no STATE movement, stop waiting and diagnose** — route to [`../olares-doctor/SKILL.md`](../olares-doctor/SKILL.md) (app stuck / won't start), which orchestrates the pod/log evidence.

## "What apps do I have?" routing

| Question | Right verb | Why |
|---|---|---|
| "Show me my apps" / "我的应用" | `market list --mine` (alias `-m`) | Matches the Market UI's "My Terminus" tab exactly — includes in-flight installs, failed rows, transitional states. Wider than "completed installs only" |
| "Runtime status of `<app>`" / "is it installed yet" | `market status <app>` or `market status <app> --watch` | Focused view: `STATE / OPERATION / PROGRESS / SOURCE`. The single-app form does cross-source fallback if `-s` doesn't find the row |
| "Which apps are running right now" | `market status [-a]` then filter `STATE=running` | `status` returns installed-app rows, not a running-only view. For resource ranking use `dashboard applications`; for title search across visible installed apps use `search app` |
| "Browse the catalog" | `market list` (no `--mine`) | Hits `/market/data`, not `/market/state`. Browse-and-discover, not inventory |

> **`list --mine` is NOT the same as "completed installs only".** It hides only the 6 SPA-hidden `uninstalledAppStates` (`pendingCanceled`, `downloadingCanceled`, `downloadFailed`, `installFailed`, `installingCanceled`, `uninstalled`). In-flight installs and post-install failures stay visible because the user clicked something and expects to monitor / retry / cancel them.

## Output conventions

- `-o table` (default) or `-o json` — present on EVERY market verb.
- `-q / --quiet` suppresses all stdout / stderr; exit code carries the signal.
- `--no-headers` — only `list` and `categories` (row-oriented browse). NOT on `get` (key:value detail, no headers separable) or `status` (runtime probe, always headered). No-op silently on mutating verbs.
- Mutating verbs emit a stable `OperationResult` JSON shape with `finalState` / `finalOpType` for scripted parsing.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `no apps in source 'X'` (stderr) | Unknown source id, or empty result | Run `market list -a` to enumerate configured sources |
| `app 'X' is not installed (run 'olares-cli market install X' to install it)` | `status <app>` could not find the row anywhere | Install first, or verify the app name spelling |
| `App is installed under source 'Y' (not 'X')` (stderr) | `-s X` but the row is actually in source Y | Drop `-s`, or pass the correct source. The CLI still renders the row |
| `missing required env var(s): KEY1, KEY2 ...` | `install` for an app that declares required envs | Re-run with `--env KEY=VALUE` (repeatable) for each missing var |
| `app 'X' supports multiple compute modes; re-run with --compute-mode <type> ...` | 1.12.6+ `install` (non-interactive) of an app runnable on several accelerators | Re-run with `--compute-mode <type>` (e.g. `nvidia`); a TTY would prompt instead |
| `app 'X' requires a compute binding ... re-run with --compute-binding ...` | 1.12.6+ `resume` (non-interactive) of a GPU app needing a device | Re-run with `--compute-binding <node>:<device>[:<mem>]` (mem accepts Gi/Mi; repeat the flag once per card for multi-GPU apps) from `settings compute list`; a TTY would prompt instead (comma-separated for multi-card) |
| `the supplied --compute-binding (...) was rejected ...: <reason>` | 1.12.6+ `resume` with an explicit binding the backend refused | `<reason>` mirrors the SPA wording for that `validation.code` — e.g. `aggregate-vram-insufficient` (combined VRAM), `device-vram-insufficient`, `node-pressure` (lists Memory/CPU/Disk Total/Used/Needed), or a raw structural code like `multi-card-not-supported` / `gpu-type-mismatch`. Pick different/more cards per the available list |
| `--compute-mode/--compute-binding requires Olares 1.12.6+ ...` | Flag passed against a 1.12.5 backend | Drop the flag — 1.12.5 uses a different, unchanged code path |
| Lifecycle watcher hangs near `*Failed` state | Backend failed but kept the failure row visible | Inspect `market status <app>` for the failure detail; cancel with `market cancel <app>` if applicable |
| `cannot --watch 'status' (no app argument)` | `market status` without an app + `--watch` | Use `status <app> --watch` |
| 401 / 403 from any verb after refresh | Token rotation / consistent server-side rejection | See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) |
| `--cascade auto-enabled ...` (stderr) | 1.12.5: C/S v2 multi-chart app on single-user cluster | Informational — the CLI picked `--cascade=true`; pass `--cascade=false` to override (1.12.5 only) |
| `--cascade force-enabled ... (CS/shared apps always cascade)` (stderr) | 1.12.6: target is a CS/shared app (apiVersion v2 / shared) | Informational — on 1.12.6 CS/shared apps are always cascaded; `--cascade=false` cannot disable it |
| `'X' has no in-progress operation for this user; nothing to cancel` | `cancel` on 1.12.6 with the per-user row gone and no `--source` | Nothing to cancel; if you must still cancel a stuck op, pass `--source <id>` |

For the full auth-error matrix see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
