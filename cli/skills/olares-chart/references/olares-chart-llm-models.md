# Deploy an LLM model via the llm-init base apps (env-driven)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This reference covers a *special porting pattern*: instead of authoring a chart, you serve any Hugging Face / Ollama model by **cloning one of four ready-made `llm-init` base charts and filling its env** — no image build, no template edits. GPU image + accelerator sizing concepts are in [olares-chart-gpu.md](olares-chart-gpu.md) / [olares-chart-accelerator.md](olares-chart-accelerator.md).

The four base apps wrap the `llm-init` sidecar (downloads the model, writes a readiness sentinel, serves an OpenAI/Anthropic-compatible `/v1/*` surface) in front of one inference engine. They are `templateOnly: true` + `allowMultipleInstall: true`, so the model identity, source and tuning are **100% install-time env** on `.Values.olaresEnv`:

| base chart | engine | eats | client port |
|---|---|---|---|
| `llamacppllmbasev3` | llama.cpp | one GGUF file | 8090 |
| `ollamallmbasev3` | Ollama | Library tag or one GGUF | 8090 |
| `vllmllmbasev3` | vLLM | full HF safetensors repo | 8090 |
| `sglangllmbasev3` | SGLang | full HF safetensors repo | 8090 |

## When to use

- "Run / serve / host `<some HF or Ollama model>` on my Olares", "give me an OpenAI endpoint for `<model>`", "deploy a local LLM / embedding model".
- You do NOT care which engine — let the format + hardware pick it.

> Anything about authoring your own chart -> parent [`../SKILL.md`](../SKILL.md). App lifecycle verbs (clone/install/upgrade) -> [`../../olares-market/SKILL.md`](../../olares-market/SKILL.md).

## Contents

1. Five-step flow
2. Step 1 — find the model (Hugging Face)
3. Step 2 — evaluate the hardware
4. Step 3 — pick the engine
5. Step 4 — fill the env (the core)
6. Step 5 — install it (base apps are not on the Market yet)
7. Manage / switch the model
8. Errors → fixes

## 1. Five-step flow

```mermaid
flowchart LR
  find["1 find model (HF API)"] --> hw["2 hardware fit (GPU mem)"]
  hw --> pick["3 pick engine (by format)"]
  pick --> env["4 fill env"]
  env --> install["5 package + upload + clone --env"]
```

## 2. Step 1 — find the model (Hugging Face)

No `olares-cli` HF command exists; query the Hub API directly (agent-driven):

- Search: `GET https://huggingface.co/api/models?search=<q>&filter=text-generation&sort=downloads`.
- Inspect one repo: `GET https://huggingface.co/api/models/<owner>/<repo>` (`siblings`, `tags`) and `?blobs=true` for file sizes. Read the model card README for params, recommended VRAM, modality.

Record four facts that drive everything below: **params** (e.g. 7B), **format** (GGUF single file vs safetensors repo), **quant** (Q4_K_M / AWQ / GPTQ / FP8 / fp16), **modality** (text / vision / embedding).

## 3. Step 2 — evaluate the hardware

Read the node's real GPU memory before promising a model fits:

```bash
olares-cli dashboard overview gpu -o json     # per-GPU graphics + tasks (memory)
olares-cli cluster node get <node> -o json    # K8s node detail (capacity/allocatable)
```

Estimate the floor and compare to free VRAM (see [olares-chart-accelerator.md](olares-chart-accelerator.md) §C):

```
GPU memory ≈ weights + KV-cache/activations + ~1–2Gi runtime overhead
weights ≈ params × bytes-per-param   (fp16 ≈ 2B, int8/Q8 ≈ 1B, 4-bit ≈ 0.5B)
```

e.g. 7B fp16 ≈ 14Gi weights → ~16Gi floor; 7B Q4 ≈ 3.5Gi → ~6Gi floor. If it won't fit: pick a smaller quant, offload to CPU (llama.cpp `-ngl` partial / omit), or choose a smaller model.

## 4. Step 3 — pick the engine

GGUF world (Ollama + llama.cpp) and safetensors world (vLLM + SGLang) barely overlap — a model is rarely usable by all four (model-landscape §4):

| model situation | engine | MODEL_SOURCE shape |
|---|---|---|
| single GGUF (a `*-GGUF` quant repo), low VRAM / CPU ok | `llamacpp` | `hf://owner/repo-GGUF --include file.gguf` |
| want Ollama-native `/api/*`, Library tag or GGUF | `ollama` | `ollama://tag` or `hf://...-GGUF --include ...gguf` |
| full HF safetensors + enough GPU, high throughput / TP | `vllm` | `hf://owner/repo` |
| full HF safetensors + want SGLang runtime | `sglang` | `hf://owner/repo` |
| AWQ / GPTQ / FP8 quant repo (safetensors) | `vllm` or `sglang` | `hf://owner/repo` |

> SGLang does **not** eat GGUF; vLLM eats GGUF only experimentally — don't. safetensors → llama.cpp/Ollama needs offline conversion — don't; find a community `*-GGUF` instead.

## 5. Step 4 — fill the env (the core)

Required-per-model envs and how each engine differs:

| env | meaning | per-engine rule |
|---|---|---|
| `MODEL_SOURCE` | download channel | vLLM/SGLang: `hf://owner/repo` (whole repo, **no `--include`**). llama.cpp: `hf://owner/repo-GGUF --include <file>.gguf` (sharded GGUF: give the `*-00001-of-*` name). Ollama: `ollama://tag` / `hf://...-GGUF --include ...gguf`. Mirror via `HF_ENDPOINT` — `--endpoint` inside the value is blacklisted (fail-fast). |
| `MODEL_NAME` | client `model` alias; also fed to the engine | llama.cpp template runs `llama-server -hf "$MODEL_NAME"` → must be `owner/repo` or `owner/repo:quant` matching `MODEL_SOURCE`. vLLM `--model` / SGLang `--model-path` → `owner/repo` matching `MODEL_SOURCE`. Ollama: free alias (may differ from the upstream tag). |
| `MODEL_MODE` | `chat` \| `embedding` | embedding: llama.cpp auto-adds `--embedding`, SGLang auto-adds `--is-embedding`; **vLLM needs `--task embed` in `ENGINE_ARGS` yourself**. |
| `MODEL_SUPPORTS` | capability seed | Pick coarse GROUP tokens (multi-select, comma): `vision` / `tools` / `thinking` / `embedding`. The chart expands each to the `supports_*` keys llm-init validates (`tools` → function-calling keys, `thinking` → reasoning keys, `embedding` → none). Required field; leave only `embedding` (or empty) when the model has no extra caps. |
| `ENGINE_ARGS` | engine-native startup flags (string) | vLLM: `--max-model-len 8192 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 [--quantization awq\|gptq\|fp8]`. SGLang: `--context-length 8192 --mem-fraction-static 0.8 --tp 1`. llama.cpp: `-c 8192 -ngl all -fa on` (drop `-ngl` for CPU). Ollama: `OLLAMA_NUM_CTX=8192 OLLAMA_KEEP_ALIVE=30m` (`KEY=VALUE` list). Unknown tokens pass through, never fail. |
| `<ENGINE>_REQUIRED_GPU_MEMORY` | per-instance GPU quota → `nvidia.com/gpumem` | `LLAMACPP_/OLLAMA_/VLLM_/SGLANG_REQUIRED_GPU_MEMORY`. Accepts `8Gi` / `8192` / `8192Mi` (bare MiB). Set it to the Step-2 floor. Non-editable after install. |
| `HF_ENDPOINT` / `HF_TOKEN` | mirror / private repo | auto-injected from `OLARES_USER_HUGGINGFACE_*`; set a token only for gated/private repos. Read only when an `hf://` source exists. |

`LOG_LEVEL` (debug/info/warn/error) and the `*_CPU_REQUEST` / `*_MEMORY_*` envs default sanely — leave them.

Worked example (Qwen2.5-7B four ways, model-landscape §7):

```bash
# vLLM — full safetensors repo
MODEL_SOURCE=hf://Qwen/Qwen2.5-7B-Instruct  MODEL_NAME=Qwen/Qwen2.5-7B-Instruct
MODEL_MODE=chat  ENGINE_ARGS=--max-model-len 8192 --gpu-memory-utilization 0.9  VLLM_REQUIRED_GPU_MEMORY=16Gi

# llama.cpp — one GGUF
MODEL_SOURCE=hf://bartowski/Qwen2.5-7B-Instruct-GGUF --include Qwen2.5-7B-Instruct-Q4_K_M.gguf
MODEL_NAME=bartowski/Qwen2.5-7B-Instruct-GGUF:Q4_K_M  MODEL_MODE=chat  ENGINE_ARGS=-c 8192 -ngl all -fa on  LLAMACPP_REQUIRED_GPU_MEMORY=6Gi

# Ollama — Library tag
MODEL_SOURCE=ollama://qwen2.5:7b-instruct  MODEL_NAME=qwen2.5:7b-instruct  MODEL_MODE=chat  OLLAMA_REQUIRED_GPU_MEMORY=8Gi
```

## 6. Step 5 — install it (base apps are not on the Market yet)

The four charts are not published, so deploy them as local uploaded charts. Get them from [terminus-apps](https://github.com/Above-Os/terminus-apps) (`git clone` or `gh`), then:

```bash
olares-cli chart package ./llamacppllmbasev3 -o ./dist     # -> dist/llamacppllmbasev3-<ver>.tgz
olares-cli market upload ./dist/llamacppllmbasev3-<ver>.tgz # lands in source 'upload'
olares-cli market clone llamacppllmbasev3 -s upload \
  --title "Qwen2.5 7B Q4" \
  --env MODEL_SOURCE='hf://bartowski/Qwen2.5-7B-Instruct-GGUF --include Qwen2.5-7B-Instruct-Q4_K_M.gguf' \
  --env MODEL_NAME='bartowski/Qwen2.5-7B-Instruct-GGUF:Q4_K_M' \
  --env MODEL_MODE=chat --env MODEL_SUPPORTS=tools \
  --env ENGINE_ARGS='-c 8192 -ngl all -fa on' \
  --env LLAMACPP_REQUIRED_GPU_MEMORY=6Gi --watch
```

- `templateOnly` apps are created via `clone` (the CLI sends `templateClone:true` on 1.12.6+); `clone` mints a per-instance name and `--watch` tracks it. Single local test instance can also use `market install <base> -s upload --env ...`.
- `lint` the chart first if you edited it: `olares-cli chart lint ./<base>`.
- A long `downloading` state is the multi-GB engine image pull (then the model), not a hang — watch byte progress + speed via the `image-service` logs ([olares-chart-deploy.md](olares-chart-deploy.md) §3).

## 7. Manage / switch the model

- Change the model/tuning later: edit the envs and re-apply via the Market lifecycle ([`../../olares-market/SKILL.md`](../../olares-market/SKILL.md)); the shared HF cache (`appCommon/huggingface`) keeps old snapshots so swapping `MODEL_SOURCE` back is instant.
- The capability card (`mode` / `supports` / `context_size` / pricing) is editable at runtime on the llm-init dashboard (its `/v1/*` entrance) via `PUT /api/model-spec`.

## 8. Errors → fixes

| Symptom | Cause | Fix |
|---|---|---|
| `missing required env var(s): ...` | a required env was omitted | add `--env KEY=VALUE` for each |
| `app '<base>' is not cloneable` | not a template/multi-instance row | confirm `market get <base> -s upload -o json` shows `templateOnly`/`allowMultipleInstall` |
| engine pod OOM / CUDA OOM at load | `*_REQUIRED_GPU_MEMORY` or model too big for the GPU | smaller quant, lower `--max-model-len` / `--gpu-memory-utilization`, or a smaller model |
| llama.cpp won't start, bad `-hf` | `MODEL_NAME` not `owner/repo[:quant]` | set `MODEL_NAME` to the repo (and quant) matching `MODEL_SOURCE` |
| SGLang/vLLM can't load a GGUF | wrong engine for the format | use `llamacpp`/`ollama` for GGUF; safetensors for vLLM/SGLang |
| uploaded chart invisible | wrong source | always pair `market upload` with `-s upload` on install/clone |
