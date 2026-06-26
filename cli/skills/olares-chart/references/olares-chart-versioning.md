# Chart version fields — the fixed values every chart writes

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This collects the fixed version fields every chart writes — `apiVersion: v3`, `olaresManifest.version: '0.12.0'`, and the `olares` system-dependency floor — plus the several distinct "version" fields a chart carries and how they relate.

## Olares system version

The semver scheme (stable / RC / daily), `.Values.sysVersion`, the `-0` prerelease-matching rule, and how to read the target version (`profile list` VERSION column, `--refresh-version`, `settings me version`) are the platform **Olares version & semver model** (loaded via the SKILL.md prerequisite).

## Porting baseline: Olares >= 1.12.6

**This skill (porting apps) targets Olares >= 1.12.6.** This baseline applies only to porting — other `olares-cli` features have no such floor. The reason: the userspace backends a ported app commonly relies on — `drive/Common` (`appCommon`), archive, and NFS — are gated at `1.12.6`. Check the target before porting (`olares-cli profile list` VERSION column; `--refresh-version` or `settings me version` for a live re-fetch).

## apiVersion: v3

Every chart sets `apiVersion: v3` at the top of `OlaresManifest.yaml`. `from-compose` does not write it, so hand-add it after scaffolding:

```yaml
apiVersion: v3
olaresManifest.version: '0.12.0'
olaresManifest.type: app
metadata:
  name: myapp
  ...
```

What `apiVersion: v3` governs:

- **Declarative env rules** — an app-local `envName` must not start with `OLARES_USER`; user/system variables are mapped via `valueFrom`. Full env model: [olares-chart-env.md](olares-chart-env.md).
- **Chart scan** — `lint` rejects templates that inline `OLARES_USER...` env names.
- **Admin-only install** — a normal-user install is rejected; only the admin can install the app.
- **Namespace depends on `isShared`** in the manifest:
  - Without `isShared: true` → installs into `<app>-<adminUsername>` (the admin's personal namespace). Use this when the app just needs admin-only install but manages its own users internally.
  - With `isShared: true` → installs into `<app>-shared` (cluster-wide, cross-namespace access enabled). Use this for heavy shared backends (GPU inference servers, shared databases) that other apps consume. See [olares-chart-shared.md](olares-chart-shared.md).

## The version fields in a chart (don't confuse them)

A chart carries several "version" fields with different jobs. Their values are fixed for every app:

| Field | Where | What it is | Value |
|---|---|---|---|
| `apiVersion` | `OlaresManifest.yaml` (top level) | manifest API generation | `v3` |
| `olaresManifest.version` | `OlaresManifest.yaml` | manifest schema | `0.12.0` |
| `metadata.version` | `OlaresManifest.yaml` | **Chart version** (Market package) | semver, **equal `Chart.yaml` `version`** |
| `version` | `Chart.yaml` | Helm chart version | `== metadata.version` |
| `spec.versionName` | `OlaresManifest.yaml` | **upstream app** version (display) | tracks `Chart.yaml` `appVersion` — convention, not enforced |

> **Name clash:** `Chart.yaml` has its own `apiVersion` (Helm's chart API) — unrelated to the OlaresManifest `apiVersion`. Don't copy one into the other.

### olaresManifest.version: 0.12.0

`0.12.0` is the manifest schema every chart uses (`from-compose` emits it). It defines:

- **Resource envelope** under `spec.resources[]` / `spec.accelerator[]` (mode-keyed: `cpu`, `nvidia`, ...).
- **GPU / accelerator** declared via `spec.accelerator` with a mode → arch cross-check at `lint`.
- **`permission.externalData`** (`.Values.sharedlib`).
- **`workloadReplicas`** — a required map of every Deployment/StatefulSet → replica count, with each workload's `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}`. See [olares-chart-manifest.md](olares-chart-manifest.md) (Workloads & replicas).

Accelerator mode declaration and sizing are in [olares-chart-accelerator.md](olares-chart-accelerator.md).

## System dependency: the constraint

Every chart declares the `olares` `type: system` dependency in `options.dependencies` — the entry's shape and the "author it yourself" rule are in [olares-chart-manifest.md](olares-chart-manifest.md) (System dependency: olares). What this section adds is the **constraint semantics**:

- At install, app-service matches the version constraint against the running Olares version (semver); a mismatch blocks install.
- Declare exactly `>=1.12.6-0`. The `-0` prerelease suffix is required so daily/RC builds (e.g. `1.12.6-20260327`) match — without it they fail to match an otherwise-satisfied version.

`lint` requires both this entry and `workloadReplicas` to be present.

## Caveats

- The `>= 1.12.6` baseline is a **porting** concern; it does not apply to other `olares-cli` commands.
- `profile list`'s version is **cached** — use `--refresh-version` (or `settings me version`) if the target was just upgraded.
