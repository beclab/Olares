---
name: olares-settings
version: 0.4.0
description: "olares-cli settings command tree: profile-based reads of every section the SPA's Settings page exposes (https://docs.olares.com/manual/olares/settings/) plus the Phase 2-6 mutating verbs that don't require JWS-signed bodies (me sso revoke, me password set with version-aware MD5+salt; appearance language set; search rebuild + excludes/dirs add+rm; integration accounts add awss3|tencent + accounts delete; apps suspend/resume + per-app env get/set + per-app secrets list/set/delete + per-app permissions/providers/entrances/domain/policy/auth-level reads + per-entrance domain set/finish + per-entrance policy set + per-entrance auth-level set; vpn devices rename/delete/tags + routes enable/disable + ssh enable/disable + subroutes enable/disable + per-app acl get/set/add/remove/clear + public-domain-policy set; network reverse-proxy set; advanced env system|user list/set; backup plans delete/pause/resume + snapshots run/cancel + password set; restore plans check-url + create-from-snapshot + create-from-url + cancel). Phase 1 surface (read-only): users / appearance / apps / integration / vpn / network / gpu / video / search / backup / restore / advanced + a non-canonical `me` self-service tree (whoami / version / check-update / login-history / sso list). Covers role caching (owner / admin / normal) on the active profile, soft preflight + `profile whoami --refresh` recovery, the `-o table | json` output convention, and the diverse upstream wire formats the CLI normalizes (BFL envelope on /api/*, app-service ListResult on /api/users-v2 + /api/myapps, raw Headscale JSON on /headscale/*, BFL envelope on /apis/backup/v1/*, terminusd-proxied envelopes for advanced status / containerd registries / images, /admin/secret/<app> for per-app secrets, BFL envelope on /api/applications/<app>/<entrance>/setup/* for per-entrance config). Use whenever the user mentions Olares Settings, Settings UI, the SPA Settings page, role / owner / admin / normal, integration accounts, login history, SSO tokens, GPU mode, search index, backup plans, restore plans, containerd registries, password change, app secrets, app env vars, app suspend/resume, app permissions, app entrances, app custom domain, app two-factor policy, app authorization level, VPN ACLs, reverse-proxy mode, restore from URL, or sees errors like 'this command needs role X to ...', 'HTTP 403 while attempting to ...', 'run olares-cli profile whoami --refresh', or wants to know what `olares-cli settings me whoami` / `settings users me` / `profile whoami` actually print."
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

> The shape is always `olares-cli settings <area> <verb>` (or `<area> <noun> <verb>` when the area has multiple sub-resources, e.g. `settings vpn devices list`, `settings backup plans list`). Every verb honours the global `--profile` flag inherited from the umbrella.

## Authentication transport

Every request goes through the factory-injected `*http.Client` and the resolved profile from `cmdutil.Factory`. There is no kubeconfig dependency.

- Base URL: `<rp.DesktopURL>` (e.g. `https://desktop.<terminus>`) — the same origin the SPA uses.
- Auth header: `X-Authorization: <access_token>` (NOT `Authorization: Bearer …`). Injected by the factory's `refreshingTransport` (see [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go)); the settings area `prepare()` helpers never call `req.Header.Set("X-Authorization", …)` themselves.
- **Expired access_tokens are auto-rotated.** When the server returns 401/403, the transport hits `/api/refresh`, persists the new token, and retries the original request once — transparently to the caller. Users do NOT need to run `profile login` just because their access_token aged out; only when the *refresh_token* itself is invalidated. Full mechanics — concurrency, cross-process flock, typed `*credential.ErrTokenInvalidated` / `*credential.ErrNotLoggedIn` errors — are documented in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) under "Automatic token refresh". **Do not write retry loops on top of these typed errors** — once you see one, only `profile login` / `profile import` will help.
- Two ingress prefixes show up in this subtree:
  - `/api/*` (the bulk of the surface) — terminates at user-service, which proxies BFL / app-service / Headscale / terminusd / search3 / HAMI etc.
  - `/apis/backup/v1/*` (`settings backup`, `settings restore`) — terminates at BFL's backup-server directly.
- 401 / 403 that survive auto-refresh (i.e. the server still says no after the new token was issued) are translated into a CLI-friendly hint via `WrapPermissionErr` + `PreflightRole` (see [`cli/cmd/ctl/settings/preflight.go`](cli/cmd/ctl/settings/preflight.go)) — that's the **stale role cache** path, distinct from the **stale access_token** path which the transport already handled. **Token recovery is not handled here — defer to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).**

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

### Soft preflight

Verbs that gate on role call `settings.PreflightRole(...)` at the top of their `RunE`. Behavior (see [`cli/cmd/ctl/settings/preflight.go`](cli/cmd/ctl/settings/preflight.go)):

- Cached role missing / unknown → no opinion, server decides.
- Cached role is **at or above** the required rank → pass through.
- Cached role is **below** the required rank → short-circuit with the `profile whoami --refresh` hint and don't issue the API call.

### 403/401 wrap

Every gated verb wraps its outbound error with `settings.WrapPermissionErr`, which detects HTTP 401 / 403 in the upstream message and appends the same "run `profile whoami --refresh` and retry" hint. This closes the loop for the **stale cache** case (cache said you're admin, server has since demoted you to normal).

> The user only ever needs to learn one trick: **`olares-cli profile whoami --refresh`**. Both the preflight branch and the post-flight branch funnel into that same recovery.

## Output convention

Every read verb accepts `-o / --output {table,json}` (default `table`):

- `table` is wrapped in [`text/tabwriter`](https://pkg.go.dev/text/tabwriter); columns differ per verb but always print a clear "no X" sentinel when the result set is empty.
- `json` round-trips the upstream's already-unwrapped data verbatim. **Use `-o json` whenever the agent needs to feed the result into another tool** (jq, downstream scripts, etc.) — column ordering, truncation, and human-friendly relabeling only happen in table mode.
- A handful of verbs (e.g. `settings video config get`, `settings advanced status`) deliberately downgrade the table output to a one-line summary because the upstream config is large + provider-versioned. The hint to switch to JSON is printed inline.

## Per-area wire format normalization

Different upstreams return JSON in different envelopes. Each area has its own `common.go` that picks the matching decoder; **the table below is the cheat sheet for "what's the wire format for area X"**.

| Area | Endpoint family | Decoder | Notes |
|---|---|---|---|
| `me` | `/api/olares-info`, `/api/checkLastOsVersion`, `/api/users/<u>/login-records`, `/api/device/sso` | BFL envelope (`doGetEnvelope`) for most; `/login-records` is unwrapped server-side and decoded directly | login-history derives `<u>` from the OlaresID local part |
| `users` | `/api/users/v2` (list), `/api/users/:name` (single) | `decodeListResult` for app-service `{code:200, data:[...], totals:N}` on list; direct `UserInfo` decode on single | `users/v2` enforces server-side role filtering — normal users see only themselves |
| `apps` | `/api/myapps` | BFL envelope | `apps get <name>` filters the list client-side because there's no per-app endpoint; honours `--all` / `--show-system` to mirror the SPA's filters |
| `vpn` | `/headscale/machine`, `/headscale/machine/:id/routes` | Raw Headscale JSON (no envelope) | Discovered URL: SPA hits `<DesktopURL>/headscale/...` directly without `/api` prefix |
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

## Phase 1 verb cheatsheet

Phase 1 covers the **read-only** surface for every area. The Phase 2 mutating verbs that have already landed (low-risk owner-and-self CRUD) are listed in the next section. Verbs scheduled for Phase 3-6 still print "this verb lands in phase N" in their `--help`; treat them as not yet implemented.

### `me` — self-service (any authenticated user)

```bash
olares-cli settings me whoami                     # cached role + olaresId
olares-cli settings me whoami --refresh           # force a re-read of /api/backend/v1/user-info
olares-cli settings me whoami -o json
olares-cli settings me version                    # Olares OS version + osBuild + arch
olares-cli settings me check-update               # current_version, new_version, is_new
olares-cli settings me login-history              # last login records (for the active profile only)
olares-cli settings me login-history --limit 50
olares-cli settings me sso list                   # SSO tokens (with `ID` column for `me sso revoke`)
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
```

### `vpn` — Headscale mesh

```bash
olares-cli settings vpn devices list              # raw Headscale machines
olares-cli settings vpn devices routes <device-id>
olares-cli settings vpn public-domain-policy get  # deny_all flag (0/1)
```

### `network`

```bash
olares-cli settings network reverse-proxy get     # mode collapsed into public-ip / frp / cloudflare / off
olares-cli settings network frp list              # registry of FRP servers
olares-cli settings network external-network get  # spec.disabled + status (phase / message / updatedAt)
olares-cli settings network ssl status            # task-state mapped to human label
olares-cli settings network hosts-file get        # entries from /system/hosts-file (terminusd)
```

### `appearance` / `integration` / `gpu` / `video` / `search` / `advanced`

```bash
olares-cli settings appearance get
olares-cli settings integration accounts list
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

## Phase 2 verb cheatsheet (low-risk owner-and-self CRUD)

Every Phase 2 verb hits the same `<DesktopURL>` ingress over `X-Authorization`; none of them require a JWS-signed body. The endpoints are picked deliberately to mirror the SPA's *self-service* / *owner-only* dialogs that don't carry secrets in their request bodies (the one exception being `me password set`, which hashes locally before sending — see below).

### `me` — self-service writes

```bash
# Revoke an SSO token. The ID comes from the new "ID" column on `me sso list`.
olares-cli settings me sso revoke <sso-id>

# Change the current user's password.
olares-cli settings me password set                         # interactive (hidden)
printf '%s\n%s\n' "$CURRENT" "$NEW" |
  olares-cli settings me password set --passwords-stdin     # automation
```

`me password set` hashes both passwords with the SPA's `saltedMD5` scheme (`md5(password+"@Olares2025")`) when the target OS reports `>= 1.12.0-0`, otherwise sends the raw password — implemented in [`cli/cmd/ctl/settings/me/passwordhash.go`](cli/cmd/ctl/settings/me/passwordhash.go) (mirrors [`apps/packages/app/src/utils/salted-md5.ts`](apps/packages/app/src/utils/salted-md5.ts) and [`apps/packages/app/src/utils/account.ts`](apps/packages/app/src/utils/account.ts) bit-for-bit, including the dash-prerelease quirk in `compareOlaresVersion`). Raw passwords never leave the machine. **After a successful password change, the existing access token may still work for a while; if a later CLI call returns 401, run `olares-cli profile login --olares-id <olares-id>`.**

### `appearance` — language

```bash
olares-cli settings appearance language set --value en-US   # POST /api/wallpaper/update/language
```

The list of supported codes is whatever the SPA's i18n bundle ships; the server is the source of truth and rejects unknown values with a clear message.

### `search` — index control

```bash
olares-cli settings search rebuild                          # POST /api/search/task/rebuild
olares-cli settings search excludes add "node_modules" "*.tmp"
olares-cli settings search excludes rm  "*.tmp"
olares-cli settings search dirs add /home/alice/Documents
olares-cli settings search dirs rm  /home/alice/Documents
```

`excludes add` / `dirs add` go through `PUT .../part`; the `rm` counterparts hit `DELETE .../part`. Both helpers dedupe and trim their inputs before sending so the upstream sees a clean list.

### `vpn` — Headscale device + policy writes (Phase 3c1)

```bash
olares-cli settings vpn devices rename <device-id> <new-name>
olares-cli settings vpn devices delete <device-id>            # prompts; pass --yes for automation
olares-cli settings vpn devices tags   set <device-id> --tag ops --tag laptop
olares-cli settings vpn devices tags   set <device-id>        # zero --tag flags clears the list

olares-cli settings vpn routes enable  <route-id>             # IDs come from `devices routes <id>`
olares-cli settings vpn routes disable <route-id>

olares-cli settings vpn public-domain-policy set --deny-all   # block non-whitelisted entrances
olares-cli settings vpn public-domain-policy set --allow-all  # default Olares behavior
```

`devices delete` is destructive: it disconnects the device from the mesh and invalidates any TermiPass session bound to it. The CLI prompts for `[y/N]` confirmation by default; for unattended scripts, pass `--yes` (non-TTY stdin without `--yes` is a hard error so a missed pipe doesn't silently destroy state). Tag values are normalized client-side into the `tag:<name>` form Headscale stores, so callers should pass the bare name (`--tag ops`, not `--tag tag:ops`) — both forms work, the CLI dedupes and normalizes either way.

The Phase 3c1 surface plus the Phase 3c2 SSH / subroutes toggles below cover the boolean ACL flips. The richer per-app ACL editor (`/api/acl/app/status`) lands in Phase 3c3 — see `vpn acl` further down.

### `vpn ssh` / `vpn subroutes` — boolean ACL toggles (Phase 3c2)

```bash
olares-cli settings vpn ssh status                  # GET /api/acl/ssh/status
olares-cli settings vpn ssh enable                  # POST /api/acl/ssh/enable
olares-cli settings vpn ssh disable                 # POST /api/acl/ssh/disable

olares-cli settings vpn subroutes status            # GET /api/acl/subroutes/status (raw upstream JSON)
olares-cli settings vpn subroutes enable
olares-cli settings vpn subroutes disable
```

Both toggles send an explicit empty `{}` body to match the SPA's request shape, even though the upstream doesn't read the body. `subroutes status` always renders raw JSON because the upstream shape isn't strongly typed.

### `vpn acl` — per-app ACL editor (Phase 3c3)

```bash
olares-cli settings vpn acl get my-app                              # GET /api/acl/app/status?name=<app>
olares-cli settings vpn acl get my-app -o json

olares-cli settings vpn acl set    my-app --tcp 80,443 --udp 53    # full replace
olares-cli settings vpn acl add    my-app --tcp 8080               # read-modify-write merge
olares-cli settings vpn acl remove my-app --tcp 80                 # read-modify-write drop
olares-cli settings vpn acl rm     my-app --udp 53                 # alias of `remove`
olares-cli settings vpn acl clear  my-app                          # POST with acls:[] (prompts; --yes for automation)
```

`--tcp` / `--udp` accept either repeated flags (`--tcp 80 --tcp 443`) or comma-separated values (`--tcp 80,443`); both forms are deduped client-side. Port strings are passed verbatim — Headscale accepts single ports, ranges (`8000-8100`), `*`, etc., and the CLI doesn't second-guess the format.

The upstream replaces the **whole** per-app ACL vector on every POST; there is no add / remove endpoint. `vpn acl add` and `vpn acl remove` are read-modify-write sugar over the same POST so unrelated entries survive untouched (matching how the SPA's add / remove buttons work). When `vpn acl get` returns no rows for an app, that's the upstream saying "no ACL configured" (HTTP 200 with `code != 0`) — the CLI surfaces this as an empty list rather than a hard error, the same way the SPA does.

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

### `apps` — lifecycle + per-app config (Phase 3a + 3b)

```bash
# Lifecycle (mirrors user-service /api/app/{suspend,resume})
olares-cli settings apps suspend my-app             # POST /api/app/suspend  body {name, all:false}
olares-cli settings apps resume  my-app             # GET  /api/app/resume/<name>

# Per-app environment variables (read-modify-write merge)
olares-cli settings apps env get my-app
olares-cli settings apps env set my-app \
  --var API_URL=https://api.example.com \
  --var LOG_LEVEL=debug

# Per-app secret store (/admin/secret/<app>; same X-Authorization Bearer)
olares-cli settings apps secrets list my-app
olares-cli settings apps secrets set  my-app --key API_KEY --value abc123
echo -n "$API_KEY" | olares-cli settings apps secrets set my-app --key API_KEY --value-stdin
olares-cli settings apps secrets delete my-app --key API_KEY              # prompts; --yes for automation
```

`apps env set` always reads the current env vector first (so unrelated values stay intact) and then PUTs the merged result back. Variables the SPA flags as `editable: false` are rejected by the upstream. `apps secrets set` upserts by default (POST first, fall back to PUT on "already exists"); use `--create-only` / `--update-only` when you need strict semantics. The CLI never echoes secret values in the table output (`-o table` shows only keys); pass `-o json` if you need the raw payload, but treat that as a credential dump.

### `apps` — per-app permissions / entrances / per-entrance config (Phase 3b/3d)

```bash
# Read-only inspection of declared permissions and live entrance vector
olares-cli settings apps permissions     my-app          # GET /api/applications/permissions/<app>
olares-cli settings apps providers list  my-app          # GET /api/applications/provider/registry/<app>
olares-cli settings apps entrances list  my-app          # GET /api/applications/<app>/entrances (fresher than `apps get`)

# Custom domain on a single entrance (read-modify-write)
olares-cli settings apps domain get   my-app www
olares-cli settings apps domain set   my-app www \
  --third-level mysite \
  --third-party www.example.com \
  --cert-file /path/to/fullchain.pem \
  --key-file  /path/to/privkey.pem
olares-cli settings apps domain set   my-app www --clear-third-level --clear-third-party
olares-cli settings apps domain finish my-app www        # GET .../setup/domain/finish (verify CNAME)

# Two-factor / one-time-link policy on a single entrance (read-modify-write)
olares-cli settings apps policy get my-app www
olares-cli settings apps policy set my-app www \
  --default-policy two_factor \
  --one-time true \
  --valid-duration 3600 \
  --sub-policy "/api/private=two_factor" \
  --sub-policy "/admin=password"
olares-cli settings apps policy set my-app www --sub-policies-file ./sub-policies.json
olares-cli settings apps policy set my-app www --clear-sub-policies

# Authorization level on a single entrance (private | public | internal)
olares-cli settings apps auth-level set my-app www --level public
```

`apps domain set` and `apps policy set` both fetch the current setup first and only overwrite the fields you pass, so you don't accidentally drop a cert when you only meant to flip a sub-domain. `--cert-file` / `--key-file` are read locally and POSTed as PEM strings — never paste cert/key bodies into the command line. `--clear-sub-policies` sends `null` (matches the SPA's "remove all sub-policies" semantics); use `--sub-policies-file` for full-replace, `--sub-policy` for per-path adjustments. The `auth-level` verb is write-only (re-read with `apps entrances list` or `apps get` to see the result). All four verbs target the BFL-style `/api/applications/<app>/<entrance>/setup/{domain,policy,auth-level}` routes.

### `network` — reverse-proxy mode (Phase 4)

```bash
olares-cli settings network reverse-proxy set --mode public-ip --ip 203.0.113.5
olares-cli settings network reverse-proxy set --mode frp \
  --frp-server frp.example.com --frp-port 7000 \
  --frp-auth-method token --frp-auth-token "$FRP_TOKEN"
olares-cli settings network reverse-proxy set --mode cloudflare-tunnel
olares-cli settings network reverse-proxy set --mode off
```

`reverse-proxy set` is a read-modify-write that fetches the existing config, applies the user's `--mode` + per-field flags, and POSTs the merged config back to `/api/reverse-proxy`. Unrelated fields (e.g. an FRP token you don't want to retype when switching modes) survive untouched.

### `advanced env` — system + user environment variables (Phase 4)

```bash
olares-cli settings advanced env system list
olares-cli settings advanced env system set --var FOO=bar --var BAZ=qux
olares-cli settings advanced env user   list
olares-cli settings advanced env user   set --var EDITOR=nvim
```

Same read-modify-write semantics as `apps env set`. The upstream rejects writes to system entries the SPA flagged as `editable: false`.

### `backup` — plan + snapshot + password writes (Phase 6)

```bash
# Plan lifecycle (no create / update — those need a richer flag UX, see "Out of scope" below)
olares-cli settings backup plans pause  <plan-id>
olares-cli settings backup plans resume <plan-id>
olares-cli settings backup plans delete <plan-id>                    # prompts; --yes for automation

# Snapshots
olares-cli settings backup snapshots run    <backup-id>              # POST .../snapshots {event:"create"}
olares-cli settings backup snapshots cancel <backup-id> <snapshot-id>  # DELETE-with-body {event:"cancel"}; prompts

# Repository password (separate from the plan record itself)
olares-cli settings backup password set my-plan                      # interactive (hidden + confirm)
echo -n "$PW" | olares-cli settings backup password set my-plan --password-stdin
```

`backup plans delete` orphans every snapshot the plan produced — the prompt body is explicit about this. Snapshot cancel is a `DELETE` with a `{"event": "cancel"}` body, mirroring backup-server's axios-style API. Repository passwords cannot be recovered: losing one means losing the ability to decrypt existing snapshots. The upstream encrypts them at rest, but the CLI defends against accidental disclosure by prompting for the password without echo (or reading once from `--password-stdin`).

### `restore` — plan create / cancel + URL pre-flight (Phase 6)

```bash
# Probe a remote backup URL before creating a restore plan
olares-cli settings restore plans check-url \
  --backup-url s3:s3.amazonaws.com/bucket/repo \
  --password "$REPO_PW"

# Create from an existing snapshot (id from `settings backup snapshots list`)
olares-cli settings restore plans create-from-snapshot \
  --snapshot-id <sid> --path /restore/here

# Create from a custom URL + password
olares-cli settings restore plans create-from-url \
  --backup-url s3:s3.amazonaws.com/bucket/repo \
  --password-stdin --path /restore/here --dir subdir

# Cancel a running restore
olares-cli settings restore plans cancel <plan-id>                   # prompts; --yes for automation
```

`check-url` lists candidate snapshots without committing to a restore. The two `create-from-*` verbs both POST to `/apis/backup/v1/plans/restore`; pick the variant that matches what you have on hand (snapshot id vs raw URL). Cancel uses the same DELETE-with-body shape as `backup snapshots cancel`. **There is intentionally no `update` or non-cancel `delete` verb on restore plans — backup-server has no routes for them.**

## What's NOT shipped yet

Anything not listed above either needs more design work or requires JWS-signed bodies the CLI can't produce yet:

- **App lifecycle (install / uninstall / upgrade / start / stop / cancel / clone)** — these route through the market service, not user-service; use `olares-cli market install|uninstall|upgrade|start|stop|cancel|clone` instead of `settings apps`. (Per-app suspend/resume + secrets/env/permissions/entrances/domain/policy/auth-level + per-app VPN ACLs do ship under `settings apps` / `settings vpn acl`.)
- **Network writes that require a JWS-signed device-id header** — hosts-file write, FRP server register/delete, SSL enable/disable/update, external-network master switch (the SPA reads/writes these via `X-Signature` headers the CLI doesn't produce yet).
- **Containerd registry mutations** — `registries mirrors put/delete`, `images delete/prune` (also `X-Signature`-gated).
- **Hardware / restart-class** — reboot, shutdown, ssh-password, OS upgrade — these go through TermiPass-issued JWS over a QR callback URL today; the CLI will gain support once we have a JWS key sourcing path.
- **Collect logs** — `POST /api/command/collectLogs` is `X-Signature`-gated.
- **Backup plan create / update** — full `BackupPolicy` + `LocationConfig` vector; needs either a `--from-file plan.json` mode or an upstream "create from defaults" shortcut before shipping.
- **Restore plan update / non-cancel delete** — backup-server has no routes for these.

Every area's `--help` lists what's currently implemented vs deferred so the agent can stay calibrated.

## Common errors → fixes

| Error message | Cause | Fix |
|---|---|---|
| `refresh token for <id> became invalid at <ts>; please run: olares-cli profile login --olares-id <id>` | `/api/refresh` itself returned 401/403 — the grant is dead (typed `*credential.ErrTokenInvalidated`) | `olares-cli profile login --olares-id <id>`. Defer the full recovery flow to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md). |
| `no access token for <id>; run: olares-cli profile login --olares-id <id>` | Profile selected but keychain has no entry (typed `*credential.ErrNotLoggedIn`) | `olares-cli profile login` or `profile import`. |
| `server rejected the access token (HTTP 401/403)` | Server still rejects after auto-refresh — rare, usually server-side state drift | Defer to [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (login + profile rules). |
| `this command needs role "<R>" or higher to <verb>, but profile "<id>" is cached as "<r>"` | Cached role below the verb's requirement | If your role on the server changed, run `olares-cli profile whoami --refresh`. Otherwise ask the owner to grant you the right role. |
| `HTTP 403 while attempting to <verb>` (with the same refresh hint appended) | Server rejected even though cache said OK — usually a stale **role** cache (NOT a stale token; the transport already handled that) | Run `olares-cli profile whoami --refresh`, then retry the verb. |
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
- For writes that take secrets (`integration accounts add awss3|tencent`, `me password set`, `apps secrets set`, `backup password set`, `restore plans check-url`, `restore plans create-from-url`), **always** read the secret from an env var or stdin pipe — never paste it into the chat or expand it inline in an `olares-cli ...` command line you suggest. Bash history retention is the user's responsibility; the agent should default to `printf '%s\n' "$VAR" | ... --password-stdin` (or the verb's equivalent `--value-stdin` / `--passwords-stdin` flag) style invocations whenever a verb supports it.
- `me password set` hashes locally; the raw password never leaves the machine. The agent should still avoid logging the input — even hidden-prompted data ends up on screen recordings and accessibility tooling. Never `echo` the result of a hash either: a leaked salted-MD5 is functionally as dangerous as the raw password against this backend.
- Read-only verbs do **not** carry "this will change X" prompts — only Phase 2+ writes do, and the prompts they do carry come from the upstream server's own response messages. Don't fabricate one for read verbs.
- The `me whoami --refresh` recovery path is the only authentication-adjacent action this skill should ever recommend. **All** other auth recovery (login expiry, profile import, 2FA) belongs in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
