# @olares/cli

Node wrapper for [`olares-cli`](https://github.com/beclab/Olares/tree/main/cli) — the official CLI for installing and operating [Olares](https://olares.com), an AI-native, self-hosted personal cloud.

This package downloads the platform-specific Go binary on `postinstall` and exposes it as `olares-cli` on your PATH. The binary itself is a single static Go executable; this Node wrapper just makes it `npx`-able and `npm install -g`-able.

## Quick start

```bash
# One-off — no install:
npx @olares/cli@latest <verb>

# Persistent global install (recommended for dev boxes, CI, macOS, Windows):
npm install -g @olares/cli@latest
olares-cli <verb>

# Bootstrap a fresh Linux server into a full Olares instance:
npx @olares/cli@latest install
```

After install, your AI coding agent can also pull the matching skill bundles:

```bash
npx skills add beclab/Olares -y -g
```

## Where the binary lives

| You ran | Binary ends up at |
| --- | --- |
| `npx @olares/cli@latest install` | `/usr/local/bin/olares-cli` (laid by Olares OS bootstrap) |
| `npm install -g @olares/cli` (no existing `olares-cli` on PATH) | `<npm prefix>/bin/olares-cli` (symlink managed by npm) |
| `npm install -g @olares/cli` (with existing `olares-cli` already at the target path) | npm aborts with `EEXIST`. Your existing binary is **never overwritten**. Use `--prefix` (see below) to side-step. |
| `npx @olares/cli@latest <verb>` | `~/.npm/_npx/<hash>/.../vendor/olares-cli` (temporary) |

## Uninstall

```bash
npm uninstall -g @olares/cli
```

npm cleans the symlink and the package files itself. There is no extra cleanup step.

## On an Olares host: install into a separate prefix

If you are already on an Olares host (where `/usr/local/bin/olares-cli` exists), `npm install -g` will refuse with `EEXIST`. To install the npm copy side-by-side without touching the OS bundle:

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # PATH order decides which copy wins
```

Don't use `npm install -g --force` on an Olares host — it would clobber the OS-managed binary.

## Environment

- `OLARES_CLI_DOWNLOAD_MIRROR` — base URL for downloading the prebuilt binary if `https://github.com/beclab/Olares/releases/download/...` is unreachable (defaults to `https://cdn.olares.com`).
- `OLARES_CLI_SKIP_DOWNLOAD=1` — install the JS shim only, no binary fetch.

## Links

- GitHub: <https://github.com/beclab/Olares>
- CLI docs: <https://github.com/beclab/Olares/tree/main/cli#readme>
- Olares product site: <https://olares.com>

## License

AGPL-3.0-or-later. See [LICENSE](https://github.com/beclab/Olares/blob/main/LICENSE).
