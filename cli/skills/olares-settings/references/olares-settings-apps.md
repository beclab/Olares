# settings apps (post-install configuration)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings apps --help` and `olares-cli settings apps <verb> --help`.

The **post-install** surface for an Olares app. Inspect the app, list its entrances, edit per-entrance domain / policy / auth-level / env vars, suspend / resume.

> **NOT for install / uninstall / upgrade / clone / start / stop** — use [`olares-market`](../../olares-market/SKILL.md) for app lifecycle. `settings apps` is for tweaking an app that's already installed.

## Verbs at a glance

| Verb | Floor | Status | Purpose |
|---|---|---|---|
| `list [--all] [--show-system]` | normal | VERIFIED | Installed apps for the current user. Default filter mirrors the SPA |
| `get <app>` | normal | VERIFIED | Detail view (filtered client-side; no per-app endpoint) |
| `entrances list <app>` | normal | VERIFIED | **Live** entrance vector — names, state, auth-level |
| `env get <app>` | normal | VERIFIED | Current per-app env vector |
| `env set <app> KEY=VALUE [KEY=VALUE...]` | normal | UNVERIFIED | Replace the env vector |
| `domain get <app> <entrance>` | normal | VERIFIED | Per-entrance custom-domain setup |
| `domain list <app>` | normal | VERIFIED | Every entrance's domain setup |
| `domain set <app> <entrance> [flags]` | normal | UNVERIFIED | RMW update |
| `domain finish <app> <entrance>` | normal | UNVERIFIED | Confirm third-party CNAME after DNS propagates |
| `policy get <app> <entrance>` | normal | VERIFIED | Per-entrance auth policy |
| `policy list <app>` | normal | VERIFIED | Every entrance's policy |
| `policy set <app> <entrance> [flags]` | normal | UNVERIFIED | RMW update |
| `auth-level set <app> <entrance> --level X` | normal | UNVERIFIED | `private` / `public` / `internal` |
| `suspend <app> [--all]` | normal | UNVERIFIED | Suspend running app |
| `resume <app>` | normal | UNVERIFIED | Resume suspended app |

## The per-entrance editing pipeline

Most per-entrance edits follow a 4-step pattern:

```bash
# 1. Discover entrances on the app.
olares-cli settings apps entrances list firefox
# → produces ENTRANCE / STATE / AUTH LEVEL / DOMAIN columns

# 2. Inspect current setup of the entrance you want to edit.
olares-cli settings apps domain get firefox www
olares-cli settings apps policy get firefox www

# 3. RMW update (unspecified flags survive).
olares-cli settings apps domain set firefox www --third-level my-firefox
olares-cli settings apps policy set firefox www --default-policy two_factor

# 4. For third-party domains: confirm CNAME after DNS propagation.
olares-cli settings apps domain set firefox www --third-party firefox.example.com \
  --cert-file /path/to/cert.pem --key-file /path/to/key.pem
olares-cli settings apps domain finish firefox www
```

## `domain set` — RMW semantics + cert/key handling

```bash
# Update third-level only (host under .<terminus>).
olares-cli settings apps domain set firefox www --third-level my-firefox

# Update third-party domain — REQUIRES --cert-file AND --key-file.
olares-cli settings apps domain set firefox www \
  --third-party firefox.example.com \
  --cert-file /etc/letsencrypt/live/firefox.example.com/fullchain.pem \
  --key-file /etc/letsencrypt/live/firefox.example.com/privkey.pem

# Explicitly drop a domain dimension (RMW would otherwise preserve it).
olares-cli settings apps domain set firefox www --clear-third-party
olares-cli settings apps domain set firefox www --clear-third-level
```

- **Unspecified flags survive** — RMW under the hood. Pass `--clear-*` to drop a dimension.
- **Third-party domains REQUIRE both `--cert-file` AND `--key-file`** (unless `--clear-third-party`). The PEM bytes are POSTed verbatim as multi-line strings.
- After `domain set --third-party`, run `domain finish` once the CNAME propagates. Without `finish`, the SPA shows "pending" status.

## `policy set` — replace sub-policies

```bash
# Update default policy.
olares-cli settings apps policy set firefox www --default-policy two_factor

# Set one-time-link mode (factor cap).
olares-cli settings apps policy set firefox www --one-time true --valid-duration 3600

# Replace sub-policies (any --sub-policy flag drops the existing set).
olares-cli settings apps policy set firefox www \
  --sub-policy "uri=/admin,policy=two_factor" \
  --sub-policy "uri=/api,policy=public"

# Explicitly drop sub-policies.
olares-cli settings apps policy set firefox www --clear-sub-policies

# Bulk file form.
olares-cli settings apps policy set firefox www --sub-policies-file ./sub-policies.json
```

- `--default-policy` values: `system` | `one_factor` | `two_factor` | `public`
- Default-policy / one-time / valid-duration flags follow RMW semantics.
- **Sub-policy entries are REPLACED in full whenever any sub-policy flag is passed.** This is intentional — partial sub-policy edits don't compose safely. Pass `--clear-sub-policies` to drop the existing set without adding new ones.

## `auth-level set` — no GET endpoint upstream

```bash
olares-cli settings apps auth-level set firefox www --level public
```

| Level | Reachability |
|---|---|
| `private` | Only the app's owner |
| `public` | Any authenticated user |
| `internal` | Intra-cluster traffic only |

> **There is no `auth-level get` verb** because no GET endpoint exists upstream. To inspect the current level, run `apps entrances list <app>` and read the `AUTH LEVEL` column.

## `env get` / `env set`

```bash
olares-cli settings apps env get gitea
olares-cli settings apps env set gitea GITEA_TOKEN=abc DB_PASS=xyz
```

- `env set` REPLACES the full env vector. Read current env first if you only want to add a single var.
- For secrets, pipe via env var or stdin redirection. Don't paste the value into chat.

## `list` filters

```bash
olares-cli settings apps list                  # SPA-equivalent filtered view (current user, no system apps)
olares-cli settings apps list --show-system    # include system apps
olares-cli settings apps list --all            # every state + every kind (cluster-wide for admins)
```

`get <app>` filters client-side — there is no per-app endpoint upstream. For multi-instance / cloned apps, pass the per-instance name (e.g. `windowsefe992`), not the source name (`windows`).

## Agent best practices

- **Always run `entrances list <app>` before any per-entrance edit.** Don't assume entrance names; the user-facing names (e.g. `www`) often differ from the chart-defined service names.
- **For `policy set`**, surface `policy get` output to the user BEFORE applying — the sub-policy replacement semantics are easy to misuse.
- **For `domain set --third-party`**, remind the user to set up the CNAME at their DNS provider AND run `domain finish` once it propagates.
- **For UNVERIFIED verbs** (`env set`, `domain set/finish`, `policy set`, `auth-level set`), the result is provisional (not yet smoke-tested against a live instance) — confirm the outcome after running.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `entrance '<name>' not found on app '<app>'` | Typo / chart vs user-facing name mismatch | `apps entrances list <app>` to enumerate |
| `--cert-file and --key-file are required when --third-party is set` | Third-party domain without cert | Provide both, or use `--clear-third-party` |
| `--default-policy: invalid value 'X' (allowed: system, one_factor, two_factor, public)` | Typo | Use one of the four valid values |
| `auth-level get is not supported (no upstream endpoint); use 'apps entrances list <app>' to read the AUTH LEVEL column` | Tried to GET auth-level | Read from `entrances list` |
