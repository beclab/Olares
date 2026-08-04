# knowledge download file tools

> **Flags:** `olares-cli knowledge download file <verb> --help`.

Two different addressing schemes live here: `exists` pre-checks a **URL**
before create (url family), while `check` / `remove` operate on an existing
**resource path** on the PVC (`drive/Home/...`). Identity is always the
gateway-injected `X-Bfl-User`; the CLI never sets the user.

## exists (URL pre-check)

```bash
olares-cli knowledge download file exists 'https://host/big.zip' --path drive/Home/Downloads/
olares-cli knowledge download file exists 'https://host/v?a=1&b=2' -o json
```

`GET /api/url/file-exists`. Quote URLs containing `?`, `&` or `=`. The
server resolves the target name from the URL (or `--name`) under `--path`
for `--app` and reports `Exists` plus a `Conflict` path when it collides.

## check (resource path)

```bash
olares-cli knowledge download file check --path drive/Home/Downloads/clip.mp4
```

`GET /api/download/file_check`. `--path` is a file-manager resource path.
The success body is top-level `{code, exist}` (note `exist`, not the usual
`data` wrapper). Table prints `Exist: <bool>`.

## remove (resource path)

```bash
olares-cli knowledge download file remove --path drive/Home/Downloads/clip.mp4
```

`DELETE /api/download/file_remove`. Deletes the file-manager resource; a
path that does not exist is still reported as success. Prints
`removed <path>`.
