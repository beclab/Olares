# files edit

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files edit --help`.

Edit a single existing file in-place by opening it in `$EDITOR`. **UPDATE-only verb** — the server-side `PUT /api/resources/<encPath>` handler is wired as "replace bytes of an existing file"; it does NOT create new files. To materialize a new file, use `files upload` from a local source.

## Safety constraints

- **Interactive TTY required.** `edit` spawns `$EDITOR` foreground; CI / pipes / heredocs are refused cleanly with a `download` + `upload` recovery hint.
- **Three-tier size cap** (default 1 MiB, configurable via `--max-size`): pre-fetch Stat, during-fetch `io.LimitReader` defense, post-edit local file size.
- **Text-only guard by default**: an extension deny-list (jpg/png/gif/heic/pdf/docx/mp4/mp3/zip/tar.gz/exe/so/sqlite/ttf/...) plus a NUL-byte sniff over the first 8 KiB. Pass `--allow-binary` to disable both.
- **No ETag / If-Match support on the wire** — concurrent edits from two clients follow last-writer-wins. Same as the LarePass GUI.

## Supported namespaces

`drive/Home/<sub>/<file>`, `drive/Data/<sub>/<file>`, `sync/<repo_id>/<sub>/<file>`, `cache/<node>/<sub>/<file>`, `external/<node>/<volume>/<sub>/<file>`.

**Cloud drives (awss3 / google / dropbox / tencent) are refused** — the PUT writeback shape is not wire-verified per cloud driver. Use this workflow instead:

```bash
olares-cli files download <cloud-path> <local>
$EDITOR <local>
olares-cli files upload <local> <cloud-path>
```

## Examples

```bash
olares-cli files edit drive/Home/Documents/notes.md
olares-cli files edit drive/Home/.config/app.yaml --editor nano
olares-cli files edit sync/<repo_id>/Notes/draft.md
olares-cli files edit drive/Home/Logs/today.log --max-size 5242880  # 5 MiB
olares-cli files edit external/<node>/usb1/config.json --keep-temp
```

## Editor cascade

Matches `git commit` / `crontab -e`:

```
--editor flag  →  $VISUAL  →  $EDITOR  →  vi (POSIX) / notepad (Windows)
```

The binary is resolved up-front BEFORE the CLI dials the server. A missing / mistyped editor fails fast without pulling the remote file.

## Agent notes

- If the user exits the editor without changes, **no PUT is issued** (byte-for-byte comparison; robust against editors that always rewrite). No warning needed.
- **Concurrent-delete detection**: if Stat says the file exists but the subsequent GET returns 404, the verb refuses with `file disappeared between stat and fetch`. Do NOT retry — re-pull the parent and ask the user what to do.
- `--keep-temp` is the right escape hatch when an unexpected size-cap rejection or NUL-byte sniff blocks the writeback — point the user at the temp path so they can recover bytes manually.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `stdin/stdout is not a terminal` | Non-TTY context | Use `download` + local edit + `upload` |
| `file too large: <N> bytes > max-size <M>` | Three-tier cap pre-fetch | Pass `--max-size 0` (unbounded) or `--max-size <bigger>` |
| `binary content detected (extension/NUL)` | Text-only guard | Confirm intent, then `--allow-binary` |
| `file disappeared between stat and fetch` | Someone else deleted the file between probes | Re-pull parent, ask user |
| `cloud drive edit is not supported` | awss3/google/dropbox/tencent target | Use download → edit → upload workflow |
