---
name: olares-doctor
version: 1.0.0
description: "Runtime diagnosis for Olares apps and the system via olares-cli — symptom-to-root-cause routing for an app that won't install, won't start, crashes, is `running` but unreachable, or is slow; plus the `doctor` command tree (images). Use when an app or Olares is misbehaving and you need to find out why; both catalog (market) and dev (chart) apps hand runtime failures here."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# doctor (runtime diagnosis)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli doctor --help`.**

> **Source of truth for flags is always `olares-cli doctor <verb> --help`.** This file only carries what `--help` cannot give: symptom-to-reference routing, the `doctor` command index, and which sibling skill owns each evidence-gathering command.

> **This skill is a thin router.** It does not restate platform facts or duplicate sibling-skill commands — it maps a symptom to the right reference, then orchestrates `olares-cluster` / `olares-dashboard` / `olares-market` to gather evidence. Backend facts (state machine, TTLs, `running` semantics, download serialization) live once in [`../olares-shared/references/olares-platform-appstate.md`](../olares-shared/references/olares-platform-appstate.md).

## When to use

Diagnosing **why** an app or the system misbehaves — not performing lifecycle actions (that is `market`/`chart`) and not authoring charts (`chart`):

- An install/upgrade is stuck or never reaches `running`; an app won't start.
- An app crashes / restarts repeatedly (CrashLoopBackOff, exit codes, config errors).
- An image won't pull (`ImagePullBackOff` / `ErrImagePull` / wrong arch), or you want to find unused local images.
- An app is `running` but its entrance is unreachable / errors / times out.
- The system or an app is slow, or a GPU/resource binding is rejected (`node-pressure`).

**Both catalog apps (installed via `market`) and your own dev apps (deployed via `chart`) route runtime failures here.** Once the root cause is found, the *fix* for a dev app you authored is usually a chart edit — hand back to [`../olares-chart/SKILL.md`](../olares-chart/SKILL.md).

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** `doctor` answers *"why is this broken and what do I do next?"* It reads evidence, never mutates. The four-skill develop->deploy->debug combo is `chart` + `market` + `olares-shared` + `doctor`.

## Symptom -> reference routing

Pick the row that matches the reported symptom; each reference carries the locate-commands, the decision criteria, and the next step.

| Symptom | Reference |
|---|---|
| Install/upgrade stuck; never reaches `running`; sits in `pending` / `downloading` / `installing` / `initializing`; a fresh install ended in `stopped` | [references/olares-doctor-app-stuck.md](references/olares-doctor-app-stuck.md) |
| App crashes / restarts (CrashLoopBackOff, non-zero exit, `CreateContainerConfigError`, permission errors) | [references/olares-doctor-app-crash.md](references/olares-doctor-app-crash.md) |
| Image won't pull (`ImagePullBackOff` / `ErrImagePull` / `InvalidImageName` / arch mismatch); finding unused local images | [references/olares-doctor-image.md](references/olares-doctor-image.md) |
| App is `running` but the entrance is unreachable / 5xx / times out / blank | [references/olares-doctor-running-unhealthy.md](references/olares-doctor-running-unhealthy.md) |
| System or app slow; resource pressure; GPU/compute binding rejected (`node-pressure`) | [references/olares-doctor-resources.md](references/olares-doctor-resources.md) |

> **First, rule out the normal queue.** Before declaring an install stuck, check whether another app is `downloading` — app-service runs **one download at a time**, so a `pending` row is often just queuing (see the appstate reference and the app-stuck reference).

## `olares-cli doctor` command tree

`doctor` also hosts read-only diagnostic commands that combine multiple Olares API surfaces. **Source of truth for flags is always `olares-cli doctor <verb> --help`.** They mutate nothing.

| Command | What it does | Reference |
|---|---|---|
| `doctor images` | Lists local containerd images annotated with how many workloads reference each one; `--unused` shows zero-reference prune candidates (largest-first, with reclaimable size). Always full-scans the control node; `-n` / `-l` scope the workload reference count. | [references/olares-doctor-image.md](references/olares-doctor-image.md) |
| `doctor thirdleveldomain` | Audits Application `customDomain.third_level_domain` per user zone (kubeconfig): flags duplicate prefixes and reserved names (`auth` / `desktop` / `wizard`). `--force-dedupe` keeps one duplicate per zone, clears the rest, and clears reserved names. | — |

## How doctor gathers evidence (orchestration, not ownership)

`doctor` does not own these commands — it routes to the skill that does. The references spell out the exact invocations:

| Need | Skill | Typical commands |
|---|---|---|
| App lifecycle state / source / watch | [`../olares-market/SKILL.md`](../olares-market/SKILL.md) | `market status <app> [-a]`, `market list --mine` |
| Pod / container status, logs, events, workloads | [`../olares-cluster/SKILL.md`](../olares-cluster/SKILL.md) | `cluster application status <ns>`, `cluster pod list/get/logs/events`, `cluster workload ...` |
| Resource usage / pressure / GPU | [`../olares-dashboard/SKILL.md`](../olares-dashboard/SKILL.md) | `dashboard overview [memory\|cpu\|disk\|gpu]`, `dashboard applications` |
| The app's namespace (`<app>-<owner>` vs `<app>-shared`, v2 multi-namespace) | [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md#finding-an-apps-namespace) | — |

> **Admin caveat:** `os-framework` / `os-platform` system pods are typically visible only to an **admin** profile. On `HTTP 403` / `404`, fall back to the app's own pod logs and report that platform logs need admin.

## Common errors

These are diagnosis-flow pitfalls, not CLI flag errors (those live in each sibling skill). Auth errors (`ErrTokenInvalidated` / 401 / 403 after refresh) are always [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).

| Symptom | Cause | Fix |
|---|---|---|
| Declared an install "stuck" while it's `pending` | Another app is `downloading` — app-service serializes downloads (one at a time) | Not stuck; it's queuing. Confirm with `market status -a`, then wait |
| Treated a `--watch` timeout as a failure | A short foreground window expired; the op is still progressing (long backend TTLs) | Not a failure. Re-judge by STATE (`market status <app>`), not the timeout — see [olares-doctor-app-stuck.md](references/olares-doctor-app-stuck.md) |
| `running` but the app doesn't work | `running` only proves entrance TCP-reachability, not health | Climb the health ladder — [olares-doctor-running-unhealthy.md](references/olares-doctor-running-unhealthy.md) |
| A fresh install ended in `stopped` (no `*Failed`) | Scheduling failure tears down via `Stopping -> stopped`, never `installFailed` | Read pod events for the scheduling reason — [olares-doctor-resources.md](references/olares-doctor-resources.md) |
| `HTTP 403` / `404` reading `os-framework` / `os-platform` pods | Active profile isn't admin | Fall back to the app's own pod logs; report that platform logs need an admin profile |
| Found the root cause but unsure how to fix a dev app | The fix for an app you authored is a chart edit, not a CLI action | Hand back to [`../olares-chart/SKILL.md`](../olares-chart/SKILL.md) |
