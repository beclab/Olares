# knowledge download inspect & prefs

> **Flags:** `olares-cli knowledge download inspect --help`, `olares-cli knowledge download prefs get|set --help`.

## inspect

```bash
olares-cli knowledge download inspect 'https://www.youtube.com/watch?v=…'
olares-cli knowledge download inspect 'https://example.com/file.zip' -o json
```

Returns provider (`yt-dlp` / `aria2` / `huggingface` / …), title, and (for yt-dlp) `available_qualities`. Probe failures often still return HTTP 200 with `Error` / `error_category` set — treat as a hint, not a gate before `create`.

If `Available: false` for yt-dlp, the yt-dlp daemon is unreachable (often not installed). Create for yt-dlp URLs will fail until it is available; aria2 / huggingface URLs are unaffected.

## When the URL needs a login

These signals mean the URL is downloadable but the server has no session for it — not that the URL is bad:

| Signal | Where it shows up |
|---|---|
| `error_code` 501 | inspect data, or the create response |
| `error_code` 507 / 511 / 512 | inspect data |
| `error_category` `authorization_failed` / `private_resource` / `bot_detected` | inspect data, or `info` on a failed task |

Do not stop here. Cookies are an [`olares-settings`](../../olares-settings/SKILL.md) concern; import them and retry:

```bash
olares-cli settings integration cookie import --domain youtube.com --file cookies.txt
olares-cli knowledge download inspect 'https://www.youtube.com/watch?v=…'
```

The CLI prints this command for you, with the domain already filled in from the URL. If the user's `cookies.txt` export turns out to be missing the login, open [`olares-settings`](../../olares-settings/SKILL.md) and use the `header` import path under `settings integration cookie` — that is the fix.

## prefs get / set

Per-(user, app) default yt-dlp quality used when `create` omits `--quality` / `--format-id`.

```bash
olares-cli knowledge download prefs get --app wise
olares-cli knowledge download prefs set --app wise --quality 1080p
```

Allowed `--quality` values: `best`, `2160p`, `1080p`, `720p`, `480p`, `360p`, `audio`. Empty is not valid on set — use `best` for “no override”.
