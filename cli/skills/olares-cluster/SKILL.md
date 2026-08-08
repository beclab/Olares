---
name: olares-cluster
version: 4.5.0
description: "Olares ControlHub K8s runtime view via olares-cli cluster — inspect pods, containers, workloads, logs, jobs, cronjobs, namespaces, nodes, and middleware; exec, scale, restart, or delete K8s objects. Use for raw runtime objects and logs, not app lifecycle (market), resource metrics (dashboard), or host install."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# cluster (per-user K8s view)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Load the shared [platform model](../olares-shared/references/olares-platform.md) when the task depends on app namespaces, application spaces, cross-namespace DNS, or system middleware. Use `olares-cli cluster <noun> <verb> --help` for syntax.

## When to use

- Inspect K8s runtime objects, YAML, events, logs, or identity context.
- Execute a one-shot command in a container.
- Scale, restart, stop, start, or delete a workload; suspend/resume cronjobs; rerun jobs.
- Watch pod, workload, application, or log state.

> **Diagnosing *why* an app is broken** (stuck install, crash loop, `running` but unreachable, image won't pull, resource pressure) is [`../olares-doctor/SKILL.md`](../olares-doctor/SKILL.md) — it orchestrates these `cluster` commands into symptom→root-cause routing. `cluster` stays the raw runtime view and the place that mutates K8s objects.

`workload stop/start` changes controller replicas; it does not update the Market lifecycle row. Use `market stop/resume` for app-level lifecycle.

## The identity-vs-server-decides principle (cross-cutting)

- Identity is the active profile; there is no per-invocation profile override.
- The server decides visibility and authorization. `cluster context` is identity display, not a permission preflight.
- Surface a 403 as authoritative. The exception is `exec`, whose namespace safety gate is described in its reference.

## Verb index

| Noun | Verbs | Read when triggered |
|---|---|---|
| `context` | (single verb) | `olares-cli cluster context --help` |
| `pod` | `list`, `get`, `yaml`, `events`, `logs`, `delete`, `restart`, `exec` | `exec` requires Olares 1.12.7+; [pod operations](references/olares-cluster-pod.md); [exec safety](references/olares-cluster-exec.md) |
| `container` | `list`, `env`, `logs`, `exec` | `exec` requires Olares 1.12.7+; [exec safety](references/olares-cluster-exec.md) |
| `workload` (`wl`) | `list`, `images`, `get`, `yaml`, `rollout-status`, `scale`, `restart`, `stop`, `start`, `delete` | [workload operations](references/olares-cluster-workload.md) |
| `application` (`app`) | `list`, `get`, `workloads`, `pods`, `status` | [application aggregation](references/olares-cluster-application.md) |
| `namespace` (alias `ns`) | `list`, `get` | `olares-cli cluster namespace --help` |
| `node` (alias `nodes`) | `list`, `get` | `olares-cli cluster node --help` |
| `job` (`jobs`) | `list`, `get`, `yaml`, `pods`, `events`, `rerun` | [job operations](references/olares-cluster-job.md) |
| `cronjob` (`cronjobs`, `cj`) | `list`, `get`, `yaml`, `jobs`, `suspend`, `resume` | [cronjob operations](references/olares-cluster-cronjob.md) |
| `middleware` (`mw`) | `list` | [middleware model](references/olares-cluster-middleware.md) |

## Watch and asynchronous semantics

- Watches and log follow poll; they are not streaming API watches.
- Transient network/5xx failures are retried within a bounded streak. Terminal 4xx responses stop immediately; 408 and 429 remain retryable.
- Ctrl-C stops the local poll cleanly. It does not undo a mutation already accepted by the server.
- A rollout/status watch ending is not proof that the application entrance is healthy; use `olares-doctor` when runtime symptoms remain.

## Safety and escalation

- Confirm every destructive runtime mutation before invoking it, even if `--yes` is available.
- `exec` is a remote code-execution boundary. Confirm namespace, pod, container, command, and whether writes are expected. Prefer one-shot execution; interactive TTY belongs to the human.
- Do not delete or restart objects merely because a diagnostic found them unhealthy. Establish controller ownership and the app-level consequence first.
- A 404 may hide a namespace the profile cannot see. Compare against `cluster application list`; do not broaden scope or switch identity without user approval.
- For complete image-reference diagnostics use `doctor images`, not a paginated workload listing.
- Stop on ambiguous namespace/resource identity, missing permission, or any request to access another user's container.
