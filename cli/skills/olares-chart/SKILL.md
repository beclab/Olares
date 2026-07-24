---
name: olares-chart
version: 4.10.2
description: "Olares app packaging and chart authoring via olares-cli chart — port a repo, docker-compose, or generic Helm chart; build/push the image; author, lint, package, and deploy an OlaresManifest; wire storage, middleware, entrances, env, and GPU; edit the chart after diagnosis. Runtime failure diagnosis is olares-doctor; public Market submission is olares-publish."
compatibility: Requires olares-cli on PATH; chart authoring is local-only, deploy needs login
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# Deploy your code or any project to your Olares

> **Source of truth for flags is always `olares-cli chart <verb> --help`.** This file only carries what `--help` cannot: the two coupled axes of a port, where to start on each, the per-axis concerns to get right, and the gotchas `lint` won't catch.

> **Porting baseline: Olares >= 1.12.6.** Check the target with `olares-cli profile list` (VERSION column). Full version rules — `apiVersion: v3`, the chart version fields, the `olares` `type: system` dependency — are in [references/olares-chart-versioning.md](references/olares-chart-versioning.md).

> **Canonical manifest combination:** new ports use `OlaresManifest.yaml apiVersion: v3` + `olaresManifest.version: 0.12.0` + an `olares` `type: system` dependency at `>=1.12.6-0`; current `from-compose` emits all three. Never downgrade these to satisfy `lint`—check for an old CLI or skill. `Chart.yaml apiVersion: v2` is a separate Helm field and remains `v2`.

> **Platform model (read once, no login needed for authoring).** Porting decisions rely on the Olares storage model, uid-1000 run identity, app/namespace & networking, system middleware, and version model — all defined once in [`../olares-shared/references/olares-platform.md`](../olares-shared/references/olares-platform.md). Packaging an image and authoring/validating the chart need no login; only **deploy to your Olares** (`market upload` + `install`) does.

## When to use

- Turn a repo / docker-compose / generic Helm chart into an Olares app, or validate an OlaresManifest; package its image; wire storage / middleware / entrances / env / GPU
- Deploy / run the app on **your own** Olares (`market upload` + `install`); after `olares-doctor` identifies a chart-owned root cause, edit, lint, and redeploy the chart
- Serve a specific LLM / embedding model (HF or Ollama) with no chart authoring — clone an `llm-init` base app and fill env ([llm-models.md](references/olares-chart-llm-models.md)); capability/context tuning + day-2 ops in [llm-ops.md](references/olares-chart-llm-ops.md)

> Anything outside this scope -> see the **Skill suite map** in [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) (already loaded as the suite prerequisite).

> **Mental model:** this skill authors, validates, and **deploys your own chart to your own Olares**, and owns the **fixes that mean editing your chart**. **Diagnosing a runtime failure** (won't install / start, crashes, `running` but unreachable, image won't pull) is [`../olares-doctor/SKILL.md`](../olares-doctor/SKILL.md) — it finds the root cause and points back here for the chart edit. Listing / selling it on the **public** Olares Market is [`../olares-publish/SKILL.md`](../olares-publish/SKILL.md); installing or managing an **already-published** catalog app is [`../olares-market/SKILL.md`](../olares-market/SKILL.md); inspecting pods/logs of an unrelated running app is [`../olares-cluster/SKILL.md`](../olares-cluster/SKILL.md).

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
| A pullable image exists | check its arch; rebuild if it doesn't match this node (`olares-cli cluster node list`) | every workload has a pullable, arch-correct image |

## Axis 2 — Deployment (the chart)

The target is a `lint`-passing Olares chart. `from-compose` (kompose) is **just one entry method** — a bare repo, a generic Helm chart, or an already-Olares chart each begin elsewhere (see the state table below). Local authoring (`from-compose` / `lint` / `package`) needs **no login**.

| Deployment state | Do this | Ready when |
|---|---|---|
| Source only (no compose) | author a docker-compose from the code ([compose.md](references/olares-chart-compose.md)) | — |
| A docker-compose | `chart from-compose` then refine ([from-compose.md](references/olares-chart-from-compose.md)) | — |
| A generic Helm chart (no OlaresManifest) | hand-author `OlaresManifest.yaml` + refine (skip `from-compose`) | — |
| Already an Olares chart | go straight to validation | a chart that passes `chart lint` |

## Deploy to your Olares (the done step)

Both axes ready → **deploy to the current Olares automatically**. `lint` proves the chart is structurally valid; it does **not** prove the app pulls its images, wires its middleware, and reaches `running` — the deploy loop does. **After `lint` passes, proceed without asking:** check login → package → `market upload` → `market install -s upload --watch` → on failure fetch logs → diagnose → fix chart + re-lint → retry. Only stop to ask when the profile fails olares-shared's [auth-readiness gate](../olares-shared/SKILL.md#auth-readiness-gate) (`invalidated` / `never`) — `logged-in` / `expired` both proceed. Full procedure: [references/olares-chart-deploy.md](references/olares-chart-deploy.md).

For deploying to your own Olares, **metadata can stay a stub** as long as `lint` passes; functional refinement (storage / middleware / entrances) is still required.

## The concerns to get right (reflect on these per axis)

`from-compose` produces a chart that **lints but is not yet a good Olares app**. These are the same concerns viewed two ways — what triggers them, what they let you do, and when they're done — collapsed into one row each. Drive each to "get this right", and loop back here when its trigger reappears.

| Axis | Concern | Get this right | Loop back when | Reference |
|---|---|---|---|---|
| packaging | **Image** | pullable, pinned to a version tag (never `:latest`), arch-correct for **this node** | `ImagePullBackOff` / wrong arch, or a deploy constraint forces a rebuild | [image.md](references/olares-chart-image.md) |
| packaging+deployment | **Run identity** | uid 1000; `spec.runAsUser: true`; initContainer `chown` for root-owned volumes; no root main on non-trusted images (OPA) | EACCES on appData/appCache/userData; admission denies a root third-party image | [run-as-user.md](references/olares-chart-run-as-user.md) |
| deployment | **Storage** | every compose volume → the right userspace area (Data/Cache/Home/Common/External), matching `permission`, leftover kompose PVCs deleted | a volume isn't persisting or lands in the wrong area | [manifest.md](references/olares-chart-manifest.md) §2 |
| deployment | **Middleware & deps** | no bundled `postgres`/`redis`/`mongo`/…; wire to system middleware; SQLite→Postgres where supported; companion apps as `type: application` deps | a bundled db/queue remains, or a companion should be a dependency | [middleware.md](references/olares-chart-middleware.md) |
| deployment | **Env** | app config in `envs[]` (v3 `valueFrom`, no inline `OLARES_USER`); install-time `required` prompts; middleware/system/user vars via `.Values.olaresEnv`; platform context via `.Values.*` | install fails on `appenv` 422, or config must be user-supplied | [env.md](references/olares-chart-env.md), [env-defaults.md](references/olares-chart-env-defaults.md), [system-values.md](references/olares-chart-system-values.md) |
| deployment | **Entrances & ports** | ≥1 `entrances[]`; HTTP via entrances, non-HTTP via `ports[]`; internal-only services `invisible: true` | a service is unreachable, or an internal port is exposed as a desktop entrance | [manifest.md](references/olares-chart-manifest.md) §4 |
| packaging+deployment | **GPU / models** | build a CUDA image without a local GPU; download model weights via initContainer into the shared `appCommon` Hugging Face cache | AI app needs a CUDA build, model provisioning, or a shared model cache | [gpu.md](references/olares-chart-gpu.md) |
| deployment | **LLM model serving** | serve any HF/Ollama model without authoring — pick an engine by format, fill env, clone an `llm-init` base app (llama.cpp / Ollama / vLLM / SGLang); set `MODEL_SUPPORTS` from the model card | user wants to run/serve a specific LLM or embedding model, not author a new app | [llm-models.md](references/olares-chart-llm-models.md) + [llm-ops.md](references/olares-chart-llm-ops.md) |
| deployment | **Accelerator** | **GPU/accelerator apps only:** declare `spec.accelerator` modes per repo support; set `requiredGPUMemory`. A non-accelerator app needs no `mode` — use the flat `spec.requiredCpu/limitedCpu/requiredMemory/limitedMemory/requiredDisk` envelope (mutually exclusive with `spec.accelerator`) | app targets a GPU/accelerator device, or `lint` flags `spec.resources` | [accelerator.md](references/olares-chart-accelerator.md) |
| packaging+deployment | **DinD** | a privileged `beclab/docker` daemon sidecar (`ENABLE_DIND`, `DOCKER_HOST`); main container stays non-privileged | a terminal/agent app must run `docker` / `docker compose` | [dind.md](references/olares-chart-dind.md) |
| deployment | **Shared backend** | `options.shared: true` (on a v3 app) ⇒ admin-only install into `<app>-shared`; consumers reach it via cross-namespace Service DNS; flag the admin-install | a heavy/accelerator backend serves many users over shared data | [shared.md](references/olares-chart-shared.md) |
| deployment | **Version & deps fields** | fixed values every chart writes: `apiVersion: v3`; `olaresManifest.version: '0.12.0'`; `metadata.version` == `Chart.yaml` `version`; `options.dependencies` includes `olares >=1.12.6-0` (`type: system`) — authored by you; **bump `metadata.version` (= `Chart.yaml` `version`) on every (re)upload** | every manifest edit — confirm these fields + the `olares` system dep (`lint` rejects a missing system dep); every (re)upload — bump the version first | [versioning.md](references/olares-chart-versioning.md) |
| deployment | **Workloads / replicas** | `workloadReplicas` lists every Deployment/StatefulSet → count (authored by you); each `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}` + matching `values.yaml` | every time you author/add/rename a workload — run the three-point self-check; a hardcoded `replicas` silently no-ops suspend/resume / staged install | [manifest.md](references/olares-chart-manifest.md) Workloads & replicas |
| deployment | **Metadata** | stub OK for local deploy (`Utilities`, default icon) while `lint` passes; full `metadata.*` + listing images only when publishing | `lint` flags missing metadata, or you want a public listing | [manifest.md](references/olares-chart-manifest.md) §1 |
| deployment | **Validate-local** | `olares-cli chart lint ./<app>` passes, then `chart package` | a refinement changed the manifest/templates | [lint.md](references/olares-chart-lint.md) |
| deploy | **Deploy** | `market upload` + `market install`, then diagnose from logs — automatic after `lint` passes (login required) | proving the chart actually runs on the developer's Olares | [deploy.md](references/olares-chart-deploy.md) |

> **Deploy** leans on sibling skills: [`olares-shared`](../olares-shared/SKILL.md) (login check), [`olares-market`](../olares-market/SKILL.md) (upload / install / cleanup), [`olares-cluster`](../olares-cluster/SKILL.md) (logs), and [`olares-doctor`](../olares-doctor/SKILL.md) (runtime root-cause diagnosis when a deploy fails or the app won't start). **After `lint` passes, drive the deploy loop automatically without asking** (proceed unless olares-shared's auth-readiness gate says stop); **never log in on the developer's behalf without asking.** One way to sequence the whole assembly (and the file tree `from-compose` emits) lives in [references/olares-chart-workflow.md](references/olares-chart-workflow.md) — a reference, not a required order.

## CLI verbs

The only `olares-cli chart` subcommands (source of truth: `--help`). Everything else above is docker or sibling skills.

| Verb | What it does | Reference |
|---|---|---|
| `from-compose` (alias `init`) | kompose-convert compose file(s) into an Olares chart skeleton | [from-compose.md](references/olares-chart-from-compose.md) |
| `lint` | validate a chart dir / `.tgz` with the Market ingest pipeline | [lint.md](references/olares-chart-lint.md) |
| `package` | package a chart dir into a `<name>-<version>.tgz` for upload (mirrors `helm package`, no helm binary needed) | [workflow.md](references/olares-chart-workflow.md) (D4) |

## Special porting patterns & official ports (beclab/apps)

Most of this skill assumes a web app with an HTTP entrance. When the upstream doesn't fit, match a known pattern first; if still unsure, see how the official ports solved it.

- **Headless CLI / service (no web UI)** — no GUI to point an entrance at: add a web-terminal sidecar as a **visible** entrance + expose the API/MCP port as an `invisible` internal entrance. → [archetype-headless.md](references/olares-chart-archetype-headless.md)
- **GUI desktop app (browser-streamed)** — a native Linux desktop app with no web UI: wrap it in a web-desktop base image (Selkies default, or KasmVNC for old hardware/static UIs), point one visible window entrance at HTTP `:3000`, and device-gate optional iGPU/VAAPI acceleration on `.Values.deviceName`. → [archetype-gui.md](references/olares-chart-archetype-gui.md)
- **Still no idea?** look at how the official ports wire it (manifest fields, entrance shapes, middleware/app dependencies, GPU specs) in [beclab/apps](https://github.com/beclab/apps) before guessing:

```bash
gh search code --repo beclab/apps <keyword>      # find charts using a pattern (e.g. type: application, accelerator, appCommon)
# then browse https://github.com/beclab/apps/tree/main/<app> — its OlaresManifest.yaml + templates/
```

This is the canonical source for cross-app wiring this skill intentionally does not hardcode.

**When picking a reference app, skip these — they mislead more than they help:**

- **Apps with a `.suspend` (or `.remove`) control file in the OAC root** — suspended / no longer distributed; not a current, reliable pattern.
- **Shared / cluster-scoped charts** that express sharing with `spec.subCharts[].shared: true` + `options.appScope.clusterScoped: true` + `appRef` (the `ollamaserver`/`ollamav2` shape). Copy the shared-app pattern from an `apiVersion: v3` app, not from these. See [shared.md](references/olares-chart-shared.md).

## Gotchas (what `lint` won't catch)

`lint` validates structure, not Olares correctness. Beyond the concerns table above, these blind spots bite and are entirely on you:

- **`metadata.name` must match the chart folder and `Chart.yaml` `name`**, and be `^[a-z][a-z0-9]{0,29}$`. Keep `metadata.appid` equal to `metadata.name` (`from-compose` sets it; it backs the entrance domain `<appid>.<zone>`). Rename all four together. **`lint` does NOT require `metadata.appid`** — an empty value passes (the loader normalizes it to `md5(metadata.name)[:8]`), but **`market upload` rejects a missing `appid`**, so set it explicitly or a lint-clean chart still fails to upload.
- **Declared `.Values.userspace.appData`/`appCache`/`userData` mounts MUST have the matching `permission` field**, or the app-data cross-check fails.
- **`hostPath` volumes + rolling updates are incompatible** — replace host mounts with the userspace volumes above.
- **The entrance proxy caps every request at `options.apiTimeout` seconds (default 15s)** — long LLM streams / big uploads / slow reports get cut at the entrance (504 / closed connection) even when the pod is healthy. Set `options.apiTimeout: 0` to disable, or a large bounded value; a *negative* value is not "unlimited" (it falls back to 15s). See the Manifest refinement areas.
