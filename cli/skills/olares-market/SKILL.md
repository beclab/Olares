---
name: olares-market
version: 4.10.0
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

## Verb index

| Family | Verbs | Read when triggered |
|---|---|---|
| catalog + inventory | `list`, `get`, `categories`, `status` | [list, `--mine`, and status](references/olares-market-list.md) |
| lifecycle | `install`, `upgrade`, `uninstall`, `clone`, `stop`, `resume`, `cancel` | Canceling `resuming` / `upgrading` requires Olares 1.12.7+; [lifecycle decisions](references/olares-market-lifecycle.md) |
| restart | `restart` | [restart, compute binding, and baseline watch](references/olares-market-restart.md) |
| charts | `upload`, `download`, `delete` | [chart management](references/olares-market-charts.md) |
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

## `--watch` semantics (lifecycle verbs)

- One-shot mutations return after acknowledgement; the server usually continues asynchronously.
- Watch is polling. Use a short foreground window; a timeout means only "not terminal yet", not failure.
- Judge movement by state transitions, not progress percentage. If state stops moving, route to [`olares-doctor`](../olares-doctor/SKILL.md).
- `stop` on stopped and `resume` on running may finish as idempotent no-ops. Restart and upgrade capture a pre-request `statusTime` baseline and require a newer state before accepting success.

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
