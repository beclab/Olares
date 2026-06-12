# dashboard --watch / windows / NDJSON contract

> **Prerequisite:** Read the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags:** `olares-cli dashboard --help` (every flag here is a `dashboard`-root global flag inherited by every leaf).

`--watch` opts every leaf verb into an HTTP-polling loop that mirrors the SPA's `setTimeout` cadence. **Polling, not streaming** — every iteration is a fresh HTTP request, no chunked transfer encoding.

## `--watch` behaviour

```
invoke leaf cmd
   │
   ▼
--watch set? ── no ──> single execution → emit one envelope → exit 0/1
   │
   yes
   ▼
Runner.Run loop
   │
   ▼
RunOnce iteration N
   │
   ├── success         → emit envelope (Shape A or B) with meta.iteration=N
   │                       JSON mode → NDJSON line on stdout
   │                       table mode → clear-screen redraw
   │
   ├── transient err   → emit envelope with meta.error="...", items:[]
   │                       continue
   │
   └── ErrTokenInvalidated / ErrNotLoggedIn
                       → exit non-zero IMMEDIATELY
   ▼
3 consecutive errors? ──> exit non-zero
   │
   no
   ▼
--watch-iterations cap reached? ──> exit 0
   │
   no
   ▼
SIGINT (Ctrl-C) ──> exit 0 gracefully
   │
   no
   ▼
wait --watch-interval, loop back
```

## NDJSON contract (JSON mode)

- **One envelope per iteration; nothing else on stdout.** Pretty `-o json` is for one-shot mode; `--watch -o json` is **always NDJSON** (line-delimited, no leading newline, no trailing array bracket).
- A failed iteration emits `items: []` + `meta.error: "<msg>"` — **NDJSON keeps streaming**, the failure does not raise to process exit unless it's auth-class or the 3-consecutive cap trips.
- `meta.iteration` is **1-based** and present in every payload.
- Each envelope's `meta.fetched_at` reflects that iteration's HTTP-request start time.

```jsonc
{ "kind": "dashboard.overview.cpu", "meta": { "iteration": 1, ... }, "items": [...] }
{ "kind": "dashboard.overview.cpu", "meta": { "iteration": 2, "error": "...", ... }, "items": [] }
{ "kind": "dashboard.overview.cpu", "meta": { "iteration": 3, ... }, "items": [...] }
```

## Watch loop knobs

| Flag | Purpose | Default |
|---|---|---|
| `--watch` | Enable the loop | off |
| `--watch-interval D` | Time between iterations | command's `recommended_poll_seconds` (e.g. 30s for GPU, 60s for most others) |
| `--watch-iterations N` | Stop after N iterations | unbounded |
| `--watch-timeout D` | Stop after this much wall-clock time | unbounded |

**All three `--watch-*` knobs are REJECTED without `--watch`** — don't silently waste a flag.

## Window semantics — `--since` vs `--start/--end`

Mutually exclusive. Pick one.

| Form | Semantics | Behaviour under `--watch` |
|---|---|---|
| `--since 5m` | Window = `[now - 5m, now]` | **Sliding** — each iteration recomputes `now` |
| `--start <RFC3339> --end <RFC3339>` | Window = `[start, end]` | **Fixed** — same window every iteration |

If neither is set, each leaf has its own SPA-derived default (e.g. 8h for GPU detail page, 1h for GPU task detail page, no window for cluster-level instant snapshots).

The chosen window appears in `meta.window` so an agent can replay the exact same query without recomputing:

```json
"meta": {
  "window": { "since": "8h", "start": "...", "end": "...", "step": "30m" }
}
```

## Examples

```bash
# Tail CPU every 30s as NDJSON until Ctrl-C.
olares-cli dashboard overview cpu --watch -o json --watch-interval 30s

# Tail CPU but stop after 10 iterations.
olares-cli dashboard overview cpu --watch -o json --watch-iterations 10

# Watch a GPU detail page for 5 minutes with a sliding 15m window.
olares-cli dashboard overview gpu graphics <uuid> --watch -o json --since 15m --watch-timeout 5m

# Fixed window across all iterations (useful for retrospective tailing).
olares-cli dashboard overview ranking --watch -o json \
  --start 2026-05-28T00:00:00Z --end 2026-05-28T01:00:00Z
```

## SIGINT / graceful shutdown

- Ctrl-C (SIGINT) and SIGTERM trigger a graceful stop — the current iteration is allowed to finish, then the loop exits with `0`.
- Inside a `for line in cmd.stdout` loop, the parent agent process sees the stream end cleanly (no half-line, no half-envelope).

## `--user <olaresId>` interaction

- `--user` only changes upstream filter args. The auth header is still the active profile's token.
- **Platform-admin only.** Non-admin callers get an immediate error; the loop never starts.
- The admin check is cached per-Client (`sync.Once`) — there is zero per-iteration auth-check overhead.

## `--timezone <IANA>` interaction

- Affects DISPLAY only — `meta.window` / table-side timestamps / header rendering.
- The wire format sent to HAMI's monitor query endpoints stays in the backend pod's TZ (default `Asia/Shanghai`) so trends resolve correctly even when `--timezone` differs.
- Override the backend TZ with `OLARES_HAMI_BACKEND_TZ=<IANA>` env if your HAMI deployment is in a different zone (hidden from `--help`; only operators of non-default HAMI deployments need it).

## Agent best practices

- **Use `--watch -o json` for live tailing.** Always NDJSON, always one envelope per line.
- **Respect `meta.recommended_poll_seconds`.** It mirrors the SPA's polling cadence; overriding with a much shorter `--watch-interval` adds upstream load.
- **Handle the empty-envelope iterations.** `meta.error` is normal during transient upstream issues; do NOT abort the consumer loop unless 3 consecutive errors trip the safety net (the CLI exits non-zero in that case anyway).
- **Stop polling on `meta.empty_reason == not_olares_one`.** It's a hard gate that won't change without a hardware change.
- **For `vgpu_unavailable`**, the CLI keeps streaming — your agent should keep iterating too. Real failures still surface via process exit on auth-class errors or 3 consecutive iteration failures.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `--watch-iterations requires --watch` | Knob without gate | Add `--watch` or drop the knob |
| `--since and --start/--end are mutually exclusive` | Both window forms set | Pick one |
| Process exits non-zero immediately on the first iteration | `ErrNotLoggedIn` / `ErrTokenInvalidated` | Run `olares-cli profile login` and retry |
| Process exits non-zero after a few iterations | 3 consecutive transient errors | Investigate upstream (HAMI / capi); restart the loop |
