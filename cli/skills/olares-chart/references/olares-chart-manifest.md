# Refining OlaresManifest.yaml — the four refinement areas

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and [olares-chart-publish-targets.md](olares-chart-publish-targets.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a working chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. The four areas below are what kompose cannot decide. **§1 Metadata depth depends on release target; §2–§4 are functional and required for both targets.**

## Schema version: 0.8.0 (legacy) vs 0.12.0 (`--new-schema`)

`olaresManifest.version` declares which manifest schema the chart uses. `from-compose` emits **0.8.0** by default and **0.12.0** with `--new-schema`. The difference is not cosmetic — some fields only exist on 0.12.0:

| | 0.8.0 (legacy, default) | 0.12.0 (`--new-schema`) |
|---|---|---|
| Resource envelope | flat `spec.requiredCpu` / `requiredMemory` / `requiredDisk` / `limitedCpu` / ... | `spec.resources[]` / `spec.accelerator[]` (mode-keyed: `cpu`, `nvidia`, ...) |
| GPU / accelerator | not expressible cleanly | `spec.accelerator` with mode → arch cross-check at `lint` |
| `permission.externalData` (`.Values.sharedlib`) | rejected | supported |

**Use 0.12.0 when** the app declares GPU/accelerator resources, needs `permission.externalData`, or you want the modern resource envelope (recommended for new Market apps). Otherwise 0.8.0 is fine. To switch an existing stub, re-scaffold with `from-compose --new-schema` (or edit `olaresManifest.version` and migrate the resource fields). Declaring `spec.accelerator` modes (nvidia/amd/apple-m/cpu, mode→arch cross-check, GPU-memory sizing, and how much to request) is in [olares-chart-gpu.md](olares-chart-gpu.md) §C–D; the per-target view is in [olares-chart-publish-targets.md](olares-chart-publish-targets.md).

> **`olaresManifest.version` is not `apiVersion`.** The top-level `apiVersion` is a separate axis — this skill sets it to **`v3`** (hand-add it; `from-compose` omits it). v3 works with either schema version and enables the declarative env rules. Both, plus `metadata.version` vs `spec.versionName` and the `type: system` dependency, are covered in [olares-chart-versioning.md](olares-chart-versioning.md).

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

Each compose volume became a raw `persistentvolumeclaim-*.yaml`. Decide which userspace area each one maps to (table below), then **delete the PVC template you replace** and rewrite the container's `volumeMounts` to an `emptyDir`/`hostPath` pointing at the injected userspace path. Olares exposes five mountable areas:

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

**Default rule: replace bundled databases with system middleware.** A compose `postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` service MUST be removed and wired to Olares-managed middleware. A self-hosted db is the **exception**, not the default — it needs an explicit, named reason (see the escape hatch below).

> **Anti-pattern:** do NOT keep the `postgres`/`redis` workload that `from-compose` scaffolded just because it renders and passes `lint`. `lint` does not check for this — a bundled db lints fine but wastes resources, skips backups/HA, and is the most common porting mistake. Removing it is part of the job, not optional.

### SQLite default → prefer PostgreSQL

Many upstream demos / compose files default to **SQLite** for zero-config convenience. On Olares the data file lands on a userspace `hostPath` (potentially networked/shared storage), where SQLite's file locking and concurrent writes are **prone to corruption**. A demo defaulting to SQLite does not mean production should.

**When porting, check whether the upstream supports an external database and switch to system PostgreSQL if it does:**

1. Inspect DB config knobs: compose / `.env` vars like `DATABASE_URL`, `DB_TYPE` / `DB_CONNECTION` / `DB_ENGINE`, `*_DB_HOST`; the upstream docs / sample config for a postgres section; whether the ORM is multi-driver (Django, Rails, Prisma, Sequelize, SQLAlchemy commonly are).
2. **Supports Postgres → wire it to system middleware** (§3 above). Typically just set the connection env, e.g. `DATABASE_URL=postgres://{{ .Values.postgres.username }}:{{ .Values.postgres.password }}@{{ .Values.postgres.host }}:{{ .Values.postgres.port }}/{{ .Values.postgres.databases.myapp }}`, and drop the SQLite volume.
3. **Only supports SQLite (cannot switch) → fallback:** keep SQLite but put the db file under `.Values.userspace.appData` (never `userData`), run a single replica with `strategy: Recreate` (no concurrent writers), and note the corruption risk to the user. Record this as an exception ("upstream has no external-db support").

> Rule of thumb: **if it can be configured for Postgres, configure it for Postgres.** Keep SQLite only when the upstream genuinely has no external-database option.

For each db/queue service:

1. Delete that service's `deployment-*.yaml` (or statefulset) and its PVC.
2. Add a `middleware:` block.
3. Add an `options.dependencies` entry of type `middleware` (set `mandatory: true` if install must wait for it).
4. Repoint the app's env vars at the injected `.Values.<mw>.*`.

```yaml
middleware:
  postgres:
    username: myapp
    password: myapp
    databases:
    - name: myapp            # → reference as .Values.postgres.databases.myapp
      extensions:            # optional — only what the upstream app needs (see catalog below)
      - vector
      scripts:               # optional — run after the db is created, in order
      - BEGIN;
      - ALTER EXTENSION vector OWNER TO $dbusername;
      - COMMIT;
  redis:
    namespace: db0           # logical Redis db namespace for this app
options:
  dependencies:
  - name: olares
    version: ">=1.0.0-0"
    type: system
  - name: mysql              # middleware deps: set mandatory when install must wait for it
    version: ">=8.0.0-0"
    type: middleware
    mandatory: true
```

> Use a **single-instance** postgres database — do not declare `distributed`. (Distributed citus exists but is out of scope for ported apps.)
> `scripts` run after the database is created; `$dbusername` and `$databasename` are substituted with the real values. `extensions` are created with `CREATE EXTENSION IF NOT EXISTS <name> CASCADE`.

### PostgreSQL extension catalog

The system postgres (PostgreSQL 17, Citus image) ships these extra extensions — declare the name in `extensions:`. **Only add what the upstream app requires.**

| Declare as | Created extension | Purpose |
|---|---|---|
| `vector` / `pgvector` | `vector` | pgvector vector search (HNSW/IVFFlat) |
| `vectors` / `pgvecto.rs` | `vectors` | pgvecto.rs vector search (preloaded `vectors.so`) |
| `vchord` | `vchord` | VectorChord vector search (preloaded `vchord.so`) |
| `hll` | `hll` | HyperLogLog cardinality estimation |
| `topn` | `topn` | Top-N approximation |
| `postgis` | `postgis` (+ `postgis_topology`, ...) | Geospatial |
| `zhparser` | `zhparser` | Chinese full-text search parser |

- Standard PostgreSQL 17 **contrib** extensions also work (e.g. `earthdistance` + `cube`, `pg_trgm`, `uuid-ossp`, `hstore`, `pgcrypto`, `btree_gin`, `btree_gist`) — declared the same way, dependencies pulled in via `CASCADE`.
- Three vector engines coexist (`vector` / `vectors` / `vchord`) — pick the one the upstream expects; **do not mix them**.

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

> **Env wiring is its own topic.** The `.Values.<mw>.*` mappings above, plus app config the user supplies at install (admin credentials), reused `OLARES_SYSTEM_*`/`OLARES_USER_*` vars, and the `envs[]` `required`/`type`/`regex` rules, are all covered in [olares-chart-env.md](olares-chart-env.md).

> **PostgreSQL and Redis are always available — no admin pre-install, no extra steps.** Given the extension catalog above (vector/vectors/vchord/postgis/zhparser/... + standard contrib), the system postgres covers nearly every app, so reach for it by default. MongoDB, MySQL, MariaDB, MinIO, RabbitMQ require an admin to install them from the Market first.
>
> **Escape hatch (rare):** keep a self-hosted db ONLY when the app needs a specific version or extension that the system middleware genuinely cannot provide — and only after checking the extension catalog above. State the exact missing version/extension when you do; "it's simpler" is not a valid reason.

### Depend on an already-ported app (don't bundle it)

`middleware` covers Olares-managed databases/queues. The other case is a **companion application** the upstream bundles — a meta-search backend like **searxng**, an embeddings/inference service, an auth provider — that Olares **already ships as a Market app**. Don't copy that app's workload into your chart; declare a dependency on the existing one.

1. **Find the exact name + a usable version.** Search the Market with the [`olares-market`](../olares-market/SKILL.md) skill (`market list` to browse, `market get <app>` for detail), or read the app's `OlaresManifest.yaml` in [beclab/apps](https://github.com/beclab/apps) and take `metadata.name` / `metadata.version`.
2. **Declare it** under `options.dependencies` with `type: application`:
   ```yaml
   options:
     dependencies:
     - name: olares
       version: ">=1.0.0-0"
       type: system
     - name: searxng           # exact Market app name
       version: ">=1.0.0-0"    # semver constraint matched against the installed app
       type: application
       mandatory: true         # install is blocked until this app is present
   ```
3. **`mandatory` semantics.** When `mandatory: true` and the app is not installed, app-service refuses install with `dependency application <name> not existed`; an installed-but-version-mismatched app errors too. Leave it `false` (or omit) for a soft/optional dependency. `selfRely: true` makes the chart satisfy the dependency itself (skip the check) — rarely needed for ported apps.

> **Reaching the dependency:** it runs as its own app in its own namespace; your app talks to it over the endpoint that app exposes, not an in-chart Service. The exact env/URL wiring is app-specific — copy it from that app's official chart in [beclab/apps](https://github.com/beclab/apps) rather than guessing a cluster DNS name.

> **middleware vs application:** `type: middleware` = an Olares-managed datastore wired via `.Values.<mw>.*` (§3 above); `type: application` = a separate, full Olares app you depend on. Use middleware for databases/queues, application for companion apps.

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
