# Middleware & dependencies (refinement area 3)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This is refinement area 3 of the Manifest — the database/queue and companion-app half. Env wiring of the `.Values.<mw>.*` values below belongs to the Env area.

> **Functional requirement, not cosmetic** — always required.

## Use the system service, don't bundle one

**Default rule: replace bundled databases with system middleware.** A compose `postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` service MUST be removed and wired to Olares-managed middleware. A self-hosted db is the **exception**, not the default — it needs an explicit, named reason (see the escape hatch below).

> **Anti-pattern:** do NOT keep the `postgres`/`redis` workload that `from-compose` scaffolded just because it renders and passes `lint`. `lint` does not check for this — a bundled db lints fine but wastes resources, skips backups/HA, and is the most common porting mistake. Removing it is part of the job, not optional.

## SQLite default → prefer PostgreSQL

Many upstream demos / compose files default to **SQLite** for zero-config convenience. On Olares the data file lands on a userspace `hostPath` (potentially networked/shared storage), where SQLite's file locking and concurrent writes are **prone to corruption**. A demo defaulting to SQLite does not mean production should.

**When porting, check whether the upstream supports an external database and switch to system PostgreSQL if it does:**

1. Inspect DB config knobs: compose / `.env` vars like `DATABASE_URL`, `DB_TYPE` / `DB_CONNECTION` / `DB_ENGINE`, `*_DB_HOST`; the upstream docs / sample config for a postgres section; whether the ORM is multi-driver (Django, Rails, Prisma, Sequelize, SQLAlchemy commonly are).
2. **Supports Postgres → wire it to system middleware** (below). Typically just set the connection env, e.g. `DATABASE_URL=postgres://{{ .Values.postgres.username }}:{{ .Values.postgres.password }}@{{ .Values.postgres.host }}:{{ .Values.postgres.port }}/{{ .Values.postgres.databases.myapp }}`, and drop the SQLite volume.
3. **Only supports SQLite (cannot switch) → fallback:** keep SQLite but put the db file under `.Values.userspace.appData` (never `userData`), run a single replica with `strategy: Recreate` (no concurrent writers), and note the corruption risk to the user. Record this as an exception ("upstream has no external-db support").

> Rule of thumb: **if it can be configured for Postgres, configure it for Postgres.** Keep SQLite only when the upstream genuinely has no external-database option.

## Wiring system middleware

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
  - name: olares             # required system dep — author per the Manifest refinement areas (System dependency: olares)
    version: ">=1.12.6-0"
    type: system
  - name: mysql              # middleware deps: set mandatory when install must wait for it
    version: ">=8.0.0-0"
    type: middleware
    mandatory: true
```

> Use a **single-instance** postgres database — do not declare `distributed`. (Distributed citus exists but is out of scope for ported apps.)
> `scripts` run after the database is created; `$dbusername` and `$databasename` are substituted with the real values. `extensions` are created with `CREATE EXTENSION IF NOT EXISTS <name> CASCADE`.

## PostgreSQL extension catalog

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

> **Env wiring is its own topic.** The `.Values.<mw>.*` mappings above, plus app config the user supplies at install (admin credentials), reused `OLARES_SYSTEM_*`/`OLARES_USER_*` vars, and the `envs[]` `required`/`type`/`regex` rules, are all covered in the Env area.

> **Which middleware is always available vs admin-installed** is the platform **System middleware model** (loaded via the SKILL.md prerequisite): PostgreSQL + Redis are always on; MongoDB/MySQL/MariaDB/MinIO/RabbitMQ/NATS need an admin install first. Given the extension catalog above, the system postgres covers nearly every app, so reach for it by default.
>
> **Escape hatch (rare):** keep a self-hosted db ONLY when the app needs a specific version or extension that the system middleware genuinely cannot provide — and only after checking the extension catalog above. State the exact missing version/extension when you do; "it's simpler" is not a valid reason.

## Depend on an already-ported app (don't bundle it)

`middleware` covers Olares-managed databases/queues. The other case is a **companion application** the upstream bundles — a meta-search backend like **searxng**, an embeddings/inference service, an auth provider — that Olares **already ships as a Market app**. Don't copy that app's workload into your chart; declare a dependency on the existing one.

1. **Find the exact name + a usable version.** Search the Market with the [`olares-market`](../../olares-market/SKILL.md) skill (`market list` to browse, `market get <app>` for detail), or read the app's `OlaresManifest.yaml` in [beclab/apps](https://github.com/beclab/apps) and take `metadata.name` / `metadata.version`.
2. **Declare it** under `options.dependencies` with `type: application`:
   ```yaml
   options:
     dependencies:
     - name: olares            # required system dep — author per the Manifest refinement areas (System dependency: olares)
       version: ">=1.12.6-0"
       type: system
     - name: searxng           # exact Market app name
       version: ">=1.0.0-0"    # semver constraint matched against the installed app
       type: application
       mandatory: true         # install is blocked until this app is present
   ```
3. **`mandatory` semantics.** When `mandatory: true` and the app is not installed, app-service refuses install with `dependency application <name> not existed`; an installed-but-version-mismatched app errors too. Leave it `false` (or omit) for a soft/optional dependency. `selfRely: true` makes the chart satisfy the dependency itself (skip the check) — rarely needed for ported apps.

> **Reaching the dependency:** it runs as its own app in its own namespace; your app talks to it over the endpoint that app exposes, not an in-chart Service. The exact env/URL wiring is app-specific — copy it from that app's official chart in [beclab/apps](https://github.com/beclab/apps) rather than guessing a cluster DNS name.

> **If the dependency is a v3 shared backend** (a cluster-wide ollama / vLLM / LLM gateway), app-service injects its Services into your chart as `.Values.svcs.<svc>_host` / `.Values.svcs.<svc>_ports`, so you reach it by cross-namespace Service DNS — no entrance/URL. Authoring the shared backend itself (admin-only, `<app>-shared` namespace) is covered in the Shared backend pattern.

> **middleware vs application:** `type: middleware` = an Olares-managed datastore wired via `.Values.<mw>.*`; `type: application` = a separate, full Olares app you depend on. Use middleware for databases/queues, application for companion apps.
