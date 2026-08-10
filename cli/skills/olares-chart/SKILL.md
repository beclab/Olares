---
name: olares-chart
version: 4.17.0
description: "Olares app packaging and chart authoring via olares-cli chart — port a repo, docker-compose, or generic Helm chart; build/push the image; author, lint, package, and deploy an OlaresManifest; wire storage, middleware, entrances, env, and GPU; edit the chart after diagnosis. Runtime failure diagnosis is olares-doctor; public Market submission is olares-publish."
compatibility: Requires olares-cli on PATH; chart authoring is local-only, building an image for a specific Olares and deploying both need login
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# Deploy your code or any project to your Olares

Flags and syntax come from `olares-cli chart <verb> --help`. Read the shared [Olares platform model](../olares-shared/references/olares-platform.md) before porting: chart decisions depend on its storage, uid-1000, namespace, middleware and version semantics.

Authoring (`from-compose`, `lint`, `package`) is local-only. Building for a specific Olares needs the target node architecture before the first image build, and deployment needs the active profile.

Porting targets Olares 1.12.6+; load [versioning](references/olares-chart-versioning.md) before writing manifest/chart version and dependency fields.

## When to use

- Turn a repo / docker-compose / generic Helm chart into an Olares app, or validate an OlaresManifest; package its image; wire storage / middleware / entrances / env / GPU
- Deploy / run the app on **your own** Olares (`market upload` + `install`); after `olares-doctor` identifies a chart-owned root cause, edit, lint, and redeploy the chart
- Serve a generation/chat model with an official base app, integrate `llm-init`, or route an embedding app to the appropriate Market install

This skill owns changes to your chart. [`olares-doctor`](../olares-doctor/SKILL.md) finds runtime root causes, [`olares-market`](../olares-market/SKILL.md) manages published apps, and [`olares-publish`](../olares-publish/SKILL.md) prepares a public listing.

## The shape of the work — two axes

Porting an app is **not** a fixed `from-compose → lint → deploy` pipeline — it is driving two **orthogonal but coupled** axes each to its own *ready* state, looping back as constraints surface (an image's baked-in uid/paths constrain the chart's mounts/permissions; a deploy constraint can send you back to rebuild the image). Start wherever your app already stands, not at a fixed step 1. Once both axes are ready, **deploy to the current Olares** — an automatic upload + install + diagnose loop.

- **Packaging — the image:** the app built into a pullable, arch-correct artifact. Olares only pulls, never builds.
- **Deployment — the chart:** a `lint`-passing OlaresManifest + templates. `from-compose` is only **one** way in.

**First move (not a pipeline):** locate where the app already sits on the packaging and deployment state tables → drive the concerns to ready, looping as constraints surface → deploy to your Olares.

## Axis 1 — Packaging (the image)

Olares **pulls images from a registry and never builds from source**, so every workload must reference a publicly pullable, node-arch-correct image. Image work is **agent-driven**: resolve the **target Olares node's architecture** with `olares-cli cluster node list`, then ask which registry the developer uses (Docker Hub / ghcr), check docker is usable and logged in, and **build + push yourself** — only `docker login` stays manual, and only when not already authenticated ([references/olares-chart-image.md](references/olares-chart-image.md)). Build for the target node's arch (single-arch), never the development host's implicit/default arch; multi-arch is only for publishing.

> **The target architecture is a build input, so resolving it cannot be deferred.** `spec.supportArch` says what the chart *claims*; nothing opens the image to check what it *is*. A wrong guess therefore survives build, push and `lint`, and first appears as `exec format error` in the cluster — and on Apple Silicon an unresolved target silently becomes arm64. Query the node or have the developer state the arch; with **neither, ask instead of guessing.** Only a chart nobody is deploying yet needs no target.

| Packaging state | Do this | Ready when |
|---|---|---|
| No Dockerfile (just source) | author a Dockerfile, then build+push | — |
| Dockerfile, but no pullable image | build+push (Docker Hub or ghcr) | — |
| A pullable image exists | check its arch; rebuild if it doesn't match the target Olares node (`olares-cli cluster node list`) | every workload has a pullable, arch-correct image |

## Axis 2 — Deployment (the chart)

The target is a `lint`-passing Olares chart. `from-compose` (kompose) is **just one entry method** — a bare repo, a generic Helm chart, or an already-Olares chart each begin elsewhere (see the state table below). Local authoring (`from-compose` / `lint` / `package`) needs **no login**.

| Deployment state | Do this | Ready when |
|---|---|---|
| Source only (no compose) | author a docker-compose from the code ([compose.md](references/olares-chart-compose.md)) | — |
| A docker-compose | `chart from-compose` then refine ([from-compose.md](references/olares-chart-from-compose.md)) | — |
| A generic Helm chart (no OlaresManifest) | hand-author `OlaresManifest.yaml` + refine (skip `from-compose`) | — |
| Uploaded to the Olares, but no local copy left | `market download <app>` + unpack the `.tgz`, then refine that ([olares-market](../olares-market/SKILL.md), under `charts` → chart management) | — |
| Already an Olares chart | go straight to validation | a chart that passes `chart lint` |

## Deploy to your Olares (the done step)

Both axes ready → **deploy to the current Olares automatically**. `lint` proves the chart is structurally valid; it does **not** prove the app pulls its images, wires its middleware, and reaches `running` — the deploy loop does. **After `lint` passes, proceed without asking:** check login → verify `spec.supportArch` intersects `cluster node list` → package → `market upload` → `market install -s upload --watch` → on failure fetch logs → diagnose → fix chart + re-lint → retry. An architecture upload rejection is terminal until the manifest/image changes; never bump and retry the unchanged package. Only stop to ask when the profile fails olares-shared's [auth-readiness gate](../olares-shared/SKILL.md#auth-readiness-gate) (`invalidated` / `never`) — `logged-in` / `expired` both proceed. Full procedure: [references/olares-chart-deploy.md](references/olares-chart-deploy.md).

For deploying to your own Olares, **metadata can stay a stub** as long as `lint` passes; functional refinement (storage / middleware / entrances) is still required.

## Concern router

`from-compose` produces a skeleton that may lint without being a correct Olares app. Use this index to load only the references triggered by the current port. All 18 original concerns remain represented.

### Every port

| Trigger | Read |
|---|---|
| Build or select every workload image | [image](references/olares-chart-image.md) |
| Decide process uid, then mounted-volume ownership — two independent questions, see below | [run identity](references/olares-chart-run-as-user.md) |
| Map persistence, entrances, metadata and workload replicas | [manifest](references/olares-chart-manifest.md) |
| Map configuration and platform values | [environment](references/olares-chart-env.md), then [defaults](references/olares-chart-env-defaults.md) or [system values](references/olares-chart-system-values.md) only when needed |
| Handle passwords, API keys or generated keys | [secrets](references/olares-chart-secrets.md) |
| Set manifest/chart versions and dependencies | [versioning](references/olares-chart-versioning.md) |
| Validate after a chart change | [lint](references/olares-chart-lint.md) |
| Prove the chart on the target Olares | [deploy](references/olares-chart-deploy.md) |

The manifest reference covers four concerns separately: storage, entrances/ports, workloads/replicas and metadata. Together with image, run identity, env, secrets, versioning, validation and deployment, these are the 11 concerns every port checks.

### Run identity: answer two questions

Q1 (what uid the process ends up as) and Q2 (who owns the directories it writes) are **independent**: every Q1 answer can be paired with either Q2 answer. Answer both, then open the run identity reference for the how.

| Question | Answer | Do |
|---|---|---|
| **Q1** What is the image's effective uid? | 1000 | `spec.runAsUser: true`; on non-primary workloads also set `securityContext.runAsUser: 1000` yourself |
| | 0, and the entrypoint drops via `PUID`/`PGID` | Leave `spec.runAsUser` off, set `PUID=PGID=1000`, verify the final process |
| | 0 and stays root, or any other non-1000 uid | Try `securityContext.runAsUser: 1000`; if the app breaks, rebuild the image |
| **Q2** Does it write a userspace mount? | no | Done |
| | yes | Add a **non-recursive** `init-permissions` initContainer (`beclab/` image, container-level `runAsUser: 0`) — on the `PUID`/`PGID` path read the startup log first, that entrypoint often does it already |

Two red lines: never `chown -R` at runtime, and never set an explicit root `securityContext` — the `beclab/` permissions initContainer is the only exception.

### Conditional

| Trigger | Read |
|---|---|
| Compose bundles a database/queue, or an app dependency is needed | [middleware and dependencies](references/olares-chart-middleware.md) |
| CUDA image, model provisioning or shared model cache | [GPU and models](references/olares-chart-gpu.md) |
| Generation/chat, custom `llm-init`, or embedding serving | [model routing](references/olares-chart-llm-models.md), [model operations](references/olares-chart-llm-ops.md), [custom integration](references/olares-chart-llm-init-integration.md) |
| GPU/accelerator scheduling modes or resource envelope | [accelerator](references/olares-chart-accelerator.md) |
| The app must run Docker or Compose | [DinD](references/olares-chart-dind.md) |
| One heavy backend serves multiple users | [shared backend](references/olares-chart-shared.md) |
| The running app needs a memorable route or custom FQDN | [custom URL](references/olares-chart-custom-domain.md) |

After `lint` passes, drive the deploy/debug loop within the authorised chart task without asking at every install, upgrade, restart, uninstall or clean reinstall. Stop for login, missing registry credentials, an ambiguous target or work outside that task scope. The full assembly sequence is in [workflow](references/olares-chart-workflow.md).

## CLI verbs

The only `olares-cli chart` subcommands (source of truth: `--help`). Everything else above is docker or sibling skills.

| Verb | What it does | Reference |
|---|---|---|
| `from-compose` (alias `init`) | kompose-convert compose file(s) into an Olares chart skeleton | [from-compose.md](references/olares-chart-from-compose.md) |
| `lint` | validate a chart dir / `.tgz` with the Market ingest pipeline | [lint.md](references/olares-chart-lint.md) |
| `package` | package a chart dir into a `<name>-<version>.tgz` for upload (mirrors `helm package`, no helm binary needed) | [workflow.md](references/olares-chart-workflow.md) (D4) |

## Special porting patterns

Most of this skill assumes a web app with an HTTP entrance. When the upstream doesn't fit, match a known pattern first; if still unsure, see how the official ports solved it.

- **Headless CLI / service (no web UI)** — no GUI to point an entrance at: add a web-terminal sidecar as a **visible** entrance + expose the API/MCP port as an `invisible` internal entrance. → [archetype-headless.md](references/olares-chart-archetype-headless.md)
- **GUI desktop app (browser-streamed)** — a native Linux desktop app with no web UI: wrap it in a web-desktop base image (Selkies default, or KasmVNC for old hardware/static UIs), point one visible window entrance at HTTP `:3000`, and device-gate optional iGPU/VAAPI acceleration on `.Values.deviceName`. → [archetype-gui.md](references/olares-chart-archetype-gui.md)
- If no documented pattern fits, inspect a current active port in [beclab/apps](https://github.com/beclab/apps) before guessing:

```bash
gh search code --repo beclab/apps <keyword>      # find charts using a pattern (e.g. type: application, accelerator, appCommon)
# then browse https://github.com/beclab/apps/tree/main/<app> — its OlaresManifest.yaml + templates/
```

Skip references that would mislead:

- **Apps with a `.suspend` (or `.remove`) control file in the OAC root** — suspended / no longer distributed; not a current, reliable pattern.
- **Shared / cluster-scoped charts** that express sharing with `spec.subCharts[].shared: true` + `options.appScope.clusterScoped: true` + `appRef` (the `ollamaserver`/`ollamav2` shape). Copy the shared-app pattern from an `apiVersion: v3` app, not from these. See [shared.md](references/olares-chart-shared.md).

## Gotchas (what `lint` won't catch)

`lint` validates structure, not Olares correctness. Beyond the concerns table above, these blind spots bite and are entirely on you:

- **`metadata.name` must match the chart folder and `Chart.yaml` `name`**, and be `^[a-z][a-z0-9]{0,29}$`. Keep `metadata.appid` equal to `metadata.name` (`from-compose` sets it). Rename all four together. **`lint` does NOT require `metadata.appid`** — a chart lints without it, but **`market upload` rejects a missing `appid`**, so set it explicitly or a lint-clean chart still fails to upload. It does not decide the entrance host: the platform derives that from the app name, so read the real value from the `URL` column of `settings apps list` rather than computing it.
- **Cluster upload requires `spec.supportArch` to intersect at least one current node architecture.** Query `olares-cli cluster node list`; ensure the referenced images support the same target architecture. If upload reports `architecture_incompatible`, fix and repackage before retrying. If it reports `cluster_arch_unavailable`, keep the package/version unchanged and wait for node discovery to recover.
- **Declared `.Values.userspace.appData`/`appCache`/`userData` mounts MUST have the matching `permission` field**, or the app-data cross-check fails.
- **`hostPath` volumes + rolling updates are incompatible** — replace host mounts with the userspace volumes above.
- **The entrance proxy caps every request at `options.apiTimeout` seconds (default 15s)** — long LLM streams / big uploads / slow reports get cut at the entrance (504 / closed connection) even when the pod is healthy. Set `options.apiTimeout: 0` to disable, or a large bounded value; a *negative* value is not "unlimited" (it falls back to 15s). See the Manifest refinement areas.
