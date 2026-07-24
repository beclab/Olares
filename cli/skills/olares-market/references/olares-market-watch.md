# market `--watch` behavior, agent discipline, stuck states, errors

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. Verb details are in the parent lifecycle row; state-machine facts (states, transitions, fail TTLs, `running` semantics) live in the shared **application state machine**.

## `--watch` interaction with each verb

| Verb | Terminal-success buckets | Idempotent shortcut |
|---|---|---|
| `install` | `running` | — (no watcher shortcut; requires `OpType=install`) |
| `upgrade` | `running` (matchOpType=upgrade) | — (handled by pre-flight) |
| `uninstall` | `uninstalled`, row disappears (or `*Canceled` when an in-flight app was auto-canceled) | App already uninstalled → returns immediately; in-flight apps are canceled first, then uninstalled if still present |
| `clone` | `running` on the new clone name | — |
| `stop` | `stopped` | Already stopped → returns immediately |
| `resume` | `running` | Already running → returns immediately |
| `restart` | `running` with `statusTime` newer than the pre-request baseline | — (the initial `running` row must not short-circuit the stop-then-resume cycle) |
| `cancel` | Any "stopped moving" state | — |

### Per-op foreground watch windows

`--watch` defaults to a 15m timeout, but progressing states have very long backend TTLs (`downloading` = 30 days; `installing` 30m; `initializing`/`upgrading` 1h — see the shared **application state machine**). Don't sit on the default. Use a short foreground window sized to the verb, then switch to polling:

| Verb / phase | Suggested foreground `--watch-timeout` | After timeout |
|---|---|---|
| `stop` / `cancel` / `resume` / `restart` / `uninstall` | `30s` | poll `market status <app> --watch --watch-interval 5s` |
| `install` deploy phase (post-download) / `upgrade` / `clone` | `1m` | poll `status`, then diagnose if STATE doesn't move |
| `install` while STATE is `downloading` | judge by pull progress, not a timeout (see below) | keep polling patiently — a 30-day TTL means it won't self-fail |

A timed-out short window is **not** a failure — it just means "not terminal yet". Re-judge by the STATE row, never by the PROGRESS number (unreliable).

### `install` download phase is special

When STATE is `downloading`, the app is pulling images and may legitimately stay there for many minutes (multi-GB images), with a 30-day backend TTL — so it will not self-fail. Poll patiently (`market status <app> --watch --watch-interval 5s`); only once it **leaves** `downloading` (into `installing`/`initializing`) do the 1m deploy-phase window and the "stuck" rules apply. A `downloading` row that never advances AND whose byte-level pull progress is flat is a *stalled* pull, not a slow one — diagnose via [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md) (it shows where real pull progress lives). Judge by STATE, not PROGRESS.

### Verifying an app is actually healthy

`state=running` only proves each entrance is **TCP-reachable**, not that the app serves correctly. Treat `running` as necessary-but-not-sufficient. If a freshly-installed app is `running` but anything looks off (running-but-unreachable, crashloop, soft-hang), hand it to [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md), which owns the full health ladder and triage.

## Agent best practices

- **For "install X and tell me when it's running"** → `market install X --watch -o json`, then parse `.finalState`.
- **For "upgrade X if there's a newer version"** → `market get X -o json` to check version, then `market upgrade X --watch`. The pre-flight will short-circuit if there's no newer chart.
- **For "re-apply an upload chart I just re-uploaded"** → `market upgrade X -s upload --version <same-version> --watch`. The same-version upgrade is allowed for `-s upload` (gate 3 exception) and is the right verb once the app already exists (`running` / `upgradeFailed` / ...) — `install` would be rejected by app-service in those states.
- **For "stop everything for this user"** → `market list --mine -o json | jq -r '.[].name'` + a shell loop calling `market stop`. The cluster doesn't expose a bulk-stop verb.
- **For "install a custom chart"** → `market upload ./mychart.tgz` (always lands in source `upload`), then `market install <name> -s upload`.
- **For ambiguous source rows on uninstall/stop/resume**: the verb already resolves automatically. Don't pass `-s` even when the SPA shows it under multiple sources.
- **Don't block on a long foreground watch.** Use a short `--watch-timeout` per the table above; on timeout switch to `market status <app> --watch --watch-interval 5s`, or fire-and-forget the mutation and poll `market status <app>` periodically. Judge by STATE, not PROGRESS.
- **For "install X and watch it without hanging"** → `market install X --watch --watch-timeout 1m -o json`; if it returns non-terminal (still `downloading`/`installing`), poll `market status X --watch --watch-interval 5s` and only diagnose once a short window passes with no STATE movement.

## Stuck in installing / initializing

A long `installing` / `initializing` is NOT a failure — app-service polls a long TTL before giving up, and two non-obvious traps make `--watch` misleading here (a soft-hang is never fast-failed; a scheduling failure ends in `stopped`, not `installFailed` — both are in the shared **application state machine**). Operational discipline: let a **short window** (~1m, post-download) run; if STATE hasn't moved, **stop watching passively and diagnose**.

**Hand stuck/won't-start installs to [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md)** — it owns the symptom→root-cause routing (queue vs stalled download vs scheduling failure vs soft-hang), the namespace resolution, and the exact pod/log/event commands. This reference does not duplicate that triage.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `missing required env var(s): KEY1, KEY2 ...` (install) | App declares required envs | Re-run with `--env KEY=VALUE` per missing var |
| `app 'X' is not in an upgradable state (current: Y)` | Pre-flight gate 2 | Wait for terminal state, or run `cancel` first |
| `target version '1.2.3' is already installed — nothing to do` | Pre-flight gate 3 | Nothing to upgrade. **Does NOT fire for `-s upload`** — same-version upgrade is allowed there to re-apply an overwritten chart |
| `chart is marked 'suspend' or 'remove' in source 'X' ...` | Pre-flight gate 4 (`app_simple_info.app_labels` contains `suspend` or `remove`) | Upstream withdrew the app; the SPA hides its Upgrade button too. Contact the app maintainer |
| `app 'X' is not cloneable` | `clone` against an app that is neither multi-instance nor a template | Check `market get X -o json` for `allowMultipleInstall` / `templateOnly` |
| `--title is required` | `clone` without `--title` | Add `--title "..."` |
| Watcher hangs near `*Failed` | Backend op failed | `market status <app>` to inspect; `market cancel <app>` if applicable |
| `--cascade auto-enabled ...` (stderr) | 1.12.5 C/S v2 single-user cluster | Informational; override with `--cascade=false` if needed |
| `--cascade force-enabled ... (CS/shared apps always cascade)` (stderr) | 1.12.6 CS/shared app | Informational; `--cascade=false` cannot disable cascade on 1.12.6 |
