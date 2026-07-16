# knowledge download torrent

> **Flags:** `olares-cli knowledge download torrent <verb> --help`.

## inspect

```bash
olares-cli knowledge download torrent inspect --file ./x.torrent
olares-cli knowledge download torrent inspect --file ./x.torrent -o json
```

Reads the local `.torrent`, base64-encodes it and uploads it
(`POST /api/download/torrent/inspect`). Returns the info hash, mode
(`single`/`multi`), piece layout and a **1-based** file list. Those indices
feed `torrent files --select` and `create --select-files`.

## stats / peers

```bash
olares-cli knowledge download torrent stats 42
olares-cli knowledge download torrent peers 42 -o json
```

`stats` shows live BitTorrent counters (speeds, share ratio, seeders,
pieces have/total, ETA, `is_seeding`). `peers` lists connected peers as
`IP:PORT  PROGRESS  DOWN  UP  SEEDER` (progress is shown as a percentage).

## files

```bash
olares-cli knowledge download torrent files 42 --select 1,3,5
olares-cli knowledge download torrent files 42 --select all
```

Sets the **full** selection of a multi-file torrent (not a delta),
`PUT /api/download/<id>/torrent/files`. `--select` takes comma-separated
1-based indices; `--select all` sends an empty selection so every file is
kept.

## seed stop / resume

```bash
olares-cli knowledge download torrent seed stop 42
olares-cli knowledge download torrent seed resume 42
```

`POST /api/download/<id>/torrent/seed/stop|resume`. HTTP 409 means the task
is in the wrong state: `stop` needs a task that is currently seeding,
`resume` needs a completed task.

## create with torrent / magnet

```bash
olares-cli knowledge download create 'magnet:?xt=urn:btih:...'
olares-cli knowledge download create --torrent ./x.torrent --select-files 1,3
```

A magnet link is an ordinary URL argument. `--torrent` uploads a local
`.torrent` (base64, `extra.torrent_file_b64`) and lets the URL argument be
omitted. `--select-files` passes 1-based indices through as
`extra.selected_files`.
