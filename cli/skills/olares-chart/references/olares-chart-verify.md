# Verify: the app serves and opens from the Desktop

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first; this follows a [deploy](olares-chart-deploy.md) that reached `running`.

## `running` is not the same as serving

`running` only says app-service scaled the workload up: an app whose first write to a userspace mount was refused still reads `running` and answers 403. Before calling it deployed, read the app's own evidence — `cluster pod list -n <ns>` for READY, `cluster container logs <ns>/<pod>/<container>` for its startup and first-request lines, and `settings apps list` for the real host in its `URL` column (**never compute it**; `entrances list` leaves that column empty) — then request that host. Anything unreachable from there is [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md)'s routing.

## Done means it opens from the Desktop

A request phrased as "finish when it can be opened from Olares Desktop" is answered by the entrance record, not by the market row. Read it:

```bash
olares-cli settings apps get <app> -o json
```

Every check below must hold, and the final report quotes the values:

| Field | Must be |
|---|---|
| `.state` and each visible `.entrances[].state` | `running` |
| `.entrances[].invisible` | `false` for the entrance the user clicks |
| `.entrances[].openMethod` | what was asked — `window` for a separate browser tab; see [entrances](olares-chart-entrances.md) |
| `.entrances[].icon` | the intended icon; `curl -sIL <icon>` answers `200` with `content-type: image/*` |
| `.entrances[].url` | a host; `curl -sI https://<url>` answers |

For a `private` entrance an unauthenticated `curl` gets `303` (or `302`) to `https://auth.<olares-id-domain>/?...&rd=https://<url>/` — that redirect is the expected answer, and it proves the route and the Olares login are in place. A `404`, `502` or `503` there means the entrance does not reach the app: back to [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md).

The app's own page behind the login is proven by its startup and first-request log lines above, or by the user opening it; say which one the report relies on.

**`cluster pod exec` requires Olares >= 1.12.7** — it is the direct way to read a mount's owner or curl the app on localhost, and it is gated. The baseline this skill scaffolds against is `>= 1.12.6`, where the commands above are the entire verification surface.
