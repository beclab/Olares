# files smb

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli files smb --help` (parent), `olares-cli files smb mount --help`, `olares-cli files smb history --help`.

Mount **external** SMB shares into the per-user files-backend's `external/<node>/...` namespace. CLI counterpart of the LarePass "Connect to Server" modal.

> **Don't confuse with `files share smb`** — that creates an OUTBOUND share (expose a directory over Samba). `files smb` consumes INBOUND shares (mount a network share into Olares).

## Sub-commands

| Sub-command | Purpose |
|---|---|
| `smb mount <smb-url>` | Mount a remote SMB share, materializes at `external/<node>/<entry>/` |
| `smb unmount <name>` | Unmount a previously-mounted entry |
| `smb history list` | List the per-node "Favorite Servers" |
| `smb history add <smb-url>` | Stash a favorite for later (optional credentials) |
| `smb history rm <smb-url>...` | Drop favorites by URL |

## Safety constraints

- **Mount / unmount mutate the per-node state — confirm intent with the user.**
- **Credentials in `-p / --password` end up in shell history.** For scripts use `--password-stdin`; for interactive use, omit both and the CLI prompts without echo.
- **History entries can carry credentials.** Treat them as sensitive — adding credentials to history is convenient but the entries are stored server-side.

## Mount flow with host-only address (discovery)

```
POST /api/mount/[<node>/]?external_type=smb
body: {smbPath, user, password}
reply:
  code 200 → mounted; visible at external/<node>/<entry>/
  code 300 → smbPath was host-only; data is the list of discovered shares
```

When the user passes a HOST-only URL (e.g. `//host.local`), the server returns `code 300` with the discovered shares. The CLI prints the list and asks the user to re-run with one of them:

```bash
# Step 1 — host-only triggers discovery.
olares-cli files smb mount //host.local
# → server returned 3 shares: //host.local/Public, //host.local/Movies, //host.local/Backups

# Step 2 — re-run with the chosen share path.
olares-cli files smb mount //host.local/Public -u alice -p s3cret
```

## Examples

```bash
# Mount with credentials.
olares-cli files smb mount //host.local/Public -u alice -p s3cret

# Mount via stdin password (script-friendly).
printf '%s' "$SMB_PASSWORD" | olares-cli files smb mount //host.local/Public -u alice --password-stdin

# Stash a favorite (credentials optional; prompted at mount time if omitted).
olares-cli files smb history add //host.local/Public

# List favorites for the current node.
olares-cli files smb history list

# Inspect the mounted entries (every external mount is just a child of external/<node>/).
olares-cli files ls external/<node>/

# Unmount when done.
olares-cli files smb unmount <entry-name>

# Remove a favorite by URL.
olares-cli files smb history rm //host.local/Public
```

## Wire shape

```
POST   /api/mount/[<node>/]?external_type=smb            (mount)
POST   /api/unmount/external/<node>/<name>/?external_type=smb  (unmount)
GET    /api/smb_history/<node>/                          (history list)
PUT    /api/smb_history/<node>/  body: array             (history upsert)
DELETE /api/smb_history/<node>/  body: array of {url}    (history rm)
```

## Agent notes

- **After a successful mount, the entry lives at `external/<node>/<entry>/`** and is consumed by every other `files` verb the same way as any other namespace. Use `files ls external/<node>/` to confirm the new entry name (it's usually a sanitized version of the SMB path).
- **`--node` is rarely needed** — defaults to the active node from `/api/nodes/`. Pass it explicitly only if the user has a multi-node Olares and wants the mount to land on a specific node.
- **Mount failures with code 300 are not errors** — they're discovery responses. Surface the share list to the user verbatim and ask them which one to mount.
- **Don't try to mkdir under `external/<node>/`** — that's quirk #3 (virtual layer). New volumes come from mount, not mkdir.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `code 300` returned from mount | host-only URL; server returned discovered shares | Pick one from the list, re-run mount with full `//host/share` path |
| `authentication failed` | Wrong username / password | Check credentials; ensure the SMB server actually accepts them |
| `host not reachable` | Network / DNS issue | Verify the host is on the same network; check `ping <host>` |
| Mount succeeded but the entry isn't at `external/<node>/<expected-name>/` | Backend sanitized the name | `files ls external/<node>/` to find the actual entry name |
| `name already mounted` | The same share was mounted before (possibly under a different node) | `files smb history list`, then `files smb unmount` the stale one |
