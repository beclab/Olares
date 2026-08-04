# doctor: slow / resource pressure / scheduling rejected

> **Prerequisite:** read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.

Symptom: the system or an app feels slow, a pod can't schedule, or a GPU/compute binding is rejected with `node-pressure`. This reference orchestrates `dashboard` (usage) and `cluster` (scheduling) to locate the constrained resource.

## Where is the pressure?

```bash
# System + per-user resource snapshot (physical / user / ranking sections).
olares-cli dashboard overview -o json
# Drill into one dimension.
olares-cli dashboard overview memory -o json
olares-cli dashboard overview cpu -o json
olares-cli dashboard overview disk -o json
# Which app/workload is consuming the most.
olares-cli dashboard applications -o json
```

(Envelope shape, `--watch`, and empty-data semantics: [`../../olares-dashboard/SKILL.md`](../../olares-dashboard/SKILL.md).)

## Pod can't schedule

A pod stuck `Pending` (and a fresh install that ended in `stopped`, the app-stuck scheduling trap) is usually a scheduling constraint:

```bash
olares-cli cluster pod events <ns>/<pod>     # "Insufficient cpu/memory", taints, node affinity, no GPU
olares-cli cluster node list -o json         # allocatable vs requested per node, arch
```

| Event reason | Root cause | Next step |
|---|---|---|
| `Insufficient memory` / `cpu` | Node lacks headroom | Free resources (stop other apps), or lower the chart resource envelope |
| `Insufficient <gpu resource>` / no schedulable node | No matching GPU/accelerator capacity | Pick a node/device with capacity; check `settings compute list` |
| `node-pressure` on a GPU `resume` (Memory/CPU/Disk Total/Used/Needed) | The compute binding can't fit | Free resources or pick different cards with `market resume --compute-binding` |

## GPU / VRAM rejections

GPU binding rejections (`aggregate-vram-insufficient`, `device-vram-insufficient`, `node-pressure`, `multi-card-not-supported`, `gpu-type-mismatch`) come back from `market install`/`resume`. Re-pick devices with `market resume --compute-binding`; list operable devices with `olares-cli settings compute list`.

> Pruning unused images to reclaim disk is **doctor: image / pull failures** (`doctor images --unused`).
