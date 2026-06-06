# Version rules — Olares version, apiVersion, and the chart version fields

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This collects the version-related rules that bite when porting: the Olares system version, the porting baseline, the `apiVersion: v3` skill rule, and the several distinct "version" fields a chart carries.

## Olares system version

Olares releases follow [semver](https://docs.olares.com/developer/install/versioning.html) — `Major.Minor.Patch[-PreRelease]`:

| Release type | Example |
|---|---|
| Stable | `1.12.6` |
| Release candidate | `1.12.0-rc.0` |
| Daily build | `1.12.0-20241201` |

The running version lives in the `Terminus` CR `spec.version` and is injected into every chart as `.Values.sysVersion` ([helm_utils.go](../../../../framework/app-service/pkg/appinstaller/helm_utils.go)). "At least" comparisons strip the prerelease/build segment, so a daily build like `1.12.6-20260327` still counts as `>= 1.12.6`.

## Porting baseline: Olares >= 1.12.6

**This skill (porting apps) targets Olares >= 1.12.6.** This baseline applies only to porting — other `olares-cli` features have no such floor. The reason: the userspace backends a ported app commonly relies on — `drive/Common` (`appCommon`), archive, and NFS — are gated at `1.12.6` ([cli/cmd/ctl/files/common_gate.go](../../../../cli/cmd/ctl/files/common_gate.go), `commonNamespaceMinOlaresVersion = "1.12.6"`).

Check the target Olares version before porting:

```bash
olares-cli profile list        # VERSION column shows each profile's Olares version
olares-cli profile list --refresh-version   # re-fetch for the active profile
olares-cli settings me version               # live fetch of the running version
```

`profile list` shows a `VERSION` column — the cached `BackendVersion`, populated at login from `/api/olares-info` `osVersion` (which comes from the Terminus CR). See [`olares-shared`](../../olares-shared/SKILL.md) for profile management.

## apiVersion: v3 (skill rule)

`OlaresManifest.yaml` carries a top-level `apiVersion`. The toolchain accepts `v1`, `v2`, or `v3` (empty defaults to `v1`); only unknown values are rejected by `lint`. `from-compose` does **not** write an `apiVersion`, so a freshly scaffolded chart is implicitly `v1`.

> **Skill rule: set `apiVersion: v3` for every ported app.** Hand-add it after `from-compose` (it is not added or required by `lint` — this is a skill convention, not a CLI lock):
>
> ```yaml
> apiVersion: v3
> olaresManifest.version: '0.8.0'   # independent axis — see below
> olaresManifest.type: app
> metadata:
>   name: myapp
>   ...
> ```

What `v3` turns on (enforced by the toolchain only when `apiVersion: v3`):

- **Declarative env rules** — an app-local `envName` must not start with `OLARES_USER`; user/system variables are mapped via `valueFrom` ([oac/internal/manifest/envs.go](../../../../framework/oac/internal/manifest/envs.go)). Full env model: [olares-chart-env.md](olares-chart-env.md).
- **Chart scan** — `lint` rejects templates that inline `OLARES_USER...` env names.

`apiVersion` is independent of `olaresManifest.version` — `v3` works with both `0.8.0` and `0.12.0`.

## The version fields in a chart (don't confuse them)

A chart carries several "version" fields with different jobs and different rules:

| Field | Where | What it is | Rule |
|---|---|---|---|
| `apiVersion` | `OlaresManifest.yaml` (top level) | manifest API generation | skill sets `v3` (toolchain allows `v1`/`v2`/`v3`, default `v1`) |
| `olaresManifest.version` | `OlaresManifest.yaml` | manifest **schema** version | `0.8.0` (legacy) vs `0.12.0` (`--new-schema`); install minimum `>= 0.7.2` ([config/app.go](../../../../framework/app-service/pkg/utils/config/app.go) `MinCfgFileVersion`) |
| `metadata.version` | `OlaresManifest.yaml` | **Chart version** (Market package) | must be semver and **equal `Chart.yaml` `version`** ([chartfolder/folder.go](../../../../framework/oac/internal/chartfolder/folder.go)) |
| `version` | `Chart.yaml` | Helm chart version | `== metadata.version` |
| `spec.versionName` | `OlaresManifest.yaml` | **upstream app** version (display) | tracks `Chart.yaml` `appVersion` — convention, not enforced |

> **Name clash:** `Chart.yaml` also has its own `apiVersion` (`v2`) — that is the **Helm** chart API and is unrelated to the OlaresManifest `apiVersion`. Don't copy one into the other.

Schema `0.8.0` vs `0.12.0` details (when each is required, e.g. GPU/accelerator and `externalData`) are in [olares-chart-manifest.md](olares-chart-manifest.md).

## Declaring system-version compatibility

State the minimum Olares your app needs via `options.dependencies` with `type: system`:

```yaml
options:
  dependencies:
  - name: olares
    version: ">=1.12.6-0"   # -0 includes daily/prerelease builds (e.g. 1.12.6-20260327)
    type: system
```

At install, app-service matches the constraint against the Terminus version with `MatchVersion` (Masterminds semver, [app_utils.go](../../../../framework/app-service/pkg/utils/app_utils.go)); a mismatch blocks install. `lint` only checks the dependency's structure, not the semver constraint. Given the porting baseline above, declare `>=1.12.6-0` (bump it higher if you use a feature from a later release).

## Caveats

- The `>= 1.12.6` baseline is a **porting** concern; it does not apply to other `olares-cli` commands.
- `profile list`'s version is **cached** — use `--refresh-version` (or `settings me version`) if the target was just upgraded.
- `lint` does **not** enforce `apiVersion: v3` and does **not** validate the `type: system` semver constraint — both are only checked downstream (skill discipline / install time).
- Use the `-0` prerelease suffix in system constraints, or daily/RC builds (which carry a prerelease segment) will fail to match an otherwise-satisfied version.
