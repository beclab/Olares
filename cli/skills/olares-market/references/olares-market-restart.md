# market `restart`

> Read the parent [`../SKILL.md`](../SKILL.md) first. Flags remain authoritative in `olares-cli market restart --help`.

`restart` asks app-service to run a stop-then-resume cycle for an installed app:

```bash
olares-cli market restart firefox
olares-cli market restart firefox --watch --watch-timeout 1m -o json
olares-cli market restart comfyui --compute-binding node-1:gpu-0 --watch
```

## Version and source

- Requires **Olares >= 1.12.6**. Older releases do not expose `POST /apps/restart`.
- Source is implicit: the CLI resolves the installed app's per-user state row and sends its source. `restart` does not expose `-s`.
- The request body uses the same `{app_name, source, computeBinding?}` shape as the overlay feature's restart path.
- If the version cannot be determined, confirm the active profile is logged in and run `olares-cli profile list --refresh-version`. If the detected version is below 1.12.6, upgrade Olares.

## Compute binding

`--compute-binding <node>:<device>[:<mem>]` is repeatable and follows the same validation and interactive/non-interactive selection rules as `market resume`. A GPU app may require a binding; explicit rejected bindings surface the backend reason.

## Watch semantics

A completed restart looks like the pre-request row: `state=running`, `opType=resume`. The watcher therefore:

1. Captures the row's `statusTime` before sending the restart.
2. Accepts `running` only when its `statusTime` is strictly newer than that baseline.
3. Watches both phases for `stopFailed`, `resumeFailed`, `resumingCanceled`, or `resumingCancelFailed`.

`restart --watch` has **no** resume-style idempotent `already running` shortcut. A short timeout is not failure; poll `market status <app>` if the cycle is still progressing.
