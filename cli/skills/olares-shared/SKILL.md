---
name: olares-shared
version: 1.1.0
description: "Shared olares-cli foundation: profile model, first-time login (profile login with password + TOTP), bootstrapping a profile from an existing refresh token (profile import), switching/listing/removing profiles, the global --profile flag, where access/refresh tokens live in the OS keychain, automatic access_token refresh via /api/refresh (transparent reactive retry on 401/403; pro-active JWT-exp refresh for streaming uploads; cross-goroutine + cross-process deduplication via flock), and how to recover from auth errors (HTTP 401/403, refresh token invalidated, not logged in, two-factor authentication required). Use whenever the user is configuring olares-cli for the first time, logging in or importing credentials, switching/listing/removing profiles, asking how token refresh works, scripting parallel olares-cli invocations, or seeing errors like 'server rejected the access token', 'refresh token for X became invalid', 'no access token for X', 'already authenticated', or 'two-factor authentication required'; also use when the user asks about refresh tokens, the keychain, olaresId, or profile management."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli profile --help"
---

# olares-cli shared rules

This skill explains: what a profile is, how to obtain access credentials for it, where those credentials live, and how to recover when the server rejects a token. **Every `olares-cli files ...` (and other business) command depends on the profile selection + auth flow described here.**

## Profile model

One profile = one Olares instance + one user identity. The identity is uniquely keyed by an **olaresId** (e.g. `alice@olares.com`). Each profile owns its own access_token / refresh_token pair, stored in the OS keychain.

The `profile` command tree exposes 5 verbs (see [`cmd/ctl/profile/root.go`](cli/cmd/ctl/profile/root.go)):

| Command | Purpose |
|---------|---------|
| `olares-cli profile login` | Authenticate with a password (+ TOTP if 2FA is on); auto-creates the profile on first run (mode A) |
| `olares-cli profile import` | Bootstrap an access_token from an existing refresh_token (mode B) |
| `olares-cli profile list` | List every profile, mark the current one, show login status per profile |
| `olares-cli profile use <name\|->` | Switch the current profile; `-` reverts to the previous one (analogous to `cd -`) |
| `olares-cli profile remove <name>` | Delete a profile and its stored token in one shot |

> **There is no `olares-cli auth login` / `auth logout` namespace.** Every auth-related action lives under `profile`. "Logout" is `profile remove`.

## Global `--profile` flag

The root command registers a persistent `--profile <olaresId>` flag (see [`cmd/ctl/root.go`](cli/cmd/ctl/root.go) L57). It overrides the currently-selected profile for one invocation without flipping the persisted current pointer:

```bash
# Doesn't change current; this single ls runs against alice's credentials.
olares-cli files ls drive/Home/ --profile alice@olares.com
```

Use this for: scripting parallel operations against multiple profiles, sanity-checking a specific profile's status, and avoiding pollution of the interactive terminal's current pointer.

## First-time login (mode A: password + optional TOTP)

```bash
olares-cli profile login --olares-id <olaresId>
```

Behavior (see [`cmd/ctl/profile/login.go`](cli/cmd/ctl/profile/login.go) and [`cmd/ctl/profile/credentials.go`](cli/cmd/ctl/profile/credentials.go)):

- Profile does not exist → auto-created
- Profile exists, token expired or invalidated → reuse the profile entry, write a fresh token
- Profile exists, token still valid → **rejected** with a hint to run `olares-cli profile remove <id>` first

### Password

- Default: read from the controlling TTY with echo disabled
- Scripts: `--password-stdin`, e.g. `printf '%s' "$PW" | olares-cli profile login --olares-id <id> --password-stdin`
- **There is no `--password <plaintext>` flag.** The CLI deliberately omits it so passwords never leak into shell history or `ps` output.

### 2FA / TOTP

When the server's `/api/firstfactor` returns `fa2=true`, a second factor is required:

- TTY: the CLI prompts `two-factor code for <id>:` automatically; the user types the 6-digit code
- Non-TTY (`--password-stdin`, CI, etc.): you MUST pass `--totp <code>` up front, otherwise the command fails with `two-factor authentication required: re-run with --totp <code>`

> **The CLI does not try to guess whether 2FA is enabled.** It probes every login with a targetURL that triggers Authelia's 2FA policy. Accounts without 2FA pass through transparently; accounts with 2FA get prompted for TOTP and then proceed.

### Agent-driven login (recommended pattern)

When you (an AI agent) drive the login on the user's behalf, **do not** pass the password or TOTP as plaintext command-line arguments. Recommended flow:

1. Spawn `olares-cli profile login --olares-id <id>` as a background process so it parks at the password prompt.
2. Forward the prompt verbatim to the user and wait for them to type the password / TOTP into their own terminal.
3. After the command exits, read its output to confirm whether the login succeeded.

Alternatively, instruct the user to run the login in their terminal themselves; the agent then takes over for follow-up command orchestration.

## Bootstrap from an existing refresh_token (mode B)

If the user already has a refresh_token (from LarePass, the wizard activation flow, or any other source), there is no need to run through password + 2FA again:

```bash
olares-cli profile import --olares-id <olaresId> --refresh-token <tok>
```

The CLI exchanges the refresh_token for an access_token via a single `/api/refresh` call (see [`cmd/ctl/profile/import.go`](cli/cmd/ctl/profile/import.go)) and writes both into the keychain. The "reject if a valid token already exists" rule from `login` applies here too.

> **Never write `--refresh-token <tok>` as plaintext in scripts.** Read it from an environment variable or a secret manager:
> ```bash
> olares-cli profile import --olares-id <id> --refresh-token "$OLARES_REFRESH_TOKEN"
> ```

## Switching and inspecting profiles

### `profile list`

Output (see [`cmd/ctl/profile/list.go`](cli/cmd/ctl/profile/list.go)):

```
   NAME             OLARES-ID              STATUS
*  alice            alice@olares.com       logged-in (23h59m)
   bob              bob@olares.com         expired
   eve              eve@olares.com         invalidated
   frank            frank@olares.com       never
```

| STATUS | Meaning | Recovery |
|--------|---------|----------|
| `logged-in (Xh Ym)` | Token is valid; column shows time-to-expiry | — |
| `logged-in` | Token is present but its JWT has no exp claim, so we can't verify locally | Trust until the server says no |
| `expired` | Token JWT exp is in the past | `profile login` to re-authenticate |
| `invalidated` | The server explicitly rejected the refresh leg (`/api/refresh` returned 401/403) | `profile login` directly — no need to `profile remove` first |
| `never` | No token has ever been stored for this profile | `profile login` or `profile import` |

The leading `*` marks the current profile.

### `profile use <name|->`

```bash
olares-cli profile use alice            # by NAME alias
olares-cli profile use alice@olares.com # by olaresId (also accepted)
olares-cli profile use -                # back to the previous current (errors when none)
```

Updates `currentProfile` and `previousProfile` inside `~/.config/olares-cli/config.json`.

### `profile remove <name>`

```bash
olares-cli profile remove alice
```

Performs four actions atomically:

1. Removes the profile entry from `config.json`.
2. Deletes the stored token for that olaresId from the keychain.
3. If the removed profile was current, current falls back to the previous (when valid) or to the first remaining profile.
4. If the removed profile was the last one, the keychain namespace itself is purged so no orphan entries remain in Keychain Access.app / regedit / etc.

## Token storage

| OS | Backend | Location |
|------|---------|----------|
| darwin | macOS Keychain | service `olares-cli`, account = olaresId |
| linux | AES-256-GCM file | under `~/.local/share/olares-cli/` |
| windows | DPAPI | `HKCU\Software\OlaresCli\keychain` |

**The plaintext `~/.olares-cli/tokens.json` from older builds is deprecated** — tokens written there by previous versions are no longer read. If the user upgraded and suddenly appears "logged out", the correct fix is `profile login` to repopulate the new storage.

After `login` / `import` succeeds, the CLI prints a line like `token stored via <backend> (service "olares-cli", account "<id>")`. That message is the source of truth for "where did my token actually land". If the backend resolves to `file-fallback` (sandboxed / CI environments without access to a system keychain), be aware: that token is now sitting in a file with **different security properties than the system keychain**.

## Re-authentication rules (critical)

`profile login` and `profile import` both apply the same logic per olaresId:

```
                    ┌───────────────────────────────────────┐
profile not exist ──▶│ Auto-create and write the new token   │
                    └───────────────────────────────────────┘
                    ┌──────────────────────────────────────────────┐
token expired     ──▶│ Reuse the profile entry, write a new token   │
                    └──────────────────────────────────────────────┘
                    ┌──────────────────────────────────────────────┐
token invalidated ──▶│ Reuse the profile entry, write a new token   │
                    └──────────────────────────────────────────────┘
                    ┌──────────────────────────────────────────────────────────┐
token still valid ──▶│ Reject; tell the user to run profile remove <id> first   │
                    └──────────────────────────────────────────────────────────┘
```

Logic lives in [`cmd/ctl/profile/credentials.go`](cli/cmd/ctl/profile/credentials.go) `ensureProfileWritable`. If a script needs unconditional overwrite, it MUST `profile remove` first and then `profile login` / `profile import`.

## Automatic token refresh

**The CLI rotates expired access_tokens transparently.** Users do NOT need to run `profile login` just because their access_token aged out — only when the *refresh_token* itself becomes invalid.

The refresh logic lives in [`cli/pkg/cmdutil/factory.go`](cli/pkg/cmdutil/factory.go) (`refreshingTransport`) and [`cli/pkg/credential/refresher.go`](cli/pkg/credential/refresher.go). Every `*http.Client` the Factory hands out has it wired in.

### Two trigger paths

| Trigger | Applies to | Behavior |
|---------|------------|----------|
| **Reactive (401/403 + retry)** | Replayable bodies — every JSON / `files cat` / `files download` / `files rm` / `market` JSON verb | Send with current token. On 401/403, `/api/refresh`, retry the same request once with the new token. One extra round-trip in the rare expiry case; zero overhead in steady state. |
| **Pro-active (JWT exp + skew)** | Non-replayable bodies — `files upload` chunks (`*os.File`) | Decode the access_token's JWT exp before sending. If within 60s of expiry (or already past), `/api/refresh` first, send with the new token. Required because once a streaming body is consumed by the first send we can't replay it on a 401. |

The pro-active skew is hardcoded at 60s in `cli/pkg/cmdutil/factory.go` (`preflightSkew`) — comfortably absorbs client↔server clock drift plus the time from local decode to the request landing on the server. Tokens issued without an `exp` claim, or values that don't decode as a JWT at all, skip the pre-flight gracefully and fall back to the reactive path.

### Concurrency

Across goroutines AND across concurrent `olares-cli` processes, **`/api/refresh` is hit at most once per stale token**:

1. Process-wide `sync.Mutex` — losers wait, then read whatever the winner persisted.
2. Compare-after-Get against the keychain — short-circuits when a sibling already rotated the token.
3. On-disk `flock` under `<config-dir>/locks/<sanitized-olaresId>.refresh.lock` — serializes across processes; bounded by a 30s acquire timeout so a stuck peer can't hang the CLI.
4. Re-check inside the flock — collapses any final race that snuck in between the in-process and on-disk locks.

For most users this is invisible. It matters when you script multiple `olares-cli` invocations in parallel: they will not stampede `/api/refresh`.

### When refresh itself fails

| Outcome | What the user sees | Fix |
|---------|---------------------|-----|
| `/api/refresh` returns 200 + new tokens | (silent — request retried, command succeeds) | — |
| `/api/refresh` returns 401/403 (refresh_token revoked / expired / rotated by another login) | `refresh token for <id> became invalid at <ts>; please run: olares-cli profile login --olares-id <id>` (typed `*ErrTokenInvalidated`) | `olares-cli profile login --olares-id <id>` |
| No token in keychain at all | `no access token for <id>; run: olares-cli profile login --olares-id <id>` (typed `*ErrNotLoggedIn`) | `olares-cli profile login` (or `profile import`) |
| `/api/refresh` returns 5xx / network error | Surfaced verbatim from the transport | Retry the command — the grant itself is still valid |

The keychain entry is stamped `InvalidatedAt` on the 401 path so subsequent commands skip the network round-trip and go straight to the CTA. `profile list` shows these as `invalidated`.

> **Do not implement custom retry/backoff loops on top of these errors.** The transport already handles the recoverable cases; once you see `ErrTokenInvalidated` or `ErrNotLoggedIn`, only `profile login` / `profile import` will help.

## Auth error recovery table

| Error message (excerpt) | Meaning | Fix |
|-------------------------|---------|-----|
| `refresh token for <id> became invalid at <ts>` | `/api/refresh` itself returned 401/403 — the grant is dead | `olares-cli profile login --olares-id <id>` |
| `no access token for <id>` | Profile selected but keychain has no entry | `olares-cli profile login` or `profile import` |
| `server rejected the access token (HTTP 401)` / `(HTTP 403)` | After auto-refresh the server still rejects (rare; usually a server-side state drift) | `olares-cli profile login --olares-id <id>` |
| `--olares-id is required` | login / import was invoked without olaresId | Add `--olares-id <id>` |
| `already authenticated for <id> (expires in ...)` | A still-valid token exists for this olaresId | `olares-cli profile remove <id>` and re-run login / import |
| `a token is already stored for <id> but its expiry can't be determined client-side` | Token present but JWT carries no exp claim, so we conservatively reject | Same: `profile remove <id>` and re-run |
| `two-factor authentication required: re-run with --totp <code>` | 2FA is on and we're in a non-TTY context (no way to prompt) | Re-run with `--totp <code>`, or run interactively in a TTY |
| `password is empty` / `TOTP code is empty` | stdin / TTY returned an empty string | Check for premature EOF or an empty pipe |
| `profile <name> not found` | `profile use` / `profile remove` referenced an unknown profile | `profile list` to see the actual names |

> **Do not silently retry auth errors.** 401/403 after auto-refresh and `already authenticated` are deterministic — follow the table; blind retries make the situation worse.

## dev / internal flags

For internal debugging or self-hosted dev environments only — **never include these in user-facing examples or scripts**:

| Flag | Use |
|------|-----|
| `--auth-url-override <url>` | Hard-pin the Authelia URL instead of deriving it from olaresId |
| `--local-url-prefix <label>` | Inject an extra label between the auth subdomain and the terminus name |
| `--insecure-skip-verify` | Disable TLS verification (only for self-signed local environments) |

## Security rules

- **Never** invent a `--password <plaintext>` argument (it does not exist). Passwords go through the TTY or `--password-stdin` fed by a secret pipe.
- **Never** echo `access_token` / `refresh_token` to the terminal. When passing a refresh_token to `profile import`, source it from an environment variable or external secret store: `--refresh-token "$OLARES_REFRESH_TOKEN"`.
- **Confirm intent before write/delete actions** (`profile remove`, `files rm`, `files upload --overwrite`, ...). Do not act unilaterally on the user's behalf.
- **TOTP is not a password** — it is single-use and short-lived, so the CLI deliberately echoes it to make manual entry less error-prone (matching `gh auth login`, `aws sso login`, kubectl OIDC plugins). That said, never persist a TOTP in a shared script.
