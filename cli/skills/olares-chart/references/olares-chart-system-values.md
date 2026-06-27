# System-injected Helm values — the platform render context

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This is the companion to [olares-chart-env.md](olares-chart-env.md): env covers `.Values.olaresEnv.*`; this file covers everything **else** app-service injects at render time.

When app-service installs an app it injects a large set of `.Values.*` into the Helm render — the running app's identity, storage paths, domains, OIDC, middleware connections, and cluster facts. These split into **two categories**:

| Category | Helm key shape | Owns it | Reference |
|---|---|---|---|
| **Environment variables** | `.Values.olaresEnv.<name>` | system / user / app env (three levels) | [olares-chart-env.md](olares-chart-env.md) |
| **System-injected values** | `.Values.<many>` (below) | app-service, at render time | this file |

> **Nothing is auto-injected into the container.** A system-injected value only lives in `.Values.*` — you MUST map it into the workload yourself (a container `env:`, a volume `hostPath`, a connection string, an entrance host). The same explicit-mapping rule as `olaresEnv`.

> **Render/lint placeholders differ from a real install — in both directions.** During `lint` and dry-run rendering, app-service fills *placeholder* values so templates render without a cluster; a real install only sets feature-gated keys when the feature is actually used. So:
> - **Present in render, may be absent at install:** `oidc.*`, `os.*`, `dep.*`, `svcs.*`, and the middleware keys exist as placeholders but are only set for real when the app enables OIDC / declares provider permissions / has the relevant dependency / declares the middleware. A template that references them unconditionally renders fine in `lint` but resolves to empty at install when the feature is off — guard with `if` / `default`.
> - **Absent in render, present at install:** the placeholder `userspace` only carries `appCache` + `userData`, so `.Values.userspace.appData` / `.appCommon` render empty under `lint` but are filled at install when the matching `permission` is granted.
> - `lint` does **not** flag either case; only a real install surfaces it.

## A. Identity / cluster / system

Populated in `BuildBaseHelmValues` for every install (mostly from app-service's own environment and cluster queries).

| Value | Meaning / source | Possible values |
|---|---|---|
| `.Values.bfl.username` | the app owner's Olares username | free-form |
| `.Values.isAdmin` | `true` if the owner is an admin | `true` \| `false` |
| `.Values.admin` | admin username (rewritten to the owner of an already-installed cluster-scoped instance when present) | free-form |
| `.Values.sysVersion` | running Olares version, from the `Terminus` CR — see [platform.md](../../olares-shared/references/olares-platform.md) (version model) | semver, e.g. `1.12.6` |
| `.Values.GPU.Type` / `.Values.gpu` | selected GPU flavour the platform injected at install | `cpu` \| `nvidia` \| `amd-gpu` \| `amd-apu` \| `strix-halo` \| `nvidia-gb10` \| `apple-m` \| `mthreads-m1000` — the runtime flavour set is a **superset** of the `lint`-accepted `spec.accelerator` modes in [olares-chart-accelerator.md](olares-chart-accelerator.md) (some real node flavours aren't valid manifest modes) |
| `.Values.GPU.Cuda` | CUDA version (`OLARES_SYSTEM_CUDA_VERSION`) | version string, may be empty on non-NVIDIA |
| `.Values.cluster.arch` | node CPU arch | `amd64` \| `arm64` |
| `.Values.nodes` | per-node hardware metadata; advanced, rarely templated. Each entry: `cudaVersion`, `cpu[]` (`coreNumber`/`arch`/`frequency`/`model`/`modelName`/`vendor`), `memory.total`, `gpus[]` (`vendor`/`arch`/`model`/`memory`/`modelName`) | list of NodeInfo |
| `.Values.deviceName` | device name reported by olaresd | free-form |
| `.Values.rootPath` | Olares root path on the host (defaults to the platform default) | path |
| `.Values.downloadCdnURL` | system CDN base (`OLARES_SYSTEM_CDN_SERVICE`) | URL |
| `.Values.fs_type` | rootfs type (`OLARES_SYSTEM_ROOTFS_TYPE`) | `fs` (default, local) \| `jfs` (JuiceFS) — templates branch on `{{ if eq (.Values.fs_type \| default "fs") "jfs" }}` |
| `.Values.sharedlib` | External storage base path (`/Files/External/<deviceName>/`) — the External area in [platform.md](../../olares-shared/references/olares-platform.md) (storage model) | path |
| `.Values.workloads.<name>.replicaCount` | per-workload replica count; present whenever the manifest declares `workloadReplicas` (required). Each listed Deployment/StatefulSet **must** wire `spec.replicas: {{ .Values.workloads.<name>.replicaCount }}` — app-service overrides this value for install (staged at 0) and suspend/resume, so a hardcoded `replicas` makes those operations inert. See [olares-chart-manifest.md](olares-chart-manifest.md) (Workloads & replicas). | integer |

```yaml
        env:
        - name: APP_VERSION
          value: "{{ .Values.sysVersion }}"
        - name: IS_ADMIN
          value: "{{ .Values.isAdmin }}"
```

## B. Runtime context (per-install)

Populated in `SetValues` from the install's owner, entrances, permissions, and dependencies.

| Value | Meaning |
|---|---|
| `.Values.user.zone` | the owner's zone (used to build app URLs) |
| `.Values.domain.<entranceName>` | the public host for each entrance — `<appid>.<zone>` (or `<appid><i>.<zone>` for multiple entrances) |
| `.Values.schedule.nodeName` | node the pod is pinned to (set when the app requests storage permissions, so it co-locates with its node-local data) |
| `.Values.userspace.appData` / `.appCache` / `.userData` / `.appCommon` | userspace mount host paths — granted per declared `permission`; defined in [platform.md](../../olares-shared/references/olares-platform.md) (storage model) and used in [manifest.md](olares-chart-manifest.md) §2. `appData` / `appCache` already include the per-app suffix (`.../Data/<appName>`, `.../Cache/<appName>`) |
| `.Values.os.appKey` / `.Values.os.appSecret` | provider-access credentials, issued when the app declares provider permissions |
| `.Values.dep.<entrance>_host` / `_port` | host/port of a cluster-scoped app this app depends on |
| `.Values.svcs.<svc>_host` / `_ports` | Services of a depended-on v3 shared app — see [platform.md](../../olares-shared/references/olares-platform.md) (app/networking) and [olares-chart-shared.md](olares-chart-shared.md) |
| `.Values.oidc.client.id` / `.client.secret` / `.issuer` | OIDC client credentials + issuer URL; **only present when `options.oidc.enabled: true`** |

```yaml
        env:
        - name: PUBLIC_URL
          value: "https://{{ .Values.domain.myapp }}"
        - name: OIDC_ISSUER
          value: "{{ .Values.oidc.issuer }}"
```

A `userspace` path is most often consumed as a `hostPath` volume (the value already ends in `/<appName>`; append a further subdir only if you want to organize within it):

```yaml
      volumes:
      - name: app-data
        hostPath:
          path: {{ .Values.userspace.appData }}
          type: DirectoryOrCreate
```

## C. System middleware connections

When the manifest declares a `middleware:` block, app-service injects each datastore's connection info as `.Values.<mw>.*` (one of `postgres` / `redis` / `mongodb` / `mysql` / `mariadb` / `minio` / `rabbitmq` / `nats` / `elasticsearch` / `clickhouse`), each carrying `host` / `port` (+ `username` / `password` for most) plus a type-specific sub-key (`databases` / `buckets` / `vhosts` / `indexes` / `subjects` / `refs`).

> **Authoritative reference:** the exact key shapes, the `{{ .Values.<mw>.* }}` mapping examples, the PostgreSQL extension catalog, and which middleware needs an admin pre-install all live in [olares-chart-middleware.md](olares-chart-middleware.md). Do not duplicate them here.

## Caveats

- **Map it in the template.** A value in `.Values.*` does nothing until you write it into the workload (env / volume / connection string).
- **Conditional keys.** `oidc.*` only exists when OIDC is enabled; `userspace.*` keys only for the permissions actually granted; `dep.*` / `svcs.*` only when the app depends on the relevant app; `workloads.*` only with declared `WorkloadReplicas`. Guard optional keys in templates.
- **`lint` won't catch a missing/typo'd mapping.** As with env, only an actual install surfaces a wrong key.
