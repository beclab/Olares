# Shared apps (apiVersion v3, admin-installed, cluster-wide)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first, and [versioning.md](olares-chart-versioning.md) (`apiVersion: v3`, Olares >= 1.12.6).

A **shared app** is installed **once by an admin** and serves the whole Olares cluster: every user reaches the same instance, and other apps consume its services across namespaces. Reach for it when the app:

- uses **accelerator/heavy resources** (a single GPU-backed inference server shared by everyone, not one per user),
- has **its own account system / multi-tenancy**, is meant to be **used by many people**, and needs to **share one data set**.

Typical shape: a local model / inference backend (ollama, vLLM, an LLM gateway). Per-user front-ends are **separate "reference" apps** that depend on the shared backend.

## What `apiVersion: v3` means

In Olares >= 1.12.6 the install handler routes purely on `apiVersion` ([handler_installer_install.go](../../../../framework/app-service/pkg/apiserver/handler_installer_install.go) `switch apiVersion`). `case V3` ⇒:

- **Admin-only install** — non-admin callers are rejected ("only admin users can install v3 / shared apps").
- **Deterministic shared namespace** `<app>-shared` (not the per-user `<app>-<owner>`); single source of truth is `V3AppNamespace` in [pkg/utils/app/app.go](../../../../framework/app-service/pkg/utils/app/app.go).
- **Cluster-wide ApplicationManager** named `<app>-shared-<app>`.
- **`NeedsSharedAccess` is force-set true** so the app is a first-class destination for cross-namespace traffic (service-mesh sidecar + NetworkPolicy).

So for ported apps targeting 1.12.6+, **`apiVersion: v3` is itself the "shared app" declaration** — there is no separate `subCharts shared:true` / `appScope.clusterScoped` wiring to add (that was the legacy v2 form, below).

> **There is no `sharedEntrances` in 1.12.6+.** Earlier shared apps exposed cross-user entrances on a `shared.<zone>` domain; a network redesign removed that. Consumers now reach the shared app's in-cluster Service directly (below). Do **not** add `sharedEntrances` — if you see it in an existing chart (e.g. a stray entry in `ollamav3`), treat it as a leftover and drop it.

## Manifest (annotated, based on `ollamav3`)

```yaml
olaresManifest.version: '0.12.0'      # modern schema (accelerator / externalData)
olaresManifest.type: app
apiVersion: 'v3'                       # => admin-installed, cluster-wide shared app
metadata:
  name: ollamav3
  # ...
spec:
  onlyAdmin: true                      # explicit admin-only (v3 is already admin-gated; set it anyway)
  accelerator:                         # heavy/GPU resource envelope — see gpu.md §C-D
    - mode: nvidia
      requiredMemory: 5Gi
      limitedMemory: 40Gi
      requiredGPUMemory: 1Gi
      limitedGPUMemory: 24Gi
    - mode: strix-halo
      # ...
permission:
  appData: true
  appCache: true
options:
  allowMultipleInstall: true
  conflicts:
    - name: ollamav2                   # conflict with the non-shared / v2 variant
      type: application
  dependencies:
    - name: olares
      version: '>=1.12.6-0'            # REQUIRED floor: shared/v3 lands in 1.12.6
      type: system
entrances:
  - name: terminal                     # normal entrances are for the admin / operator
    host: terminal
    port: 80
    title: Ollama V3
    openMethod: window
  - name: ollama
    host: ollama
    port: 11434
    authLevel: internal
    invisible: true                    # internal service, listed in Settings only
```

Required / expected fields:

| Field | Why |
|---|---|
| `apiVersion: 'v3'` | makes it an admin-installed, cluster-wide shared app |
| `olaresManifest.version: '0.12.0'` | needed for `spec.accelerator` / `permission.externalData` (see [manifest.md](olares-chart-manifest.md)) |
| `options.dependencies` `olares` `>=1.12.6-0` (`type: system`) | **mandatory** — the shared/v3 model only exists on Olares >= 1.12.6 ([versioning.md](olares-chart-versioning.md)) |
| `spec.onlyAdmin: true` | explicit admin-only install (generally set on shared apps) |
| `spec.accelerator` | GPU/accelerator envelope for the heavy backend — sizing in [gpu.md](olares-chart-gpu.md) §C-D |
| `options.conflicts` | avoid co-installing the per-user / older variant of the same backend |
| `middleware` | shared backends can use system middleware normally (e.g. `llmgatewayv3` uses postgres) — see [manifest.md §3](olares-chart-manifest.md) |

## How consumers (reference apps) reach it

A shared app is consumed by **separate reference/client apps**, not by per-user copies of itself:

1. The client declares the dependency: `options.dependencies` with `type: application` on the shared app.
2. At render time app-service injects the shared namespace's Services into the client's Helm values as `.Values.svcs.<svcName>_host` (= `<svcName>.<app>-shared`) and `.Values.svcs.<svcName>_ports` ([helm_utils.go](../../../../framework/app-service/pkg/appinstaller/helm_utils.go)). The client points its config at that host — a plain in-cluster Service DNS, **no entrance/URL involved**.
3. Cross-namespace reachability is automatic: the shared app's `NeedsSharedAccess` (force-true for v3) gets it the mesh sidecar + NetworkPolicy that allow other namespaces to call it.

## Legacy v2 shared form (context only)

Pre-v3 shared apps (still in the catalog: `ytdlp`, `searxngv2`, `vllm*`) expressed sharing inside an `apiVersion: 'v2'` app by marking one sub-chart shared:

```yaml
spec:
  subCharts:
    - name: <app>server   # the heavy/shared workload -> lands in <chart>-shared namespace
      shared: true
    - name: <app>         # the per-user part
options:
  appScope:
    clusterScoped: true
    appRef: [ "<consumer>.*" ]   # which apps may reference it
```

For new ports targeting 1.12.6+ use the `apiVersion: v3` form above instead; this is here only to read existing charts.

## Caveats

- **Only an admin can install a v3 app** — surface this to the user; a normal user install will 403.
- The namespace is **always** `<app>-shared` regardless of what the manifest says; app-service rewrites it.
- Do not add `sharedEntrances` (removed in 1.12.6+).
- A shared app is still subject to the rest of the skill: run-as-user 1000 ([run-as-user.md](olares-chart-run-as-user.md)), pinned image tags, accelerator sizing ([gpu.md](olares-chart-gpu.md)), env rules ([env.md](olares-chart-env.md)).
