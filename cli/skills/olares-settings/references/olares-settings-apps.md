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
| `env set <app> --var KEY=VALUE [--var ...]` | normal | VERIFIED | Patch the named variables only |
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


Editing one entrance — its custom domain, its auth policy, its auth level — is a pipeline of its own: [per-entrance configuration](olares-settings-apps-entrance.md).
## `env get` / `env set`

```bash
olares-cli settings apps env get gitea
olares-cli settings apps env set gitea --var GITEA_TOKEN=abc --var DB_PASS=xyz
```

- **Values go through `--var KEY=VALUE`, repeated per variable** — positional pairs are not accepted. Only the first `=` splits, so a value may contain one: `--var "GREETING=hi=there"`.
- `env set` sends only the variables you name; the upstream patches those entries and leaves the rest of the vector alone, so there is no need to read the current env first.
- **Only the variables the app's chart declares can be set, and only the editable ones** — `env get --output json` shows both. An undeclared name is dropped by the upstream, so `env set` checks the response and fails naming the ignored keys; a read-only one fails the whole request with `app env '<name>' is not editable`.
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
- **For UNVERIFIED verbs** (`domain set/finish`, `policy set`, `auth-level set`), the result is provisional — confirm the outcome after running.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `entrance '<name>' not found on app '<app>'` | Typo / chart vs user-facing name mismatch | `apps entrances list <app>` to enumerate |
| `--cert-file and --key-file are required when --third-party is set` | Third-party domain without cert | Provide both, or use `--clear-third-party` |
| `custom domain can not be set when auth level is private` | Third-party domain on a `private` entrance | `auth-level set <app> <entrance> --level public`, then retry |
| `app not set custom domain` from `domain finish` | `finish` ran on an entrance with no third-party domain — usually the wrong entrance | `domain get` that entrance and check `third_party_domain` first |
| `--default-policy: invalid value 'X' (allowed: system, one_factor, two_factor, public)` | Typo | Use one of the four valid values |
| `auth-level get is not supported (no upstream endpoint); use 'apps entrances list <app>' to read the AUTH LEVEL column` | Tried to GET auth-level | Read from `entrances list` |
