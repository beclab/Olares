# cluster pod

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster pod --help` and `olares-cli cluster pod <verb> --help`.

Inspect Pods visible to the active profile. **Cross-namespace by default** — list verbs do NOT pass a namespace, so responses are the union of every namespace the active profile can see (matches the SPA Pods page). Pass `-n / --namespace` to scope.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | One-table view; adds a `NAMESPACE` column in cross-ns mode |
| `get <ns/name>` | Vertical summary + per-container table; `-w` polls |
| `yaml <ns/name>` | Full K8s-native YAML (JSON-to-YAML round-trip, faithful to every server field) |
| `events <ns/name>` | Filtered to `involvedObject.kind=Pod, name=<pod>`. Sorted oldest-first |
| `logs <ns/pod>` | Plain-text body to stdout; `-f` polls (sinceTime advances each tick); multi-container pods need `-c` |
| `delete <ns/name>` | **Destructive** — wrapped in `ConfirmDestructive`. Controller-managed pods will be recreated |
| `restart <ns/name>` | **Destructive** — alias verb, wire-identical to `delete`. SPA pairs them so we do too |

## Safety constraints

`delete` / `restart` follow the parent SKILL.md's Mutating verb safety contract (y/N confirm, `--yes` for scripts, server decides). Pod-specific points:

- `--grace-period -1` (default) honors the pod's `terminationGracePeriodSeconds`; `0` forces immediate kill.
- **Most Olares pods are controller-managed** — `delete` will recreate them. That's usually the user's actual intent ("restart this pod") but confirm before invoking.

## Logs polling semantics

`logs -f` is **poll-based, not streamed**. Each tick re-requests with `sinceTime=<previous fetch start>` so no bytes are lost.

- `--tail N` — initial fetch only; after the first tick, `--follow` advances by `sinceTime` regardless of `--tail`.
- `--since D` — initial fetch window (e.g. `5m`, `1h`); `0` = unlimited.
- `--previous` — fetches the previous container instance's buffer (after a crash). **Mutually exclusive with `--follow`** (upstream API restriction).
- `--timestamps` (default true) — server prefixes each line with an RFC3339 timestamp. Matches the SPA so output is correlatable across windows.
- `--limit-bytes N` — cap response body size (0 = unlimited).

## Examples

```bash
# Cross-ns listing.
olares-cli cluster pod list

# Scope to a namespace.
olares-cli cluster pod list -n user-system-alice

# Watch a single pod every 2s.
olares-cli cluster pod get user-system-alice/my-pod -w

# Tail logs of a specific container.
olares-cli cluster pod logs user-system-alice/my-pod -c app -f --since 5m

# Last 50 lines of the previous container instance after a crash.
olares-cli cluster pod logs user-system-alice/my-pod -c app --previous --tail 50

# Destructive — must confirm.
olares-cli cluster pod delete user-system-alice/my-pod         # prompts y/N
olares-cli cluster pod delete user-system-alice/my-pod --yes   # script mode
```

## Agent notes

- For "tail the logs" requests in interactive sessions, prefer `--tail 200 --since 5m -f` so the user sees recent context immediately + new lines as they arrive.
- **Multi-container pods**: `logs <ns/pod>` errors without `-c`; the error message lists the containers. Show that to the user and ask which one.
- `pod restart` is just `pod delete` with a different label — explain it that way when the user asks. The pod will be **recreated** by its controller (Deployment / StatefulSet / DaemonSet); a bare Pod has no recreation and just disappears.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `pod has multiple containers; please specify --container` | Multi-container pod without `-c` | Add `-c <name>` (the error lists them) |
| `--previous is incompatible with --follow` | Both flags set | Drop one |
| 404 on a known-existing pod | Cross-tenant visibility or wrong namespace | `cluster pod list -n <ns>` to confirm the active profile can see it |
