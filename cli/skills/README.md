# olares-cli skills

LLM-readable skills (one folder per command tree) that teach OpenClaw — and any other Claude-style agent — how to drive `olares-cli` against a live Olares instance. Each folder contains a single `SKILL.md` whose YAML frontmatter declares the skill's name, version, and runtime requirements.

These skills are published to [ClawHub](https://clawhub.ai/), the public registry for OpenClaw skills, under their canonical names.

## Layout

```
cli/skills/
├── README.md          # this file
├── publish.sh         # publish helper (used locally and from CI)
├── olares-shared/
│   ├── SKILL.md       # foundation: profile model, login, token refresh
│   └── references/
│       └── olares-platform.md   # cross-skill platform model (storage, uid 1000, namespaces, middleware, versions)
├── olares-files/
│   ├── SKILL.md       # cross-cutting concepts + verb index
│   └── references/    # one file per non-trivial subcommand
│       ├── olares-files-ls.md
│       ├── olares-files-upload.md
│       └── ...
├── olares-knowledge/  # olares-cli knowledge (download-server task centre under knowledge download; Olares >= 1.12.7)
│   ├── SKILL.md
│   └── references/    # knowledge download lifecycle / inspect+prefs
├── olares-search/     # olares-cli search (Desktop global / full-content search; single leaf command)
├── olares-market/     # olares-cli market
├── olares-settings/   # olares-cli settings
├── olares-dashboard/  # olares-cli dashboard
├── olares-cluster/    # olares-cli cluster (per-user K8s view)
├── olares-doctor/     # runtime diagnosis (thin router over cluster / market / dashboard)
│   ├── SKILL.md
│   └── references/    # one file per symptom (app stuck / crash / image / unhealthy / resources)
├── olares-chart/      # olares-cli chart (chart authoring + deploy to your Olares)
│   ├── SKILL.md
│   └── references/    # one file per refinement area / capability
└── olares-publish/    # public Olares Market distribution (beclab/apps PR, paid apps)
    ├── SKILL.md
    └── references/    # market-ready targets / submit / paid-apps
```

**Install the whole suite together.** These skills are designed as one set and cross-reference each other by relative path (e.g. `../olares-shared/SKILL.md`, `../olares-shared/references/olares-platform.md`). Installing only a subset leaves those links dangling. `olares-shared` is the foundation — every runtime skill cross-references it for profile selection, login, and HTTP 401/403 recovery, **and it hosts the cross-skill platform model** ([`olares-shared/references/olares-platform.md`](olares-shared/references/olares-platform.md)) that `files` / `chart` / `cluster` link to (one hop from their `SKILL.md`) instead of re-describing it.

`olares-chart` is a partial exception on **login**, not on linking: its authoring verbs (`from-compose` / `lint` / `package`) are local-only and need **no profile / login / cluster**, so it never logs in to author a chart. It still reads `olares-platform.md` for platform facts (no login needed) and only requires `olares-shared` login when **deploying a chart to a real Olares** (`market upload` + `install`). `olares-publish` is the public-distribution counterpart: it picks up after the app already runs locally (via `olares-chart`) and covers market-ready polish, the `beclab/apps` PR, and paid apps.

ClawHub publishes the entire skill directory (including `references/`), so reference files ship automatically without any change to `publish.sh`.

## Writing style

`SKILL.md` is NOT meant to compete with `olares-cli <command> --help`. The body references `olares-cli <command> --help` for authoritative flag syntax. The body should only carry what `--help` cannot give:

- **When to use** — this skill's scope / trigger phrases, plus the suite-map pointer for outbound routing (see "Cross-skill shared concepts")
- **Cross-cutting concepts** referenced by ≥2 subcommands (e.g. olares-files' 3-segment frontend path)
- **Client-side hard constraints that bite users** (quirks the GUI enforces, server-side auto-rename traps, …)
- **Error → fix matrix** that is not in `--help`
- **Verb index** — one row per verb pointing at `--help` and (if it exists) `references/<verb>.md`
- **Ground every behavioral claim in the implementation.** Before documenting what a verb / flag / error does, confirm it against the code (status derivation, typed error names, retry / timeout, argument arity). Model agent stop / continue rules on the tool's real control flow — typed errors, auto-retry, transient vs terminal — not on a surface status string. Keep this verification in your process only: still **no Go source-path citations** in the shipped skill (see "What to leave out" below).
- **One representation per fact within a file.** Don't place two tables / sections that encode the same thing. When a decision derives from a fact table, express it as prose that points at that table, not a second parallel table.

Each non-trivial subcommand gets a `references/<skill>-<verb>.md` file that adds — on top of `--help` — safety constraints, agent-facing multi-step flows, and common-error troubleshooting tables. Do NOT re-list flag descriptions; trust `--help`.

What to leave out of SKILL.md (and references):

- Per-flag descriptions (in `--help`)
- Source-path citations like `cli/cmd/ctl/files/path.go` — agents don't review Go source
- Internal package walkthroughs / "Source layout" sections
- "What's NOT here yet" / future-work sections — keep skills focused on current capability

Target sizes: SKILL.md ≤ 250 lines (≤ 300 for the most complex command tree). Each reference: ≤ 150 lines.

**Reference depth (one level deep):** all reference files link **directly from a `SKILL.md`**, never reference→reference. A concept buried two hops down is unreliable. Keep each reference short enough to be read whole (≤ 150 lines); when one outgrows that, split it into sibling references rather than adding an in-file table of contents — the `##` headings already are the structure, so a TOC just duplicates them and spends the line budget.

These rules apply to **every** skill, including `olares-chart` — even though it is a local-only chart-authoring skill (no live profile / login) rather than a CLI-driving one, it still avoids Go source-path citations and keeps each reference ≤ 150 lines.

## Cross-skill shared concepts (single source of truth)

Facts used by **≥2 skills** are defined **once** and linked, never copied. This keeps the suite consistent and token-efficient (the lark-cli pattern these skills are modeled on; see also [Anthropic's skill-authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)).

- **One canonical home per shared concept.** The cross-skill Olares platform model (userspace storage, uid-1000 run identity, app/namespace & networking, system middleware, version/semver) lives once in [`olares-shared/references/olares-platform.md`](olares-shared/references/olares-platform.md). The application state machine (states, transitions, fail TTLs, single-download serialization, `running` semantics, progress caveats) lives once in [`olares-shared/references/olares-platform-appstate.md`](olares-shared/references/olares-platform-appstate.md). Auth/profile facts likewise live once in `olares-shared/SKILL.md`.
- **A cross-skill decision or rule is a shared concept too — give it one named, linkable home and link to it; never re-derive it per skill.** Example: the proceed-vs-stop auth check lives once as the `## Auth-readiness gate` in `olares-shared/SKILL.md`; chart / settings / files link to that anchor instead of re-listing `logged-in / expired / invalidated / never`. A rule copied into N skills drifts in N places — an `expired`-handling fix once meant editing four chart files because each had re-derived the same decision.
- **Link the shared source one hop from each consumer's `SKILL.md`** with a short must-read prerequisite (`> Platform model (read once): … see ../olares-shared/references/olares-platform.md`). This is the lark-cli `CRITICAL — MUST Read ../lark-shared/SKILL.md` convention.
- **Do NOT deep-link the shared source from reference files.** A reference that points at another skill's reference is a two-hop nested read. Instead, refer to the shared concept **by name** (matching the source's heading) and rely on the `SKILL.md` prerequisite to have loaded it. (See the chart references, which name "the platform **Userspace storage model**" rather than linking it.)
- **Self-containment is traded for a suite contract.** Strictly, Skills are self-contained and "cannot reference files in other skill folders". We deliberately cross-link because these skills **ship and install as one suite** (stated under Layout). A standalone install leaves cross-skill links dangling — that is the documented trade-off, not an accident.
- **When a fact is genuinely two skills' own angle, let each keep its own framing.** `files` describes the storage areas as *addressing* (`drive/Home`), `chart` as *mounting* (`.Values.userspace.appData`). That is not duplication to dedupe — only the underlying platform facts (backends, durability, uid, version gates) are centralized in `olares-platform.md`.
- **Routing has one source of truth too: the Skill suite map** in [`olares-shared/SKILL.md`](olares-shared/SKILL.md). The canonical intent->skill scope for the whole suite lives there once. Each skill folds routing into its `## When to use` section — trigger phrases, then the suite-map pointer (`> Anything outside this scope -> see the Skill suite map …`), plus an optional `> Mental model` one-liner — with no separate `## Routing` section and no per-sibling ✅/❌ rows. `olares-shared` is a must-read prerequisite for the runtime skills, so the map is always already in context.

## Runtime requirement

Each `SKILL.md` declares:

```yaml
compatibility: Requires olares-cli on PATH
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
```

`description` must stay ≤ 1024 characters (OpenCode limit). Put detailed trigger phrases in the skill body's `## When to use` section.

ClawHub does **not** install the `olares-cli` binary for you — it is part of every Olares device, so the `bins:` line just gates the skill behind "you must be on a host that has olares-cli on PATH". The binary itself ships through Olares' regular release channels (see [`cli/.goreleaser.yaml`](../.goreleaser.yaml) and [`.github/workflows/release-cli.yaml`](../../.github/workflows/release-cli.yaml)).

## Publishing to ClawHub

### Prerequisites

1. Account on [clawhub.ai](https://clawhub.ai/) with a linked GitHub account that is at least 1 week old (ClawHub anti-abuse policy).
2. **Node.js 22+** (or 20.10+ with `--experimental-import-attributes`). One of `clawhub`'s transitive deps uses ES2025 import attributes (`import x from '...' with { type: 'json' }`); older Node prints `SyntaxError: Unexpected token 'with'` on every command.
3. `clawhub` CLI installed: `npm i -g clawhub`.
4. Either an interactive `clawhub login` session, or a non-interactive token exported as `CLAWHUB_TOKEN`.

### Local validation (no network)

`clawhub skill publish` does not have a `--dry-run` flag. The `--dry-run` mode here is a **local-only** sanity check: parses each `SKILL.md` frontmatter, verifies that `name` matches the folder slug, that `version` is valid semver, that `description` is ≤ 1024 characters, and that `metadata.openclaw.requires.bins` includes `olares-cli`. It then prints the `clawhub skill publish` command that would actually run.

```bash
./cli/skills/publish.sh --dry-run                  # validate all 11
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
./cli/skills/publish.sh                            # publish all 11
./cli/skills/publish.sh olares-files olares-market # publish a subset
```

Versions come from each skill's frontmatter `version:` field — bump the field before publishing a new release. Slugs and display names are baked into [`publish.sh`](publish.sh).

## Slug policy

The 11 skills publish under their canonical short names:

| Slug              | Display name                                |
|-------------------|---------------------------------------------|
| `olares-shared`   | Olares Shared (olares-cli foundation)       |
| `olares-files`    | Olares Files (olares-cli files)             |
| `olares-knowledge`| Olares Knowledge (olares-cli knowledge)     |
| `olares-search`   | Olares Search (olares-cli search)           |
| `olares-market`   | Olares Market (olares-cli market)           |
| `olares-settings` | Olares Settings (olares-cli settings)       |
| `olares-dashboard`| Olares Dashboard (olares-cli dashboard)     |
| `olares-cluster`  | Olares Cluster (olares-cli cluster)         |
| `olares-doctor`   | Olares Doctor (runtime diagnosis)           |
| `olares-chart`    | Olares Chart (olares-cli chart)             |
| `olares-publish`  | Olares Publish (Olares Market distribution) |

If a slug is ever taken on ClawHub, fall back to the `olares-cli-` prefix (e.g. `olares-cli-shared`) and update **all** cross-references to `../olares-shared/SKILL.md` inside the runtime skills accordingly.
