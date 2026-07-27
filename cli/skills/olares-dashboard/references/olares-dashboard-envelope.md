# dashboard JSON envelope and empty data semantics

> **Prerequisite:** read [`../SKILL.md`](../SKILL.md) first.

Every `dashboard` command emits one of two frozen JSON shapes. Agents pin on `kind`, `raw`, and `meta.empty_reason`; do not pin on `display`.

## Shape A — leaf items

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
    { "raw": { "...": "upstream wire shape" }, "display": { "...": "table strings" } }
  ]
}
```

- `raw` is canonical machine data — numbers stay numbers, timestamps stay Unix seconds, temperatures stay Celsius.
- `display` is human presentation only and follows `--temp-unit` / `--timezone`.
- `meta.recommended_poll_seconds` is the SPA cadence; respect it for `--watch`.
- `meta.iteration` is 1-based and present in every `--watch` payload.

## Shape B — sections envelope

Parent commands aggregate multiple sub-views:

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

Sections are fetched concurrently. A single failed section degrades to `meta.error` on that section; the other sections still return. Surface partial outputs, do not blackout the whole envelope. To enumerate every live kind: `olares-cli dashboard schema -o json`.

## Empty data and gates

Optional hardware and integrations have three legitimate empty states:

| Upstream | `meta.empty` | `meta.empty_reason` | Meaning |
|---|---|---|---|
| HTTP 404 | `true` | `no_<feature>_integration` | Integration absent, e.g. HAMI vGPU not installed |
| HTTP 200, empty body | `true` | `no_<feature>_detected` | Integration present but hardware empty |
| HTTP 200, non-empty | `false` | `""` | Normal — `items[]` populated |
| Any 4xx / 5xx | n/a | n/a | Real failure, carried by `meta.error` |

Specific reasons:

| Reason | Where |
|---|---|
| `not_olares_one` | fan default / live / curve when active device is not Olares One |
| `no_fan_integration` | fan live HTTP 404 fallback |
| `no_vgpu_integration` | gpu list / tasks / get / task HTTP 404 |
| `vgpu_unavailable` | gpu list / tasks / get / task HTTP 5xx; `meta.error` carries upstream text |
| `no_gpu_detected` | gpu list / tasks / get / task HTTP 200 empty body |

Fan is hard-gated: non-Olares-One devices return `empty_reason=not_olares_one` before any fetch and still exit 0. GPU is soft-gated: the CLI always queries HAMI, but `meta.note` records SPA-hidden advisories such as non-admin profile or no node label.

## Agent decision tree

```text
inspect meta.empty + meta.empty_reason + meta.note ->
  not_olares_one        -> skip fan on this device
  no_*_integration      -> upstream component absent
  vgpu_unavailable      -> transient; retry/check meta.http_status + meta.error
  no_*_detected         -> integration up but hardware empty
  (none) + meta.note    -> data is present, but SPA would have hidden it
  (none) + (no note)    -> items[] populated, proceed normally
```
