# cluster middleware (alias `mw`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster middleware --help`, `olares-cli cluster middleware list --help`.

Olares-managed databases / queues / object stores. **NOT a K8s native resource** — uses the `/middleware/v1/*` aggregator with a custom envelope `{code, data:[MiddlewareItem]}`, not the K8s list/get shape.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | TYPE / NAME / NAMESPACE / NODES / ADMIN-USER (read-only inventory) |

> The CLI exposes **only** the read-only inventory verb. Password rotation, instance creation, and other mutations are not in `olares-cli cluster middleware`; manage those from LarePass / ControlHub SPA, or the per-middleware admin tooling (psql, mongosh, redis-cli, etc.) once you know the connection details.

## Output formats

- Table mode shows the middleware type, name, namespace, node count, and admin username.
- JSON mode provides structured middleware instance details for scripts and automation.

## Examples

```bash
# List middleware instances.
olares-cli cluster middleware list

# Filter by type (postgres / mysql / mongodb / redis / minio / nats / ...).
olares-cli cluster middleware list -t postgres

# Structured instance details.
olares-cli cluster middleware list -o json
```

## Agent notes

- For "what databases does this Olares have?" / "what middleware can I use?" questions, **`list` is the answer**. Filter by `-t TYPE` to scope.
- Manage database passwords in LarePass / the ControlHub SPA, or with the middleware's own admin tools.
- JSON output includes instance names, namespaces, and endpoints for scripts and automation.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| Empty `list` output | No middleware visible to this profile; OR the aggregator is unreachable | `cluster context` to confirm cluster connectivity; LarePass settings for middleware enrollment |
