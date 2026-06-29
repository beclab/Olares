---
name: olares-dashboard
version: 4.2.0
description: "Olares Dashboard via olares-cli dashboard — CPU, memory, disk, network, pods, fan, GPU, ranking, applications; JSON envelope and --watch. Use for Olares Dashboard, overview, resource usage, Olares One fan."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# dashboard (overview + applications, AI-agent first)

**CRITICAL — before doing anything, load the `olares-shared` skill first (profile selection, login, token refresh, auth-error recovery). Flag reference: `olares-cli dashboard --help`.**

> **Source of truth for flags is always `olares-cli dashboard --help` (global flags) and `olares-cli dashboard <verb> --help` (per-leaf flags).** This file only carries what `--help` cannot give: the dual-shape JSON envelope contract, three-state empty-data semantics, capability gates, watch / window rules, and the verb index.

## When to use

This subtree is an **AI-agent-first JSON mirror of the Olares Dashboard SPA's Overview and Applications routes**. Use it when:

- The user asks for runtime metrics: CPU / memory / disk / pods / network / fan / GPU.
- The user wants the workload-grain or application-grain resource ranking.
- The user wants the JSON form of what the SPA Overview / Applications pages show.
- The user wants `--watch` for live-tailing one of the above.
- Errors: `fan is only available on Olares One devices`, `gpu data temporarily unavailable`
- Empty-data reasons: `no_<feature>_integration`, `no_<feature>_detected`, `vgpu_unavailable`; windows: `--since` vs `--start`/`--end`

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** dashboard answers *"what's the resource usage and health"*. For inventory and lifecycle, route elsewhere. When the metrics reveal a problem (resource pressure, an app that's `running` but slow/unreachable), hand off to [`../olares-doctor/SKILL.md`](../olares-doctor/SKILL.md) for root-cause diagnosis.

## JSON envelope and empty data

Every command emits a stable JSON envelope. Agents should pin on `kind`, `raw`, and `meta.empty_reason`; never pin on table/display strings. Empty hardware/integration states are predictable `exit 0` envelopes, not failures.

For the exact Shape A / Shape B wire forms, fan/GPU gate semantics, and the decision tree, read [references/olares-dashboard-envelope.md](references/olares-dashboard-envelope.md).

## Verb index

For flags & examples, **always start with `olares-cli dashboard <verb> --help`**.

| Verb | Purpose | `--help` first, then... |
|---|---|---|
| `applications` (alias `apps`) | Workload-grain application table (mirrors SPA Applications page) | `olares-cli dashboard applications --help` |
| `overview` (no subverb) | Default sections envelope (physical + user + ranking) | [references/olares-dashboard-overview.md](references/olares-dashboard-overview.md) + [references/olares-dashboard-envelope.md](references/olares-dashboard-envelope.md) |
| `overview <section>` | Per-section snapshot (10 sub-verbs: cpu / memory / disk / pods / network / fan / gpu / physical / user / ranking) | [references/olares-dashboard-overview.md](references/olares-dashboard-overview.md) + [references/olares-dashboard-envelope.md](references/olares-dashboard-envelope.md) |
| `schema` | Introspect the JSON Schemas served by `olares-cli dashboard` | `olares-cli dashboard schema --help` |

For `--watch`, `--since` / `--start/--end`, `--user`, `--timezone`, and the NDJSON contract, see [references/olares-dashboard-watch.md](references/olares-dashboard-watch.md).

## Global flags (cross-cutting)

Root flags are inherited by every leaf; trust `olares-cli dashboard --help` for the complete list. Defaults are sensible — don't pass `--head`, pagination, time-window, temp-unit, timezone, user, or watch knobs unless the user asks. For watch/window semantics, read [references/olares-dashboard-watch.md](references/olares-dashboard-watch.md).

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
| `--watch-iterations requires --watch` (or `--watch-interval` / `--watch-timeout` similarly) | Polling knob without gate flag | Add `--watch` or drop the knob |
| `--since and --start/--end are mutually exclusive` | Both window forms set | Pick one |
| 401/403 from any dashboard verb | Token rotation / invalidation | See [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) |

For the full auth-error matrix see [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md).
