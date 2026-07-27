# cluster application (alias `app`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster application --help` and `olares-cli cluster application <verb> --help`.

"Application space" = a K8s Namespace seen through the KubeSphere lens (grouped by workspace). Same underlying resource as `cluster namespace`, but framed for app-level navigation. Read-only.

> **NOT to be confused with app-store lifecycle.** For install / uninstall / upgrade / stop / resume of an Olares app, use [`olares-market`](../../olares-market/SKILL.md). This tree is the runtime-state view of the resulting K8s namespaces.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | One row per Namespace, grouped by KubeSphere workspace |
| `get <namespace>` | Vertical detail with KubeSphere-flavored labels (workspace, alias, creator) lifted to the top |
| `workloads <namespace> [...]` | Pivot — delegates to `workload list -n <ns>` |
| `pods <namespace> [...]` | Pivot — delegates to `pod list -n <ns>` |
| `status <namespace>` | **CLI-original aggregation** — parallel fan-out producing 3 sections: workload READY per kind + pod phase buckets + recent events |

## The `status` aggregation (unique to this verb)

```bash
olares-cli cluster application status user-system-alice
```

Produces three sections in parallel:

1. **Workload READY counts per kind** — Deployment ready/desired, StatefulSet ready/desired, DaemonSet ready/desired
2. **Pod phase buckets** — Running / Pending / Failed / Succeeded / Unknown counts
3. **Recent events** — default 5 most recent (newest-first); `--events N` to override

Per-lane errors render as `(failed: ...)` — **one section's failure does not blackout the rest**. This means partial output is normal and useful; don't treat any failed section as a hard error of the whole verb.

`--watch` re-polls on `--interval` (default `2s`); there is no timeout flag — stop it with Ctrl-C.

## Examples

```bash
# Browse application spaces.
olares-cli cluster application list

# Detail of one namespace with KubeSphere framing.
olares-cli cluster application get user-system-alice

# Aggregated status snapshot.
olares-cli cluster application status user-system-alice

# Watch the status until things converge (or Ctrl-C).
olares-cli cluster application status user-system-alice -w --interval 3s

# Pivot helpers.
olares-cli cluster application workloads user-system-alice --kind deploy
olares-cli cluster application pods user-system-alice
```

## Agent notes

- For "what's the state of `<app>`?" questions, **`application status -w` is the right starting verb** — it gives a one-screen overview before drilling into specific pods.
- The pivot verbs (`workloads` / `pods`) accept all the underlying flags (`--kind`, `-l`, `--page`, etc.). For full flag reference, see `olares-cli cluster workload list --help` and `olares-cli cluster pod list --help`.
- **`list` is the right starting point for "what apps do I have on this cluster?"** — it shows all namespaces the active profile can see, grouped by workspace.
- If the user asks "is this app healthy?", run `application status` and report each section's outcome. **Don't silently swallow a `(failed: ...)` section** — surface the section name and the failure to the user.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `(failed: ...)` in one section of `status` | One sub-endpoint errored; others succeeded | Report the partial result; re-run with `-o json` for raw details if needed |
| 404 on `application get <ns>` | Active profile can't see this namespace, OR the namespace doesn't exist | `application list` to see what's visible |
