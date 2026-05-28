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

- **Release-grade** (stable `1.12.7`, or pre-releases `-rc1` / `-beta.1` / `-alpha2`) → left alone; npm hits `EEXIST`, the wizard exits with Option 1 / Option 2 workaround block — each option already lists the matching `npx skills add beclab/Olares -y -g` follow-up, so you can copy the steps verbatim.
- **Dev / test / dirty** (`0.0.0-development` placeholder, `git describe` outputs like `1.12.7-3-gabc1234-dirty`, check.yaml's `1.12.7-12345678` PR builds, unparseable output) → removed so the npm copy can install over the same path. If `unlink` fails for permission reasons, the wizard exits with a one-line hint to re-run with `sudo` rather than silently failing.

## Environment

- `OLARES_CLI_DOWNLOAD_MIRROR` — base URL for downloading the prebuilt binary if `https://github.com/beclab/Olares/releases/download/...` is unreachable (defaults to `https://cdn.olares.com`).
- `OLARES_CLI_SKIP_DOWNLOAD=1` — install the JS shim only, no binary fetch.

## Links

- GitHub: <https://github.com/beclab/Olares>
- CLI docs: <https://github.com/beclab/Olares/tree/main/cli#readme>
- Olares product site: <https://olares.com>

## License

AGPL-3.0-or-later. See [LICENSE](https://github.com/beclab/Olares/blob/main/LICENSE).
