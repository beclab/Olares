# cluster workload (alias `wl`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster workload --help` and `olares-cli cluster workload <verb> --help`.

"Workload" = the K8s controller resources Deployment, StatefulSet, and DaemonSet — the same set the ControlHub SPA exposes under "Workloads" inside an Application Space. **The most flag-sensitive noun in this tree** because every verb except `list` requires `--kind`.

## `--kind` requirement

| Verb | `--kind` value |
|---|---|
| `list` | `all` (default — fans out one request per kind and merges with a `KIND` column) OR `deployment` / `statefulset` / `daemonset` (singular / plural / short forms `deploy`, `sts`, `ds` are accepted) |
| `get` / `yaml` / `rollout-status` / `scale` / `restart` / `stop` / `start` / `delete` | **REQUIRED** — pick exactly one of `deployment` / `statefulset` / `daemonset`. No `all` here |

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | Multi-kind union by default; `--kind X` to scope. Compact view (NAME/READY/AGE); for container images use the `images` verb |
| `images [IMAGE]` | Paginated list of image references from pod templates across Deployment/StatefulSet/DaemonSet/Job/CronJob (`--kind` also accepts `job` / `cronjob`). Pass an IMAGE arg to filter to "where is this image referenced?" (tag/digest-normalized; always full-scans so it can't miss later-page refs) |
| `get <ns/name> --kind X` | Vertical summary with kind-aware READY counts (`readyReplicas/replicas` or `numberReady/desiredNumberScheduled` for DaemonSet) |
| `yaml <ns/name> --kind X` | Full K8s-native YAML |
| `rollout-status <ns/name> --kind X` | Reports whether the rollout has converged (kind-aware). Without `-w`: one GET, exit 0 if converged or 2 if not. With `-w`: poll on `--interval` until converged, `--timeout` (default 10m), or Ctrl-C. Emits only on state change |
| `scale <ns/name> --kind X --replicas N` | **Destructive.** PATCH merge-patch+json `{"spec":{"replicas":N}}`. DaemonSet rejected. `--replicas=0` triggers `ConfirmDestructive`. `-w` chains into `rollout-status -w` automatically |
| `restart <ns/name> --kind X` | **Destructive.** 3-step: (1) GET selector; (2) GET pods by selector; (3) parallel DELETE pods. `--concurrency` (default 5) bounds parallel deletes. NOT the kubectl `restartedAt` annotation trick |
| `stop <ns/name> --kind X` | **Destructive.** Alias for `scale --replicas=0`. DaemonSet rejected (delete the workload instead) |
| `start <ns/name> --kind X --replicas N` | Non-destructive. Alias for `scale --replicas=N`. `--replicas` REQUIRED (no cached previous count). No `--yes` |
| `delete <ns/name> --kind X` | **Destructive.** CLI-original (SPA has no direct workload-delete button). `--propagation foreground` (default) waits for the cascade |

## Safety constraints

- **Every destructive verb confirms with the user.** `--yes` opts out for scripts.
- **DaemonSet has no `replicas`** — `scale` / `stop` reject it client-side. `start` requires `--replicas` so it implicitly excludes DaemonSet too.
- **`restart` deletes pods one-by-one (parallel-bounded).** During the operation pods are recreated by the controller. For Deployments with a single replica, this means a brief downtime.
- **`delete --propagation foreground` waits for the cascade.** Pass `background` only when you intentionally want fire-and-forget.

## Examples

```bash
# Cross-kind listing.
olares-cli cluster workload list

# Just StatefulSets in a namespace.
olares-cli cluster workload list -n user-system-alice --kind sts

# List workload image references (all kinds incl. Job/CronJob), paginated.
olares-cli cluster workload images --limit 50 --page 1

# Where is a given image referenced? (tag/digest-normalized; full scan)
olares-cli cluster workload images docker.io/library/nginx:latest

# Local images + workload reference counts (full scan; pause excluded).
olares-cli doctor images -o json
# Only the orphans (zero refs), biggest first, with reclaimable-size footer.
olares-cli doctor images --unused

# Get + watch rollout to convergence.
olares-cli cluster workload get user-system-alice/api --kind deploy
olares-cli cluster workload rollout-status user-system-alice/api --kind deploy -w --interval 3s --timeout 5m

# Scale + watch (auto-chains into rollout-status).
olares-cli cluster workload scale user-system-alice/api --kind deploy --replicas 3 -w

# Stop (scale to 0) — confirms.
olares-cli cluster workload stop user-system-alice/api --kind deploy

# Restart — kills pods, controller recreates them.
olares-cli cluster workload restart user-system-alice/api --kind deploy --concurrency 3

# Delete the controller itself.
olares-cli cluster workload delete user-system-alice/api --kind deploy --propagation foreground
```

## The "scale + watch" pattern (recommended)

```bash
olares-cli cluster workload scale <ns/name> --kind deploy --replicas 5 -w
```

This combines the PATCH and the rollout-status poll into one invocation. The agent should always offer `-w` to users who say "scale and tell me when it's done".

## Agent notes

- **`--kind` errors are the most common gotcha.** Always include it in `get` / `yaml` / mutating verbs. For `list`, omit it (or use `all`) when the user doesn't specify a type.
- **`start <name> --replicas N` requires the user to know N.** If the user says "start this back up", ask what replica count they want before invoking. There is intentionally no "remember previous replicas" cache.
- For "restart this app" requests, **prefer `pod delete` on a specific pod** when the user just wants a single pod to bounce; reach for `workload restart` only when they actually want every replica to recycle.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `--kind is required` | Missing `--kind` on a non-list verb | Add `--kind deploy|sts|ds` |
| `DaemonSet does not have replicas; use 'cluster workload delete --kind ds' instead` | `scale` / `stop` on a DS | Use `delete` |
| `--replicas is required` (on `start`) | Tried to "start" without saying how many | Ask the user; there is no cached previous count |
| 404 on a known-existing workload | Wrong `--kind` | Try with the correct kind, or `list --kind all -n <ns>` |
| `aborted by user` | Destructive prompt rejected | If intentional, re-run with `--yes` |
