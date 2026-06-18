# files extract

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files extract --help`.
> **Needs Olares >= 1.12.6** (the archive surface) — see the parent's [version gate](../SKILL.md#version-gate-olares--1126).

Unpack one remote archive into one remote directory. **Asynchronous**: the POST returns a `task_id` and the writing runs on the server's per-node task queue (manage it with `files task`).

```
extract <archive> <dst-dir>/
```

`<archive>` is the existing archive file; **`<dst-dir>/` MUST end with `/`** (drop-into-directory mode, mirroring `cp` / `mv`). The destination dir may already exist (conflict policy decides) or be created on the fly; its parent must exist.

## Namespace allow-list

Same as `compress`: `drive/Home`, `drive/Data`, `drive/Common`, `cache/<node>`, `external/<node>/<volume>`. `sync` + all cloud refused.

## Formats

Supported: `zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz`. Derived from `<archive>`'s extension (`.zip` / `.7z` / `.zip.001` / `.tar.gz` / ...) when `--format` is omitted.

## Key flags

| Flag | Meaning |
|---|---|
| `--format` | Override the format (else derived from `<archive>` extension) |
| `--conflict POLICY` | On collision at the destination: `rename` (default) / `overwrite` / `skip` |
| `--preserve-symlinks` | Land symlinks as symlinks (default: dereference) |
| `--password-stdin` | Read the password from STDIN (zip / 7z only) |
| `--wait` | Block until the task finishes, printing progress |
| `--node` | Override the `{node}` segment |

## Password handling

Encrypted archives surface as an HTTP 400 BEFORE the task is queued. On a TTY the CLI prompts for the password and retries (up to 5 attempts); in a non-TTY context supply it via `--password-stdin`.

## Safety constraints

- **Mutates the server (writes files into `<dst-dir>/`) — confirm intent with the user.**
- **`--conflict overwrite` can clobber existing entries** in the destination — confirm first.

## Preflight

The archive must exist and be a file (not a directory); the destination's parent directory must exist. The leaf dir itself is auto-created by the extract writer.

## Examples

```bash
# Unpack a zip into a sibling directory.
olares-cli files extract drive/Home/Backups/2026-Q1.zip drive/Home/Backups/2026-Q1/

# Encrypted 7z, blocking until done.
echo "s3cret" | olares-cli files extract drive/Home/Vault/data.7z drive/Home/Vault/unpacked/ --password-stdin --wait

# Overwrite colliding entries.
olares-cli files extract drive/Home/Backups/2026-Q1.zip drive/Home/Backups/2026-Q1/ --conflict overwrite
```

## Wire shape

```
POST /api/archive/<node>/extract
  body: {source, destination, format, preserveSymlinks, conflict}
  headers: X-Archive-Password (zip / 7z only)
  reply: {task_id}
```

## Agent notes

- **Without `--wait` the extraction is NOT done when the command returns** — manage via `files task ... <task_id> --node <node>`.
- **Ctrl-C during `--wait` does NOT cancel the server-side task** — use `files task cancel`.
- To preview an archive WITHOUT unpacking, use `files archive entries` / `files archive cat`.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `require Olares >= 1.12.6` | Backend predates the archive surface | Upgrade, or `--olares-version` |
| `must end with '/' to declare directory intent` | `<dst-dir>` had no trailing slash | Add `/` |
| `archive ... does not exist on the server` | Wrong source path | `files ls` the parent and confirm |
| `is a directory on the server, not a file` | `<archive>` pointed at a dir | Point at the archive file |
| `archive requires a password` / `password is incorrect` | Encrypted zip / 7z | Supply via `--password-stdin` (or the interactive prompt) |
| `a multi-volume archive's part is missing` | Split archive missing a `.z01`/`.001` part | Upload all parts next to the main archive |
