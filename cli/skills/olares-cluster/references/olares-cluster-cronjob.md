# cluster cronjob (aliases `cronjobs` / `cj`)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster cronjob --help` and `olares-cli cluster cronjob <verb> --help`.

K8s CronJobs (`apis/batch/v1beta1` — different from `cluster job`'s `batch/v1`). Scheduled Job templates.

## Verbs at a glance

| Verb | Purpose |
|---|---|
| `list` | NAMESPACE / NAME / SCHEDULE / SUSPEND / ACTIVE / LAST-SCHEDULE / AGE |
| `get <ns/name>` | Vertical summary + `ConcurrencyPolicy` + Active Jobs + Job Template Selector |
| `yaml <ns/name>` | Full K8s-native YAML |
| `jobs <ns/name>` | Two-step: GET cronjob for `spec.jobTemplate.metadata.labels`, then list Jobs by derived labelSelector. `--limit N` caps results |
| `suspend <ns/name>` | **Destructive.** PATCH `{"spec":{"suspend":true}}` (merge-patch+json). No-op short-circuit when already suspended |
| `resume <ns/name>` | Non-destructive. PATCH `{"spec":{"suspend":false}}`. No-op short-circuit when already active. No `--yes` (re-enabling isn't destructive) |

## Safety constraints

- **`suspend` is destructive — confirm with the user.** Suspending stops the CronJob from spawning new Jobs at scheduled times; **already-running Jobs are NOT affected**, only future schedule fires.
- **`suspend` is reversible via `resume`** — flag it as "pause schedule" to the user, not "delete".
- **No-op short-circuit**: if the cronjob is already in the target state, the verb prints a notice and skips the PATCH — saves an API round-trip and a wasted resource-version bump.

## The `jobs` label derivation

`cluster cronjob jobs <ns/name>` does NOT have a server endpoint that lists "jobs owned by this cronjob" directly. The CLI:

1. GETs the cronjob to read `spec.jobTemplate.metadata.labels`
2. Uses those labels as a `labelSelector` against `/apis/batch/v1/namespaces/<ns>/jobs`

**If the jobTemplate carries no labels, the verb errors clearly** rather than fanning to "every job in the namespace" (which would be the misleading behavior of a naive selector-builder).

## Examples

```bash
# List scheduled jobs.
olares-cli cluster cronjob list

# Full schedule + concurrency + active jobs.
olares-cli cluster cronjob get user-system-alice/nightly-backup

# Recent Jobs spawned by this CronJob.
olares-cli cluster cronjob jobs user-system-alice/nightly-backup --limit 10

# Pause the schedule (confirms; new schedule fires are skipped).
olares-cli cluster cronjob suspend user-system-alice/nightly-backup

# Re-enable.
olares-cli cluster cronjob resume user-system-alice/nightly-backup
```

## Agent notes

- For "stop running this every night" requests, **suspend, don't delete**. Suspend preserves the schedule definition and history; deletion is one-way.
- **The "is the cronjob actually firing?" question is two-step**: `get` to see `SUSPEND` + `LastScheduleTime` and `Active`, then `jobs` to inspect recent spawns. If `SUSPEND=true` AND there are recent Active jobs, those finished AFTER the schedule was paused — they'll complete normally but no new ones spawn.
- **After a long pause, expect a single catch-up fire** when resumed (depending on `ConcurrencyPolicy` and `StartingDeadlineSeconds`). Warn the user before `resume` on a CronJob that's been suspended for hours.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `jobTemplate has no labels — cannot derive selector` (from `jobs`) | The CronJob's job template doesn't set labels | Inspect manually with `cluster cronjob yaml` and the user's own labels |
| `cronjob already suspended` / `already active` | No-op short-circuit | Notice, no action needed |
| 404 on `cronjob get` | Wrong API group? — Olares uses `batch/v1beta1` here, not `batch/v1` | The CLI uses the right one; if 404 surfaces, the CronJob actually doesn't exist for this profile |
