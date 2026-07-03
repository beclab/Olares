# cluster exec

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & wire shapes:** `olares-cli cluster pod exec --help` and `olares-cli cluster container exec --help`.

Run a command inside a container over the native K8s exec WebSocket. Available as both `cluster pod exec` and `cluster container exec` (same wire, same semantics; the only difference is positional grammar — see below). **One-shot is the agent path; `-it` is human-only.**

## Two modes

| Mode | For | Shape |
|---|---|---|
| **One-shot** (default) | agents | argv after `--`, runs to completion, returns captured output |
| **Interactive** (`-i -t` / `-it`) | humans | allocates a TTY and attaches your terminal, like `kubectl exec -it` |

- **One-shot** returns `{stdout, stderr, exitCode, truncated, durationMs}` (via `-o json`). Judge success by `exitCode` (0 = ok); stdout/stderr are separated. Bounded by `--timeout` (default 60s, `0` = no limit) and `--max-output-bytes` (default 2MiB, `0` = unlimited; on overflow output is cut and `truncated:true`).
- **Interactive** needs a real terminal (no confirmation prompt). A non-TTY caller (an AI tool call) is refused — the TTY requirement itself keeps agents on the one-shot path. With `-it` and **no target**, an interactive picker lists every container visible to your profile (type to filter, arrows to move, enter to select); `-n <ns>` scopes it.

## Container identification

- `pod exec` auto-selects the sole container of a single-container pod when `-c` is omitted; a multi-container pod errors with the candidate list.
- `container exec` requires the container to be identified explicitly: the 3rd path segment `<ns>/<pod>/<ctr>`, or `<ns>/<pod>` + `-c`, or bare `<pod>` + `-n`/`-c`.

## Agent notes

- **Pass argv after `--`** (no implicit shell). For pipes / vars / multi-step, use `-- sh -c '...'`. exec is **stateless** — chain steps in one call rather than expecting `cd` / exports to persist. (Filesystem effects like `apk add` DO persist in the running container; shell / process state does not.)
- **Edit files without an interactive editor** (one-shot is non-interactive; `-i` is interactive-only):
  - heredoc: `-- sh -c 'cat > /path <<EOF`↵`<content>`↵`EOF'`
  - in-place: `-- sh -c "sed -i 's/old/new/' /path"`
  - whole file: `-- sh -c 'printf "%s" "<content>" | tee /path'`
- **Long-running / streaming / watch — make it bounded, never open-ended.** One-shot buffers to completion, so a command that never returns (`watch`, `tail -f`, `top`, `journalctl -f`, a bare server) just blocks until `--timeout` kills it with no exit code. Instead:
  - Snapshot + poll: run a one-shot that returns (`-- ps aux`, `-- tail -n 200 app.log`) and re-invoke every few seconds rather than following a stream.
  - Bound a stream: wrap follow-mode in `timeout`, e.g. `-- sh -c 'timeout 10 tail -f /var/log/app.log'` (returns after 10s).
  - Genuinely long job: raise `--timeout` (e.g. `600`, or `0` to disable), or detach in-container and poll — `-- sh -c 'nohup ./job.sh >/tmp/job.log 2>&1 & echo $!'` then poll `/tmp/job.log` + `kill -0 <pid>`.
- **Fixes are ephemeral:** changes inside a running container revert on pod restart / recreation (rollout, eviction, node drain). For a durable fix, change the image / ConfigMap / Deployment spec via the `workload` path — do not report an in-container change as permanent.

## Examples

```bash
# One-shot, JSON result (agent path).
olares-cli cluster container exec user-system-alice/my-pod/app -o json -- cat /etc/hosts

# Pipe / multi-step in a single stateless call.
olares-cli cluster pod exec user-system-alice/my-pod -c app -- sh -c 'grep -c ERROR /var/log/*.log'

# Bounded sample of a stream.
olares-cli cluster pod exec user-system-alice/my-pod -- sh -c 'timeout 10 tail -f /var/log/app.log'

# Interactive shell (human; needs a real terminal). No target → container picker.
olares-cli cluster pod exec -it
olares-cli cluster pod exec user-system-alice/my-pod -c app -it
```

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| exit code `127` / `126` | command not found / not executable | check the path; the image may lack it |
| `no such file or directory` on `sh` | distroless / scratch image (no shell) | run the binary directly, no `sh -c` wrapper |
| `EROFS` / `EACCES` on writes | read-only or permission-restricted filesystem | can't fix in-container; change the image / spec |
| `-t/--tty requires an interactive terminal` | `-it` from a non-TTY caller (agent) | drop `-it`, run one-shot with `-- CMD` |
| command hangs until `[timed out]` | a never-returning command (`watch`, `-f`) under one-shot | bound it (`timeout`, snapshot+poll) or raise `--timeout` |
| `cluster exec requires Olares >= 1.12.7 ...` | backend older than the ControlHub exec route, or version undetectable | upgrade Olares; if undetectable, log in and `olares-cli profile list --refresh-version` |
| `HTTP 403` on dial | active profile lacks `pods/exec` (server-side SAR) | use a profile/role that has exec; server-side audited by ks-apiserver |
