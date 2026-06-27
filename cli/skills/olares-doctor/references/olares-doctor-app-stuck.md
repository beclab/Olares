# doctor: app stuck / won't install / never reaches running

> **Prerequisite:** read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Backend facts** (state machine, TTLs, single-download serialization, `Pending`->`stopped` trap) live in the shared **application state machine**. This reference is the diagnostic procedure on top of those facts.

Symptom: an install/upgrade is stuck, or an app never reaches `running` — it sits in `pending` / `downloading` / `installing` / `initializing`, or a fresh install quietly ended in `stopped`.

## Step 0 — read the state row, then rule out the normal queue

```bash
olares-cli market status <app> -o json     # STATE / OPERATION / SOURCE
olares-cli market status -a                 # is ANY app currently downloading?
```

app-service admits **one `downloading` app at a time**; everything else waits in `pending`. So a `pending` row while another app downloads is **normal queuing, not stuck** — it advances when the in-flight download finishes. Only treat `pending` as a problem if nothing is downloading and it never moves.

Then branch on STATE.

## `downloading` — slow pull vs stalled pull

A `downloading` row has a **30-day** backend TTL, so it will never self-fail — judge it by byte-level pull progress, not by waiting. The market PROGRESS field is unreliable; real progress is in the per-node `image-service` DaemonSet (it pulls via containerd):

```bash
olares-cli cluster pod list -n os-framework | grep image-service
olares-cli cluster pod logs os-framework/<image-service-pod> -f | grep -E "progress=|downloading,ref:"
```

- `download image <ref> progress=<pct>, imageSize=<bytes>, offset=<bytes>` = whole-image percent; `status: downloading,ref: layer-... offset:/Total:` = active layer. Sample `offset` across two ticks: rising = healthy (large image), **flat over many ticks = a stalled pull** (registry/mirror/network) — check the mirror/connectivity.
- Model weights (an engine downloading the model *after* its image is up) are a **separate** phase, not in image-service — watch the app's own init container, e.g. `cluster pod logs <ns>/<app-pod> -c llm-init`.
- A stalled pull that can't be resolved → hand off to **doctor: image / pull failures**.

## `installing` / `initializing` stuck — two traps

`installing` (30m TTL) and `initializing` (1h TTL) only fast-fail on hard pod conditions that persist past a 5-minute grace; otherwise they poll the long TTL. So a stuck row needs pod-level inspection, not more waiting. Resolve the namespace (`<app>-<owner>`, or `<app>-shared` for a shared app; a v2 app spans several namespaces — see finding an app's namespace), then:

```bash
olares-cli cluster application status <ns>          # workload readiness at a glance
olares-cli cluster pod list -n <ns> -o json         # status.containerStatuses[].{ready,restartCount,state}
olares-cli cluster pod events <ns>/<pod>            # scheduling / pull / mount events
```

Two non-obvious traps:

- **Scheduling failure does NOT become `installFailed`.** A pod that can't schedule (stays `Pending`) is torn down via `Stopping -> stopped`. So **a fresh install that ended in `stopped` is the red flag** — read `cluster pod events` for the scheduling reason (insufficient CPU/memory/GPU, taints) and hand off to **doctor: resources / scheduling**.
- **Soft-hang:** the pod is `Running` but the app never becomes serve-ready (no crash). The row stays `initializing` and is never fast-failed — only the pod logs reveal it. If the container is up but the entrance never opens, hand off to **doctor: running but unhealthy**.

## Decision -> next step

| Finding | Route to |
|---|---|
| Another app is downloading; this one is `pending` | Normal queue — wait, no action |
| `downloading`, `offset` rising | Healthy large pull — keep polling |
| `downloading`, `offset` flat | **doctor: image / pull failures** (stalled pull) |
| Pod `Pending` / fresh install in `stopped` | **doctor: resources / scheduling** |
| Pod restarting / CrashLoopBackOff | **doctor: app crash** |
| Pod Running but entrance never opens | **doctor: running but unhealthy** |

For orchestration-level errors, also check `os-framework/app-service-0` (`cluster container logs os-framework/app-service-0/app-service`) — admin-only; fall back to the app's own pod logs if 403/404. **If this is a chart you are authoring**, fix per [`../../olares-chart/SKILL.md`](../../olares-chart/SKILL.md).
