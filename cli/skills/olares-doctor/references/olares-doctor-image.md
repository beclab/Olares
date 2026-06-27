# doctor: image problems (pull failures, unused images)

> **Prerequisite:** read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.

Two distinct concerns: an image that **won't pull** (blocks an install), and local images that are **unused** (reclaim disk).

## Image won't pull

Symptom on a pod: `ImagePullBackOff`, `ErrImagePull`, `InvalidImageName`, or a runtime arch error (`no match for platform`, `exec format error`). These are hard, never-self-healing conditions — app-service fast-fails them after the 5-minute grace.

```bash
olares-cli cluster pod list -n <ns> -o json     # state.waiting.reason on the failing container
olares-cli cluster pod events <ns>/<pod>        # "Failed to pull image ...": the registry/auth/arch detail
```

| Reason | Root cause | Next step |
|---|---|---|
| `ImagePullBackOff` / `ErrImagePull` | Image missing, private without creds, or registry/mirror unreachable | Confirm the ref is public & pullable; if a mirror is down, treat it as the app-stuck stalled-pull path |
| `InvalidImageName` | Malformed image ref in the workload spec | For a chart you author, fix the image reference |
| `no match for platform` / `exec format error` | Image arch != node arch | Rebuild for this node's arch (`cluster node list` for arch) |

A slow-but-progressing pull is NOT a failure — use the app-stuck `downloading` procedure to distinguish slow from stalled.

## Unused local images (`doctor images`)

`doctor images` lists local containerd images annotated with how many workloads reference each one. **Source of truth for flags: `olares-cli doctor images --help`.**

```bash
olares-cli doctor images                 # all local images + reference counts
olares-cli doctor images --unused        # zero-reference prune candidates, largest-first, with reclaimable size
olares-cli doctor images -n <ns>         # scope the reference count to one namespace
olares-cli doctor images -o json
```

- Always full-scans the **control node**; the unused verdict for a listed image is cluster-wide-reference-checked, so it is safe — but images living only on a worker node won't appear (not a full-cluster census).
- References are counted from Deployment / StatefulSet / DaemonSet / Job / CronJob **specs**, not from bare/static pods or the digest a running container is pinned to. A tag bumped to a new digest can leave the old, still-running image looking unused — **cross-check running pods before reclaiming.**
- `pause`/sandbox images are excluded (runtime-pinned, never prunable).

This command is read-only — it identifies candidates; pruning itself is a separate, deliberate action.
