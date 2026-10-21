# files ls

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first for the profile model, the 3-segment frontend path, and the 5 client-side quirks.
> **Flags & wire shape:** `olares-cli files ls --help` (single source of truth).

List a directory on the per-user files-backend. Uniform across all 10 namespaces.

## Examples

```bash
olares-cli files ls drive/Home/
olares-cli files ls drive/Home/Documents
olares-cli files ls sync/<repo_id>/
olares-cli files ls awss3/<account>/<bucket>
olares-cli files ls cache/<node>/
olares-cli files ls external/<node>/           # virtual volume-listing layer (see SKILL.md quirk #3)
olares-cli files ls drive/Home/Documents -o json  # raw envelope, pretty-printed
```

## One level, always

`ls` lists the directory you named and stops. There is no recursive flag and no depth option — `--help` shows only `-o` and `-h`. A subtree is walked by calling `ls` again on each directory you got back, so a script that reads one listing and prints "the whole tree" is printing one level and calling it a tree.

Before writing such a walker, ask whether you need one. If you are looking for a file some other tool just produced, that tool's own response says where it is; browsing to a plausible-looking file is not the same as finding the right one.

## Output shape

Default table: `MODE  SIZE  TYPE  MODIFIED  NAME`. Directories sort before files; directory names get a trailing `/`. Empty directories print `(empty)`.

Dot-prefixed entries are listed like any other. `.by-id/` and similar sidecar directories are visible without a flag — nothing here hides them, so their absence from a listing means they are absent from the directory.

`-o json` prints the backend's own envelope, which is not the same shape on every namespace — `olares-cli files ls --help` spells out how the cloud drives differ. The table hides that difference, so code written against one namespace's JSON breaks on another. Two fallbacks cover it:

- Children are `.items[]` on `drive/`, `sync/`, `cache/`, `external/` and `share/`, and `.data[]` on the cloud drives (`awss3/`, `google/`, `dropbox/`, `tencent/`). Read `.items[]` with a fallback to `.data[]`.
- The byte count is `size` on the first group and `fileSize` on the cloud drives. Read `size` with a fallback to `fileSize`. Reading only `size` is why a cloud listing reports every file as 0 bytes — a wrong number, not an error, so nothing tells you.

Cloud drives also drop the parent summary and send `modified` and `mode` as the empty string. Treat an empty `modified` as unknown rather than as the epoch.

## Agent notes

- `ls` is the canonical discovery verb. Use it before any write to confirm parent existence and the exact basename casing.
- `ls external/<node>/` is the right way to discover attached volumes (`hdd1`, `usb1`, `smb-...`) before targeting `external/<node>/<volume>/<sub>/`.
- `ls cache/<node>/` is the right way to discover what's under a node before sharing — share-create rejects bare `cache/<node>/`.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `drive extend must be Home, Data, or Common (got "home")` | `drive/home/...` instead of `drive/Home/...` | Use exact casing: `Home`, `Data`, or `Common` |
| Empty `items` / `data` array on a known-non-empty dir | Wrong identity — the active profile can't see this scope | `olares-cli profile list` and switch with `profile use` |
| Every file reports 0 bytes on a cloud drive | Read `size`, which that namespace does not send | Fall back to `fileSize` |
| 401/403 | Token rotation or invalidation | See [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) |
