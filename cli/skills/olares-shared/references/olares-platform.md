# Olares platform concepts (single source of truth)

> Cross-skill platform facts that ≥2 `olares-cli` skills rely on. Each section notes which skill reads it from which angle. Skills link here from their `SKILL.md` (one hop) instead of re-describing these facts. Pure platform model — no login/profile content (that is in [`../SKILL.md`](../SKILL.md)).

## Contents

- [Userspace storage model](#userspace-storage-model) — `files` (addressing) + `chart` (mounting)
- [Run identity: uid/gid 1000](#run-identity-uidgid-1000) — `chart` (runAsUser) + `files` (chown)
- [System-managed Home directories](#system-managed-home-directories) — `files` (protected names) + `chart` (reserved caches)
- [App, namespace & networking model](#app-namespace--networking-model) — `chart` (shared apps / deps) + `cluster` (application space)
- [System middleware model](#system-middleware-model) — `chart` (wire to it) + `cluster` (middleware aggregator)
- [Olares version & semver model](#olares-version--semver-model) — `chart` (version floor) + `shared` (profile VERSION) + `settings` (me version)

## Userspace storage model

Olares exposes five mountable storage areas. Each has a stable frontend path (how `files` addresses it), a Helm value (how `chart` mounts it), a backend, and a scope.

| Area | Frontend path (`files`) | Chart mount value | `permission` field | Scope | Backend / traits |
|---|---|---|---|---|---|
| **Home** | `drive/Home` | `.Values.userspace.userData` | `userData` (list paths) | user-level (shared by the user's apps granted the perm) | JuiceFS — cross-node, backed up; for **user-visible** files |
| **Data** | `drive/Data` | `.Values.userspace.appData` | `appData: true` | per-app (auto `/<appName>`) | JuiceFS — cross-node, backed up; for **app-private persistent state** (db files, config) |
| **Cache** | `cache/<node>` | `.Values.userspace.appCache` | `appCache: true` | per-app (auto `/<appName>`) | **node-local PV** (`/olares/userdata/Cache/`) — pins the pod to that node via `schedule.nodeName`; fast, regenerable, not guaranteed durable/backed-up |
| **Common** | `drive/Common` | `.Values.userspace.appCommon` | `appCommon: true` | **cross-app shared** (no `appName` suffix) | JuiceFS; reserved `huggingface`/`ollama`/`llama.cpp`/`comfyui` shared caches; needs Olares ≥ 1.12.6 |
| **External** | `external/<node>/<volume>` | `.Values.sharedlib` | `externalData: true` | user's external storage | SMB/NFS/USB volumes attached via LarePass; needs schema ≥ 0.12.0 |

Key facts:

- **Per-app vs shared vs user.** `appData`/`appCache` auto-append `/<appName>` (app-private); `appCommon` is a bare cross-app dir (every app with the perm sees the same dir — that is what makes shared model caches work); `userData` is the user's `/Home`.
- **Backend decides scheduling + durability.** `userData`/`appData`/`appCommon` are JuiceFS (cross-node, backed up). `appCache` is a node-local PV, so app-service pins the pod to that node — fast local disk, treat as disposable.
- **Owner is uid 1000.** All five are read/written as uid/gid 1000 (see next section).
- **Version gates.** `appCommon` needs Olares ≥ 1.12.6; `externalData`/`sharedlib` needs `olaresManifest.version` ≥ 0.12.0.
- **Drive's `extend` must be `Home` or `Data` exactly** — `home` is rejected with `invalid drive type`.

## Run identity: uid/gid 1000

Every userspace area above is owned and accessed as **uid/gid 1000**. This is the shared root cause behind two skills:

- `chart` — the app process must run as 1000 (`spec.runAsUser: true` injects `pod.spec.securityContext.runAsUser: 1000`). If the image runs as root or another uid, writes to userspace mounts fail with `Permission denied` or never persist. `appCommon` is created `chown 1000:1000`.
- `files` — `chown` UID conventions: 0 (root) vs 1000 (the userspace owner). Only `drive/Home`, `drive/Data`, `cache/<node>` accept `chown`.

OPA admission: a non-trusted image running as root (or `privileged`/`runAsNonRoot: false`) is denied. Init containers may run as root only with a trusted `beclab/` image. Chart-side alignment recipes (Dockerfile `USER`, initContainer `chown`, OPA boundaries) live in `chart`'s run-as-user reference.

## System-managed Home directories

These eleven names directly under `drive/Home/` are LarePass bootstrap directories that user apps look up by exact name:

```
Pictures  Music  Movies  Downloads  Documents  Code  Cache  Data  Home  Ollama  Huggingface
```

Platform invariants:

- They are created and managed by LarePass; apps depend on the exact names (e.g. the model-runtime app's `Ollama` cache, the LarePass UI's `Pictures` sidebar tile).
- Casing is significant: `Huggingface` is one word (not `HuggingFace`).
- `files` enforces them as protected names: `rename` / `rm` / `mv source` refuse them at the **first level under `drive/Home/` only**; `cp` (copy) is allowed; nested content (`drive/Home/Pictures/Trip2024/`) is fully editable; other namespaces (`drive/Data/Pictures`, `sync/...`) are unaffected.

## App, namespace & networking model

How Olares apps are placed in namespaces and reach each other:

- **Per-user app** → namespace `<app>-<owner>`. This is the default and what a normal install produces.
- **Shared app** (`apiVersion: v3` on Olares ≥ 1.12.6) → deterministic namespace `<app>-shared`, admin-only install, cluster-wide. app-service rewrites the namespace regardless of what the manifest says.
- **Application space** (`cluster` framing) is a KubeSphere-grouped K8s namespace; the same namespace, grouped by workspace.
- **Cross-namespace reachability.** A v3 shared app gets a service-mesh sidecar + NetworkPolicy so other namespaces can call it. Consumers reach it by plain in-cluster Service DNS — **no entrance/URL**.
- **Dependency injection.** When app B declares a `type: application` dependency on a shared app A, app-service injects A's Services into B's Helm values as `.Values.svcs.<svcName>_host` (= `<svcName>.<app>-shared`) and `.Values.svcs.<svcName>_ports`.
- **`sharedEntrances` was removed in 1.12.6+.** Do not add it; treat any leftover as stale.

Used by: `chart` (authoring shared apps and `application` dependencies) and `cluster` (the application-space / namespace runtime view).

## System middleware model

Olares manages shared datastores so apps do not bundle their own:

- **Always available, no admin pre-install:** PostgreSQL (v17, Citus image, ships pgvector/pgvecto.rs/vchord/postgis/zhparser + standard contrib) and Redis. Reach for these by default.
- **Require an admin to install from the Market first:** MongoDB, MySQL, MariaDB, MinIO, RabbitMQ, NATS.
- These are NOT native K8s resources — `cluster` surfaces them through a separate `/middleware/v1/*` aggregator (the `middleware` noun), not the K8s API.

Used by: `chart` (replace a bundled db, wire to `.Values.<mw>.*`, declare an `options.dependencies` `type: middleware` entry) and `cluster` (listing Olares-managed middleware). Single-instance only for ported apps (do not declare `distributed`).

## Olares version & semver model

Olares releases follow [semver](https://docs.olares.com/developer/install/versioning.html) — `Major.Minor.Patch[-PreRelease]` (stable `1.12.6`, RC `1.12.0-rc.0`, daily `1.12.0-20241201`).

- The running version lives in the `Terminus` CR `spec.version` and is injected into every chart as `.Values.sysVersion`.
- **"At least" comparisons strip the prerelease/build segment**, so a daily build `1.12.6-20260327` still counts as `>= 1.12.6`. Always use the `-0` suffix in constraints (`>=1.12.6-0`) or prerelease/daily/RC builds fail to match.
- Reading the target version: `olares-cli profile list` (cached `VERSION` column, populated at login from `/api/olares-info`), `olares-cli profile list --refresh-version`, or live `olares-cli settings me version`.

Used by: `chart` (the `>= 1.12.6` porting floor + `options.dependencies` `type: system`), `shared` (the profile `VERSION` column), and `settings` (`me version`).
