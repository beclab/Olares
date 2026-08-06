# files archive

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli files archive --help`, `olares-cli files archive entries --help`, `olares-cli files archive cat --help`.
> **Needs Olares >= 1.12.6** (the archive surface) — see the parent's [version gate](../SKILL.md#version-gate-olares--1126).

The **read-only** counterpart to `compress` / `extract`: inspect an archive WITHOUT unpacking it. Both sub-verbs are synchronous (streaming), unlike compress/extract.

## Sub-commands

| Sub-command | Purpose |
|---|---|
| `archive entries <archive>` | Stream the archive's entry list (table by default; `--json` for NDJSON) |
| `archive cat <archive> <inner-path>` | Stream one member's bytes to stdout (or to a file via `-o`) |

## Namespace allow-list

Same as compress / extract: `drive/Home`, `drive/Data`, `drive/Common`, `cache/<node>`, `external/<node>/<volume>`. `sync` + all cloud refused.

## Preview / read constraints

- **Bare single-stream compressors (`bzip2` / `xz`) have no listable entry table** — `entries` and `cat` are refused for them. To get their single payload, `files extract` the archive instead. The `tar.*` compounds (`tar.gz` / `tar.bz2` / `tar.xz` / `tgz`) ARE real tar containers and remain fully inspectable.
- Format is inferred server-side from the source extension; `--format` is accepted locally only to pre-validate flag combinations (e.g. `--password-stdin` only on zip / 7z).

## Key flags

| Flag | Applies to | Meaning |
|---|---|---|
| `--json` | `entries` | One JSON object per line (`path` / `size` / `modified` / `is_dir` / `encrypted`) |
| `--max-entries N` | `entries` | Stop after N entries (0 = no limit); head-style preview of huge archives |
| `-o, --output FILE` | `cat` | Write the member's bytes to a local file (atomic tmp+rename) instead of stdout |
| `--password-stdin` | both | Read the password from STDIN (zip / 7z only) |
| `--format` | both | Override format detection |
| `--node` | both | Override the `{node}` segment |

## Examples

```bash
# Tabular preview.
olares-cli files archive entries drive/Home/Backups/2026-Q1.zip

# JSON pipeline.
olares-cli files archive entries drive/Home/Backups/2026-Q1.zip --json | jq '.path'

# Head-style preview of a huge archive.
olares-cli files archive entries drive/Home/Backups/huge.7z --max-entries 50 --password-stdin

# Dump one member to stdout (binary-safe).
olares-cli files archive cat drive/Home/Backups/2026-Q1.zip notes.md

# Save one member to a local file.
olares-cli files archive cat drive/Home/Vault/data.7z bin/payload --password-stdin -o ./payload < pw.txt
```

## Wire shape

```
GET /api/archive/<node>/entries?source=<archive>   (Content-Type: application/x-ndjson)
GET /api/archive/<node>/entry?source=<archive>&path=<inner-path>  (application/octet-stream)
```

## Agent notes

- **`cat` is binary-safe** — pipe into `less` / `hexdump` / `head -c`. In default (no `-o`) mode stdout stays a clean byte stream, so a status line is printed only with `-o`.
- **`entries --json` keeps stdout NDJSON-clean** — the truncation notice for `--max-entries` goes to stderr in JSON mode.
- **`<inner-path>` for `cat` is the in-archive path** (leading `/` is normalized away). Use `archive entries` first to discover member paths.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| Backend version could not be determined | Profile version cache is missing or stale | Confirm `profile login`, then run `olares-cli profile list --refresh-version` |
| `require Olares >= 1.12.6`, with a detected older version | Backend predates the archive surface | Upgrade Olares |
| `previewing "bzip2"/"xz" archives is not supported` | Raw single-stream compressor, no entries | `files extract` it instead |
| `entry not found inside the archive` | Wrong `<inner-path>` | `files archive entries` to list members |
| `archive requires a password` / `password is incorrect` | Encrypted zip / 7z | Supply via `--password-stdin` |
| `entry exceeds the server's single-shot read limit (HTTP 413)` | Member too large for `cat` | `files extract` the whole archive instead |
