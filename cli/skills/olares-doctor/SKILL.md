---
name: olares-doctor
version: 1.3.0
description: "Runtime diagnosis for Olares apps and the system via olares-cli — find the root cause when an app won't install or start, crashes, cannot pull an image, is `running` but unreachable, or is slow; includes doctor images and thirdleveldomain. Use for diagnosing catalog and dev app failures, not for authoring or editing charts."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# doctor (runtime diagnosis)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

This skill is a thin diagnostic router over Market, Cluster, and Dashboard. Load the shared [application-state model](../olares-shared/references/olares-platform-appstate.md) when interpreting lifecycle states, TTLs, serialized downloads, or `running`. Use `olares-cli doctor <verb> --help` for syntax.

## When to use

- An install/upgrade is stuck or never reaches `running`; an app won't start.
- An app crashes / restarts repeatedly (CrashLoopBackOff, exit codes, config errors).
- An image won't pull (`ImagePullBackOff` / `ErrImagePull` / wrong arch), or you want to find unused local images.
- An app is `running` but its entrance is unreachable / errors / times out.
- The system or an app is slow, or a GPU/resource binding is rejected (`node-pressure`).

**Both catalog apps (installed via `market`) and your own dev apps (deployed via `chart`) route runtime failures here.** Once the root cause is found, the *fix* for a dev app you authored is usually a chart edit — hand back to [`../olares-chart/SKILL.md`](../olares-chart/SKILL.md).

> **Mental model:** `doctor` answers *"why is this broken and what do I do next?"* Diagnosis is read-only by default; the only mutation is the explicitly approved `thirdleveldomain --force-dedupe` repair. The four-skill develop->deploy->debug combo is `chart` + `market` + `olares-shared` + `doctor`.

## Symptom routing

| Symptom | Reference |
|---|---|
| Install/upgrade stuck; never reaches `running`; sits in `pending` / `downloading` / `installing` / `initializing`; a fresh install ended in `stopped` | [references/olares-doctor-app-stuck.md](references/olares-doctor-app-stuck.md) |
| App crashes / restarts (CrashLoopBackOff, non-zero exit, `CreateContainerConfigError`, permission errors) | [references/olares-doctor-app-crash.md](references/olares-doctor-app-crash.md) |
| Image won't pull (`ImagePullBackOff` / `ErrImagePull` / `InvalidImageName` / arch mismatch); finding unused local images | [references/olares-doctor-image.md](references/olares-doctor-image.md) |
| App is `running` but the entrance is unreachable / 5xx / times out / blank | [references/olares-doctor-running-unhealthy.md](references/olares-doctor-running-unhealthy.md) |
| System or app slow; resource pressure; GPU/compute binding rejected (`node-pressure`) | [references/olares-doctor-resources.md](references/olares-doctor-resources.md) |

A **model** that is configured but does not answer is diagnosed one layer up first: [`olares-router`](../olares-router/SKILL.md) separates the gateway, its access control and the model application's own download/engine state from the pod-level failures here, and routes back when the cause is below the application.

> **First, rule out the normal queue.** Before declaring an install stuck, check whether another app is `downloading` — app-service runs **one download at a time**, so a `pending` row is often just queuing (see the appstate reference and the app-stuck reference).

## Verb index

| Command | Purpose | Read when triggered |
|---|---|---|
| `images` | Full local image inventory annotated with workload references; unused candidates | [image diagnosis](references/olares-doctor-image.md) |
| `thirdleveldomain` | Audit duplicate/reserved third-level domains; optional repair | [domain audit and repair](references/olares-doctor-thirdleveldomain.md) |

## How doctor gathers evidence (orchestration, not ownership)

- Lifecycle state/source comes from [`olares-market`](../olares-market/SKILL.md).
- Pods, events, logs, and workloads come from [`olares-cluster`](../olares-cluster/SKILL.md).
- Pressure and utilization come from [`olares-dashboard`](../olares-dashboard/SKILL.md).
- Namespace discovery follows the shared [platform model](../olares-shared/references/olares-platform.md).

Correlate evidence by time and object ownership. A Market timeout is not failure; `running` proves only entrance TCP reachability; a fresh install can settle at `stopped` after scheduling failure without a `*Failed` lifecycle state.

## Safety and escalation

- Diagnosis is read-only by default. Do not restart, delete, scale, cancel, prune, or edit a chart as an automatic diagnostic step.
- `thirdleveldomain --force-dedupe` mutates Application resources. Show the proposed changes and obtain explicit approval first.
- An unused-image report is evidence, not authorization to remove images.
- System namespace evidence commonly requires admin visibility. On 403/404, use the app's own evidence and report the missing visibility; do not switch identities without approval.
- Stop when the app/user/namespace is ambiguous, required logs need a higher role, or the fix crosses into chart editing, lifecycle mutation, or host administration.
