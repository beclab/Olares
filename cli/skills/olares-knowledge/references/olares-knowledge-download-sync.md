# knowledge download sync

> **Flags:** `olares-cli knowledge download unfinished --help`, `... sync --help`.

## unfinished

```bash
olares-cli knowledge download unfinished
olares-cli knowledge download unfinished --app wise -o json
```

Lists tasks that are **not** in a terminal state (downloading / pending /
paused / …), `GET /api/download/unfinished`. Same table columns as `list`.
`--app` filters by namespace.

## sync

```bash
olares-cli knowledge download sync                 # first page from the start
olares-cli knowledge download sync --after 128     # only tasks with id > 128
olares-cli knowledge download sync --all -o json   # drain every page
```

`GET /api/download/sync` is an **incremental pull keyed on task id**, not a
change feed. It returns tasks whose id is greater than `--after`, in ascending
id order, **including finished ones**. Because the cursor is the id, sync does
**not** surface progress updates to tasks you already saw — only tasks with a
larger id. To follow new tasks over time, remember the largest id you saw and
pass it as `--after` next time.

### Cursor paging

The response carries `has_more` and `next_cursor` (the last item's id) inside
`data`. Without `--all`, one page is fetched and, if `has_more` is true, the
next `--after` value is printed. With `--all`, the CLI follows the cursor until
`has_more` is false and prints the combined result. `--limit` sets the page
size (server default 100, max 500).

> Live progress / task-change streaming (the download-server WebSocket) is not
> exposed by the CLI; poll `sync` or `list` for the current state.
