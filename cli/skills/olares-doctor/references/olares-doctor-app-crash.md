# doctor: app crashes / restarts / won't start

> **Prerequisite:** read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Backend facts** (fast-fail grace, `running` semantics): [`../../olares-shared/references/olares-platform-appstate.md`](../../olares-shared/references/olares-platform-appstate.md).

Symptom: the app's container keeps restarting (CrashLoopBackOff), exits non-zero, or fails to start with a config/permission error.

## Catch it early — don't wait out the grace window

During `initializing`, app-service polls entrance TCP reachability and only declares failure once `hasUnrecoverablePod` sees `CrashLoopBackOff` with `RestartCount >= 5` **persisting past a 5-minute grace**. So the market row legitimately stays `initializing` for minutes while the container is already crashlooping — **the fast signal is the pod, not the row.** Resolve the namespace (`<app>-<owner>` / `<app>-shared`; see [finding an app's namespace](../../olares-shared/references/olares-platform.md#finding-an-apps-namespace)) and watch the **main** container directly:

```bash
olares-cli cluster pod list -n <ns> -o json    # status.containerStatuses[].{ready,restartCount,state.waiting.reason}
```

`restartCount` climbing or `state.waiting.reason == CrashLoopBackOff` on the main container = start diagnosing now.

## Get the crash reason

```bash
# Current logs, and the buffer from the instance that just died (where the real traceback usually is).
olares-cli cluster pod logs <ns>/<pod> -c <main-container>
olares-cli cluster pod logs <ns>/<pod> -c <main-container> --previous   # mutually exclusive with -f
olares-cli cluster pod get <ns>/<pod>          # exit code / reason / last state
olares-cli cluster pod events <ns>/<pod>       # mount / config / pull events
```

(`cluster pod` flags & semantics: [`../../olares-cluster/references/olares-cluster-pod.md`](../../olares-cluster/references/olares-cluster-pod.md).)

## Common root causes -> next step

| What you see | Root cause | Next step |
|---|---|---|
| `CreateContainerConfigError` | Missing/!invalid env, secret, or configmap referenced by the container | Read events for the missing key; for a catalog app re-check required envs (`market install ... --env`); for a chart you author, fix the manifest env wiring |
| `CrashLoopBackOff`, app traceback in logs | App-level error (bad config, missing dependency, unreachable middleware) | Read the traceback; wire middleware/env correctly |
| Exit 0 / `Completed` with **empty logs**, or app reads a bogus port/host | k8s service-link env collision (`<SVC>_PORT=tcp://...` clobbers app config) | Chart fix: `enableServiceLinks: false` — [`../../olares-chart/references/olares-chart-env.md`](../../olares-chart/references/olares-chart-env.md) |
| `Permission denied` / EACCES writing data, or data not persisting | uid != 1000 on userspace mounts (root-owned dirs, missing `runAsUser`) | Chart fix: [`../../olares-chart/references/olares-chart-run-as-user.md`](../../olares-chart/references/olares-chart-run-as-user.md) |
| Admission denied: untrusted image runs as root | OPA blocks a root third-party main container | Chart fix: force uid 1000 / initContainer chown — [`../../olares-chart/references/olares-chart-run-as-user.md`](../../olares-chart/references/olares-chart-run-as-user.md) |
| Image can't be pulled (`ImagePullBackOff` / arch) | Not a crash — a pull problem | [olares-doctor-image.md](olares-doctor-image.md) |

> **Diagnosis vs fix:** this reference finds the root cause for any app. When the app is **one you are authoring**, the fix is a chart edit — hand back to [`../../olares-chart/SKILL.md`](../../olares-chart/SKILL.md) (then re-lint + re-deploy). For a published catalog app, the fix is config (`settings` / `market install --env`) or contacting the maintainer.
