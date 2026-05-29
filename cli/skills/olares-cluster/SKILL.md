---
name: olares-cluster
version: 4.0.0
description: "Olares ControlHub view (olares-cli cluster) — per-Olares-ID command-line mirror of the Olares ControlHub SPA's Cluster page, the Kubernetes view of an Olares instance. Reads pods, containers, workloads (Deployment / StatefulSet / DaemonSet), application spaces (KubeSphere-grouped namespaces), namespaces, nodes, jobs, cronjobs, and Olares-managed middleware (databases, queues, object stores) that the active Olares ID is allowed to see. Mutating verbs (scale / restart / stop / start / delete on workloads, delete / restart on pods, suspend / resume on cronjobs, rerun on jobs) all go through a confirmation prompt that --yes skips. Watch verbs (pod get -w, workload rollout-status -w, application status -w, pod / container logs -f) poll on --interval. Use when the user mentions Olares, Olares ID, Olares ControlHub, olares-cli cluster, or asks 'what's running on my Olares', 'tail logs of <pod>', 'restart / scale / delete this workload', 'who am I on this ControlHub', 'suspend this cronjob', or 'rerun this job'. Do NOT use for Olares app-store lifecycle (use olares-cli market) or host-side install / node join / OS upgrade (use olares-cli node / os / gpu)."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli cluster --help"
---

# cluster (per-user K8s view)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for the profile model, login flow, automatic token refresh, and the auth-error recovery table.**

> **Source of truth for flags & wire shapes is always `olares-cli cluster <noun> <verb> --help`.** This file only carries what `--help` cannot give: routing, the mental model of nouns, the identity-vs-server principle, the mutating-verb safety contract, cross-verb output conventions, and the common-errors → fix table.

## When to use this skill

Against the cluster the active profile can see:

- "What pods / containers / workloads / jobs / cronjobs / namespaces / nodes are running?"
- "Tail / show logs of `<pod>` (or `<container>` of `<pod>`)"
- "Restart / scale / stop / start / delete `<workload>`" — the K8s controller, not the Olares app
- "Suspend / resume `<cronjob>`" or "rerun `<job>`"
- "Who am I on this cluster, what's my role?" (`cluster context`)
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
| **Pod** | `<ns>/<pod>` (or `-n NS <pod>`) | One running pod with one or more containers |
| **Container** | `<ns>/<pod>/<container>` (or `-n NS <pod> -c NAME`) | A single container inside a pod (logs / env target) |
| **Workload** | `<ns>/<name>` + `--kind deployment\|statefulset\|daemonset` | The controller that owns pods. Subject of `scale` / `restart` / `stop` / `start` / `rollout-status` |
| **Application space** | `<namespace>` | A KubeSphere-grouped K8s namespace; the "Olares Application Space" framing groups namespaces by workspace |
| **Namespace** | `<name>` | The same K8s namespace, kubectl-style framing (no workspace grouping) |
| **Node** | `<name>` | A K8s node visible to the active profile. **Different** from `olares-cli node` (host maintenance) |
| **Job** | `<ns>/<name>` | A one-shot batch run (`apis/batch/v1`) |
| **CronJob** | `<ns>/<name>` | A scheduled Job template (`apis/batch/v1`) |
| **Middleware** | `--type T --name N --namespace NS` | An Olares-managed database / queue / object store; NOT a K8s native resource (separate `/middleware/v1/*` aggregator) |

### Resource relationships

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

## The identity-vs-server-decides principle (cross-cutting)

1. **Identity = the currently-selected profile.** Switch with `olares-cli profile use <name>` ahead of time. There is no per-invocation `--profile` override — agents must commit to one role up-front.
2. **The server decides what the active profile can see; the CLI never preflights.** Pass the request, render whatever the server returns. **A 403 is the authoritative "no" — surface it.** Never gate a call against the locally cached `cluster context`; that cache is for display only.
3. All requests go through `https://control-hub.<terminus>` and ride the active profile's `access_token` via the auto-refreshing transport. See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for refresh mechanics and `ErrTokenInvalidated` recovery.
4. The same nginx fans out four prefixes: `/capi/*` (Olares aggregator), `/api/v1/*` + `/apis/<g>/<v>/*` (K8s native), `/kapis/*` (KubeSphere paginated), `/middleware/v1/*` (Olares middleware aggregator). The right helper is picked per-call by the CLI.

> `cluster context` is identity-only — it does NOT own resources. Use it once at the start of a session to confirm "I am `<id>`, role `<X>`, can see `<workspaces>`". Never gate other verbs on its cached value.

## Verb index

For flags, examples, and wire shapes, **always start with `olares-cli cluster <noun> <verb> --help`**.

| Noun | Verbs | `--help` first, then... |
|---|---|---|
| `context` | (single verb) | `olares-cli cluster context --help` |
| `pod` | `list`, `get`, `yaml`, `events`, `logs`, `delete`, `restart` | [references/olares-cluster-pod.md](references/olares-cluster-pod.md) |
| `container` | `list`, `env`, `logs` | `olares-cli cluster container --help` |
| `workload` (alias `wl`) | `list`, `get`, `yaml`, `rollout-status`, `scale`, `restart`, `stop`, `start`, `delete` | [references/olares-cluster-workload.md](references/olares-cluster-workload.md) |
| `application` (alias `app`) | `list`, `get`, `workloads`, `pods`, `status` | [references/olares-cluster-application.md](references/olares-cluster-application.md) |
| `namespace` (alias `ns`) | `list`, `get` | `olares-cli cluster namespace --help` |
| `node` (alias `nodes`) | `list`, `get` | `olares-cli cluster node --help` |
| `job` (alias `jobs`) | `list`, `get`, `yaml`, `pods`, `events`, `rerun` | [references/olares-cluster-job.md](references/olares-cluster-job.md) |
| `cronjob` (aliases `cronjobs` / `cj`) | `list`, `get`, `yaml`, `jobs`, `suspend`, `resume` | [references/olares-cluster-cronjob.md](references/olares-cluster-cronjob.md) |
| `middleware` (alias `mw`) | `list` | [references/olares-cluster-middleware.md](references/olares-cluster-middleware.md) |

## Mutating verb safety contract (cross-cutting)

Every mutating verb — `pod delete` / `pod restart`, all of `workload scale|restart|stop|start|delete`, `cronjob suspend`, `job rerun` — follows the same contract:

1. **Wrapped in a `ConfirmDestructive` y/N prompt.** Even for "reversible" changes — the prompt is the safety net. **`--yes` / `-y` opts out for scripts.**
2. **No client-side authorization preflight** — server is the only authority. A 403 surfaces; the CLI does not gate against `cluster context`.
3. **Stable JSON summary on success**, not the apiserver's response. JSON consumers care whether the change took, not every field of the object.

Non-destructive verbs (`cronjob resume`, `workload start`, `pod logs`) are NOT wrapped. **Confirm intent with the user BEFORE invoking any destructive verb, even when scripts pass `--yes`.**

## Output conventions

- `-o table` (default): tabwriter columns. List verbs add a `NAMESPACE` column when scope is cross-namespace; `get` verbs render a vertical key/value layout. Paginated lists print `(showing X of Y total — pass --limit Y to see more)` to stderr when truncated.
- `-o json`: pretty-printed JSON. List/get verbs decode through minimal typed structs and re-emit only the fields the CLI knows about. The four `* yaml` verbs forward server bytes verbatim through JSON→YAML.
- `-q` / `--quiet`: suppress all stdout; exit code carries success/failure.
- `--no-headers`: omit table headers (handy for shell pipelines).

### Pagination (`--page N` / `--all`)

Every `list` verb under `pod` / `cronjob` / `job` / `namespace` / `node` / `workload` (and the `application pods` / `application workloads` wrappers) supports pagination. Defaults: `--limit 100`, `--page 1`. Pass `--page N` to walk pages, or `--all` to drain every page.

### `--watch` / `--follow` semantics (uniform)

`pod get -w`, `workload rollout-status -w`, `application status -w`, `pod logs -f`, `container logs -f` all share the same plumbing:

- **Polling, never streaming.** Avoids chunked transfer encoding; matches `olares-cli market --watch`.
- `signal.NotifyContext(os.Interrupt, SIGTERM)` for graceful Ctrl-C; exits nil on voluntary stop.
- Tolerates up to 5 consecutive transient errors (network blips, 5xx) before aborting. **Terminal 4xx responses short-circuit immediately** — a NotFound / Forbidden / Unauthorized won't fix itself on the next poll. `408` and `429` are still retried.
- TTY detection: clear-screen-redraw for table, raw stream for piped output, JSONL for `-o json`.
- `--interval D` / `--timeout D` are **rejected with an error** when their gate flag (`-w` or `-f`) isn't also set — don't silently waste a flag.

## Common errors → fixes

| Error message starts with | Meaning | Fix |
|---|---|---|
| `server rejected the request (HTTP 401: ...); please run: olares-cli profile login --olares-id <id>` | Auto-refresh failed, OR refreshed token still rejected | Run the suggested `profile login` |
| `server rejected the request (HTTP 403: ...)` | The active profile's role can't perform this | `cluster context --refresh` to confirm the cached role matches the server. If still 403, the user genuinely lacks permission |
| `... HTTP 404 (NotFound): ...` on a list verb | Namespace doesn't exist OR the user can't see it (KubeSphere often returns 404 instead of 403 for "no access") | `cluster application list` to see what the server thinks is visible |
| `--field-selector: field "..." is not supported (supported: ...)` (from `cluster pod list`) | The pod list `--field-selector` accepts only a translatable subset of kubectl selectors (KubeSphere doesn't speak the raw `fieldSelector=` wire syntax) | Use one of the supported fields (`status.phase`, `spec.nodeName`, `metadata.name`, `metadata.namespace`), or drop `--field-selector` and filter client-side |
| `--field-selector: "..." uses the '!=' operator which the upstream KubeSphere pods endpoint does not support` | KubeSphere only matches equality | Rephrase as a positive match, or filter the output through `jq` |
| `aborted by user` / `stdin is not a terminal — pass --yes to confirm: ...` | Destructive prompt rejected, or non-TTY context without `--yes` | Interactive: answer `y`. Scripted: add `--yes` |
| `--interval requires --follow` / `--interval requires --watch` / `--timeout requires --watch` | Polling cadence flags set without their gate flag | Add `-f` / `-w`, or drop the offending flag |
| `decode ... response: ...` | Endpoint returned something we couldn't parse | Re-run with `-o json` to see the raw shape; may indicate a server-side schema change |
| `refresh token for ... became invalid at ...` (typed `*credential.ErrTokenInvalidated`) | The refresh_token itself is dead — auto-refresh can't recover | `olares-cli profile login` |

For the full auth-error matrix see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
