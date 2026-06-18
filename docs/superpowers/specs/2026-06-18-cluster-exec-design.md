# Design: `olares-cli cluster {pod,container} exec`

Status: Approved (brainstorming)
Date: 2026-06-18
Area: `cli/cmd/ctl/cluster`, `apps/docker/system-frontend/nginx`

## Problem

`olares-cli cluster {pod,container} logs` lets an AI agent read container logs to
troubleshoot. But logs are passive: sometimes the agent (or a human) needs to run
commands *inside* a container to investigate further — check files, processes,
connectivity, env, etc. There is no exec capability in the CLI today.

## Goal

Add an `exec` verb under both `cluster pod` and `cluster container` (mirroring
`logs`) that can:

- run a single command in a container and return its output + exit code
  (one-shot, the default — optimized for AI iterative troubleshooting), and
- attach an interactive TTY shell for a human (`-it`, like `kubectl exec -it`).

The intent is to let an AI agent enter a container to **both troubleshoot and
perform repairs/fixes** (edit config, restart a process, install a tool, etc.),
not just read state.

Non-goals (YAGNI): `node exec` (host shell), `cp` (file copy), port-forward,
**stateful exec sessions** (see "Stateless vs stateful" below).

## Decisions (locked during brainstorming)

1. Both modes: one-shot default + `-it` interactive.
2. Wire protocol: **native Kubernetes exec** over WebSocket
   (`/api/v1/namespaces/<ns>/pods/<pod>/exec`, subprotocol `v5.channel.k8s.io`,
   auto-fallback to `v4`). Chosen over the KubeSphere terminal WS because native
   exec gives separated stdout/stderr **and a real exit code** — both critical
   for an AI to judge success deterministically.
3. Interactive `-it` requires a `y/N` confirmation before entering.
4. The required edge nginx change ships **within this design** (not a separate PR).

## Feasibility (verified)

Native exec WS is feasible end-to-end. Every downstream layer already supports
WebSocket upgrade; the only gap is the edge nginx whitelist.

| Layer | Component | WS upgrade | Evidence |
|---|---|---|---|
| Edge | control-hub nginx (`dashboard-control-hub.conf`) | NO for `/api` | whitelist regex only matches `kapis/terminal`, `*watch` |
| Middle | system-server kube-rbac-proxy `:28080` (the `monitoring` ExternalName target) | YES | Go `httputil.ReverseProxy`, upgrade-aware and path-agnostic (`rbac-proxy-single.go`) |
| Aggregator | ks-apiserver `/api` filter | YES | `UpgradeAwareHandler` + `UpgradeTransport` (`kubeapiserver.go:56-57`) |
| Source | kube-apiserver | YES | native `v4/v5.channel.k8s.io` exec |

Two corroborations: (a) the SPA terminal already works through this chain via
`/kapis/terminal.../exec` (so the whole chain passes WS), and (b) `cluster pod
logs` already hits `/api/v1/.../pods/.../log` through the same chain (so `/api/v1`
is within the rbac-proxy allowPaths). Therefore `/api/v1/.../exec` is both an
allowed path and on a WS-capable chain; the sole blocker is the edge nginx.

CLI deps are already present: `k8s.io/client-go v0.34.2` (has the
`tools/remotecommand` WebSocket executor) and `github.com/gorilla/websocket`.

## CLI surface

```
olares-cli cluster pod exec <ns/pod | pod> [-n NS] [-c CONTAINER] [flags] -- CMD [args...]
olares-cli cluster container exec <ns/pod/container | ns/pod | pod> [-c CONTAINER] [flags] -- CMD [args...]
```

Container selection reuses the `logs` rules: single-container auto-select;
multi-container errors with the list; `-c` overrides.

Flags:

- `-i, --stdin` — keep stdin open to the container.
- `-t, --tty` — allocate a TTY. `-it` is the interactive shorthand.
- `--timeout DUR` — one-shot only; bounded wait (default `60s`). On expiry:
  abort, return captured partial output + a typed timeout error, `exitCode=null`.
- `--max-output-bytes N` — one-shot only; per-stream cap (default `2MiB`).
  Truncate and mark when exceeded.
- `-o, --output table|json` — one-shot only (see AI contract below).
- `-q, --quiet` — suppress stdout; exit code carries success/failure.
- `-y, --yes` — skip the `-it` confirmation prompt.

Argument passing: everything after `--` is the argv passed verbatim to the
container (no implicit shell). For pipes/redirects/vars, the caller writes
`-- sh -c '...'` explicitly.

### One-shot (default)

- No TTY, stdin closed (unless `--stdin`). Output is clean text, no ANSI/prompt/echo.
- stdout and stderr captured separately.
- The in-container command's exit code becomes the CLI process exit code.
- A non-zero command exit is a *normal result*, not a CLI error: do NOT print the
  cobra `Error:` banner for it. Reserve error reporting for transport/auth/
  protocol failures (the existing `clusterclient.HTTPError` path).

### Interactive (`-it`)

- Requires a real local TTY; in a non-TTY context (e.g. an AI tool call) `-it`
  is refused with guidance to use one-shot instead.
- Prompts `y/N` before attaching (shared `ConfirmDestructive` style). `--yes`
  skips it. Confirmation + TTY requirement naturally keep AI agents on one-shot.
- Wires local stdin/stdout/stderr to the stream; handles SIGWINCH resize.
- Default command is `sh` when none is given.

## AI-friendly contract

1. `-o json` one-shot result shape:

   ```json
   {
     "namespace": "...",
     "pod": "...",
     "container": "...",
     "command": ["cat", "/etc/hosts"],
     "stdout": "...",
     "stderr": "...",
     "exitCode": 0,
     "truncated": false,
     "durationMs": 123
   }
   ```

   `exitCode` is `null` on timeout/abort. `truncated` is true when either stream
   hit `--max-output-bytes`.

2. Exit code is the authoritative success signal. JSON consumers read `exitCode`;
   shell consumers read `$?`.
3. One-shot is non-interactive by default: no TTY, stdin closed → deterministic,
   parseable output for agents; bounded by `--timeout` and `--max-output-bytes`
   so a hung or chatty command can't stall or flood the agent.
4. Update the `olares-cluster` skill (`cli/skills/olares-cluster/`): add `exec`
   to the verb index and document, for the AI:
   - the JSON contract + exit-code semantics;
   - **ephemeral vs durable fixes** (in-container changes revert on restart;
     durable fixes go through `workload`/ConfigMap/image);
   - compose multi-step repairs with `-- sh -c '...'` (stateless);
   - canonical file-edit recipes (`--stdin` + `cat >`, `sed -i`, `tee`);
   - that exec is RBAC-gated and server-side audited.

## Designing for AI repair work

The premise is an AI agent that *fixes* things inside containers, not only reads
them. That raises issues a read-only logs flow never hits:

### Stateless vs stateful

One-shot exec is **stateless**: each command is a fresh process, so `cd`, shell
variables, and interactive editors do not persist across calls. We deliberately
keep it stateless rather than building stateful sessions:

- It covers the vast majority of repairs because any sequence can be composed in
  one call: `-- sh -c 'cd /app && sed -i ... && kill -HUP 1'`. Filesystem effects
  (e.g. `apk add curl`) DO persist in the running container across calls — only
  process/shell state doesn't.
- Stateless commands are self-contained, deterministic, and reproducible — which
  is *easier* for an AI to reason about than hidden session state (current cwd,
  prior exports).
- A stateful session would require a long-lived connection held by a background
  daemon between short-lived CLI invocations, plus a session registry, output
  cursoring, timeouts/GC, and concurrency handling — a large, fragile subsystem.
  Deferred as YAGNI; revisit only if a concrete need appears.

Guidance to encode in the skill: for multi-step repairs, compose with
`-- sh -c '...'` rather than expecting state to carry between calls.

### Ephemeral vs durable fixes (CRITICAL)

Changes made inside a running container are **ephemeral** — a pod restart or any
controller-driven recreation (rollout, eviction, node drain) reverts them. exec
is a *hotfix* tool. Durable fixes must change the source of truth: the image, a
ConfigMap/Secret, env, or the Deployment/StatefulSet spec (the `cluster workload`
path). The skill MUST state this explicitly so the AI does not report a transient
in-container change as a permanent fix.

### Writing files / editing config

Editing is core to repair, but there is no interactive editor. Supported patterns:

- `--stdin` streams the CLI's stdin into the container, enabling
  `exec ... -i -- sh -c 'cat > /etc/app.conf'` with the new content piped in.
- In-place edits via `-- sh -c "sed -i 's/old/new/' /etc/app.conf"` or
  `-- sh -c 'printf "%s" "..." | tee /path'`.

The skill documents these as the canonical "AI edits a file" recipes.

### Environment constraints → clear errors

Repairs fail in predictable ways that need legible messages, not raw dumps:

- No shell/tools in image (distroless/scratch): "no `sh` in container" hint.
- Read-only root filesystem / non-root user / restrictive securityContext:
  surface `EROFS` / `EACCES` with a "container filesystem is read-only or
  permission-restricted" hint.

### Safety stance for mutating exec

One-shot stays **non-prompting** (per-command y/N would defeat AI autonomy, and
command intent can't be reliably classified as destructive). The guardrails are:

- `pods/exec` RBAC (server-side SAR) gates *who* can exec at all.
- Server-side audit: ks-apiserver already audit-logs exec subresource calls, so
  every AI exec is attributable without building new audit plumbing. The skill
  notes that exec actions are audited.
- `-it` interactive (human) keeps its y/N confirmation.

## Implementation approach (WS client)

- **Preferred:** `k8s.io/client-go/tools/remotecommand` WebSocket executor. It
  handles channel framing, stdout/stderr demux, TTY, resize, and exit-code
  parsing. Build a `*rest.Config` pointed at `ControlHubURL` and inject the
  existing `X-Authorization` refreshing transport via `WrapTransport`. Risk:
  adapting client-go's WS handshake to a non-Bearer `X-Authorization` header.
- **Fallback:** `github.com/gorilla/websocket` direct handshake (trivially adds
  the `X-Authorization` header) plus a ~100-line v4 channel framer. Maximum
  control, fewest moving parts, if header injection into client-go proves awkward.
- Decision deferred to the implementation plan; try preferred first.

## Auth & safety

- Reuse the Factory refreshing transport; handshake carries `X-Authorization`.
  401/403 map to the existing friendly CTA (`profile login` /
  `cluster context --refresh`).
- Server enforces `pods/exec` (create) via SubjectAccessReview; no client-side
  preflight (identity-vs-server principle). Missing permission → 403 surfaced.
- One-shot does NOT prompt (matches `kubectl exec`; relies on RBAC). Only `-it`
  prompts.

## Error taxonomy

| Condition | Surfaced as |
|---|---|
| Command not found in container | exit code `127` (normal result, not CLI error) |
| Not executable | exit code `126` |
| No shell in image (distroless/scratch) | exit/stderr maps to a "no `sh` in container" hint |
| Read-only fs / permission denied on write | `EROFS`/`EACCES` surfaced with "filesystem read-only or permission-restricted" hint |
| Pod/container not running | `HTTPError` (4xx, terminal) with pod state hint |
| No `pods/exec` permission | 403 + `profile login` / `context --refresh` CTA |
| Old Olares (edge nginx lacks exec WS whitelist) | handshake not upgraded → typed "exec not supported on this Olares version; please upgrade" |
| `--timeout` exceeded | typed timeout error + partial output, `exitCode=null` |

## Server-side change (required, in-scope)

Edit both server blocks of
`apps/docker/system-frontend/nginx/dashboard-control-hub.conf` to add the exec
path to the WS-upgrade regex location:

```nginx
location ~ /(kapis/terminal|api/v1/watch|apis/apps/v1/watch|api/v1/namespaces/[^/]+/pods/[^/]+/exec) {
    proxy_pass http://SettingsServer;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Host $http_host;
}
```

This is a deploy-time change (ships with the `system-frontend` image / an Olares
release). CLI does capability detection: if the handshake comes back without an
upgrade, surface the "not supported on this version" error above.
```
