---
name: olares-chart
version: 1.17.0
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

> **Porting baseline: Olares >= 1.12.6.** This skill targets Olares >= 1.12.6 (other `olares-cli` features have no such floor). Check the target with `olares-cli profile list` (VERSION column). Full version rules — `apiVersion: v3`, the several chart version fields, `type: system` dependency — are in [references/olares-chart-versioning.md](references/olares-chart-versioning.md).

## When to use

- Olares chart, OlaresManifest / OlaresManifest.yaml, olares-cli chart, chart from-compose, chart lint, chart package
- Turn docker-compose, a generic Helm chart, or a bare source repo into a **lint-passing** Olares app chart — or a **market-ready** one when distributing publicly
- **Local run** on your own Olares (upload + install); **market distribute** (full metadata, multi-arch, PR to `beclab/apps`)
- Building/pushing docker image (amd64 vs arm64), no official image, wrong arch
- GPU / CUDA app: building a CUDA image without a GPU on the build machine, `TORCH_CUDA_ARCH_LIST`, nvidia mode = amd64; model download / Hugging Face weights / shared model cache via `appCommon` (`drive/Common`)
- Accelerator compute resources: declare `spec.accelerator` modes (nvidia/amd-gpu/amd-apu/strix-halo/nvidia-gb10/apple-m/cpu, per what the repo supports), GPU memory via `requiredGPUMemory`, and how much CPU/memory/GPU to request for a ported project
- Install/runtime failures: ImagePullBackOff, app failed to install or start, market / app-service / chartrepo logs
- **Permission denied / EACCES** on userspace volumes, third-party image runs as root or non-1000 uid, `spec.runAsUser`, initContainer volume `chown`
- Headless / CLI app, MCP server, or tool with no web UI (terminal entrance + invisible entrance)
- Run `docker` / `docker compose` from inside a terminal/agent app: privileged Docker-in-Docker sidecar gated by `ENABLE_DIND`, `DOCKER_HOST=tcp://localhost:2375`, trusted `beclab/docker` daemon image
- Depend on another already-ported app instead of bundling it (`options.dependencies` `type: application`, e.g. searxng companion service)
- Shared app: a heavy/accelerator backend with its own accounts, used by many people, sharing one data set — installed **once by an admin** cluster-wide via `apiVersion: v3` (admin-only, `<app>-shared` namespace), consumed by reference apps over cross-namespace Service DNS (e.g. ollama / vLLM / LLM gateway)
- Environment variables: declare app config in `envs[]`, prompt the user at install (e.g. init admin username/password via `required`), map system/user vars (`OLARES_SYSTEM_*` / `OLARES_USER_*`) through `valueFrom`, `.Values.olaresEnv`, env `type`/`regex`/`options` validation
- Version rules: Olares system version (semver), `apiVersion: v3`, `olaresManifest.version` 0.8.0 vs 0.12.0, `metadata.version` vs `spec.versionName`, `type: system` dependency, checking the target with `profile list`
- Three axes: **packaging** (Dockerfile/image), **deployment** (compose/chart), **publishing** (release target); four post-kompose refinement areas (metadata depth gated by target, storage, middleware, entrances)
- Optional live validation (requires login): package + market upload/install — see [`olares-shared`](../olares-shared/SKILL.md), [`olares-market`](../olares-market/SKILL.md), [`olares-cluster`](../olares-cluster/SKILL.md)

## Start here: establish your release target

Before the packaging/deployment state tables, decide **who consumes the chart** and what "done" means. Infer from user language; ask if ambiguous. Full decision tree and checklists: [references/olares-chart-publish-targets.md](references/olares-chart-publish-targets.md).

| Release target | User signals | Done when |
|---|---|---|
| **local-run** (default for most users) | "run on my Olares", "upload and install", "just for myself" | `lint` OK → package → upload + install reaches `running` on the developer's Olares |
| **market-distribute** | "publish to Market", "submit to beclab/apps", "上架" | local validation passes **plus** market-ready metadata/images/arch → PR merged into `beclab/apps:main` |

**What differs by target** (full matrix in [publish-targets.md](references/olares-chart-publish-targets.md)): **image arch** (local-run = single-arch for this node via `cluster node list`; market = multi-arch + `spec.supportArch`); **metadata depth** (local = stub OK if `lint` passes; market = full metadata, dual-version categories, listing images, developer links); and the **publish path** (local = upload + install; market = the same validation first, then a `beclab/apps` PR). **Refine §2–4 (storage / middleware / entrances) is functional and identical for both.**

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
| **Env** | deployment | declare app config (`envs[]`), prompt the user at install, map system/user vars via `valueFrom`, `type`/`regex` validation | no | install fails on `appenv` 422 (missing/invalid env), or config must be user-supplied | [env.md](references/olares-chart-env.md) |
| **Run-as-user** | packaging + deployment | align image uid with Olares userspace (1000): Dockerfile `USER`, `spec.runAsUser`, initContainer `chown` | no | EACCES on appData/appCache/userData, OPA root deny on third-party image | [run-as-user.md](references/olares-chart-run-as-user.md) |
| **GPU / models** | packaging + deployment | build a CUDA image without a local GPU; download model weights via initContainer into the shared `appCommon` Hugging Face cache | no | AI app needs CUDA build or model provisioning, custom-kernel arch flags, shared model cache | [gpu.md](references/olares-chart-gpu.md) |
| **DinD** | packaging + deployment | give a terminal/agent app a `docker` CLI via a privileged `beclab/docker` daemon sidecar (`ENABLE_DIND`, `DOCKER_HOST`) while keeping the main container non-privileged | no | app must run `docker` / `docker compose` / containerized dev stacks from the terminal | [dind.md](references/olares-chart-dind.md) |
| **Shared** | deployment | make the app an admin-installed, cluster-wide shared backend (`apiVersion: v3` ⇒ `<app>-shared` namespace, admin-only, cross-namespace Service DNS for consumers) | no | heavy/accelerator backend with its own accounts, multi-user, shared data; consumed by separate reference apps | [shared.md](references/olares-chart-shared.md) |
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
| `package` | package a chart dir into a `<name>-<version>.tgz` for upload (mirrors `helm package`, no helm binary needed) | [references/olares-chart-workflow.md](references/olares-chart-workflow.md) (D4 / M3) |

## Typical assembly & conversion output

One way to compose the capabilities (packaging → deployment authoring → publish-local → publish-market), plus the file tree `from-compose` emits, lives in [references/olares-chart-workflow.md](references/olares-chart-workflow.md). It is **not** a fixed pipeline — start wherever your state tables put you, and loop across the coupling edges as failures surface. Step D1 scaffolds a chart that **already passes `lint`** but is not yet a good app; the value you add is the refinement below.

## The four refinement areas (the actual work)

kompose cannot decide these — you must. Full field-by-field mapping and edit recipes are in [references/olares-chart-manifest.md](references/olares-chart-manifest.md).

1. **Metadata** — kompose leaves a stub (`title=name`, default icon, `Utilities` category, no developer info). **Depth depends on release target:** local-run can keep the stub if `lint` passes; market-distribute requires full `metadata.{title,icon,description,categories}` and `spec.{developer,website,sourceCode,submitter,fullDescription}` plus listing images.
2. **Storage** — compose `volumes:` become raw PVCs. Decide each one: app-private state → `.Values.userspace.appData` / `.Values.userspace.appCache` (set `permission.appData/appCache: true`); user-visible files → `.Values.userspace.userData` + list the path under `permission.userData`. Delete the kompose PVCs you replaced and rewrite the `volumeMounts`. Align run identity (uid 1000) with [run-as-user.md](references/olares-chart-run-as-user.md). **Same for both release targets.**
3. **Middleware & dependencies** — drop any bundled `postgres`/`redis`/`mongo`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats` workload + its PVC and wire to Olares **system middleware** (`middleware:` block + an `options.dependencies` `type: middleware` entry, env repointed at `.Values.<mw>.*`); `lint` won't flag a bundled db, so it's on you. If the upstream defaults to **SQLite**, switch to Postgres where supported. If it bundles a **companion app** already in the Market (e.g. searxng), depend on it via `options.dependencies` `type: application` instead. Full rules, Postgres extension catalog, and the escape hatch: [manifest.md §3](references/olares-chart-manifest.md). **Same for both release targets.**
4. **Entrances & ports** — keep/add one `entrances[]` per user-facing HTTP service (tune `host`/`port`/`title`/`authLevel`); expose non-HTTP services via `ports[]` (`exposePort`). Mark internal-only services `invisible: true` or drop their entrance. **Same for both release targets.**

> **Env wiring (cross-cutting):** any config the app needs — user-supplied at install (admin credentials via `required`), reused Olares system/user vars (`valueFrom`), middleware connection strings — is declared in `envs[]` and surfaced as `.Values.olaresEnv.<name>`, then mapped into the workload's `env:`. Full rules (system/user/app levels, `required` vs optional, `type`/`regex` validation, default variable lists): [references/olares-chart-env.md](references/olares-chart-env.md).

## Reference official ports (beclab/apps)

When you need inspiration or are unsure how Olares expects something wired (manifest fields, entrance shapes, middleware/app dependencies, GPU specs), look at how the official ports do it in [beclab/apps](https://github.com/beclab/apps) before guessing:

```bash
gh search code --repo beclab/apps <keyword>      # find charts using a pattern (e.g. type: application, accelerator, appCommon)
# then browse https://github.com/beclab/apps/tree/main/<app> — its OlaresManifest.yaml + templates/
```

This is the canonical source for cross-app wiring (how a dependency app is reached, real `entrances`/`ports`, accelerator modes) that this skill intentionally does not hardcode.

## Hard constraints that bite

- **Pin every `image:` to a specific version tag** (e.g. `nginx:1.27`, `<user>/<repo>:1.2.3`). **Never use `:latest`** or an untagged image (implicit `latest`): it drifts, making installs non-reproducible and rollbacks/caching unreliable — and Olares pulls images, it never rebuilds. `lint` does **not** catch this; it is the author's responsibility.
- **Set `apiVersion: v3` in `OlaresManifest.yaml`** (skill rule for 1.12.6+ ports). `from-compose` omits it (implicit `v1`), so hand-add it after scaffolding. v3 enables the declarative env rules (`valueFrom`, no inline `OLARES_USER`); `lint` allows `v1`/`v2`/`v3` and does **not** force it. On Olares >= 1.12.6 the install handler treats **v3 as an admin-installed, cluster-wide shared app** (`<app>-shared` namespace, admin-only) — so flag the admin-install requirement to the user, and for a multi-user shared backend follow [shared.md](references/olares-chart-shared.md). Details: [versioning.md](references/olares-chart-versioning.md).
- **`metadata.name` must match the chart folder name and `Chart.yaml` `name`**, and be `^[a-z][a-z0-9]{0,29}$`. `from-compose --name` keeps them consistent; if you rename the folder, fix all three.
- **At least one entrance is required.** Never delete the last `entrances[]` entry.
- **If a template uses `.Values.userspace.appData`/`appCache`/`userData`, the matching `permission` field MUST be declared**, or `lint` fails the app-data cross-check.
- **`hostPath` volumes + rolling updates are incompatible** — `lint` rejects them. Replace host mounts with the userspace volumes above.
- **A Docker-in-Docker daemon sidecar must use a trusted `beclab/docker` image and be the only `privileged: true` container**, paired with `strategy: Recreate` (it relies on `hostPath`); the main container stays non-privileged. Non-trusted privileged images are denied by OPA. Details: [dind.md](references/olares-chart-dind.md).
- **`metadata.version` (Chart Version) and `Chart.yaml` `version` must match**, and `spec.versionName` should track the upstream app version (`Chart.yaml` `appVersion`).

## Common errors → where to fix

- **`lint` failures** (workload not named after the app, app-data/permission mismatch, version mismatch, hostPath + rolling update, namespace, missing resource limits): full failure→fix table in [lint.md](references/olares-chart-lint.md).
- **Install OK but Permission denied / data not persisted**, or **admission denied (untrusted image + root)**: image uid ≠ 1000 — [run-as-user.md](references/olares-chart-run-as-user.md).
- **`ImagePullBackOff` / wrong arch**: [image.md](references/olares-chart-image.md).
- **Install/start failure needing logs**: diagnosis step in [publish-verify.md](references/olares-chart-publish-verify.md).
