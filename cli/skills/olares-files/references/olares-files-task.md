# files task

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shape:** `olares-cli files task --help`, `olares-cli files task cancel --help`.
> **Needs Olares >= 1.12.6** (shares the archive surface).

Control the per-node task queue that backs the asynchronous file verbs (`compress` / `extract`). They return a `task_id`; this is how you act on it AFTER it's queued.

## Sub-commands

| Sub-command | Purpose |
|---|---|
| `task cancel <task-id>` | Drop one queued / running task |
| `task cancel --all` | Drop EVERY task on the node |
| `task pause <task-id>` | Suspend an in-flight task |
| `task resume <task-id>` | Resume a paused task |

## Per-node model

Tasks are **per-node**. The `<node>` segment must match the node the task was queued on — `compress` / `extract` print it in their "queued ... task: <id> ... node=<node>" line. When `--node` is omitted, it resolves the master node from `/api/nodes/` (same cascade as `files cp`).

## Controllability precheck

For a single task, `pause` / `resume` / `cancel` first read the task's state and refuse client-side when:

- it is already terminal (`completed` / `failed` / `cancelled`) — nothing to act on, or
- the server reports `pause_able=false` (its type/phase is not interruptible).

Pass `-f / --force` to skip the precheck and send the request anyway.

## Safety constraints

- **`task cancel` mutates server state — confirm intent with the user.**
- **`task cancel --all` drops every task on the node, including ones started elsewhere.** It prompts for confirmation; in a non-TTY context it refuses without `--force`.

## Examples

```bash
# Cancel a specific task on the default (master) node.
olares-cli files task cancel 6f1c2e3a-...

# Cancel a task on an explicit node.
olares-cli files task cancel 6f1c2e3a-... --node olares

# Cancel everything on a node (asks for confirmation).
olares-cli files task cancel --all --node olares

# Pause / resume a long-running compress.
olares-cli files task pause 6f1c2e3a-...
olares-cli files task resume 6f1c2e3a-...
```

## Wire shape

```
DELETE /api/task/<node>/?task_id=<id>        (cancel one)
DELETE /api/task/<node>/?all=1               (cancel all)
POST   /api/task/<node>/?task_id=<id>&op=pause
POST   /api/task/<node>/?task_id=<id>&op=resume
```

## Agent notes

- **This is the correct way to stop a runaway `compress` / `extract`.** Ctrl-C on a `--wait` only detaches the local poll; the server-side task keeps running until `task cancel`.
- **Keep the `<node>` from the queue line.** A `task` verb against the wrong node won't find the `task_id`.
- A single `task cancel/pause/resume` without `--force` may exit non-zero with a "already <status>" / "not controllable" message — that's the precheck, not a transport failure.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `pass either a <task-id> or --all, not both` / `need a <task-id>` | `cancel` invoked with both or neither | Pass exactly one |
| `task ... is already <status>; nothing to <op>` | Task already terminal | Nothing to do, or `--force` |
| `not controllable: ... pause_able=false` | Task type/phase isn't interruptible | `--force` to send anyway |
| `refusing to cancel all tasks without --force` (no TTY) | `--all` in a non-interactive context | Re-run interactively or add `--force` |
