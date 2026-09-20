# knowledge download file tools

> **Flags:** `olares-cli knowledge download file <verb> --help`.

Two different addressing schemes live here: `exists` pre-checks a **URL**
before create (url family), while `remove` operates on an existing
**resource path** on the PVC (`drive/Home/...`). Identity is always the
gateway-injected `X-Bfl-User`; the CLI never sets the user.

## exists (URL pre-check)

```bash
olares-cli knowledge download file exists 'https://host/big.zip' --path drive/Home/Downloads/
olares-cli knowledge download file exists 'https://host/v?a=1&b=2' -o json
olares-cli knowledge download file exists 'https://huggingface.co/org/model' --hf-dest cache
```

`GET /api/url/file-exists`. Quote URLs containing `?`, `&` or `=`. The
server resolves the target name from the URL (or `--name`) under `--path`
for `--app` and reports `Exists` plus a `Conflict` path when it collides.

`--hf-dest` is optional for HuggingFace URLs: `cache` or `local` (query
`hf_dest`). Omit it to keep the server default.

## remove (resource path)

```bash
olares-cli knowledge download file remove --path drive/Home/Downloads/clip.mp4
```

`DELETE /api/download/file_remove`. Deletes the file-manager resource; a
path that does not exist is still reported as success. Prints
`removed <path>`.
