---
name: olares-cluster
version: 0.5.0
description: "olares-cli cluster: per-user Kubernetes view of an Olares cluster via the ControlHub backend. Read pods / containers / workloads (Deployment/StatefulSet/DaemonSet) / application spaces (KubeSphere-grouped namespaces) / namespaces / nodes / jobs / cronjobs / Olares-managed middleware (databases, queues, object stores). Mutate via scale / restart / stop / start / delete on workloads, delete / restart on pods, suspend / resume on cronjobs, rerun on jobs, and password set on middleware — every mutating verb is wrapped in a confirmation prompt that --yes opts out of. Watch verbs (pod get -w, workload rollout-status -w, application status -w, pod/container logs -f) poll on --interval, never stream. Use whenever the user asks `what's running on my cluster?`, `tail logs of <pod>`, `restart / scale / delete this workload`, `who am I on this cluster?`, `suspend this cronjob`, `rerun this job`, or `rotate the password on this database`. Do NOT use for app-store lifecycle (use olares-cli market) or host-side install / node-join / OS upgrade (use olares-cli node / os / gpu — those go through kubeconfig, not a profile)."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli cluster --help"
---

# cluster (per-user K8s view)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md). It owns the profile model, login flow, automatic token refresh, and the auth-error recovery table that every command here depends on.**

## When to use this skill

Use `olares-cli cluster ...` when the user asks, against the cluster the active profile can see:

- "What pods / containers / workloads / jobs / cronjobs / namespaces / nodes are running?"
- "Tail / show logs of `<pod>` (or `<container>` of `<pod>`)"
- "Restart / scale / stop / start / delete `<workload>`" — the K8s controller, not the Olares app
- "Suspend / resume `<cronjob>`" or "rerun `<job>`"
- "What workspaces / application spaces can I see?"
- "Who am I on this cluster, what's my role?" (`cluster context`)
- "Rotate the admin password on this `<middleware>`"
- "What does this object's YAML look like?" (`cluster <noun> yaml`)

## When NOT to use — route to a sibling skill

| User intent | Use instead | Why |
|---|---|---|
| Install / uninstall / upgrade / start / stop an Olares **app** | [`olares-market`](../olares-market/SKILL.md) | App-store lifecycle, not K8s object lifecycle |
| Edit app entrances / domains / env / policy / ACL from the **user** perspective | [`olares-settings`](../olares-settings/SKILL.md) | The settings UI mirror, scoped to the user's apps |
| Browse / sync drive files | [`olares-files`](../olares-files/SKILL.md) | File API, not K8s |
| Cluster install / node join / OS upgrade / GPU drivers | `olares-cli node`, `olares-cli os`, `olares-cli gpu` | Kubeconfig-based host maintenance, NOT profile-based |
| Profile management, login, token refresh | [`olares-shared`](../olares-shared/SKILL.md) | Auth lives there |

> **Mental model:** if the question is *runtime state* of an existing cluster, you are here. If it's *lifecycle* of an Olares app or *day-zero* host setup, you are not.

## Core concepts

| Noun | Identifier grammar | What it is |
|---|---|---|
| **Pod** | `<ns>/<pod>` (or `-n NS <pod>`) | One running pod with one or more containers. |
| **Container** | `<ns>/<pod>/<container>` (or `-n NS <pod> -c NAME`) | A single container inside a pod (logs / env target). |
| **Workload** | `<ns>/<name>` + `--kind deployment\|statefulset\|daemonset` | The controller that owns pods. Subject of `scale` / `restart` / `stop` / `start` / `rollout-status`. |
| **Application space** | `<namespace>` | A KubeSphere-grouped K8s namespace. The "Olares Application Space" framing groups namespaces by workspace so the user sees apps, not raw K8s. |
| **Namespace** | `<name>` | The same K8s namespace, but with kubectl-style framing (no workspace grouping). |
| **Node** | `<name>` | A K8s node visible to the active profile. **Different** from `olares-cli node` (host maintenance). |
| **Job** | `<ns>/<name>` | A one-shot batch run (`apis/batch/v1`). |
| **CronJob** | `<ns>/<name>` | A scheduled Job template (`apis/batch/v1beta1`). |
| **Middleware** | `--type T --name N --namespace NS` | An Olares-managed database / queue / object store. NOT a K8s native resource — uses a separate `/middleware/v1/*` aggregator. |

## Resource relationships

```
Application space (namespace)
├── Workload  (Deployment | StatefulSet | DaemonSet)
│   └── Pod
│       └── Container       (logs | env)
├── Job
│   └── Pod
└── CronJob
    └── Job
        └── Pod

Cluster
├── Node                    (per-user view, not host maintenance)
└── Middleware              (DB / queue / object store, separate aggregator)

cluster context             (identity / role / accessible workspaces)
```

`cluster context` is identity-only; it does not own resources. Use it once at the start of a session to confirm "I am `<id>`, role `<X>`, can see `<workspaces>`" — never to gate other verbs (see "Server decides" below).

## Prerequisites and invariants

1. **Identity = the currently-selected profile.** Switch with `olares-cli profile use <name>` ahead of time. There is intentionally no per-invocation `--profile` override — agents must commit to one role up-front. See [`olares-shared`](../olares-shared/SKILL.md).
2. **The server decides what the active profile can see; the CLI never preflights.** Pass the request, render whatever the server returns. A 403 is the authoritative "no" — surface it. Never gate a call against the locally cached `cluster context`; that cache is for display only.
3. All requests go through `https://control-hub.<terminus>` and ride the active profile's `access_token` via `refreshingTransport`, which auto-rotates on 401/403 (full mechanics in [`olares-shared`](../olares-shared/SKILL.md) under "Automatic token refresh"). **Do not write retry loops on top of `*credential.ErrTokenInvalidated` / `*credential.ErrNotLoggedIn`** — once you see one, only `profile login` / `profile import` will help.
4. The same nginx fans out four prefixes: `/capi/*` (Olares custom aggregator), `/api/v1/*` and `/apis/<group>/<version>/*` (K8s native proxy), `/kapis/*` (KubeSphere paginated), `/middleware/v1/*` (Olares middleware aggregator). The right helper is picked per-call in [`cli/pkg/clusterclient/`](cli/pkg/clusterclient) — verbs do not auto-detect.

## Identity (`cluster context`)

| Command | Endpoint | What it does |
|---|---|---|
| `cluster context [--refresh] [-o table\|json]` | `GET /capi/app/detail` | Identity + global role + accessible workspaces / system namespaces / granted clusters. Cache-first; `--refresh` forces a roundtrip and updates the cache. |

> Cached `ClusterContext` exists ONLY so this verb can render without a roundtrip and so error helpers can include the cached role in messages. Never gate other verbs on it.

## Pods (`cluster pod ...`)

| Command | Endpoint | What it does |
|---|---|---|
| `cluster pod list [-n NS] [-l SEL] [--field-selector S] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/[namespaces/<ns>/]pods` | Lists pods. Cross-namespace by default; `NAMESPACE` column appears when scope is wider than one namespace. |
| `cluster pod get <ns/name> [-w] [--interval D]` | `GET /api/v1/namespaces/<ns>/pods/<name>` | Vertical summary + per-container table. With `-w`: clear-screen-redraw on TTY (table) / JSONL (json). |
| `cluster pod yaml <ns/name>` | same | JSON-to-YAML round-trip via `sigs.k8s.io/yaml`; faithful to every server field. |
| `cluster pod events <ns/name> [--limit N]` | `GET /api/v1/namespaces/<ns>/events` | Server returns every event in the namespace; CLI filters client-side to `involvedObject.kind=Pod, name=<pod>`. Sorted oldest-first. |
| `cluster pod logs <ns/pod> [-c NAME] [--tail N] [--since D] [-f] [--interval D] [--previous] [--timestamps]` | `GET /api/v1/.../pods/<name>/log?container=<c>` | Plain-text body forwarded to stdout. Multi-container pods require `-c`. `--follow` is poll-based (`sinceTime` advances each tick). `--previous` reads the prior instance's buffer (mutually exclusive with `--follow`). |
| `cluster pod delete <ns/name> [--yes] [--grace-period N]` | `DELETE /api/v1/namespaces/<ns>/pods/<name>` | Wrapped in `ConfirmDestructive`. `--grace-period -1` (default) honors the pod's `terminationGracePeriodSeconds`; `0` forces immediate kill. Controller-managed pods will be recreated. |
| `cluster pod restart <ns/name> [--yes] [--grace-period N]` | (same DELETE) | Alias verb — wire-identical to `pod delete`. The SPA pairs them so we do too. |

## Containers (`cluster container ...`)

Per-pod projection over the same `/api/v1/.../pods/<name>` body — no new HTTP surface.

| Command | What it does |
|---|---|
| `cluster container list <ns/pod>` | One row per `spec.containers[*]` fused with the matching `status.containerStatuses[*]`: CONTAINER \| IMAGE \| READY \| RESTARTS \| STATE \| PORTS. |
| `cluster container env <ns/pod> [--container NAME]` | Lists explicit `env: [...]` per container. `valueFrom` renders as `(from configMapKey/secretKey/fieldRef ...)` — values are NOT resolved. `envFrom` is intentionally not enumerated. JSON mode always emits a `{containers: [...]}` envelope (parseable even when empty). |
| `cluster container logs <ns/pod/container \| ns/pod \| pod> [-c NAME] [--tail N] [--since D] [-f] [--interval D] [--previous] [--timestamps]` | Same wire endpoint as `pod logs`, plus an extra 3-segment positional `<ns>/<pod>/<container>` so users who already know the container name can skip `-c`. |

## Workloads (`cluster workload ...`, alias `wl`)

| Command | Endpoint | What it does |
|---|---|---|
| `cluster workload list [-n NS] [--kind all\|deployment\|statefulset\|daemonset] [-l SEL] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/[namespaces/<ns>/]<kind>` | `--kind=all` (default) fans out one request per kind and merges into one table with a `KIND` column. Single-kind drops `KIND`. Singular / plural / short forms accepted (`deploy`, `sts`, `ds`). |
| `cluster workload get <ns/name> --kind X` | `GET /apis/apps/v1/namespaces/<ns>/<kind>/<name>` | `--kind` REQUIRED (no `all`). Vertical summary with kind-aware READY (`readyReplicas/replicas` or `numberReady/desiredNumberScheduled`), Availability, UpdateStrategy, Selector. |
| `cluster workload yaml <ns/name> --kind X` | same | JSON-to-YAML round-trip. |
| `cluster workload rollout-status <ns/name> --kind X [-w] [--interval D] [--timeout D]` | same | Reports whether the rollout has converged (kind-aware, mirrors `kubectl rollout status`). Without `-w`: one GET, exits 0 if converged or 2 if not. With `-w`: re-poll on `--interval` until converged, `--timeout` (default 10m), or Ctrl-C. Only emits on state change. |
| `cluster workload scale <ns/name> --kind X --replicas N [-w] [--interval D] [--timeout D] [--yes]` | `PATCH /apis/apps/v1/namespaces/<ns>/<kind>/<name>` body `{"spec":{"replicas":N}}` Content-Type `application/merge-patch+json` | DaemonSet rejected (no replicas). `--replicas=0` triggers `ConfirmDestructive`. With `-w` chains into `rollout-status -w`. |
| `cluster workload restart <ns/name> --kind X [--yes] [--concurrency N]` | (1) GET selector; (2) GET pods by selector; (3) parallel DELETE pods | SPA-aligned. The controller recreates each pod from the workload template. `--concurrency` (default 5) bounds parallel deletes. NOT the kubectl `restartedAt` annotation trick. |
| `cluster workload stop <ns/name> --kind X [-w] [--yes]` | (alias for `scale --replicas=0`) | DaemonSet rejected (delete the workload instead). Justified verb because the SPA exposes a labeled "STOP". |
| `cluster workload start <ns/name> --kind X --replicas N [-w]` | (alias for `scale --replicas=N`) | `--replicas` REQUIRED (no cached previous count). No `--yes` (non-destructive). |
| `cluster workload delete <ns/name> --kind X [--yes] [--propagation foreground\|background\|orphan]` | `DELETE /apis/apps/v1/namespaces/<ns>/<kind>/<name>?propagationPolicy=<P>` | CLI-original (the SPA has no direct workload-delete button). `Foreground` (default) waits for the cascade. |

## Application spaces (`cluster application ...`, alias `app`)

| Command | Endpoint | What it does |
|---|---|---|
| `cluster application list` | `GET /capi/namespaces/group` | One row per Namespace, grouped by KubeSphere workspace. |
| `cluster application get <namespace>` | `GET /api/v1/namespaces/<ns>` | Vertical detail with KubeSphere-flavored labels (workspace, alias, creator) lifted to the top. |
| `cluster application workloads <namespace> [...]` | (delegates to `workload list -n <ns>`) | Pivot from "what app spaces?" → "what workloads here?". |
| `cluster application pods <namespace> [...]` | (delegates to `pod list -n <ns>`) | Symmetric pivot for pods. |
| `cluster application status <namespace> [-w] [--interval D] [--events N]` | parallel fan-out over `/kapis/.../{deployments,statefulsets,daemonsets,pods}` + `/api/v1/.../events` | **CLI-original aggregation** — three sections: workload READY counts per kind, pod phase buckets, recent events (default 5, newest-first). Per-lane errors render as `(failed: ...)`; one section's failure does not blackout the rest. |

## Namespaces (`cluster namespace ...`, alias `ns`)

K8s framing of the same resource the application tree exposes.

| Command | Endpoint | What it does |
|---|---|---|
| `cluster namespace list [-l SEL] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/namespaces` | kubectl-style table: NAME / PHASE / WORKSPACE / AGE. WORKSPACE comes from the `kubesphere.io/workspace` label. |
| `cluster namespace get <name>` | `GET /api/v1/namespaces/<ns>` | Vertical K8s-style detail with full labels + annotations blocks. |

## Nodes (`cluster node ...`, alias `nodes`)

Per-user K8s view; **not** the host-side `olares-cli node` tree.

| Command | Endpoint | What it does |
|---|---|---|
| `cluster node list [-l SEL] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/nodes` | kubectl-shaped: NAME / STATUS / ROLES / AGE / VERSION / INTERNAL-IP. STATUS = Ready / `Ready,SchedulingDisabled` / NotReady / Unknown. |
| `cluster node get <name>` | same `/<node>` | Vertical detail with Capacity / Allocatable, Conditions, Taints, Addresses, full labels. |

## Jobs (`cluster job ...`, alias `jobs`)

K8s Jobs (`apis/batch/v1`).

| Command | Endpoint | What it does |
|---|---|---|
| `cluster job list [-n NS] [-l SEL] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/[namespaces/<ns>/]jobs` | NAMESPACE / NAME / COMPLETIONS / STATUS (Complete/Failed/Suspended/Running/...) / DURATION / AGE. |
| `cluster job get <ns/name>` | `GET /apis/batch/v1/namespaces/<ns>/jobs/<name>` | Vertical summary + Conditions + `Controlled By: CronJob/<name>` if present. |
| `cluster job yaml <ns/name>` | same | JSON-to-YAML round-trip. |
| `cluster job pods <ns/name> [-l ADDITIONAL] [...]` | (1) `job.Get` for `metadata.uid`; (2) `pod list` with `controller-uid=<uid>` | Two-step. `--label` is ANDed onto the controller-uid clause server-side. |
| `cluster job events <ns/name> [--limit N]` | `GET /api/v1/namespaces/<ns>/events` | Same shape as `pod events`, filtered to `involvedObject.kind=Job`. Shares `clusteropts.Event` + render/sort/URL helpers with `pod events` and `application status`. |
| `cluster job rerun <ns/name> [--yes]` | (1) `job.Get` for `resourceVersion`; (2) `POST /kapis/operations.kubesphere.io/v1alpha2/namespaces/<ns>/jobs/<name>?action=rerun&resourceVersion=<rv>` | KubeSphere operations action; server spawns a new pod attempt. `ConfirmDestructive`-wrapped. |

## CronJobs (`cluster cronjob ...`, aliases `cronjobs` / `cj`)

K8s CronJobs (`apis/batch/v1beta1` — different from Jobs).

| Command | Endpoint | What it does |
|---|---|---|
| `cluster cronjob list [-n NS] [-l SEL] [--page N\|--all]` | `GET /kapis/resources.kubesphere.io/v1alpha3/[namespaces/<ns>/]cronjobs` | NAMESPACE / NAME / SCHEDULE / SUSPEND / ACTIVE / LAST-SCHEDULE / AGE. |
| `cluster cronjob get <ns/name>` | `GET /apis/batch/v1beta1/namespaces/<ns>/cronjobs/<name>` | Vertical summary + ConcurrencyPolicy + Active Jobs + Job Template Selector. |
| `cluster cronjob yaml <ns/name>` | same | JSON-to-YAML round-trip. |
| `cluster cronjob jobs <ns/name> [--limit N]` | (1) `cronjob.Get` for `spec.jobTemplate.metadata.labels`; (2) `GET /apis/batch/v1/.../jobs?labelSelector=<derived>` | Two-step. Errors clearly if the jobTemplate carries no labels (rather than fanning to "every job in the namespace"). |
| `cluster cronjob suspend <ns/name> [--yes]` | `PATCH /apis/batch/v1beta1/.../cronjobs/<name>` body `{"spec":{"suspend":true}}` Content-Type `application/merge-patch+json` | `ConfirmDestructive`. No-op short-circuit when already suspended. |
| `cluster cronjob resume <ns/name>` | same path body `{"spec":{"suspend":false}}` | NO `--yes` (re-enabling is non-destructive). No-op short-circuit when already active. |

## Middleware (`cluster middleware ...`, alias `mw`)

Olares-managed databases / queues / object stores via the `/middleware/v1/*` aggregator. NOT a K8s native resource.

| Command | Endpoint | What it does |
|---|---|---|
| `cluster middleware list [-t TYPE] [--show-passwords]` | `GET /middleware/v1/list` | Custom envelope `{code, data:[...]}` — NOT K8s. TYPE / NAME / NAMESPACE / NODES / ADMIN-USER. **Admin password is never printed in table mode**; in `-o json` it's `<redacted>` unless `--show-passwords` is explicitly set. |
| `cluster middleware password set --type X --name N --namespace NS --user U [--password P] [--yes]` | `POST /middleware/v1/<type>/password` | Sub-noun `password` (future-proof for `password rotate` / `reveal`). **`--password` is OPTIONAL and SHOULD usually be omitted** — when not provided, the verb prompts twice (no echo, must match). Passing it on the command line leaks the secret into shell history. `ConfirmDestructive`-wrapped. JSON output never echoes the password. |

## Output conventions

Same `-o table | json` flag set as `settings` and `market`.

- `-o table` (default): tabwriter columns. List verbs add a `NAMESPACE` column when scope is cross-namespace; `get` verbs use a vertical key/value layout plus secondary tables; paginated lists print `(showing X of Y total — pass --limit Y to see more)` to stderr when truncated.
- `-o json`: pretty-printed JSON. List/get verbs decode through minimal typed structs and re-emit only the fields the CLI knows about. The four `* yaml` verbs are the exception — they forward server bytes verbatim through JSON→YAML.
- `-q` / `--quiet`: suppress all stdout; exit code carries success/failure. Useful for `cluster pod get foo/bar -q && echo ok`.
- `--no-headers`: omit table headers (handy for shell pipelines).
- **Mutating verbs synthesize a stable JSON summary** (e.g. `{operation, kind, namespace, name, replicas}`) rather than forwarding the apiserver's post-write response — JSON consumers care about whether the change took, not about every field of the object.

### Pagination (`--page N` / `--all`)

Every `list` verb under `pod` / `cronjob` / `job` / `namespace` / `node` / `workload` (and the `application pods` / `application workloads` wrappers) supports pagination. Defaults: `--limit 100`, `--page 1`. Pass `--page N` to walk pages, or `--all` to drain every page. Helper lives in [`cli/cmd/ctl/cluster/internal/clusteropts/pagination.go`](cli/cmd/ctl/cluster/internal/clusteropts/pagination.go).

### `--watch` / `--follow` semantics (uniform)

`pod get -w`, `workload rollout-status -w`, `application status -w`, `pod logs -f`, `container logs -f` all share the same plumbing:

- **Polling, never streaming.** Avoids chunked transfer encoding and matches `olares-cli market --watch`.
- `signal.NotifyContext(os.Interrupt, SIGTERM)` for graceful Ctrl-C; exits nil so scripts don't get a non-zero from a voluntary stop.
- Tolerates up to 5 consecutive transient errors before aborting; auth failures propagate immediately.
- TTY detection (`golang.org/x/term.IsTerminal`): clear-screen-redraw for table, raw stream for piped output, JSONL for `-o json`.
- `--interval D` / `--timeout D` are **rejected with an error** when their gate flag (`-w` or `-f`) isn't also set. Don't silently waste a flag.

## Mutating verb safety contract

Every mutating verb (`pod delete` / `pod restart`, all of `workload scale|restart|stop|start|delete`, `cronjob suspend`, `job rerun`, `middleware password set`) follows the same contract:

1. **Wrapped in `ConfirmDestructive`.** Even for "reversible" changes — the prompt is the safety net, not the apiserver's typed error. `--yes` / `-y` opts out so scripts work non-interactively. `cronjob resume`, `workload start`, and `pod logs` are NOT wrapped (non-destructive).
2. **No client-side authorization.** Server is the only authority. A 403 from the server is the answer; surface it. Never preflight against the cached cluster context.
3. **Stable JSON summary**, not the apiserver's response (see "Output conventions" above).

## Common errors → fixes

| Error message starts with | Meaning | Fix |
|---|---|---|
| `server rejected the request (HTTP 401: ...); please run: olares-cli profile login --olares-id <id>` | Auto-refresh failed, OR refreshed token still rejected. | Run the suggested `profile login`. |
| `server rejected the request (HTTP 403: ...)` | The active profile's role can't perform this. | Try `cluster context --refresh` to confirm the cached role matches the server. If still 403, the user genuinely lacks permission. |
| `... HTTP 404 (NotFound): ...` on a list verb | Namespace doesn't exist OR the user can't see it (KubeSphere often returns 404 instead of 403 for "no access"). | `cluster application list` to see what the server thinks is visible. |
| `aborted by user` / `stdin is not a terminal — pass --yes to confirm: ...` | Destructive prompt rejected, or non-TTY context without `--yes`. | Interactive: answer `y`. Scripted: add `--yes`. |
| `passwords do not match` (from `middleware password set`) | The two no-echo prompts disagreed. | Re-run. |
| `--interval requires --follow` / `--interval requires --watch` / `--timeout requires --watch` | Polling cadence flags set without their gate flag. | Add `-f` / `-w`, or drop the offending flag. |
| `decode ... response: ...` | Endpoint returned something we couldn't parse. | Re-run with `-o json` to see the raw shape. May indicate a server-side schema change. |
| `refresh token for ... became invalid at ...` (typed `*credential.ErrTokenInvalidated`) | The refresh_token itself is dead — auto-refresh can't recover. | `olares-cli profile login`. See [`olares-shared`](../olares-shared/SKILL.md). |

For the full auth-error matrix (`no access token for <id>`, `already authenticated`, `two-factor authentication required`, etc.) see [`olares-shared`](../olares-shared/SKILL.md).

## What's NOT here yet

| Want to ... | Status |
|---|---|
| Resolve `valueFrom` env refs to actual ConfigMap / Secret values | Not yet — `container env` shows the reference (`secretKey foo/k`) but does not GET the target. Future `--resolve` flag. |
| Enumerate `envFrom` (implicit configMapRef / secretRef sets) | Not yet — only explicit `env: [...]` declarations are listed. |
| Bulk verbs (e.g. `pod delete --all -l X`, `workload delete --all`) | Not yet — pattern from `workload restart` (bounded concurrency + `ConfirmDestructive` showing the count) would slot straight in. |
| `cronjob trigger-now` — fire a CronJob's job template once on demand | Not yet — the SPA has no dedicated endpoint; would need to clone `spec.jobTemplate` into a fresh Job ourselves. |

## Source layout (developer orientation)

For agents this is reference material, not a routing aid; skim only when modifying the verb implementation.

- [`cli/cmd/ctl/cluster/internal/clusteropts/`](cli/cmd/ctl/cluster/internal/clusteropts) — shared `ClusterOptions` (output flags + `Prepare()` factory), `ConfirmDestructive`, `SplitNsName`, `Age` / `DashIfEmpty`, `JSONToYAML`, `SleepContext`, `PaginationOptions` + `FetchAllKubeSphere`, and the unified `Event` type + `RenderEventsTable` + `EventsListPath` shared across `pod` / `job` / `application` events.
- [`cli/cmd/ctl/cluster/{pod,container,workload,application,namespace,node,job,cronjob,middleware}/`](cli/cmd/ctl/cluster) — one package per noun. Cross-package exports: `pod.RunList` / `pod.RunLogs` / `pod.RunDelete` / `pod.Get`, `job.Get`, `cronjob.Get`, `workload.RunScale` / `RunList` / `NormalizeKind`. Wrapper verbs (`application pods` / `application workloads`, `job pods`, `cronjob jobs`, `container logs`) reuse those exports rather than re-implementing the wire calls.
- [`cli/pkg/clusterclient/`](cli/pkg/clusterclient) — `Client`, `DoJSON`, `DoJSONWithContentType`, `DoRaw`, plus envelope decoders: `GetKubeSphereList[T]` (`/kapis/*` `{items, totalItems}`), `GetK8sList[T]` / `GetK8sObject` (`/api/v1/*`, `/apis/*` native), `GetRaw` (raw bytes for logs / yaml), `Patch[T]`. Per-call typed structs live in each verb's `types.go` — we do NOT vendor `k8s.io/api`.
- [`cli/pkg/clusterctx/`](cli/pkg/clusterctx) — `cluster context` business logic; mirrors [`cli/pkg/whoami`](cli/pkg/whoami).
- [`cli/pkg/credential/default_provider.go`](cli/pkg/credential/default_provider.go), [`cli/pkg/olares/id.go`](cli/pkg/olares/id.go) — `ResolvedProfile.ControlHubURL` derivation from OlaresID.

## See also

- [`olares-shared`](../olares-shared/SKILL.md) — profile model, login, automatic token refresh, full auth-error recovery table. **Read this one first.**
- [`olares-market`](../olares-market/SKILL.md) — Olares app lifecycle (install / uninstall / upgrade / start / stop / cancel / clone).
- [`olares-settings`](../olares-settings/SKILL.md) — settings UI mirror (users, appearance, vpn, network, gpu, video, search, backup, restore, advanced, integration, apps).
- [`olares-files`](../olares-files/SKILL.md) — drive / sync file browser.
- [`olares-dashboard`](../olares-dashboard/SKILL.md) — dashboard SPA proxy.
