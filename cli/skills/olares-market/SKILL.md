---
name: olares-market
version: 0.0.0-cli.0
description: "Olares Market via olares-cli market — install, upgrade, uninstall, clone, stop, resume, restart apps; catalog, status, chart upload/download, --watch. Use for Olares app store, my apps, 我的应用, install app, restart app, upload chart, download an app chart."
compatibility: Requires olares-cli on PATH and active Olares profile
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# market (App-store v2)

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing, active-profile selection, platform entry points, and the auth proceed/stop gate. Load its auth reference only when login, profile switching, token storage, or auth recovery is actually needed.

Use `olares-cli market <verb> --help` for authoritative syntax.

## When to use

- Browse the catalog or inspect "my apps".
- Install, upgrade, uninstall, clone, stop, resume, restart, or cancel an app.
- Inspect lifecycle status or manage uploaded charts.

> **Mental model:** `market` is **lifecycle and inventory** at the app-store level (install / upgrade / chart push). For runtime K8s objects, settings, or metrics, route to a sibling.

A **model** application is installed here like any other — `install` for a pinned model, `clone` for an engine base whose model is chosen on the form. What is different is what happens next: [`olares-router`](../olares-router/SKILL.md) notices the application and creates the gateway provider that routes to it, then owns the model's own download, engine state and card.

## Fast paths

| Task | Read | First command |
|---|---|---|
| Install a catalog app and know when it started | [watch and diagnosis routing](references/olares-market-watch.md) | `olares-cli market install <app> --watch --watch-timeout 1m -o json`, then read `.status` |
| See what this user has installed | this file | `olares-cli market list --mine -o json` |
| Follow one app's lifecycle | this file | `olares-cli market status <app>` |
| Take an app off, pause it, or stop one mid-flight | [uninstall, stop, resume, cancel](references/olares-market-lifecycle-remove.md) | `olares-cli market status <app> -o json` to see whether it has settled |
| Put a chart you built on this Olares | [publishing a chart](references/olares-market-chart-publish.md) | `olares-cli market upload ./chart.tgz` then `olares-cli market install <name> -s upload` |

`--watch` polls, so a timeout means "not terminal yet" rather than failure, and `running` means every entrance answers TCP rather than that the app works. An install that goes wrong usually lands on **`downloadFailed`**, not `installFailed` — the common faults (an image tag that does not exist, an unreachable registry) fail during the pull, before the install starts — and that difference decides what can clear it: `downloadFailed` accepts only a fresh `install`, while `installFailed` accepts an `uninstall`.

> **`.status` judges the command, `.finalState` names the app's landing state, and only the first is verb-independent.** Under `--watch`, `.status` is `success` once the row settles the way *this* verb intended, so that plus the exit code is what a script tests. `.finalState` is worth reading when the state itself matters — but `running` is the settling state only for `install`, `upgrade` and `restart`; a successful `stop` lands on `stopped` and a successful `uninstall` on `uninstalled`, so a `running` check copied from an install example reports both as failures. Without `--watch` there is no `.finalState` and `.status` is `accepted`, meaning the server took the request, not that the app is up.

## Verb index

| Family | Verbs | Read when triggered |
|---|---|---|
| catalog + inventory | `list`, `get`, `categories`, `status` | [list, `--mine`, and status](references/olares-market-list.md) |
| lifecycle — putting an app on | `install`, `upgrade`, `clone` | All three take `-s` and `--compute-mode`, and `clone`'s new name is `.targetApp` — [install, upgrade, clone](references/olares-market-lifecycle-add.md) |
| lifecycle — taking one off or pausing it | `uninstall`, `stop`, `resume`, `cancel` | None take `-s`; `uninstall` cancels first when the app is mid-flight, and cancelling `resuming` / `upgrading` needs Olares 1.12.7+ — [uninstall, stop, resume, cancel](references/olares-market-lifecycle-remove.md) |
| restart | `restart` | [restart, compute binding, and baseline watch](references/olares-market-restart.md) |
| charts | `upload`, `delete` | Both pin the bucket to `upload`; a published version's bytes are immutable — [publishing a chart](references/olares-market-chart-publish.md) |
| charts | `download` | The read side, and the only one of the three that takes `-s` — [pulling a chart back out](references/olares-market-chart-download.md) |
| watching / stuck operations | lifecycle `--watch`, `status --watch` | [watch and diagnosis routing](references/olares-market-watch.md) |

## Source resolution (cross-cutting)

- `market.olares` is the public catalog and the default for browse and source-aware lifecycle verbs.
- `upload` is the Local Sources bucket. `upload` and `delete` always use it; `download` defaults to it.
- `cli` and `studio` are legacy/development sources used for read-only inventory.
- `--all-sources` is read-only. `uninstall`, `stop`, `resume`, and `restart` resolve the user's state row rather than accepting a source.
- Pass `cancel --source` only when the state row is gone or unreadable and the user explicitly knows the source.

## App lifecycle / state machine

Load the shared [application-state model](../olares-shared/references/olares-platform-appstate.md) when interpreting states, transitions, fail TTLs, serialized downloads, or `running`.

`State` and `OpType` are separate. After a mutation, an old terminal-looking state can remain visible before `OpType` changes. Mutating watchers therefore require the requested operation to be observed before accepting success; `uninstall` may also succeed when the row disappears. `status` and `cancel` are intentionally operation-agnostic.

## Inventory decisions

- "My apps" → `market list --mine`; this includes useful transitional and failure rows, not only completed installs.
- One app's lifecycle → `market status <app>`.
- Running apps → `market status --all-sources`, then filter `STATE=running`.
- Catalog discovery → `market list`.

## Safety and escalation

- Treat the user's named lifecycle request as task-scope authorisation. Within an authorised deploy/debug loop, install, upgrade, restart, uninstall and clean reinstall do not need repeated confirmation.
- Ask again when the app/source is ambiguous, deletion expands to user data, or the action is outside the authorised task.
- Compute bindings come from `settings compute list`; do not invent node/device identifiers.
- A stuck lifecycle operation is a cross-command diagnosis: inspect status, then use `olares-doctor` for pods, events, logs, images, and resources.
- Stop when credentials, required environment values, compute choices, or the target app/source are ambiguous. Follow the shared auth gate for authentication failures.
