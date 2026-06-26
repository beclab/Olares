# Refining OlaresManifest.yaml — the four judgment calls

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a publishable chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. This doc covers, in order: the fixed **header fields**, the **two required manifest entries** every chart must author (`workloadReplicas` and the `olares` system dependency), then the **four refinement areas** kompose cannot decide (§1 Metadata can stay a stub for deploying to your Olares; §2–§4 are functional and always required). Full market-ready metadata is only for publishing — see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

## Header fields

The manifest top sets `apiVersion: v3` and `olaresManifest.version: '0.12.0'`. `0.12.0` carries `spec.accelerator` and `permission.externalData`. The fixed version-field values, the `olares` system dependency, and the field map are in [olares-chart-versioning.md](olares-chart-versioning.md); accelerator sizing is in [olares-chart-gpu.md](olares-chart-gpu.md).

> **Two required manifest entries — author both yourself for every chart; `lint` rejects either if missing:** the `workloadReplicas` map (below) and the `olares` system dependency under `options.dependencies` (next section). Both live at the **top level of `OlaresManifest.yaml`** — siblings of `metadata`, `spec`, and `options`. In particular `workloadReplicas` is **not** nested under the manifest `spec`.

## OlaresManifest skeleton (top-level keys)

The top-level shape, in canonical order — every key below is a sibling (each at column 0). Field detail lives in the linked sections; this is just the map.

```yaml
# OlaresManifest.yaml — top-level keys, in order (all siblings)
olaresManifest.version: '0.12.0'
olaresManifest.type: app
apiVersion: v3
metadata: { ... }          # §1
entrances: [ ... ]         # §4
spec: { ... }              # §1 spec + resources/accelerator
permission: { ... }        # §2 — only areas you mount
middleware: { ... }        # §3 — only if wiring system middleware
options: { ... }           # olares system dep (required) + middleware/app deps
workloadReplicas: { ... }  # required; sibling of spec
# optional/niche: ports, envs, overlayGateway, tailscale
```

## Workloads & replicas (required)

Every chart **must** declare a top-level `workloadReplicas` map — one entry per **Deployment / StatefulSet** the chart renders → its replica count. **This is a mandatory part of authoring the chart, and it is on you to get right: declare it and confirm it yourself for every chart, whether you scaffolded with `from-compose` or hand-authored, and re-check it whenever you add, rename, or remove a workload. Do not assume a scaffolder produced it correctly, and do not treat a passing `lint` as proof it is present and wired** — `lint` rejecting an omitted map is a backstop, not the reason you write it.

```yaml
workloadReplicas:
  myapp: 1          # every Deployment/StatefulSet name must appear here
  worker: 1
```

**Self-check (inspect the files directly — do not rely on any tool):**

1. `OlaresManifest.yaml` top-level `workloadReplicas` lists **every** Deployment/StatefulSet the chart renders (by rendered `metadata.name`) → its replica count. **DaemonSets are excluded** (one-per-node, not replica-controlled).
2. Every listed workload's template sets `spec.replicas: {{ .Values.workloads.<name>.replicaCount }}` — **never a hardcoded number**. This `spec.replicas` is the Deployment/StatefulSet **template's** spec under `templates/`, a different field from the manifest's top-level `spec` — don't conflate the two.
3. `values.yaml` carries a matching `workloads.<name>.replicaCount` for each. The `.Values.workloads.*` value is documented in [olares-chart-system-values.md](olares-chart-system-values.md).

**Why the template wiring matters (non-obvious).** app-service drives the whole lifecycle through this Helm value: install is two-phase (helm renders at `replicas: 0`, then scales up), suspend scales every listed workload to `0`, resume scales back to the declared counts. If a template **hardcodes** `replicas`, **suspend/resume and the staged install silently stop working** — the value override has nothing to drive.

## System dependency: olares (required)

Every chart **must** declare the `olares` system dependency in `options.dependencies` — author it yourself, under the same "don't assume the scaffolder added it, don't treat a passing `lint` as proof" rule as `workloadReplicas` above.

```yaml
options:
  dependencies:
  - name: olares
    version: '>=1.12.6-0'   # -0 covers daily/prerelease builds
    type: system
```

The constraint and the `-0` prerelease rule are in [olares-chart-versioning.md](olares-chart-versioning.md).

## 1. Metadata

The stub sets `title=name`, the default icon, `categories: [Utilities]`, and no developer info. Fill from the upstream project (or ask the user):

```yaml
metadata:
  name: myapp                 # must match folder + Chart.yaml name; do not change casually
  appid: myapp                # app identifier; set = name. from-compose scaffolds it; backs the entrance domain <appid>.<zone>
  title: My App               # stub title=name is OK for local deploy
  description: One-line summary
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp  # default OK for local deploy
  version: 0.0.1              # Chart Version — MUST equal Chart.yaml `version`; bump on each (re)upload
  categories:
  - Utilities                 # stub OK for local deploy; lint does not enum-check
spec:
  versionName: "1.2.3"        # upstream app version; tracks Chart.yaml `appVersion`
  developer: Upstream Author
  submitter: Your Name
  website: https://project.example
  sourceCode: https://github.com/org/project
  fullDescription: |
    Longer Market description.
```

> **Resource envelope (optional, under `spec`):** a non-accelerator app sets flat `spec.requiredCpu` / `limitedCpu` / `requiredMemory` / `limitedMemory` / `requiredDisk` (no `mode`); a GPU/accelerator app uses `spec.accelerator[]` instead — the two are mutually exclusive. See [olares-chart-accelerator.md](olares-chart-accelerator.md) §A.1.

> **`appid` vs `lint`:** `chart lint` only requires `name`, `icon`, `description`, `title`, `version` (`appid` is `omitempty` in the schema, so a chart lints without it). But `from-compose` always writes `appid: <name>`, and the platform uses it as the app's identity (e.g. the entrance host `<appid>.<zone>` — see [olares-chart-system-values.md](olares-chart-system-values.md)). **Keep `appid` present and equal to `metadata.name`** when hand-authoring a manifest (e.g. from a generic Helm chart); rename it alongside `name` / the folder / `Chart.yaml`. **`lint` passes without it; `market upload` does not** — omitting it produces `upload payload missing ... metadata.appid`.

### Keep as stub (deploy to your Olares)

Keep the stub: `Utilities` category, default icon, and empty `spec.developer`/`submitter`/`website`/`sourceCode`/`fullDescription`/`featuredImage`/`promoteImage`/`locale`/`supportArch` are all fine (skip `supportArch` unless using accelerator modes). Optional polish: set `metadata.title`/`description` to something readable and `spec.versionName` to the upstream version.

## 2. Storage (compose volumes → Olares userspace)

Each compose volume became a raw `persistentvolumeclaim-*.yaml`. Decide per volume, then **delete the PVC template you replace** and rewrite the container's `volumeMounts` to an `emptyDir`/`hostPath` pointing at the injected userspace path.

| Volume holds | Mount it on | Declare in `permission` |
|---|---|---|
| App-private state (config, db files you keep self-hosted) | `.Values.userspace.appData` | `appData: true` |
| Regenerable cache | `.Values.userspace.appCache` | `appCache: true` |
| Files the user should see in Files app | `.Values.userspace.userData` + subpath | add the path under `userData:` |

```yaml
permission:
  appData: true
  appCache: true
  userData:
  - Home/Documents/MyApp/
```

In the deployment template, replace the PVC mount with the injected path:

```yaml
        volumeMounts:
        - name: app-data
          mountPath: /var/lib/myapp
      volumes:
      - name: app-data
        hostPath:
          path: {{ .Values.userspace.appData }}/myapp   # appData/appCache are host paths injected by Olares
          type: DirectoryOrCreate
```

> Anything declared in a template (`.Values.userspace.appData/appCache/userData`) MUST have the matching `permission` field, or `lint`'s app-data cross-check fails. Drop leftover kompose PVCs.

> **Coupling with packaging:** storage and permission are constrained by how the **image** was built. If the image hardcodes a write path Olares won't grant, or runs as root where Olares expects non-root, the fix may be to **rebuild the image** (back to the Image capability in [olares-chart-image.md](olares-chart-image.md)) so it writes under an injected userspace path and runs as a normal user — not just to edit this manifest.

## 3. Middleware (use the system service, don't bundle one)

A compose `postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` service should usually be removed and replaced by Olares-managed middleware. For each:

1. Delete that service's `deployment-*.yaml` (or statefulset) and its PVC.
2. Add a `middleware:` block.
3. Add an `options.dependencies` entry of type `middleware` (set `mandatory: true` if install must wait for it).
4. Repoint the app's env vars at the injected `.Values.<mw>.*`.

```yaml
middleware:
  postgres:
    username: myapp
    databases:
    - name: myapp            # → reference as .Values.postgres.databases.myapp
  redis:
    namespace: db0
options:
  dependencies:
  - name: olares
    version: ">=1.0.0-0"
    type: system
```

Env wiring in the deployment (PostgreSQL example; Redis/Mongo/MySQL/MariaDB/MinIO/RabbitMQ are analogous):

```yaml
        env:
        - name: DB_HOST
          value: "{{ .Values.postgres.host }}"
        - name: DB_PORT
          value: "{{ .Values.postgres.port }}"
        - name: DB_USER
          value: "{{ .Values.postgres.username }}"
        - name: DB_PASSWORD
          value: "{{ .Values.postgres.password }}"
        - name: DB_NAME
          value: "{{ .Values.postgres.databases.myapp }}"
```

> MongoDB, MySQL, MariaDB, MinIO, RabbitMQ must be installed by an admin from the Market before client apps can use them; PostgreSQL/Redis are always available. Keep a self-hosted db only if the app needs a version/extension the system middleware can't provide.

> The `olares` `type: system` dependency (see "System dependency: olares") is a **separate, always-required** entry in `options.dependencies` — keep it when you add or remove middleware / application dependencies.

## 4. Entrances & ports

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
