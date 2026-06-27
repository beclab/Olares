# Shared apps (apiVersion v3, admin-installed, cluster-wide)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first, and [versioning.md](olares-chart-versioning.md) (`apiVersion: v3`, Olares >= 1.12.6).

A **shared app** is installed **once by an admin** and serves the whole Olares cluster: every user reaches the same instance, and other apps consume its services across namespaces. Reach for it when the app:

- uses **accelerator/heavy resources** (a single GPU-backed inference server shared by everyone, not one per user),
- has **its own account system / multi-tenancy**, is meant to be **used by many people**, and needs to **share one data set**.

Typical shape: a local model / inference backend (ollama, vLLM, an LLM gateway). Per-user front-ends are **separate "reference" apps** that depend on the shared backend.

## What makes an app shared

`apiVersion: v3` by itself only selects the manifest schema (env rules, accelerator, `workloadReplicas`); it does **not** make an app shared or admin-only. The shared semantics are triggered by **`options.shared: true`** (only honored on a v3 app). `options.shared: true` ⇒:

- **Admin-only install** — non-admin callers are rejected ("only admin users can install shared apps").
- **Deterministic shared namespace** `<app>-shared` (not the per-user `<app>-<owner>`), owned by the cluster owner so it stays stable across admins.
- **Cluster-wide ApplicationManager** named `<app>-shared-<app>`.
- **Cross-namespace shared access is force-enabled** so the app is a first-class destination for cross-namespace traffic (service-mesh sidecar + NetworkPolicy).

So **`options.shared: true` is the "shared app" declaration** (on top of `apiVersion: v3`) — there is no separate sub-chart / app-scope wiring to add. (`spec.onlyAdmin: true` is a separate, independent admin-only-install gate that any app can set without being shared.)

> **There is no `sharedEntrances`.** Consumers reach the shared app's in-cluster Service directly (below). Do **not** add `sharedEntrances` — if you see it in an existing chart, treat it as a leftover and drop it.

## Manifest (annotated, based on `ollamav3`)

```yaml
olaresManifest.version: '0.12.0'      # accelerator / externalData
olaresManifest.type: app
apiVersion: 'v3'                       # schema generation (env/accelerator); shared-ness comes from options.shared below
metadata:
  name: ollamav3
  # ...
spec:
  onlyAdmin: true                      # independent admin-only-install gate (optional; shared already implies admin-only)
  accelerator:                         # heavy/GPU resource envelope — see accelerator.md
    - mode: nvidia
      requiredMemory: 5Gi
      limitedMemory: 40Gi
      requiredGPUMemory: 1Gi
      limitedGPUMemory: 24Gi
    - mode: apple-m
      # ...
permission:
  appData: true
  appCache: true
options:
  shared: true                         # THIS is what makes the app a cluster-wide shared singleton (<app>-shared)
  allowMultipleInstall: true
  conflicts:
    - name: ollamav2                   # conflict with the non-shared variant
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
| `apiVersion: 'v3'` | selects the v3 manifest schema (env/accelerator rules) — prerequisite for `options.shared` |
| `options.shared: true` | **what actually makes the app shared**: admin-only install, `<app>-shared` namespace, cross-namespace access |
| `olaresManifest.version: '0.12.0'` | needed for `spec.accelerator` / `permission.externalData` (see [manifest.md](olares-chart-manifest.md)) |
| `options.dependencies` `olares` `>=1.12.6-0` (`type: system`) | **mandatory** system dependency every chart declares ([versioning.md](olares-chart-versioning.md)) |
| `spec.onlyAdmin: true` | independent admin-only-install gate (redundant for shared apps, which are already admin-only) |
| `spec.accelerator` | GPU/accelerator envelope for the heavy backend — sizing in [olares-chart-accelerator.md](olares-chart-accelerator.md) |
| `options.conflicts` | avoid co-installing the per-user / older variant of the same backend |
| `middleware` | shared backends can use system middleware normally (e.g. `llmgatewayv3` uses postgres) — see [olares-chart-middleware.md](olares-chart-middleware.md) |

## How consumers (reference apps) reach it

A shared app is consumed by **separate reference/client apps**, not by per-user copies of itself. The client declares an `options.dependencies` `type: application` on the shared app; app-service then injects the shared namespace's Services into the client's Helm values and grants cross-namespace reachability automatically. The full mechanism (`.Values.svcs.<svc>_host` = `<svc>.<app>-shared`, mesh sidecar + NetworkPolicy, no entrance/URL) is the platform **App, namespace & networking model** (loaded via the SKILL.md prerequisite).

## Caveats

- **Only an admin can install a shared app** — surface this to the user; a normal user install will 403.
- **`<app>-shared` namespace requires `options.shared: true`** in the manifest. Without it, even an `apiVersion: v3` app installs into the **installing user's** `<app>-<owner>` namespace. Set `options.shared: true` when the app genuinely needs cluster-wide shared storage/services accessible across namespaces; omit it (and use `spec.onlyAdmin: true` instead) for an app that just needs admin-only install but manages its own users internally.
- Do not add `sharedEntrances` (not supported).
- A shared app is still subject to the rest of the skill: run-as-user 1000 ([run-as-user.md](olares-chart-run-as-user.md)), pinned image tags, accelerator sizing ([accelerator.md](olares-chart-accelerator.md)), env rules ([env.md](olares-chart-env.md)).
