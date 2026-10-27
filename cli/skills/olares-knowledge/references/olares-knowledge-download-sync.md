# knowledge download sync

> **Flags:** `olares-cli knowledge download unfinished --help`, `... sync --help`.

## unfinished

```bash
olares-cli knowledge download unfinished
olares-cli knowledge download unfinished --provider yt-dlp -o json
```

Lists tasks that are **not** in a terminal state (downloading / pending /
paused / …), `GET /api/download/unfinished`. Same table columns as `list`.

The server endpoint requires a single `provider` query param (one of:
`yt-dlp`, `aria2`, `huggingface`) and does **not** accept an `app`
filter. Pass `--provider` to scope to one; omit it and the CLI fans out
one request per provider and merges the result (sorted by `updated_at`).
Only your own tasks are returned.

## sync

```bash
olares-cli knowledge download sync                                      # full drain from the start
olares-cli knowledge download sync --since 2026-07-20T14:00                # local time, matches the table column
olares-cli knowledge download sync --since "2026-07-20 14:00" --since-id 128
olares-cli knowledge download sync --since 2026-07-20T06:00:00Z         # zoned RFC3339 (UTC)
olares-cli knowledge download sync --all -o json                       # drain every page
```

`GET /api/download/sync` is an **incremental pull keyed on a composite
`(updated_at, id)` cursor**. It returns tasks whose `updated_at` is newer than
`--since`, or equal to `--since` with an id greater than `--since-id`, in
`(updated_at, id)` ascending order, **including finished ones**. Because the
cursor is `updated_at`, sync **does** surface progress updates to tasks you
already saw (any change bumps `updated_at`). To follow changes over time,
remember the cursor of the last row you saw (printed after each page) and pass
it back as `--since` / `--since-id`.

`--since` accepts the **local time** shown in the table (`2026-07-20T14:00`, or
`"2026-07-20 14:00"` — quote it because of the space), a bare date
(`2026-07-20`), or a zoned RFC3339 value (`2026-07-20T06:00:00Z`). A value
**without a zone is read in your local timezone**, so you can paste what the
`UPDATED` column shows without doing any UTC conversion. Omit `--since` for a
full drain from the beginning. `--since-id` is the id tie-breaker for rows whose
`updated_at` equals `--since`.

> The `UPDATED` column in `list` / `sync` is local time. The next-cursor line
> printed after a page uses RFC3339 (UTC) so it round-trips exactly — you can
> paste it straight back into `--since`.

### Cursor paging

The success body is the top-level `{code, list, has_more}` envelope (the same
`list` slot as the `list` endpoint, plus `has_more`). The server does **not**
echo a cursor: the CLI derives the next `(--since, --since-id)` from the last
returned row. Without `--all`, one page is fetched and, if `has_more` is true,
the next `--since` / `--since-id` values are printed. With `--all`, the CLI
advances the cursor until `has_more` is false and prints the combined result.
`--limit` sets the page size (server default 100, max 100).

> Live progress / task-change streaming (the download-server WebSocket) is not
> exposed by the CLI; poll `sync` or `list` for the current state.
