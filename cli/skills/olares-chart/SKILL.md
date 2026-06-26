---
name: olares-chart
version: 4.10.0
description: "Use when deploying a repo, docker-compose, or generic Helm chart to your own Olares, packaging an Olares app image, authoring or validating an OlaresManifest, wiring storage / system middleware / entrances / env / GPU, or fixing a failed install (ImagePullBackOff, permission denied / EACCES, app won't start or won't reach running). Publishing to the public Olares Market is the olares-publish skill."
compatibility: Requires olares-cli on PATH; chart authoring is local-only, deploy needs login
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# chart (compose → Olares app chart)

> **Source of truth for flags is always `olares-cli chart <verb> --help`.** This file only carries what `--help` cannot: the two coupled axes of a port, where to start on each, the per-axis concerns to get right, and the gotchas `lint` won't catch.

> **Porting baseline: Olares >= 1.12.6.** Check the target with `olares-cli profile list` (VERSION column). Full version rules — `apiVersion: v3`, the chart version fields, the `olares` `type: system` dependency — are in [references/olares-chart-versioning.md](references/olares-chart-versioning.md).

> **Platform model (read once, no login needed for authoring).** Porting decisions rely on the Olares storage model, uid-1000 run identity, app/namespace & networking, system middleware, and version model — all defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md). Packaging an image and authoring/validating the chart need no login; only **deploy to your Olares** (`market upload` + `install`) does.

## When to use

- Turn a repo / docker-compose / generic Helm chart into an Olares app, or validate an OlaresManifest; package its image; wire storage / middleware / entrances / env / GPU
- Deploy / run the app on **your own** Olares (`market upload` + `install`), or fix a failed install (`ImagePullBackOff`, `EACCES`, app won't start)
- Serve a specific LLM / embedding model (HF or Ollama) with no chart authoring — clone an `llm-init` base app and fill env ([llm-models.md](references/olares-chart-llm-models.md))

## Start here: establish your state

Before doing anything, state where the app stands on **two orthogonal axes**. Each axis has its own "ready" target; pick the capability that advances whichever axis is behind. The axes are orthogonal (having a compose says nothing about whether an image needs building) but **coupled**: an image's baked-in paths / run-user constrain what the manifest can mount and which `permission`s it grants, and Olares deployment constraints (non-root, userspace volumes, system middleware) can send you back to rebuild the image rather than just edit the manifest. Drive both axes to ready and loop across these edges as constraints surface.

## The shape of the work — two axes

Porting an app is **not** a fixed `from-compose → lint → deploy` pipeline — it is driving two **orthogonal but coupled** axes each to its own *ready* state, looping back as constraints surface (an image's baked-in uid/paths constrain the chart's mounts/permissions; a deploy constraint can send you back to rebuild the image). Start wherever your app already stands, not at a fixed step 1. Once both axes are ready, **deploy to the current Olares** — an automatic upload + install + diagnose loop.

- **Packaging — the image:** the app built into a pullable, arch-correct artifact. Olares only pulls, never builds.
- **Deployment — the chart:** a `lint`-passing OlaresManifest + templates. `from-compose` is only **one** way in.

**First move (not a pipeline):** locate where the app already sits on the packaging and deployment state tables → drive the concerns to ready, looping as constraints surface → deploy to your Olares.

## Axis 1 — Packaging (the image)

Olares **pulls images from a registry and never builds from source**, so every workload must reference a publicly pullable, node-arch-correct image. Image work is **agent-driven**: ask which registry the developer uses (Docker Hub / ghcr), check docker is usable and logged in, then **build + push yourself** — only `docker login` stays manual, and only when not already authenticated ([references/olares-chart-image.md](references/olares-chart-image.md)). No Olares login needed. Build for **this node's arch** (single-arch); multi-arch is only for publishing.

| Packaging state | Do this | Ready when |
|---|---|---|
| No Dockerfile (just source) | author a Dockerfile, then build+push | — |
| Dockerfile, but no pullable image | build+push (Docker Hub or ghcr) | — |
| A pullable image exists | check its arch; rebuild / make multi-arch if it doesn't match the node | every workload has a pullable, arch-correct image |

## Axis 2 — Deployment (the chart)

The target is a `lint`-passing Olares chart. `from-compose` (kompose) is **just one entry method** — a bare repo, a generic Helm chart, or an already-Olares chart each begin elsewhere (see the state table below). Local authoring (`from-compose` / `lint` / `package`) needs **no login**.

| Deployment state | Do this | Ready when |
|---|---|---|
| Source only (no compose) | author a docker-compose from the code | — |
| A CLI / headless service (no web UI) | classify it and apply an archetype recipe ([archetypes.md](references/olares-chart-archetypes.md)) | a chart that passes `chart lint` |
| A docker-compose | `chart from-compose` then refine | — |
| A generic Helm chart (no OlaresManifest) | hand-author `OlaresManifest.yaml` + refine (skip `from-compose`) | — |
| Already an Olares chart | go straight to validation | a chart that passes `chart lint` |

The image-building work is **guided** — you check/install docker and drive `docker login` / build / push **with the developer, never on their behalf** ([references/olares-chart-image.md](references/olares-chart-image.md)). Local authoring (`from-compose` / `lint` / `package`) needs **no login**; live validation does (see below).

Both axes ready → **deploy to the current Olares automatically**. `lint` proves the chart is structurally valid; it does **not** prove the app pulls its images, wires its middleware, and reaches `running` — the deploy loop does. **After `lint` passes, proceed without asking:** check login → package → `market upload` → `market install -s upload --watch` → on failure fetch logs → diagnose → fix chart + re-lint → retry. Only stop to ask when the profile is not logged in, or a failure is clearly not a chart problem. Full procedure: [references/olares-chart-deploy.md](references/olares-chart-deploy.md).

For deploying to your own Olares, **metadata can stay a stub** as long as `lint` passes; functional refinement (storage / middleware / entrances) is still required.

> **Validate-live** leans on sibling skills: [`olares-shared`](../olares-shared/SKILL.md) (login check), [`olares-market`](../olares-market/SKILL.md) (upload / install / cleanup), [`olares-cluster`](../olares-cluster/SKILL.md) (logs). **Never log in or upload on the developer's behalf without asking first.**

## App archetypes (thin upstream context)

| Axis | Concern | Get this right | Loop back when | Reference |
|---|---|---|---|---|
| packaging | **Image** | pullable, pinned to a version tag (never `:latest`), arch-correct for **this node** | `ImagePullBackOff` / wrong arch, or a deploy constraint forces a rebuild | [image.md](references/olares-chart-image.md) |
| packaging+deployment | **Run identity** | uid 1000; `spec.runAsUser: true`; initContainer `chown` for root-owned volumes; no root main on non-trusted images (OPA) | EACCES on appData/appCache/userData; admission denies a root third-party image | [run-as-user.md](references/olares-chart-run-as-user.md) |
| deployment | **Storage** | every compose volume → the right userspace area (Data/Cache/Home/Common/External), matching `permission`, leftover kompose PVCs deleted | a volume isn't persisting or lands in the wrong area | [manifest.md](references/olares-chart-manifest.md) §2 |
| deployment | **Middleware & deps** | no bundled `postgres`/`redis`/`mongo`/…; wire to system middleware; SQLite→Postgres where supported; companion apps as `type: application` deps | a bundled db/queue remains, or a companion should be a dependency | [middleware.md](references/olares-chart-middleware.md) |
| deployment | **Env** | app config in `envs[]` (v3 `valueFrom`, no inline `OLARES_USER`); install-time `required` prompts; middleware/system/user vars via `.Values.olaresEnv`; platform context via `.Values.*` | install fails on `appenv` 422, or config must be user-supplied | [env.md](references/olares-chart-env.md), [env-defaults.md](references/olares-chart-env-defaults.md), [system-values.md](references/olares-chart-system-values.md) |
| deployment | **Entrances & ports** | ≥1 `entrances[]`; HTTP via entrances, non-HTTP via `ports[]`; internal-only services `invisible: true` | a service is unreachable, or an internal port is exposed as a desktop entrance | [manifest.md](references/olares-chart-manifest.md) §4 |
| packaging+deployment | **GPU / models** | build a CUDA image without a local GPU; download model weights via initContainer into the shared `appCommon` Hugging Face cache | AI app needs a CUDA build, model provisioning, or a shared model cache | [gpu.md](references/olares-chart-gpu.md) |
| deployment | **LLM model serving** | serve any HF/Ollama model without authoring — pick an engine by format, fill env, clone an `llm-init` base app (llama.cpp / Ollama / vLLM / SGLang); set `MODEL_SUPPORTS` from the model card | user wants to run/serve a specific LLM or embedding model, not author a new app | [llm-models.md](references/olares-chart-llm-models.md) |
| deployment | **Accelerator** | declare `spec.accelerator` modes per repo support; set `requiredGPUMemory`; a sane CPU/memory envelope | GPU/accelerator app needs a resource envelope, or `lint` flags `spec.resources` | [accelerator.md](references/olares-chart-accelerator.md) |
| packaging+deployment | **DinD** | a privileged `beclab/docker` daemon sidecar (`ENABLE_DIND`, `DOCKER_HOST`); main container stays non-privileged | a terminal/agent app must run `docker` / `docker compose` | [dind.md](references/olares-chart-dind.md) |
| deployment | **Shared backend** | `apiVersion: v3` ⇒ admin-only install into `<app>-shared`; consumers reach it via cross-namespace Service DNS; flag the admin-install | a heavy/accelerator backend serves many users over shared data | [shared.md](references/olares-chart-shared.md) |
| deployment | **Version & deps fields** | fixed values every chart writes: `apiVersion: v3`; `olaresManifest.version: '0.12.0'`; `metadata.version` == `Chart.yaml` `version`; `options.dependencies` includes `olares >=1.12.6-0` (`type: system`) — authored by you | every manifest edit — confirm these fields + the `olares` system dep (`lint` rejects a missing system dep) | [versioning.md](references/olares-chart-versioning.md) |
| deployment | **Workloads / replicas** | `workloadReplicas` lists every Deployment/StatefulSet → count (authored by you); each `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}` + matching `values.yaml` | every time you author/add/rename a workload — run the three-point self-check; a hardcoded `replicas` silently no-ops suspend/resume / staged install | [manifest.md](references/olares-chart-manifest.md) Workloads & replicas |
| deployment | **Metadata** | stub OK for local deploy (`Utilities`, default icon) while `lint` passes; full `metadata.*` + listing images only when publishing | `lint` flags missing metadata, or you want a public listing | [manifest.md](references/olares-chart-manifest.md) §1 |
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

This is the canonical source for cross-app wiring this skill intentionally does not hardcode.

## What the conversion produces

- **Apps with a `.suspend` (or `.remove`) control file in the OAC root** — suspended / no longer distributed; not a current, reliable pattern.
- **Shared / cluster-scoped charts** that express sharing with `spec.subCharts[].shared: true` + `options.appScope.clusterScoped: true` + `appRef` (the `ollamaserver`/`ollamav2` shape). Copy the shared-app pattern from an `apiVersion: v3` app, not from these. See [shared.md](references/olares-chart-shared.md).

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
