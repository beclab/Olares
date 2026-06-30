# files share

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli files share --help` (parent), `olares-cli files share <flavor> --help` (each leaf), `olares-cli files share set-members --help`, …

Create and manage shares for directories. **All three flavors are directory-only** — the CLI Stats the target before posting and refuses files / non-existent paths up front. To share a single file, place it in a dedicated directory and share that.

## Three flavors at a glance

| Flavor | Audience | Recipient model | Update verb |
|---|---|---|---|
| `internal` | Other Olares users on the same node | Olares user names → per-user permission | `set-members` |
| `public` | Anyone with the link + password | Opaque link, recipients open it at `<host>/sharable-link/<id>/` | `set-password` |
| `smb` | Local network (Finder / Explorer / etc.) | SMB-account IDs (managed via `smb-users`), or `--public` for "anyone on the LAN" | `set-smb` |

## Per-flavor namespace allow-list

| Flavor | Allowed | Notes |
|---|---|---|
| `internal` | `drive`, `sync`, `external`, `cache` | Cloud rejected — cross-cloud-account share doesn't work. `external/<node>/` bare-root and `cache/<node>/` bare-root rejected (quirks #3, #5) — point at a sub-path. **`drive/Common` refused** — public-only |
| `smb` | `drive`, `external`, `cache` | Sync rejected (Seafile has its own mount story, not Samba) + cloud rejected. Same bare-root guards as internal. **`drive/Common` refused** — public-only |
| `public` | `drive` ONLY | Tightest of the three. Sync / external / cache / cloud all refused (the GUI restricts Public to drive only). **`drive/Common` is allowed here** — `public` is the only share flavor it supports. Error messages route the user to `share internal` (sync) or `share internal` / `share smb` (external / cache) |

## `--users` format (internal / smb / set-members)

```
name1:perm1,name2:perm2,name3   (perm defaults to "view" if omitted)
```

Permissions: `view` / `upload` / `edit` / `admin` (or `0..4`). Empty perm falls back to `view`.

For SMB shares, "name" is an **SMB-account ID** (see `share smb-users list`), not an Olares user name.

## Wire shape (all create flavors)

```
POST /api/share/share_path/<fileType>/<extend><subPath>/
body: {name, share_type, permission, password, ...}
```

Response carries the new `share id`, plus per-flavor extras (`smb_link` / `smb_user` / `smb_password` for SMB; the Public-link URL is constructed by the LarePass app's `shareBaseUrl + /sharable-link/<id>/` pattern).

Management verbs (`list` / `get` / `rm`) take the share id and are share-type-agnostic.

## Update verbs (REPLACES, not appends)

| Verb | What it changes | Important semantic |
|---|---|---|
| `set-password` | Public-link password | One field; rejects non-Public shares up front |
| `set-members` | Internal share member list | **Drops every member not listed in `--users`.** Pass `--clear` to drop them all. Wire has no "add member"; for additive updates, list every existing member + the new one |
| `set-smb` | SMB account list OR public-SMB toggle | Same replace semantics as `set-members`. `--public` flips to "anyone on the LAN" mode |

## Examples

```bash
# Internal share, two members (alice can edit, bob can view).
olares-cli files share internal drive/Home/Backups/ \
    --users alice:edit,bob:view

# Public link valid 7 days, password auto-generated and printed.
olares-cli files share public drive/Home/Photos/ --expire-days 7

# Public upload-only inbox with explicit password + size cap.
olares-cli files share public drive/Home/Inbox/ --upload-only \
    --password drop --expire-days 30 --upload-size-limit 100M

# SMB share for two SMB users.
olares-cli files share smb drive/Home/Movies/ \
    --users smb-uid-1:edit,smb-uid-2:edit

# Roll a Public link's password.
olares-cli files share set-password <share-id>

# Promote bob from view to admin on an Internal share (carry alice through unchanged!).
olares-cli files share set-members <share-id> \
    --users alice:edit,bob:admin

# Drop every member (share stays, becomes private to its owner).
olares-cli files share set-members <share-id> --clear

# Switch an SMB share to public-SMB.
olares-cli files share set-smb <share-id> --public

# List, inspect, remove.
olares-cli files share list --shared-by-me
olares-cli files share get <share-id>
olares-cli files share rm <share-id>
```

## Public-link specifics

- **Password is required** — either pass `--password <pw>` or let the CLI auto-generate an 8-char random password (which it then prints).
- **Expiration is required** — pass exactly one of `--expire-days N` or `--expire-time <RFC3339>`. Public links without an expiration are not supported by the backend.
- **`--upload-only`** locks recipients out of listing / download — they can only drop files in.
- **`--upload-size-limit`** accepts human-readable sizes: `100M`, `1G`, `500K`, `512` (raw bytes). `0` / omitted = no per-upload cap.

## Agent flows

### Create a Public share and reply with the URL

```bash
SHARE_ID=$(olares-cli files share public drive/Home/Photos/ --expire-days 7 --password "$PW" --json | jq -r '.id')
echo "Share link: https://<your-host>/sharable-link/$SHARE_ID/"
```

The `<your-host>` part is read from the LarePass app's `shareBaseUrl`; if the user doesn't already know it, point them at LarePass settings.

### Add a member to an existing Internal share without dropping existing ones

```bash
# First fetch the current members.
CURRENT=$(olares-cli files share get <share-id> --json | jq -r '.share_members | map(.share_member + ":" + .permission_label) | join(",")')
# Then re-list them PLUS the new member.
olares-cli files share set-members <share-id> --users "$CURRENT,carol:view"
```

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `share target is a file, not a directory` | Tried to share a single file | Wrap it in a directory and share that |
| `Public only supports the drive namespace` | `share public sync/...` / `share public external/...` / etc. | Use `share internal` instead (or `share smb` for external/cache) |
| `cloud namespaces are not supported` | `share <flavor> awss3/...` etc. | Move the data into drive first, then share |
| `drive/Common (the app common data area) only supports outbound public links` | `share internal` / `share smb` on a `drive/Common` path | Use `files share public` (the only flavor Common allows) |
| `refusing to share external/<node>/` | Quirk #3 bare root | Point at `external/<node>/<volume>/<sub>/` |
| `refusing to share cache/<node>/` | Quirk #5 node-picker layer | Point at `cache/<node>/<sub>/` |
| `--users and --clear are mutually exclusive` | Both passed to `set-members` | Pick one |
| `share-type mismatch` (e.g. set-password on Internal) | Wrong update verb for the flavor | Use the matching verb from the flavor table |
