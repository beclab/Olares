# Serve a generation model from the official Market

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This workflow clones one of four published generation templates from the default `market.olares` source. Embeddings use the dedicated admin-only, shared application instead: let [`olares-market`](../../olares-market/SKILL.md) own its lifecycle with `olares-cli market install embeddinggemmav3`, add no arbitrary model env, and for a non-interactive install pass an explicit `--compute-mode` when required, choosing only `cpu`, `intel`, `nvidia`, or `nvidia-gb10`.

## 1. Confirm the target

Use a logged-in administrator profile and confirm:

- Olares is `>=1.12.6-0`.
- The target node offers `nvidia` or `nvidia-gb10`.
- The administrator accepts that the cloned template is admin-only and shared.

Inspect the profile and accelerator before choosing a model:

```bash
olares-cli profile list
olares-cli cluster node list
olares-cli dashboard overview gpu -o json
```

The official generation templates support no other compute mode. If neither mode is available, stop and report that this workflow does not apply.

## 2. Inspect the model

Read the Hugging Face or Ollama metadata and record:

1. parameter count;
2. artifact format: GGUF or a full safetensors repository;
3. quantization: for example Q4_K_M, AWQ, GPTQ, FP8, or fp16;
4. modality: text-only or multimodal.

For Hugging Face, inspect `GET https://huggingface.co/api/models/<owner>/<repo>?blobs=true` and the model card. Compare the weight size and recommended memory with the real accelerator capacity before cloning.

## 3. Select the engine

| Model artifact | Published template | Engine |
|---|---|---|
| One GGUF file | `llamacppllmbasev3` | llama.cpp |
| Ollama Library tag or one GGUF file | `ollamallmbasev3` | Ollama |
| Full safetensors repository, including AWQ/GPTQ/FP8 | `vllmllmbasev3` | vLLM |
| Full safetensors repository | `sglangllmbasev3` | SGLang |

Keep GGUF on llama.cpp or Ollama. Use vLLM or SGLang for full safetensors repositories.

## 4. Fill the published environment

Every official generation clone uses:

- `MODEL_MODE=chat`;
- `MODEL_SUPPORTS=vision`, `tools`, `thinking`, `none`, or a comma-separated combination from [capability mapping](olares-chart-llm-ops.md#capability-mapping);
- the engine-specific values below.

| Template | `MODEL_SOURCE` | `MODEL_NAME` | `ENGINE_ARGS` | Required memory env |
|---|---|---|---|---|
| llama.cpp | `hf://owner/repo-GGUF --include file.gguf` | `owner/repo-GGUF:quant` | for example `-c 8192 -ngl all -fa on` | `LLAMACPP_REQUIRED_GPU_MEMORY` |
| Ollama | `ollama://tag` or an HF GGUF source | client model alias | for example `OLLAMA_NUM_CTX=8192 OLLAMA_KEEP_ALIVE=30m` | `OLLAMA_REQUIRED_GPU_MEMORY` |
| vLLM | `hf://owner/repo` | `owner/repo` | for example `--max-model-len 8192 --gpu-memory-utilization 0.9 --tensor-parallel-size 1` | `VLLM_REQUIRED_GPU_MEMORY` |
| SGLang | `hf://owner/repo` | `owner/repo` | for example `--context-length 8192 --mem-fraction-static 0.8 --tp 1` | `SGLANG_REQUIRED_GPU_MEMORY` |

Pass the same per-engine `<ENGINE>_REQUIRED_GPU_MEMORY` env with either compute mode. Its scheduler meaning follows the selected mode:

```text
--compute-mode nvidia      --env <ENGINE>_REQUIRED_GPU_MEMORY=<discrete-GPU-memory floor>
--compute-mode nvidia-gb10 --env <ENGINE>_REQUIRED_GPU_MEMORY=<unified pod-memory floor>
```

For `nvidia`, the value sizes `nvidia.com/gpumem`. For `nvidia-gb10`, it sizes pod memory instead; GB10 has no separate GPU-memory quota.

The published manifests already set `options.apiTimeout: 0`. Clone them directly; no chart edit is required for long generations.

## 5. Clone and watch

This example serves a GGUF model through llama.cpp on a discrete NVIDIA GPU:

```bash
olares-cli market clone llamacppllmbasev3 \
  --title "Qwen2.5 7B Q4" \
  --compute-mode nvidia \
  --env MODEL_SOURCE='hf://bartowski/Qwen2.5-7B-Instruct-GGUF --include Qwen2.5-7B-Instruct-Q4_K_M.gguf' \
  --env MODEL_NAME='bartowski/Qwen2.5-7B-Instruct-GGUF:Q4_K_M' \
  --env MODEL_MODE=chat \
  --env MODEL_SUPPORTS=tools \
  --env ENGINE_ARGS='-c 8192 -ngl all -fa on' \
  --env LLAMACPP_REQUIRED_GPU_MEMORY=6Gi \
  --watch \
  --watch-timeout 30m
```

Omitting `--source` intentionally uses `market.olares`. For a compatible GB10 node, change only `--compute-mode` to `nvidia-gb10` and set the same `LLAMACPP_REQUIRED_GPU_MEMORY` env to the required unified pod-memory floor.

After the clone request, follow the returned target application name with `olares-cli market status <target-app> --watch`. If download, scheduling, startup, or reachability stalls, use [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md) for runtime evidence and diagnosis. Capability selection, context sizing, updates, and common corrections continue in [llm-ops.md](olares-chart-llm-ops.md).
