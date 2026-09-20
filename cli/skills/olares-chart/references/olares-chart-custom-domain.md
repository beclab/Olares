# Custom URL: route ID or your own domain (post-deploy)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This is the last step of a port — the app is already installed and `running`, and the developer wants a URL they can say out loud.
> **Verb mechanics** (flags, RMW merge, wire fields) live in the [`olares-settings`](../../olares-settings/SKILL.md) skill, under `apps` → **`domain set` — RMW semantics + cert/key handling**. This file owns the **process**: which of the two pipelines applies, which stage the entrance is in, and what to ask the developer for next.

Both pipelines run against a **live app** through `olares-cli settings apps domain`, so they need login and an entrance name. Never guess the entrance:

```bash
olares-cli settings apps entrances list <app>     # entrance NAME + AUTH LEVEL (its STATE and URL columns read '-')
```

**Either pipeline's `domain set` is an app upgrade.** app-service moves the app to `Upgrading` and reconciles asynchronously, which has two consequences in both pipelines: the call is refused with `<op> operation is not allowed for <state> state` while the app is mid-operation, and the routing is live only once the app reads `running` again — not when the command returns.

## Two pipelines, not one flow

The two options share a verb and nothing else. A custom **route ID** stays inside the Olares zone, so the platform owns DNS and TLS and there is no DNS or certificate work at all — one `domain set` and you are done. A **third-party domain** is the developer's own FQDN: they own DNS, they supply the certificate, part of the work happens in a control panel this CLI cannot reach, and it runs a five-stage pipeline that starts with the auth level and ends with polling for `cname_status=active`.

## Pipeline A — custom route ID

The user manual calls this the **custom route ID**; the CLI flag and the wire field call the same thing `--third-level` / `third_level_domain`. Say both names once, or the developer will look for a feature that does not exist.

```bash
olares-cli settings apps domain set <app> <entrance> --third-level <prefix>
```

The old URL keeps working; the new one is `https://<prefix>.<zone>`. There is no CNAME and no `finish` in this pipeline — an agent that waits for `cname_status` here waits forever. Wait for the app to read `running` again (the upgrade above), then confirm the new URL over HTTPS.

**Conflicts are rejected by the write itself**, so read the error instead of pre-checking: `auth`, `desktop` and `wizard` are reserved, a prefix already used in the caller's zone (or colliding with an app's default domain) is refused, and a third-party domain must be unique across **every user**, not just your own apps. `doctor thirdleveldomain` in the [`olares-doctor`](../../olares-doctor/SKILL.md) skill audits a zone after the fact; it is not the gate.

## Pipeline B — the developer's own domain

### Stage detection: read before you write

`domain get` is the whole state machine. Read it first and act on **one** stage only. Handing over every command at once is how a session ends up stalled at a CNAME nobody added, or re-running `set` against a domain that was already activating.

```bash
olares-cli settings apps domain get <app> <entrance>
```

| What `domain get` shows | Stage | The only next step |
|---|---|---|
| `third_party_domain` empty | not started | Read `AUTH LEVEL` from `entrances list`. Not `public` → set it first (below) |
| `third_party_domain` set, `cname_target_status` empty or `unset` | waiting on the developer's DNS | Give them the CNAME **verbatim** (below) and stop. Do not run `finish` yet |
| `cname_target_status=set`, `cname_status=pending` | activating | Poll `domain get`. Nothing to change |
| `cname_status=active` | done | Confirm over HTTPS and stop |
| `cname_status=cert-not-found` / `cert-invalid` | the certificate, not DNS | Re-check the PEM pair (below), then `domain set` again |
| `cname_status=pending` for far longer than the registrar's TTL | most likely no record, or one aimed elsewhere | Compare the live record with `cname_target`, fix it, `finish` again. A missing record has no failure state — this is where it lands |
| `cname_status=timeout` (rarely `error`) | verification gave up | Same comparison; if the record already matches, treat it as platform-side |
| anything else | unknown to this skill | Show the value as-is. Never fold an unrecognized status into "probably fine" |

### The CNAME is where this goes wrong

`cname_target` is the **Olares zone** (e.g. `laresprime.olares.com`), and it is issued by the server. Pass it through untouched, and split the record for the developer explicitly, because a DNS panel asks for two fields and both are easy to fill in wrong:

| Record field | Value | Common mistake |
|---|---|---|
| Type | `CNAME` | — |
| Name / Host | the record name **relative to the DNS zone** — `media` for `media.n1.monster` when the managed zone is `n1.monster`; `foo.bar` when it is `example.com` | using `@`, or copying a full FQDN into a provider that expects a relative name |
| Value / Target | `cname_target` **exactly as printed** | using the app's current Olares URL, the `<appid>` host, or the FQDN being set up |

Never derive the target from the app's access URL. It is not the same string, and the resulting record resolves to something that will never activate.

### Sequence

```bash
# 1. Auth level must be public (see the constraint below).
olares-cli settings apps auth-level set <app> <entrance> --level public

# 2. Register the domain together with its cert pair.
olares-cli settings apps domain set <app> <entrance> \
  --third-party <fqdn> --cert-file <cert.pem> --key-file <key.pem>

# 3. Read the target, hand the developer the record, and wait for them.
olares-cli settings apps domain get <app> <entrance>

# 4. Only once they confirm the record exists:
olares-cli settings apps domain finish <app> <entrance>

# 5. Poll. Propagation is minutes to hours, set by their registrar.
olares-cli settings apps domain get <app> <entrance>
```

**`finish` does not check DNS.** It flips `cname_target_status` to `set` and `cname_status` to `pending`, and asks the platform to start verifying — that is all. Running it before the record exists is not an error and produces no warning, and **a missing or broken record never turns into a failed status**: when the platform's lookup comes back empty or errored it abandons the pass without writing anything back, so the entrance simply stays `pending`. Judge that case by elapsed time, and gate `finish` on the developer confirming the record.

## Hard constraints

- **Auth level must be `public`.** A custom domain cannot carry Olares authentication, so BFL rejects the write while the entrance is `private` (`custom domain can not be set when auth level is private`). `internal` is not reachable from the internet, which defeats the purpose.
- **The certificate and its private key must be RSA, PEM, and readable by you.** `--cert-file` / `--key-file` are read verbatim from disk, so a root-only path (usually `/etc/letsencrypt/live/...`) must be **copied** — not moved, or renewal breaks — into a private temporary directory. Keep the key copy at `0600` and delete it after `domain set` returns. The cert is normally the full chain.
- **Certbot defaults to ECDSA; ask for `--key-type rsa` at issue time.** Its default key arrives as ECDSA inside a `-----BEGIN PRIVATE KEY-----` (PKCS#8) block, and the platform's validator assumes PKCS#8 means RSA — the result is an unreadable failure, not a clean rejection. Say this before issuance so the challenge is not repeated.
- **`domain set` is read-modify-write.** Passing only `--third-level` leaves an existing third-party domain in place, and vice versa. Dropping one dimension needs `--clear-third-level` / `--clear-third-party`.
- **The domain must be fully qualified.** `example.com` and `media.example.com` are fine; a bare label or a trailing-dot form is rejected before anything is stored.
- **One entrance at a time.** Domain setup is per-entrance, and an app with several entrances needs the whole pipeline repeated per entrance that should get its own hostname.

## Auth level persistence: know the version boundary

On the **1.12.7** line, a runtime auth-level change is stored in an override slot — per-user `Spec.UserSettings[caller]` for shared apps, app-global `Spec.Settings["authLevel"]` otherwise — so it survives the reconciler reprojecting chart values over `Spec.Entrances`. The custom domain is stored the same way. Two consequences:

- The chart's `entrances[].authLevel` is the **install-time default**, nothing more. Do not edit the chart and redeploy to make a `public` entrance stick; the override already handles it.
- **On 1.12.6 and older** there is no override slot, so a reconcile or upgrade can put the chart's value back and return the entrance to `private`, taking the custom domain's requirement with it. Re-read the auth level after any upgrade there.

Either way, **read state back rather than assuming a side effect**: re-run `entrances list`, and call `market resume` only if the app actually reads `stopped`. An unconditional resume is a surprise restart at worst.

## Tools you may need but must not install

If the developer already has a valid certificate, none of this applies — take the files and move on. When one has to be issued and they choose certbot, `command -v certbot`; if it is missing, say what it is for and let them install it. Do not install it for them.

Two facts about issuing are worth passing on, because each costs a DNS round trip to discover: a DNS-01 challenge needs a **TXT** record, which is a different record from the **CNAME** that puts the domain into service (the TXT one can be removed afterwards), and RSA must be asked for explicitly per the constraint above.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `custom domain can not be set when auth level is private` | Pipeline B started before the auth level was changed | `auth-level set --level public`, then retry |
| `third_level_domain "x" is reserved and cannot be used` | The prefix is `auth`, `desktop` or `wizard` | Pick another prefix |
| `... is already used by entrance "e" of app "a"` / `... conflicts with the default domain of ...` | The prefix is taken in this zone, or the FQDN is already claimed — third-party domains are unique across every user | Pick another value; `doctor thirdleveldomain` lists what a zone already holds |
| `<op> operation is not allowed for <state> state` | `domain set` arrived while the app was mid-operation | Wait until the app reads `running`, then retry |
| `--third-party requires both --cert-file and --key-file` | Domain passed without its cert pair | Supply both, or `--clear-third-party` to drop the domain |
| `app not set custom domain` from `finish` | `finish` ran on an entrance with no third-party domain — often a typo'd entrance, or `set` silently targeted a different one | `domain get` the entrance and confirm `third_party_domain` before finishing |
| Cert file unreadable | The PEM lives in a root-only directory | Copy it into a private temporary directory; preserve `0600` on the key, delete the copy after use, and do not move the renewal source |
| `cname_status` stuck at `pending` | The record is missing, still propagating, or points somewhere other than `cname_target` | Re-read `cname_target` and compare it against what the panel actually holds |
