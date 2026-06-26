---
name: olares-chart
version: 4.6.0
description: "Help a developer turn their own code or any open-source project into an app that runs on their own Olares. Two coupled axes: packaging the container image and authoring/refining the Olares app chart (OlaresManifest), then deploying it to the current Olares with an automatic upload + install + diagnose loop. Use when deploying a repo, docker-compose, or Helm chart to Olares, packaging an Olares app, wiring storage / system middleware / entrances / env / GPU, or fixing a failed install (ImagePullBackOff, permission denied / EACCES, app won't start). Publishing to the public Olares Market is the olares-publish skill."
compatibility: Requires olares-cli on PATH; chart authoring is local-only, deploy needs login
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# chart (compose → Olares app chart)

> **Source of truth for flags is always `olares-cli chart <verb> --help`.** This file only carries what `--help` cannot: when to use this skill, the convert→refine→lint loop, and the four judgment calls that turn a raw conversion into a publishable chart.

## When to use

- Olares chart, OlaresManifest / OlaresManifest.yaml, olares-cli chart, chart from-compose, chart lint, chart package
- Turn docker-compose, a generic Helm chart, or a bare source repo into a publishable Olares app chart
- Packaging for Olares Market / app store, building/pushing docker image (amd64 vs arm64), no official image, wrong arch
- Install/runtime failures: ImagePullBackOff, app failed to install or start, market / app-service / chartrepo logs
- Headless / CLI app, MCP server, or tool with no web UI (terminal entrance + invisible entrance)
- Two axes: **packaging** (Dockerfile/image) and **deployment** (compose/chart); four post-kompose judgment calls (metadata, storage, middleware, entrances)
- Optional live validation (requires login): package + market upload/install — see [`olares-shared`](../olares-shared/SKILL.md), [`olares-market`](../olares-market/SKILL.md), [`olares-cluster`](../olares-cluster/SKILL.md)

## Start here: establish your state

Before doing anything, state where the app stands on **two orthogonal axes**. Each axis has its own "ready" target; pick the capability that advances whichever axis is behind. The axes are orthogonal (having a compose says nothing about whether an image needs building) but **coupled**: an image's baked-in paths / run-user constrain what the manifest can mount and which `permission`s it grants, and Olares deployment constraints (non-root, userspace volumes, system middleware) can send you back to rebuild the image rather than just edit the manifest. Drive both axes to ready and loop across these edges as constraints surface.

**Packaging axis — the image** (how the app is built into a runnable artifact). Olares **pulls images from a registry and never builds from source**, so every workload must end up referencing a publicly pullable, node-arch-correct image.

| Packaging state | Do this | Ready when |
|---|---|---|
| No Dockerfile (just source) | author a Dockerfile, then build+push | — |
| Dockerfile, but no pullable image | build+push (Docker Hub or ghcr) | — |
| A pullable image exists | check its arch; rebuild / make multi-arch if it doesn't match the node | every workload has a pullable, arch-correct image |

**Deployment axis — the orchestration** (how the app is deployed). The target is a `lint`-passing Olares chart.

| Deployment state | Do this | Ready when |
|---|---|---|
| Source only (no compose) | author a docker-compose from the code | — |
| A CLI / headless service (no web UI) | classify it and apply an archetype recipe ([archetypes.md](references/olares-chart-archetypes.md)) | a chart that passes `chart lint` |
| A docker-compose | `chart from-compose` then refine | — |
| A generic Helm chart (no OlaresManifest) | hand-author `OlaresManifest.yaml` + refine (skip `from-compose`) | — |
| Already an Olares chart | go straight to validation | a chart that passes `chart lint` |

The image-building work is **guided** — you check/install docker and drive `docker login` / build / push **with the developer, never on their behalf** ([references/olares-chart-image.md](references/olares-chart-image.md)). Local authoring (`from-compose` / `lint` / `package`) needs **no login**; live validation does (see below).

## Capabilities (composable, loopable)

| Capability | Axis | What it does | Login? | Loop back here when | Reference |
|---|---|---|---|---|---|
| **Image** | packaging | author a Dockerfile if needed, build + push a pullable, arch-correct (multi-arch) image | no (docker + registry) | install hits `ImagePullBackOff` / wrong arch, or a deploy constraint forces a rebuild | [image.md](references/olares-chart-image.md) |
| **Compose** | deployment | obtain or author a docker-compose from the code | no | — | [compose.md](references/olares-chart-compose.md) |
| **Convert** | deployment | `chart from-compose` scaffolds an Olares chart | no | — | [from-compose.md](references/olares-chart-from-compose.md) |
| **Refine** | deployment | the four judgment calls / hand-author `OlaresManifest.yaml` | no | `lint` fails, or install fails on env/wiring | [manifest.md](references/olares-chart-manifest.md) |
| **Validate-local** | deployment | `chart lint` + `chart package` | no | — | [lint.md](references/olares-chart-lint.md) |
| **Validate-live** | both | `market upload` + `market install` + diagnose from logs | yes | — | [publish-verify.md](references/olares-chart-publish-verify.md) |

> **Validate-live** leans on sibling skills: [`olares-shared`](../olares-shared/SKILL.md) (login check), [`olares-market`](../olares-market/SKILL.md) (upload / install / cleanup), [`olares-cluster`](../olares-cluster/SKILL.md) (logs). **Never log in or upload on the developer's behalf without asking first.**

## App archetypes (thin upstream context)

| Axis | Concern | Get this right | Loop back when | Reference |
|---|---|---|---|---|
| packaging | **Image** | pullable, pinned to a version tag (never `:latest`), arch-correct for **this node** (multi-arch only when publishing) | `ImagePullBackOff` / wrong arch, or a deploy constraint forces a rebuild | [image.md](references/olares-chart-image.md) |
| packaging+deployment | **Run identity** | process runs as uid 1000; `spec.runAsUser: true`; initContainer `chown` for root-owned volumes; no root main on non-trusted images (OPA) | EACCES on appData/appCache/userData, admission denies a root third-party image | [run-as-user.md](references/olares-chart-run-as-user.md) |
| deployment | **Storage** | every compose volume mapped to the right userspace area (Data / Cache / Home / Common / External), `permission` declared to match, leftover kompose PVCs deleted | a volume isn't persisting or lands in the wrong area | [manifest.md](references/olares-chart-manifest.md) §2 |
| deployment | **Middleware & deps** | no bundled `postgres`/`redis`/`mongo`/…; wired to system middleware; SQLite→Postgres where supported; companion apps as `type: application` deps | a bundled db/queue is still in the chart, or a companion should be a dependency | [middleware.md](references/olares-chart-middleware.md) |
| deployment | **Env** | app config in `envs[]` (v3 `valueFrom`, no inline `OLARES_USER`); install-time `required` prompts; middleware/system/user vars mapped via `.Values.olaresEnv`; platform render context (identity, domain, userspace, oidc, middleware) consumed via `.Values.*` | install fails on `appenv` 422, or config must be user-supplied | [env.md](references/olares-chart-env.md), [env-defaults.md](references/olares-chart-env-defaults.md), [system-values.md](references/olares-chart-system-values.md) |
| deployment | **Entrances & ports** | ≥1 `entrances[]`; HTTP via entrances, non-HTTP via `ports[]`; internal-only services `invisible: true` | a service is unreachable, or an internal port is exposed as a desktop entrance | [manifest.md](references/olares-chart-manifest.md) §4 |
| packaging+deployment | **GPU / models** | build a CUDA image without a local GPU (custom-kernel arch flags); download model weights via initContainer into the shared `appCommon` Hugging Face cache | AI app needs a CUDA build, model provisioning, or a shared model cache | [gpu.md](references/olares-chart-gpu.md) |
| deployment | **LLM model serving** | serve any HF/Ollama model with no chart authoring — pick an engine by format, fill env, clone an `llm-init` base app (llama.cpp / Ollama / vLLM / SGLang); set the `MODEL_SUPPORTS` capability field from the model card (don't claim `vision` on a text-only GGUF) | user wants to run/serve/host a specific LLM or embedding model, not author a new app | [llm-models.md](references/olares-chart-llm-models.md) |
| deployment | **Accelerator** | declare `spec.accelerator` modes (nvidia/amd-gpu/apple-m/cpu/…) per what the repo supports; set `requiredGPUMemory`; a sane CPU/memory envelope | GPU/accelerator app needs a resource envelope, or `lint` flags `spec.resources` | [accelerator.md](references/olares-chart-accelerator.md) |
| packaging+deployment | **DinD** | a privileged `beclab/docker` daemon sidecar (`ENABLE_DIND`, `DOCKER_HOST`) while the main container stays non-privileged | a terminal/agent app must run `docker` / `docker compose` | [dind.md](references/olares-chart-dind.md) |
| deployment | **Shared backend** | `apiVersion: v3` ⇒ admin-only install into `<app>-shared`; consumers reach it over cross-namespace Service DNS; flag the admin-install to the user | a heavy/accelerator backend serves many users over shared data | [shared.md](references/olares-chart-shared.md) |
| deployment | **Version rules** | `apiVersion: v3`; `olaresManifest.version` always `0.12.0`; `metadata.version` == `Chart.yaml` `version`; `options.dependencies` `olares >=1.12.6-0` (`type: system`) | install rejects the manifest, or behavior differs by Olares version | [versioning.md](references/olares-chart-versioning.md) |
| deployment | **Metadata** | stub OK for local deploy (`Utilities`, default icon) as long as `lint` passes; full `metadata.*` + listing images only when publishing to the Market | `lint` flags missing metadata, or you want a public listing | [manifest.md](references/olares-chart-manifest.md) §1 |
| deployment | **Validate-local** | `olares-cli chart lint ./<app>` passes, then `chart package` | a refinement changed the manifest/templates | [lint.md](references/olares-chart-lint.md) |
| deploy | **Deploy** | `market upload` + `market install`, then diagnose from logs — automatic after `lint` passes (login required) | proving the chart actually runs on the developer's Olares | [deploy.md](references/olares-chart-deploy.md) |

| Archetype | Olares mapping | Reference |
|---|---|---|
| Headless CLI / service (no web UI) — CLI tool, MCP server, API daemon | a **web terminal** as a visible desktop entrance + the service/MCP port as an **invisible** internal entrance (Settings-only) | [archetypes.md](references/olares-chart-archetypes.md) |

## Routing (this skill vs siblings)

Use this skill to **author/validate your own** Olares chart from a repo, compose, or Helm chart (see the state tables above for where to start). Hand off to a sibling when the task is not authoring:

| User intent | Where |
|---|---|
| Turn a repo / compose / Helm chart into an Olares app, or validate one you authored | ✅ this skill |
| "My chart won't install / the app won't start — why?" | ✅ this skill — diagnosis step of [references/olares-chart-publish-verify.md](references/olares-chart-publish-verify.md) (then [`olares-cluster`](../olares-cluster/SKILL.md) for deeper log digging) |
| "Just install / upgrade an existing catalog app" (not validating your own chart) | [`olares-market`](../olares-market/SKILL.md) |
| "Inspect pods / logs of an unrelated running app" | [`olares-cluster`](../olares-cluster/SKILL.md) |

> **Mental model:** the heart of this skill is **authoring** (produce a valid chart on disk). Live validation — uploading and running the chart you just authored to prove it works — is an opt-in extension here that orchestrates the sibling skills; routine app-store lifecycle on apps you did not author belongs to `olares-market`.

## CLI verbs

The only `olares-cli chart` subcommands (source of truth: `--help`). Everything else above is docker or sibling skills.

| Verb | What it does | Reference |
|---|---|---|
| `from-compose` (alias `init`) | kompose-convert compose file(s) into an Olares chart skeleton | [references/olares-chart-from-compose.md](references/olares-chart-from-compose.md) |
| `lint` | validate a chart dir / `.tgz` with the Market ingest pipeline | [references/olares-chart-lint.md](references/olares-chart-lint.md) |
| `package` | package a chart dir into a `<name>-<version>.tgz` for upload (mirrors `helm package`, no helm binary needed) | [references/olares-chart-publish-verify.md](references/olares-chart-publish-verify.md) |

## A typical assembly (compose with a build-only image)

One way to compose the capabilities; not a fixed pipeline — start wherever your state tables put you, and loop across the coupling edges as failures surface.

```
 Packaging (guided, with the developer; only if an image is missing / wrong-arch):
 P1. docker?    docker version && docker buildx version   # else guide install
 P2. registry   docker login  (Docker Hub)  |  docker login ghcr.io  (ghcr, PAT write:packages)
 P3. build+push docker buildx build --platform linux/amd64,linux/arm64 -t <user>/<repo>:<tag> --push <ctx>
                -> wire <user>/<repo>:<tag> into every build-only `image:` in the compose

 Deployment authoring (no login):
 D1. scaffold   olares-cli chart from-compose --name <app> -f docker-compose.yml
 D2. refine     edit OlaresManifest.yaml + templates/ for the 4 judgment areas
 D3. lint       olares-cli chart lint ./<app>        # loop D2<->D3 until OK
 D4. package    olares-cli chart package ./<app>

 Live validation (requires login + developer consent):
 V1. logged-in? olares-cli profile list              # if not: tell developer, stop
 V2. ask        confirm with the developer before uploading to a real Olares
 V3. upload     olares-cli market upload ./<app>-<ver>.tgz
 V4. run        olares-cli market install <app> -s upload --version <ver> --watch -o json
 V5. on failure fetch market / chartrepo / app-service / app-pod logs and diagnose
 V6. decide     loop back: chart problem -> D2 ; image problem -> P3 ; else report & ask
 V7. cleanup    olares-cli market uninstall <app> --watch ; olares-cli market delete <app>
```

Step D1 produces a chart that **already passes `lint`** but is NOT yet a good app: kompose translates containers literally and cannot make product decisions. The value you add is D2. Treat the generated `OlaresManifest.yaml` as a stub. The V steps are optional and cross into sibling skills — full procedure, log targets, and the admin caveat live in [references/olares-chart-publish-verify.md](references/olares-chart-publish-verify.md); only proceed past D3 with the developer's consent.

## What the conversion produces

```
<output>/
├── Chart.yaml              # helm chart metadata (name/version pinned to 0.0.1)
├── OlaresManifest.yaml     # Olares app manifest — the file you refine
├── values.yaml             # empty; fill if you template values
└── templates/
    ├── deployment-<app>.yaml          # the primary workload, renamed to <app>
    ├── deployment-<svc>.yaml          # one per extra compose service
    ├── service-<svc>.yaml             # one per exposed compose service
    └── persistentvolumeclaim-*.yaml   # one per named/anonymous compose volume
```

- Every resource is namespaced with `namespace: '{{ .Release.Namespace }}'`.
- Default CPU/memory requests+limits are stamped onto every container.
- One **entrance** is auto-detected (the `olares.service.type: Entrance`-labeled service, else the first service with a port, else a `port: 80` placeholder).
- `olaresManifest.version` is `0.8.0` (legacy) unless you pass `--new-schema` (`0.12.0`, resources under `spec.accelerator`).

> **Tip:** label the service you want exposed in the compose file with `labels: { olares.service.type: Entrance }` so the right workload becomes the entrance and gets renamed to the app name.

## The four judgment calls (the actual work)

kompose cannot decide these — you must. Full field-by-field mapping and edit recipes are in [references/olares-chart-manifest.md](references/olares-chart-manifest.md).

1. **Metadata** — kompose leaves a stub (`title=name`, default icon, `Utilities` category, no developer info). Set `metadata.{title,icon,description,categories}` and `spec.{developer,website,sourceCode,submitter,fullDescription}` from the upstream project (or ask the user).
2. **Storage** — compose `volumes:` become raw PVCs. Decide each one: app-private state → `.Values.userspace.appData` / `.Values.userspace.appCache` (set `permission.appData/appCache: true`); user-visible files → `.Values.userspace.userData` + list the path under `permission.userData`. Delete the kompose PVCs you replaced and rewrite the `volumeMounts`.
3. **Middleware** — a compose `postgres`/`redis`/`mongo`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` service should usually be dropped and replaced by Olares system middleware: add a `middleware:` block + an `options.dependencies` entry (type `middleware`), delete that workload + its PVC, and repoint the app's env vars at `.Values.<mw>.*`.
4. **Entrances & ports** — keep/add one `entrances[]` per user-facing HTTP service (tune `host`/`port`/`title`/`authLevel`); expose non-HTTP services via `ports[]` (`exposePort`). Mark internal-only services `invisible: true` or drop their entrance.

## Hard constraints that bite

- **`metadata.name` must match the chart folder name and `Chart.yaml` `name`**, and be `^[a-z][a-z0-9]{0,29}$`. `from-compose --name` keeps them consistent; if you rename the folder, fix all three.
- **At least one entrance is required.** Never delete the last `entrances[]` entry.
- **If a template uses `.Values.userspace.appData`/`appCache`/`userData`, the matching `permission` field MUST be declared**, or `lint` fails the app-data cross-check.
- **`hostPath` volumes + rolling updates are incompatible** — `lint` rejects them. Replace host mounts with the userspace volumes above.
- **`metadata.version` (Chart Version) and `Chart.yaml` `version` must match**, and `spec.versionName` should track the upstream app version (`Chart.yaml` `appVersion`).

## Common errors → fix

| `lint` says | Cause | Fix |
|---|---|---|
| `must have a Deployment or StatefulSet named "<app>"` | no workload named after the app | one workload must be `metadata.name: <app>` — `from-compose` renames the entrance workload automatically; preserve that |
| app-data permission mismatch | template uses `.Values.userspace.*` but `permission` doesn't declare it (or vice versa) | align `permission.appData/appCache/userData` with what the templates mount |
| version mismatch | `Chart.yaml` `version` ≠ `metadata.version` | make them equal |
| hostPath + rolling update | a template mounts a `hostPath` | switch to a userspace volume |
| manifest structural error | a required manifest field is missing/invalid after editing | re-check against [references/olares-chart-manifest.md](references/olares-chart-manifest.md) |
