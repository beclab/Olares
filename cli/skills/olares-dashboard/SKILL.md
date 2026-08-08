---
name: olares-dashboard
version: 4.3.0
description: "Olares Dashboard via olares-cli dashboard — CPU, memory, disk, network, pod counts, fan, GPU, application/resource rankings, JSON envelopes, and --watch. Use for Dashboard metrics, overview, resource usage, and Olares One fan; not for pod logs or K8s object inspection (olares-cluster)."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# dashboard (overview + applications, AI-agent first)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Use `olares-cli dashboard <verb> --help` for syntax.

## When to use

- Read CPU, memory, disk, network, pod, fan, or GPU metrics.
- Rank applications/workloads by resource use.
- Observe dashboard metrics over time or consume their stable JSON envelopes.

> **Mental model:** dashboard answers *"what's the resource usage and health"*. For inventory and lifecycle, route elsewhere. When the metrics reveal a problem (resource pressure, an app that's `running` but slow/unreachable), hand off to [`../olares-doctor/SKILL.md`](../olares-doctor/SKILL.md) for root-cause diagnosis.

## Verb index

| Verb | Purpose | Read when triggered |
|---|---|---|
| `applications` (`apps`) | Workload-grain resource table | `dashboard applications --help` |
| `overview` | Physical, user, and ranking sections | [overview section meanings](references/olares-dashboard-overview.md); [envelope and empty states](references/olares-dashboard-envelope.md) |
| `overview cpu|memory|disk|pods|network|fan|gpu|physical|user|ranking` | One section | [overview section meanings](references/olares-dashboard-overview.md); [envelope and capability gates](references/olares-dashboard-envelope.md) |
| `schema` | Served JSON schemas | `dashboard schema --help` |
| any metric with watch/window intent | Repeated snapshots / historical window | [watch, windows, NDJSON](references/olares-dashboard-watch.md) |

## Envelope and capability semantics

- Pin automation to `kind`, `raw`, and `meta.empty_reason`, not table labels.
- No hardware, no integration, or a capability gate is usually an `exit 0` empty envelope, not a command failure.
- Aggregate verbs (`overview`, `overview disk|fan|gpu`) exit non-zero only when *every* section carries `meta.error`; a partly degraded envelope is still `exit 0` real data. The envelope reaches stdout either way, so read the per-section error rather than treating the exit code as the whole story.
- A whole-instance outage surfaces as `HTTP 530` on every section — the Olares is unreachable, so confirm it is online before diagnosing dashboard itself.
- Fan is meaningful only on Olares One. Stop probing after the device gate says it is unavailable.
- A hidden GPU sidebar for a non-admin profile is advisory if data is still returned. `vgpu_unavailable` is transient evidence, not proof that the device has no GPU.

## Watch and diagnosis

- Watch emits repeated snapshots; JSON mode is NDJSON. Ctrl-C ends observation, not any workload.
- A watch aborts after repeated iteration failures. Separate transient metric collection failure from actual resource pressure.
- Choose either a relative or absolute time window based on the user's question; do not combine them.
- Dashboard locates pressure; it does not explain pod scheduling, crashes, image pulls, or entrance failures. Route those to `olares-doctor`.

## Safety and escalation

- Dashboard is read-only, but cross-user metrics require platform-admin authority. Do not switch profiles or broaden the target user without approval.
- Do not report an empty integration envelope as "healthy" or "zero usage"; preserve its empty reason.
- Stop and escalate persistent upstream metric errors after a bounded retry, including the affected section and time window.
