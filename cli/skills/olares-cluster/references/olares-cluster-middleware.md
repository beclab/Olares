# cluster middleware (alias `mw`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster middleware --help`, `olares-cli cluster middleware list --help`.

Olares-managed databases / queues / object stores. **NOT a K8s native resource** — uses the `/middleware/v1/*` aggregator with a custom envelope `{code, data:[MiddlewareItem]}`, not the K8s list/get shape.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | TYPE / NAME / NAMESPACE / NODES / ADMIN-USER (read-only inventory) |

> The CLI exposes **only** the read-only inventory verb. Password rotation, instance creation, and other mutations are not in `olares-cli cluster middleware`; manage those from LarePass / ControlHub SPA, or the per-middleware admin tooling (psql, mongosh, redis-cli, etc.) once you know the connection details.

## Sensitive-field handling on `list`

- **Table mode never shows admin passwords.** Always.
- **JSON mode redacts admin passwords as `<redacted>` by default.** Pass `--show-passwords` to include them in the JSON output.

This is a deliberate guard: a casual `cluster mw list -o json` should not splatter credentials onto a terminal or into a piped log. Use `--show-passwords` consciously — and only when you know where the output is going.

## Examples

```bash
# List middleware instances (passwords redacted by default).
olares-cli cluster middleware list

# Filter by type (postgres / mysql / mongodb / redis / minio / nats / ...).
olares-cli cluster middleware list -t postgres

# JSON with passwords (use carefully — output may end up in logs).
olares-cli cluster middleware list -o json --show-passwords
```

## Agent notes

- For "what databases does this Olares have?" / "what middleware can I use?" questions, **`list` is the answer**. Filter by `-t TYPE` to scope.
- For "rotate the database password" / "change the admin password" requests, **tell the user this CLI cannot do that anymore** — redirect them to LarePass / the ControlHub SPA, or to the middleware's own admin shell.
- Avoid piping `--show-passwords` JSON anywhere persistent (logs, files, CI output). The redaction default exists for a reason.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| Empty `list` output | No middleware visible to this profile; OR the aggregator is unreachable | `cluster context` to confirm cluster connectivity; LarePass settings for middleware enrollment |
