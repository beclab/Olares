# settings users

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) (especially "Role caching + admin/normal floor") first.
> **Flags & examples:** `olares-cli settings users --help` and `olares-cli settings users <verb> --help`.

Instance roster — list / get / create / delete users on the active Olares. **Admin floor** (`owner` / `admin` only).

> **Not the same as `olares-cli user create`** (kube / cluster CR). `settings users` hits Termipass user-service routes; `user` writes a cluster-level CR. Mention this if the user mixes them up.

## Verbs at a glance

| Verb | Floor | Notes |
|---|---|---|
| `list` | admin | Server-side role-filtered (normal users see only themselves) |
| `get <username>` | admin | Single user record (email / olaresId / role / quota) |
| `me` | normal | Alias of `profile whoami` (same driver) |
| `create <username>` | admin | Two flag groups: `--defaults` OR explicit `--role` / `--cpu` / `--memory-gb` |
| `delete <username>` | admin | **Destructive**. Owner accounts refused. Whole-word `yes` prompt unless `--yes` |

## `create` — two mutually-exclusive flag groups

```bash
# Group A: SPA preset (normal, 1 CPU, 4G memory; auto-generated password; DID precheck)
olares-cli settings users create bob --defaults
olares-cli settings users create bob --defaults --watch

# Group B: explicit
olares-cli settings users create alice --role admin --cpu 2 --memory-gb 8
olares-cli settings users create alice --role normal --cpu 1 --memory-gb 4 --watch
```

- `--defaults` is mutually exclusive with `--role` / `--cpu` / `--memory-gb`.
- `--role` may only be `admin` or `normal`. **The SPA account dialog cannot create an owner**, and neither can the CLI.
- The initial password is **always auto-generated** by the CLI — no `--password` flag exists. The password is printed once to stdout on success; treat transcripts as sensitive.
- DID precheck runs before the create POST. A username that conflicts with an existing DID fails fast with a copy-pasteable error.

## `delete` — destructive, owner-safe

```bash
olares-cli settings users delete bob                       # whole-word "yes" prompt
olares-cli settings users delete bob --yes                 # script mode
olares-cli settings users delete bob --yes --watch         # block until Deleted
```

- **Owner accounts cannot be deleted** — the CLI rejects via GET before sending DELETE, matching `olares-cli user delete`.
- The prompt expects the **whole word `yes`** (like ssh `yes/no`), not `y`. Reduces accidental confirmation.
- Default is accepted-then-exit — the row may still appear briefly in `users list` while controllers tear it down. Use `--watch` to block until `Deleted`.

## `--watch` (block until terminal state)

Same shape as [`olares-cli market --watch`](../../olares-market/SKILL.md): opt-in `-w/--watch`, companion `--watch-timeout` (default 15m), `--watch-interval` (default 2s). Polls user-service's `/status` endpoint. Gives up after 5 consecutive transport errors. SIGINT-graceful.

| Op | Terminal-success | Terminal-failure | absentMeansSuccess |
|---|---|---|---|
| `create` | `Created` | `Failed` / `Deleted`* | false |
| `delete` | `Deleted` | (timeout only) | true (defensive HTTP 404) |

\* `Deleted` during a create watch means the row vanished mid-watch (controller cleanup raced) — reported as failure so JSON consumers see a non-zero exit.

## JSON output shapes

```jsonc
// Without --watch
{"name":"alice","original_password":"...","status":"Accepted"}
{"name":"bob","status":"Accepted"}

// With --watch (final_status added)
{"name":"alice","original_password":"...","status":"Created","final_status":"Created","wizard_url":"https://wizard-alice.example.com"}
{"name":"bob","status":"Deleted","final_status":"Deleted"}
```

`wizard_url` appears on a successful create `--watch` — surface it to the user so they can finish setting up the new account.

## Agent best practices

- **Always run `users get <name>` before `users delete <name>`** in interactive sessions to confirm role and quota — gives the user a chance to back out if they targeted the wrong account.
- For "create user and tell me when they can log in" → `users create <name> --defaults --watch -o json | jq '.wizard_url'`.
- **NEVER paste the auto-generated password into chat.** Recommend the user copy it directly from the CLI output.
- `--watch` is **opt-in**: without it, `users create` / `users delete` return as soon as the request is accepted. Scripts that need blocking semantics must pass `--watch` explicitly.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `cannot create user: --role required (or use --defaults)` | Group A vs B confusion | Pick one set of flags |
| `cannot delete owner user` | DELETE blocked client-side | Owner accounts are unmanageable; use a different verb / SPA action |
| `confirmation rejected: type the whole word 'yes' to delete` | Y/N-style answer to whole-word prompt | Type `yes`, or pass `--yes` |
| `did precheck failed: <reason>` | Username conflicts with an existing DID | Pick a different username |
| Watcher hangs near `Failed` | Backend create / provisioning failed | Check `users get <name>` for `status`; retry or rollback |
