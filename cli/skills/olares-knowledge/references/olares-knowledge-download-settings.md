# knowledge download settings

> **Flags:** `olares-cli knowledge download settings <verb> --help`.

Read or update **download-server global settings** (aria2 / yt-dlp concurrency
and seeding limits). These are server-wide, not per-user, so changing them may
require administrator privileges. The CLI does **not** pre-check this and
defers to the server, which returns an error if the caller is not allowed.

## get

```bash
olares-cli knowledge download settings get
olares-cli knowledge download settings get -o json
```

`GET /api/system/settings`. Fields: `aria2_max_concurrent`,
`aria2_max_conn_per_server`, `aria2_split`, `ytdlp_concurrent`,
`seed_ratio_limit`, `seed_time_limit` (seconds).

## set

```bash
olares-cli knowledge download settings set --aria2-max-concurrent 5
olares-cli knowledge download settings set --ytdlp-concurrent 2 --seed-ratio-limit 1.5
```

`PUT /api/system/settings` as a **partial patch**: only the flags you pass are
sent, and unset fields are left unchanged. Provide at least one flag, otherwise
the command errors. The updated snapshot is printed on success.
