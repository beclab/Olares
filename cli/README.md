# Olares CLI

[![License: AGPL-3.0-or-later](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8.svg)](https://go.dev)
[![npm @olares/cli](https://img.shields.io/npm/v/@olares/cli?label=npm%20%40olares%2Fcli)](https://www.npmjs.com/package/@olares/cli)

`olares-cli` is the official CLI for installing and operating [Olares](https://olares.com) — an AI-native, self-hosted personal cloud. It is a single static Go binary that drives every part of the Olares product (OS bootstrap, app market, file storage, dashboard, settings, cluster, profile auth), and is designed to be driven equally well by humans typing commands and by AI coding agents reading [`SKILL.md`](#agent-skills) bundles.

The product surface this CLI mirrors:

- **Olares OS** — Kubernetes-based personal cloud you install on a Linux host
- **Olares ID** — your identity within Olares (`<name>@olares.com`)
- **Olares Dashboard / ControlHub / Files / Market / Settings** — the SPAs you can also drive from the CLI

## Why olares-cli?

- **Agent-native** — every command tree ships a maintained `SKILL.md`, compiled into the binary, so any agent (Cursor, Claude Code, Codex, OpenClaw, ...) can discover verbs, flags, and recovery flows. `olares-cli skills install` writes them out; `skills list` / `skills read` serve them without touching disk
- **Wide coverage** — Olares OS install/upgrade plus the full identity, files, market, dashboard, settings, and ControlHub surface from one binary
- **Olares-native auth** — refresh tokens live in the OS keychain; access tokens auto-refresh on 401/403; profile model maps cleanly to multi-instance / multi-ID setups
- **Distributed two ways** — persistent `npm install -g @olares/cli` client on macOS / Windows / Linux; or zero-install `npx @olares/cli <verb>` for one-offs. A first-run wizard, `npx @olares/cli install`, does both `npm install -g` and skill installation in one shot.

## Install — which path fits you?

| Your situation | Use this | Why |
| --- | --- | --- |
| I'm already on an Olares host and want to use its CLI | Use `/usr/local/bin/olares-cli` directly (and see ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) if you need the agent verbs the OS bundle doesn't ship) | It's the OS-bundled copy, kept in sync via `olares-cli upgrade`. |
| I'm a first-time user and want one command to set up the CLI + skills | `npx @olares/cli@latest install` <br>([Scenario A](#first-run-wizard-scenario-a-recommended)) | Runs `npm install -g` and `olares-cli skills install` for you. Does not install Olares OS. |
| I want to control a remote Olares from my dev box, CI, macOS, or Windows | `npm install -g @olares/cli@latest` <br>([Scenario B](#client-on-a-non-olares-machine-scenario-b)) | Persistent install; `olares-cli` on PATH. |
| I just want to run one command quickly without installing | `npx @olares/cli@latest <verb>` <br>([Scenario C](#one-off-ops-scenario-c)) | No persistent files; ~1-2 s cold-start per invocation. Keychain/token caches persist per-user. |

### First-run wizard (Scenario A, recommended)

```bash
npx @olares/cli@latest install
```

The `install` verb is handled entirely by the Node shim, never by the Go binary. It runs two steps for you:

1. `npm install -g @olares/cli` (or upgrade if you already have it).
2. `olares-cli skills install` to write out the twelve `olares-*` agent skills the binary carries.

After it prints `You are all set!`, do the auth step yourself — it's interactive and ties to your Olares ID:

```bash
olares-cli profile login --olares-id <your-olares-id>
olares-cli profile current
```

> **What this command does NOT do:** install Olares OS. The Linux host bootstrap stays `curl -fsSL https://olares.sh | bash` (see [docs/manual/get-started](https://docs.olares.com/manual/get-started/install-olares/linux.html)). It also does not configure any app credentials — Olares uses the Olares ID directly, so no `config init` step is needed.
>
> **On a Linux host with an existing `olares-cli` in `/usr/local/bin` or `/usr/bin`:** the wizard reads its `--version` and decides keep-vs-replace.
> - **Release-grade** (stable `1.12.7`, or pre-releases `-rc1` / `-beta.1` / `-alpha2`) → kept; if `npm config get prefix` resolves to the same `bin` directory (typical Olares host: `/usr/local`), the wizard short-circuits the `npm install -g` attempt — no full install timeout — and exits with a side-by-side install block (`npm install -g ... --prefix=$HOME/.olares-cli-npm` + `PATH` export + `olares-cli skills install`) you can copy verbatim. The OS bundle is canonical for system-layer verbs and only `olares-cli upgrade` is supposed to replace it. See ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) for the long-form reference.
> - **Dev / test / dirty** (the Makefile placeholder `0.0.0-development`, `git describe` outputs like `1.12.7-3-gabc1234-dirty`, check.yaml's `1.12.7-12345678` PR builds, unparseable output) → removed so npm can install over the same path. If removal needs root, the wizard exits with a one-line sudo hint instead of silently failing.
>
> **Permission errors on Linux** (`EACCES` while npm writes to `/usr/lib/node_modules` or `/usr/local/lib/node_modules`): typical for distro-packaged Node (`apt install nodejs`) where the global prefix is root-owned. The wizard surfaces the offending npm `stderr` plus a one-time fix that switches npm to a user-owned prefix (`npm config set prefix ~/.npm-global` + `PATH`) so global installs no longer need `sudo`.

### Local dev (install your build + local skills)

For working on the CLI itself. Requires Go 1.24+ and a full repo clone — [`cli/go.mod`](go.mod) has a relative `replace` into `../framework/oac`, so a `cli/`-only clone won't build.

```bash
cd cli
sudo make install          # or: make install PREFIX="$HOME/.local"
olares-cli skills install  # write out the skills the binary you just built carries
make uninstall             # remove the binary again
```

There is no separate step for the skills: `make install` compiles `cli/skills/` into the binary, so installing them writes your checkout's copy, edits included. Notes:

- **Never run step 2 with `sudo`.** It creates root-owned `~/.agents/skills/olares-*` directories that later break reinstalls with `EACCES`. Only step 1 (`make install` into a system prefix) may need `sudo`.
- On Windows, step 1 needs Git Bash / MSYS (for `make` and `install`); step 2 runs natively and falls back to copying where symlinks need privileges.

**Editing skills without rebuilding.** Point each agent at the checkout by hand, so a saved `SKILL.md` is what the agent reads on its next turn:

```bash
ln -s "$(pwd)/skills/olares-shared" ~/.agents/skills/olares-shared   # repeat per skill
```

`skills install` refuses to write over links like these and names them, because repointing them at the store would silently end the live-editing loop. Remove them, or pass `--force`, when you want the installed copy back.

### Client on a non-Olares machine (Scenario B)

```bash
# macOS / Windows / Linux dev box that talks to a remote Olares.
npm install -g @olares/cli@latest

# The package's only PATH-exposed bin is `olares-cli`, managed by npm itself.
# If an existing `olares-cli` is already at npm's target path (i.e. you're on
# a Linux Olares host where the OS bundle owns /usr/local/bin/olares-cli), npm
# refuses the install with EEXIST — your existing binary is never overwritten.
# See "On a Linux Olares host" below for the side-by-side workaround.

olares-cli profile login <your-olares-id>
olares-cli files ls /drive/Home
```

### One-off ops (Scenario C)

```bash
# No persistent install; each invocation re-uses the npm cache. The OS keychain
# persists across invocations, so log in once — subsequent commands re-use it.
npx @olares/cli@latest profile login <your-olares-id>
npx @olares/cli@latest profile current
npx @olares/cli@latest files ls /drive/Home
```

### Capabilities & limits of each install method

| Method | What it gives you | Binary path | What it can't / won't do |
| --- | --- | --- | --- |
| **A — `npx @olares/cli@latest install`** | Superset of B: runs `npm install -g @olares/cli` (or upgrade) and `olares-cli skills install` in one shot. End state is the same as B, plus the twelve `olares-*` skills pre-installed. | Same as B once the wizard finishes. The wizard itself is the Node shim from the npx cache. | Does not install Olares OS (still `curl -fsSL https://olares.sh \| bash`). Does not run `profile login` for you (interactive + needs your Olares ID). On a Linux host with an existing `/usr/local/bin/olares-cli` or `/usr/bin/olares-cli`, the wizard reads its `--version`: release-grade copies (stable / `rc` / `beta` / `alpha`) are kept and npm prints the [`--prefix` / `npx`](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) workarounds; dev / test / dirty / `0.0.0-development` builds are removed so the npm copy can take over (re-run with `sudo` if the wizard exits with a permission hint). |
| **B — `npm install -g @olares/cli@latest`** | Persistent `olares-cli` CLI on PATH (macOS / Windows / Linux). Use to talk to a *remote* Olares (login, files, market, dashboard, cluster, settings). | `<npm prefix>/bin/olares-cli` (symlink managed by npm itself). On a Linux Olares host where `/usr/local/bin/olares-cli` already exists, npm aborts with `EEXIST` — see ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) for the side-by-side workaround. | The npm wrapper auto-sets `OLARES_CLI_REMOTE_ONLY=1`, so host-side verbs (`uninstall`, `upgrade`, `node`, `os`, `gpu`, `disk`, `wizard`, `user`, `osinfo`, `amdgpu`) are hidden from `--help` and return `unknown command`. The `install` verb is intercepted by the Node shim and runs the Scenario A wizard. All of these are reachable only on an Olares host through the OS-bundled `/usr/local/bin/olares-cli`. |
| **C — `npx @olares/cli@latest <verb>`** | Zero-install, runs any *remote/identity* verb without touching PATH. Great for CI one-shots, ephemeral containers, "just try it". | `~/.npm/_npx/<hash>/.../vendor/olares-cli` (only during the npx subprocess; cleared after) | Same host-side-verbs restriction as B (same Node shim). Each invocation pays a ~1-2 s npx cold-start. Long watches (`olares-cli market list --watch`, `olares-cli cluster pod logs -f`) work but pay the cost up front. Keychain/token caches persist across npx invocations. |

### Install AI agent skills

The Scenario A wizard runs this for you. If you went the Scenario B / C route, run it manually:

```bash
olares-cli skills install
```

Writes the twelve `olares-*` skills to `~/.agents/skills` and links them from each agent skills directory that already exists (Cursor, Claude Code, Codex, OpenClaw, ...). Directories are never created for an agent that isn't installed here, and the bytes come from the binary — no GitHub, no Node, nothing that can fetch a version disagreeing with the verbs you have. Run it again after upgrading; the CLI says so on startup when the installed copy came from a different release.

```bash
olares-cli skills list                       # what this binary carries
olares-cli skills read olares-shared         # serve one without writing anything
olares-cli skills export ./skills            # write them somewhere specific
```

`skills` is available on both channels, including through `npx @olares/cli`.

> **For AI agents:** the human must run `olares-cli profile login <their-olares-id>` themselves — auth opens a browser. Verify with `olares-cli profile current` + `olares-cli dashboard overview`. Load `olares-shared` first; it documents the auth model for the other skills.

## Agent skills

Each skill ships a single `SKILL.md` plus a `references/` folder, all loaded on demand by `olares-shared`-aware agents. All twelve are compiled into the binary, and they carry its release as their version, so a skill and the verbs it documents cannot disagree.

| Skill | Surface | Use when the user mentions... |
| --- | --- | --- |
| [`olares-shared`](skills/olares-shared/SKILL.md) | Suite routing, platform model, profile auth and login recovery | profile, login, logout, 2FA / TOTP, keychain, auth errors, which skill owns a task |
| [`olares-files`](skills/olares-files/SKILL.md) | `olares-cli files` — read/write files in the Olares Files SPA | files, drive, home, upload, download, share, SMB / NFS, archives, Seafile sync |
| [`olares-market`](skills/olares-market/SKILL.md) | `olares-cli market` — install / upgrade / list apps | market, apps, install app, upgrade app, clone app, charts |
| [`olares-settings`](skills/olares-settings/SKILL.md) | `olares-cli settings` — read & mutate the Olares Settings SPA | settings, account, users, VPN, network, backup, GPU, integrations |
| [`olares-dashboard`](skills/olares-dashboard/SKILL.md) | `olares-cli dashboard` — Overview / Apps / GPU views | dashboard, overview, resource usage, rankings, fan |
| [`olares-cluster`](skills/olares-cluster/SKILL.md) | `olares-cli cluster` — Olares ControlHub Kubernetes view | ControlHub, cluster, pods, workloads, namespaces, nodes, logs, exec |
| [`olares-doctor`](skills/olares-doctor/SKILL.md) | Runtime diagnosis across the trees above | won't install, won't start, crash loop, image pull, running but unreachable, slow |
| [`olares-router`](skills/olares-router/SKILL.md) | `olares-cli router` — Router and the model applications behind it | models, Router, AI gateway, API key for a model, model quota, chat / embed / transcribe / OCR |
| [`olares-knowledge`](skills/olares-knowledge/SKILL.md) | `olares-cli knowledge` — download orchestration and Wise tasks | download a URL, yt-dlp, aria2, torrent, Hugging Face, Wise |
| [`olares-search`](skills/olares-search/SKILL.md) | `olares-cli search` — Desktop global search | search files, find a document, search installed apps, Sync / Google Drive / Dropbox |
| [`olares-chart`](skills/olares-chart/SKILL.md) | `olares-cli chart` — author, validate and deploy an app chart | port an app, docker-compose, Helm chart, OlaresManifest, storage / entrance / GPU wiring |
| [`olares-publish`](skills/olares-publish/SKILL.md) | Public Olares Market submission | publish an app, Market listing, icon and screenshots, release targets |

Skills are also published on [ClawHub](https://clawhub.io) (search "olares"), which is where an agent can get them on a machine with no `olares-cli` yet. The binary is the primary channel: a registry copy is whatever was last pushed there, while `skills install` writes the copy that matches the binary running it.

## Three-layer command system

```
olares-cli <area> [<noun>] <verb> [flags]
```

- **System layer** (root-level, no `<area>` prefix): `install`, `uninstall`, `upgrade`, `start`, `stop`, `status`, `backup`, `precheck`, `prepare`, `download`, `change-ip`, `release`, `printinfo`, `logs`, `node`, `gpu`, `amdgpu`, `disk`, `osinfo`, `wizard`. These manage the host running Olares OS itself and require root / kubeconfig access — they are not driven by an Olares ID. *Channel availability*: the Go binary only registers them when `OLARES_CLI_REMOTE_ONLY` is unset, i.e. only when invoked from an Olares host's OS-bundled `/usr/local/bin/olares-cli`. Through `npm install -g @olares/cli` or `npx @olares/cli`, the Node shim sets `OLARES_CLI_REMOTE_ONLY=1` and they are hidden. The lone exception is `install`, which the Node shim itself intercepts and routes to the [first-run wizard](#first-run-wizard-scenario-a-recommended) — it never reaches the Go binary on the npm channel.
- **Identity-bound layer** (`<area>` = `profile` / `files` / `market` / `settings` / `dashboard` / `cluster` / `doctor` / `router`): act on behalf of the currently-selected Olares ID against a running Olares HTTP API. Pick the identity once with `olares-cli profile use <name>`, then every verb in this layer uses it. Reachable through both `npm install -g` and `npx`.
- **Local layer** (`<area>` = `skills` / `chart`): touch this machine's files only — no Olares instance, no identity, no network. `skills` reads the suite compiled into the binary and writes it out; `chart` authors and validates a chart before anything is deployed. Reachable on every channel, and `skills` needs no profile even to `list`.

For every command, `--help` is the source of truth for flags and wire shapes:

```bash
olares-cli --help
olares-cli files --help
olares-cli files ls --help
```

Examples for workload image inspection:

```bash
olares-cli cluster workload list --limit 20 --page 1 -o json
olares-cli cluster workload images --limit 50 --page 1
olares-cli cluster workload images docker.io/library/nginx:latest
olares-cli doctor images -o json
# Orphans only, biggest first, with reclaimable-size footer.
olares-cli doctor images --unused
# Per-user third_level_domain duplicates / reserved names (kubeconfig).
olares-cli doctor thirdleveldomain
olares-cli doctor thirdleveldomain --force-dedupe
```

## Output formats

Most identity-layer verbs accept `--output table` (default, human-readable) and `--output json` (machine-readable). Use `--output json` whenever a script or agent needs to parse the result; the JSON schema is intentionally stable across minor versions.

```bash
olares-cli files ls /drive/Home --output json
olares-cli market list --output json | jq '.items[] | {name, version, status}'
```

## Uninstall

Pick the reverse operation that matches how you installed.

### Remove the CLI client

```bash
npm uninstall -g @olares/cli
# npm cleans the `olares-cli` symlink and the package files itself —
# there is no extra cleanup step.
```

### Clear the npx cache

```bash
# npx auto-evicts the cache after a few days. To force-clear sooner:
rm -rf ~/.npm/_npx/                       # nukes all npx-cached packages
# Or, more surgical:
ls ~/.npm/_npx/                            # find the hash dir for @olares/cli
rm -rf ~/.npm/_npx/<hash>/
```

### Remove agent skills

```bash
rm -rf ~/.agents/skills/olares-*          # the store; agents' links go dangling
rm ~/.claude/skills/olares-*              # and the links themselves, per agent
```

### Wipe stored credentials

```bash
olares-cli profile list                    # see what's stored
olares-cli profile remove <name>           # delete one profile + its keychain token
```

Credentials live in the OS-native keychain (macOS Keychain / Windows DPAPI / Linux secret-service or filesystem fallback at `~/.olares/credentials/`). `profile remove` is always the right cleaning verb — don't hand-edit those files.

## On a Linux Olares host: install side-by-side with the OS bundle

This section applies **only to Linux hosts that have Olares OS installed** (where `/usr/local/bin/olares-cli` already exists). macOS / Windows / non-Olares Linux dev boxes never hit this scenario.

### Why you might want this

The OS-bundled `olares-cli` is pinned to the version that shipped with your Olares OS release (e.g. **1.12.5**). Older bundles do **not** include the agent / identity verbs (`profile`, `files`, `market`, `dashboard`, `settings`, `cluster`) — those land in newer npm releases first. If you want to drive a remote (or your own) Olares from the same Linux host, install the latest npm copy alongside the OS bundle. Two ways:

### Option 1 — Install under a separate prefix

A plain `npm install -g @olares/cli` aborts with `EEXIST` because `/usr/local/bin/olares-cli` already exists. Use a separate prefix to coexist:

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # PATH order decides which copy wins
olares-cli --version                            # now resolves to the npm copy
# Revert: reorder PATH or `rm -rf ~/.olares-cli-npm/`
```

Both binaries are then on disk:

- `/usr/local/bin/olares-cli` — OS bundle. System layer (`install`, `uninstall`, `upgrade`, `start`, `stop`, ...).
- `~/.olares-cli-npm/bin/olares-cli` — npm copy. Identity layer (`profile`, `files`, `market`, `dashboard`, `settings`, `cluster`); system layer is hidden by `OLARES_CLI_REMOTE_ONLY=1`.

### Option 2 — Use `npx` for one-offs

No persistent install, no PATH gymnastics:

```bash
npx @olares/cli@latest profile current
npx @olares/cli@latest files ls /drive/Home
```

> Do **not** `npm install -g @olares/cli --force` on an Olares host — that would clobber the OS-managed `/usr/local/bin/olares-cli`. The OS bundle is canonical for system-layer verbs on that host and is upgraded via `olares-cli upgrade`. Without `--force`, npm already aborts safely with `EEXIST`.

> The Scenario A wizard automates this safety net: it reads the existing binary's `--version` and only refuses to replace it when the version is release-grade (stable / `rc` / `beta` / `alpha`). If you intentionally placed a `make install` dev build at `/usr/local/bin/olares-cli`, the wizard will remove it and let npm install over the path — re-run with `sudo` if removal needs root.

## Build from source

Requires **Go 1.24+**.

```bash
cd cli
go build -o olares-cli ./cmd/main.go
./olares-cli --help
```

The npm package downloads pre-built binaries from GitHub Releases on `postinstall`; you only need a local Go toolchain if you're modifying the CLI itself.

## Repository layout

```
cli/
├── cmd/                  # CLI entrypoint and Cobra command tree
│   ├── main.go
│   └── ctl/              # one folder per top-level command (os, node, gpu, profile,
│                         #   market, files, dashboard, settings, cluster, ...)
├── pkg/                  # install engine + remote API clients
│   ├── core/             # pipeline / module / task / action framework
│   ├── pipelines/        # top-level pipelines invoked by install/start/upgrade/...
│   └── ...               # one package per concern (k3s, etcd, gpu, storage, terminus, ...)
├── internal/             # non-exported helpers (keychain, lockfile, files client, ...)
├── apis/                 # kubekey v1alpha2 CRD types
├── skills/               # AI-agent SKILL.md bundles, compiled into the binary by embed.go
├── npm/                  # @olares/cli npm wrapper (postinstall downloads the Go binary)
├── version/              # VERSION / VENDOR ldflag targets
├── .goreleaser.yaml
└── go.mod
```

The install engine in `pkg/core` runs a `Pipeline → Module → Task → Action` stack. Each mode-1 command moves the host between five lifecycle stages:

- **prechecked** — `olares-cli precheck` validates the environment against install requirements; gating step before any state-changing action.
- **downloaded** — `olares-cli download` (`component` / `wizard`) fetches the install assets; `olares-cli download check` verifies completeness.
- **prepared** — `olares-cli prepare` lays out dependencies.
- **installed** — `olares-cli install` brings up Kubernetes and Olares core; `olares-cli upgrade` moves an installed host to a newer version; `olares-cli start` / `stop` / `status` toggle the runtime; `olares-cli uninstall` rolls back (optionally to a specific phase); `olares-cli change-ip` repairs after an IP change.
- **activated** — `olares-cli wizard activate <olaresId>` enrols the first user against BFL/Auth, after which the profile-based commands become usable.

## Security & risks

- **Credentials** — refresh tokens are stored in the OS-native keychain (macOS Keychain / Windows DPAPI / Linux secret-service); access tokens are derived on demand and never persisted. `olares-cli profile remove` is the canonical way to wipe them.
- **Profile isolation** — there is no per-invocation `--profile` flag. Identity is single-source via `olares-cli profile use <name>`; agents and scripts must commit to one identity up front rather than silently hopping mid-pipeline.
- **`--yes` contract** — every mutating verb on the identity layer (delete / restart / scale / install / upgrade) prompts for confirmation by default. `--yes` is the agreed-on bypass; treat it as a safety check, not a style preference.
- **`metadata.requires.bins` is advisory** — skills declare `["olares-cli"]` as a host requirement so agents can warn when the binary is missing, but skill discovery does *not* auto-install the CLI. Install it explicitly via one of the methods above.
- **Code signing** — on macOS / Windows the npm-downloaded binary is currently unsigned; Gatekeeper or SmartScreen may warn on first run. Verify the download via `sha256sum` against the matching GitHub Release if you need to be sure.

## License

[AGPL-3.0-or-later](../LICENSE).
