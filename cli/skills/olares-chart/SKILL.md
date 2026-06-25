---
name: olares-chart
version: 4.7.0
description: "Help a developer turn their own code or any open-source project into an app that runs on their own Olares. Two coupled axes: packaging the container image and authoring/refining the Olares app chart (OlaresManifest), then deploying it to the current Olares with an automatic upload + install + diagnose loop. Use when deploying a repo, docker-compose, or Helm chart to Olares, packaging an Olares app, wiring storage / system middleware / entrances / env / GPU, or fixing a failed install (ImagePullBackOff, permission denied / EACCES, app won't start). Publishing to the public Olares Market is the olares-publish skill."
compatibility: Requires olares-cli on PATH; chart authoring is local-only, deploy needs login
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# Deploy your code or any project to your Olares

> **Source of truth for flags is always `olares-cli chart <verb> --help`.** This file only carries what `--help` cannot: the two coupled axes of a port, where to start on each, the per-axis concerns to get right, and the gotchas `lint` won't catch.

> **Porting baseline: Olares >= 1.12.6.** This skill targets Olares >= 1.12.6 (other `olares-cli` features have no such floor). Check the target with `olares-cli profile list` (VERSION column). Full version rules — `apiVersion: v3`, the several chart version fields, `type: system` dependency — are in [references/olares-chart-versioning.md](references/olares-chart-versioning.md).

> **Platform model (read once, no login needed for authoring).** The porting decisions below rely on the Olares storage model, uid-1000 run identity, app/namespace & networking, system middleware, and version model — all defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md). **Packaging an image and authoring/validating the chart need no profile login.** Only the final **deploy to your Olares** (`market upload` + `install`) requires login.

## When to use

- Turn a repo / docker-compose / generic Helm chart into an Olares app, or validate an OlaresManifest you authored
- Package an Olares app image; wire storage / system middleware / entrances / env / GPU
- Deploy / run the app on **your own** Olares (`market upload` + `install`)
- Serve a specific LLM / embedding model (HF or Ollama) with no chart authoring — clone an `llm-init` base app and fill env ([llm-models.md](references/olares-chart-llm-models.md))
- Fix a failed install (`ImagePullBackOff`, permission denied / `EACCES`, app won't start)

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** this skill authors, validates, and **deploys your own chart to your own Olares**. Listing / selling it on the **public** Olares Market is [`../olares-publish/SKILL.md`](../olares-publish/SKILL.md); installing or managing an **already-published** catalog app is [`../olares-market/SKILL.md`](../olares-market/SKILL.md); inspecting pods/logs of an unrelated running app is [`../olares-cluster/SKILL.md`](../olares-cluster/SKILL.md).

## The shape of the work — two axes

Porting an app is **not** a fixed `from-compose → lint → deploy` pipeline — it is driving two **orthogonal but coupled** axes each to its own *ready* state, looping back across the coupling edges as constraints surface (an image's baked-in uid/paths constrain the chart's mounts/permissions; a deploy constraint can send you back to rebuild the image). Start wherever your app already stands, not at a fixed step 1. Once both axes are ready, **deploy to the current Olares** — an automatic upload + install + diagnose loop.

- **Packaging — the image:** the app built into a pullable, arch-correct artifact. Olares only pulls, never builds.
- **Deployment — the chart:** a `lint`-passing OlaresManifest + templates. `from-compose` is only **one** way in.

**First move (not a pipeline):** locate where the app already sits on the packaging and deployment state tables → drive the concerns to ready, looping as constraints surface → deploy to your Olares.

> **Publishing to the public Olares Market is out of scope here.** Once the app runs on your Olares, listing it publicly (market-ready metadata, multi-arch, the `beclab/apps` PR, paid apps) is the [`../olares-publish/SKILL.md`](../olares-publish/SKILL.md) skill.

## Axis 1 — Packaging (the image)

Olares **pulls images from a registry and never builds from source**, so every workload must reference a publicly pullable, node-arch-correct image. Image work is **agent-driven**: ask which registry the developer uses (Docker Hub / ghcr), check docker is usable and already logged in, then **build + push yourself** — only the credential entry (`docker login`) stays manual, and only when not already authenticated ([references/olares-chart-image.md](references/olares-chart-image.md)). No Olares login needed for this. Build for **this node's arch** (single-arch); multi-arch is only needed when publishing to the public Market (olares-publish).

| Packaging state | Do this | Ready when |
|---|---|---|
| No Dockerfile (just source) | author a Dockerfile, then build+push | — |
| Dockerfile, but no pullable image | build+push (Docker Hub or ghcr) | — |
| A pullable image exists | check its arch; rebuild if it doesn't match this node (`olares-cli cluster node list`) | every workload has a pullable, arch-correct image |

## Axis 2 — Deployment (the chart)

The target is a `lint`-passing Olares chart. `from-compose` (kompose) is **just one entry method** for a compose-based start — a bare repo, a generic Helm chart, or an already-Olares chart each begin elsewhere. Local authoring (`from-compose` / `lint` / `package`) needs **no login**.

| Deployment state | Do this | Ready when |
|---|---|---|
| Source only (no compose) | author a docker-compose from the code ([compose.md](references/olares-chart-compose.md)) | — |
| A docker-compose | `chart from-compose` then refine ([from-compose.md](references/olares-chart-from-compose.md)) | — |
| A generic Helm chart (no OlaresManifest) | hand-author `OlaresManifest.yaml` + refine (skip `from-compose`) | — |
| Already an Olares chart | go straight to validation | a chart that passes `chart lint` |

## Deploy to your Olares (the done step)

Both axes ready → **deploy to the current Olares automatically**. `lint` proves the chart is structurally valid; it does **not** prove the app pulls its images, wires its middleware, and reaches `running`. The deploy loop does — by pushing the chart to the developer's Olares and watching it install. **After `lint` passes, proceed without asking:** check login → package → `market upload` → `market install -s upload --watch` → on failure fetch logs → diagnose → fix chart + re-lint → retry. Only stop to ask when the profile is not logged in, or when a failure is clearly not a chart problem. Full procedure: [references/olares-chart-deploy.md](references/olares-chart-deploy.md).

For deploying to your own Olares, **metadata can stay a stub** (`Utilities` category, default icon, empty `spec.developer`/`fullDescription`/…) as long as `lint` passes — functional refinement (storage / middleware / entrances) is still required.

## The concerns to get right (reflect on these per axis)

`from-compose` produces a chart that **lints but is not yet a good Olares app**. These are the same concerns viewed two ways — what triggers them, what they let you do, and when they're done — collapsed into one row each. Drive each to "get this right", and loop back here when its trigger reappears.

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
| deployment | **Workloads / replicas** | `workloadReplicas` lists every Deployment/StatefulSet → count (required on 0.12.0); each workload's `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}` + matching `values.yaml` | `lint` rejects a missing map, or suspend/resume / staged install silently no-op on a hardcoded `replicas` | [manifest.md](references/olares-chart-manifest.md) Workloads & replicas |
| deployment | **Metadata** | stub OK for local deploy (`Utilities`, default icon) as long as `lint` passes; full `metadata.*` + listing images only when publishing to the Market | `lint` flags missing metadata, or you want a public listing | [manifest.md](references/olares-chart-manifest.md) §1 |
| deployment | **Validate-local** | `olares-cli chart lint ./<app>` passes, then `chart package` | a refinement changed the manifest/templates | [lint.md](references/olares-chart-lint.md) |
| deploy | **Deploy** | `market upload` + `market install`, then diagnose from logs — automatic after `lint` passes (login required) | proving the chart actually runs on the developer's Olares | [deploy.md](references/olares-chart-deploy.md) |

> **Deploy** leans on sibling skills: [`olares-shared`](../olares-shared/SKILL.md) (login check), [`olares-market`](../olares-market/SKILL.md) (upload / install / cleanup), [`olares-cluster`](../olares-cluster/SKILL.md) (logs). **After `lint` passes, drive the deploy loop automatically without asking; never log in on the developer's behalf without asking.** One way to sequence the whole assembly (and the file tree `from-compose` emits) lives in [references/olares-chart-workflow.md](references/olares-chart-workflow.md) — a reference, not a required order.

## CLI verbs

The only `olares-cli chart` subcommands (source of truth: `--help`). Everything else above is docker or sibling skills.

| Verb | What it does | Reference |
|---|---|---|
| `from-compose` (alias `init`) | kompose-convert compose file(s) into an Olares chart skeleton | [from-compose.md](references/olares-chart-from-compose.md) |
| `lint` | validate a chart dir / `.tgz` with the Market ingest pipeline | [lint.md](references/olares-chart-lint.md) |
| `package` | package a chart dir into a `<name>-<version>.tgz` for upload (mirrors `helm package`, no helm binary needed) | [workflow.md](references/olares-chart-workflow.md) (D4) |

## Special porting patterns & official ports (beclab/apps)

Most of this skill assumes a web app with an HTTP entrance. When the upstream doesn't fit, match a known pattern first; if still unsure, see how the official ports solved it.

- **Headless CLI / service (no web UI)** — no GUI to point an entrance at: add a web-terminal sidecar as a **visible** entrance + expose the API/MCP port as an `invisible` internal entrance. → [archetypes.md](references/olares-chart-archetypes.md)
- **GUI desktop app (browser-streamed)** — a native Linux desktop app with no web UI: wrap it in a web-desktop base image (Selkies default, or KasmVNC for old hardware/static UIs), point one visible window entrance at HTTP `:3000`, and device-gate optional iGPU/VAAPI acceleration on `.Values.deviceName`. → [archetypes.md](references/olares-chart-archetypes.md)
- **Still no idea?** look at how the official ports wire it (manifest fields, entrance shapes, middleware/app dependencies, GPU specs) in [beclab/apps](https://github.com/beclab/apps) before guessing:

```bash
gh search code --repo beclab/apps <keyword>      # find charts using a pattern (e.g. type: application, accelerator, appCommon)
# then browse https://github.com/beclab/apps/tree/main/<app> — its OlaresManifest.yaml + templates/
```

This is the canonical source for cross-app wiring (how a dependency app is reached, real `entrances`/`ports`, accelerator modes) that this skill intentionally does not hardcode.

**When picking a reference app, skip these — they mislead more than they help:**

- **Apps with a `.suspend` (or `.remove`) control file in the OAC root** — suspended / no longer distributed; not a current, reliable pattern.
- **Legacy v2 shared / cluster-scoped charts** that express sharing with `spec.subCharts[].shared: true` + `options.appScope.clusterScoped: true` + `appRef` (the `ollamaserver`/`ollamav2` shape). For Olares >= 1.12.6 this is superseded by `apiVersion: v3`; copy the shared-app pattern from a current `v3` app, not from these. See [shared.md](references/olares-chart-shared.md).

## Gotchas (what `lint` won't catch)

`lint` validates structure, not Olares correctness. Beyond the concerns table above, these blind spots bite and are entirely on you:

- **`metadata.name` must match the chart folder and `Chart.yaml` `name`**, and be `^[a-z][a-z0-9]{0,29}$`. Keep `metadata.appid` equal to `metadata.name` (`from-compose` sets it; it backs the entrance domain `<appid>.<zone>`). Rename all four together.
- **Declared `.Values.userspace.appData`/`appCache`/`userData` mounts MUST have the matching `permission` field**, or the app-data cross-check fails.
- **`hostPath` volumes + rolling updates are incompatible** — replace host mounts with the userspace volumes above.
