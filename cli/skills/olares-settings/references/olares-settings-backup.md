# settings backup

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings backup --help` and `olares-cli settings backup <noun> <verb> --help`.

Manage backup plans (BFL backup-server) and the repository encryption password.

> **Different ingress prefix** — `<SettingsURL>/apis/backup/v1/*` rather than `/api/*`. The CLI handles this transparently, but errors carrying the path will mention `/apis/backup/v1/...`.

## Sub-tree

| Verb | Floor | Status | Purpose |
|---|---|---|---|
| `plans list` | normal | VERIFIED | List backup plans (with pagination) |
| `snapshots list <backup-id>` | normal | VERIFIED | List snapshots for one plan |
| `password set <name>` | normal | UNVERIFIED | Set / rotate the repository encryption password |

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

## `password set <name>` (UNVERIFIED)

```bash
# Interactive TTY prompt (default; no echo).
olares-cli settings backup password set my-plan

# Stdin pipe (scripted).
echo -n "my-secret-password" | olares-cli settings backup password set my-plan --password-stdin
```

- Sets / rotates the encryption password for one backup plan's repository.
- **Default reads from a TTY without echo.** `--password-stdin` reads once from stdin (newline-terminated; trim newline if your producer adds one).
- **Losing the password means losing the ability to decrypt existing snapshots.** The upstream cannot recover this password — make sure the user has it stored somewhere safe (password manager) BEFORE the call.

## Plan create / update — out of scope (today)

`plans create` and `plans update` are **not implemented yet**. They need either:

- A `--from-file plan.json` mode (still pending design — the full `BackupPolicy` + `LocationConfig` vector is large), or
- An upstream "create from defaults" shortcut that doesn't yet exist.

Until then, **create / update backup plans via the SPA** (Settings → Backup). The CLI's read verbs can still inspect plans created via the SPA.

## Agent best practices

- **For "rotate the backup password"** → always run interactively via TTY prompt unless explicitly scripted. The `--password-stdin` flow is for CI / non-interactive contexts only.
- **For "show me my backups"** → `plans list` is the right verb. If the user then asks "what snapshots are in plan X", `snapshots list <backup-id>` (use the ID from `plans list -o json`).
- **NEVER paste the backup password into the agent transcript or shell history.** Even in scripted form, prefer reading from a secret manager / env var that doesn't persist in scrollback.
- **Surface the irreversibility warning before any `password set` interactive flow.** The user must understand that the password is the only key to existing encrypted snapshots — there is no recovery path.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `plan '<name>' not found` | Wrong backup-id or name | `plans list` to enumerate |
| `snapshots list: backup-id required` | Missing positional | Provide the backup-id from `plans list` |
| `failed to read password from stdin: EOF` | `--password-stdin` invoked without piped input | Pipe from `echo -n` / a file / a secret manager |
| `GET /apis/backup/v1/plans/backup: upstream returned code N` | Backup-server unreachable / 5xx | Check `settings advanced status` for backup-server health |
