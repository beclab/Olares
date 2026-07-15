# download lifecycle

> **Flags:** `olares-cli download create|list|info|pause|resume|cancel|remove --help`.

## create

```bash
olares-cli download create 'https://example.com/video' --app wise
olares-cli download create 'https://…' --path drive/Home/Downloads/ --name clip.mp4 --quality 1080p
olares-cli download create 'https://…' --format-id 'bv*+ba/b' -o json
```

- `--quality` → `extra.ytdlp_quality`; `--format-id` → `extra.format_id`.
- `--extra` is a JSON object of string values merged into `extra`. `--quality` / `--format-id` are applied after and override matching keys.
- `--path` accepts `drive/Home/...` / `drive/Data/...` or a files resource URL. Empty path lets the server pick a default for some providers.
- Success table line: `Created task <id> status=… provider=… name=…`. Use `-o json` for the full task row.

## list / info

```bash
olares-cli download list --app wise
olares-cli download list --status downloading --page 1 --page-size 20 -o json
olares-cli download info 42
```

Table columns: `ID`, `STATUS`, `PROVIDER`, `PERCENT`, `NAME`, `APP`, `UPDATED`. Footer shows `N of total` when the server returns `total`.

## pause / resume / cancel

```bash
olares-cli download pause 42
olares-cli download resume 42
olares-cli download cancel 42
```

No body. 409 means the task is in the yt-dlp mover phase — wait and retry.

## remove

```bash
olares-cli download remove 42
olares-cli download remove 42 --remove-file
```

`--remove-file` sets `remove_flag=true` (delete artefact on PVC). Default keeps the file and only drops the task row.
