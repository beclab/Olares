# Olares CLI

[![License: AGPL-3.0-or-later](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8.svg)](https://go.dev)
[![npm @olares/cli](https://img.shields.io/npm/v/@olares/cli?label=npm%20%40olares%2Fcli)](https://www.npmjs.com/package/@olares/cli)

`olares-cli` is the official CLI for installing and operating [Olares](https://olares.com) — an AI-native, self-hosted personal cloud. One static Go binary drives the whole product: OS bootstrap, Olares ID, Files, Market, Dashboard, Settings, ControlHub, Router. It also carries the twelve `olares-*` agent skills that document it, compiled in, so the instructions an agent reads and the verbs it can call always come from the same release.

Three ways to read the rest of this page, depending on who you are:

- [**For users**](#for-users) — install the CLI, log in, upgrade it
- [**For AI agents**](#for-ai-agents) — get the skills onto this machine and keep them current
- [**For developers**](#for-developers) — build the CLI, install your build, edit skills live

## Which copy of olares-cli am I running?

There are four, and they differ in who upgrades them and which verbs they expose. `command -v olares-cli` and `olares-cli --version` tell you which one you have.

| Copy | Path | Upgraded by | Verbs |
| --- | --- | --- | --- |
| **OS bundle** — a Linux host running Olares OS | `/usr/local/bin/olares-cli` | `olares-cli upgrade`, and Olares OS releases | all of them, including the [system layer](#three-layer-command-system) |
| **npm global** | `<npm prefix>/bin/olares-cli` | `npm install -g @olares/cli@<tag>` | everything except the system layer |
| **npx**, for the length of one command | `~/.npm/_npx/<hash>/…/vendor/olares-cli`, temporary | nothing — each run resolves the tag you name | same as npm global |
| **your own build** | wherever `make install` put it | you | all of them |

The npm channel hides the whole system layer — everything that manages an Olares host, from `upgrade` and `start` through `node`, `gpu` and `disk` — because those verbs need a host filesystem laid down by the OS installer. The Node shim sets `OLARES_CLI_REMOTE_ONLY=1` and the Go binary then never registers them ([cmd/ctl/root.go](cmd/ctl/root.go)), so they do not appear in `--help` and return `unknown command`. `install` is the one exception: the shim keeps that name for the setup wizard, so it never reaches the Go binary on that channel.

## For users

### Install

**On a Linux host running Olares OS** you already have it, at `/usr/local/bin/olares-cli`, and it stays in step with the OS. Go straight to [logging in](#log-in). If that bundle is older than the identity verbs you need, see [On a Linux Olares host](#on-a-linux-olares-host).

**Anywhere else** — macOS, Windows, a Linux dev box, CI — one command, Node 18+:

```bash
npx @olares/cli@latest install
```

`install` is handled by the Node shim rather than the Go binary. It does two things: `npm install -g` the same version as the wizard you just invoked, then `olares-cli skills install`. So the tag you name decides the whole setup — `@latest` leaves you with a `latest` CLI and `latest` skills, `@next` with both from `next`.

**For a single command, installing nothing:**

```bash
npx @olares/cli@latest files ls /drive/Home
```

Each invocation pays roughly a second or two of npx cold start and leaves nothing on PATH. Your login survives anyway, because it lives in the OS keychain rather than in the package.

> This wizard does **not** install Olares OS. Bootstrapping a Linux host is `curl -fsSL https://olares.sh | bash` — see [docs/manual/get-started](https://docs.olares.com/manual/get-started/install-olares/linux.html). There is also no `config init` step: Olares authenticates with your Olares ID directly.

### Log in

Interactive — it opens a browser and may ask for a TOTP code — which is why the wizard leaves it to you:

```bash
olares-cli profile login --olares-id <your-olares-id>
olares-cli profile current
```

### Upgrade

Two npm tags, and the difference matters:

| Tag | What it points at |
| --- | --- |
| `@latest` | the promoted stable release. Promotion is a manual step, so this can sit several releases behind |
| `@next` | every release, as CI publishes it |

```bash
npx @olares/cli@latest install     # re-run the wizard: CLI and skills together
npm install -g @olares/cli@next    # or move just the CLI, to a tag you pick
olares-cli skills install          # then bring the skills along
```

That last line is the step people forget, so the CLI reminds you: while the skills on disk name a different release than the binary, every command other than `skills` itself prints one line about it on stderr. On an Olares host the OS bundle upgrades through `olares-cli upgrade` instead, which does not touch the skills — run `skills install` after it too.

### Uninstall

```bash
npm uninstall -g @olares/cli      # npm removes the symlink and the package files itself
rm -rf ~/.npm/_npx/               # optional: the npx cache, which auto-evicts anyway
rm -rf ~/.agents/skills/olares-*  # the skills store
rm -f ~/.claude/skills/olares-*   # and each agent's links into it
olares-cli profile remove <name>  # one profile and its stored credentials
```

Credentials live in the OS-native keychain (macOS Keychain, Windows DPAPI, Linux secret-service, or `~/.olares/credentials/` as a fallback). `profile remove` is the only supported way to clear them; don't hand-edit those files.

## For AI agents

The suite is compiled into the binary ([skills/embed.go](skills/embed.go)), which is the point: reading a skill cannot fetch a version that disagrees with the verbs available here. Every verb below touches local files only — no profile, no network, no cluster — so all of them work on the npm and npx channels too.

### Read them without writing anything

```bash
olares-cli skills list                            # names, versions, one-line summaries
olares-cli skills list -o json
olares-cli skills read olares-shared              # a SKILL.md, byte for byte, on stdout
olares-cli skills list olares-router/references   # what a skill links to
olares-cli skills read olares-router references/olares-router-calling.md
```

Load `olares-shared` first: it routes a request to the right skill and describes the auth model the others assume.

### Put them where this machine's agents look

```bash
olares-cli skills install
```

One copy goes to `~/.agents/skills`, and every agent skills directory that **already exists** here is linked to it — Cursor, Claude Code, Codex, OpenClaw, and the rest. A directory is never created for an agent that is not installed, so nothing is left behind for tools you do not use. Where symlinks are unavailable (Windows without developer mode) the files are copied instead, and the report says which happened where.

To write the suite somewhere of your own choosing — a container image, a project directory, a harness that boots from a fixed path:

```bash
olares-cli skills export ./skills
```

### Keep them current

Run `skills install` again after every CLI upgrade. Until you do, each command prints a line on stderr saying the installed skills name a different release; `OLARES_CLI_NO_SKILL_NOTICE=1` silences it where the extra output cannot be tolerated.

If `olares-cli` is not on the machine at all, install it first ([For users](#for-users)) and then run `skills install`. Skill discovery never installs the CLI for you — `metadata.requires.bins` is advisory, so an agent can warn instead of guessing.

### Where a copy can come from, and what each costs

| Route | Writes | Version you get | Notice sees it |
| --- | --- | --- | --- |
| `olares-cli skills install` | `~/.agents/skills` + existing agent directories | exactly the binary's | yes |
| `olares-cli skills export <dir>` | wherever you say | exactly the binary's | no |
| `olares-cli skills read <skill>` | nothing | exactly the binary's | n/a |
| `npx skills add beclab/Olares` | the `skills` CLI decides | this repository's `main`, whatever it is today | no |
| [ClawHub](https://clawhub.ai) (search "olares") | the ClawHub CLI decides | whatever was last accepted there — see below | no |

The first three are the same bytes, so they cannot disagree with the verbs the binary has. The last two can, and silently: a skill declares `requires.bins: [olares-cli]`, which any build satisfies, so an agent reading `main`'s instructions against a six-month-old binary gets told to run flags that do not exist. That is what "notice sees it" means — the [drift notice](#keep-them-current) compares the store against the binary, so it catches a stale `skills install` and nothing else.

**ClawHub is not being updated.** A skill's version now names the release it ships in (`1.12.7-cli.4`), which is numerically below the per-skill numbering the registry already holds (`olares-chart` reached `4.18.0`), so a push is refused by a registry that requires increasing versions. `publish.sh` still works if that is ever resolved; until then, treat what is there as a copy from before the suite moved into the binary.

There is also no Claude Code plugin marketplace here, deliberately. A marketplace entry points at a git ref, which puts it in the bottom half of that table: a version nobody can check against the binary in front of it.

> Logging in is the human's job: `olares-cli profile login --olares-id <id>` opens a browser. Verify with `olares-cli profile current`, then `olares-cli dashboard overview`.

### The suite

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

Each is one `SKILL.md` plus a `references/` folder, loaded on demand.

## For developers

### Build

Go **1.25+** ([go.mod](go.mod)). Clone the whole monorepo: `go.mod` carries relative `replace` directives into `../framework/oac` and `../framework/app-gateway`, so a `cli/`-only checkout cannot resolve its own dependencies.

```bash
cd cli
make build                      # ./olares-cli, version stamped from `git describe`
go build -o olares-cli ./cmd    # or plain go, leaving the version at its source default
./olares-cli --help
```

Either way `cli/skills/` is compiled in, so a skill you just edited is part of the build you just made.

### Install your build

```bash
sudo make install                    # /usr/local/bin/olares-cli
make install PREFIX="$HOME/.local"   # or a prefix you own, no sudo needed
olares-cli skills install            # write out the skills this build carries
make uninstall                       # remove the binary again
```

**Never `sudo` that third line.** It leaves root-owned `~/.agents/skills/olares-*` directories that break the next install with `EACCES`. Only an install into a system prefix needs root.

Updating afterwards is the same three steps: `git pull && sudo make install && olares-cli skills install`.

On Windows, `make install` needs Git Bash / MSYS; `skills install` runs natively and falls back to copying where linking needs privileges.

### Edit skills without rebuilding

Point the agents at your checkout, so a saved `SKILL.md` is what they read on the next turn:

```bash
ln -s "$(pwd)/skills/olares-shared" ~/.agents/skills/olares-shared   # repeat per skill
```

`skills install` never writes over links like these, because repointing them at the installed copies would end the live-editing loop without saying so. Where it stops depends on where the link is: one in `~/.agents/skills`, as above, is the copy every agent reads, so the whole install refuses and names it; one in a single agent's directory only skips that agent, and the rest of the machine still gets the update. Remove the link, or pass `--force`, when you want the installed copies back.

### Test

```bash
go test ./skills/... ./cmd/ctl/skills/...   # embedding, export, install, the drift notice
go test ./cmd/ctl -run TestEveryCommandTheSkillsDocumentResolves
python3 -m pip install -r skills/requirements.txt
python3 -m unittest skills/test_validate.py
python3 skills/validate.py                  # frontmatter, and one version across the suite
bash skills/publish.sh --dry-run
```

That is what [skills-ci.yml](../.github/workflows/skills-ci.yml) runs on a pull request.

### What `--version` reports

| Build | Version |
| --- | --- |
| `make build` / `make install` | `git describe --tags --always --dirty`, e.g. `1.12.7-20260821-9-gea3fb979f` |
| plain `go build` | `0.0.0-development`, the default in [version/version.go](version/version.go) |
| a release | the release version, stamped through ldflags in CI |

Both dev forms matter beyond your own shell: the setup wizard reads `--version` on `/usr/local/bin/olares-cli` and replaces anything that is not release-grade, so a `make install` build parked there is removed rather than preserved. The embedded skills carry the npm release version (`1.12.7-cli.4`) either way, and CI asserts that a published binary's skills name the release it shipped with.

## Three-layer command system

```
olares-cli <area> [<noun>] <verb> [flags]
```

- **System layer** (root-level, no `<area>` prefix): `install`, `uninstall`, `upgrade`, `start`, `stop`, `status`, `backup`, `precheck`, `prepare`, `download`, `change-ip`, `release`, `printinfo`, `logs`, `node`, `gpu`, `amdgpu`, `disk`, `osinfo`, `wizard`, `user`. These manage the host running Olares OS and need root / kubeconfig access; no Olares ID is involved. Registered only when `OLARES_CLI_REMOTE_ONLY` is unset — in practice, only through an Olares host's `/usr/local/bin/olares-cli`.
- **Identity layer** (`<area>` = `profile`, `files`, `market`, `settings`, `dashboard`, `cluster`, `doctor`, `router`, `search`, `knowledge`): acts as the currently selected Olares ID against a running Olares HTTP API. Choose the identity once with `olares-cli profile use <name>`; every verb in this layer then uses it. Available on every channel.
- **Local layer** (`<area>` = `skills`, `chart`, `preinstall`): this machine's files only — no Olares instance, no identity, no network. `skills` reads the suite compiled into the binary and writes it out; `chart` authors and validates an app chart before anything is deployed.

`--help` is the source of truth for flags and wire shapes:

```bash
olares-cli --help
olares-cli files --help
olares-cli files ls --help
```

## Output formats

Most identity-layer verbs accept `--output table` (default, human-readable) and `--output json`. Use JSON whenever a script or an agent has to parse the result; the schema is intentionally stable across minor versions.

```bash
olares-cli files ls /drive/Home --output json
olares-cli market list --output json | jq '.items[] | {name, version, status}'
olares-cli cluster workload images --limit 50 --page 1 -o json
```

## On a Linux Olares host

This applies **only** to Linux hosts with Olares OS installed, where `/usr/local/bin/olares-cli` already exists. macOS, Windows and non-Olares Linux boxes never hit it.

The OS bundle is pinned to whatever shipped with your Olares OS release, and identity verbs land in npm releases first, so a host bundle can be missing verbs you want. The bundle is canonical for the system layer and `olares-cli upgrade` is what is supposed to replace it — so install the npm copy *beside* it rather than over it. A plain `npm install -g @olares/cli` aborts with `EEXIST` for exactly that reason, and your existing binary is never overwritten.

```bash
# Option 1 — a separate prefix, PATH order decides which copy wins.
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
olares-cli --version
olares-cli skills install
# Revert by reordering PATH, or `rm -rf ~/.olares-cli-npm/`.

# Option 2 — npx, no persistent install and no PATH changes.
npx @olares/cli@latest profile current
```

Never pass `--force` to `npm install -g` here: it would clobber the OS-managed binary.

The setup wizard automates that safety net. It reads `--version` on the existing binary and:

- **release-grade** (stable `1.12.7`, or `-rc1` / `-beta.1` / `-alpha2`) → kept. If `npm config get prefix` resolves to the same `bin` directory — typical on an Olares host, where it is `/usr/local` — the wizard skips the doomed `npm install -g` attempt entirely rather than waiting out its timeout, and exits printing the Option 1 block above with the version it would have installed.
- **dev, test or unparseable** (`0.0.0-development`, `git describe` output like `1.12.7-3-gabc1234-dirty`, a CI build's `1.12.7-12345678`) → removed, so npm can install over the same path. If removal needs root, the wizard exits with a one-line sudo hint instead of failing quietly.

**Permission errors on Linux** (`EACCES` while npm writes to `/usr/lib/node_modules` or `/usr/local/lib/node_modules`) are typical for distro-packaged Node, where the global prefix is root-owned. The wizard surfaces npm's own `stderr` plus a one-time fix that moves npm to a user-owned prefix (`npm config set prefix ~/.npm-global`, then extend `PATH`) so global installs stop needing `sudo`.

## Repository layout

```
cli/
├── cmd/                  # CLI entrypoint and Cobra command tree
│   ├── main.go
│   └── ctl/              # one folder per top-level command (os, node, gpu, profile,
│                         #   market, files, dashboard, settings, cluster, skills, ...)
├── pkg/                  # install engine + remote API clients
│   ├── core/             # pipeline / module / task / action framework
│   ├── pipelines/        # top-level pipelines invoked by install/start/upgrade/...
│   └── ...               # one package per concern (k3s, etcd, gpu, storage, terminus, ...)
├── internal/             # non-exported helpers (keychain, lockfile, files client, ...)
├── apis/                 # kubekey v1alpha2 CRD types
├── skills/               # agent SKILL.md bundles, compiled into the binary by embed.go
├── npm/                  # @olares/cli wrapper (postinstall fetches the Go binary)
├── version/              # VERSION / VENDOR ldflag targets
├── .goreleaser.yaml
└── go.mod
```

The npm wrapper's `postinstall` downloads a prebuilt binary from GitHub Releases, falling back to `https://cdn.olares.com`; `OLARES_CLI_DOWNLOAD_MIRROR` overrides the second, and `OLARES_CLI_SKIP_DOWNLOAD=1` installs the shim alone. A local Go toolchain is needed only to modify the CLI itself.

The install engine in `pkg/core` runs a `Pipeline → Module → Task → Action` stack, moving a host through five lifecycle stages:

- **prechecked** — `olares-cli precheck` validates the environment; the gate before any state-changing action.
- **downloaded** — `olares-cli download` (`component` / `wizard`) fetches install assets; `download check` verifies completeness.
- **prepared** — `olares-cli prepare` lays out dependencies.
- **installed** — `olares-cli install` brings up Kubernetes and Olares core; `upgrade` moves an installed host forward; `start` / `stop` / `status` toggle the runtime; `uninstall` rolls back, optionally to a phase; `change-ip` repairs after an address change.
- **activated** — `olares-cli wizard activate <olaresId>` enrols the first user against BFL/Auth, after which the identity layer becomes usable.

## Security & risks

- **Credentials** — refresh tokens live in the OS-native keychain; access tokens are derived on demand and never persisted. `olares-cli profile remove` is the canonical way to wipe them.
- **Profile isolation** — there is no per-invocation `--profile` flag. Identity is single-source via `olares-cli profile use <name>`, so agents and scripts commit to one identity up front instead of hopping mid-pipeline.
- **`--yes` contract** — every mutating verb on the identity layer prompts by default. `--yes` is the agreed bypass; treat it as a safety check, not a style preference.
- **Skills are advisory about their host** — a skill declares `["olares-cli"]` under `metadata.requires.bins` so an agent can warn when the binary is missing. Discovery never installs it.
- **Code signing** — on macOS and Windows the downloaded binary is currently unsigned, so Gatekeeper or SmartScreen may warn on first run. Verify with `sha256sum` against the matching GitHub Release if you need certainty.

## License

[AGPL-3.0-or-later](../LICENSE).
