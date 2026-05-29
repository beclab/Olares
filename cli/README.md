# Olares CLI

[![License: AGPL-3.0-or-later](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8.svg)](https://go.dev)
[![npm @olares/cli](https://img.shields.io/npm/v/@olares/cli?label=npm%20%40olares%2Fcli)](https://www.npmjs.com/package/@olares/cli)

`olares-cli` is the command-line tool for [Olares](../README.md), an open-source personal cloud OS. It is a single Go binary that installs and operates the whole system: the OS, app market, files, dashboard, settings, cluster, and your identity. You can type its commands directly to manage the cluster, or use it with Agent Skills in your own AI runtime.

## Modes

`olares-cli` runs in three modes that differ in where they run and how they authenticate.

| Mode | Where it runs | How it authenticates |
| --- | --- | --- |
| Host mode | On the machine running Olares OS | Host root and kubeconfig, with no login needed |
| User mode | On any computer with `olares-cli` installed, on behalf of a logged-in user | An Olares profile and access token, through the same HTTP API as the web UI and LarePass |
| In-cluster mode (planned) | Inside an Olares app | Credentials injected as environment variables, with scope set by the app's `OlaresManifest` |

## Install

When you install Olares OS, either with the one-line command or from the ISO image, the CLI ships with it at `/usr/local/bin/olares-cli`, so you don't need to install it manually.

To use the CLI on another computer, or to drive Olares with an AI agent, install the standalone Olares CLI with npm. Pick the method that fits your needs:

### Method 1: Set up the CLI and Agent Skills

If you want an AI agent to drive Olares, this is the easiest way to start. The wizard sets up the CLI and all six Agent Skills together.

```bash
npx @olares/cli@latest install
```

Once it finishes, log in to wrap up the setup. The login is interactive and uses your Olares ID:

```bash
olares-cli profile login --olares-id <your-olares-id>   # asks for your password, and a TOTP code if 2FA is on
olares-cli profile list                                 # check that the profile is logged in
```

The wizard won't install Olares OS, and it won't log you in either.

### Method 2: Install the CLI only

```bash
npm install -g @olares/cli@latest
olares-cli profile login --olares-id <your-olares-id>
olares-cli files ls /drive/Home
```

This puts `olares-cli` on your PATH so you can manage a remote Olares. To add the Agent Skills later, see [Agent Skills](#agent-skills).

### Method 3: Run a single command

```bash
npx @olares/cli@latest profile login --olares-id <your-olares-id>
npx @olares/cli@latest files ls /drive/Home
```

This doesn't install anything. Each call runs from the npm cache, so it takes a second or two to start up. Your login is saved in the OS keychain, so you only sign in once.

### Special case: On a machine that runs Olares OS

This only comes up on a Linux machine that already runs Olares OS, where `/usr/local/bin/olares-cli` is already in place. You won't run into it on macOS or Windows, since their bundled binary lives inside a Linux environment.

If you want a newer standalone CLI than the bundled one, install an npm copy alongside it under a separate prefix:

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # PATH order decides which copy runs
olares-cli --version                            # now points to the npm copy
```

Now you have two binaries: `/usr/local/bin/olares-cli` for host commands and `~/.olares-cli-npm/bin/olares-cli` for user commands. If you only need to run something once, `npx @olares/cli@latest <verb>` saves you the PATH setup.

If you run a plain `npm install -g @olares/cli` here, it stops with an `EEXIST` error. That's expected. npm won't overwrite a binary it didn't install, so your bundled CLI stays safe. Never add `--force` to push past it, since that would overwrite the bundled CLI, which should only ever be updated through `olares-cli upgrade`.

## Agent Skills

The user and in-cluster modes are built for AI agents rather than typing commands by hand. To support that, `olares-cli` ships a set of Agent Skills, one per group of commands. Each one is a `SKILL.md` bundle with a `references/` folder that teaches an agent what each command does, which flags matter, how authentication works, and how to recover from common errors.

| Skill | Description |
| --- | --- |
| [`olares-shared`](skills/olares-shared/SKILL.md) | Profile model, login flows, token storage, automatic refresh, and auth-error recovery. The base for every other skill. |
| [`olares-files`](skills/olares-files/SKILL.md) | List, upload, download, edit, share, mount SMB, and manage Sync repos. |
| [`olares-market`](skills/olares-market/SKILL.md) | Browse, install, upgrade, uninstall, and upload local charts. |
| [`olares-settings`](skills/olares-settings/SKILL.md) | Read and modify the settings the web UI exposes. |
| [`olares-dashboard`](skills/olares-dashboard/SKILL.md) | Overview and app metrics, with a stable JSON schema. |
| [`olares-cluster`](skills/olares-cluster/SKILL.md) | Read and modify pods, workloads, nodes, jobs, cronjobs, and middleware passwords. |

Install `olares-shared` first. Every other skill assumes it for the profile model, token refresh, and auth-error recovery. An agent that loads only `olares-files`, for example, hits auth errors with no way to recover.

The wizard installs all six skills for you. To install them on their own, run:

```bash
npx skills add beclab/Olares -y -g
```

This adds the skills to your agent, such as Cursor, Claude Code, Codex, or OpenClaw, and the agent loads the matching one when you describe an Olares task. Since `olares-shared` ships in the same bundle, the shared-first requirement is met automatically. The skills are also published on [ClawHub](https://clawhub.io), which reads the same `SKILL.md` files, so install from whichever your agent supports.

When an agent uses these skills, you still run `olares-cli profile login --olares-id <id>` yourself, because it asks for your password and a TOTP code. The agent can then check the result with `olares-cli profile list`.

With the skills loaded, you drive Olares in natural language and the agent picks the command to run:

```plain
# Lists files through olares-files
List the files in the Home folder on my Olares device

# Installs an app through olares-market
Install Firefox from Market and tell me when it's ready

# Checks resource usage through olares-dashboard
Show me which apps are using more than 1 GB of memory
```

## Authentication

User mode signs in with a profile, which pairs one Olares instance with one user identity.

Log in once:

```bash
olares-cli profile login --olares-id alice@olares.com
```

The CLI asks for your password, then for a six-digit TOTP code if two-factor authentication is on. After that it refreshes tokens for you, and you log in again only when the refresh token expires.

| Task | Command |
| --- | --- |
| List profiles and their login status | `olares-cli profile list` |
| Show the current identity | `olares-cli profile whoami` |
| Switch to another profile | `olares-cli profile use <name>` |
| Switch back to the previous profile | `olares-cli profile use -` |
| Remove a profile and its token | `olares-cli profile remove <name>` |

Tokens are stored under the keychain service `olares-cli`, with your Olares ID as the account name, and they're never written in plaintext. To clear one, run `olares-cli profile remove` rather than editing files by hand.

| OS | Where tokens are stored |
| --- | --- |
| macOS | Keychain |
| Linux | AES-256-GCM file under `~/.local/share/olares-cli/` |
| Windows | DPAPI under `HKCU\Software\OlaresCli\keychain` |

## Command structure

```
olares-cli <area> [<noun>] <verb> [flags]
```

Commands fall into two groups that line up with host mode and user mode:

- **Host commands**: `install`, `upgrade`, `start`, `stop`, `status`, `precheck`, `node`, `gpu`, `disk`, `logs`, and similar. They manage the machine that runs Olares OS and need root or kubeconfig access. They run only from the bundled CLI on the host, and are hidden when you run the CLI through npm or npx.
- **User commands**: `profile`, `files`, `market`, `settings`, `dashboard`, and `cluster`. They act as the selected Olares user against a running Olares. Set the identity once with `olares-cli profile use <name>`, and every user command then uses it. These work from both the host CLI and the npm or npx CLI.

For any command, `--help` shows its flags:

```bash
olares-cli --help
olares-cli files --help
olares-cli files ls --help
```

## Output formats

Most user commands accept `-o table` for readable output and `-o json` for output that a script or agent can parse. `table` is the default. The JSON shape stays stable across minor versions.

```bash
olares-cli files ls /drive/Home -o json
olares-cli market list -o json | jq '.items[] | {name, version, status}'
```

## Uninstall

However you installed the CLI, here's how to undo it.

```bash
# Remove the CLI
npm uninstall -g @olares/cli

# Clear the npx cache. npx also clears it on its own after a few days.
rm -rf ~/.npm/_npx/

# Remove the Agent Skills
npx skills remove beclab/Olares -y -g

# Remove stored credentials
olares-cli profile list             # see what is stored
olares-cli profile remove <name>    # remove one profile and its token
```

## Build from source

Use this for local development or to pin a specific commit. You need Go 1.24+ and the full Olares repo, because `cli/go.mod` points to `../framework/oac` and a clone of `cli/` alone will not build.

```bash
git clone https://github.com/beclab/Olares.git
cd Olares/cli
go build -o olares-cli ./cmd/main.go
./olares-cli --help

# Or install to a prefix:
sudo make install                       # installs to /usr/local/bin/olares-cli
sudo make install PREFIX=$HOME/.local   # or a prefix you choose

# Install the Agent Skills. The wizard does this, a source build does not.
npx skills add beclab/Olares -y -g
```

A source build reports its version from `git describe`, like `1.12.7-3-gabc1234-dirty`, which is how you can tell it apart from an official release or an npm copy.

## Repository layout

```
cli/
├── cmd/        # CLI entry point and command tree, one folder per command under ctl/
├── pkg/        # install engine and remote API clients
├── internal/   # internal helpers: keychain, lockfile, files client
├── apis/       # kubekey v1alpha2 CRD types
├── skills/     # agent SKILL.md bundles, one per user command area
├── npm/        # @olares/cli npm wrapper, downloads the Go binary on postinstall
├── version/    # VERSION and VENDOR build targets
└── go.mod
```

The install engine in `pkg/core` runs a pipeline of `Pipeline → Module → Task → Action`. A host moves through five stages: `precheck` checks the environment, `download` fetches assets, `prepare` sets up dependencies, `install` brings up Kubernetes and the Olares core, and `wizard activate <olaresId>` registers the first user so the user commands become available. After install, `upgrade`, `start`, `stop`, and `uninstall` manage the running system.

## Security and risks

- **Credentials**: refresh tokens are kept in the OS keychain, as listed under [Authentication](#authentication). Access tokens are created on demand and never written to disk. Use `olares-cli profile remove` to clear them.
- **One identity at a time**: there is no per-command `--profile` flag. You set the identity once with `olares-cli profile use <name>`, so an agent or script should pick one identity and stay with it.
- **Confirmation by default**: any command that changes state, such as delete, restart, scale, install, or upgrade, asks before it runs. `--yes` skips the prompt. Treat it as a safety check, not a default.
- **Skills do not install the CLI**: a skill lists `olares-cli` as a requirement so an agent can warn when it is missing, but installing the CLI is a separate step.
- **Code signing**: on macOS and Windows the npm-downloaded binary is not signed yet, so Gatekeeper or SmartScreen may warn on first run. Check the download against the matching GitHub Release if you want to be sure.

## License

[AGPL-3.0-or-later](../LICENSE).
