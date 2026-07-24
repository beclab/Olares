# LLM Base Skill Design

## Goal

Make the `olares-chart` LLM guidance concise and executable by separating the official Market workflow from custom `llm-init` chart integration.

## Verified Product State

- Olares `>=1.12.6-0` publishes four generation engine templates in `market.olares`: `llamacppllmbasev3`, `ollamallmbasev3`, `vllmllmbasev3`, and `sglangllmbasev3`.
- The four templates are `templateOnly`, multi-instance, admin-only, shared, and support only `nvidia` and `nvidia-gb10`.
- Their manifests already set `options.apiTimeout: 0`.
- `embeddinggemmav3` is the dedicated EmbeddingGemma application. It is admin-only and shared, supports `cpu`, `intel`, `nvidia`, and `nvidia-gb10`, and has no arbitrary model-selection environment variables.
- The published engine templates still expose embedding compatibility options, but this skill will not recommend that route.
- Raw `llm-init` accepts optional `MODEL_SUPPORTS` values as `supports_*` keys. The published charts expose coarse `vision`, `tools`, `thinking`, and `none` groups and expand them before invoking `llm-init`.

## Information Architecture

### `cli/skills/olares-chart/SKILL.md`

Act only as a router:

- Generation model serving points to `olares-chart-llm-models.md`.
- Embedding points to the official `embeddinggemmav3` application and the `olares-market` lifecycle.
- Custom charts that embed `llm-init` point to `olares-chart-llm-init-integration.md`.

The parent skill must not duplicate engine selection, environment tables, or model sizing guidance.

### `references/olares-chart-llm-models.md`

Own the official generation workflow:

1. Confirm a logged-in admin profile and Olares `>=1.12.6-0`.
2. Inspect the model format and the real accelerator capacity.
3. Select one of the four generation templates.
4. Populate only the environment variables exposed by the published chart.
5. Clone directly from `market.olares` with an explicit `--compute-mode`.
6. Watch the new instance and inspect its workload in parallel.

This file must not contain local package/upload instructions, CPU fallbacks, arbitrary embedding guidance, raw download-only configuration, or unsupported mmproj automation.

### `references/olares-chart-llm-ops.md`

Own post-clone operation only:

- Capability-group selection for the published charts.
- Context-window and GPU-memory tuning.
- Model switching through the Market lifecycle.
- Model card updates through the supported `llm-init` API.
- Runtime symptoms and links to `olares-doctor`.

It must not document custom sidecar integration or model provisioning features that are absent from the published install schema.

### `references/olares-chart-llm-init-integration.md`

Own the narrow custom-chart contract:

- Distinguish chart-level convenience values from raw `llm-init` environment values.
- Document the stable model source forms, HF mirror/token behavior, shared cache, readiness sentinel, and download-only mode.
- Treat indexed multi-source configuration as a chart-author feature.
- Link to the `llm-init` repository for complete environment and API references instead of copying them.
- Do not document `embed` or `clipembed`; embedding remains routed to the dedicated application.
- Do not promise automatic mmproj sentinel or wrapper wiring until the implementation and tests provide it.

## Canonical User Behavior

### Generation

- Use only the four official templates.
- Use `MODEL_MODE=chat`.
- Use the published chart groups `vision`, `tools`, `thinking`, or `none` for `MODEL_SUPPORTS`.
- Use llama.cpp or Ollama for GGUF and vLLM or SGLang for full safetensors repositories.
- Run `market clone` without `-s upload`; the default source is `market.olares`.
- Pass `--compute-mode nvidia` or `--compute-mode nvidia-gb10`.
- Do not repackage a template to change `apiTimeout`; the official manifests already disable the timeout.
- If the node has neither supported mode, report that the official generation templates do not apply instead of inventing a CPU path.

### Embedding

- Route to `market install embeddinggemmav3`.
- Do not present the four generation templates as the embedding solution.
- Do not offer arbitrary embedding model environment variables because the published application does not expose them.
- Preserve the application constraints: admin-only, shared, and explicit accelerator selection when required.

### Custom `llm-init` Integration

- Raw `MODEL_SUPPORTS` is optional and accepts only validated `supports_*` keys; unknown keys fail at startup.
- An empty `ENGINE_KIND` is download-only and exposes no `/v1/*` data plane.
- Download-only supports `hf://` and `https://` sources, not `ollama://`.
- Official generation templates hard-wire their engine and do not expose `ENGINE_KIND`; download-only therefore requires a custom chart.
- Source indexing and cache sharing are chart-author concerns, not clone-time options for the official templates.

## CLI Help Corrections

Update `market clone` and `market get` help to match implementation:

- A clone source is a catalog application and does not need to be installed.
- Cloneability is computed from `allowMultipleInstall` or `templateOnly`; the raw JSON payload is not guaranteed to contain a top-level `cloneable` field.
- The operation result field is `targetApp`, not `cloneTarget`.
- The table output from `market get` is the supported human-readable cloneability check.

## Failure Handling

- Missing `--compute-mode` in non-interactive use should point to the supported modes returned by `computeModeSelect`.
- A non-admin request should stop with the admin-only/shared constraint instead of suggesting a namespace workaround.
- A GGUF/safetensors mismatch should route back to engine selection.
- A requested CPU generation deployment should report unsupported official-template hardware.
- A custom download-only request should route to the integration reference.
- A VLM request must not rely on undocumented automatic mmproj wiring.

## Verification

Apply documentation TDD:

1. Run fresh-context baseline scenarios against the current skill and record the wrong decisions.
2. Micro-test the generation, embedding, and custom-integration recipes against a no-guidance control, with at least five fresh samples per wording variant.
3. Run the same scenarios against the revised skill and inspect every decision manually.

The scenarios must cover:

- GGUF generation.
- Safetensors generation.
- Dedicated embedding installation.
- Unsupported generation hardware.
- Custom download-only integration.
- Scripted clone output and cloneability checks.
- VLM/mmproj guidance.

Add focused CLI tests for the corrected help contracts, then run the affected Market package tests and the skill validation/link checks. Do not perform a real model installation or clone as part of verification.

## Scope

In scope:

- `olares-chart` LLM routing and references.
- One new custom-integration reference.
- `market clone/get` help accuracy.
- Skill and CLI regression tests.
- Updating the existing PR with separate documentation and CLI-help commits.

Out of scope:

- Changing the published `beclab/apps` manifests.
- Adding or removing `llm-init` engine adapters.
- Implementing mmproj automation.
- Supporting arbitrary embedding models through the four generation templates.
- Performing a heavyweight model deployment during verification.
