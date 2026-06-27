# Olares application state machine (single source of truth)

> Cross-skill platform facts about how app-service drives an app through its lifecycle. `market` (operational gating + `--watch` discipline), `chart` (deploy loop), and `doctor` (diagnosis) all link here (one hop) instead of re-stating the state machine. Pure backend contract — no login/profile content (that is in [`../SKILL.md`](../SKILL.md)).
>
> Source of truth in the backend: `framework/app-service/pkg/appstate/state_transition.go` (states, transitions, allowed ops, TTLs), `pending_app.go` (download serialization), `appinstaller/helm_ops_install.go` (`WaitForLaunch` / `WaitForStartUp`), `images/puller.go` (progress).

## Lifecycle state machine

A normal install advances through a fixed pipeline; each arrow is the only legal forward edge:

```
pending -> downloading -> installing -> initializing -> running
```

- **`installing` may skip `initializing`** straight to `running` (system-middleware fast-path).
- **`upgrade` of a stopped app** re-renders the chart at `replicas=0` and lands **back in `stopped`** (nothing to launch), not `running`.
- The state enum groups into four buckets (the CLI mirrors these):

| Bucket | States | Meaning for an agent |
|---|---|---|
| **Progressing** | `pending`, `downloading`, `installing`, `initializing`, `upgrading`, `applyingEnv`, `resuming`, `stopping`, `uninstalling`, every `*Canceling` | Backend is actively working — keep polling, do NOT treat as failure |
| **Terminal success** | `running`, `stopped`, `uninstalled` | Operation finished cleanly |
| **Terminal failure** | `downloadFailed`, `installFailed`, `upgradeFailed`, `stopFailed`, `resumeFailed`, `applyEnvFailed`, `uninstallFailed` | Operation finished with a hard error |
| **Canceled / cancel-failed** | `*Canceled` (the op was canceled before completing — app is NOT really installed), `*CancelFailed` (the cancel itself failed — conservatively treat as "still there") | A `cancel` landed (or itself failed) |

Used by: `market` (bucket classification for watchers), `doctor` (is this state a failure, a queue, or normal progress?).

## Per-state allowed operations (the gate)

app-service rejects any operation not allowed in the current state (`OperationAllowedInState`). This is why "just retry install" often fails — the row must be in an install-accepting state first.

| Current state | Operations app-service accepts |
|---|---|
| no row / `uninstalled` / `pendingCanceled` / `downloadingCanceled` / `installingCanceled` | `install` |
| `pending` / `downloading` / `installing` / `initializing` / `upgrading` / `applyingEnv` | `cancel` only |
| `resuming` | `cancel`, `stop` |
| `running` | `uninstall`, `upgrade`, `env`, `stop` |
| `stopped` | `uninstall`, `upgrade`, `env`, `resume` |
| `downloadFailed` | `install` |
| `installFailed` | `install`, `uninstall` |
| `upgradeFailed` | `upgrade`, `uninstall` |
| `applyEnvFailed` | `env`, `uninstall` |
| `stopFailed` | `stop`, `upgrade`, `uninstall` |
| `resumeFailed` | `resume`, `upgrade`, `env`, `uninstall` |
| `uninstallFailed` | `uninstall` |
| `pendingCancelFailed` / `downloadingCancelFailed` | `cancel` |
| `installingCancelFailed` / `upgradingCancelFailed` / `applyingEnvCancelFailed` | `cancel`, `uninstall` |
| any `*Canceling` / `uninstalling` | none — only wait for the terminal state |

Key consequences: an **in-flight** app accepts only `cancel` (never a direct `uninstall` — the CLI auto-cancels first); `install` against an already-existing settled app is rejected (use `upgrade`).

Used by: `market` (verb pre-flight gating), `chart` (install-vs-upgrade verb choice), `doctor` ("why was my operation rejected?").

## Backend fail TTLs (how long a state can sit before app-service gives up)

Each progressing state has its own timeout before the backend itself fails the op (`StateToDurationMap`, default fallback 10m):

| State | Backend TTL |
|---|---|
| `pending` | 24h |
| `downloading` | **30 days** |
| `installing` | 30m |
| `initializing` | 1h |
| `upgrading` | 1h |
| `applyingEnv` | 30m |
| (any other) | 10m |

The `downloading` 30-day TTL is the headline fact: **a slow/large image pull will never self-fail in any reasonable agent timeframe**, so a foreground `--watch` that sits in `downloading` is not a hang to wait out — judge it by image-pull progress, not by waiting for a terminal state.

Used by: `market` (why `--watch` blocks so long), `doctor` (distinguishing a stalled pull from a slow one).

## Serialized downloads (only one app downloads at a time)

app-service admits **at most one app in `downloading`** cluster-wide (`pending_app.go`: it counts `Downloading` rows and only proceeds when `count < 1`, else the app waits in line in `pending`).

- A batch of installs therefore drains **serially**: one downloads while the rest sit in `pending`. This is normal queuing, NOT a stuck install.
- An app parked in `pending` while another app is `downloading` needs no intervention — it advances once the in-flight download finishes.

Used by: `market` (a `pending` row is expected during batch installs), `doctor` (rule out the queue before declaring an app stuck — check whether any other app is `downloading`).

## What `running` really means (TCP-reachable, not "healthy")

`running` is set by `WaitForLaunch`, which polls every 1s and only checks that **each entrance's host:port accepts a TCP connection** (`apputils.TryConnect`). It does **not** issue an HTTP request, check readiness, or verify the app actually serves correctly.

- So `state=running` means **L4 reachable**, not "the app works". An app can be `running` yet return 5xx, show a blank page, or still be warming up. Full health needs a layer-by-layer check (pod Ready / RESTARTS stable / entrance HTTP 200 / logs clean) on top of the row.
- **`StudioSource` (Devbox) apps skip the launch probe entirely** — they reach `running` without any TCP check.
- **Fast-fail during `installing` / `initializing`:** `WaitForStartUp` / `WaitForLaunch` bail early (instead of polling the full 30m/1h TTL) when `hasUnrecoverablePod` reports a condition that persists past a **5-minute grace** (`unrecoverableGrace`):
  - hard, never-self-healing pod errors: `ImagePullBackOff`, `ErrImagePull`, `InvalidImageName`, `CreateContainerConfigError`, `Unschedulable`;
  - `CrashLoopBackOff` with `RestartCount >= 5` (`crashLoopRestartThreshold`).
- A crash that has NOT yet crossed that threshold/grace does not fast-fail — the row legitimately stays `initializing` for several minutes while the container is already crashlooping.

Used by: `market` (the "running ≠ healthy" verification ladder), `doctor` (a `running` app with a broken entrance is a runtime-health case, not an install failure).

## `progress` is not a reliable signal

The `PROGRESS` field on the state row cannot be trusted for fine-grained tracking:

- Image-pull progress is a no-op in app-service (`images/puller.go` `Progress()` returns `""`).
- The install progress is initialized to a hardcoded `"0.00"`.

**Judge by STATE transitions, not by the progress number.** Byte-level image-pull progress comes from the per-node `image-service` DaemonSet logs / `imagemanagers` CRD, not from this field (see `doctor`).

Used by: `market` (don't poll on progress), `doctor` (where real pull progress actually lives).

## Two non-obvious terminal behaviors

- **A scheduling failure does not become `installFailed`.** When a pod can't be scheduled (stays `Pending`), app-service tears the install down through `Stopping -> stopped`, not `installFailed`. A watcher that only looks for `*Failed` will miss it — a fresh install that ends in `stopped` is a red flag, not a success.
- **`cancel` is teardown-vs-stop depending on phase.** Canceling `pending` / `downloading` / `installing` **tears the partial install down (namespace deleted)** — functionally equivalent to uninstalled. Canceling `initializing` / `upgrading` / `applyingEnv` / `resuming` only **stops** the app (it lands in `stopped`, still installed). `market uninstall` relies on this split when it auto-orchestrates an in-flight uninstall.

Used by: `market` (uninstall auto-orchestration, cancel outcome), `doctor` (a just-installed app sitting in `stopped` is the scheduling-failure trap).
