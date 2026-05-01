---
name: olares-cluster
version: 0.4.1
description: "olares-cli cluster command tree: per-user K8s view of an Olares cluster via the ControlHub BFF (https://control-hub.<terminus>). Read + mutate surface: `cluster context` (identity / globalrole / accessible workspaces from /capi/app/detail), `cluster pod list / get / yaml / events / logs / delete / restart` (`get --watch` polls and repaints; `delete`/`restart` issue DELETE under ConfirmDestructive), `cluster container list / env / logs` (per-pod container projection — logs delegates to pod.RunLogs), `cluster workload list / get / yaml / rollout-status / scale / restart / stop / start / delete` (Deployment / StatefulSet / DaemonSet, --kind all by default; `rollout-status --watch` enforces kubectl-style convergence; `scale` PATCHes `application/merge-patch+json` and chains into rollout-status with --watch; `restart` is SPA-aligned GET-selector → list-pods → parallel DELETE; `stop`/`start` are aliases over `scale`; `delete` cascades by --propagation), `cluster application list / get / workloads / pods / status` (Olares ApplicationSpaces grouped by KubeSphere workspace via /capi/namespaces/group; `status` is a CLI-original parallel fan-out across workloads + pods + events with optional --watch), `cluster namespace list / get` (raw K8s framing), `cluster node list / get` (per-user node view with kubectl-style STATUS/ROLES/AGE/VERSION columns), `cluster job list / get / yaml / pods / events / rerun` (apis/batch/v1; `pods` filters by `controller-uid=<uid>`; `rerun` calls KubeSphere operations API `kapis/operations.kubesphere.io/v1alpha2`), `cluster cronjob list / get / yaml / jobs / suspend / resume` (apis/batch/v1beta1; `jobs` derives child-Job selector from spec.jobTemplate.metadata.labels; `suspend`/`resume` PATCH `application/merge-patch+json`), `cluster middleware list / password set` (Olares-managed databases / queues / object stores; `password set` POSTs to `/middleware/v1/<type>/password` with no-echo prompt + ConfirmDestructive). Per-user resource scoping is ALWAYS enforced server-side; CLI verbs MUST NOT consult the locally cached cluster context to gate calls — the cache is for display only. Authentication uses the active profile's access_token via the factory's refreshingTransport (auto-rotates on 401/403). Wire formats handled: KubeSphere {items, totalItems} on /kapis/*, K8s native {kind, apiVersion, metadata, items|spec|status} on /api/v1/* and /apis/*, K8s native logs (text/plain) on /api/v1/.../log, ControlHub /capi/* custom shapes (no envelope), Olares /middleware/v1/* envelope ({code, data:[]}), KubeSphere operations actions on /kapis/operations.kubesphere.io/v1alpha2/* (no body, query params), JSON merge patches on /apis/* (Content-Type application/merge-patch+json). --follow on logs and --watch on get / rollout-status / application status all use polling (sinceTime advance / repeat-and-render) — uniform with `olares-cli market --watch`, no long-lived chunked streams. Mutating verbs are wrapped in ConfirmDestructive (lifted from settings/vpn — `--yes` / `-y` opts out). Use whenever the user asks about pods / containers / workloads / namespaces / application spaces / nodes / middleware / jobs / cronjobs / global roles / scaling / restarting / deleting K8s objects on the per-user cluster view, NOT for app-store lifecycle (use `olares-cli market`) or host-side install/upgrade (use `olares-cli node` / `gpu` / `os`)."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli cluster --help"
---

# cluster (Olares per-user K8s view)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for the profile selection, login, and HTTP 401/403 recovery rules that every command here depends on.**

## What this command tree is

`olares-cli cluster ...` is the CLI mirror of the ControlHub SPA at [`apps/packages/app/src/apps/controlHub`](apps/packages/app/src/apps/controlHub) — the per-user view of an Olares cluster's Kubernetes resources. Identity and transport come from the active profile, the same way [`olares-cli settings`](../olares-settings/SKILL.md) and [`olares-cli market`](../olares-market/SKILL.md) work.

> Boundary: this tree is the **runtime-state view**. App-store lifecycle (install / uninstall / start / stop / upgrade) belongs to [`olares-cli market`](../olares-market/SKILL.md); host-side maintenance (cluster install, node operations, GPU drivers, OS upgrades) belongs to `olares-cli node` / `gpu` / `os` and uses kubeconfig, NOT a profile. If the user asks "is my pod running?" answer here; if they ask "install Joplin" answer in `market`; if they ask "join a worker node" answer in `node`.

## Authentication transport

Every request goes through the factory-injected `*http.Client` and the resolved profile from `cmdutil.Factory`. There is no kubeconfig dependency.

- Base URL: **`rp.ControlHubURL`** = `https://control-hub.<terminus>`. Derived once in [`cli/pkg/credential/default_provider.go`](cli/pkg/credential/default_provider.go) from the OlaresID, via [`olares.ID.ControlHubURL`](cli/pkg/olares/id.go). The same nginx fans out to several path prefixes:
  - `/capi/*` — Olares custom aggregator (per-user app/workspace metadata, e.g. `/capi/app/detail`, `/capi/namespaces/group`).
  - `/api/v1/*` and `/apis/<group>/<version>/*` — wildcard K8s API proxy. Returns native K8s shapes (`{kind, apiVersion, metadata, spec, status}` for objects, `{kind, apiVersion, metadata, items}` for lists). PATCH targets here use `Content-Type: application/merge-patch+json` for `workload scale` / `cronjob suspend`/`resume`.
  - `/kapis/*` — KubeSphere aggregated API. Returns paginated `{items, totalItems}` envelopes. The KubeSphere **operations** API lives at `/kapis/operations.kubesphere.io/v1alpha2/...` and is what `cluster job rerun` POSTs to.
  - `/middleware/*` — Olares middleware controller (`/middleware/v1/list` for read; `/middleware/v1/<type>/password` for password rotation).
  - `/user-service/*` — BFL (re-uses the same auth chain as desktop / settings; not yet exercised here).
- Auth header: `X-Authorization: <access_token>` (NOT `Authorization: Bearer …`). Injected by the factory's `refreshingTransport` (see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go)); the `cluster` package's [`Prepare()`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) helper never calls `req.Header.Set("X-Authorization", …)` itself.
- **Expired access_tokens are auto-rotated.** When the server returns 401/403, the transport hits `/api/refresh`, persists the new token, and retries the original request once — transparently to the caller. Users do NOT need to run `profile login` just because their access_token aged out; only when the *refresh_token* itself is invalidated. Full mechanics — concurrency, cross-process flock, typed `*credential.ErrTokenInvalidated` / `*credential.ErrNotLoggedIn` errors — are documented in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) under "Automatic token refresh". **Do not write retry loops on top of these typed errors** — once you see one, only `profile login` / `profile import` will help.
- 401 / 403 that survive auto-refresh are reformatted into a CLI-friendly hint via [`reformatClusterAuthErr`](cli/pkg/clusterclient/client.go) — same wording template as the settings + market clients so the user only has to memorize one CTA.

## Security model — server decides, CLI never gates

This is a **hard requirement**, not a guideline. Every verb in `cluster ...` MUST:

1. Pass the request to the server with whatever scope the caller asked for (a namespace flag, a label selector, a positional name).
2. Trust the server's response. If the server returned an item, render it; if the server returned a 403, surface the error.
3. NEVER consult the locally cached cluster context (`ProfileConfig.ClusterContext` in [`cli/pkg/cliconfig/config.go`](cli/pkg/cliconfig/config.go)) to decide whether to make the call, what namespaces to filter to, or whether the user "should" see something.

This applies to mutating verbs too (`workload scale` / `restart` / `delete`, `pod delete` / `restart`, `cronjob suspend` / `resume`, `job rerun`, `middleware password set`). The CLI's only gate is `ConfirmDestructive` — a **UX guard**, not an authorization check. A 403 from the server is the authoritative "no"; we surface it verbatim and never preflight against the cached role.

### Why
- The ControlHub backend already enforces per-user scoping based on the access token. A cache-based local check adds attack surface — a user who tampers with `~/.olares-cli/config.json` could trick the CLI into showing UI as if they had `platform-admin`, even though the server still rejects every call.
- Caches drift silently after role changes. The server is always right.
- Mirroring `kubectl`'s split between client (issue requests, render responses) and server (authorize) keeps the model simple and audit-friendly.

The cached `ClusterContext` exists only so [`cluster context`](cli/cmd/ctl/cluster/context.go) can render identity / role / workspaces without a roundtrip and so error-wrap helpers can include the cached role in their messages.

## Top-level commands

### Identity

| Command | Endpoint | Notes |
|---|---|---|
| `cluster context [--refresh] [-o table\|json]` | `GET /capi/app/detail` | Identity + globalrole + accessible workspaces / system namespaces / granted clusters. Cache-first; `--refresh` forces a roundtrip and updates the cache. **Display only — never gates other verbs.** |

### Pods (`cluster pod ...`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster pod list [-n NS] [-l SEL] [--field-selector SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/pods` (or `/.../namespaces/<ns>/pods`) | Cross-namespace by default; the server returns the union of every namespace your token can see. NAMESPACE column appears in cross-namespace mode. |
| `cluster pod get <ns/name \| name> [-n NS] [-w] [--interval D]` | `GET /api/v1/namespaces/<ns>/pods/<name>` | Vertical key/value summary + per-container table in `-o table`. JSON forwarded verbatim in `-o json`. With `--watch` (`-w`) the GET repeats on `--interval` (default 2s); table mode clear-screen-redraws when stdout is a TTY, JSON mode emits one object per tick (JSONL). |
| `cluster pod yaml <ns/name \| name> [-n NS]` | `GET /api/v1/namespaces/<ns>/pods/<name>` | JSON-to-YAML round-trip via `sigs.k8s.io/yaml`. Faithful to every field the server returned (NOT decoded through the typed `Pod` struct). |
| `cluster pod events <ns/name \| name> [-n NS] [--limit N]` | `GET /api/v1/namespaces/<ns>/events` | Fetches all events in the namespace, then filters client-side to `involvedObject.kind=Pod, name=<pod>`. Sorted oldest-first by `lastTimestamp`. |
| `cluster pod logs <ns/pod \| pod> [-n NS] [-c NAME] [--tail N] [--since D] [-f] [--interval D] [--previous] [--timestamps]` | `GET /api/v1/namespaces/<ns>/pods/<name>/log?container=<c>` | Plain-text body forwarded to stdout. Container is auto-selected when the pod has exactly one; multi-container pods require `-c/--container` (the verb errors with the available list). `--follow` is poll-based: first fetch uses `--tail` (default 200) / `--since`, subsequent ticks use `sinceTime=<previous fetch start>`. `--previous` reads the previous container instance's buffer (after a crash) and is mutually exclusive with `--follow`. `--timestamps` (default true) asks the server to RFC3339-prefix every line — same default the SPA pins. Tolerates up to 5 consecutive transient errors before giving up; auth failures propagate immediately. Ctrl-C exits cleanly. |
| `cluster pod delete <ns/name \| name> [-n NS] [--yes] [--grace-period N]` | `DELETE /api/v1/namespaces/<ns>/pods/<name>` (optional `?gracePeriodSeconds=N`) | Wrapped in `ConfirmDestructive`. `--grace-period -1` (default) lets the apiserver use the pod's `terminationGracePeriodSeconds`; `0` forces immediate kill (matches `kubectl delete --grace-period=0`). For controller-managed pods the controller will recreate them (which is also what `restart` relies on). |
| `cluster pod restart <ns/name \| name> [-n NS] [--yes] [--grace-period N]` | (same DELETE as `delete` — alias verb) | Wire-identical to `pod delete`; the SPA's `restartPods` is bit-identical to `deletePod`. The verb name is the only difference; we offer both because the SPA pairs them. |

### Containers (`cluster container ...`)

Per-pod projection over the same `/api/v1/namespaces/<ns>/pods/<name>` body — no new HTTP surface.

| Command | Notes |
|---|---|
| `cluster container list <ns/pod \| pod> [-n NS]` | One row per `spec.containers[*]` fused with the matching `status.containerStatuses[*]`: CONTAINER \| IMAGE \| READY \| RESTARTS \| STATE \| PORTS. |
| `cluster container env <ns/pod \| pod> [-n NS] [--container NAME]` | Lists explicit env vars per container. `valueFrom` references render as `(from configMapKey/secretKey/fieldRef/resourceFieldRef ...)` — values are NOT resolved (no extra GETs against ConfigMap / Secret). `envFrom` (implicit imports) is intentionally NOT enumerated. |
| `cluster container logs <ns/pod/container \| ns/pod \| pod> [-n NS] [-c NAME] [--tail N] [--since D] [-f] [--interval D] [--previous] [--timestamps]` | Same wire endpoint as `cluster pod logs` (delegates to `pod.RunLogs`); the container alias adds an optional 3-segment positional `<ns>/<pod>/<container>` so users who already know the container name can skip `--container`. Container is mandatory here either way (this is the verb named after it). All other flags pass straight through and behave bit-exactly like `cluster pod logs`. |

### Workloads (`cluster workload ...`, alias `wl`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster workload list [-n NS] [--kind all\|deployment\|statefulset\|daemonset] [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/<kind>` (or `/.../namespaces/<ns>/<kind>`) | `--kind` defaults to `all` and fans out one request per kind in `[deployments, statefulsets, daemonsets]`, merging into a single table with a KIND column. Single-kind requests drop the KIND column. Singular / plural / short forms accepted (`deploy` / `sts` / `ds`). |
| `cluster workload get <ns/name \| name> [-n NS] --kind X` | `GET /apis/apps/v1/namespaces/<ns>/<kind>/<name>` | K8s native. `--kind` REQUIRED here (cannot be `all`). Vertical summary in table mode includes READY (kind-aware: `readyReplicas/replicas` for Deployment/StatefulSet, `numberReady/desiredNumberScheduled` for DaemonSet) + Availability + UpdateStrategy + Selector (paste straight into `cluster pod list -l ...`). |
| `cluster workload yaml <ns/name \| name> [-n NS] --kind X` | same endpoint as get | JSON-to-YAML round-trip; faithful to every field the server returned. |
| `cluster workload rollout-status <ns/name \| name> [-n NS] --kind X [-w] [--interval D] [--timeout D]` | `GET /apis/apps/v1/namespaces/<ns>/<kind>/<name>` | Reports whether the rollout has converged. **Convergence rule** (kind-aware, mirrors `kubectl rollout status`): `observedGeneration == metadata.generation` AND, for Deployment/StatefulSet, `updatedReplicas == spec.replicas` AND `readyReplicas == spec.replicas`; for DaemonSet, `updatedNumberScheduled == desiredNumberScheduled` AND `numberReady == desiredNumberScheduled`. Without `--watch`: one GET, prints a one-line status, exits 0 if converged or 2 if not (sentinel `clusteropts.ErrReported`). With `--watch`: re-poll on `--interval` (default 2s) until converged, `--timeout` (default 10m) elapses, or Ctrl-C. Each state change emits one line (table) or one JSON object (JSONL); steady ticks are silent. |
| `cluster workload scale <ns/name \| name> [-n NS] --kind X --replicas N [-w] [--interval D] [--timeout D] [--yes]` | `PATCH /apis/apps/v1/namespaces/<ns>/<kind>/<name>` body `{"spec":{"replicas":N}}` Content-Type `application/merge-patch+json` | DaemonSet rejected up-front (no replicas concept). `--replicas=0` triggers `ConfirmDestructive` (scale-to-zero pauses traffic — same as `stop`); other counts are reversible and silent. With `--watch` the verb chains into `rollout-status --watch` so users get "scaled and Ready" in one command. |
| `cluster workload restart <ns/name \| name> [-n NS] --kind X [--yes] [--concurrency N]` | (1) `GET /apis/apps/v1/namespaces/<ns>/<kind>/<name>` to read `spec.selector.matchLabels`; (2) `GET /api/v1/namespaces/<ns>/pods?labelSelector=<rebuilt>`; (3) parallel `DELETE /api/v1/namespaces/<ns>/pods/<name>` | SPA-aligned (matches `confirmHandler2` in `apps/.../Workloads/Detail.vue`). The controller recreates each pod from the workload template — we deliberately do NOT use the kubectl `restartedAt` annotation trick. `--concurrency` (default 5) bounds parallel deletes. `ConfirmDestructive` shows count + (truncated) pod names. |
| `cluster workload stop <ns/name \| name> [-n NS] --kind X [-w] [--yes]` | (alias for `scale --replicas=0`) | Justified verb because the SPA exposes a labeled "STOP" button. DaemonSet rejected (delete the workload instead). `--watch` chains into rollout-status convergence. |
| `cluster workload start <ns/name \| name> [-n NS] --kind X --replicas N [-w]` | (alias for `scale --replicas=N`) | Mirror of `stop`. `--replicas` REQUIRED (no cached previous count); must be `>= 1`. No `--yes` because starting a stopped workload is non-destructive. |
| `cluster workload delete <ns/name \| name> [-n NS] --kind X [--yes] [--propagation foreground\|background\|orphan]` | `DELETE /apis/apps/v1/namespaces/<ns>/<kind>/<name>?propagationPolicy=<P>` | CLI-original (the SPA has no direct workload-delete button). `--propagation=Foreground` (default) waits for the cascade; `Background` returns immediately; `Orphan` leaves dependents. `ConfirmDestructive`. |

### Application spaces (`cluster application ...`, alias `app`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster application list` | `GET /capi/namespaces/group` | One row per Namespace, grouped by KubeSphere workspace. Default `--label kubesphere.io/workspace!=kubesphere.io/devopsproject` matches the SPA. JSON output preserves the workspace grouping. |
| `cluster application get <namespace>` | `GET /api/v1/namespaces/<ns>` | Vertical Namespace detail with KubeSphere-flavored labels (workspace, alias, creator) lifted to top of the table; full label set rendered as a sub-block. |
| `cluster application workloads <namespace> [--kind ...] [-l ...] [--limit ...]` | (delegates to `cluster workload list -n <ns>`) | Convenience pivot from `application list` → "what workloads run here?". No client-side scope expansion — same server-side rules. |
| `cluster application pods <namespace> [-l ...] [--field-selector ...] [--limit ...]` | (delegates to `cluster pod list -n <ns>`) | Symmetric pivot for pods. |
| `cluster application status <namespace> [-w] [--interval D] [--events N]` | parallel fan-out: `/kapis/.../namespaces/<ns>/{deployments,statefulsets,daemonsets}` + `/kapis/.../namespaces/<ns>/pods` + `/api/v1/namespaces/<ns>/events` | **CLI-original aggregation** — the SPA spreads this across separate tabs. Three sections: workloads READY counts per kind, pod phase buckets (Running / Pending / Succeeded / Failed / Unknown / TOTAL), and the most recent `--events` events (default 5) sorted newest-first. Per-lane errors don't black out the rest of the snapshot — failed lanes render as `(failed: ...)`. With `--watch`: TTY → clear-screen-redraw on each tick; pipe → repeated tables; JSON mode → one object per tick (JSONL). Uses `golang.org/x/sync/errgroup` for the parallel fetch. |

### Namespaces (`cluster namespace ...`, alias `ns`)

K8s-flavored framing of the same resource the application tree exposes.

| Command | Endpoint | Notes |
|---|---|---|
| `cluster namespace list [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/namespaces` | Flat kubectl-style table: NAME / PHASE / WORKSPACE / AGE. WORKSPACE comes from the `kubesphere.io/workspace` label. |
| `cluster namespace get <name>` | `GET /api/v1/namespaces/<ns>` | Vertical K8s-style detail with full labels + annotations blocks. Use `cluster application get` for the workspace-first framing. |

### Nodes (`cluster node ...`, alias `nodes`)

Per-user view of the cluster's nodes — different from `olares-cli node` which uses kubeconfig for host-side maintenance.

| Command | Endpoint | Notes |
|---|---|---|
| `cluster node list [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/nodes` | kubectl-shaped table: NAME / STATUS / ROLES / AGE / VERSION / INTERNAL-IP. STATUS = Ready / `Ready,SchedulingDisabled` / NotReady / Unknown derived from the Ready condition + `spec.unschedulable`. |
| `cluster node get <name>` | `GET /kapis/resources.kubesphere.io/v1alpha3/nodes/<node>` | Vertical detail with Capacity / Allocatable (well-known keys: cpu / memory / pods / ephemeral-storage), Conditions, Taints, Addresses, full label list. |

### Jobs (`cluster job ...`, alias `jobs`)

K8s Jobs (`apis/batch/v1`).

| Command | Endpoint | Notes |
|---|---|---|
| `cluster job list [-n NS] [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/jobs` (or `/.../namespaces/<ns>/jobs`) | Cross-namespace by default. Columns: NAMESPACE \| NAME \| COMPLETIONS (`succeeded/completions`) \| STATUS (Complete / Failed / Suspended / Running / Failing / Pending — derived from `Status.Conditions[*]`) \| DURATION (`completionTime - startTime`, or `now - startTime` while running) \| AGE. |
| `cluster job get <ns/name \| name> [-n NS]` | `GET /apis/batch/v1/namespaces/<ns>/jobs/<name>` | Vertical summary including Active / Succeeded / Failed counts, Conditions list, and a `Controlled By: CronJob/<name>` line when the job has a CronJob owner reference. Exports `job.Get` for `job pods`/`rerun` to share. |
| `cluster job yaml <ns/name \| name> [-n NS]` | same endpoint | JSON-to-YAML round-trip. |
| `cluster job pods <ns/name \| name> [-n NS] [-l ADDITIONAL] [--field-selector SEL] [--limit N]` | (1) `job.Get` for `metadata.uid`; (2) `pod.RunList` with `labelSelector=controller-uid=<uid>[,<extra>]` | Two-step "find pods this job spawned". Mirrors the SPA's lazy-load tree. `--label` is ANDed onto the controller-uid clause server-side. Server-decides scoping. |
| `cluster job events <ns/name \| name> [-n NS] [--limit N]` | `GET /api/v1/namespaces/<ns>/events` | Same shape as `cluster pod events` filtered to `involvedObject.kind=Job, name=<job>`. Shares the `clusteropts.Event` typed view, `RenderEventsTable`, and `EventsListPath` URL builder with `cluster pod events` and `cluster application status`. Sorted oldest-first by `lastTimestamp`. |
| `cluster job rerun <ns/name \| name> [-n NS] [--yes]` | (1) `job.Get` for `metadata.resourceVersion`; (2) `POST /kapis/operations.kubesphere.io/v1alpha2/namespaces/<ns>/jobs/<name>?action=rerun&resourceVersion=<rv>` (no body) | KubeSphere operations action — server spawns a new pod attempt and updates `Status.Active` accordingly. `ConfirmDestructive`-wrapped (rerun is mutating). |

### CronJobs (`cluster cronjob ...`, aliases `cronjobs` / `cj`)

K8s CronJobs (`apis/batch/v1beta1` — different API version from Jobs).

| Command | Endpoint | Notes |
|---|---|---|
| `cluster cronjob list [-n NS] [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/cronjobs` (or `/.../namespaces/<ns>/cronjobs`) | Columns: NAMESPACE \| NAME \| SCHEDULE \| SUSPEND \| ACTIVE (count of `status.active[]`) \| LAST-SCHEDULE \| AGE. |
| `cluster cronjob get <ns/name \| name> [-n NS]` | `GET /apis/batch/v1beta1/namespaces/<ns>/cronjobs/<name>` | Vertical summary including SCHEDULE / SUSPEND / ConcurrencyPolicy / Active Jobs (sorted name list) / LAST-SCHEDULE / Job Template Selector (the labelSelector `cronjob jobs` will use). |
| `cluster cronjob yaml <ns/name \| name> [-n NS]` | same endpoint | JSON-to-YAML round-trip. |
| `cluster cronjob jobs <ns/name \| name> [-n NS] [--limit N]` | (1) `cronjob.Get` for `spec.jobTemplate.metadata.labels`; (2) `GET /apis/batch/v1/namespaces/<ns>/jobs?labelSelector=<derived>` | Two-step. Errors out clearly if the jobTemplate carries no labels (rather than fanning to "every job in the namespace"). Renders the same NAME / COMPLETIONS / STATUS / AGE columns as `cluster job list`. |
| `cluster cronjob suspend <ns/name \| name> [-n NS] [--yes]` | `PATCH /apis/batch/v1beta1/namespaces/<ns>/cronjobs/<name>` body `{"spec":{"suspend":true}}` Content-Type `application/merge-patch+json` | `ConfirmDestructive` (pauses scheduled runs). No-op short-circuit when already suspended (with a stderr note). |
| `cluster cronjob resume <ns/name \| name> [-n NS]` | same path body `{"spec":{"suspend":false}}` | NO `--yes` (re-enabling a paused schedule is non-destructive). No-op short-circuit when already active. Shares the `runToggle` body with `suspend`. |

### Middleware (`cluster middleware ...`, alias `mw`)

Olares-managed databases / queues / object stores via the `/middleware/v1/*` aggregator.

| Command | Endpoint | Notes |
|---|---|---|
| `cluster middleware list [-t TYPE] [--show-passwords]` | `GET /middleware/v1/list` | Custom envelope `{code, data:[MiddlewareItem]}` — NOT a K8s shape. Table columns: TYPE / NAME / NAMESPACE / NODES / ADMIN-USER. **Admin password is never printed in table mode**; in `-o json` it's redacted as `<redacted>` unless `--show-passwords` is explicitly set. `-t` filters client-side (case-insensitive) so a single fetch can be re-projected by type. |
| `cluster middleware password set --type X --name N --namespace NS --user U [--password P] [--yes]` | `POST /middleware/v1/<type>/password` body `{name, namespace, middleware, user, password}` (custom envelope `{code, message}`) | Sub-noun `password` (future-proof for `password rotate` / `password reveal`). `--password` is OPTIONAL and SHOULD usually be omitted: when not provided, the verb prompts twice (no echo, via `golang.org/x/term.ReadPassword`) and requires both entries to match. Passing `--password` on the command line leaks the secret into shell history. `ConfirmDestructive`-wrapped (a wrong `--name` will break the running instance). Type validation is client-side (against the same enum the SPA uses) so a typo fails before the prompt. The server's `code != 0/200` envelope is unwrapped into a returned error. JSON output never echoes the password (security: a `-o json` redirected to a log file would leak it). |

> The shape is always `olares-cli cluster <noun> <verb>`. Every verb runs against the currently-selected profile; switch with `olares-cli profile use <name>` ahead of time (there is no per-invocation override flag).

## Output convention

Same `-o table | json` flag set as `settings` and `market` (see [`cli/cmd/ctl/cluster/internal/clusteropts/options.go`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) for `AddOutputFlags`):

- `-o table` (default): tabwriter columns. List verbs add a `NAMESPACE` column when scope is cross-namespace; `get` verbs use a vertical key/value layout plus secondary tables (containers / conditions); paginated lists print a `(showing X of Y total — pass --limit Y to see more)` hint to stderr when truncated.
- `-o json`: pretty-printed JSON. Pod / event verbs decode through minimal typed structs and re-emit only the fields the CLI knows about; `cluster pod yaml` / `cluster job yaml` / `cluster workload yaml` / `cluster cronjob yaml` are the exception — they forward the server's bytes verbatim through a JSON→YAML conversion. **Mutating verbs synthesize a stable summary object** (e.g. `{operation, kind, namespace, name, replicas}` for `scale`) rather than forwarding the post-write apiserver response — JSON consumers care about whether the change took, not about every field of the object.
- `--quiet`: suppress all stdout; exit code indicates success/failure. Useful for `cluster pod get foo/bar -q && echo ok`.
- `--no-headers`: omit table headers (handy for shell pipelines).

### `--watch` semantics (uniform across the tree)

Verbs that support `--watch` (`pod get`, `workload rollout-status`, `application status`) all use the same plumbing:

- Polling, never streaming. Avoids chunked transfer encoding and matches `olares-cli market --watch`.
- `signal.NotifyContext(os.Interrupt, SIGTERM)` for graceful Ctrl-C; exits nil so scripts don't get a non-zero from a voluntary stop.
- Tolerates up to 5 consecutive transient errors before aborting (auth failures propagate immediately).
- TTY detection (`golang.org/x/term.IsTerminal`): clear-screen-redraw for table mode, raw stream for piped output, JSONL for `-o json`.
- "Only emit on change" for `rollout-status` (state-key squelching) so the output tracks real progress 1:1.

## Wire format mapping (which envelope when)

Every verb picks the right decode path based on the endpoint prefix; the package does not auto-detect. See [`cli/pkg/clusterclient/decode.go`](cli/pkg/clusterclient/decode.go) for the helpers:

| Endpoint prefix | Wire shape | Helper |
|---|---|---|
| `/kapis/resources.kubesphere.io/v1alpha3/...` | `{items: [...], totalItems: N}` | `clusterclient.GetKubeSphereList[T]` |
| `/kapis/operations.kubesphere.io/v1alpha2/...` | Opaque (action triggers; SPA discards body) | `clusterclient.Client.DoJSON(ctx, "POST", path, nil, nil)` (nil body, success = 2xx) |
| `/api/v1/...`, `/apis/<g>/<v>/...` (GET) | K8s native list `{kind, apiVersion, items, metadata}` OR object `{kind, apiVersion, metadata, spec, status}` | `clusterclient.GetK8sList[T]` (lists) / `clusterclient.GetK8sObject` (objects) |
| `/api/v1/.../pods/<name>/log` | `text/plain` | `clusterclient.GetRaw` (forwarded to stdout, never decoded) |
| `/api/v1/...`, `/apis/<g>/<v>/...` (PATCH) | K8s native object response (or empty) | `clusterclient.Patch[T]` / `clusterclient.Client.DoJSONWithContentType` — Content-Type **`application/merge-patch+json`** for `workload scale`, `cronjob suspend`/`resume` |
| `/api/v1/...`, `/apis/<g>/<v>/...` (DELETE) | `metav1.Status` (success ignored) | `clusterclient.Client.DoJSON(ctx, "DELETE", path, nil, nil)` |
| `/capi/app/detail`, `/capi/namespaces/group` | Custom (typed object or array, no envelope) | `clusterclient.Client.DoJSON` straight into a per-call typed struct |
| `/middleware/v1/list`, `/middleware/v1/<type>/password` | Custom envelope `{code, data:[...] or message}` (NOT K8s) | `clusterclient.Client.DoJSON` into a per-package envelope wrapper that unwraps `code != 0/200` into a returned error |
| Anything that should be forwarded byte-for-byte | Raw bytes | `clusterclient.GetRaw` (used by `cluster {pod,workload,job,cronjob} yaml`) |

Per-call typed structs live in the verb files (e.g. `Pod` in [`cli/cmd/ctl/cluster/pod/types.go`](cli/cmd/ctl/cluster/pod/types.go), `NamespaceGroup` in [`cli/cmd/ctl/cluster/application/list.go`](cli/cmd/ctl/cluster/application/list.go), `Job` / `CronJob` in their package `types.go`). They model only the fields the verb renders — we do NOT vendor `k8s.io/api` for shape.

## Mutating verb checklist

Every mutating verb in this tree follows the same template. When adding a new one:

1. Wrap in `clusteropts.ConfirmDestructive(os.Stderr, os.Stdin, message, assumeYes)` — even for "reversible" destructive UX (the prompt is the safety net, not the apiserver's typed error).
2. Expose a `--yes` / `-y` flag that maps to the `assumeYes` argument so scripts can opt out.
3. The wire endpoint and Content-Type live in this skill's wire-format table — match it precisely. PATCH bodies use `application/merge-patch+json` unless the SPA explicitly uses a different merge algorithm.
4. Synthesize a stable JSON-mode result struct (`{operation, ..., serverResourceVersion}`) — never forward the apiserver's response wholesale. Keeps consumers stable across K8s versions and avoids accidentally leaking secrets.
5. Server is the only authority. **Do NOT** preflight against the cached cluster context, even for "obvious" cases like "operator's role is `member` so they can't delete". A 403 from the server is the answer; surface it.

## Common errors → fixes

| Error message starts with | What it means | What to do |
|---|---|---|
| `server rejected the request (HTTP 401: …); please run: olares-cli profile login --olares-id <id>` | The token couldn't be refreshed, OR the server rejected even the refreshed token. | Run the suggested `olares-cli profile login`. If it keeps happening immediately after, check the OlaresID is correct (`olares-cli profile list`). |
| `server rejected the request (HTTP 403: …)` | The server says the role on this profile cannot perform this action / see this resource. **Do NOT second-guess client-side.** | Suggest `cluster context --refresh` to confirm the cached role matches the server (drift is the most common confusing case). If still 403, the user genuinely lacks the permission — escalate to whoever owns Olares admin. |
| `list pods: GET …: HTTP 404 (NotFound): …` | The namespace doesn't exist, OR the user can't see it (KubeSphere often returns 404 for "you don't have access" rather than 403). | Run `cluster application list` to see what namespaces the server thinks are visible to this profile. |
| `aborted by user` / `stdin is not a terminal — pass --yes to confirm: …` | The destructive-verb prompt was rejected (or the verb ran in a non-TTY context without `--yes`). | If interactive, answer `y`. If scripted, add `--yes`. |
| `passwords do not match` (from `middleware password set`) | The two no-echo prompts didn't agree. | Re-run the verb. |
| `decode … response: …` | The endpoint returned something we couldn't parse. | Re-run with `-o json` (or just look at the body in the error) to see the raw response shape. May indicate a server-side schema change. |
| `refresh token for … became invalid at …` (typed `*credential.ErrTokenInvalidated`) | The refresh_token itself is dead — auto-refresh can't recover. | `olares-cli profile login` (full re-auth). See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for full mechanics. |

## What's NOT here yet (and where to look instead)

| Want to … | Use |
|---|---|
| Install / uninstall / start / stop / upgrade an Olares app | `olares-cli market install / uninstall / start / stop / upgrade` (see [`olares-market`](../olares-market/SKILL.md)) |
| List / get Olares apps from the user's perspective (entrances, env, domain, policy) | `olares-cli settings apps ...` (see [`olares-settings`](../olares-settings/SKILL.md)) |
| Manage VPN devices / ACLs | `olares-cli settings vpn ...` (see [`olares-settings`](../olares-settings/SKILL.md)) |
| Cluster install / node join / upgrade | `olares-cli node ...` and `olares-cli os ...` (kubeconfig-based, NOT profile-based) |
| Resolve `valueFrom` env refs to actual ConfigMap / Secret values | Not yet — `cluster container env` shows the reference (`secretKey foo/k`) but does not GET the target. Future `--resolve` flag. |
| List `envFrom` (implicit configMapRef / secretRef sets) on a container | Not yet — only explicit `env: [...]` declarations are enumerated by `cluster container env`. Add when a verb actually needs the implicit set. |
| Bulk verbs (e.g. `cluster pod delete --all -l X`, `cluster workload delete --all`) | Not yet — pattern lifted from `cluster workload restart` (bounded concurrency + ConfirmDestructive showing the count) would slot straight in. Add when an operator workflow demands it. |
| `cronjob trigger-now` — fire a CronJob's job template once on demand | Not yet — the SPA has no dedicated endpoint; would need to clone `spec.jobTemplate` into a fresh Job ourselves. Defer until a real user need surfaces. |

## File map

| File | Purpose |
|---|---|
| [`cli/cmd/ctl/cluster/root.go`](cli/cmd/ctl/cluster/root.go) | Umbrella command, registers sub-trees. |
| [`cli/cmd/ctl/cluster/context.go`](cli/cmd/ctl/cluster/context.go) | `cluster context` — cobra glue around [`pkg/clusterctx`](cli/pkg/clusterctx). |
| [`cli/cmd/ctl/cluster/internal/clusteropts/options.go`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) | Shared `ClusterOptions` (output flags + `Prepare()` factory for `clusterclient.Client`). Lives under `internal/` to break the umbrella ↔ subpackage import cycle. |
| [`cli/cmd/ctl/cluster/internal/clusteropts/confirm.go`](cli/cmd/ctl/cluster/internal/clusteropts/confirm.go) | `ConfirmDestructive(prompt, in, message, assumeYes)` — lifted from `settings/vpn/common.go`; the destructive-verb UX guard for every mutating cluster verb. TTY check + `--yes` short-circuit + literal y/yes match. |
| [`cli/cmd/ctl/cluster/internal/clusteropts/argparse.go`](cli/cmd/ctl/cluster/internal/clusteropts/argparse.go) | `SplitNsName(nsFlag, arg)` — single source for the `<ns>/<name>` or `-n + name` positional grammar shared by every per-noun get/yaml/pods/events verb. |
| [`cli/cmd/ctl/cluster/internal/clusteropts/format.go`](cli/cmd/ctl/cluster/internal/clusteropts/format.go) | `DashIfEmpty(s)` and `Age(ts, now)` — single-source row-rendering helpers shared by every list/get table renderer. |
| [`cli/cmd/ctl/cluster/internal/clusteropts/yaml.go`](cli/cmd/ctl/cluster/internal/clusteropts/yaml.go) | `JSONToYAML(body)` — single source for the K8s native JSON-to-YAML conversion used by every `cluster <noun> yaml` verb. |
| [`cli/cmd/ctl/cluster/pod/`](cli/cmd/ctl/cluster/pod) | `cluster pod` verbs (`list`, `get` with `--watch`, `yaml`, `events`, `logs`, `delete`, `restart`); exports `pod.Get` + `pod.RunList` + `pod.RunLogs` + `pod.LogsOptions` + `pod.RunDelete` for sibling packages. The shared `Event` typed view + sort/render helpers live in `clusteropts/events.go`. |
| [`cli/cmd/ctl/cluster/container/`](cli/cmd/ctl/cluster/container) | `cluster container` verbs (`list`, `env`, `logs`). `list`/`env` are pure projections over `pod.Get`; `logs` is a thin alias over `pod.RunLogs` that adds the 3-segment `<ns>/<pod>/<container>` positional grammar. |
| [`cli/cmd/ctl/cluster/workload/`](cli/cmd/ctl/cluster/workload) | `cluster workload` verbs (`list`, `get`, `yaml`, `rollout-status`, `scale`, `restart`, `stop`, `start`, `delete`); covers Deployment + StatefulSet + DaemonSet via per-call `--kind`. Exports `workload.RunList`, `workload.NormalizeKind`, `workload.SingularKind`, `workload.RunScale`. `rollout-status` and `scale --watch` share a single convergence loop (no duplicated polling code). `restart` is SPA-aligned (GET selector → list pods → bounded-parallel DELETE). |
| [`cli/cmd/ctl/cluster/application/`](cli/cmd/ctl/cluster/application) | `cluster application` verbs (`list`, `get`, `workloads`, `pods`, `status`); workloads/pods delegate to the sibling packages. `status` is CLI-original — parallel `errgroup` fan-out across the underlying GETs with three-section table / one-object-per-tick JSONL. |
| [`cli/cmd/ctl/cluster/namespace/`](cli/cmd/ctl/cluster/namespace) | `cluster namespace` verbs (`list`, `get`) — K8s framing of the same resource as `application`. |
| [`cli/cmd/ctl/cluster/node/`](cli/cmd/ctl/cluster/node) | `cluster node` verbs (`list`, `get`). Per-user K8s view; not the host-side `olares-cli node` tree. |
| [`cli/cmd/ctl/cluster/job/`](cli/cmd/ctl/cluster/job) | `cluster job` verbs (`list`, `get`, `yaml`, `pods`, `events`, `rerun`). Exports `job.Get` for siblings. `pods` delegates to `pod.RunList` with `controller-uid=<uid>`; `events` reuses `clusteropts.Event` + the shared render/sort/URL helpers; `rerun` POSTs to the KubeSphere operations API. |
| [`cli/cmd/ctl/cluster/cronjob/`](cli/cmd/ctl/cluster/cronjob) | `cluster cronjob` verbs (`list`, `get`, `yaml`, `jobs`, `suspend`, `resume`). Exports `cronjob.Get` for siblings. `jobs` delegates to a K8s native `apis/batch/v1/.../jobs?labelSelector=…` GET (rebuilt from `spec.jobTemplate.metadata.labels`); `suspend`/`resume` PATCH `application/merge-patch+json` and share `runToggle`. |
| [`cli/cmd/ctl/cluster/middleware/`](cli/cmd/ctl/cluster/middleware) | `cluster middleware` verbs (`list`, plus the `password` sub-noun). Custom envelope; passwords redacted by default. |
| [`cli/cmd/ctl/cluster/middleware/password/`](cli/cmd/ctl/cluster/middleware/password) | `cluster middleware password set` — `POST /middleware/v1/<type>/password` with no-echo prompt + `ConfirmDestructive`. |
| [`cli/pkg/clusterclient/`](cli/pkg/clusterclient) | HTTP wrapper (`Client`, `DoJSON`, `DoJSONWithContentType`, `DoRaw`) + envelope decode helpers (`ListResponse[T]`, `K8sList[T]`, `GetKubeSphereList`, `GetK8sList`, `GetK8sObject`, `GetRaw`, `Patch[T]`). |
| [`cli/pkg/clusterctx/`](cli/pkg/clusterctx) | `cluster context` business logic (Endpoint, `Info`, `Display`, `FetchAndCache`, `Run`). Mirrors [`cli/pkg/whoami`](cli/pkg/whoami). |
| [`cli/pkg/cliconfig/config.go`](cli/pkg/cliconfig/config.go) | `ProfileConfig.ClusterContext` + `SetClusterContext` cache. Display only. |
| [`cli/pkg/credential/default_provider.go`](cli/pkg/credential/default_provider.go) | `ResolvedProfile.ControlHubURL` derivation. |
| [`cli/pkg/olares/id.go`](cli/pkg/olares/id.go) | `ID.ControlHubURL(localPrefix)` URL builder. |

## What this skill does NOT cover

- App lifecycle (install/uninstall/upgrade/start/stop/cancel/clone) — use [`olares-market`](../olares-market/SKILL.md).
- Settings UI mirror (users, appearance, vpn, network, gpu, video, search, backup, restore, advanced, integration, apps from settings perspective) — use [`olares-settings`](../olares-settings/SKILL.md).
- File browser (drive / sync) — use [`olares-files`](../olares-files/SKILL.md).
- Dashboard SPA proxy — use [`olares-dashboard`](../olares-dashboard/SKILL.md).
- Shared profile / login / token refresh mechanics — use [`olares-shared`](../olares-shared/SKILL.md). **Read this one first.**
