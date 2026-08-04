# knowledge download settings

> **Flags:** `olares-cli knowledge download settings <verb> --help`.

Read or update **download-server global settings**. Today the manager exposes a
single global knob, `aria2_max_concurrent` (aria2 `max-concurrent-downloads`).
These are server-wide, not per-user, so changing them requires administrator
privileges. The CLI does **not** pre-check this and defers to the server, which
returns an error if the caller is not allowed.

## get

```bash
olares-cli knowledge download settings get
olares-cli knowledge download settings get -o json
```

`GET /api/system/settings`. Field: `aria2_max_concurrent`.

## set

```bash
olares-cli knowledge download settings set --aria2-max-concurrent 5
```

`PUT /api/system/settings`. The manager applies exactly **one key/value pair
per request** (body `{"key":"aria2_max_concurrent","value":5}`), so the CLI
exposes only `--aria2-max-concurrent` (server-validated range `[1, 16]`). The
updated snapshot is printed on success.
