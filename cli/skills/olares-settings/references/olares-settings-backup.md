# settings backup

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings backup --help` and `olares-cli settings backup <noun> <verb> --help`.

Inspect backup plans and their snapshots, including backup status, schedules, storage locations, and snapshot history.

> **Different ingress prefix** — `<SettingsURL>/apis/backup/v1/*` rather than `/api/*`. The CLI handles this transparently, but errors carrying the path will mention `/apis/backup/v1/...`.

## Sub-tree

| Verb | Floor | Status | Purpose |
|---|---|---|---|
| `plans list` | normal | VERIFIED | List backup plans (with pagination) |
| `snapshots list <backup-id>` | normal | VERIFIED | List snapshots for one plan |

## `plans list`

```bash
olares-cli settings backup plans list                     # default 50
olares-cli settings backup plans list --limit 100
olares-cli settings backup plans list --offset 50 --limit 50
olares-cli settings backup plans list -o json
```

Returns each plan with name, ID, schedule, target (cloud / local / S3 / ...), status. Fields not surfaced in the table view are available via `-o json` (the SPA's axios interceptor unwraps `data.data`, which is why upstream code reads `{backups: [...]}` directly).

`--offset` / `--limit` mirror the SPA's pagination behavior. Default `--limit` is 50.

## `snapshots list <backup-id>`

```bash
olares-cli settings backup snapshots list <backup-id> --limit 50
olares-cli settings backup snapshots list <backup-id> -o json
```

Requires the **backup ID** (from `plans list`, NOT the plan name). Returns the snapshot timeline for that plan: snapshot ID, taken-at timestamp, size, status, restic / kopia snapshot identifier.

## Plan create / update

The CLI does **not** create or update backup plans — the only backup verbs are the read paths (`plans list`, `snapshots list`). **Create / edit plans in the SPA** (Settings → Backup); the CLI's read verbs then inspect what the SPA created.

## Agent best practices

- **For "show me my backups"** → `plans list` is the right verb. If the user then asks "what snapshots are in plan X", `snapshots list <backup-id>` (use the ID from `plans list -o json`).
- **NEVER paste the backup password into the agent transcript or shell history.** Even in scripted form, prefer reading from a secret manager / env var that doesn't persist in scrollback.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `plan '<name>' not found` | Wrong backup-id or name | `plans list` to enumerate |
| `snapshots list: backup-id required` | Missing positional | Provide the backup-id from `plans list` |
| `GET /apis/backup/v1/plans/backup: upstream returned code N` | Backup-server unreachable / 5xx | Check `settings advanced status` for backup-server health |
