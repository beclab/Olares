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

- **Agent-native** — every command tree ships a maintained `SKILL.md` so any agent (Cursor, Claude Code, Codex, OpenClaw, ...) can discover verbs, flags, and recovery flows
- **Wide coverage** — Olares OS install/upgrade plus the full identity, files, market, dashboard, settings, and ControlHub surface from one binary
- **Olares-native auth** — refresh tokens live in the OS keychain; access tokens auto-refresh on 401/403; profile model maps cleanly to multi-instance / multi-ID setups
- **Distributed two ways** — persistent `npm install -g @olares/cli` client on macOS / Windows / Linux; or zero-install `npx @olares/cli <verb>` for one-offs. A first-run wizard, `npx @olares/cli install`, does both `npm install -g` and skill installation in one shot.

## Install — which path fits you?

| Your situation | Use this | Why |
| --- | --- | --- |
| I'm already on an Olares host and want to use its CLI | Use `/usr/local/bin/olares-cli` directly (and see ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) if you need the agent verbs the OS bundle doesn't ship) | It's the OS-bundled copy, kept in sync via `olares-cli upgrade`. |
| I'm a first-time user and want one command to set up the CLI + skills | `npx @olares/cli@latest install` <br>([Scenario A](#first-run-wizard-scenario-a-recommended)) | Runs `npm install -g` and `skills add beclab/Olares` for you. Does not install Olares OS. |
| I want to control a remote Olares from my dev box, CI, macOS, or Windows | `npm install -g @olares/cli@latest` <br>([Scenario B](#client-on-a-non-olares-machine-scenario-b)) | Persistent install; `olares-cli` on PATH. |
| I just want to run one command quickly without installing | `npx @olares/cli@latest <verb>` <br>([Scenario C](#one-off-ops-scenario-c)) | No persistent files; ~1-2 s cold-start per invocation. Keychain/token caches persist per-user. |

### First-run wizard (Scenario A, recommended)

```bash
npx @olares/cli@latest install
```

The `install` verb is handled entirely by the Node shim, never by the Go binary. It runs two steps for you:

1. `npm install -g @olares/cli` (or upgrade if you already have it).
2. `npx skills add beclab/Olares -y -g` to install the six `olares-*` agent skills.

After it prints `You are all set!`, do the auth step yourself — it's interactive and ties to your Olares ID:

```bash
olares-cli profile login --olares-id <your-olares-id>
olares-cli profile current
```

> **What this command does NOT do:** install Olares OS. The Linux host bootstrap stays `curl -fsSL https://olares.sh | bash` (see [docs/manual/get-started](https://docs.olares.com/manual/get-started/install-olares/linux.html)). It also does not configure any app credentials — Olares uses the Olares ID directly, so no `config init` step is needed.
>
> **On a Linux Olares host (`/usr/local/bin/olares-cli` already present):** the wizard detects the resulting `EEXIST` and prints two safe workarounds (`--prefix` side-by-side install, or stay on `npx`). It will not overwrite the OS bundle. See ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle).

### Or build from source

For local development or pinning to a specific commit. Requires Go 1.24+ and the full Olares repo — the CLI's [`cli/go.mod`](go.mod) has a relative `replace` directive into `../framework/oac`, so a `cli/`-only clone won't build.

```bash
git clone https://github.com/beclab/Olares.git
cd Olares/cli

sudo make install                                    # → /usr/local/bin/olares-cli
sudo make install PREFIX=$HOME/.local                # or pick your own prefix
make uninstall                                       # remove the same binary

# Install the agent skills (Scenario A does this for you; from-source doesn't):
npx skills add beclab/Olares -y -g
```

The resulting binary reports `git describe --tags --always --dirty` as its version (e.g. `1.12.7-3-gabc1234-dirty`), distinguishable from official releases (stable / `rc` / `beta` / `alpha` semver) and from `npm install -g` copies.

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
| **A — `npx @olares/cli@latest install`** | Superset of B: runs `npm install -g @olares/cli` (or upgrade) and `npx skills add beclab/Olares -y -g` in one shot. End state is the same as B, plus the six `olares-*` skills pre-installed. | Same as B once the wizard finishes. The wizard itself is the Node shim from the npx cache. | Does not install Olares OS (still `curl -fsSL https://olares.sh \| bash`). Does not run `profile login` for you (interactive + needs your Olares ID). On a Linux Olares host with `/usr/local/bin/olares-cli` present, npm hits `EEXIST` — the wizard detects this and prints the [`--prefix` / `npx`](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) workarounds instead of failing silently. |
| **B — `npm install -g @olares/cli@latest`** | Persistent `olares-cli` CLI on PATH (macOS / Windows / Linux). Use to talk to a *remote* Olares (login, files, market, dashboard, cluster, settings). | `<npm prefix>/bin/olares-cli` (symlink managed by npm itself). On a Linux Olares host where `/usr/local/bin/olares-cli` already exists, npm aborts with `EEXIST` — see ["On a Linux Olares host"](#on-a-linux-olares-host-install-side-by-side-with-the-os-bundle) for the side-by-side workaround. | The npm wrapper auto-sets `OLARES_CLI_REMOTE_ONLY=1`, so host-side verbs (`uninstall`, `upgrade`, `node`, `os`, `gpu`, `disk`, `wizard`, `user`, `osinfo`, `amdgpu`) are hidden from `--help` and return `unknown command`. The `install` verb is intercepted by the Node shim and runs the Scenario A wizard. All of these are reachable only on an Olares host through the OS-bundled `/usr/local/bin/olares-cli`. |
| **C — `npx @olares/cli@latest <verb>`** | Zero-install, runs any *remote/identity* verb without touching PATH. Great for CI one-shots, ephemeral containers, "just try it". | `~/.npm/_npx/<hash>/.../vendor/olares-cli` (only during the npx subprocess; cleared after) | Same host-side-verbs restriction as B (same Node shim). Each invocation pays a ~1-2 s npx cold-start. Long watches (`olares-cli market list --watch`, `olares-cli cluster pod logs -f`) work but pay the cost up front. Keychain/token caches persist across npx invocations. |

### Install AI agent skills

The Scenario A wizard runs this for you. If you went the Scenario B / C route, run it manually:

```bash
npx skills add beclab/Olares -y -g
```

Installs the 6 `olares-*` skill bundles via [vercel-labs/skills](https://github.com/vercel-labs/skills) into your active agent (Cursor, Claude Code, Codex, OpenClaw, and ~50 others). The agent then auto-loads the right skill when you mention Olares-flavoured tasks.

> Note: the `@olares` npm scope and the `beclab` GitHub org are independent naming spaces. The canonical source repo is `github.com/beclab/Olares`, so the `skills add` command (which fetches via GitHub) uses `beclab/Olares`, while the npm package is published as `@olares/cli`.

> **For AI agents:** the human must run `olares-cli profile login <their-olares-id>` themselves — auth opens a browser. Verify with `olares-cli profile current` + `olares-cli dashboard overview`. Load `olares-shared` first; it documents the auth model for the other skills.

## Agent skills

Each skill ships a single `SKILL.md` plus a `references/` folder, all loaded on demand by `olares-shared`-aware agents.

| Skill | Surface | Use when the user mentions... |
| --- | --- | --- |
| [`olares-shared`](skills/olares-shared/SKILL.md) | Profile auth, login, refresh tokens, error recovery | profile, login, logout, 2FA / TOTP, keychain, auth errors |
| [`olares-files`](skills/olares-files/SKILL.md) | `olares-cli files` — read/write files in the Olares Files SPA | files, drive, home, upload, download, chown, cp, mv |
| [`olares-market`](skills/olares-market/SKILL.md) | `olares-cli market` — install / upgrade / list apps | market, apps, install app, upgrade app, charts |
| [`olares-settings`](skills/olares-settings/SKILL.md) | `olares-cli settings` — read & mutate the Olares Settings SPA | settings, account, GPU, app settings, integrations |
| [`olares-dashboard`](skills/olares-dashboard/SKILL.md) | `olares-cli dashboard` — Overview / Apps / GPU views | dashboard, overview, resource usage |
| [`olares-cluster`](skills/olares-cluster/SKILL.md) | `olares-cli cluster` — Olares ControlHub Kubernetes view | ControlHub, cluster, pods, workloads, namespaces, nodes, logs |

Skills are also published on [ClawHub](https://clawhub.io) (search "olares"); both channels read the same `SKILL.md` files, so you only need one of them installed.

## Three-layer command system

```
olares-cli <area> [<noun>] <verb> [flags]
```

- **System layer** (root-level, no `<area>` prefix): `install`, `uninstall`, `upgrade`, `start`, `stop`, `status`, `backup`, `precheck`, `prepare`, `download`, `change-ip`, `release`, `printinfo`, `logs`, `node`, `gpu`, `amdgpu`, `disk`, `osinfo`, `wizard`. These manage the host running Olares OS itself and require root / kubeconfig access — they are not driven by an Olares ID. *Channel availability*: the Go binary only registers them when `OLARES_CLI_REMOTE_ONLY` is unset, i.e. only when invoked from an Olares host's OS-bundled `/usr/local/bin/olares-cli`. Through `npm install -g @olares/cli` or `npx @olares/cli`, the Node shim sets `OLARES_CLI_REMOTE_ONLY=1` and they are hidden. The lone exception is `install`, which the Node shim itself intercepts and routes to the [first-run wizard](#first-run-wizard-scenario-a-recommended) — it never reaches the Go binary on the npm channel.
- **Identity-bound layer** (`<area>` = `profile` / `files` / `market` / `settings` / `dashboard` / `cluster`): act on behalf of the currently-selected Olares ID against a running Olares HTTP API. Pick the identity once with `olares-cli profile use <name>`, then every verb in this layer uses it. Reachable through both `npm install -g` and `npx`.

For every command, `--help` is the source of truth for flags and wire shapes:

```bash
olares-cli --help
olares-cli files --help
olares-cli files ls --help
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
npx skills remove beclab/Olares -y -g     # mirror of `skills add`
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
├── skills/               # AI-agent SKILL.md bundles, one per profile-based command tree
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
