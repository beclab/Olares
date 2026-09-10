---
name: olares-shared
version: 0.0.0-cli.0
description: "Foundation for the olares-cli skill suite: choose the right domain skill, understand the Olares platform model each task depends on, and decide whether the active profile can proceed or needs login recovery. Use first for runtime skills, and directly for Olares ID, profile, login, 2FA/TOTP, refresh token, keychain, auth errors, or uncertainty about which Olares skill owns a task."
compatibility: Requires olares-cli on PATH
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# olares-cli shared rules

Read this thin front door before a runtime skill. It supplies the active-profile model, platform entry points and the auth proceed/stop rule — what a task needs once it knows which skill it belongs to. Load detailed references only when the current task triggers them.

- [Suite map](references/olares-suite-map.md): which skill owns which task, and how to install `olares-cli` when it is not on PATH. Read it when the task has not been routed to a skill yet; a task that has been does not need it again.

## Reading the answer

Every profile-backed tree spells machine-readable output the same way: `-o json` (`--json` is the same request, kept for older scripts). A verb that fails under it answers on **stderr** with `{"error":{"code","message","retryable","action"}}`, and stderr carries nothing else, so it parses whole.

`.error.code` is worth branching on where it is set, and today that is the failures whose recovery differs from every other failure's: `auth_no_profile`, `auth_not_logged_in`, `auth_token_expired`, `auth_token_invalidated` and `timeout`, plus whatever Router reports from upstream. Everything else arrives as `unclassified` with the whole story in `.message`. So read the code first, and fall back to the message rather than assuming a code you did not get means something. `retryable` and `action` are present only when the failure knows them — their absence means unknown, not "no".

Some trees also carry their own result document (Market's lifecycle verbs report `.status` and `.finalState`); where they do, that is the one to read.

## Platform entry points

- [Olares platform model](references/olares-platform.md): userspace storage, uid/gid 1000, protected Home directories, app/namespace networking, system middleware and Olares version semantics. Read it for files, chart, cluster or settings tasks that touch those concepts.
- [Application state machine](references/olares-platform-appstate.md): lifecycle transitions, allowed operations, backend timeouts, serialized downloads, `running` semantics and unreliable progress. Read it for market operations or runtime diagnosis.
- [Profile and authentication](references/olares-auth.md): login/import flows, profile statuses, token storage and refresh behavior. Read it only for profile work or auth recovery.

## Active profile

One profile selects one Olares instance and one Olares identity. Every profile-backed command targets that selection; there is no per-command `--profile` override.

| Command | Purpose |
|---|---|
| `olares-cli profile list` | Show profiles, current selection, auth status and cached Olares version |
| `olares-cli profile whoami` | Identity and role of the selected profile; the admin/normal answer other skills gate on |
| `olares-cli profile use <name\|->` | Switch selection; `-` returns to the previous profile |
| `olares-cli profile login` | Authenticate with password and optional TOTP |
| `olares-cli profile import` | Bootstrap from a refresh token |
| `olares-cli profile remove <name>` | Remove a profile and its credential |

Use `olares-cli profile --help` for flags and [the auth reference](references/olares-auth.md) for login or recovery.

## Auth-readiness gate

Proceed by default:

- `logged-in` and `expired` proceed; an expired access token normally refreshes on the next request.
- `never` and `invalidated` stop for `profile login` or `profile import`.
- For `unknown` or an unparseable token, run the business command and react to its typed error instead of guessing.

Do not preflight every command. The CLI refreshes and retries an authentication rejection once. Stop for login when the CLI explicitly says the credential is absent/invalidated or prints a login action after a persistent 401/459. A 403 permission denial, network error or 5xx is not a login signal. Never build a retry loop around auth errors.

## Security and task scope

- **Never** place a password in command arguments. Use the interactive prompt or `--password-stdin`.
- **Never** print access or refresh tokens. Source imports from a secret environment variable or secret manager.
- Ask before login, credential replacement, an ambiguous target or an action outside the user's authorised task scope.
- Within an authorised chart deploy/debug task, install, upgrade, restart, uninstall and clean reinstall are normal loop steps; do not ask again for each command.
