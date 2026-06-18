---
name: olares-settings
version: 4.2.1
description: "Olares Settings via olares-cli settings — mirror of Settings SPA: users, apps, VPN, backup, integration, GPU, search, me/whoami. Use for Olares Settings, role, VPN ACL, backup, integration accounts, language."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# settings (Olares Settings UI mirror)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli settings --help`.**

> **Source of truth for flags is always `olares-cli settings <area> <verb> --help`.** This file only carries what `--help` cannot give: routing, the 13-section index, the role-caching / admin-vs-normal floor, the wire-format cheat sheet, and the common-errors table.

> **Platform model:** the Olares version/semver scheme behind `settings me version` is defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md#olares-version--semver-model).

## When to use

- Olares Settings UI (https://docs.olares.com/manual/olares/settings/), olares-cli settings, role (owner / admin / normal), who am I on this Olares instance
- Areas: users, appearance, apps (entrances / env / domain / policy), integration (awss3 / tencent), VPN (devices / ACL / SSH), network, GPU, video, search, backup / restore, advanced (containerd registries, env)
- Mutating: language set, search rebuild, integration add/delete, VPN SSH / ACL, users create/delete

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** `settings` covers configuration that the Olares Settings SPA exposes — **post-install per-app config**, mesh / VPN, backup, accounts, system appearance. Lifecycle and runtime live in sibling skills.

## 13 area sub-trees

The umbrella registers the 12 canonical Settings docs sections, plus a 13th non-canonical `me` self-service tree:

```
users         appearance   apps          integration   vpn         network
gpu           video        search        backup        restore     advanced
+ me (self-service: whoami / version / check-update / sso list)
```

Shape is always `olares-cli settings <area> <verb>` (or `<area> <noun> <verb>` when the area has multiple sub-resources, e.g. `vpn devices list`, `backup plans list`).

### Per-area `--help` and references

For every area, **start with `olares-cli settings <area> --help`**. References below cover the non-trivial sub-trees:

| Area | `--help` first, then... |
|---|---|
| `me` | `olares-cli settings me --help` (whoami / version / check-update / sso list) |
| `users` | [references/olares-settings-users.md](references/olares-settings-users.md) |
| `apps` | [references/olares-settings-apps.md](references/olares-settings-apps.md) (entrances / domain / policy / auth-level pipeline) |
| `vpn` | [references/olares-settings-vpn.md](references/olares-settings-vpn.md) (ACL deltas, SSH / subroutes / public-domain-policy) |
| `integration` | [references/olares-settings-integration.md](references/olares-settings-integration.md) (`accounts add awss3|tencent`) |
| `backup` | [references/olares-settings-backup.md](references/olares-settings-backup.md) (plans / snapshots / password) |
| `appearance` | `olares-cli settings appearance --help` (`get`, `language set`) |
| `network` | `olares-cli settings network --help` (read-only reverse-proxy / frp / hosts-file; writes blocked by JWS gap — see below) |
| `gpu` | `olares-cli settings gpu --help` (read-only `list`) |
| `video` | `olares-cli settings video --help` (read-only `config get`) |
| `search` | `olares-cli settings search --help` (`status`, `rebuild`, `dirs list/add/rm`). `dirs` = the **full-content** index directories (filenames are indexed broadly by default; full-text defaults to Drive `/Documents/` only — add more with `dirs add`). `status` shows the index `Status` plus a full-text-extraction `Failures` count (`-o json` for per-file detail). Exclude-pattern view/edit is SPA-only today. Index coverage model lives in [`olares-search`](../olares-search/SKILL.md) |
| `restore` | `olares-cli settings restore --help` (read-only `plans list`) |
| `advanced` | `olares-cli settings advanced --help` (read-only `status`, `registries list`, `images list`, `env (system|user) list`) |

## Role caching + admin/normal floor

A profile carries the role its user has on the Olares instance — `owner`, `admin`, or `normal` — cached locally so the CLI can short-circuit gated verbs without a round-trip. Populated on `profile login` / `profile import`, refreshed by `olares-cli profile whoami --refresh`.

### Three whoami aliases, one driver

```bash
olares-cli profile whoami
olares-cli settings users me
olares-cli settings me whoami
```

All three delegate to the same driver — same output, same caching, same `--refresh`. **Never** suggest the user "should use the other one" — they are aliases on purpose.

### Floor table (admin-gated vs normal)

| Floor | Verbs |
|---|---|
| **Admin (owner / admin)** | `users list / get / create / delete`; `network reverse-proxy get`, `network frp list`, `network hosts-file get`; `gpu list`; `advanced status / registries list / images list`; `vpn ssh status/enable/disable`, `vpn subroutes status`, `vpn acl all/get/add/remove`, `vpn public-domain-policy get` |
| **Normal (any authenticated user)** | `me whoami / version / check-update / sso list`; `apps list/get`, `apps entrances list`, `apps env get`, `apps domain get`, `apps policy get`; `vpn devices list / routes <id>`; `appearance get`, `appearance language set`; `integration accounts list / list-by-type / get / add / delete`; `video config get`; `search status / dirs list / rebuild`; `backup plans list`, `backup snapshots list`; `restore plans list`; `advanced env (system|user) list` |

### Soft preflight behavior

- Admin-floor verbs check the cached role and fail fast with `role required: this command needs role "<R>" or higher to <verb>, but profile "<id>" is cached as "<r>"` **before any HTTP call**.
- Server-side 401 / 403 (e.g. role changed since last cache write) still gets the same refresh-and-retry hint, even on verbs that don't preflight.
- If role isn't cached yet, preflight is **soft**: it lets the call through and lets the server be authoritative.
- **The standard refresh path is `olares-cli profile whoami --refresh`** — recommend it whenever a settings verb returns a permission-shaped error.

## Output convention

Every read verb accepts `-o / --output {table,json}` (default `table`):

- `table` is tabwriter-formatted; columns differ per verb but always print a clear "no X" sentinel when the result set is empty.
- `json` round-trips the upstream's already-unwrapped data verbatim. **Use `-o json` whenever the agent needs to feed the result into another tool** — column ordering, truncation, and human-friendly relabeling only happen in table mode.
- A handful of verbs (`video config get`, `advanced status`) downgrade table output to a one-line summary because the upstream config is large; the hint to switch to JSON is printed inline.

## Wire-format cheat sheet

Different upstream services return JSON in different envelopes. The CLI normalizes them per-area; this table is the cheat sheet for "what's the wire format for area X" when you need to fish out a non-tabled field via `-o json`:

| Area | Endpoint family | Wire envelope |
|---|---|---|
| `me`, `apps`, `network`, `appearance`, `integration`, `gpu`, `video`, `search`, `advanced` | `/api/*` (user-service / BFL / terminusd) | Unwrapped BFL `{data: ...}` envelope |
| `users` | `/api/users/v2`, `/api/users/:name`, `POST/DELETE /api/users/...` | List-result decoder; mutating returns axios-inner `{name}` |
| `vpn` devices / ACL | `<SettingsURL>/headscale/machine`, `/headscale/machine/:id/routes` | Raw Headscale JSON (NO envelope), `route.id` is a string |
| `vpn` public-domain-policy | `/api/launcher-public-domain-access-policy` | Already-unwrapped `{deny_all: 0|1}` |
| `backup`, `restore` | `<SettingsURL>/apis/backup/v1/*` | BFL envelope; **different ingress prefix** (`/apis/backup/v1`, not `/api`) |
| `video config get` | `/api/files/video/config` | BFL envelope, inner data as `json.RawMessage` (provider-versioned) |

## Currently-implemented mutating verbs

Verbs marked **VERIFIED** have been confirmed against a live Olares instance. Verbs flagged **UNVERIFIED** ship in the binary but are tracked as experimental; the CLI emits a one-line "experimental" stderr hint when they run.

| Area | Verb | Status |
|---|---|---|
| `appearance` | `language set <code>` | VERIFIED |
| `users` | `create` / `delete` (with `--watch`) | VERIFIED |
| `search` | `rebuild`, `dirs add / remove` | VERIFIED (rebuild + dirs writes) |
| `vpn ssh` | `enable` / `disable` | VERIFIED |
| `vpn acl` | `add` / `remove` | VERIFIED |
| `integration accounts` | `add awss3` / `add tencent` / `delete` | VERIFIED |
| `apps` | `suspend [--all]` / `resume`, `env set`, `domain set/finish`, `policy set`, `auth-level set` | UNVERIFIED |
| `backup` | `password set` | UNVERIFIED |

**Not yet implemented** (and the CLI deliberately does NOT register them):

- App lifecycle (`install` / `uninstall` / `upgrade` / `start` / `stop` / `cancel` / `clone`) → use [`olares-market`](../olares-market/SKILL.md) instead
- Per-app secrets / permissions / providers (Infisical-backed) → admin / chart-side tooling
- Network writes requiring a JWS-signed device-id header (`hosts-file set`, `frp set`, `ssl enable/disable/update`)
- Containerd registry mutations (`registries mirrors put/delete`, `images delete/prune`) — JWS-gated
- Hardware / restart-class (reboot, shutdown, ssh-password, OS upgrade) — JWS-gated via TermiPass QR callback
- Backup plan create / update (needs full `BackupPolicy` + `LocationConfig` vector design)
- Restore plan update / non-cancel delete — backup-server has no routes

**Don't suggest the deferred verbs today** — they will error with "command not found".

## Security rules

- **Never** echo `<access_token>` or any field returned by `me sso list` into the terminal beyond what the table view shows. SSO tokens identify a TermiPass-bound device session and should never be logged or pasted into chat.
- For writes that take secrets (`integration accounts add awss3|tencent`, future `backup password set`), **always read the secret from an env var or stdin pipe** — never paste it into chat or expand it inline in a suggested `olares-cli ...` command line.
- `users create` / `users delete` are destructive. `delete` needs the whole word `yes` unless `--yes`. `delete` refuses owner accounts (rejected before DELETE). `create` always prints the initial password once to stdout — treat transcripts accordingly.
- Read-only verbs do NOT carry "this will change X" prompts. **Don't fabricate one for read verbs.**
- The `profile whoami --refresh` recovery path is the only authentication-adjacent action this skill recommends. **All** other auth recovery belongs in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).

## Common errors → fixes

| Error | Cause | Fix |
|---|---|---|
| `this command needs role "<R>" or higher to <verb>, but profile "<id>" is cached as "<r>"` | Cached role below the verb's floor | If your role on the server changed: `olares-cli profile whoami --refresh`. Otherwise ask owner to grant the role |
| `HTTP 403 while attempting to <verb>` (with refresh hint) | Server rejected even though cache said OK — stale role cache | `olares-cli profile whoami --refresh`, retry the verb |
| `refresh token for <id> became invalid at <ts>; please run: olares-cli profile login --olares-id <id>` | The refresh_token itself is dead — see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) | `olares-cli profile login --olares-id <id>` |
| `no access token for <id>; run: olares-cli profile login --olares-id <id>` | Keychain has no entry for the active profile | `olares-cli profile login` or `profile import` |
| `unsupported --output "<x>" (allowed: table, json)` | Typo on `-o` | Use `-o table` or `-o json` |
| `GET <path>: upstream returned code <N>: <msg>` | user-service / BFL / backup-server returned non-success envelope | Read the message verbatim; it almost always carries actionable detail (e.g. "user not found") |

For the full auth-error matrix see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
