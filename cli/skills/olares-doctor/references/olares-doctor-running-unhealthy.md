# doctor: running but unreachable / unhealthy

> **Prerequisite:** read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Backend fact:** `state=running` only proves entrance **TCP reachability**, not health — shared **application state machine**.

Symptom: `market status` shows `state=running`, but the app's page is blank, returns 5xx, times out, or the entrance doesn't respond. `running` was set the moment each entrance's host:port accepted a TCP connection — the L4 socket is open, but L7 (HTTP) may be broken. Climb the ladder to find which layer fails.

## Health ladder

```bash
# 1. Market row — necessary, not sufficient.
olares-cli market status <app> -o json                 # state=running ?

# 2. Pod readiness — is the container actually Ready, restarts stable?
olares-cli cluster application status <ns>             # Deployment X/Y Ready
olares-cli cluster pod list -n <ns> -o json            # ready=true, restartCount stable

# 3. App logs — fatal errors, panics, repeated restarts behind a still-open socket.
olares-cli cluster pod logs <ns>/<pod> -c <main-container> --tail 200
```

(Namespace resolution: finding an app's namespace.)

| Finding | Root cause | Next step |
|---|---|---|
| Pod not `Ready` (`0/1`) but socket open | App listens but readiness/HTTP not up (soft-hang) | Read logs for startup blockers (waiting on middleware, slow migration) |
| Pod restarting behind the open socket | Crash after the port opened | **doctor: app crash** |
| Pod Ready, logs clean, but entrance 504 / closes at ~15s on a long request | Entrance proxy route timeout `options.apiTimeout` defaults to 15s | Chart fix: manifest `options.apiTimeout` |
| Pod Ready, logs clean, entrance still wrong host/port | Entrance wired to the wrong service port | Chart fix: manifest entrance host/port |
| `StudioSource` (Devbox) app reads `running` immediately | Studio apps skip the launch probe entirely (`running` proves nothing about reachability) | Verify by pod readiness + HTTP directly |

> **Diagnosis vs fix:** this reference locates the failing layer. Entrance/timeout fixes for an app **you author** are chart edits — hand back to [`../../olares-chart/SKILL.md`](../../olares-chart/SKILL.md). For a catalog app, an entrance/domain tweak may be a `settings` change instead.
