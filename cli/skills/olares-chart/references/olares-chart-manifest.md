# Refining OlaresManifest.yaml — the four judgment calls

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a publishable chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. The four areas below are what kompose cannot decide. **§1 Metadata can stay a stub for deploying to your Olares; §2–§4 are functional and always required. Full market-ready metadata is only for publishing — see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).**

## Schema version and apiVersion

`olaresManifest.version` (the manifest **schema**: always `0.12.0` for new apps) and the top-level `apiVersion` (skill sets **`v3`**) are separate axes from the chart/app versions. `0.12.0` carries `spec.accelerator` and `permission.externalData`. The full schema description, the version-field map, and the `type: system` dependency are in [olares-chart-versioning.md](olares-chart-versioning.md); accelerator sizing is in [olares-chart-gpu.md](olares-chart-gpu.md).

## Workloads & replicas (required on 0.12.0)

On `0.12.0` the manifest **must** declare a top-level `workloadReplicas` map — one entry per **Deployment / StatefulSet** the chart renders → its replica count (`lint` rejects a modern non-v2 manifest that omits it). `from-compose` scaffolds this for you; you only touch it when you add or rename workloads.

```yaml
workloadReplicas:
  myapp: 1          # every Deployment/StatefulSet name must appear here
  worker: 1
```

Three rules that bite:

- **Key = rendered workload name.** Each key must equal the rendered `metadata.name` of a Deployment/StatefulSet. **DaemonSets are excluded** (they are one-per-node, not replica-controlled).
- **Template wiring contract.** Every listed workload's `spec.replicas` **must** be `{{ .Values.workloads.<name>.replicaCount }}`, and `values.yaml` must carry a matching `workloads.<name>.replicaCount`. The `.Values.workloads.*` value is documented in [olares-chart-system-values.md](olares-chart-system-values.md).
- **Why it matters (non-obvious).** app-service drives the whole lifecycle through this Helm value: install is two-phase (helm renders at `replicas: 0`, then scales up), suspend scales every listed workload to `0`, resume scales back to the declared counts. If a template **hardcodes** `replicas`, `lint` may still pass but **suspend/resume and the staged install silently stop working** — the value override has nothing to drive. `from-compose` already wires the scaffolded workloads; any workload you add by hand must be wired the same way.

## 1. Metadata

The stub sets `title=name`, the default icon, `categories: [Utilities]`, and no developer info. Fill from the upstream project (or ask the user):

```yaml
metadata:
  name: myapp                 # must match folder + Chart.yaml name; do not change casually
  title: My App               # ≤30 chars
  description: One-line summary shown under the title
  icon: https://.../icon.png  # PNG/WEBP, 256x256, ≤512KB
  version: 0.0.1              # Chart Version — MUST equal Chart.yaml `version`
  categories:                 # see manifest docs; include both 1.11 + 1.12 values for compatibility
  - Utilities
  - Utilities_v112
spec:
  versionName: "1.2.3"        # upstream app version; tracks Chart.yaml `appVersion`
  developer: Upstream Author
  submitter: Your Name
  website: https://project.example
  sourceCode: https://github.com/org/project
  fullDescription: |
    Longer Market description.
```

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
