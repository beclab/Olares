# Olares profile and authentication

Read this reference when setting up or switching an Olares identity, handling an authentication error, or explaining token storage and refresh. Routine business commands only need the proceed/stop gate in [`../SKILL.md`](../SKILL.md).

## Profile model

One profile is one Olares instance plus one user identity, keyed by **olaresId** such as `alice@olares.com`. Each profile owns its access/refresh token pair. The selected profile determines the target and identity for every profile-backed command; there is no per-invocation `--profile` override.

Use `olares-cli profile --help` for flags. The profile verbs are:

| Command | Purpose |
|---|---|
| `profile login` | Authenticate with password and optional TOTP; create the profile when absent |
| `profile import` | Bootstrap from an existing refresh token |
| `profile list` | List profiles, local auth status and cached Olares version |
| `profile use <name\|->` | Select a profile; `-` returns to the previous one |
| `profile remove <name>` | Remove the profile and its stored token |

There is no `auth login` or `auth logout` namespace. “Logout” is `profile remove`.

## Login with password

```bash
olares-cli profile login --olares-id <olaresId>
```

Interactive login prompts for the password without echo and then prompts for TOTP when 2FA is enabled. For scripted input, use `--password-stdin`; when 2FA is enabled, also pass the short-lived `--totp <code>` because there is no second prompt.

When driving login for a user, do not request or place a password in command arguments. Start the interactive command so it waits at the password prompt, forward that prompt to the user, then read the result after the process exits. Never log in on the user's behalf unless they asked.

TOTP is short-lived, not reusable configuration. Never persist it in a shared script, file or task record.

## Import an existing refresh token

```bash
olares-cli profile import \
  --olares-id <olaresId> \
  --refresh-token "$OLARES_REFRESH_TOKEN"
```

Source the token from an environment variable or secret manager. Never paste it into a command literal or echo it to the terminal.

## Inspecting profile status

`profile list` reports what the local token store can prove without a network call:

| Status | Meaning | Agent action |
|---|---|---|
| `logged-in` | Access token is locally usable, or has no parseable expiry | Proceed |
| `expired` | Access token expiry is in the past | Proceed; the next command normally refreshes it |
| `invalidated` | The server rejected the refresh grant | Stop and run `profile login` or `profile import` |
| `never` | No token has been stored | Stop and run `profile login` or `profile import` |
| `unknown` / `logged-in (unparseable token)` | Token storage or JWT parsing could not establish status | Run the command; re-login if the typed auth failure persists |

The `VERSION` column is the cached Olares backend version. `profile list --refresh-version` refreshes it. The leading `*` marks the selected profile.

## Re-authentication

`profile login` and `profile import` replace expired, invalidated or absent credentials in place. A still-valid token is protected from accidental overwrite; follow the CLI's instruction to remove it first only when intentionally replacing that identity.

Older builds stored plaintext credentials in `~/.olares-cli/tokens.json`. That store is deprecated. If an upgraded installation appears logged out, authenticate again rather than copying the old file.

## Token storage

| OS | Backend |
|---|---|
| macOS | Keychain, service `olares-cli`, account = olaresId |
| Linux | AES-256-GCM file under the Olares CLI data directory |
| Windows | DPAPI-protected value under the current user's registry hive |

After login/import, the CLI reports which backend stored the token. A `file-fallback` backend has different security properties from an OS keychain and should be treated accordingly.

## Automatic refresh

The CLI transparently rotates expired access tokens. Replayable requests refresh and retry once after an authentication rejection. Streaming uploads refresh before sending a chunk when the token is close to expiry because a consumed file stream cannot be replayed safely.

Refresh is coordinated across goroutines and concurrent CLI processes, so agents must not add their own authentication retry loop. Follow an explicit login/import action from the CLI after a missing/invalidated credential or persistent 401/459 instead of retrying the business command. A 403 permission denial is not fixed by logging in again.

