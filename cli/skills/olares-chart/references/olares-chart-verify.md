# Verify: the app serves and opens from the Desktop

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first; this follows a [deploy](olares-chart-deploy.md) that reached `running`.

## `running` is not the same as serving

`running` only says app-service scaled the workload up: an app whose first write to a userspace mount was refused still reads `running` and answers 403. Before calling it deployed, read the app's own evidence — `cluster pod list -n <ns>` for READY, `cluster container logs <ns>/<pod>/<container>` for its startup and first-request lines, and `settings apps list` for the real host in its `URL` column (**never compute it**; `entrances list` leaves that column empty) — then request that host. Anything unreachable from there is [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md)'s routing.

**`cluster pod exec` requires Olares >= 1.12.7** — it is the direct way to read a mount's owner or curl the app on localhost, and it is gated. The baseline this skill scaffolds against is `>= 1.12.6`, where the commands above are the entire verification surface.
