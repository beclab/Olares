---
name: olares-cluster
version: 0.2.0
description: "olares-cli cluster command tree: per-user K8s view of an Olares cluster via the ControlHub BFF (https://control-hub.<terminus>). Read-only surface: `cluster context` (identity / globalrole / accessible workspaces from /capi/app/detail), `cluster pod list / get / yaml / events`, `cluster container list / env` (per-pod container projection), `cluster workload list / get / yaml` (Deployment / StatefulSet / DaemonSet, --kind all by default, K8s native get), `cluster application list / get / workloads / pods` (Olares ApplicationSpaces grouped by KubeSphere workspace via /capi/namespaces/group; the workloads/pods aliases delegate to cluster workload list / cluster pod list with -n forced), `cluster namespace list / get` (raw K8s framing), `cluster node list / get` (per-user node view with kubectl-style STATUS/ROLES/AGE/VERSION columns), `cluster middleware list` (Olares-managed databases / queues / object stores, redacts admin password by default). Per-user resource scoping is ALWAYS enforced server-side; CLI verbs MUST NOT consult the locally cached cluster context to gate calls — the cache is for display only. Authentication uses the active profile's access_token via the factory's refreshingTransport (auto-rotates on 401/403). Wire formats handled: KubeSphere {items, totalItems} on /kapis/*, K8s native {kind, apiVersion, metadata, items|spec|status} on /api/v1/* and /apis/*, ControlHub /capi/* custom shapes (no envelope), Olares /middleware/v1/* envelope ({code, data:[]}). Use whenever the user asks about pods / containers / workloads / namespaces / application spaces / nodes / middleware / global roles on the per-user cluster view, NOT for app-store lifecycle (use `olares-cli market`) or host-side install/upgrade (use `olares-cli node` / `gpu` / `os`)."
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
  - `/api/v1/*` and `/apis/<group>/<version>/*` — wildcard K8s API proxy. Returns native K8s shapes (`{kind, apiVersion, metadata, spec, status}` for objects, `{kind, apiVersion, metadata, items}` for lists).
  - `/kapis/*` — KubeSphere aggregated API. Returns paginated `{items, totalItems}` envelopes.
  - `/middleware/*` — Olares middleware controller (not yet exercised by any verb).
  - `/user-service/*` — BFL (re-uses the same auth chain as desktop / settings; not yet exercised here).
- Auth header: `X-Authorization: <access_token>` (NOT `Authorization: Bearer …`). Injected by the factory's `refreshingTransport` (see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go)); the `cluster` package's [`Prepare()`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) helper never calls `req.Header.Set("X-Authorization", …)` itself.
- **Expired access_tokens are auto-rotated.** When the server returns 401/403, the transport hits `/api/refresh`, persists the new token, and retries the original request once — transparently to the caller. Users do NOT need to run `profile login` just because their access_token aged out; only when the *refresh_token* itself is invalidated. Full mechanics — concurrency, cross-process flock, typed `*credential.ErrTokenInvalidated` / `*credential.ErrNotLoggedIn` errors — are documented in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) under "Automatic token refresh". **Do not write retry loops on top of these typed errors** — once you see one, only `profile login` / `profile import` will help.
- 401 / 403 that survive auto-refresh are reformatted into a CLI-friendly hint via [`reformatClusterAuthErr`](cli/pkg/clusterclient/client.go) — same wording template as the settings + market clients so the user only has to memorize one CTA.

## Security model — server decides, CLI never gates

This is a **hard requirement**, not a guideline. Every verb in `cluster ...` MUST:

1. Pass the request to the server with whatever scope the caller asked for (a namespace flag, a label selector, a positional name).
2. Trust the server's response. If the server returned an item, render it; if the server returned a 403, surface the error.
3. NEVER consult the locally cached cluster context (`ProfileConfig.ClusterContext` in [`cli/pkg/cliconfig/config.go`](cli/pkg/cliconfig/config.go)) to decide whether to make the call, what namespaces to filter to, or whether the user "should" see something.

### Why
- The ControlHub backend already enforces per-user scoping based on the access token. A cache-based local check adds attack surface — a user who tampers with `~/.olares-cli/config.json` could trick the CLI into showing UI as if they had `platform-admin`, even though the server still rejects every call.
- Caches drift silently after role changes. The server is always right.
- Mirroring `kubectl`'s split between client (issue requests, render responses) and server (authorize) keeps the model simple and audit-friendly.

The cached `ClusterContext` exists only so [`cluster context`](cli/cmd/ctl/cluster/context.go) can render identity / role / workspaces without a roundtrip and so error-wrap helpers can include the cached role in their messages.

## Top-level commands (today)

### Identity

| Command | Endpoint | Notes |
|---|---|---|
| `cluster context [--refresh] [-o table\|json]` | `GET /capi/app/detail` | Identity + globalrole + accessible workspaces / system namespaces / granted clusters. Cache-first; `--refresh` forces a roundtrip and updates the cache. **Display only — never gates other verbs.** |

### Pods (`cluster pod ...`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster pod list [-n NS] [-l SEL] [--field-selector SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/pods` (or `/.../namespaces/<ns>/pods`) | Cross-namespace by default; the server returns the union of every namespace your token can see. NAMESPACE column appears in cross-namespace mode. |
| `cluster pod get <ns/name \| name> [-n NS]` | `GET /api/v1/namespaces/<ns>/pods/<name>` | Vertical key/value summary + per-container table in `-o table`. JSON forwarded verbatim in `-o json`. |
| `cluster pod yaml <ns/name \| name> [-n NS]` | `GET /api/v1/namespaces/<ns>/pods/<name>` | JSON-to-YAML round-trip via `sigs.k8s.io/yaml`. Faithful to every field the server returned (NOT decoded through the typed `Pod` struct). |
| `cluster pod events <ns/name \| name> [-n NS] [--limit N]` | `GET /api/v1/namespaces/<ns>/events` | Fetches all events in the namespace, then filters client-side to `involvedObject.kind=Pod, name=<pod>`. Sorted oldest-first by `lastTimestamp`. |

### Containers (`cluster container ...`)

Per-pod projection over the same `/api/v1/namespaces/<ns>/pods/<name>` body — no new HTTP surface.

| Command | Notes |
|---|---|
| `cluster container list <ns/pod \| pod> [-n NS]` | One row per `spec.containers[*]` fused with the matching `status.containerStatuses[*]`: CONTAINER \| IMAGE \| READY \| RESTARTS \| STATE \| PORTS. |
| `cluster container env <ns/pod \| pod> [-n NS] [--container NAME]` | Lists explicit env vars per container. `valueFrom` references render as `(from configMapKey/secretKey/fieldRef/resourceFieldRef ...)` — values are NOT resolved (no extra GETs against ConfigMap / Secret). `envFrom` (implicit imports) is intentionally NOT enumerated. |

### Workloads (`cluster workload ...`, alias `wl`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster workload list [-n NS] [--kind all\|deployment\|statefulset\|daemonset] [-l SEL] [--limit N]` | `GET /kapis/resources.kubesphere.io/v1alpha3/<kind>` (or `/.../namespaces/<ns>/<kind>`) | `--kind` defaults to `all` and fans out one request per kind in `[deployments, statefulsets, daemonsets]`, merging into a single table with a KIND column. Single-kind requests drop the KIND column. Singular / plural / short forms accepted (`deploy` / `sts` / `ds`). |
| `cluster workload get <ns/name \| name> [-n NS] --kind X` | `GET /apis/apps/v1/namespaces/<ns>/<kind>/<name>` | K8s native. `--kind` REQUIRED here (cannot be `all`). Vertical summary in table mode includes READY (kind-aware: `readyReplicas/replicas` for Deployment/StatefulSet, `numberReady/desiredNumberScheduled` for DaemonSet) + Availability + UpdateStrategy + Selector (paste straight into `cluster pod list -l ...`). |
| `cluster workload yaml <ns/name \| name> [-n NS] --kind X` | same endpoint as get | JSON-to-YAML round-trip; faithful to every field the server returned. |

### Application spaces (`cluster application ...`, alias `app`)

| Command | Endpoint | Notes |
|---|---|---|
| `cluster application list` | `GET /capi/namespaces/group` | One row per Namespace, grouped by KubeSphere workspace. Default `--label kubesphere.io/workspace!=kubesphere.io/devopsproject` matches the SPA. JSON output preserves the workspace grouping. |
| `cluster application get <namespace>` | `GET /api/v1/namespaces/<ns>` | Vertical Namespace detail with KubeSphere-flavored labels (workspace, alias, creator) lifted to top of the table; full label set rendered as a sub-block. |
| `cluster application workloads <namespace> [--kind ...] [-l ...] [--limit ...]` | (delegates to `cluster workload list -n <ns>`) | Convenience pivot from `application list` → "what workloads run here?". No client-side scope expansion — same server-side rules. |
| `cluster application pods <namespace> [-l ...] [--field-selector ...] [--limit ...]` | (delegates to `cluster pod list -n <ns>`) | Symmetric pivot for pods. |

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

### Middleware (`cluster middleware ...`, alias `mw`)

Olares-managed databases / queues / object stores via the `/middleware/v1/*` aggregator.

| Command | Endpoint | Notes |
|---|---|---|
| `cluster middleware list [-t TYPE] [--show-passwords]` | `GET /middleware/v1/list` | Custom envelope `{code, data:[MiddlewareItem]}` — NOT a K8s shape. Table columns: TYPE / NAME / NAMESPACE / NODES / ADMIN-USER. **Admin password is never printed in table mode**; in `-o json` it's redacted as `<redacted>` unless `--show-passwords` is explicitly set. `-t` filters client-side (case-insensitive) so a single fetch can be re-projected by type. |

> The shape is always `olares-cli cluster <noun> <verb>`. The umbrella always honors the global `--profile` flag.

## Output convention

Same `-o table | json` flag set as `settings` and `market` (see [`cli/cmd/ctl/cluster/internal/clusteropts/options.go`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) for `AddOutputFlags`):

- `-o table` (default): tabwriter columns. List verbs add a `NAMESPACE` column when scope is cross-namespace; `get` verbs use a vertical key/value layout plus secondary tables (containers / conditions); paginated lists print a `(showing X of Y total — pass --limit Y to see more)` hint to stderr when truncated.
- `-o json`: pretty-printed JSON. Pod / event verbs decode through minimal typed structs and re-emit only the fields the CLI knows about; `cluster pod yaml` is the exception — it forwards the server's bytes verbatim through a JSON→YAML conversion.
- `--quiet`: suppress all stdout; exit code indicates success/failure. Useful for `cluster pod get foo/bar -q && echo ok`.
- `--no-headers`: omit table headers (handy for shell pipelines).

## Wire format mapping (which envelope when)

Every verb picks the right decode path based on the endpoint prefix; the package does not auto-detect. See [`cli/pkg/clusterclient/decode.go`](cli/pkg/clusterclient/decode.go) for the helpers:

| Endpoint prefix | Wire shape | Helper |
|---|---|---|
| `/kapis/...` | `{items: [...], totalItems: N}` | `clusterclient.GetKubeSphereList[T]` |
| `/api/v1/...`, `/apis/<g>/<v>/...` | K8s native list `{kind, apiVersion, items, metadata}` OR object `{kind, apiVersion, metadata, spec, status}` | `clusterclient.GetK8sList[T]` (lists) / `clusterclient.GetK8sObject` (objects) |
| `/capi/app/detail`, `/capi/namespaces/group` | Custom (typed object or array, no envelope) | `clusterclient.Client.DoJSON` straight into a per-call typed struct |
| `/middleware/v1/list`, `/middleware/v1/...` | Custom envelope `{code, data:[...], message?}` (NOT K8s) | `clusterclient.Client.DoJSON` into a per-package envelope wrapper that unwraps `code != 0/200` into a returned error |
| Anything that should be forwarded byte-for-byte | Raw bytes | `clusterclient.GetRaw` (used by `cluster pod yaml` / `cluster workload yaml`) |

Per-call typed structs live in the verb files (e.g. `Pod` in [`cli/cmd/ctl/cluster/pod/types.go`](cli/cmd/ctl/cluster/pod/types.go), `NamespaceGroup` in [`cli/cmd/ctl/cluster/application/list.go`](cli/cmd/ctl/cluster/application/list.go)). They model only the fields the verb renders — we do NOT vendor `k8s.io/api` for shape.

## Common errors → fixes

| Error message starts with | What it means | What to do |
|---|---|---|
| `server rejected the request (HTTP 401: …); please run: olares-cli profile login --olares-id <id>` | The token couldn't be refreshed, OR the server rejected even the refreshed token. | Run the suggested `olares-cli profile login`. If it keeps happening immediately after, check the OlaresID is correct (`olares-cli profile list`). |
| `server rejected the request (HTTP 403: …)` | The server says the role on this profile cannot perform this action / see this resource. **Do NOT second-guess client-side.** | Suggest `cluster context --refresh` to confirm the cached role matches the server (drift is the most common confusing case). If still 403, the user genuinely lacks the permission — escalate to whoever owns Olares admin. |
| `list pods: GET …: HTTP 404 (NotFound): …` | The namespace doesn't exist, OR the user can't see it (KubeSphere often returns 404 for "you don't have access" rather than 403). | Run `cluster application list` to see what namespaces the server thinks are visible to this profile. |
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
| Rotate a middleware admin password | `cluster middleware password set` will ship in Phase 6 (destructive) using the SPA's `POST /middleware/v1/<type>/password` endpoint plus `confirmDestructive`. |
| Tail container logs / follow events | Phase 2 (logs) and Phase 3 (`--watch`) — not implemented yet. For now use `cluster pod events` repeatedly or `kubectl logs` against the underlying cluster directly. |
| Scale workloads / restart pods / delete pods | Phase 4 (scale/restart) and Phase 6 (destructive) — not implemented yet. Mutating verbs will reuse the same `confirmDestructive` pattern as `settings vpn devices delete` and the merge-patch+json plumbing the SPA's `patchWorkloadsControler` uses. |

## File map

| File | Purpose |
|---|---|
| [`cli/cmd/ctl/cluster/root.go`](cli/cmd/ctl/cluster/root.go) | Umbrella command, registers sub-trees. |
| [`cli/cmd/ctl/cluster/context.go`](cli/cmd/ctl/cluster/context.go) | `cluster context` — cobra glue around [`pkg/clusterctx`](cli/pkg/clusterctx). |
| [`cli/cmd/ctl/cluster/internal/clusteropts/options.go`](cli/cmd/ctl/cluster/internal/clusteropts/options.go) | Shared `ClusterOptions` (output flags + `Prepare()` factory for `clusterclient.Client`). Lives under `internal/` to break the umbrella ↔ subpackage import cycle. |
| [`cli/cmd/ctl/cluster/pod/`](cli/cmd/ctl/cluster/pod) | `cluster pod` verbs (`list`, `get`, `yaml`, `events`); exports `pod.Get` + `pod.SplitNsName` + `pod.RunList` for sibling packages. |
| [`cli/cmd/ctl/cluster/container/`](cli/cmd/ctl/cluster/container) | `cluster container` verbs (`list`, `env`). Pure projection over `pod.Get`. |
| [`cli/cmd/ctl/cluster/workload/`](cli/cmd/ctl/cluster/workload) | `cluster workload` verbs (`list`, `get`, `yaml`); covers Deployment + StatefulSet + DaemonSet via per-call `--kind`. Exports `workload.RunList`, `workload.NormalizeKind`, `workload.SingularKind`. |
| [`cli/cmd/ctl/cluster/application/`](cli/cmd/ctl/cluster/application) | `cluster application` verbs (`list`, `get`, `workloads`, `pods`); workloads/pods delegate to the sibling packages. |
| [`cli/cmd/ctl/cluster/namespace/`](cli/cmd/ctl/cluster/namespace) | `cluster namespace` verbs (`list`, `get`) — K8s framing of the same resource as `application`. |
| [`cli/cmd/ctl/cluster/node/`](cli/cmd/ctl/cluster/node) | `cluster node` verbs (`list`, `get`). Per-user K8s view; not the host-side `olares-cli node` tree. |
| [`cli/cmd/ctl/cluster/middleware/`](cli/cmd/ctl/cluster/middleware) | `cluster middleware` verbs (`list`). Custom envelope; passwords redacted by default. |
| [`cli/pkg/clusterclient/`](cli/pkg/clusterclient) | HTTP wrapper (`Client`, `DoJSON`, `DoRaw`) + envelope decode helpers (`ListResponse[T]`, `K8sList[T]`, `GetKubeSphereList`, `GetK8sList`, `GetK8sObject`, `GetRaw`). |
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
