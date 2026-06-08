# Version rules — Olares version, apiVersion, and the chart version fields

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This collects the version-related rules that bite when porting: the Olares system version, the porting baseline, the `apiVersion` install-routing axis (v1 per-user vs v3 shared), and the several distinct "version" fields a chart carries.

## Olares system version

The semver scheme (stable / RC / daily), `.Values.sysVersion`, the `-0` prerelease-matching rule, and how to read the target version (`profile list` VERSION column, `--refresh-version`, `settings me version`) are the platform **Olares version & semver model** (loaded via the SKILL.md prerequisite).

## Porting baseline: Olares >= 1.12.6

**This skill (porting apps) targets Olares >= 1.12.6.** This baseline applies only to porting — other `olares-cli` features have no such floor. The reason: the userspace backends a ported app commonly relies on — `drive/Common` (`appCommon`), archive, and NFS — are gated at `1.12.6`. Check the target before porting (`olares-cli profile list` VERSION column; `--refresh-version` or `settings me version` for a live re-fetch).

## apiVersion: the install-routing axis (v1 per-user vs v3 shared)

`OlaresManifest.yaml` carries a top-level `apiVersion`. This is an **install-routing** axis, **not** a manifest-format/schema axis (that is `olaresManifest.version`, below). The toolchain accepts `v1`, `v2`, or `v3` (empty defaults to `v1`); only unknown values are rejected by `lint`. `from-compose` does **not** write an `apiVersion`, so a freshly scaffolded chart is implicitly `v1`.

On Olares >= 1.12.6 the install handler routes purely on `apiVersion`:

| `apiVersion` | Install model | Namespace | Who can install |
|---|---|---|---|
| `v1` (default / empty) | per-user | `<app>-<owner>` | any user |
| `v3` | admin-installed, cluster-wide **shared** | `<app>-shared` | **admin only** (normal-user install is 403'd) |

> **Skill rule: choose `apiVersion` by app shape — do NOT default everything to `v3`.**
>
> - **Per-user app (the common porting case):** leave `apiVersion` at the default (`v1`). The app installs once per user into `<app>-<owner>`. `from-compose` already scaffolds this; nothing to add.
> - **Deliberate shared backend** (heavy/accelerator, its own multi-tenancy, one shared dataset for everyone): set `apiVersion: v3`. This is admin-only and lands in `<app>-shared`. Full pattern: [olares-chart-shared.md](olares-chart-shared.md).

What `v3` additionally turns on (enforced by the toolchain only when `apiVersion: v3`):

- **Declarative env rules** — an app-local `envName` must not start with `OLARES_USER`; user/system variables are mapped via `valueFrom`. (`valueFrom` works the same on v1, it is only the prefix ban that is v3-only.) Full env model: [olares-chart-env.md](olares-chart-env.md).
- **Chart scan** — `lint` rejects templates that inline `OLARES_USER...` env names.

`apiVersion` is independent of `olaresManifest.version` — both `v1` and `v3` work with `0.8.0` and `0.12.0`.

## The version fields in a chart (don't confuse them)

A chart carries several "version" fields with different jobs and different rules:

| Field | Where | What it is | Rule |
|---|---|---|---|
| `apiVersion` | `OlaresManifest.yaml` (top level) | install-routing axis | `v1` per-user (default) vs `v3` admin-only shared; toolchain allows `v1`/`v2`/`v3`, default `v1` |
| `olaresManifest.version` | `OlaresManifest.yaml` | manifest **schema** version | `0.8.0` (legacy) vs `0.12.0` (`--new-schema`); install minimum `>= 0.7.2` |
| `metadata.version` | `OlaresManifest.yaml` | **Chart version** (Market package) | must be semver and **equal `Chart.yaml` `version`** |
| `version` | `Chart.yaml` | Helm chart version | `== metadata.version` |
| `spec.versionName` | `OlaresManifest.yaml` | **upstream app** version (display) | tracks `Chart.yaml` `appVersion` — convention, not enforced |

> **Name clash:** `Chart.yaml` also has its own `apiVersion` (`v2`) — that is the **Helm** chart API and is unrelated to the OlaresManifest `apiVersion`. Don't copy one into the other.

### Schema version: 0.8.0 (legacy) vs 0.12.0 (`--new-schema`)

`olaresManifest.version` declares which manifest schema the chart uses. `from-compose` emits **0.8.0** by default and **0.12.0** with `--new-schema`. The difference is not cosmetic — some fields only exist on 0.12.0:

| | 0.8.0 (legacy, default) | 0.12.0 (`--new-schema`) |
|---|---|---|
| Resource envelope | flat `spec.requiredCpu` / `requiredMemory` / `requiredDisk` / `limitedCpu` / ... | `spec.resources[]` / `spec.accelerator[]` (mode-keyed: `cpu`, `nvidia`, ...) |
| GPU / accelerator | not expressible cleanly | `spec.accelerator` with mode → arch cross-check at `lint` |
| `permission.externalData` (`.Values.sharedlib`) | rejected | supported |

**Use 0.12.0 when** the app declares GPU/accelerator resources, needs `permission.externalData`, or you want the modern resource envelope (recommended for new Market apps). Otherwise 0.8.0 is fine. To switch an existing stub, re-scaffold with `from-compose --new-schema` (or edit `olaresManifest.version` and migrate the resource fields). Accelerator mode declaration and sizing are in [olares-chart-accelerator.md](olares-chart-accelerator.md).

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
- `lint` does **not** pick `apiVersion` for you (it allows `v1`/`v2`/`v3`, default `v1`) and does **not** validate the `type: system` semver constraint — the per-user-vs-shared choice is yours, and the semver constraint is only checked at install time.
- Use the `-0` prerelease suffix in system constraints, or daily/RC builds (which carry a prerelease segment) will fail to match an otherwise-satisfied version.
