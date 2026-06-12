# cluster job (alias `jobs`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster job --help` and `olares-cli cluster job <verb> --help`.

K8s Jobs (`apis/batch/v1`) — one-shot batch runs.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | NAMESPACE / NAME / COMPLETIONS / STATUS / DURATION / AGE |
| `get <ns/name>` | Vertical summary + Conditions + `Controlled By: CronJob/<name>` if present |
| `yaml <ns/name>` | Full K8s-native YAML |
| `pods <ns/name>` | Two-step: GET job for `metadata.uid`, then `pod list` with `controller-uid=<uid>`. `-l` is ANDed onto the controller-uid clause server-side |
| `events <ns/name>` | Filtered to `involvedObject.kind=Job, name=<job>`. Same render/sort/URL helpers as `pod events` |
| `rerun <ns/name>` | **Destructive.** KubeSphere operations API — `POST /kapis/operations.kubesphere.io/v1alpha2/.../jobs/<name>?action=rerun&resourceVersion=<rv>`. Server spawns a new pod attempt |

## Safety constraints

- **`rerun` is destructive — confirm with the user.** It re-fires the Job's pod template against the cluster; for jobs with side effects (sends notifications, mutates external state) this happens AGAIN.
- The CLI fetches the Job's current `resourceVersion` before posting; concurrent reruns by a third party between Get and POST will fail with a conflict (which is the safe outcome).

## The `pods` two-step

`cluster job pods <ns/name>` does NOT have a single endpoint:

1. `cluster job.Get <ns/name>` → `metadata.uid`
2. `pod list` with `labelSelector=controller-uid=<uid>`

Pass `-l <additional-label>` to AND another selector clause onto the second call (e.g. `-l my-app=worker` to scope to a specific shard).

## Examples

```bash
# List one-shot jobs across visible namespaces.
olares-cli cluster job list

# Single Job's full state + events.
olares-cli cluster job get user-system-alice/migrate-2026-q1
olares-cli cluster job events user-system-alice/migrate-2026-q1

# Pods spawned by this Job.
olares-cli cluster job pods user-system-alice/migrate-2026-q1

# Rerun (confirms).
olares-cli cluster job rerun user-system-alice/migrate-2026-q1
```

## Agent notes

- **Distinguish Job from CronJob first.** A "scheduled migration" usually means CronJob (use `cluster cronjob`); a "one-shot batch run" is Job. If unsure, `get` returns `Controlled By: CronJob/<name>` for Jobs owned by a CronJob.
- For "why did this job fail?" questions, **chain `get` → `pods` → `pod logs`**: get shows status + conditions, pods finds the failed pod, logs reveals the error.
- **`rerun` is the right verb for "run it again with the same template".** It's NOT the same as deleting + re-applying the YAML — that would re-set `resourceVersion` and lose the job history.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `job pods` returned empty | Job has not yet spawned a pod (e.g. just created), OR pods were garbage-collected | Check `cluster job get` → `Conditions`; if older, the pods may be gone |
| `rerun` returned 409 / conflict | Concurrent modification | Re-run; if persistent, someone else is operating on the same Job |
| `aborted by user` | y/N prompt rejected | Re-run with `--yes` if intentional |
