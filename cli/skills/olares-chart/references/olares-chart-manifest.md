# Refining OlaresManifest.yaml — the four refinement areas

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and [olares-chart-publish-targets.md](olares-chart-publish-targets.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a working chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. The four areas below are what kompose cannot decide. **§1 Metadata depth depends on release target; §2–§4 are functional and required for both targets.**

## Schema version and apiVersion

`olaresManifest.version` (the manifest **schema**: `0.8.0` default vs `0.12.0` via `--new-schema`) and the top-level `apiVersion` (skill sets **`v3`**) are separate axes from the chart/app versions. Use `0.12.0` when the app needs `spec.accelerator` or `permission.externalData`. The full schema-field table, the version-field map, and the `type: system` dependency are in [olares-chart-versioning.md](olares-chart-versioning.md); accelerator sizing is in [olares-chart-gpu.md](olares-chart-gpu.md).

## 1. Metadata

Depth is gated by release target — see [olares-chart-publish-targets.md](olares-chart-publish-targets.md).

### Always required (`lint` structural check)

```yaml
metadata:
  name: myapp                 # must match folder + Chart.yaml name; do not change casually
  title: My App               # stub title=name is OK for local-run
  description: One-line summary
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp  # default OK for local-run
  version: 0.0.1              # Chart Version — MUST equal Chart.yaml `version`
  categories:
  - Utilities                 # stub OK for local-run; lint does not enum-check
spec:
  versionName: "1.2.3"        # upstream app version; tracks Chart.yaml `appVersion`
  runAsUser: true             # optional but recommended — Olares injects pod runAsUser 1000; see run-as-user.md
```

### local-run: optional (keep stub unless user cares)

Keep the stub: `Utilities` category, default icon, and empty `spec.developer`/`submitter`/`website`/`sourceCode`/`fullDescription`/`featuredImage`/`promoteImage`/`locale`/`supportArch` are all fine (skip `supportArch` unless using accelerator modes).

### market-distribute: required (Market listing + GitBot)

Fill from the upstream project (or ask the user):
```yaml
metadata:
  title: My App               # ≤30 chars
  description: One-line summary shown under the title
  icon: https://.../icon.png  # PNG/WEBP, 256x256, ≤512KB
  categories:                 # BOTH 1.11 + 1.12 values — GitBot enum-checks these
  - Productivity
  - Productivity_v112
spec:
  developer: Upstream Author
  submitter: Your Name
  website: https://project.example
  sourceCode: https://github.com/org/project
  fullDescription: |
    Longer Market description.
  locale: [ en ]
  supportArch: [ amd64, arm64 ]   # must match image platforms
  featuredImage: https://.../hero.webp
  promoteImage:
  - https://.../screenshot1.webp
```

Category values: [manifest docs](https://docs.olares.com/developer/develop/package/manifest.html#categories); listing images: [promote-apps](https://docs.olares.com/developer/develop/promote-apps.html).

## 2. Storage (compose volumes → Olares userspace)

> **Same for both release targets** — functional requirement, not cosmetic.

Each compose volume became a raw `persistentvolumeclaim-*.yaml`. Map each to a userspace area (table below), **delete the PVC template you replace**, and rewrite `volumeMounts` to the injected userspace path. Olares exposes five mountable areas:

| Dir | Mount value | Permission | Files entry | Scope | Backend / traits |
|---|---|---|---|---|---|
| **Home** | `.Values.userspace.userData` | `userData` (list paths) | `drive/Home` | user-level (shared by the user's apps that get the perm) | JuiceFS — cross-node, backed up; for **user-visible** files |
| **Cache** | `.Values.userspace.appCache` | `appCache: true` | `cache/<node>` | per-app (auto `/<appName>`) | **node-local PV** (`/olares/userdata/Cache/`) — pins the pod to that node via `schedule.nodeName`; fast, regenerable, not guaranteed durable/backed-up |
| **Data** | `.Values.userspace.appData` | `appData: true` | `drive/Data` | per-app (auto `/<appName>`) | JuiceFS — cross-node, backed up; for **app-private persistent state** (db files, config) |
| **Common** | `.Values.userspace.appCommon` | `appCommon: true` | `drive/Common` | **cross-app shared** (no `appName` suffix) | JuiceFS; reserved `huggingface`/`ollama`/`llama.cpp`/`comfyui` shared caches; needs Olares ≥ 1.12.6 |
| **External** | `.Values.sharedlib` | `externalData: true` | `external/<node>/<volume>` | user's external storage | SMB/NFS/USB volumes the user attaches via LarePass; needs schema ≥ 0.12.0 |

```yaml
permission:
  appData: true
  appCache: true
  appCommon: true             # shared Common dir; cross-app model/cache sharing
  userData:
  - Home/Documents/MyApp/
```

Key differences to remember when authoring:

- **Per-app vs shared vs user.** `appData`/`appCache` auto-append `/<appName>` (app-private); `appCommon` is bare `/rootfs/Common` (every app with the perm sees the same dir — that's what makes shared model caches work); `userData` is the user's `/Home`.
- **Backend decides scheduling + durability.** `userData`/`appData`/`appCommon` are JuiceFS (cross-node, backed up). `appCache` is a node-local PV, so app-service pins the pod to that node — fast local disk, but treat it as disposable.
- **Owner is uid 1000.** All five are read/written as uid/gid 1000 (`appCommon` is created `chown 1000:1000`). If the main process runs as 1000 it can write any of them directly — see [olares-chart-run-as-user.md](olares-chart-run-as-user.md).
- **Version gates.** `appCommon` needs Olares ≥ 1.12.6; `externalData`/`sharedlib` needs `olaresManifest.version` ≥ 0.12.0.
- **Pick by need.** Private db/config → **Data**; regenerable cache → **Cache**; user-facing files → **Home**; multi-app shared model weights / HF cache → **Common** (see [olares-chart-gpu.md](olares-chart-gpu.md) §B); external disk/network share → **External**.

In the deployment template, replace the PVC mount with the injected host path (`appData`/`appCache` are host paths):

```yaml
      volumes:
      - name: app-data
        hostPath:
          path: {{ .Values.userspace.appData }}/myapp
          type: DirectoryOrCreate
```

> Anything declared in a template (`.Values.userspace.appData/appCache/userData`) MUST have the matching `permission` field, or `lint`'s app-data cross-check fails. Drop leftover kompose PVCs.

> **Coupling with packaging — run identity:** userspace mounts require the process to run as **uid 1000** (set `spec.runAsUser: true`). For third-party or root-default images, or a hardcoded write path Olares won't grant, see [olares-chart-run-as-user.md](olares-chart-run-as-user.md) and [olares-chart-image.md](olares-chart-image.md).

## 3. Middleware & dependencies

> **Same for both release targets** — functional requirement, not cosmetic.

Replace any bundled `postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` workload with Olares **system middleware**, prefer Postgres over a bundled SQLite, and depend on an already-ported companion app instead of copying its workload. `lint` does **not** flag a bundled db, so this is on you. Full rules — the SQLite→Postgres decision, the `middleware:` block, the PostgreSQL extension catalog, `type: application` dependencies, and the self-hosted escape hatch — are in [olares-chart-middleware.md](olares-chart-middleware.md). Env wiring of the `.Values.<mw>.*` values is in [olares-chart-env.md](olares-chart-env.md).

## 4. Entrances & ports

> **Same for both release targets** — functional requirement, not cosmetic.

The stub has one auto-detected entrance. Adjust:

```yaml
entrances:
- name: myapp
  host: myapp-svc        # an existing Service name in templates/
  port: 8080             # the Service port
  title: My App
  authLevel: private     # public | private | internal
  invisible: false       # true for internal-only services
```

- **One entrance per user-facing HTTP service.** Add entries for additional UIs; set `invisible: true` (or omit the entrance) for internal-only services.
- **No web UI at all** (a CLI tool, an MCP server, an API daemon)? Don't force a fake entrance — apply the headless archetype: a web terminal as the visible entrance + the service port as an invisible internal entrance ([olares-chart-archetypes.md](olares-chart-archetypes.md)).
- **Non-HTTP services** (game server, SMTP, RDP, …) are exposed via `ports[]`, not entrances:
  ```yaml
  ports:
  - name: game
    host: game-svc
    port: 7777
    protocol: udp
    exposePort: 47777    # cluster-unique; avoid reserved 22/80/81/443/444/2379/18088
  ```
- **Outbound non-HTTP** (e.g. the app sends SMTP): `options.allowedOutboundPorts: [465, 587]`.

## After refining

```bash
olares-cli chart lint ./myapp        # loop back here on any failure
```

For **local-run**, proceed to [olares-chart-publish-verify.md](olares-chart-publish-verify.md). For **market-distribute**, complete the market-ready checklist in [olares-chart-publish-targets.md](olares-chart-publish-targets.md), then [olares-chart-market-submit.md](olares-chart-market-submit.md).
