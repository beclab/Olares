# files compress

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files compress --help`.
> **Needs Olares >= 1.12.6** (the archive surface) — see the parent's [version gate](../SKILL.md#version-gate-olares--1126).

Pack one or more remote entries into a single archive. **Asynchronous**: the POST returns a `task_id` and the byte-writing runs on the server's per-node task queue (manage it with `files task`).

```
compress <src>... <dst>
```

`<src>...` are existing remote files/dirs (dirs are recursively included); `<dst>` is the new archive's path (a file — must NOT end with `/`).

## Namespace allow-list

Only `drive/Home`, `drive/Data`, `drive/Common`, `cache/<node>`, `external/<node>/<volume>`. `sync` and all cloud drives (`awss3`/`dropbox`/`google`/`tencent`) are refused — stage into `drive/Home` first, or use the LarePass web app for cloud.

## Formats & constraints

Supported: `zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz`. Derived from `<dst>`'s extension when `--format` is omitted; pass `--format` if the destination has no canonical suffix.

- **Single-file compressors (`gzip` / `bzip2` / `xz`) wrap exactly ONE file** — they cannot pack a directory or multiple sources. Use a container format (`zip` / `7z` / `tar*`) for those.
- **Passwords and split volumes are `zip` / `7z` only.** `--password-stdin` on any other format is refused client-side; for `7z` it also enables header encryption.
- A single source that is already a compressed/archive file is refused (avoids accidental double-compression).

## Key flags

| Flag | Meaning |
|---|---|
| `--format` | Override the format (else derived from `<dst>` extension) |
| `--level N` | Compression level 0..9 (0=store, 9=max); omit for the codec default |
| `--volume-size SIZE` | Split-archive volume size with unit: `100MB` / `1.5GB` (bare number = MiB). zip / 7z only |
| `--volume-size-mb M` | Back-compat raw-MiB alias for `--volume-size` (pass at most one of the two) |
| `--conflict POLICY` | On collision at `<dst>`: `rename` (default) / `overwrite` / `skip` |
| `--preserve-symlinks` | Archive symlinks as symlinks (default: dereference) |
| `--password-stdin` | Read the archive password from STDIN (zip / 7z only) |
| `--wait` | Block until the task finishes, printing progress |
| `--node` | Override the `{node}` segment (default: master node from `/api/nodes/`) |

## Safety constraints

- **Mutates the server (writes a new archive) — confirm intent with the user.**
- **`--conflict overwrite` can clobber an existing `<dst>`** — confirm the destination first.
- **Never pass a password on the command line** — use `--password-stdin` so it doesn't leak into shell history / `ps`.

## Preflight

Every `<src>` is Stat'd (must exist; trailing-slash must match file-vs-dir kind) and `<dst>`'s parent directory must exist BEFORE the task is queued. Create the parent with `files mkdir -p` if needed.

## Examples

```bash
# Two files into a zip.
olares-cli files compress drive/Home/a.pdf drive/Home/b.pdf drive/Home/out.zip

# Whole directory into a tar.gz at max compression, block until done.
olares-cli files compress drive/Home/Photos/ drive/Home/photos.tar.gz --level 9 --wait

# Encrypted 7z with header encryption (password via stdin).
echo "s3cret" | olares-cli files compress drive/Home/Secrets/ drive/Home/secrets.7z --password-stdin

# Split-volume zip (100 MiB volumes).
olares-cli files compress drive/Home/Backups/ drive/Home/backup.zip --volume-size 100MB
```

## Wire shape

```
POST /api/archive/<node>/compress
  body: {sources, destination, format, level?, volumeSizeMB?, preserveSymlinks, conflict}
  headers: X-Archive-Password (zip / 7z only)
  reply: {task_id}
```

## Agent notes

- **Without `--wait` the archive is NOT done when the command returns** — it printed a `task_id`. Poll/manage with `files task ... <task_id> --node <node>`, or re-run with `--wait`.
- **Ctrl-C during `--wait` does NOT cancel the server-side task** — use `files task cancel <task_id> --node <node>`.
- Note the `<node>` printed in the "queued compress task" line; the `task` verbs need it.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| Backend version could not be determined | Profile version cache is missing or stale | Confirm `profile login`, then run `olares-cli profile list --refresh-version` |
| `require Olares >= 1.12.6`, with a detected older version | Backend predates the archive surface | Upgrade Olares |
| `does not support the "sync"/"<cloud>" namespace` | Outside the archive allow-list | Stage into `drive/Home` first |
| `cannot derive --format from destination` | No canonical suffix on `<dst>` | Pass `--format` |
| `only supported on ... (zip, 7z)` | Password / split-volume on a non-zip/7z format | Use zip or 7z, or drop the flag |
| single-file compressor + directory/multi-source | gzip / bzip2 / xz can't pack >1 entry | Use a container format (zip / 7z / tar*) |
| `destination's parent directory ... does not exist` | Parent not pre-created | `files mkdir -p` the parent |
