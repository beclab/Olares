# Recover a deploy whose market row is wedged or settled wrong

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first; this is the recovery half of the [deploy](olares-chart-deploy.md) loop. Which verb a state accepts is the app-service state machine; [`../../olares-market/SKILL.md`](../../olares-market/SKILL.md) owns the verbs themselves.

## A crash while the row still says `installing` / `initializing`

A crashlooping pod keeps the row in flight for minutes (the fast-fail grace in the shared **application state machine**), and an in-flight app accepts neither `install` nor `upgrade` (`stop` 404s too). Do not wait out the grace — fix the chart, bump, re-package and re-upload, then clear the row by the state it is in:

| Row state | Clear it with | Then |
|---|---|---|
| `pending` / `downloading` / `installing` | `olares-cli market uninstall <app> --watch` — cancels first; this cancel tears the partial install down | `market install <app> -s upload --version <NEW> --watch` |
| `initializing` / `upgrading` / `resuming` | `olares-cli market cancel <app> --watch` — the app lands in `stopped` | `market upgrade <app> -s upload --version <NEW> --watch`, then `market resume <app> --watch`: an upgrade from `stopped` stays `stopped` |
| `stopped` after the install, `reason` `StartUpFailed` / `InitFailed` / `Unschedulable` (app-service gave up; `install --watch` reports it) | nothing to clear — the app is installed and stopped | the same upgrade and resume; or `market uninstall` and `install` the new version |

`applyingEnv` cannot be cancelled; wait for it to end in `running` or `applyEnvFailed`.

## `stopped` after upgrade

An upgrade can leave the **market row** in `state=stopped` while the **workload** is actually `Running`. Two paths land in `stopped`: upgrading an **already-stopped** app re-renders the chart at `replicas=0` and intentionally returns to `stopped` (by design); and **canceling an in-flight** op (`initializing` / `upgrading` / `applyingEnv` / `resuming`) only *stops* the app — so if a crashing initContainer was fixed and the workload later came up on its own, the row can read `stopped` while the pod is `1/1 Running`:

```bash
olares-cli market status <app> -s upload   # state=stopped
olares-cli cluster application status <ns> # Deployment 1/1 Running
```

This is **not** a failure — the market row just needs to be resumed. Recovery:

```bash
olares-cli market resume <app> --watch
```

`resume` scales the workloads back up and waits for startup (`stopped → resuming → running`). If the pod is already running it completes quickly and flips the market row to `running`.

## `upgradeFailed`

If an upgrade instead left the app in **`upgradeFailed`** (the upgrade itself errored, not a `stopped` row), recover by fixing the chart, **bumping the version**, re-packaging + re-uploading, and re-running `market upgrade <app> -s upload --version <NEW> --watch` — `upgradeFailed` is an upgradable state. The bump is mandatory whenever the chart changed, because the old version cannot be re-uploaded; a same-version upgrade is only for retrying the stored chart after a transient failure. Do **not** fall back to `install`: app-service rejects `install` from `upgradeFailed`, which only re-wedges the row.
