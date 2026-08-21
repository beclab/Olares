# knowledge download inspect & prefs

> **Flags:** `olares-cli knowledge download inspect --help`, `olares-cli knowledge download prefs get|set --help`.

## inspect

```bash
olares-cli knowledge download inspect 'https://www.youtube.com/watch?v=…'
olares-cli knowledge download inspect 'https://example.com/file.zip' -o json
```

Returns provider (`yt-dlp` / `aria2` / `huggingface` / …), title, and (for yt-dlp) `available_qualities`. Probe failures often still return HTTP 200 with `Error` / `error_category` set — treat as a hint, not a gate before `create`. Inspect itself still exits 0.

Read `Available` and `Error` separately:

- `Available: false` (or an error mentioning yt-dlp unreachable / not installed) means the yt-dlp daemon is down. Create for yt-dlp URLs will fail until it is available (the CLI prints the market install command); aria2 / huggingface URLs are unaffected.
- `Available: true` with `Error: daemon returned code 505 (network)` means the daemon is up and the probe failed. 505 is a coarse bucket (it can wrap a site 4xx such as HTTP 412). That is **not** “daemon crashed”, and it is **not** a stop before `create`. Login walls often land here as `network` instead of `authorization_failed`, so the cookie CTA below may be absent — still import a Netscape `cookies.txt` and retry inspect.

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

The CLI prints this command only for those codes. Inspect 505 / create HTTP 412 / a waited `err_category` of `network_error` do **not** print it — still try cookies when the URL is a logged-in site. Prefer a Netscape `cookies.txt` from a cookie exporter. A `Cookie:` header import can succeed in the store and still 412 on create if the site's session is not in that line. Use the [`olares-settings`](../../olares-settings/SKILL.md) `header` path only when a Netscape export is missing the login (typical for YouTube httpOnly cookies).

## prefs get / set

Per-(user, app) default yt-dlp quality used when `create` omits `--quality` / `--format-id`.

```bash
olares-cli knowledge download prefs get --app wise
olares-cli knowledge download prefs set --app wise --quality 1080p
```

Allowed `--quality` values: `best`, `2160p`, `1080p`, `720p`, `480p`, `360p`, `audio`. Empty is not valid on set — use `best` for “no override”.
