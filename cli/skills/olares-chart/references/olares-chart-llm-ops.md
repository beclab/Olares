# Operate an official generation clone

> **Prerequisite:** complete [the official generation workflow](olares-chart-llm-models.md) first. This reference covers only capability selection, context sizing, install-time changes, and diagnosis after cloning a published template.

## Capability mapping

`MODEL_SUPPORTS` is a comma-separated list of chart-level groups. The published chart expands these groups before starting `llm-init`.

| Group | Expanded capability |
|---|---|
| `vision` | `supports_vision` |
| `tools` | `supports_function_calling`, `supports_parallel_function_calling`, `supports_tool_choice` |
| `thinking` | `supports_reasoning`, `supports_reasoning_effort` |
| `none` | no extra capability |

Choose only capabilities that the deployed engine actually exposes. A model card that mentions vision does not by itself prove that a particular serving artifact exposes vision. Use `none` for a generation model with no extra capability.

## Context sizing

Set the engine window through `ENGINE_ARGS`:

- llama.cpp: `-c`;
- Ollama: `OLLAMA_NUM_CTX`;
- vLLM: `--max-model-len`;
- SGLang: `--context-length`.

For `nvidia`, budget discrete GPU memory for weights, runtime overhead, and KV cache. Start below the model's trained maximum, retain headroom for concurrent requests, and increase only after the clone is stable. A smaller quantization or window is the normal correction for OOM.

For `nvidia-gb10`, use the same stability rule but budget against unified pod memory rather than a separate GPU-memory quota. Keep the window at or below the model's trained maximum on both modes.

Useful engine-specific controls include llama.cpp flash attention and quantized KV cache, vLLM `--kv-cache-dtype fp8`, and SGLang `--mem-fraction-static`. Apply only controls supported by the selected engine and model.

## Change what a deployed clone serves

Use the Market lifecycle in [`../../olares-market/SKILL.md`](../../olares-market/SKILL.md) to update `MODEL_SOURCE`, `MODEL_NAME`, `MODEL_SUPPORTS`, `ENGINE_ARGS`, or related published environment values. Keep the engine aligned with the artifact format when switching models.

Those values seed the first boot only. Once the clone is running, the model card inside it is what it serves from, and editing that card belongs to [`olares-router`](../../olares-router/SKILL.md) — `router model spec show/set --app` reads and replaces it, and Router has to be told to re-mirror it afterwards. Reinstalling with different environment values is the long way round to a change that reference makes directly.

## Errors and routing

| Symptom | Correction |
|---|---|
| The catalog item is not cloneable | Run `olares-cli market get <base>` against the default source and inspect the human-readable cloneability fields; confirm the selected app is one of the four published templates. |
| Non-interactive clone requires compute selection | Repeat with `--compute-mode nvidia` or `--compute-mode nvidia-gb10`, matching the target node. |
| Requested node has neither supported mode | Stop; the official generation templates do not support that hardware. |
| llama.cpp or Ollama cannot load full safetensors | Select vLLM or SGLang. |
| vLLM or SGLang cannot load GGUF | Select llama.cpp or Ollama. |
| Engine load or generation reaches OOM | Use a smaller quantization, reduce the context window or concurrency, or select a smaller model. |
| Download, scheduling, startup, or reachability stalls | Follow [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md) and diagnose the returned target application. |

Requests for a chart-owned sidecar, shared-cache provisioning, or download-only behavior belong to [custom `llm-init` integration](olares-chart-llm-init-integration.md), not post-clone operation.
