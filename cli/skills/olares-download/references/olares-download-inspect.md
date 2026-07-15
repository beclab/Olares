# download inspect & prefs

> **Flags:** `olares-cli download inspect --help`, `olares-cli download prefs get|set --help`.

## inspect

```bash
olares-cli download inspect 'https://www.youtube.com/watch?v=…'
olares-cli download inspect 'https://example.com/file.zip' -o json
```

Returns provider (`yt-dlp` / `aria2` / `huggingface` / …), title, and (for yt-dlp) `available_qualities`. Probe failures often still return HTTP 200 with `Error` / `error_category` set — treat as a hint, not a gate before `create`.

If `Available: false` for yt-dlp, the yt-dlp daemon is unreachable (often not installed). Create for yt-dlp URLs will fail until it is available; aria2 / huggingface URLs are unaffected.

## prefs get / set

Per-(user, app) default yt-dlp quality used when `create` omits `--quality` / `--format-id`.

```bash
olares-cli download prefs get --app wise
olares-cli download prefs set --app wise --quality 1080p
```

Allowed `--quality` values: `best`, `2160p`, `1080p`, `720p`, `480p`, `360p`, `audio`. Empty is not valid on set — use `best` for “no override”.
