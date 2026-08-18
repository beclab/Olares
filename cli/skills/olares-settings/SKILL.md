---
name: olares-settings
version: 4.7.0
description: "Olares Settings via olares-cli settings — mirror the Settings SPA: users, apps, VPN, network, backup, integrations, GPU/compute, search-index directories/rebuild, and me/whoami. Use for system or post-install app configuration; not for querying file contents with olares-search."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# settings (Olares Settings UI mirror)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Load the shared [platform model](../olares-shared/references/olares-platform.md) when version semantics affect the task. Use `olares-cli settings <area> <verb> --help` for syntax.

## When to use

- Mirror Settings UI configuration: users, app entrances/env/domain/policy, integrations, VPN, network, accelerators, search indexing, backup/restore, appearance, video, and advanced status.
- Inspect the current identity, Olares version, updates, or SSO sessions.
- Perform supported post-install configuration writes.

> **Mental model:** `settings` covers configuration that the Olares Settings SPA exposes — **post-install per-app config**, mesh / VPN, backup, accounts, system appearance. Lifecycle and runtime live in sibling skills.

## Verb index

| Area | Verbs / resources | Read when triggered |
|---|---|---|
| `me` | `whoami`, `version`, `check-update`, `sso list` | `settings me --help` |
| `users` | `me`, `list`, `get`, `create`, `delete` | [user lifecycle and password handling](references/olares-settings-users.md) |
| `apps` | list/get, entrances, env, domain, policy, auth-level, suspend/resume | [post-install app configuration](references/olares-settings-apps.md) |
| `vpn` | devices, hidden `routes enable/disable`, SSH, subroutes (hidden enable/disable), ACL, public-domain-policy | [VPN and ACL decisions](references/olares-settings-vpn.md) |
| `integration` | account list/get/add/delete, `cookie import/list/rm/validate` | [integration accounts](references/olares-settings-integration.md); [cookie store](references/olares-settings-cookies.md) |
| `backup` | plans, snapshots, password | [backup decisions](references/olares-settings-backup.md) |
| `appearance` | `get`, `language set` | `settings appearance --help` |
| `network` | reverse-proxy, FRP, hosts-file, overlay gateway/app | `settings network --help` |
| `gpu` / `compute` | legacy GPU list / accelerator list, unbind, set-type | `settings gpu --help`; `settings compute --help` |
| `video` | `config get` | `settings video --help` |
| `search` | `status`, `rebuild`, `dirs list/add/rm` | [`olares-search`](../olares-search/SKILL.md) for index coverage |
| `restore` | `plans list` | `settings restore --help` |
| `advanced` | status, registries, images, system/user env | `settings advanced --help` |

## Role caching + admin/normal floor

- Roles are `normal`, `admin`, and `owner`. The active profile caches its role; the server remains authoritative when no cache exists or the role changed.
- Admin areas include user administration, accelerator administration, advanced machine state, and privileged VPN/network reads. Overlay gateway master enable/disable and reverse-proxy set are owner-only.
- `profile whoami`, `settings users me`, and `settings me whoami` are equivalent identity views. Do not redirect between them as if one were more authoritative.

## Version and asynchronous decisions

- Olares 1.12.5 uses legacy `gpu list`; 1.12.6+ uses `compute`.
- Overlay gateway and per-app overlay operations require 1.12.6+.
- Overlay writes are asynchronous. `--watch` observes the gateway/app state settling; per-app overlay changes may restart a running app through Market.
- Search indexing and rebuild are asynchronous. A successful request does not mean newly indexed content is immediately searchable.

## Security rules

- Never expose access tokens, SSO session fields, initial passwords, integration credentials, backup passwords, or VPN secrets in chat or command history.
- Read secrets from stdin or environment variables supported by the selected verb.
- Treat the requested configuration change as task-scope authorisation; do not ask again for each idempotent write in that scope.
- Confirm destructive or wider consequences separately: user deletion, ACL replacement, accelerator unbind/type changes, app restart/stop, index rebuild, and gateway-wide overlay changes.
- Treat app lifecycle requests as [`olares-market`](../olares-market/SKILL.md), not settings configuration.
- A download or collection that fails for a missing login is a cookie import here (`settings integration cookie`), not an [`olares-knowledge`](../olares-knowledge/SKILL.md) problem — see [cookie store](references/olares-settings-cookies.md).
- Stop when the requested verb is not registered, the role is insufficient, the target user/app/device is ambiguous, or the action requires a JWS/device interaction the CLI cannot perform.
