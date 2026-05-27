# Olares CLI

`olares-cli` is the official CLI for installing and operating Olares. The full command reference lives at [docs.olares.com](https://docs.olares.com/developer/install/cli-1.12/olares-cli.html).

## Build

Requires **Go 1.24+**.

```bash
cd cli
go build -o olares-cli ./cmd/main.go
```

## Usage

`olares-cli` runs in three modes.

### 1. Host operator

On the same machine as the Olares OS, install / start / stop / maintain the local installation. No remote login — identity is the host itself (root + kubeconfig where applicable).

Commands: `osinfo`, `os` (`install`, `start`, `stop`, `status`, `upgrade`, `uninstall`, ...), `node`, `gpu`, `amdgpu`, `disk`.

```bash
./olares-cli --help
./olares-cli install
```

### 2. User agent

Act on behalf of a specific user against a running Olares (requires Olares 1.12.5 or newer), over the same access-token HTTP API used by the web UI and LarePass. First version assumes the CLI is on the same host as the OS.

Get credentials in one of three ways:

- `olares-cli wizard activate` — activate Olares and keep the generated mnemonic / refresh token
- `olares-cli profile import` — paste an existing refresh token
- OAuth a refresh token from the desktop LarePass app

Refresh tokens live in the OS keychain (macOS Keychain / Linux AES file / Windows DPAPI); access tokens are refreshed transparently.

Commands: `profile`, `files`, `market`, `settings`, `dashboard`, `cluster`, ...

```bash
olares-cli profile login --olares-id alice@olares.com
olares-cli market list
olares-cli files ls /drive/Home
```

### 3. In-cluster agent

Same command surface as mode 2 (requires Olares 1.12.6 or newer), but the CLI runs inside a cluster app's container (e.g. an Openclaw skill) and acts as the current user with no login: credentials are injected as environment variables according to the app's declared scope, and requests are forwarded through `user-service`.

```bash
# inside an Olares app container — no profile login needed
olares-cli files ls /drive/Home
```

### AI agent skills

Modes 2 and 3 are both driven primarily by AI agents, not humans typing commands. The [`skills/`](skills/) directory ships one SKILL.md per profile-based command tree (`olares-shared`, `olares-market`, `olares-files`, `olares-dashboard`, `olares-settings`, `olares-cluster`) — loaded by AI coding agents (Cursor, Claude, ...) on a user's machine in mode 2, and by in-cluster apps (e.g. Openclaw) in mode 3. Each skill teaches the agent what the verbs do, which flags matter, how auth/refresh works, and how to recover from common errors. All non-shared skills assume `olares-shared` is loaded first.

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
├── version/              # VERSION / VENDOR ldflag targets
├── .goreleaser.yaml
└── go.mod
```

The install engine in `pkg/core` runs a `Pipeline → Module → Task → Action` stack. Each mode-1 command moves the host between five lifecycle stages:

- **prechecked** — `olares-cli precheck` validates the environment against install requirements; gating step before any state-changing action.
- **downloaded** — `olares-cli download` (`component` / `wizard`) fetches the install assets; `olares-cli download check` verifies completeness.
- **prepared** — `olares-cli prepare` lays out dependencies.
- **installed** — `olares-cli install` brings up Kubernetes and Olares core; `olares-cli upgrade` moves an installed host to a newer version; `olares-cli start` / `stop` / `status` toggle the runtime; `olares-cli uninstall` rolls back (optionally to a specific phase); `olares-cli change-ip` repairs after an IP change.
- **activated** — `olares-cli wizard activate <olaresId>` enrols the first user against BFL/Auth, after which the profile-based commands of modes 2 and 3 become usable.
