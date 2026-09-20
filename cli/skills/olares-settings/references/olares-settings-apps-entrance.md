# settings apps: per-entrance configuration

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md), the parent [`../SKILL.md`](../SKILL.md), and [post-install app configuration](olares-settings-apps.md) first — the verb table and the env vector live there.

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
olares-cli settings apps domain set firefox www --clear-third-party   # or --clear-third-level
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
```

- `--default-policy` values: `system` | `one_factor` | `two_factor` | `public`. That flag, `--one-time` and `--valid-duration` follow RMW semantics.
- **Sub-policy entries are REPLACED in full whenever any sub-policy flag is passed** — partial sub-policy edits don't compose safely, so this is intentional. `--clear-sub-policies` drops the set without adding new entries.

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

