---
name: olares-shared
version: 4.6.0
description: "Foundation for the olares-cli skill suite: choose the right domain skill, understand the Olares platform model each task depends on, and decide whether the active profile can proceed or needs login recovery. Use first for runtime skills, and directly for Olares ID, profile, login, 2FA/TOTP, refresh token, keychain, auth errors, or uncertainty about which Olares skill owns a task."
compatibility: Requires olares-cli on PATH
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# olares-cli shared rules

Read this thin front door before a runtime skill. It supplies suite routing, the active-profile model, platform entry points and the auth proceed/stop rule. Load detailed references only when the current task triggers them.

## Platform entry points

- [Olares platform model](references/olares-platform.md): userspace storage, uid/gid 1000, protected Home directories, app/namespace networking, system middleware and Olares version semantics. Read it for files, chart, cluster or settings tasks that touch those concepts.
- [Application state machine](references/olares-platform-appstate.md): lifecycle transitions, allowed operations, backend timeouts, serialized downloads, `running` semantics and unreliable progress. Read it for market operations or runtime diagnosis.
- [Profile and authentication](references/olares-auth.md): login/import flows, profile statuses, token storage and refresh behavior. Read it only for profile work or auth recovery.

## Skill suite map

| Skill | Use it for |
|---|---|
| [`olares-shared`](SKILL.md) | Suite routing, platform entry points, profile/auth decisions |
| [`olares-market`](../olares-market/SKILL.md) | Install and manage catalog/uploaded apps; lifecycle and chart transfer |
| [`olares-settings`](../olares-settings/SKILL.md) | Post-install app/system configuration, users, VPN, backup and integrations |
| [`olares-cluster`](../olares-cluster/SKILL.md) | Runtime objects, logs, jobs, namespaces, nodes and middleware |
| [`olares-dashboard`](../olares-dashboard/SKILL.md) | CPU, memory, disk, network, pod, GPU and fan metrics |
| [`olares-files`](../olares-files/SKILL.md) | Browse and modify Drive, Sync, cache, external and cloud files |
| [`olares-knowledge`](../olares-knowledge/SKILL.md) | URL, yt-dlp, aria2, torrent and Hugging Face download tasks |
| [`olares-search`](../olares-search/SKILL.md) | Full-content file search and installed-app title search |
| [`olares-router`](../olares-router/SKILL.md) | Configure, install, call and diagnose models through Router |
| [`olares-chart`](../olares-chart/SKILL.md) | Author, validate and deploy an app's Olares chart |
| [`olares-publish`](../olares-publish/SKILL.md) | Prepare and submit a public Olares Market listing |
| [`olares-doctor`](../olares-doctor/SKILL.md) | Diagnose an app/system runtime failure and route the fix |

Porting and debugging an app commonly combines `chart` (author/fix), `market` (lifecycle), this skill's platform model and `doctor` (root-cause diagnosis).

Host installation, node joining, OS upgrades and GPU drivers use the kubeconfig-backed `olares-cli node` / `os` / `gpu` trees, not this profile-backed skill suite.

## Active profile

One profile selects one Olares instance and one Olares identity. Every profile-backed command targets that selection; there is no per-command `--profile` override.

| Command | Purpose |
|---|---|
| `olares-cli profile list` | Show profiles, current selection, auth status and cached Olares version |
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
