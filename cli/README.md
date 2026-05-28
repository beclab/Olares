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
- **Distributed three ways** — one-line `npx ... install` bootstrap; persistent `npm install -g`; or zero-install `npx <verb>` for one-offs

## Install — which path fits you?

| Your situation | Use this | Why |
| --- | --- | --- |
| I want to turn a fresh Linux server into an Olares instance | `npx @olares/cli@latest install` <br>([Scenario A](#bootstrap-an-olares-host-scenario-a)) | One-line bootstrap; after `install` completes, `olares-cli` lives at `/usr/local/bin/` and is managed by Olares OS upgrades |
| I'm already on an Olares host and want to use its CLI | Use `/usr/local/bin/olares-cli` directly | It's the OS-bundled copy, kept in sync via `olares-cli upgrade`. No need to install via npm. |
| I want to control a remote Olares from my dev box, CI, macOS, or Windows | `npm install -g @olares/cli@latest` <br>([Scenario B](#client-on-a-non-olares-machine-scenario-b)) | Persistent install; `olares-cli` on PATH. Safe even if you already have an OS bundle — postinstall auto-skips the alias on conflict. |
| I just want to run one command quickly without installing | `npx @olares/cli@latest <verb>` <br>([Scenario C](#one-off-ops-scenario-c)) | No persistent files; ~1-2 s cold-start per invocation. Keychain/token caches persist per-user. |

### Bootstrap an Olares host (Scenario A)

```bash
# On a fresh Linux server you want to turn into an Olares instance:
npx @olares/cli@latest install

# After install completes, olares-cli is at /usr/local/bin/olares-cli (on PATH).
# The npm-cached copy is no longer needed.
olares-cli profile login <your-olares-id>
olares-cli dashboard overview
```

### Client on a non-Olares machine (Scenario B)

```bash
# macOS / Windows / Linux dev box that talks to a remote Olares.
npm install -g @olares/cli@latest

# The package's only PATH-exposed bin is `olares-cli`, managed by npm itself.
# If an existing `olares-cli` is already at npm's target path (i.e. you're on
# an Olares host where the OS bundle owns /usr/local/bin/olares-cli), npm
# refuses the install with EEXIST — your existing binary is never overwritten.
# See "On an Olares host" below for the --prefix workaround in that case.

olares-cli profile login <your-olares-id>
olares-cli files ls /drive/Home
```

### One-off ops (Scenario C)

```bash
# No persistent install; each invocation re-uses the npm cache:
npx @olares/cli@latest profile current
npx @olares/cli@latest files ls /drive/Home
```

### Capabilities & limits of each install method

| Method | What it gives you | What it can't / won't do |
| --- | --- | --- |
| **A — `npx @olares/cli@latest install`** | Turns a Linux host into an Olares instance (downloads payload, installs k8s, lays `olares-cli` at `/usr/local/bin/`). Subsequent invocations are direct `olares-cli <verb>` on that host. | Requires Linux + root (or sudo) + ~50 GB free + 30-60 min runtime. Not usable on macOS / Windows / non-Linux. Doesn't help "I just want to control a remote Olares". |
| **B — `npm install -g @olares/cli@latest`** | Persistent `olares-cli` CLI on PATH (macOS / Windows / Linux). Use to talk to a *remote* Olares (login, files, market, dashboard, cluster, settings). | Doesn't install Olares OS itself — `olares-cli install` from this client targets `localhost` and will fail on a non-Linux machine. If `olares-cli` already exists at npm's target path (e.g. on an Olares host where the OS bundle owns `/usr/local/bin/olares-cli`), npm aborts with `EEXIST` — see ["On an Olares host"](#on-an-olares-host-install-into-a-separate-prefix) for the workaround. |
| **C — `npx @olares/cli@latest <verb>`** | Zero-install, runs any verb without touching PATH. Great for CI one-shots, ephemeral containers, "just try it". | Each invocation pays a ~1-2 s npx cold-start. Long watches (`olares-cli market list --watch`, `olares-cli cluster pod logs -f`) work but pay the cost up front. Keychain/token caches persist across npx invocations. |

### Install AI agent skills

```bash
npx skills add beclab/Olares -y -g
```

Installs the 6 `olares-*` skill bundles via [vercel-labs/skills](https://github.com/vercel-labs/skills) into your active agent (Cursor, Claude Code, Codex, OpenClaw, and ~50 others). The agent then auto-loads the right skill when you mention Olares-flavoured tasks.

> Note: the `@olares` npm scope and the `beclab` GitHub org are independent naming spaces. The canonical source repo is `github.com/beclab/Olares`, so the `skills add` command (which fetches via GitHub) uses `beclab/Olares`, while the npm package is published as `@olares/cli`.

## Quick start (AI agent)

If you are an AI coding agent reading this README to bootstrap an environment:

```bash
# 1. Install the CLI (pick one based on the target machine):
npx @olares/cli@latest install            # fresh Linux host → full Olares install
npm install -g @olares/cli@latest         # dev box / CI / macOS / Win → client only

# 2. Install the agent skills into your own runtime:
npx skills add beclab/Olares -y -g

# 3. Ask the human to log in (interactive — browser + OS keychain):
#    olares-cli profile login <their-olares-id>

# 4. Verify identity + reach:
olares-cli profile current
olares-cli dashboard overview
```

Then load the appropriate skill (`olares-shared` is the foundation; load it first before any of the others) and follow its instructions.

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

- **System layer** (root-level, no `<area>` prefix): `install`, `uninstall`, `upgrade`, `start`, `stop`, `status`, `backup`, `precheck`, `prepare`, `download`, `change-ip`, `release`, `printinfo`, `logs`, `node`, `gpu`, `amdgpu`, `disk`, `osinfo`, `wizard`. These manage the host running Olares OS itself and require root / kubeconfig access — they are not driven by an Olares ID.
- **Identity-bound layer** (`<area>` = `profile` / `files` / `market` / `settings` / `dashboard` / `cluster`): act on behalf of the currently-selected Olares ID against a running Olares HTTP API. Pick the identity once with `olares-cli profile use <name>`, then every verb in this layer uses it.

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

## Where is olares-cli after install?

| Scenario | `olares-cli` resolves to |
| --- | --- |
| **A** — `npx ... install` on a fresh host | `/usr/local/bin/olares-cli` (OS bundle, on PATH after install completes) |
| **B** — `npm install -g` on a host without an existing `olares-cli` | `<npm prefix>/bin/olares-cli` (symlink managed by npm itself) |
| **B-conflict** — `npm install -g` on a host that already has `olares-cli` at npm's target path | npm aborts with `EEXIST`; nothing is installed. The pre-existing binary stays put. Use the [`--prefix` workaround](#on-an-olares-host-install-into-a-separate-prefix) to coexist. |
| **C** — `npx @olares/cli@latest <verb>` | `~/.npm/_npx/<hash>/.../vendor/olares-cli` (only during the npx subprocess) |

## Uninstall

Pick the reverse operation that matches how you installed.

### Remove the CLI client (Scenario B)

```bash
npm uninstall -g @olares/cli
# npm cleans the `olares-cli` symlink and the package files itself —
# there is no extra cleanup step.
```

### Clear the npx cache (Scenario C)

```bash
# npx auto-evicts the cache after a few days. To force-clear sooner:
rm -rf ~/.npm/_npx/                       # nukes all npx-cached packages
# Or, more surgical:
ls ~/.npm/_npx/                            # find the hash dir for @olares/cli
rm -rf ~/.npm/_npx/<hash>/
```

### Remove Olares OS from a host (Scenario A — destructive!)

```bash
olares-cli uninstall                       # revert the most recent install phase
olares-cli uninstall --all                 # complete removal (Olares + k8s + data)
```

> **Warning**: `olares-cli uninstall --all` deletes **all** Olares data (user files, installed apps, profiles, k8s state). Back up first via `olares-cli backup` or the Olares dashboard. The `/usr/local/bin/olares-cli` binary itself stays (you may still want it for re-install); to drop it too: `sudo rm /usr/local/bin/olares-cli`.

### Remove agent skills

```bash
npx skills remove beclab/Olares -y -g     # mirror of `skills add`
```

### Wipe stored credentials (any scenario)

```bash
olares-cli profile list                    # see what's stored
olares-cli profile remove <name>           # delete one profile + its keychain token
```

Credentials live in the OS-native keychain (macOS Keychain / Windows DPAPI / Linux secret-service or filesystem fallback at `~/.olares/credentials/`). `profile remove` is always the right cleaning verb — don't hand-edit those files.

## On an Olares host: install into a separate prefix

If you are on an Olares host (where `/usr/local/bin/olares-cli` already exists from `olares-cli install`), a plain `npm install -g @olares/cli` aborts with `EEXIST`. To install the npm copy side-by-side without touching the OS bundle, use a separate prefix:

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # PATH order decides which copy wins
olares-cli --version                            # resolves to the npm copy
# Revert: reorder PATH or `rm -rf ~/.olares-cli-npm/`
```

> Do **not** use `npm install -g --force` to bypass `EEXIST` on an Olares host — that would clobber the OS-managed binary. The OS bundle is the canonical CLI on an Olares host and is upgraded via `olares-cli upgrade`.

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
