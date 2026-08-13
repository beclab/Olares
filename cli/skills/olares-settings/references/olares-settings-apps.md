# settings apps (post-install configuration)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings apps --help` and `olares-cli settings apps <verb> --help`.

The **post-install** surface for an Olares app. Inspect the app, list its entrances, edit per-entrance domain / policy / auth-level / env vars, suspend / resume.

> **Primary lifecycle belongs to `olares-market`**: use it for install / uninstall / upgrade / clone / stop / resume. `settings apps` is primarily post-install configuration; its `suspend` / `resume` verbs are thin Settings-SPA aliases over `market stop` / `market resume`.

## Verbs at a glance

| Verb | Floor | Status | Purpose |
|---|---|---|---|
| `list [--all] [--show-system]` | normal | VERIFIED | Installed apps for the current user. Default filter mirrors the SPA |
| `get <app>` | normal | VERIFIED | Detail view (filtered client-side; no per-app endpoint) |
| `entrances list <app>` | normal | VERIFIED | Entrance names, auth levels, visibility. Its `STATE` and `URL` columns read `-`; for the app's host use `list` / `get` |
| `env get <app>` | normal | VERIFIED | Current per-app env vector |
| `env set <app> KEY=VALUE [KEY=VALUE...]` | normal | UNVERIFIED | Replace the env vector |
| `domain get <app> <entrance>` | normal | VERIFIED | Per-entrance custom-domain setup |
| `domain list <app>` | normal | VERIFIED | Every entrance's domain setup |
| `domain set <app> <entrance> [flags]` | normal | UNVERIFIED | RMW update |
| `domain finish <app> <entrance>` | normal | UNVERIFIED | Record that the CNAME exists and start asynchronous verification |
| `policy get <app> <entrance>` | normal | VERIFIED | Per-entrance auth policy |
| `policy list <app>` | normal | VERIFIED | Every entrance's policy |
| `policy set <app> <entrance> [flags]` | normal | UNVERIFIED | RMW update |
| `auth-level set <app> <entrance> --level X` | normal | UNVERIFIED | `private` / `public` / `internal` |
| `suspend <app> [--cascade]` | normal | UNVERIFIED | Suspend (stop) a running app — thin alias over `market stop` |
| `resume <app>` | normal | UNVERIFIED | Resume a suspended app — thin alias over `market resume` |

> `suspend` / `resume` carry **no settings-side logic**: on 1.12.6 the Settings page routes stop/resume through the Market flow, so the CLI reuses `market stop` / `market resume` verbatim (renamed verb only). They inherit the full market behavior — `--cascade` / `--watch`, source-implicit resolution, and the 1.12.6 force-cascade for CS/shared apps. See [`olares-market`](../../olares-market/SKILL.md) lifecycle for the cascade and watch semantics.

## The per-entrance editing pipeline

Most per-entrance edits follow a 4-step pattern:

```bash
# 1. Discover entrances (ENTRANCE / STATE / AUTH LEVEL / DOMAIN columns).
olares-cli settings apps entrances list firefox
# 2. Inspect the entrance you want to edit.
olares-cli settings apps domain get firefox www
olares-cli settings apps policy get firefox www
# 3. RMW update (unspecified flags survive).
olares-cli settings apps domain set firefox www --third-level my-firefox
olares-cli settings apps policy set firefox www --default-policy two_factor
```

> **The two domain kinds are two different pipelines, not two steps of one.** `--third-level` is complete on its own — no DNS record, no `finish`, and no `cname_*` field ever becomes relevant. Only `--third-party` involves a certificate, a CNAME the user adds at their registrar, `finish`, and then polling. The end-to-end procedure, including which stage an entrance is in and what to ask the user for, is the **Custom URL** capability of the [`olares-chart`](../../olares-chart/SKILL.md) skill.

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
- **Third-party domains REQUIRE both `--cert-file` AND `--key-file`** (unless `--clear-third-party`) and an entrance whose auth level is already `public` — BFL rejects the write with `custom domain can not be set when auth level is private`. The PEM bytes are POSTed verbatim as multi-line strings, so the files must be readable by the CLI's own user and the key must be RSA.
- After `domain set --third-party`, the user adds a CNAME pointing at `cname_target`, and **then** `domain finish` records that they have done so. `finish` does not resolve DNS itself.
- Either domain kind makes app-service upgrade the app, so the write is refused mid-operation and the route is live only once the app runs again.

## `domain get` — reading the third-party stage

Adding or changing the third-party domain resets both status fields to empty / `unset`; `finish` moves them to `set` / `pending`, and everything after that is the platform's asynchronous check writing back. An RMW update that keeps the same domain can preserve its existing state.

| `cname_target_status` | `cname_status` | Means |
|---|---|---|
| empty or `unset` | empty or `unset` | Domain registered; **`finish` has not run**. Waiting on the user's DNS record |
| `set` | `pending` | Activation submitted; the platform is verifying the CNAME and the certificate |
| `set` | `active` | Live. The custom domain serves the entrance |
| `set` | `cert-not-found` / `cert-invalid` | The certificate side failed, not DNS |
| `set` | `timeout` | Verification gave up. A record that is simply missing stays `pending` instead — it never turns into a failure |

`cname_target` is the user's Olares **zone** (e.g. `laresprime.olares.com`) and is the CNAME's *value*. Its *name* is the custom domain relative to the DNS zone managed at that provider (`media` for `media.n1.monster` in zone `n1.monster`, or `foo.bar` in zone `example.com`). Print the target verbatim; it cannot be derived from the app's current URL. Values outside this table are possible (the field is a plain string on the wire): show them as-is rather than guessing.

## `policy set` — replace sub-policies

```bash
# Any --sub-policy flag REPLACES the existing set. Bulk form: --sub-policies-file ./sub-policies.json
olares-cli settings apps policy set firefox www \
  --sub-policy "uri=/admin,policy=two_factor" \
  --sub-policy "uri=/api,policy=public"

# Drop the set without adding new entries.
olares-cli settings apps policy set firefox www --clear-sub-policies
```

- `--default-policy` values: `system` | `one_factor` | `two_factor` | `public`. That flag, `--one-time` and `--valid-duration` follow RMW semantics.
- **Sub-policy entries are REPLACED in full whenever any sub-policy flag is passed** — partial sub-policy edits don't compose safely, so this is intentional.

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
olares-cli settings apps list --show-system    # also include system apps (Files / Settings / Vault / ...)
olares-cli settings apps list --all            # also include uninstalled / pending / installing / upgrading / reinstalling states
```

`get <app>` filters client-side — there is no per-app endpoint upstream. For multi-instance / cloned apps, pass the per-instance name (e.g. `windowsefe992`), not the source name (`windows`).

## Agent best practices

- **Always run `entrances list <app>` before any per-entrance edit.** User-facing names (e.g. `www`) often differ from the chart-defined service names.
- **For `policy set`**, surface `policy get` output to the user BEFORE applying — the replacement semantics are easy to misuse.
- **For `domain set --third-party`**, run `domain get` first and act on that one stage only (table above). Hand over the CNAME name/value split verbatim, wait for them to confirm the record, then `finish` — a full command list given up front reliably ends with `finish` run against a record nobody added yet.
- **For UNVERIFIED verbs** (`env set`, `domain set/finish`, `policy set`, `auth-level set`), the result is provisional — confirm the outcome after running.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `entrance '<name>' not found on app '<app>'` | Typo / chart vs user-facing name mismatch | `apps entrances list <app>` to enumerate |
| `--cert-file and --key-file are required when --third-party is set` | Third-party domain without cert | Provide both, or use `--clear-third-party` |
| `custom domain can not be set when auth level is private` | Third-party domain on a `private` entrance | `auth-level set <app> <entrance> --level public`, then retry |
| `app not set custom domain` from `domain finish` | `finish` ran on an entrance with no third-party domain — usually the wrong entrance | `domain get` that entrance and check `third_party_domain` first |
| `--default-policy: invalid value 'X' (allowed: system, one_factor, two_factor, public)` | Typo | Use one of the four valid values |
| `auth-level get is not supported (no upstream endpoint); use 'apps entrances list <app>' to read the AUTH LEVEL column` | Tried to GET auth-level | Read from `entrances list` |
