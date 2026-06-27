# chart lint (validate a chart)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli chart lint --help`. This file adds the meaning of each check and a failure→fix table.

`lint` runs the **same ingest pipeline the Olares Market uses to validate chart structure**, against a directory or a `.tgz` / `.tar.gz`. Local-only; no Olares login.

## lint OK ≠ market-ready

| Check | `chart lint` | GitBot PR (`CheckWithTitle`) |
|---|---|---|
| Folder layout, manifest structure, helm dry-run | ✅ | ✅ |
| `metadata.categories` enum | ❌ (stub `Utilities` passes) | ✅ |
| `featuredImage`, `promoteImage`, `fullDescription` | ❌ | recommended for listing |
| Multi-arch / `spec.supportArch` alignment | partial (cross-field if accelerator set) | expected for public Market |

**Deploy to your Olares:** `lint OK` + a live install reaching `running` → sufficient. Stub metadata is fine.

**Publishing to the public Market:** `lint OK` is necessary but not sufficient — complete the market-ready checklist in the [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md) skill before opening a PR.

```bash
olares-cli chart lint ./myapp
olares-cli chart lint ./myapp-1.0.0.tgz
olares-cli chart lint ./myapp --skip-resource --skip-app-data
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
| RBAC rules | a ServiceAccount granted forbidden cluster permissions | (always) |
| securityContext | a non-`beclab/` image running with root-equivalent privileges (`privileged`/`runAsUser: 0`/`runAsNonRoot: false`) | (always) |

> **The RBAC + securityContext checks run unconditionally.** `--with-rbac` / `--with-security-context` exist as flags (and their `--help` text claims "off by default"), but the CLI never actually disables either check — both run on every `chart lint`, so passing the flags is a no-op. Don't rely on them to *enable* anything; treat both checks as always-on.

> **lint does not check middleware usage.** A chart that bundles its own `postgres`/`redis` instead of using system middleware passes `lint` cleanly — removing the bundled db is the author's responsibility (see the Middleware & dependencies area).

## Owner scenarios

By default lint renders the chart under **both** `owner==admin` (admin install) and `owner!=admin` (regular-user install); both must pass. Pin one scenario with `--auto-owner=false --owner <u> --admin <a>` — useful when a chart templates differently for admins (e.g. shared apps using `.Values.bfl.username == .Values.admin`).

## Failure → fix

| Message | Cause | Fix |
|---|---|---|
| `must have a Deployment or StatefulSet named "<app>"` | no workload named after the app | rename the primary workload's `metadata.name` to the app name (`from-compose` does this automatically) |
| app-data / permission mismatch | template mounts `.Values.userspace.*` not declared in `permission` (or reverse) | align `permission.appData/appCache/userData` with template mounts |
| `Chart.yaml` vs manifest version mismatch | the two `version` fields differ | set them equal |
| hostPath + rolling update | a template mounts a `hostPath` with a rolling-update workload | switch to a userspace volume, or set the workload strategy to `Recreate` if a host mount is truly required |
| resource limit missing | a container has no CPU/memory limit | add `resources.limits` (or `--skip-resource` only for a quick check, not for market submit) |
| `workloadReplicas is required` (or a workload not listed) | the `workloadReplicas` map is missing, incomplete, or not wired | run the three-point self-check in the Manifest refinement areas (Workloads & replicas) |
| `options.dependencies must declare ... name="olares" ... type="system"` | the `olares` system dependency is missing from `options.dependencies` | add it per the Manifest refinement areas (System dependency: olares) |
| manifest structural error | required field missing/invalid | fix per the Manifest refinement areas |
| namespace check failed | a resource has a hardcoded namespace | use `namespace: '{{ .Release.Namespace }}'` |

## In the loop

`lint` exit code is the signal: `0` = OK (prints `<path>: OK`), non-zero = a domain error printed without a usage dump. Keep editing the manifest/templates and re-running until it prints OK.

Then:

- **Deploy to your Olares:** the Deploy step (upload + install)
- **Publish to the public Market:** market-ready checklist → re-lint → the [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md) skill
