---
name: olares-chart
version: 1.8.0
description: "Olares Chart via olares-cli chart — from-compose, lint, package; turn compose/Helm/repo into an Olares app chart. Release targets: local-run (upload on your Olares) or market-distribute (public Market). Use for OlaresManifest, docker-compose to Olares, chart lint/package, Market upload, ImagePullBackOff."
compatibility: Requires olares-cli on PATH; chart authoring is local-only
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# chart (compose → Olares app chart)

> **Source of truth for flags is always `olares-cli chart <verb> --help`.** This file only carries what `--help` cannot: when to use this skill, the convert→refine→lint loop, release-target routing, and the four refinement areas that turn a raw conversion into a working chart.

## When to use

- Olares chart, OlaresManifest / OlaresManifest.yaml, olares-cli chart, chart from-compose, chart lint, chart package
- Turn docker-compose, a generic Helm chart, or a bare source repo into a **lint-passing** Olares app chart — or a **market-ready** one when distributing publicly
- **Local run** on your own Olares (upload + install); **market distribute** (full metadata, multi-arch, PR to `beclab/apps`)
- Building/pushing docker image (amd64 vs arm64), no official image, wrong arch
- Install/runtime failures: ImagePullBackOff, app failed to install or start, market / app-service / chartrepo logs
- **Permission denied / EACCES** on userspace volumes, third-party image runs as root or non-1000 uid, `spec.runAsUser`, initContainer volume `chown`
- Headless / CLI app, MCP server, or tool with no web UI (terminal entrance + invisible entrance)
- Three axes: **packaging** (Dockerfile/image), **deployment** (compose/chart), **publishing** (release target); four post-kompose refinement areas (metadata depth gated by target, storage, middleware, entrances)
- Optional live validation (requires login): package + market upload/install — see [`olares-shared`](../olares-shared/SKILL.md), [`olares-market`](../olares-market/SKILL.md), [`olares-cluster`](../olares-cluster/SKILL.md)

## Start here: establish your release target

Before the packaging/deployment state tables, decide **who consumes the chart** and what "done" means. Infer from user language; ask if ambiguous. Full decision tree and checklists: [references/olares-chart-publish-targets.md](references/olares-chart-publish-targets.md).

| Release target | User signals | Done when |
|---|---|---|
| **local-run** (default for most users) | "run on my Olares", "upload and install", "just for myself" | `lint` OK → package → upload + install reaches `running` on the developer's Olares |
| **market-distribute** | "publish to Market", "submit to beclab/apps", "上架" | local validation passes **plus** market-ready metadata/images/arch → PR merged into `beclab/apps:main` |

**Requirements matrix** (summary — detail in [publish-targets.md](references/olares-chart-publish-targets.md)):

| Concern | local-run | market-distribute |
|---|---|---|
| Image arch | single-arch matching **this node's** arch (`cluster node list`) | multi-arch (`linux/amd64,linux/arm64`); declare `spec.supportArch` |
| Metadata §1 | stub OK if `lint` passes (`Utilities`, default icon) | full metadata, dual-version categories, `fullDescription`, developer links |
| Listing assets | skip | `featuredImage`, `promoteImage` |
| Refine §2–4 | **same for both** — storage / middleware / entrances are functional, not cosmetic |
| Validate | D3 lint → D4 package → V upload/install | same V steps first, then M market submit |

## Start here: establish your state

State where the app stands on **three orthogonal axes**. Packaging and deployment each have their own "ready" target; **publishing** gates how strictly you apply image arch and metadata depth. The axes are orthogonal (having a compose says nothing about whether an image needs building) but **coupled**: an image's baked-in paths / run-user constrain what the manifest can mount and which `permission`s it grants, and Olares deployment constraints (non-root, userspace volumes, system middleware) can send you back to rebuild the image rather than just edit the manifest. Drive all three to ready and loop across these edges as constraints surface.

**Packaging axis — the image** (how the app is built into a runnable artifact). Olares **pulls images from a registry and never builds from source**, so every workload must end up referencing a publicly pullable, node-arch-correct image.

| Packaging state | Do this | Ready when |
|---|---|---|
| No Dockerfile (just source) | author a Dockerfile, then build+push | — |
| Dockerfile, but no pullable image | build+push (Docker Hub or ghcr) | — |
| A pullable image exists | check its arch; rebuild if it doesn't match the target (node arch for local-run; multi-arch for market-distribute) | every workload has a pullable, arch-correct image |

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
| **Image** | packaging | author a Dockerfile if needed, build + push a pullable, arch-correct image (single-arch or multi-arch per release target) | no (docker + registry) | install hits `ImagePullBackOff` / wrong arch, or a deploy constraint forces a rebuild | [image.md](references/olares-chart-image.md) |
| **Compose** | deployment | obtain or author a docker-compose from the code | no | — | [compose.md](references/olares-chart-compose.md) |
| **Convert** | deployment | `chart from-compose` scaffolds an Olares chart | no | — | [from-compose.md](references/olares-chart-from-compose.md) |
| **Refine** | deployment | the four refinement areas / hand-author `OlaresManifest.yaml` | no | `lint` fails, or install fails on env/wiring | [manifest.md](references/olares-chart-manifest.md) |
| **Run-as-user** | packaging + deployment | align image uid with Olares userspace (1000): Dockerfile `USER`, `spec.runAsUser`, initContainer `chown` | no | EACCES on appData/appCache/userData, OPA root deny on third-party image | [run-as-user.md](references/olares-chart-run-as-user.md) |
| **Validate-local** | deployment | `chart lint` + `chart package` | no | — | [lint.md](references/olares-chart-lint.md) |
| **Publish-local** | publishing | `market upload` + `market install` + diagnose from logs | yes | — | [publish-verify.md](references/olares-chart-publish-verify.md) |
| **Publish-market** | publishing | market-ready checklist + `beclab/apps` PR guidance | no (GitHub) | local validation passed, user wants public Market listing | [market-submit.md](references/olares-chart-market-submit.md) |

> **Publish-local** leans on sibling skills: [`olares-shared`](../olares-shared/SKILL.md) (login check), [`olares-market`](../olares-market/SKILL.md) (upload / install / cleanup), [`olares-cluster`](../olares-cluster/SKILL.md) (logs). **Never log in or upload on the developer's behalf without asking first.**

## App archetypes (thin upstream context)

When the upstream ships no compose/chart and the deployment shape is unclear, classify it into an archetype and apply a vetted Olares recipe before refining. Full templates and the canonical example chart live in [references/olares-chart-archetypes.md](references/olares-chart-archetypes.md).

| Archetype | Olares mapping | Reference |
|---|---|---|
| Headless CLI / service (no web UI) — CLI tool, MCP server, API daemon | a **web terminal** as a visible desktop entrance + the service/MCP port as an **invisible** internal entrance (Settings-only) | [archetypes.md](references/olares-chart-archetypes.md) |

## Routing (this skill vs siblings)

Use this skill to **author/validate your own** Olares chart from a repo, compose, or Helm chart (see the state tables above for where to start). Hand off to a sibling when the task is not authoring:

| User intent | Where |
|---|---|
| Turn a repo / compose / Helm chart into an Olares app, or validate one you authored | ✅ this skill |
| Run the app on **my own** Olares (upload + install) | ✅ this skill — release target **local-run** |
| Publish / list the app on the **public** Olares Market | ✅ this skill — release target **market-distribute** → [market-submit.md](references/olares-chart-market-submit.md) |
| "My chart won't install / the app won't start — why?" | ✅ this skill — diagnosis step of [references/olares-chart-publish-verify.md](references/olares-chart-publish-verify.md) (then [`olares-cluster`](../olares-cluster/SKILL.md) for deeper log digging) |
| "Just install / upgrade an existing catalog app" (not validating your own chart) | [`olares-market`](../olares-market/SKILL.md) |
| "Inspect pods / logs of an unrelated running app" | [`olares-cluster`](../olares-cluster/SKILL.md) |

> **Mental model:** the heart of this skill is **authoring** (produce a valid chart on disk). **Publish-local** — uploading and running the chart on the developer's Olares — proves it works for local-run. **Publish-market** — polishing metadata and opening a PR — is the path to the public catalog. Routine app-store lifecycle on apps you did not author belongs to `olares-market`.

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
 0. target       local-run | market-distribute   # gates P3 arch flags + D2 metadata depth

 Packaging (guided, with the developer; only if an image is missing / wrong-arch):
 P1. docker?    docker version && docker buildx version   # else guide install
 P2. registry   docker login  (Docker Hub)  |  docker login ghcr.io  (ghcr, PAT write:packages)
 P3. build+push local-run:     docker buildx build --platform linux/<node-arch> -t <ref>:<tag> --push <ctx>
                market:        docker buildx build --platform linux/amd64,linux/arm64 -t <ref>:<tag> --push <ctx>
                -> wire <ref>:<tag> into every build-only `image:` in the compose

 Deployment authoring (no login):
 D1. scaffold   olares-cli chart from-compose --name <app> -f docker-compose.yml
 D2. refine     edit OlaresManifest.yaml + templates/ for the 4 refinement areas
                (metadata depth per release target — see manifest.md)
 D3. lint       olares-cli chart lint ./<app>        # loop D2<->D3 until OK
 D4. package    olares-cli chart package ./<app>

 Publish-local (requires login + developer consent; local-run done here):
 V1. logged-in? olares-cli profile list              # if not: tell developer, stop
 V2. ask        confirm with the developer before uploading to a real Olares
 V3. upload     olares-cli market upload ./<app>-<ver>.tgz
 V4. run        olares-cli market install <app> -s upload --version <ver> --watch -o json
 V5. on failure fetch market / chartrepo / app-service / app-pod logs and diagnose
 V6. decide     loop back: chart problem -> D2 ; image problem -> P3 ; uid/EACCES -> run-as-user.md ; else report & ask
 V7. cleanup    olares-cli market uninstall <app> --watch ; olares-cli market delete <app>

 Publish-market (market-distribute only; requires V pass first):
 M1. polish     full metadata, categories, listing images, spec.supportArch — publish-targets checklist
 M2. lint       olares-cli chart lint ./<app>        # re-check after polish
 M3. package    olares-cli chart package ./<app>
 M4. fork       developer forks beclab/apps, adds OAC + owners file
 M5. PR         [NEW][<app>][<ver>] Summary — see market-submit.md
 M6. wait        GitBot validates → auto-merge → app appears in Market
```

Step D1 produces a chart that **already passes `lint`** but is NOT yet a good app: kompose translates containers literally and cannot make product decisions. The value you add is D2. Treat the generated `OlaresManifest.yaml` as a stub — how much you polish §1 Metadata depends on the release target. The V steps cross into sibling skills — full procedure in [references/olares-chart-publish-verify.md](references/olares-chart-publish-verify.md). M steps are for market-distribute only — [references/olares-chart-market-submit.md](references/olares-chart-market-submit.md). Only proceed past D3 upload with the developer's consent.

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

## The four refinement areas (the actual work)

kompose cannot decide these — you must. Full field-by-field mapping and edit recipes are in [references/olares-chart-manifest.md](references/olares-chart-manifest.md).

1. **Metadata** — kompose leaves a stub (`title=name`, default icon, `Utilities` category, no developer info). **Depth depends on release target:** local-run can keep the stub if `lint` passes; market-distribute requires full `metadata.{title,icon,description,categories}` and `spec.{developer,website,sourceCode,submitter,fullDescription}` plus listing images.
2. **Storage** — compose `volumes:` become raw PVCs. Decide each one: app-private state → `.Values.userspace.appData` / `.Values.userspace.appCache` (set `permission.appData/appCache: true`); user-visible files → `.Values.userspace.userData` + list the path under `permission.userData`. Delete the kompose PVCs you replaced and rewrite the `volumeMounts`. Align run identity (uid 1000) with [run-as-user.md](references/olares-chart-run-as-user.md). **Same for both release targets.**
3. **Middleware** — a compose `postgres`/`redis`/`mongo`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` service should usually be dropped and replaced by Olares system middleware: add a `middleware:` block + an `options.dependencies` entry (type `middleware`), delete that workload + its PVC, and repoint the app's env vars at `.Values.<mw>.*`. **Same for both release targets.**
4. **Entrances & ports** — keep/add one `entrances[]` per user-facing HTTP service (tune `host`/`port`/`title`/`authLevel`); expose non-HTTP services via `ports[]` (`exposePort`). Mark internal-only services `invisible: true` or drop their entrance. **Same for both release targets.**

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
| install OK but Permission denied / data not persisted | image uid ≠ 1000 or root-owned dirs on userspace mount | [references/olares-chart-run-as-user.md](references/olares-chart-run-as-user.md) — `spec.runAsUser: true`, securityContext, or initContainer with `beclab/aboveos-busybox:1.37.0` |
| admission denied: untrusted image + root | third-party container runs as root | force uid 1000 or initContainer chown per [run-as-user.md](references/olares-chart-run-as-user.md) |
