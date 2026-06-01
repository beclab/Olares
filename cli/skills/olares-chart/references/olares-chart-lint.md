# chart lint (validate a chart before publishing)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli chart lint --help`. This file adds the meaning of each check and a failure→fix table.

`lint` runs the **same pipeline the Olares Market uses to ingest a chart**, against a directory or a `.tgz` / `.tar.gz`. Local-only; no Olares login.

```bash
olares-cli chart lint ./myapp
olares-cli chart lint ./myapp-1.0.0.tgz
olares-cli chart lint ./myapp --skip-resource --with-rbac
olares-cli chart lint ./myapp --auto-owner=false --owner alice --admin root
```

## What it checks

| Stage | Catches | Skip flag |
|---|---|---|
| Folder layout | missing `Chart.yaml` / `values.yaml` / `templates/` / `OlaresManifest.yaml` | `--skip-folder` |
| Manifest validation | structural + cross-field errors in `OlaresManifest.yaml` | `--skip-manifest` |
| Helm dry-run + workload integrity | templates don't render, or no `Deployment`/`StatefulSet` named after the app | (always) |
| Resource limits | containers missing CPU/memory limits | `--skip-resource` |
| hostPath check | `hostPath` mount + rolling update (incompatible) | `--skip-host-path` |
| Namespace check | rendered resource pinned to a non-templated namespace | `--skip-namespace` |
| App-data cross-check | `.Values.userspace.appData/appCache/userData` used in a template but not declared in `permission` (or vice versa) | `--skip-app-data` |
| Version match | `Chart.yaml` `version` ≠ `metadata.version` | `--skip-same-version` |

Off by default (opt in): `--with-rbac` (ServiceAccount forbidden-rule check), `--with-security-context` (non-beclab privileged securityContext check).

## Owner scenarios

By default lint renders the chart under **both** `owner==admin` (admin install) and `owner!=admin` (regular-user install); both must pass. Pin one scenario with `--auto-owner=false --owner <u> --admin <a>` — useful when a chart templates differently for admins (e.g. shared apps using `.Values.bfl.username == .Values.admin`).

## Failure → fix

| Message | Cause | Fix |
|---|---|---|
| `must have a Deployment or StatefulSet named "<app>"` | no workload named after the app | rename the primary workload's `metadata.name` to the app name (`from-compose` does this automatically) |
| app-data / permission mismatch | template mounts `.Values.userspace.*` not declared in `permission` (or reverse) | align `permission.appData/appCache/userData` with template mounts |
| `Chart.yaml` vs manifest version mismatch | the two `version` fields differ | set them equal |
| hostPath + rolling update | a template mounts a `hostPath` with a rolling-update workload | switch to a userspace volume, or set the workload strategy to `Recreate` if a host mount is truly required |
| resource limit missing | a container has no CPU/memory limit | add `resources.limits` (or `--skip-resource` only for a quick check, not for publish) |
| manifest structural error | required field missing/invalid | fix per [olares-chart-manifest.md](olares-chart-manifest.md) |
| namespace check failed | a resource has a hardcoded namespace | use `namespace: '{{ .Release.Namespace }}'` |

## In the loop

`lint` exit code is the signal: `0` = OK (prints `<path>: OK`), non-zero = a domain error printed without a usage dump. Keep editing the manifest/templates and re-running until it prints OK, then optionally live-validate on a real Olares — see [olares-chart-publish-verify.md](olares-chart-publish-verify.md).
