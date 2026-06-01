---
name: olares-dashboard
version: 4.0.0
description: "Olares Dashboard (olares-cli dashboard) — query the Olares Dashboard SPA's Overview and Applications routes from the command line, scoped to the active Olares ID. An AI-agent-first JSON mirror with a strict dual-shape envelope (leaf items / aggregated sections), stable Kind constants, and human-readable table output. Covers Overview leaves (CPU, memory, disk, network, pods, fan, GPU, ranking), Applications listing and detail, the --watch HTTP-polling loop (interval / iterations / timeout / NDJSON-per-iteration / SIGINT-graceful), --since (sliding window) vs --start/--end (fixed window) selection, three-state empty-data semantics (no_<feature>_integration / no_<feature>_detected / vgpu_unavailable), and capability gates (fan is hard-gated to Olares One hardware; GPU is a soft advisory mirroring the SPA sidebar). Use when the user mentions Olares, Olares ID, Olares Dashboard, Olares One (fan / cooling), olares-cli dashboard, the Olares overview / applications view, 'show CPU / memory / disk / pods / network / fan / GPU / ranking on Olares', wants JSON for an AI agent, or hits errors like 'fan is only available on Olares One devices' or 'gpu data temporarily unavailable'."
metadata:
  requires:
    bins: ["olares-cli"]
  cliHelp: "olares-cli dashboard --help"
---

# dashboard (overview + applications, AI-agent first)

**CRITICAL — before doing anything, MUST use the Read tool to read [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for profile selection, login, automatic token refresh, and the auth-error recovery table.**

> **Source of truth for flags is always `olares-cli dashboard --help` (global flags) and `olares-cli dashboard <verb> --help` (per-leaf flags).** This file only carries what `--help` cannot give: the dual-shape JSON envelope contract, three-state empty-data semantics, capability gates, watch / window rules, and the verb index.

## Routing

This subtree is an **AI-agent-first JSON mirror of the Olares Dashboard SPA's Overview and Applications routes**. Use it when:

- The user asks for runtime metrics: CPU / memory / disk / pods / network / fan / GPU.
- The user wants the workload-grain or application-grain resource ranking.
- The user wants the JSON form of what the SPA Overview / Applications pages show.
- The user wants `--watch` for live-tailing one of the above.

Sibling-skill routing:

| User intent | Use instead |
|---|---|
| "What pods / workloads / nodes / namespaces exist?" (object inventory) | [`olares-cluster`](../olares-cluster/SKILL.md) |
| "Install / start / stop an Olares app" | [`olares-market`](../olares-market/SKILL.md) |
| "Configure my user settings" | [`olares-settings`](../olares-settings/SKILL.md) |
| "Browse files / drives" | [`olares-files`](../olares-files/SKILL.md) |

> **Mental model:** dashboard answers *"what's the resource usage and health"*. For inventory and lifecycle, route elsewhere.

## JSON envelope (two shapes, frozen)

Every command emits exactly one of two shapes. The choice is fixed per command. **Agents pin on `kind` and `meta.empty_reason`; the CLI does not change these values across releases.**

### Shape A — leaf items

```json
{
  "kind": "dashboard.<area>.<verb>",
  "meta": {
    "fetched_at": "...",
    "iteration": 0,
    "recommended_poll_seconds": 60,
    "empty": false,
    "empty_reason": "",
    "error": "",
    "http_status": 200
  },
  "items": [
    { "raw": { /* upstream wire shape */ }, "display": { /* table-friendly strings */ } }
  ]
}
```

- `raw` is the canonical machine-friendly shape — numbers as numbers, timestamps as Unix seconds, temperatures as raw Celsius. **Agents pin on `raw`.**
- `display` is human-presentation only, formatted with current `--temp-unit` / `--timezone`. **Agents MUST NOT pin on `display`** — it can change with locale or format fixes.
- `meta.recommended_poll_seconds` — the page-level cadence the SPA uses; respect it when driving `--watch`.
- `meta.iteration` — 1-based, present in every `--watch` payload.

### Shape B — sections envelope

Used by **parent commands that aggregate multiple sub-views**:

| Parent command | Sections |
|---|---|
| `dashboard overview` | `physical` / `user` / `ranking` |
| `dashboard overview disk` | `main` / `partitions` |
| `dashboard overview fan` | `live` / `curve` |
| `dashboard overview gpu` | `graphics` / `tasks` |

```json
{
  "kind": "dashboard.overview",
  "meta": {...},
  "sections": {
    "physical": { "kind": "dashboard.overview.physical", "meta": {...}, "items": [...] },
    "user":     { "kind": "dashboard.overview.user",     "meta": {...}, "items": [...] },
    "ranking":  { "kind": "dashboard.overview.ranking",  "meta": {...}, "items": [...] }
  }
}
```

Sections are fetched **concurrently**. A single failed section degrades to `meta.error="..."` on that section — **the others still return**. Surface partial outputs to the user, don't blackout the whole envelope.

To enumerate every kind live: `olares-cli dashboard schema -o json`.

## Three-state empty data (`meta.empty_reason`)

Optional hardware (GPU / fan) and optional integrations have three legitimate "empty" states. The CLI distinguishes them so agents can branch without parsing prose.

| Upstream | `meta.empty` | `meta.empty_reason` | Meaning |
|---|---|---|---|
| HTTP 404 | `true` | `no_<feature>_integration` | Integration absent (e.g. HAMI vGPU not installed) |
| HTTP 200, empty body | `true` | `no_<feature>_detected` | Integration present but hardware empty |
| HTTP 200, non-empty | `false` | `""` | Normal — items[] populated |
| Any 4xx / 5xx | n/a (envelope `meta.error`) | n/a | Real failure |

**Forbidden**: turning a 404 into an error, or merging the two reasons into a single "empty=true" string. Agents currently key on the reason.

Specific reason values:

| Reason | Where |
|---|---|
| `not_olares_one` | fan default / live / curve (active device's `device_name` is not `Olares One`) |
| `no_fan_integration` | fan live (HTTP 404 fallback) |
| `no_vgpu_integration` | gpu list / tasks / get / task (HTTP 404) |
| `vgpu_unavailable` | gpu list / tasks / get / task (HTTP 5xx); `meta.error` carries the upstream message, `meta.http_status` the original status |
| `no_gpu_detected` | gpu list / tasks / get / task (HTTP 200, empty body) |

## Capability gates

The fan / GPU subtrees mirror SPA gates, but with **two different strengths**.

### Hard gate — fan (Olares One only)

`overview fan *` returns an empty envelope BEFORE any fetch on non-Olares-One hardware:

```json
{
  "kind": "dashboard.overview.fan.live",
  "meta": {
    "empty": true,
    "empty_reason": "not_olares_one",
    "note": "Fan / cooling integration is only available on Olares One devices",
    "device_name": "DIY-PC"
  },
  "items": []
}
```

Exit code is still `0` — this is a predictable state, not a failure.

### Soft gate — GPU (advisory)

The CLI ALWAYS calls HAMI for GPU verbs. The SPA-equivalent "would-have-been-hidden" reason is recorded as `meta.note` and (in table mode) as a one-liner on stderr. The actual `meta.empty` / `empty_reason` are decided by HAMI's response.

Advisories you may see in `meta.note`:

- `GPU sidebar entry is hidden for non-admin profiles in the SPA; HAMI was queried directly`
- `no node carries gpu.bytetrade.io/cuda-supported=true; SPA hides the GPU card. HAMI was queried directly`

When BOTH a HAMI failure AND a soft advisory apply, `meta.note` is `"<advisory> | <hami-explanation>"`.

## Agent decision tree (recommended)

```
inspect meta.empty + meta.empty_reason + meta.note →
  not_olares_one        → skip the fan subtree on this device
  no_*_integration      → upstream component absent (HAMI / capi system fan)
  vgpu_unavailable      → transient: retry; check meta.http_status / meta.error
  no_*_detected         → integration up but hardware empty
  (none) + meta.note    → data is present but the SPA would have hidden the entry; surface the note to the user
  (none) + (no note)    → items[] populated, proceed normally
```

## Verb index

For flags & examples, **always start with `olares-cli dashboard <verb> --help`**.

| Verb | Purpose | `--help` first, then... |
|---|---|---|
| `applications` (alias `apps`) | Workload-grain application table (mirrors SPA Applications page) | `olares-cli dashboard applications --help` |
| `overview` (no subverb) | Default sections envelope (physical + user + ranking) | [references/olares-dashboard-overview.md](references/olares-dashboard-overview.md) |
| `overview <section>` | Per-section snapshot (10 sub-verbs: cpu / memory / disk / pods / network / fan / gpu / physical / user / ranking) | [references/olares-dashboard-overview.md](references/olares-dashboard-overview.md) |
| `schema` | Introspect the JSON Schemas served by `olares-cli dashboard` | `olares-cli dashboard schema --help` |

For `--watch`, `--since` / `--start/--end`, `--user`, `--timezone`, and the NDJSON contract, see [references/olares-dashboard-watch.md](references/olares-dashboard-watch.md).

## Global flags (cross-cutting)

These are bound on the `dashboard` root and inherited by every leaf. **Defaults are sensible — don't pass them unless the user explicitly asks for the non-default.**

| Flag | Purpose |
|---|---|
| `-o, --output table\|json` | Output format (default `table`). JSON is the agent-facing form |
| `--head N` | Truncate to the first N rows after sorting (0 = no truncation) |
| `--limit N` / `--page N` | Pagination knobs (0 = upstream defaults) |
| `--since D` | Relative window (e.g. `5m`, `1h`). **Sliding when `--watch`** |
| `--start <RFC3339>` / `--end <RFC3339>` | Absolute window. **Fixed across iterations when `--watch`**. Mutually exclusive with `--since` |
| `--temp-unit C\|F\|K` | Display unit for temperature (JSON `raw` always Celsius) |
| `--timezone <IANA>` | Display TZ for table rendering (does NOT affect wire-format TZ) |
| `--user <olaresId>` | Target a different user than the active profile — **platform-admin only**, returns 403 otherwise |
| `--watch` | Enable HTTP-polling loop (NDJSON per iteration in JSON mode) |
| `--watch-interval D` / `--watch-iterations N` / `--watch-timeout D` | Watch-loop knobs (require `--watch`; rejected otherwise) |

## Exit codes & error semantics

- **Exit `0`** for every gated / advisory / empty path — these are predictable states, not failures.
- **Exit non-zero** only on:
  - Auth-class errors (`ErrTokenInvalidated` / `ErrNotLoggedIn`) — propagated immediately
  - 3 consecutive iteration failures inside a `--watch` loop
  - A real upstream `meta.error` on a one-shot invocation
- **Stderr** carries one human-readable line in non-JSON modes. JSON / NDJSON modes stay silent on stderr — agents read stdout exclusively.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `fan is only available on Olares One devices (current: <device_name>)` (stderr) | Hard gate; `meta.empty_reason=not_olares_one` | Stop probing fan on this device |
| `(advisory) GPU sidebar entry is hidden for non-admin profiles ...` (stderr) | Soft gate; data still returned | Surface to user as a note; don't treat as error |
| `gpu data temporarily unavailable: HAMI returned HTTP 500` | `vgpu_unavailable`; transient | Retry; if persistent, file a server-side issue |
| `--user requires platform-admin role` | Non-admin profile passing `--user` | Use the active profile, or switch with `olares-cli profile use` |
| `--watch-iterations requires --watch` (or `--interval` / `--timeout` similarly) | Polling knob without gate flag | Add `--watch` or drop the knob |
| `--since and --start/--end are mutually exclusive` | Both window forms set | Pick one |
| 401/403 from any dashboard verb | Token rotation / invalidation | See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) |

For the full auth-error matrix see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
