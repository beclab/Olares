---
name: olares-shared
version: 4.2.0
description: "Olares profile and auth foundation for olares-cli — prerequisite for all olares-* skills. Profile login, import, list, use, remove, keychain tokens, 401/403 recovery. Use for Olares ID, profile, login, 2FA/TOTP, refresh token, keychain, auth errors."
compatibility: Requires olares-cli on PATH
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# olares-cli shared rules

Foundation for every other `olares-cli` skill. Every business verb under `cluster` / `files` / `market` / `settings` / `dashboard` rides the active profile's token. **Read this first.**

> **This skill also hosts the cross-skill platform model** in [references/olares-platform.md](references/olares-platform.md) — the userspace storage model, uid-1000 run identity, system-managed `drive/Home` dirs, app/namespace & networking, system middleware, and version/semver. `files` / `chart` / `cluster` link there (one hop) instead of re-describing it. That reference is pure platform model and needs no login.

> **Source of truth for flags & syntax is always `olares-cli profile --help`.** This file only carries what `--help` cannot give: the profile mental model, agent-driven login flow, token-storage backends, refresh semantics, and the error → fix matrix.

## When to use

- Olares, Olares ID, olares-cli, OpenClaw on Olares, profile, login, logout, two-factor / 2FA / TOTP, refresh token, keychain
- Auth errors: `server rejected the access token`, `refresh token for X became invalid`, `no access token for X`, `already authenticated`, `two-factor authentication required`

## Profile model

One profile = one Olares instance + one user identity, keyed by **olaresId** (e.g. `alice@olares.com`). Each profile owns its own access_token / refresh_token pair, stored in the OS keychain.

| Command | Purpose |
|---------|---------|
| `olares-cli profile login` | Mode A — password (+ TOTP if 2FA is on); auto-creates the profile on first run |
| `olares-cli profile import` | Mode B — bootstrap an access_token from an existing refresh_token |
| `olares-cli profile list` | List every profile, mark the current one, show login status per profile |
| `olares-cli profile use <name\|->` | Switch the current profile; `-` reverts to the previous one (like `cd -`) |
| `olares-cli profile remove <name>` | Delete a profile and its stored token in one shot |

> **There is no `auth login` / `auth logout` namespace and no per-invocation `--profile` override flag.** Everything lives under `profile`. "Logout" is `profile remove`. Identity is whichever profile is currently selected; to target a different one, run `olares-cli profile use <name>` first.

## Login modes

### Mode A — password (+ optional TOTP)

```bash
olares-cli profile login --olares-id <olaresId>
```

- Interactive: prompts for password (echo disabled); prompts for TOTP if 2FA is enabled.
- Scripted: pipe via `--password-stdin`; if 2FA is on, you MUST also pass `--totp <code>` because there is no second prompt.
- **There is no `--password <plaintext>` flag** — passwords are never accepted on the command line.

### Mode B — existing refresh_token

```bash
olares-cli profile import --olares-id <olaresId> --refresh-token "$OLARES_REFRESH_TOKEN"
```

Exchanges the refresh_token for an access_token once via `/api/refresh` and writes both to the keychain. **Read the token from an env var or secret manager — never inline plaintext.**

### Agent-driven login (recommended)

When you (an AI agent) drive the login on the user's behalf, do NOT pass password / TOTP as command-line arguments. Spawn `olares-cli profile login --olares-id <id>` as a background process so it parks at the password prompt, forward the prompt to the user, and read its output after the command exits to confirm success.

## Switching and inspecting profiles

`profile list` output:

```
   NAME             OLARES-ID              STATUS
*  alice            alice@olares.com       logged-in (23h59m)
   bob              bob@olares.com         expired
   eve              eve@olares.com         invalidated
   frank            frank@olares.com       never
```

| STATUS | Meaning | Recovery |
|--------|---------|----------|
| `logged-in (Xh Ym)` | Token valid; column shows time-to-expiry | — |
| `logged-in` | Token present but JWT has no exp claim (can't verify locally) | Trust until the server says no |
| `expired` | JWT exp is in the past | `profile login` |
| `invalidated` | Server explicitly rejected the refresh leg | `profile login` directly (no need to `profile remove` first) |
| `never` | No token has ever been stored | `profile login` or `profile import` |

The leading `*` marks the current profile. `profile use` accepts either the NAME alias or the olaresId.

## Token storage

| OS | Backend | Location |
|------|---------|----------|
| darwin | macOS Keychain | service `olares-cli`, account = olaresId |
| linux | AES-256-GCM file | under `~/.local/share/olares-cli/` |
| windows | DPAPI | `HKCU\Software\OlaresCli\keychain` |

After `login` / `import` succeeds, the CLI prints `token stored via <backend> (service "olares-cli", account "<id>")`. If the backend resolves to `file-fallback` (sandboxed / CI environments), be aware the token now sits in a file with **different security properties than the system keychain**.

> **The plaintext `~/.olares-cli/tokens.json` from older builds is deprecated** — if a user upgraded and suddenly appears "logged out", `profile login` is the fix.

## Re-authentication rules

`profile login` and `profile import` both reject the case "a still-valid token already exists for this olaresId" — to force-overwrite, run `profile remove <id>` first. Expired / invalidated / never-logged-in profiles get the new token written in place; this lets scripts call `login` after `invalidated` without an extra `remove` step.

## Automatic token refresh

**The CLI rotates expired access_tokens transparently.** Users do NOT need to run `profile login` just because their access_token aged out — only when the *refresh_token itself* becomes invalid.

- Replayable requests (every JSON verb, `files cat`, `files download`, `files rm`, `market` verbs, …): on 401/403 the transport calls `/api/refresh` and retries once with the new token.
- Streaming uploads (`files upload` chunks): pre-decode the JWT exp; if within 60s of expiry, refresh BEFORE sending, because once a `*os.File` chunk is consumed it can't be replayed on a 401.

Across goroutines AND across concurrent `olares-cli` processes, `/api/refresh` is hit at most once per stale token (in-process mutex + cross-process flock).

> **Do not implement custom retry/backoff loops on top of auth errors.** Once you see `ErrTokenInvalidated` or `ErrNotLoggedIn`, only `profile login` / `profile import` will help.

## Auth error recovery table

| Error message (excerpt) | Meaning | Fix |
|-------------------------|---------|-----|
| `refresh token for <id> became invalid at <ts>` | `/api/refresh` returned 401/403 — the grant is dead | `olares-cli profile login --olares-id <id>` |
| `no access token for <id>` | Profile selected but keychain has no entry | `olares-cli profile login` or `profile import` |
| `server rejected the access token (HTTP 401)` / `(HTTP 403)` | After auto-refresh the server still rejects (rare) | `olares-cli profile login --olares-id <id>` |
| `--olares-id is required` | login / import invoked without olaresId | Add `--olares-id <id>` |
| `already authenticated for <id> (expires in ...)` | Still-valid token exists | `olares-cli profile remove <id>` then re-run |
| `a token is already stored for <id> but its expiry can't be determined client-side` | Token present but JWT carries no exp claim | `profile remove <id>` then re-run |
| `two-factor authentication required: re-run with --totp <code>` | 2FA on, non-TTY context | Re-run with `--totp <code>`, or run interactively |
| `password is empty` / `TOTP code is empty` | stdin / TTY returned an empty string | Check for premature EOF or an empty pipe |
| `profile <name> not found` | `profile use` / `profile remove` referenced an unknown profile | `profile list` to see the actual names |

> **Do not silently retry auth errors.** 401/403 after auto-refresh and `already authenticated` are deterministic — follow the table; blind retries make it worse.

## Security rules

- **Never** invent a `--password <plaintext>` argument (it does not exist). Passwords go through the TTY or `--password-stdin` fed by a secret pipe.
- **Never** echo `access_token` / `refresh_token` to the terminal. When passing a refresh_token to `profile import`, source it from an environment variable: `--refresh-token "$OLARES_REFRESH_TOKEN"`.
- **Confirm intent before write/delete actions** (`profile remove`, `files rm`, `files upload --overwrite`, `cluster pod delete`, …). Do not act unilaterally on the user's behalf.
- **TOTP is not a password** — it is single-use and short-lived, so the CLI echoes it to make manual entry less error-prone. Never persist a TOTP in a shared script.
