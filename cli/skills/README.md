# olares-cli skills

LLM-readable skills (one folder per command tree) that teach OpenClaw — and any other Claude-style agent — how to drive `olares-cli` against a live Olares instance. Each folder contains a single `SKILL.md` whose YAML frontmatter declares the skill's name, version, and runtime requirements.

These skills are published to [ClawHub](https://clawhub.ai/), the public registry for OpenClaw skills, under their canonical names.

## Layout

```
cli/skills/
├── README.md          # this file
├── publish.sh         # publish helper (used locally and from CI)
├── olares-shared/
│   └── SKILL.md       # foundation: profile model, login, token refresh
├── olares-files/
│   ├── SKILL.md       # cross-cutting concepts + verb index
│   └── references/    # one file per non-trivial subcommand
│       ├── olares-files-ls.md
│       ├── olares-files-upload.md
│       └── ...
├── olares-market/     # olares-cli market
├── olares-settings/   # olares-cli settings
├── olares-dashboard/  # olares-cli dashboard
└── olares-cluster/    # olares-cli cluster (per-user K8s view)
```

`olares-shared` is the foundation — every other skill cross-references it for the profile selection, login, and HTTP 401/403 recovery rules. Always install it first.

ClawHub publishes the entire skill directory (including `references/`), so reference files ship automatically without any change to `publish.sh`.

## Writing style

`SKILL.md` is NOT meant to compete with `olares-cli <command> --help`. The frontmatter already declares `cliHelp:` so the agent knows the authoritative `--help` invocation. The body should only carry what `--help` cannot give:

- **Routing** — when to use this skill vs. siblings
- **Cross-cutting concepts** referenced by ≥2 subcommands (e.g. olares-files' 3-segment frontend path)
- **Client-side hard constraints that bite users** (quirks the GUI enforces, server-side auto-rename traps, …)
- **Error → fix matrix** that is not in `--help`
- **Verb index** — one row per verb pointing at `--help` and (if it exists) `references/<verb>.md`

Each non-trivial subcommand gets a `references/<skill>-<verb>.md` file that adds — on top of `--help` — safety constraints, agent-facing multi-step flows, and common-error troubleshooting tables. Do NOT re-list flag descriptions; trust `--help`.

What to leave out of SKILL.md (and references):

- Per-flag descriptions (in `--help`)
- Source-path citations like `[cli/cmd/ctl/files/path.go](...)` — agents don't review Go source
- Internal package walkthroughs / "Source layout" sections
- "What's NOT here yet" / future-work sections — keep skills focused on current capability

Target sizes: SKILL.md ≤ 250 lines (≤ 300 for the most complex command tree). Each reference: ≤ 150 lines.

## Runtime requirement

Each `SKILL.md` declares:

```yaml
metadata:
  requires:
    bins: ["olares-cli"]
```

ClawHub does **not** install the `olares-cli` binary for you — it is part of every Olares device, so the `bins:` line just gates the skill behind "you must be on a host that has olares-cli on PATH". The binary itself ships through Olares' regular release channels (see [`cli/.goreleaser.yaml`](../.goreleaser.yaml) and [`.github/workflows/release-cli.yaml`](../../.github/workflows/release-cli.yaml)).

## Publishing to ClawHub

### Prerequisites

1. Account on [clawhub.ai](https://clawhub.ai/) with a linked GitHub account that is at least 1 week old (ClawHub anti-abuse policy).
2. **Node.js 22+** (or 20.10+ with `--experimental-import-attributes`). One of `clawhub`'s transitive deps uses ES2025 import attributes (`import x from '...' with { type: 'json' }`); older Node prints `SyntaxError: Unexpected token 'with'` on every command.
3. `clawhub` CLI installed: `npm i -g clawhub`.
4. Either an interactive `clawhub login` session, or a non-interactive token exported as `CLAWHUB_TOKEN`.

### Local validation (no network)

`clawhub skill publish` does not have a `--dry-run` flag. The `--dry-run` mode here is a **local-only** sanity check: parses each `SKILL.md` frontmatter, verifies that `name` matches the folder slug, that `version` is valid semver, and that the required `metadata.requires.bins: ["olares-cli"]` declaration is present. It then prints the `clawhub skill publish` command that would actually run.

```bash
./cli/skills/publish.sh --dry-run                  # validate all 6
./cli/skills/publish.sh --dry-run olares-shared    # validate one
```

### Server-side preview (optional)

For a real "what would the registry do" preview — including remote slug/version conflict checks — use the `sync` subcommand. It enumerates the directory, validates against the live registry schema, and reports new vs. updated skills without uploading:

```bash
# from the repo root:
clawhub sync --workdir cli --dry-run
# or, equivalently:
( cd cli && clawhub sync --dry-run )
```

**WARNING**: `clawhub sync` resolves its scan root as `<workdir>/skills` (default workdir = current directory). If that path doesn't exist it silently falls back to the OpenClaw workspace at `~/.openclaw/skills` (or `D:\openclaw\skills` on Windows), which usually contains community skills installed by OpenClaw itself. Always pass `--workdir cli` (or `cd cli` first) so it scans `cli/skills/` here — otherwise a real (non-dry-run) `sync` would re-upload third-party skills under **your** account.

Note: `clawhub sync` defaults to bumping the patch version on updates. For deterministic releases driven by the `version:` field in each `SKILL.md`, prefer the `publish.sh` (no `--dry-run`) path for actual uploads.

### Publish

```bash
./cli/skills/publish.sh                            # publish all 6
./cli/skills/publish.sh olares-files olares-market # publish a subset
```

Versions come from each skill's frontmatter `version:` field — bump the field before publishing a new release. Slugs and display names are baked into [`publish.sh`](publish.sh).

## Slug policy

The 6 skills publish under their canonical short names:

| Slug              | Display name                                |
|-------------------|---------------------------------------------|
| `olares-shared`   | Olares Shared (olares-cli foundation)       |
| `olares-files`    | Olares Files (olares-cli files)             |
| `olares-market`   | Olares Market (olares-cli market)           |
| `olares-settings` | Olares Settings (olares-cli settings)       |
| `olares-dashboard`| Olares Dashboard (olares-cli dashboard)     |
| `olares-cluster`  | Olares Cluster (olares-cli cluster)         |

If a slug is ever taken on ClawHub, fall back to the `olares-cli-` prefix (e.g. `olares-cli-shared`) and update **all** 5 cross-references to `../olares-shared/SKILL.md` inside the non-shared skills accordingly.
