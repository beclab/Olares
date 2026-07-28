# files nfs

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli files nfs --help`, `olares-cli files nfs mount --help`, `olares-cli files nfs history --help`.
> **Needs Olares >= 1.12.6.**

Mount **external** NFS exports into the per-user files-backend's `external/<node>/...` namespace. The NFS half of the LarePass "Connect to Server" modal — the SMB half is `files smb`.

> **Don't confuse with `files share`** — `share` exposes an OUTBOUND share; `nfs` consumes an INBOUND export (mounts a remote NFS export into Olares).

## How NFS differs from SMB

- **No credentials.** NFS exports mount by address alone — no username / password step.
- **Address shape.** A target is either a bare host/IP (`192.168.1.10`, triggers export discovery) or a full `host:/export` (`192.168.1.10:/data`, mounts directly). (SMB uses `//host/share`.)
- **Shared favorites book.** `nfs history` reads/writes the SAME per-node store as `smb history` (`/api/smb_history/<node>/`); NFS entries are URL-only.

## Sub-commands

| Sub-command | Purpose |
|---|---|
| `nfs mount <host \| host:/export>` | Discover a host's exports (bare host) or mount one directly (full path) |
| `nfs unmount <name>` | Unmount a previously-mounted entry (use the name from `files ls external/<node>/`) |
| `nfs history list` | List NFS-shaped favorites (pass `--all` to include SMB `//` shares too) |
| `nfs history add <host:/export \| host>` | Stash a favorite (URL-only) |
| `nfs history rm <url>...` | Drop favorites by URL |

## Discovery flow (bare host)

A bare host makes the server LIST the host's exports; the CLI prints them as remountable `host:/export` strings and exits non-zero so a script can re-target:

```bash
# Step 1 — bare host triggers discovery.
olares-cli files nfs mount 192.168.1.10
# → server returned 2 export(s): 192.168.1.10:/data, 192.168.1.10:/backups

# Step 2 — re-run with the chosen export.
olares-cli files nfs mount 192.168.1.10:/data
```

## Safety constraints

- **Mount / unmount mutate the per-node state — confirm intent with the user.**
- `nfs history rm` deletes favorites — confirm the URLs.

## Examples

```bash
# Mount a specific export.
olares-cli files nfs mount 192.168.1.10:/data

# Inspect mounted entries, then unmount by the listed name.
olares-cli files ls external/<node>/
olares-cli files nfs unmount nfs-192-168-1-10-data --node <node>

# Favorites (no credentials for NFS).
olares-cli files nfs history add 192.168.1.10:/data
olares-cli files nfs history list
olares-cli files nfs history rm 192.168.1.10:/data
```

## Wire shape (shared with files smb; dispatched by ?external_type=nfs)

```
POST   /api/mount/[<node>/]?external_type=nfs   body: {url}                → mount host:/export
POST   /api/mount/[<node>/]?external_type=nfs   body: {url, operate:"list"} → discover exports
POST   /api/unmount/external/<node>/<name>/?external_type=nfs
GET    /api/smb_history/<node>/                  (history list)
PUT    /api/smb_history/<node>/  body: array     (history upsert)
DELETE /api/smb_history/<node>/  body: array of {url}  (history rm)
```

## Agent notes

- **After mount the entry lives at `external/<node>/<entry>/`** and is consumed by every other `files` verb like any other namespace. `files ls external/<node>/` confirms the actual (sanitized) entry name to pass to `unmount`.
- **A discovery response is NOT an error** — surface the export list verbatim and ask which one to mount; the non-zero exit is just the "pick one and re-run" signal.
- **`--node` is rarely needed** — defaults to the first `/api/nodes/` entry; pass it on multi-node Olares.
- **Don't `mkdir` under `external/<node>/`** — that's [quirk #3](../SKILL.md#3-externalnode-is-a-virtual-volume-listing-layer-read-only); new volumes come from mount.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| Backend version could not be determined | Profile version cache is missing or stale | Confirm `profile login`, then run `olares-cli profile list --refresh-version` |
| `require Olares >= 1.12.6`, with a detected older version | Backend predates NFS support | Upgrade Olares |
| `looks like a URL scheme` / `is an SMB-style path` | `nfs://...` or `//host/share` passed | Use bare host or `host:/export`; use `files smb` for `//` shares |
| `nfs mount returned an export list` | Bare host → discovery | Re-run with one of the printed `host:/export` paths |
| `entry name ... must not contain '/'` | `unmount` got a path, not the entry name | Pass the bare name from `files ls external/<node>/` |
| `host not reachable` | Network / DNS | Verify the host is on the same network |
