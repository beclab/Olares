# Version rules — Olares version, apiVersion, and the chart version fields

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This collects the version-related rules that bite when porting: the Olares system version, the porting baseline, the `apiVersion: v3` skill rule, and the several distinct "version" fields a chart carries.

## Olares system version

The semver scheme (stable / RC / daily), `.Values.sysVersion`, the `-0` prerelease-matching rule, and how to read the target version (`profile list` VERSION column, `--refresh-version`, `settings me version`) are the platform **Olares version & semver model** (loaded via the SKILL.md prerequisite).

## Porting baseline: Olares >= 1.12.6

**This skill (porting apps) targets Olares >= 1.12.6.** This baseline applies only to porting — other `olares-cli` features have no such floor. The reason: the userspace backends a ported app commonly relies on — `drive/Common` (`appCommon`), archive, and NFS — are gated at `1.12.6`. Check the target before porting (`olares-cli profile list` VERSION column; `--refresh-version` or `settings me version` for a live re-fetch).

## apiVersion: v3 (skill rule)

`OlaresManifest.yaml` carries a top-level `apiVersion`. The toolchain accepts `v1`, `v2`, or `v3` (empty defaults to `v1`); only unknown values are rejected by `lint`. `from-compose` does **not** write an `apiVersion`, so a freshly scaffolded chart is implicitly `v1`.

> **Skill rule: set `apiVersion: v3` for every ported app.** Hand-add it after `from-compose` (it is not added or required by `lint` — this is a skill convention, not a CLI lock):
>
> ```yaml
> apiVersion: v3
> olaresManifest.version: '0.12.0'   # independent axis — see below
> olaresManifest.type: app
> metadata:
>   name: myapp
>   ...
> ```

What `v3` turns on (enforced by the toolchain only when `apiVersion: v3`):

- **Declarative env rules** — an app-local `envName` must not start with `OLARES_USER`; user/system variables are mapped via `valueFrom`. Full env model: [olares-chart-env.md](olares-chart-env.md).
- **Chart scan** — `lint` rejects templates that inline `OLARES_USER...` env names.
- **Admin-only install** — a normal-user install is rejected; only the admin can install a v3 app.
- **Namespace depends on `isShared`** in the manifest:
  - Without `isShared: true` → installs into `<app>-<adminUsername>` (the admin's personal namespace). Use this when the app just needs admin-only install but manages its own users internally.
  - With `isShared: true` → installs into `<app>-shared` (cluster-wide, cross-namespace access enabled). Use this for heavy shared backends (GPU inference servers, shared databases) that other apps consume. See [olares-chart-shared.md](olares-chart-shared.md).

`apiVersion` is independent of `olaresManifest.version` — `v3` is the skill rule for the API axis, while the schema axis is always `0.12.0` (see below).

## The version fields in a chart (don't confuse them)

A chart carries several "version" fields with different jobs and different rules:

| Field | Where | What it is | Rule |
|---|---|---|---|
| `apiVersion` | `OlaresManifest.yaml` (top level) | manifest API generation | skill sets `v3` (toolchain allows `v1`/`v2`/`v3`, default `v1`) |
| `olaresManifest.version` | `OlaresManifest.yaml` | manifest **schema** version | always `0.12.0` for new apps (`from-compose` emits it; install minimum `>= 0.7.2`, legacy `< 0.12.0` charts still install) |
| `metadata.version` | `OlaresManifest.yaml` | **Chart version** (Market package) | must be semver and **equal `Chart.yaml` `version`** |
| `version` | `Chart.yaml` | Helm chart version | `== metadata.version` |
| `spec.versionName` | `OlaresManifest.yaml` | **upstream app** version (display) | tracks `Chart.yaml` `appVersion` — convention, not enforced |

> **Name clash:** `Chart.yaml` also has its own `apiVersion` (`v2`) — that is the **Helm** chart API and is unrelated to the OlaresManifest `apiVersion`. Don't copy one into the other.

### Schema version: 0.12.0

`olaresManifest.version` declares which manifest schema the chart uses. **New apps always use `0.12.0`** — `from-compose` emits it unconditionally (the old `--new-schema` flag is a deprecated no-op). The 0.12.0 schema:

- **Resource envelope** lives under `spec.resources[]` / `spec.accelerator[]` (mode-keyed: `cpu`, `nvidia`, ...), not the legacy flat `spec.requiredCpu` / `requiredMemory` / ... fields.
- **GPU / accelerator** is declared via `spec.accelerator` with a mode → arch cross-check at `lint`.
- **`permission.externalData`** (`.Values.sharedlib`) is supported.
- **`workloadReplicas` is required** (non-v2): a map of every Deployment/StatefulSet → replica count, with each workload's `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}`. See [olares-chart-manifest.md](olares-chart-manifest.md) (Workloads & replicas).

Accelerator mode declaration and sizing are in [olares-chart-accelerator.md](olares-chart-accelerator.md).

> **Legacy `< 0.12.0` (e.g. `0.8.0`) charts** still install and validate (the install minimum is `>= 0.7.2`), and the validator reads their flat `spec.requiredCpu` / ... fields. You only meet them when reading old charts; do not author new ones on the legacy schema.

## Declaring system-version compatibility

State the minimum Olares your app needs via `options.dependencies` with `type: system`:

```yaml
options:
  dependencies:
  - name: olares
    version: ">=1.12.6-0"   # -0 includes daily/prerelease builds (e.g. 1.12.6-20260327)
    type: system
```

At install, app-service matches the constraint against the running Terminus version (semver); a mismatch blocks install. `lint` only checks the dependency's structure, not the semver constraint. Given the porting baseline above, declare `>=1.12.6-0` (bump it higher if you use a feature from a later release).

## Caveats

- The `>= 1.12.6` baseline is a **porting** concern; it does not apply to other `olares-cli` commands.
- `profile list`'s version is **cached** — use `--refresh-version` (or `settings me version`) if the target was just upgraded.
- `lint` does **not** enforce `apiVersion: v3` and does **not** validate the `type: system` semver constraint — both are only checked downstream (skill discipline / install time).
- Use the `-0` prerelease suffix in system constraints, or daily/RC builds (which carry a prerelease segment) will fail to match an otherwise-satisfied version.
