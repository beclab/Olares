# Run identity: UID/GID 1000 (packaging + deployment)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first (which loads the platform **Run identity** model — uid-1000 ownership of userspace volumes + the OPA root-deny rule) and answer its two run-identity questions.
> This doc covers the chart-side delta: how to align a self-built or third-party image with that convention — in the Dockerfile, in `OlaresManifest.yaml`, and in deployment templates.

Q1 (what uid the process ends up as) and Q2 (who owns the directories it writes) are **independent**. Answering Q1 never settles Q2 — every Q1 answer still has to face it, including the `PUID`/`PGID` one.

## Two fields, both called `runAsUser`

They sit under keys both spelled `spec`, and they are different fields.

| | `OlaresManifest.yaml` → `spec.runAsUser` | template → `spec.template.spec.securityContext.runAsUser` |
|---|---|---|
| Type | boolean | numeric uid |
| Consumed by | the app-service mutating webhook | kubelet |
| Value | 1000 only — the manifest cannot carry another uid | any uid, written by you |
| Reaches | the primary workload, Pod level only | exactly where you write it; container level overrides Pod level |

Two consequences that bite:

- With the manifest switch on, a **Pod-level** uid other than 1000 is silently rewritten to 1000. Container level is left alone — which is what lets a root initContainer coexist with a uid-1000 Pod.
- The switch only reaches the workload whose name equals the app name (a Deployment, else a StatefulSet of that name). A Job, CronJob, DaemonSet or a second Deployment receives nothing; write the numeric field into those templates yourself.

## Q1 — what uid does the process end up as

Probe the image before editing the chart:

```bash
docker inspect <third-party-image> --format '{{.Config.User}}'
docker run --rm --entrypoint /bin/sh <third-party-image> -c 'exec id'
```

The inspect result is the image's declared `USER`; empty means the runtime starts as root. The shell probe replaces both the image entrypoint and its baked-in `CMD` arguments. If `/bin/sh` is absent, run `docker run --rm --entrypoint id <image> -u` and repeat with `-g`; if `id` is also absent, inspect the Dockerfile/entrypoint. Do **not** use `docker run <image> id`: the image entrypoint still runs first and may consume or transform the probe.

| Image `USER` | Typical handling |
|---|---|
| `1000` (or numeric 1000) | The manifest switch below; usually sufficient |
| Non-root but **not** 1000 (e.g. `nginx` / uid 101) | Force `securityContext.runAsUser: 1000`. What usually breaks is a root-owned **runtime** path baked into the image, not the app — fix that (see self-built images below); rebuild only as a last resort. Do **not** instead `chown` the userspace mount to the image's uid: the convention is `1000:1000`, and Files and any other app holding the same permission expect it |
| Root, empty, or `0`; entrypoint supports `PUID`/`PGID` and drops privileges | The `PUID`/`PGID` pattern below |
| Root, empty, or `0`; process stays root | Rebuild/fork the image to run non-root; current OPA does not detect an implicit image user, but a root application process violates the platform run-identity convention |

### Manifest switch (preferred)

Set `spec.runAsUser: true` in `OlaresManifest.yaml`. Optionally reinforce it in the deployment template, since Kubernetes overrides the Dockerfile `USER`:

```yaml
spec:
  template:
    spec:
      securityContext:
        runAsUser: 1000
        runAsGroup: 1000
```

The switch injects a **uid and nothing else** — no `runAsGroup`, no `fsGroup` — so the process gid stays whatever the image declares. Write `runAsGroup: 1000` yourself when the gid matters.

`fsGroup` is not an alternative to Q2. Userspace mounts are `hostPath`, whose mounter reports `Managed: false`, so kubelet applies neither the gid change nor `fsGroupChangePolicy`. It is not "sometimes helpful" — on these mounts it does nothing at all.

### Root entrypoint that drops to `PUID`/`PGID`

Some third-party images deliberately start as root, prepare files or networking, then use `PUID`/`PGID` to launch the application as a non-root user. For these images, forcing Pod-level uid 1000 prevents the entrypoint from completing its initialization.

```yaml
# OlaresManifest.yaml
spec:
  runAsUser: false
```

```yaml
# Deployment env; names vary by image
- name: PUID
  value: "1000"
- name: PGID
  value: "1000"
```

Omitting `spec.runAsUser` is equivalent for webhook injection. Do not set Pod/container `securityContext.runAsUser: 0`; leave the security context absent so the image entrypoint controls its startup identity. Verify from startup logs or a running container that the final application process drops to uid 1000. If it remains root, rebuild or fork the image instead of treating this exception as permission to run the app as root.

### Self-built images

When authoring a Dockerfile (the Image capability), prefer uid/gid 1000 end-to-end:

```dockerfile
RUN addgroup -g 1000 app && adduser -u 1000 -G app -D app
RUN chown -R 1000:1000 /var/lib/myapp
USER 1000
```

An image-time `chown` only survives where nothing mounts over it: a `hostPath` mounted at that path shadows the image's directory along with its ownership, which is why a uid-1000 image can still hit `Permission denied` on its own data dir. Ownership of a mount is settled at runtime, by Q2.

`chown` the **runtime** paths too, not just the data ones. Pid files, sockets, lock files and scratch dirs under `/run`, `/var/run` or `/tmp` are written before the app does anything useful, so a base image that puts them in a root-owned directory (nginx, php-fpm, supervisor) fails at startup under `USER 1000`. In an image you build, relocate the path in its config or make it writable here; in a third-party image, point a configurable runtime path at `/tmp`, or mount a writable `emptyDir` and prepare its ownership with the initContainer below.

Before pushing, re-run the two probes above against your own ref and expect `uid=1000`.

## Q2 — does the app write a userspace mount

If it does not, Q1 was the whole job. If it does, someone has to make the directories `1000:1000` before the app writes — whichever Q1 answer applied.

On the `PUID`/`PGID` pattern that someone is often the image's own root entrypoint; check its startup log before adding a second mechanism. Every other Q1 answer needs the initContainer below, because the process is already non-root by the time it starts.

**Olares does not pre-`chown` the paths it grants.** Declaring `permission.appData` grants the mount; it does not hand you ownership of it. `DirectoryOrCreate` creates the volume root as root, and kubelet creates a missing `subPath` directory as root too, so a uid-1000 app writing to `appData` / `appCache` fails its first write with `Permission denied` without this.

### The permissions initContainer

`chown` the mount root and each subdirectory the app needs — **each one non-recursively**:

```yaml
spec:
  template:
    spec:
      initContainers:
      - name: init-permissions
        image: beclab/aboveos-busybox:1.37.0
        command:
        - sh
        - -c
        - |
          for d in /data /data/config /data/db; do
            mkdir -p "$d" && chown 1000:1000 "$d"
          done
        securityContext:
          runAsUser: 0
        volumeMounts:
        - name: app-data
          mountPath: /data
      containers:
      - name: app
        image: third-party/app:1.2.3
        volumeMounts:
        - name: app-data
          mountPath: /data
```

The initContainer and the main container mount the **same volume at the same path** — the directories prepared here are the ones the app then writes. List every userspace mount the app writes to, plus any `subPath` directory under them. Also set `spec.runAsUser: true` in `OlaresManifest.yaml` unless Q1 put you on the `PUID`/`PGID` pattern. `init-permissions` and `fix-permissions` are both names you will meet in existing charts.

**Add it now, whatever Olares version you target.** Nothing else prepares these directories: the platform injects a uid only, and install prepares just `appCommon` and its reserved sub-caches — never your app's own `appData` / `appCache` roots or the `subPath` dirs under them. Should a later release grow an equivalent injection, the two are idempotent and safe to overlap.

**Red line: no explicit root `securityContext`** anywhere else in the chart. This initContainer is the only exception, and only because `beclab/` is trusted by both the runtime admission policy and the install gate — see [lint.md](olares-chart-lint.md), which also covers why the two trust different image sets.

### Red line: no `chown -R` at runtime

A root initContainer `chown`s a root-owned directory fine, but recursing has been observed to fail with `Operation not permitted` on subdirectories the main container previously created as uid 1000 — which crash-loops the pod and leaves it `Initializing` indefinitely. So a fresh install works and the upgrade of the same chart does not. The cap-dropping layer is environment-specific, below the cluster at the node / container-runtime level; it is **not** the Olares OPA policy, which only denies untrusted-image + root/`privileged` containers and mutates nothing about capabilities.

The non-recursive form above sidesteps this whatever the layer turns out to be: it touches the mount root and the directories it creates, never the tree the app already owns. Pair it with `set -e` safely.

This red line is about the **runtime** initContainer. An image-build `RUN chown -R` is fine and often necessary — at build time there are no uid-1000-owned subdirectories to trip over.

After any template change, re-run `olares-cli chart lint ./<app>` (the Validate-local (lint) step). When a deployed app misbehaves, the symptom-to-cause routing lives in [`olares-doctor`](../../olares-doctor/SKILL.md), under **Symptom routing** → the crash / permission-error row; the fixes it points back to are the two questions above.
