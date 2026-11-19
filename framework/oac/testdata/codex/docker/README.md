# codex base image

Ubuntu 24.04 base image for the Olares `codex` app with common language
runtimes, build toolchains, database clients, and extra dev headers pre-installed.

## What it ships

**Shell / tooling:** `ca-certificates`, `curl`, `wget`, `git`, `git-lfs`,
`openssh-client`, `jq`, `yq` (kislyuk YAML CLI), `less`, `procps`, `tzdata`,
`unzip`, `zip`, `vim-tiny`, `xz-utils`, `pigz`, `file`, `rsync`, `patch`,
`tmux`, `htop`, `dnsutils`, `iproute2`, `netcat-openbsd`, `socat`, `xmlstarlet`,
`shellcheck`, `graphviz`, `moreutils`, `locales` (C.UTF-8 pre-generated).

**Build:** `build-essential`, `pkg-config`, `cmake`, `ninja-build`, `clang`,
`clang-format`, `clang-tidy`, common `-dev` headers (SSL, FFI, zlib, bz2,
readline, sqlite, lzma, xml, xslt, yaml, jpeg, png, **curl**, **ICU**, **LDAP**,
**libpq**, **default-libmysqlclient-dev**), `gettext`, `protobuf-compiler`,
`libprotobuf-dev`, `gfortran`.

**Runtimes (distro):** Python 3 (`python3`, `pip`, `venv`, `pipx`), Node.js +
`npm`, Go (`golang-go`), Rust (`rustc`, `cargo`), Ruby + `bundler`, OpenJDK 21
(headless JDK), `maven`, **Gradle**, PHP 8.3 CLI + extensions (`curl`, `intl`,
`mysql`, `pgsql`, `sqlite3`, `mbstring`, `xml`, `zip`) + `composer`, `lua5.4`,
`sqlite3`, `perl`.

**Database clients:** `postgresql-client`, `default-mysql-client`,
`redis-tools`.

**User:** non-root `codex` (UID 1000), `HOME=/opt/data`, `PATH` prepends
`/opt/data/.local/bin`. `JAVA_HOME` points at OpenJDK 21.
`PIP_BREAK_SYSTEM_PACKAGES=1` so `pip install` into the distro Python is
allowed when needed.

**Olares integration (baked at build time):** `olares-cli` at `/usr/local/bin/olares-cli`
(vendor binary from `@olares/cli`, same as claudecode). Olares skill pack under
`/skills-staging`; `init-skills` flattens into `$HOME/.codex/skills` on the PVC
(Codex user skills via `.agents/skills` at build time; `init-skills` flattens into
`$CODEX_HOME/skills` on the PVC; agent flag `codex` via `npx skills add`).

**Not included:** the `codex` CLI standalone package (installed at runtime by
`init-install-codex` via OpenAI's official `install.sh` into
`/opt/data/.local/bin` and `~/.codex/packages/standalone`). `ripgrep` is
omitted from apt on purpose: Codex bundles its own.

## Build & push

```bash
# Single-arch (local dev)
docker build -t beclab/harveyff-codex-base:0.5.9 .

# Multi-arch release (recommended, matches chart supportArch)
docker buildx create --use --name codex-builder || true
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t beclab/harveyff-codex-base:0.5.9 \
  --push .
```

## Bumping the image

When bumping apt dependencies or base Ubuntu:

1. Increment the tag (e.g. `0.5.9 -> 0.5.10`).
2. Update all references in `codex/templates/deployment.yaml`
   (init-skills + init-install-codex + main container).
3. Bump `version` in `codex/Chart.yaml` and `metadata.version` in
   `OlaresManifest.yaml`.

The Codex standalone package (rust-v0.149.1 etc.) is controlled separately by
`appVersion` in `Chart.yaml` and `metadata.version` / `spec.versionName` in
`OlaresManifest.yaml`. `init-install-codex` runs
`https://chatgpt.com/codex/install.sh --release <appVersion>` whenever
`$HOME/.install-state/codex-standalone-<version>-<target>` is missing or the
install is binary-only (no `~/.codex/packages/standalone/current`).
