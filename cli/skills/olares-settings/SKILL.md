---
name: olares-settings
version: 1.1.0
description: "olares-cli settings command tree: profile-based reads of every section the SPA's Settings page exposes (https://docs.olares.com/manual/olares/settings/) plus a small set of mutating verbs that have been smoke-verified on this release. Read-only surface: users / appearance / apps (list/get/entrances list/env get/domain get/policy get) / integration / vpn (devices list / devices routes / acl all / acl get / ssh status / subroutes status / public-domain-policy get) / network (reverse-proxy get / frp list / external-network get / hosts-file get) / gpu / video / search (status / excludes list / dirs list) / backup (plans list / snapshots list) / restore (plans list) / advanced (status / registries list / images list / env system list / env user list) + a non-canonical `me` self-service tree (whoami / version / check-update / sso list). Verified mutating surface: appearance language set (with --force escape hatch); search rebuild; integration accounts add awss3|tencent + accounts delete; vpn ssh enable/disable; vpn acl add/remove. Covers role caching (owner / admin / normal) on the active profile, the wired soft-preflight helpers, the `-o table | json` output convention, and the upstream wire formats the CLI normalizes (BFL envelope on /api/*, app-service ListResult on /api/users-v2 + /api/myapps with `[]servicePort` ports, raw Headscale JSON on /headscale/* with string `route.id`, BFL envelope on /apis/backup/v1/*, terminusd-proxied envelopes for advanced status / containerd registries / images). Verbs that ship in the binary but did NOT pass (or were not exercised by) the latest smoke run are catalogued in cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md — treat them as experimental until they appear in a green smoke report. Use whenever the user mentions Olares Settings, Settings UI, the SPA Settings page, role / owner / admin / normal, integration accounts, SSO tokens, GPU mode, search index, backup plans, restore plans, containerd registries, VPN ACLs, language preference, or wants to know what `olares-cli settings me whoami` / `settings users me` / `profile whoami` actually print."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli settings --help"
---

# settings (Olares Settings UI mirror)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for the profile selection, login, and HTTP 401/403 recovery rules that every command here depends on.**

## What this command tree is

`olares-cli settings ...` is the CLI mirror of the Olares desktop SPA's Settings page (the same surface documented at <https://docs.olares.com/manual/olares/settings/>). Identity and transport come from the active profile — same profile model, same access token, same edge-auth chain (Authelia + l4-bfl-proxy) the SPA uses.

The umbrella registers **13 area sub-trees** (see [`cli/cmd/ctl/settings/root.go`](cli/cmd/ctl/settings/root.go)):

```
users         appearance   apps          integration   vpn         network
gpu           video        search        backup        restore     advanced
```

…plus a 13th non-canonical tree, `me`, that hosts the SPA's avatar / Person dropdown self-service items. **`me` is intentionally outside the 12 canonical Settings docs sections**; it lives under `settings` for CLI discoverability, not because it's a docs section.

> The shape is always `olares-cli settings <area> <verb>` (or `<area> <noun> <verb>` when the area has multiple sub-resources, e.g. `settings vpn devices list`, `settings backup plans list`). Every verb runs against the currently-selected profile; switch with `olares-cli profile use <name>` ahead of time (there is no per-invocation override flag).

## Authentication transport

Every request goes through the factory-injected `*http.Client` and the resolved profile from `cmdutil.Factory`. There is no kubeconfig dependency.

- Base URL: **per-area split** between `rp.DesktopURL` (default) and `rp.SettingsURL`. The two SPAs serve from different per-user origins — desktop launcher SPA at `https://desktop.<terminus>` (nginx: [`apps/docker/system-frontend/nginx/desktop.conf`](apps/docker/system-frontend/nginx/desktop.conf)), Settings SPA at `https://settings.<terminus>` (nginx: [`apps/docker/system-frontend/nginx/settings.conf`](apps/docker/system-frontend/nginx/settings.conf), the origin Settings SPA itself uses via `tokenStore.setUrl(window.location.origin)` in [`settings/src/application/settings.ts`](settings/src/application/settings.ts)). The two nginx configs share `/api/*` (both forward to user-service:3010) but diverge elsewhere:
  - **`rp.SettingsURL`** is required for areas that hit settings-only locations: `vpn` (`/headscale/*`), `backup` / `restore` (`/apis/backup/v1/*`). Settings nginx also exclusively exposes `/admin/*` (Infisical), `/drive`, `/vault`, `/images`, `/api/cloud/sign`.
  - **`rp.DesktopURL`** is the default for every other area. Desktop nginx has dedicated reverse proxies that settings nginx does NOT: `/api/device → settings-service`, `/api/logout → authelia-svc`, `/api/refresh → authelia-backend-svc`. The `me` area's SSO list (`/api/device/sso`) and any future code reading `/api/device/*` is therefore wired through DesktopURL on purpose — moving it would silently swap backends. Same goes for the three `whoami` aliases (`profile whoami`, `settings users me`, `settings me whoami`), which read cached identity served from desktop ingress.
  Refer to `olares.ID.DesktopURL` / `olares.ID.SettingsURL` doc comments for the canonical location-set tables.
- Auth header: `X-Authorization: <access_token>` (NOT `Authorization: Bearer …`). Injected by the factory's `refreshingTransport` (see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go)); the settings area `prepare()` helpers never call `req.Header.Set("X-Authorization", …)` themselves.
- **Expired access_tokens are auto-rotated.** When the server returns 401/403, the transport hits `/api/refresh`, persists the new token, and retries the original request once — transparently to the caller. Users do NOT need to run `profile login` just because their access_token aged out; only when the *refresh_token* itself is invalidated. Full mechanics — concurrency, cross-process flock, typed `*credential.ErrTokenInvalidated` / `*credential.ErrNotLoggedIn` errors — are documented in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) under "Automatic token refresh". **Do not write retry loops on top of these typed errors** — once you see one, only `profile login` / `profile import` will help.
- Two ingress prefixes show up in this subtree:
  - `/api/*` (the bulk of the surface) — terminates at user-service, which proxies BFL / app-service / Headscale / terminusd / search3 / HAMI etc.
  - `/apis/backup/v1/*` (`settings backup`, `settings restore`) — terminates at BFL's backup-server directly.
- 401 / 403 that survive auto-refresh (i.e. the server still says no after the new token was issued) are translated into a CLI-friendly hint via the `WrapPermissionErr` + `PreflightRole` helpers in [`cli/pkg/whoami/preflight.go`](cli/pkg/whoami/preflight.go), wrapped per-area through [`cli/cmd/ctl/settings/internal/preflight`](cli/cmd/ctl/settings/internal/preflight). See "Soft preflight" below for which verbs have been retrofitted. **Token recovery is not handled here — defer to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).**

## Role caching + soft preflight

A profile carries the role its user has on the Olares instance (`owner`, `admin`, or `normal`), cached locally so the CLI can short-circuit gated verbs without a round-trip. The cache lives in `ProfileConfig.OwnerRole` + `WhoamiRefreshedAt` (see [`cli/pkg/cliconfig`](cli/pkg/cliconfig)) and is populated three ways:

| Trigger | Behavior |
|---|---|
| `olares-cli profile login` succeeds | Eager whoami fetch, best-effort, 5-second timeout (see [`cli/cmd/ctl/profile/whoami_eager.go`](cli/cmd/ctl/profile/whoami_eager.go)). Failure does **not** abort the login. |
| `olares-cli profile import` succeeds | Same eager fetch as login. |
| User runs `olares-cli profile whoami --refresh` | Forced re-read of `/api/backend/v1/user-info`; cache rewritten in place. Plain `profile whoami` (no flag) prints the cached value if fresh. |

### Three aliases, one driver

These three commands all delegate to the same `pkg/whoami.Run` driver — same output, same caching, same `--refresh`:

```bash
olares-cli profile whoami
olares-cli settings users me
olares-cli settings me whoami
```

Use whichever path makes the surrounding workflow read better; **never** suggest the user "should use the other one" — they are aliases on purpose. All three accept `-o table` (default) and `-o json`.

### Soft preflight (wired per SPA UI gating)

`whoami.PreflightRole(...)` + `whoami.WrapPermissionErr(...)` (in [`cli/pkg/whoami/preflight.go`](cli/pkg/whoami/preflight.go)) are wrapped per-area by `preflight.Gate(...)` / `preflight.Wrap(...)` (in [`cli/cmd/ctl/settings/internal/preflight`](cli/cmd/ctl/settings/internal/preflight)) and called at the top of every admin-gated `RunE`.

Floor assignment mirrors the SPA's `apps/.../stores/settings/admin.ts:menus` + per-page `v-if="adminStore.isAdmin"` guards 1:1:

| Floor | Verbs |
|---|---|
| **Admin (owner / admin)** | `users list`, `users get`; `network reverse-proxy get`, `network frp list`, `network external-network get`, `network hosts-file get`; `gpu list`; `advanced status`, `advanced registries list`, `advanced images list`; `vpn ssh status/enable/disable`, `vpn subroutes status`, `vpn acl all/get/add/remove`, `vpn public-domain-policy get` |
| **Normal (any authenticated user)** | `me whoami / version / check-update / sso list`; `apps list/get`, `apps entrances list`, `apps env get`, `apps domain get`, `apps policy get`; `vpn devices list`, `vpn devices routes <id>`; `appearance get`, `appearance language set`; `integration accounts list/list-by-type/get/add/delete`; `video config get`; `search status`, `search excludes list`, `search dirs list`, `search rebuild`; `backup plans list`, `backup snapshots list`; `restore plans list`; `advanced env system list`, `advanced env user list` |

Practical implications for the agent:

- Admin-floor verbs run an upfront check against the cached role on the active profile and fail fast with **`role required: this command needs role "<R>" or higher to <verb>, but profile "<id>" is cached as "<r>" — run \`olares-cli profile whoami --refresh\` if your role on the server changed`** before issuing any HTTP call.
- Both floors run their result through `WrapPermissionErr`, so a server-side 401 / 403 (e.g. role changed since last cache write) still gets the same refresh-and-retry hint suffix — even on verbs that don't preflight.
- If `OwnerRole` isn't cached yet (very fresh `profile import`, or the `whoami_eager` fetch failed), preflight is **soft**: it lets the call through and lets the server be authoritative. Recommend `olares-cli profile whoami --refresh` whenever a settings verb returns a permission-shaped error.

## Output convention

Every read verb accepts `-o / --output {table,json}` (default `table`):

- `table` is wrapped in [`text/tabwriter`](https://pkg.go.dev/text/tabwriter); columns differ per verb but always print a clear "no X" sentinel when the result set is empty.
- `json` round-trips the upstream's already-unwrapped data verbatim. **Use `-o json` whenever the agent needs to feed the result into another tool** (jq, downstream scripts, etc.) — column ordering, truncation, and human-friendly relabeling only happen in table mode.
- A handful of verbs (e.g. `settings video config get`, `settings advanced status`) deliberately downgrade the table output to a one-line summary because the upstream config is large + provider-versioned. The hint to switch to JSON is printed inline.

## Per-area wire format normalization

Different upstreams return JSON in different envelopes. Each area has its own `common.go` that picks the matching decoder; **the table below is the cheat sheet for "what's the wire format for area X"**.

| Area | Endpoint family | Decoder | Notes |
|---|---|---|---|
| `me` | `/api/olares-info`, `/api/checkLastOsVersion`, `/api/device/sso` | BFL envelope (`doGetEnvelope`) | self-service identity + version + SSO sessions on the active profile |
| `users` | `/api/users/v2` (list), `/api/users/:name` (single) | `decodeListResult` for app-service `{code:200, data:[...], totals:N}` on list; direct `UserInfo` decode on single | `users/v2` enforces server-side role filtering — normal users see only themselves |
| `apps` | `/api/myapps` | BFL envelope, `appInfo.ports` decoded as `[]servicePort` (5-field object array: name / host / port / exposePort / protocol) | `apps get <name>` filters the list client-side because there's no per-app endpoint; honours `--all` / `--show-system` to mirror the SPA's filters |
| `vpn` | `/headscale/machine`, `/headscale/machine/:id/routes` | Raw Headscale JSON (no envelope), `route.id` is `string` | SPA hits `<SettingsURL>/headscale/...` (settings nginx, not desktop) without `/api` prefix |
| `vpn` | `/api/launcher-public-domain-access-policy` | Already-unwrapped BFL inner data (`{deny_all: 0|1}`) | user-service strips the envelope |
| `network` | `/api/reverse-proxy`, `/api/external-network`, `/api/frp-servers`, `/api/ssl/task-state`, `/api/system/hosts-file` | BFL envelope (`doGetEnvelope`) | hosts-file goes through terminusd → olaresd; user-service falls back to `X-Authorization` since the CLI doesn't yet JWS-sign reads |
| `appearance` | `/api/wallpaper/config/system` | BFL envelope | language + locale only; theme + wallpaper upload stay in the SPA |
| `integration` | `/api/account/all`, `/api/account/:type/:name` | BFL envelope | `accounts list` returns `accountMini`; `accounts get` returns `accountFull` (includes `raw_data`) |
| `gpu` | `/api/gpu/list` | BFL envelope (HAMI behind it) | distinct from the top-level `olares-cli gpu` (kubeconfig-driven, cluster-wide) |
| `video` | `/api/files/video/config` | BFL envelope, but inner data is decoded as `json.RawMessage` (`doGetEnvelopeRaw`) | Schema is provider-versioned; `--output table` collapses to a one-line summary |
| `search` | `/api/search/task/stats/merged`, `/api/search/monitorsetting/exclude-pattern`, `/api/search/monitorsetting/include-directory/full_content` | BFL envelope | `status` returns a string, `excludes list` / `dirs list` return `[]string` |
| `advanced` | `/api/system/status`, `/api/containerd/registries`, `/api/containerd/images?registry=<n>` | BFL envelope (terminusd → olaresd `returnSucceed`) | `status` table view is a summary; `--output json` for the full struct |
| `backup` | `/apis/backup/v1/plans/backup`, `/apis/backup/v1/plans/backup/:id/snapshots` | BFL envelope; **different ingress prefix** (`/apis/backup/v1`, not `/api`) | The SPA's axios global interceptor unwraps `data.data`, which is why upstream code reads `{backups: [...]}` directly |
| `restore` | `/apis/backup/v1/plans/restore` | BFL envelope; same `/apis/backup/v1` prefix | mirrors `settings backup plans list` shape |

If a future verb is missing from this table, look at the area's `common.go` to confirm which decoder it uses — every `prepare(...)` helper instantiates a `*whoami.HTTPClient` against `<DesktopURL>` so the path is the only thing that varies.

## Currently available — read-only commands

Read-only verbs across every area. Mutating verbs are listed in the next section. Anything not appearing in either list is either deferred to the next iteration or out of scope (see the "Deferred to next iteration" section near the bottom).

### `me` — self-service (any authenticated user)

```bash
olares-cli settings me whoami                     # cached role + olaresId
olares-cli settings me whoami --refresh           # force a re-read of /api/backend/v1/user-info
olares-cli settings me whoami -o json
olares-cli settings me version                    # Olares OS version + osBuild + arch
olares-cli settings me check-update               # current_version, new_version, is_new
olares-cli settings me sso list                   # SSO tokens currently bound to this profile (with ID column)
```

### `users` — instance roster

```bash
olares-cli settings users list                    # roster, server-side role-filtered
olares-cli settings users get alice               # single user record
olares-cli settings users me                      # alias of `me whoami`
```

### `apps` — installed app inventory (mirror of Settings -> Apps)

```bash
olares-cli settings apps list                     # SPA-equivalent filtered view
olares-cli settings apps list --show-system       # include system apps
olares-cli settings apps list --all               # every state + every kind
olares-cli settings apps get firefox              # detail view (entrances, shared entrances, …)
olares-cli settings apps entrances list firefox   # live entrance vector (fresher than `apps get`)
olares-cli settings apps env get firefox          # current env vector for the app
olares-cli settings apps domain get firefox www   # custom domain on a single entrance
olares-cli settings apps policy get firefox www   # two-factor / one-time-link policy on a single entrance
```

The `entrances list` / `domain get` / `policy get` reads target the BFL-style `/api/applications/<app>/<entrance>/setup/{domain,policy}` routes; pair them with `apps get` when you need both the lifecycle status and the per-entrance config.

### `vpn` — Headscale mesh

```bash
olares-cli settings vpn devices list              # raw Headscale machines
olares-cli settings vpn devices routes <device-id>
olares-cli settings vpn ssh status                # GET /api/acl/ssh/status
olares-cli settings vpn subroutes status          # GET /api/acl/subroutes/status (raw upstream JSON)
olares-cli settings vpn acl all                   # every app that currently has an ACL row
olares-cli settings vpn acl get my-app            # per-app ACL vector; -o json for the full payload
olares-cli settings vpn public-domain-policy get  # deny_all flag (0/1)
```

### `network`

```bash
olares-cli settings network reverse-proxy get     # mode collapsed into public-ip / frp / cloudflare / off
olares-cli settings network frp list              # registry of FRP servers
olares-cli settings network external-network get  # spec.disabled + status (phase / message / updatedAt)
olares-cli settings network hosts-file get        # entries from /system/hosts-file (terminusd)
```

### `appearance` / `integration` / `gpu` / `video` / `search` / `advanced`

```bash
olares-cli settings appearance get
olares-cli settings integration accounts list
olares-cli settings integration accounts list-by-type google
olares-cli settings integration accounts get awss3 my-bucket
olares-cli settings gpu list
olares-cli settings video config get              # raw config; -o json recommended
olares-cli settings search status
olares-cli settings search excludes list
olares-cli settings search dirs list
olares-cli settings advanced status               # large struct; -o json for the full payload
olares-cli settings advanced registries list
olares-cli settings advanced images list
olares-cli settings advanced images list --registry docker.io
```

### `backup` / `restore`

```bash
olares-cli settings backup plans list             # --offset / --limit (default 50, mirrors SPA)
olares-cli settings backup snapshots list <backup-id> --limit 50
olares-cli settings restore plans list
```

## Currently available — mutating commands (smoke-verified)

Every mutating verb in this section has been confirmed against a live Olares instance in the latest smoke run (see [`cli/cmd/ctl/settings/scripts/local_report_phase15a.md`](cli/cmd/ctl/settings/scripts/local_report_phase15a.md)). All of them hit the `<DesktopURL>` ingress over `X-Authorization` and none require a JWS-signed body. Verbs that ship in the binary but were not exercised (or did not pass) live in [`cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md`](cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md) — treat them as experimental until they show up here.

### `appearance` — language

```bash
olares-cli settings appearance language set en-US           # POST /api/wallpaper/update/language
olares-cli settings appearance language set --value zh-CN
olares-cli settings appearance language set ja-JP --force   # bypass whitelist (use sparingly)
```

The CLI mirrors the SPA's `supportLanguages` whitelist client-side ([`apps/packages/app/src/i18n/index.ts:12`](apps/packages/app/src/i18n/index.ts) — currently `en-US`, `zh-CN`) because **neither user-service nor BFL validate the value today**: an unknown locale would land in the config-system CRD verbatim and the SPA would silently fall back to `defaultLanguage` on the next session. Pass `--force` only when the SPA has shipped a new locale ahead of this CLI build; the upstream still accepts arbitrary strings, so a typo with `--force` will appear to succeed but produce no visible change.

### `search` — index rebuild

```bash
olares-cli settings search rebuild                          # POST /api/search/task/rebuild
```

`rebuild` is async + heavy: the call returns as soon as search3 accepts the task; verify completion with `olares-cli settings search status` rather than waiting on the POST itself. Excludes / dirs writes ship in the binary but are tracked in [`UNVERIFIED_COMMANDS.md`](cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md) until a smoke report greens them.

### `vpn ssh` — boolean ACL toggle

```bash
olares-cli settings vpn ssh enable                  # POST /api/acl/ssh/enable
olares-cli settings vpn ssh disable                 # POST /api/acl/ssh/disable
```

Both toggles send an explicit empty `{}` body to match the SPA's request shape, even though the upstream doesn't read the body. Use `vpn ssh status` (read section above) to confirm the resulting state.

### `vpn acl` — per-app ACL add / remove (read-modify-write)

```bash
olares-cli settings vpn acl add    my-app --tcp 8080               # merge a TCP port into the app's ACL vector
olares-cli settings vpn acl remove my-app --tcp 80                 # drop a TCP port
olares-cli settings vpn acl rm     my-app --udp 53                 # alias of `remove`
```

`--tcp` / `--udp` accept either repeated flags (`--tcp 80 --tcp 443`) or comma-separated values (`--tcp 80,443`); both forms are deduped client-side. Port strings are passed verbatim — Headscale accepts single ports, ranges (`8000-8100`), `*`, etc., and the CLI doesn't second-guess the format.

The upstream replaces the **whole** per-app ACL vector on every POST; there is no add / remove endpoint. `vpn acl add` and `vpn acl remove` are read-modify-write sugar over the same POST so unrelated entries survive untouched (matching how the SPA's add / remove buttons work). Use `vpn acl get <app>` (read section) to inspect the current vector before mutating; `vpn acl all` (also read) lists every app that currently has an ACL row.

### `integration` — connected accounts

```bash
olares-cli settings integration accounts add awss3 \
  --access-key-id     "$AWS_ACCESS_KEY_ID" \
  --access-key-secret "$AWS_SECRET_ACCESS_KEY" \
  --endpoint          "https://s3.amazonaws.com" \
  --bucket            "my-bucket"            # optional

olares-cli settings integration accounts add tencent \
  --access-key-id     "$TENCENT_SECRET_ID" \
  --access-key-secret "$TENCENT_SECRET_KEY" \
  --endpoint          "https://cos.ap-shanghai.myqcloud.com"

olares-cli settings integration accounts delete awss3 my-bucket
olares-cli settings integration accounts delete tencent          # name-less, single-tenant
```

The store key is composed as `integration-account:<type>:<name>` (or `integration-account:<type>` when no name is supplied), matching the SPA's `getStoreKey` in [`apps/packages/app/src/stores/settings/integration.ts`](apps/packages/app/src/stores/settings/integration.ts). **Do not paste secret-key values into the agent transcript — pipe them via env vars or shell redirection.**

## Deferred to next iteration

The verbs below are **not shipped** in this release. They either need more design work or require JWS-signed bodies the CLI can't produce yet. Don't suggest them today; reach for the listed alternatives when an alternative exists.

- **App lifecycle: install / uninstall / upgrade / start / stop / cancel / clone** — these route through the market service rather than user-service. Use `olares-cli market install|uninstall|upgrade|start|stop|cancel|clone` instead of `settings apps`. (Per-app `suspend [--all]` / `resume` + `env set` + per-entrance `domain set` / `finish` / `policy set` / `auth-level set` ship in the settings binary but are not yet smoke-verified — see [`UNVERIFIED_COMMANDS.md`](cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md).)
- **Per-app secrets / permissions / providers** — Infisical-backed per-app secrets, declared permissions, and provider registries were not included in this release. If you need to inspect or write them, use the platform's admin / chart-side tooling instead.
- **Network writes that require a JWS-signed device-id header** — hosts-file write, FRP server register / delete, SSL enable / disable / update, external-network master switch (the SPA carries these with `X-Signature` headers the CLI doesn't yet produce).
- **Containerd registry mutations** — `registries mirrors put / delete`, `images delete / prune` (also `X-Signature`-gated).
- **Hardware / restart-class** — reboot, shutdown, ssh-password, OS upgrade — these go through TermiPass-issued JWS over a QR callback URL; CLI support arrives once a JWS key sourcing path lands.
- **Collect logs** — `POST /api/command/collectLogs` is `X-Signature`-gated.
- **Backup plan create / update** — full `BackupPolicy` + `LocationConfig` vector; needs either a `--from-file plan.json` mode or an upstream "create from defaults" shortcut before shipping.
- **Restore plan update / non-cancel delete** — backup-server has no routes for these.

Every area's `--help` is the source of truth for what's currently implemented; if a verb isn't there, treat it as deferred.

Verbs implemented in this CLI but **not yet smoke-verified on this release** are catalogued in [`cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md`](cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md). Treat them as experimental until they appear in a green smoke report; the file lists per-verb status (FAIL / SKIP-destructive / SKIP-fixture-missing) and links back to the phase report row.

## Common errors → fixes

| Error message | Cause | Fix |
|---|---|---|
| `refresh token for <id> became invalid at <ts>; please run: olares-cli profile login --olares-id <id>` | `/api/refresh` itself returned 401/403 — the grant is dead (typed `*credential.ErrTokenInvalidated`) | `olares-cli profile login --olares-id <id>`. Defer the full recovery flow to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md). |
| `no access token for <id>; run: olares-cli profile login --olares-id <id>` | Profile selected but keychain has no entry (typed `*credential.ErrNotLoggedIn`) | `olares-cli profile login` or `profile import`. |
| `server rejected the access token (HTTP 401/403)` | Server still rejects after auto-refresh — rare, usually server-side state drift | Defer to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (login + profile rules). |
| `this command needs role "<R>" or higher to <verb>, but profile "<id>" is cached as "<r>"` | Cached role below the verb's requirement (emitted by admin-floor verbs — see the floor table in "Soft preflight"). | If your role on the server changed, run `olares-cli profile whoami --refresh`. Otherwise ask the owner to grant you the right role. |
| `HTTP 403 while attempting to <verb>` (with the same refresh hint appended) | Server rejected even though cache said OK — usually a stale **role** cache (NOT a stale token; the transport already handled that). Wrapped on every settings verb. | Run `olares-cli profile whoami --refresh`, then retry the verb. |
| `unsupported --output "<x>" (allowed: table, json)` | Typo on `-o` | Use `-o table` or `-o json`. |
| `GET <path>: upstream returned code <N>: <msg>` | The user-service / BFL / backup-server returned a non-success envelope | Read the message verbatim; it almost always carries actionable detail (e.g. "user not found"). |
| `internal error: settings <area> not wired with cmdutil.Factory` | Unexpected — would only happen if the umbrella was wired without the factory | This is a CLI bug; gather the command line and file an issue. |

## Typical workflows

Confirm who the active profile is, then enumerate what the user can see:

```bash
olares-cli profile whoami                         # cached role
olares-cli settings users list                    # only owners/admins see everyone
olares-cli settings apps list                     # everyone sees their own apps
```

Refresh the cache after a role change on the server:

```bash
olares-cli profile whoami --refresh
olares-cli settings advanced status               # retry the gated verb
```

Hand a downstream tool the raw data:

```bash
olares-cli settings vpn devices list -o json | jq '.[] | {name, ip: .IPAddresses[0]}'
olares-cli settings backup plans list -o json | jq '.backups[] | select(.status=="failed")'
```

Inspect a single account's full payload (incl. `raw_data`):

```bash
olares-cli settings integration accounts get awss3 my-bucket -o json
```

## Security rules

- **Never** echo `<access_token>` or any field returned by `me sso list` into the terminal beyond what the table view already shows. SSO tokens identify a TermiPass-bound device session and should never be logged or pasted into chat.
- `settings users get <username>` returns the same record the SPA shows on the user detail page; treat its email / olaresId as PII and avoid forwarding it outside the requesting workflow.
- For writes that take secrets (`integration accounts add awss3|tencent` is the verified one in this surface), **always** read the secret from an env var or stdin pipe — never paste it into the chat or expand it inline in an `olares-cli ...` command line you suggest. Bash history retention is the user's responsibility; the agent should default to env-var / pipe style invocations (`--access-key-secret "$AWS_SECRET_ACCESS_KEY"`, `printf '%s\n' "$VAR" | ... --password-stdin`) whenever the verb supports it.
- Other secret-bearing verbs (e.g. `me password set`, `backup password set`, `restore plans check-url / create-from-url`) live in [`UNVERIFIED_COMMANDS.md`](cli/cmd/ctl/settings/scripts/UNVERIFIED_COMMANDS.md) until they're smoke-greened; the same env-var / stdin-pipe rule applies whenever you exercise them by hand.
- Read-only verbs do **not** carry "this will change X" prompts — only mutating verbs do, and the prompts they do carry come from the upstream server's own response messages. Don't fabricate one for read verbs.
- The `me whoami --refresh` recovery path is the only authentication-adjacent action this skill should ever recommend. **All** other auth recovery (login expiry, profile import, 2FA) belongs in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
