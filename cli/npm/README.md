# @olares/cli

Node wrapper for [`olares-cli`](https://github.com/beclab/Olares/tree/main/cli) — the official CLI for installing and operating [Olares](https://olares.com), an AI-native, self-hosted personal cloud.

This package downloads the platform-specific Go binary on `postinstall` and exposes it as `olares-cli` on your PATH. The binary itself is a single static Go executable; this Node wrapper just makes it `npx`-able and `npm install -g`-able.

## Quick start

```bash
# First-run wizard (recommended): does `npm install -g @olares/cli` and
# `npx skills add beclab/Olares -y -g` for you, in that order.
npx @olares/cli@latest install

# Or do it yourself, step by step:
npm install -g @olares/cli@latest                # persistent global install
npx skills add beclab/Olares -y -g               # six olares-* agent skills

# One-off — no install:
npx @olares/cli@latest <verb>
```

After any of those, authenticate (interactive, prompts for password + optional TOTP):

```bash
olares-cli profile login --olares-id <your-olares-id>
olares-cli profile current      # verify
```

> This package distributes the `olares-cli` binary as a **client** only. The Node wrapper auto-sets `OLARES_CLI_REMOTE_ONLY=1`, which hides the Go binary's host-side verbs (`uninstall`, `upgrade`, `node`, `os`, `gpu`, `disk`, `wizard`, `user`, `osinfo`, `amdgpu`); these are reachable only on an Olares host through `/usr/local/bin/olares-cli`. The `install` verb is intercepted by the Node shim itself and routed to the first-run wizard (it never reaches the Go binary). Installing Olares OS itself is out of scope for this package — on a Linux host run `curl -fsSL https://olares.sh | bash`.

> **Permission errors on Linux** (`EACCES` while npm writes to `/usr/lib/node_modules` or `/usr/local/lib/node_modules`): typical for distro-packaged Node (`apt install nodejs`) where the global prefix is root-owned. The wizard surfaces the offending npm `stderr` plus a one-time fix that switches npm to a user-owned prefix (`npm config set prefix ~/.npm-global` + `PATH`) so global installs no longer need `sudo` and `npx skills add -g` writes under your user (not `/root`).

## Where the binary lives

| You ran | Binary ends up at |
| --- | --- |
| On a Linux Olares host (OS bundle already there) | `/usr/local/bin/olares-cli` (managed by Olares OS upgrades) |
| `npm install -g @olares/cli` (no existing `olares-cli` on PATH) | `<npm prefix>/bin/olares-cli` (symlink managed by npm) |
| `npm install -g @olares/cli` (with existing `olares-cli` already at the target path) | npm aborts with `EEXIST`. Your existing binary is **never overwritten**. Use `--prefix` (see below) to side-step. |
| `npx @olares/cli@latest <verb>` | `~/.npm/_npx/<hash>/.../vendor/olares-cli` (temporary) |

## Uninstall

```bash
npm uninstall -g @olares/cli
```

npm cleans the symlink and the package files itself. There is no extra cleanup step.

## On a Linux Olares host: install side-by-side with the OS bundle

Linux-only: macOS / Windows / non-Olares Linux never hit this.

The OS-bundled `olares-cli` is pinned to whatever shipped with your Olares OS release (e.g. **1.12.5**). Older bundles don't include the agent / identity verbs (`profile`, `files`, `market`, `dashboard`, `settings`, `cluster`) — those land in newer npm releases first. To get them on the same Linux host as the OS bundle, install the npm copy under a separate prefix, or use `npx` for one-offs:

```bash
# Option 1 — separate prefix (npm aborts with EEXIST otherwise):
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # PATH order decides which copy wins

# Option 2 — npx, no install:
npx @olares/cli@latest profile current
```

Don't use `npm install -g --force` on an Olares host — it would clobber the OS-managed binary.

### What the `npx @olares/cli@latest install` wizard does on this path

Before running `npm install -g`, the wizard reads `--version` on the existing `/usr/local/bin/olares-cli` (or `/usr/bin/olares-cli`):

- **Release-grade** (stable `1.12.7`, or pre-releases `-rc1` / `-beta.1` / `-alpha2`) → left alone; if `npm config get prefix` points at the same `bin` directory (typical Olares host: `/usr/local`), the wizard short-circuits the `npm install -g` attempt (no full install timeout) and exits with a side-by-side install block (`npm install -g ... --prefix=$HOME/.olares-cli-npm` + `PATH` export + `npx skills add beclab/Olares -y -g`) you can copy verbatim.
- **Dev / test / dirty** (`0.0.0-development` placeholder, `git describe` outputs like `1.12.7-3-gabc1234-dirty`, check.yaml's `1.12.7-12345678` PR builds, unparseable output) → removed so the npm copy can install over the same path. If `unlink` fails for permission reasons, the wizard exits with a one-line hint to re-run with `sudo` rather than silently failing.

## Environment

- `OLARES_CLI_DOWNLOAD_MIRROR` — base URL for downloading the prebuilt binary if `https://github.com/beclab/Olares/releases/download/...` is unreachable (defaults to `https://cdn.olares.com`).
- `OLARES_CLI_SKIP_DOWNLOAD=1` — install the JS shim only, no binary fetch.

## Versioning and release (maintainers)

Two version numbers are tracked:

| Field | Example | Used for |
| --- | --- | --- |
| **npm version** | `1.12.6-cli.5` | `package.json`, npm registry, CDN/GitHub tar names |
| **binary version** (`version.VERSION`) | `1.12.6` | Olares OS upgrade line; shown by `olares-cli --version` |

`postinstall` downloads `olares-cli-v{npm_version}_{platform}.tar.gz` using the npm package version, so each `-cli.N` bump is a distinct artifact and does not overwrite prior builds.

**npm dist-tags**

| Tag | Meaning |
| --- | --- |
| `latest` | Production default (`npx @olares/cli@latest`, install wizard). Set manually via **Tag npm CLI** workflow. |
| `next` | Fresh CI publish from **Release CLI** (`npx @olares/cli@next`). Does not move `latest`. |
| `daily` | Optional tag for daily builds (manual promote). |

**Release CLI** (`release-cli.yaml`, manual dispatch): builds with GoReleaser, uploads tars to CDN, publishes `@olares/cli` with `--tag next`. Inputs: `branch` (default `main`), `version` (npm), optional `binary-version` (defaults from `version` — strips `-cli.N` suffix).

**Tag npm CLI** (`tag-npm-cli.yaml`): promotes an already-published version to a dist-tag (default `latest`). Fails if the version is not on npm.

Daily / OS release pipelines call **Release CLI** with `publish-npm: false` — they only upload binaries; npm publish stays a separate manual step.

## Links

- GitHub: <https://github.com/beclab/Olares>
- CLI docs: <https://github.com/beclab/Olares/tree/main/cli#readme>
- Olares product site: <https://olares.com>

## License

AGPL-3.0-or-later. See [LICENSE](https://github.com/beclab/Olares/blob/main/LICENSE).
