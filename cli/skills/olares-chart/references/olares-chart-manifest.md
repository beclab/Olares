# Refining OlaresManifest.yaml — the four refinement areas

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the field-by-field map from a raw `from-compose` stub to a working chart. After every change, re-run `olares-cli chart lint ./<app>` (see [olares-chart-lint.md](olares-chart-lint.md)).

The scaffolded manifest is a stub. The four areas below are what kompose cannot decide. **§1 Metadata can stay a stub for deploying to your Olares; §2–§4 are functional and always required. Full market-ready metadata is only for publishing — see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).**

## Schema version and apiVersion

`olaresManifest.version` (the manifest **schema**: `0.8.0` default vs `0.12.0` via `--new-schema`) and the top-level `apiVersion` (skill sets **`v3`**) are separate axes from the chart/app versions. Use `0.12.0` when the app needs `spec.accelerator` or `permission.externalData`. The full schema-field table, the version-field map, and the `type: system` dependency are in [olares-chart-versioning.md](olares-chart-versioning.md); accelerator sizing is in [olares-chart-gpu.md](olares-chart-gpu.md).

## 1. Metadata

For deploying to your own Olares, metadata can stay a stub as long as `lint` passes. Full market-ready metadata (custom icon, dual-version categories, listing images, marketing spec) is only needed when **publishing to the public Market** — see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

### Always required (`lint` structural check)

```yaml
metadata:
  name: myapp                 # must match folder + Chart.yaml name; do not change casually
  appid: myapp                # app identifier; set = name. from-compose scaffolds it; backs the entrance domain <appid>.<zone>
  title: My App               # stub title=name is OK for local deploy
  description: One-line summary
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp  # default OK for local deploy
  version: 0.0.1              # Chart Version — MUST equal Chart.yaml `version`
  categories:
  - Utilities                 # stub OK for local deploy; lint does not enum-check
spec:
  versionName: "1.2.3"        # upstream app version; tracks Chart.yaml `appVersion`
  runAsUser: true             # optional but recommended — Olares injects pod runAsUser 1000; see run-as-user.md
```

> **`appid` vs `lint`:** `chart lint` only requires `name`, `icon`, `description`, `title`, `version` (`appid` is `omitempty` in the schema, so a chart lints without it). But `from-compose` always writes `appid: <name>`, and the platform uses it as the app's identity (e.g. the entrance host `<appid>.<zone>` — see [olares-chart-system-values.md](olares-chart-system-values.md)). **Keep `appid` present and equal to `metadata.name`** when hand-authoring a manifest (e.g. from a generic Helm chart); rename it alongside `name` / the folder / `Chart.yaml`. **`lint` passes without it; `market upload` does not** — omitting it produces `upload payload missing ... metadata.appid`.

### Keep as stub (deploy to your Olares)

Keep the stub: `Utilities` category, default icon, and empty `spec.developer`/`submitter`/`website`/`sourceCode`/`fullDescription`/`featuredImage`/`promoteImage`/`locale`/`supportArch` are all fine (skip `supportArch` unless using accelerator modes). Optional polish: set `metadata.title`/`description` to something readable and `spec.versionName` to the upstream version.

## 2. Storage (compose volumes → Olares userspace)

> **Functional requirement, not cosmetic** — always required.

Each compose volume became a raw `persistentvolumeclaim-*.yaml`. Map each to a userspace area, **delete the PVC template you replace**, and rewrite `volumeMounts` to the injected userspace path. The five areas — their mount values, `permission` fields, backends (JuiceFS vs node-local PV), scope, uid-1000 ownership, and version gates — are defined once in the platform **Userspace storage model** (loaded via the SKILL.md prerequisite). **Pick by need:** private db/config → **Data** (`appData`); regenerable cache → **Cache** (`appCache`); user-facing files → **Home** (`userData`); multi-app shared model weights / HF cache → **Common** (`appCommon`, see [olares-chart-gpu.md](olares-chart-gpu.md) §B); external disk/network share → **External** (`sharedlib`).

```yaml
permission:
  appData: true
  appCache: true
  appCommon: true             # shared Common dir; cross-app model/cache sharing
  userData:
  - Home/Documents/MyApp/
```

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

> **Functional requirement, not cosmetic** — always required.

Replace any bundled `postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` workload with Olares **system middleware**, prefer Postgres over a bundled SQLite, and depend on an already-ported companion app instead of copying its workload. `lint` does **not** flag a bundled db, so this is on you. Full rules — the SQLite→Postgres decision, the `middleware:` block, the PostgreSQL extension catalog, `type: application` dependencies, and the self-hosted escape hatch — are in [olares-chart-middleware.md](olares-chart-middleware.md). Env wiring of the `.Values.<mw>.*` values is in [olares-chart-env.md](olares-chart-env.md).

## 4. Entrances & ports

> **Functional requirement, not cosmetic** — always required.

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
- **Long-running HTTP requests** (LLM streaming, big uploads, slow report generation): the per-app **entrance proxy** caps every request at `options.apiTimeout` **seconds** — **default 15s**, so anything slower is cut at the entrance (504 / closed connection) regardless of the app or browser. Set `options.apiTimeout: 0` to disable the cap, or a large value (e.g. `3600`) for a bounded one. It is an install-time **manifest** field (not an install-time env), so for `templateOnly` env-driven charts you must edit it in the chart manifest and re-package.

## After refining

```bash
olares-cli chart lint ./myapp        # loop back here on any failure
```

Once `lint` passes, **deploy to your Olares** — [olares-chart-deploy.md](olares-chart-deploy.md). To list it on the public Market afterwards, see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).
