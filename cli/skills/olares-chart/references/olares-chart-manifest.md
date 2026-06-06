# Refining OlaresManifest.yaml — the four refinement areas

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and [olares-chart-publish-targets.md](olares-chart-publish-targets.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a working chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. The four areas below are what kompose cannot decide. **§1 Metadata depth depends on release target; §2–§4 are functional and required for both targets.**

## 1. Metadata

Depth is gated by release target — see [olares-chart-publish-targets.md](olares-chart-publish-targets.md).

### Always required (`lint` structural check)

These must be present and valid for `chart lint` to pass:

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

- `categories` — `Utilities` alone is fine
- Default icon URL
- `spec.developer`, `submitter`, `website`, `sourceCode`, `fullDescription` — omit or leave empty
- `spec.featuredImage`, `promoteImage`, `locale`, `supportArch` — skip unless using accelerator modes

### market-distribute: required (Market listing + GitBot)

Fill from the upstream project (or ask the user):

```yaml
metadata:
  title: My App               # ≤30 chars, human-readable
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
  locale:
  - en
  supportArch:                # must match image platforms
  - amd64
  - arm64
  featuredImage: https://.../hero.webp
  promoteImage:
  - https://.../screenshot1.webp
  - https://.../screenshot2.webp
```

Category values: [manifest docs — categories](https://docs.olares.com/developer/develop/package/manifest.html#categories). Listing images: [promote-apps](https://docs.olares.com/developer/develop/promote-apps.html).

## 2. Storage (compose volumes → Olares userspace)

> **Same for both release targets** — functional requirement, not cosmetic.

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

> **Coupling with packaging — run identity:** userspace mounts require the process to run as **uid 1000**. Set `spec.runAsUser: true`; for third-party or root-default images also check [olares-chart-run-as-user.md](olares-chart-run-as-user.md) (initContainer `chown` with `beclab/aboveos-busybox:1.37.0`, `securityContext`, or Dockerfile rebuild). If the image hardcodes a write path Olares won't grant, loop back to [olares-chart-image.md](olares-chart-image.md).

## 3. Middleware (use the system service, don't bundle one)

> **Same for both release targets** — functional requirement, not cosmetic.

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
