# dashboard overview

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli dashboard overview --help` and `olares-cli dashboard overview <section> --help`.

The most flag-rich verb in this subtree. Three "framing" sections (`physical` / `user` / `ranking`) live at the top of the overview tree; seven "hardware" sections (`cpu` / `memory` / `disk` / `pods` / `network` / `fan` / `gpu`) live alongside.

## Default action (no subverb)

```bash
olares-cli dashboard overview
olares-cli dashboard overview -o json
```

Emits a **Shape B sections envelope** with the three "framing" sections fetched concurrently:

```json
{
  "kind": "dashboard.overview",
  "sections": {
    "physical": { "kind": "dashboard.overview.physical", "items": [...] },
    "user":     { "kind": "dashboard.overview.user",     "items": [...] },
    "ranking":  { "kind": "dashboard.overview.ranking",  "items": [...] }
  }
}
```

For per-hardware sections (CPU / disk / fan / GPU / etc.), invoke them explicitly — they are NOT part of the default envelope.

## 10-section matrix

| Section | Kind | When to use |
|---|---|---|
| `physical` | `dashboard.overview.physical` | 9-row cluster-level snapshot (CPU/Memory/Disk/Pods/Net + extras). The "single screen vital signs" view |
| `user` | `dashboard.overview.user` | Per-user CPU / memory quota usage. Accepts optional `<username>` positional — **admin-only** for non-self lookups. Mirrors SPA's User Resources panel |
| `ranking` | `dashboard.overview.ranking` | Workload-grain resource ranking (per-application). `--sort asc|desc` + `--head N` for top-N. Mirrors SPA's UsageRanking widget |
| `cpu` | `dashboard.overview.cpu` | Per-node CPU details (model / freq / cores / utilisation breakdown / temp / load avg) |
| `memory` | `dashboard.overview.memory` | Per-node memory breakdown. `--mode physical|swap` (default `physical`) |
| `disk` | `dashboard.overview.disk` (Shape B) | **Sections envelope**: `main` (per-disk table) + `partitions` (per-device partition tables). `disk main` and `disk partitions <device>` give individual views |
| `pods` | `dashboard.overview.pods` | Per-node pod count snapshot (last/avg/max running) |
| `network` | `dashboard.overview.network` | Per-physical-NIC table. `--test-connectivity` (default true) probes internet/IPv6 per interface |
| `fan` | `dashboard.overview.fan` (Shape B) | **Olares One only.** Sections envelope: `live` (real-time fan/temp/power) + `curve` (hardcoded fan-curve spec). Hard-gated on non-Olares-One: `meta.empty_reason=not_olares_one` |
| `gpu` | `dashboard.overview.gpu` (Shape B) | Sections envelope: `graphics` (GPU list) + `tasks` (vGPU tasks). **Soft-gated**; `meta.note` advises when the SPA would have hidden the entry. See parent SKILL.md for the soft-gate decision tree |

## Drilling into a section

```bash
# Per-section snapshots.
olares-cli dashboard overview cpu -o json
olares-cli dashboard overview memory --mode swap
olares-cli dashboard overview network --test-connectivity=false

# Sections envelope (disk has its own Shape B):
olares-cli dashboard overview disk -o json                    # both sections
olares-cli dashboard overview disk main                       # just per-disk table
olares-cli dashboard overview disk partitions sda             # per-device partitions

# Fan (Olares One only; otherwise empty envelope with reason=not_olares_one):
olares-cli dashboard overview fan -o json
olares-cli dashboard overview fan live
olares-cli dashboard overview fan curve

# GPU (soft-gated):
olares-cli dashboard overview gpu                             # graphics + tasks sections
olares-cli dashboard overview gpu graphics                    # just GPU list
olares-cli dashboard overview gpu graphics <uuid>             # per-GPU detail page (gauges + trends)
olares-cli dashboard overview gpu tasks                       # task list
olares-cli dashboard overview gpu tasks <name-or-pod-uid>     # per-task detail page

# Ranking (top-N apps by CPU).
olares-cli dashboard overview ranking --sort desc --head 5
```

## GPU detail pages

`gpu graphics <uuid>` and `gpu tasks <name-or-pod-uid>` are **full SPA-mirror detail pages** — they assemble a sections envelope with `detail` / `gauges` / `trends`:

- `detail` — HAMI metadata
- `gauges` — instant-vector queries (current values)
- `trends` — range-vector queries (default 8h for GPU, 1h for task). Override with `--since` / `--start` / `--end`

Partial-failure semantics: a single gauge / trend item failing does NOT abort the envelope; it carries `raw.error` and contributes to `meta.warnings`. **Branch on `len(meta.warnings) > 0` to detect partial data.**

For `gpu tasks <ref>`, `<ref>` accepts EITHER the TASK column (`name`) OR the POD_UID column (`podUid`) — copy-paste from either column works. If two tasks share a name, the CLI errors with the candidate pod-uids; re-run with one.

## Agent notes

- For "give me a single-screen vital-signs view", **`olares-cli dashboard overview -o json` is the right verb** (Shape B envelope with physical + user + ranking).
- For "tail CPU / memory live", **add `--watch -o json`** to a per-section command. The output is NDJSON, one envelope per iteration. See [olares-dashboard-watch.md](olares-dashboard-watch.md).
- For "what apps are using the most CPU", **`overview ranking --sort desc --head 10`** is the right entry point — it returns a workload-grain ranking with both `raw` (machine-friendly) and `display` (human-friendly) columns.
- For `--user <other>` cross-tenant queries, **only platform-admins succeed**; non-admins get an immediate error. Run `cluster context` first if you need to confirm the active role.

## Common errors

The fan / GPU gate, `--user` admin-only, and HAMI-absent errors are in the consolidated [Common errors](../SKILL.md#common-errors) table in the parent skill.
