# LLM base apps: capability/context tuning + day-2 operations

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first; this is the companion to the LLM model serving deploy flow. It holds the two env deep-dives the Step-4 table points into (`MODEL_SUPPORTS`, context length) plus everything after the first deploy.

## Capability mapping: `MODEL_SUPPORTS`

`MODEL_SUPPORTS` is a **required** field: a comma-joined list of coarse capability GROUP tokens. The chart expands each token one-to-many into the `supports_*` keys llm-init validates against its gateway mirror, then re-joins them into the CSV the container env expects. Pass multiple tokens with a literal comma in one flag value (`--env MODEL_SUPPORTS=tools,thinking`).

| token | expands to `supports_*` keys |
|---|---|
| `vision` | `supports_vision` |
| `tools` | `supports_function_calling`, `supports_parallel_function_calling`, `supports_tool_choice` |
| `thinking` | `supports_reasoning`, `supports_reasoning_effort` |
| `embedding` | (no keys — the required-field fallback) |

Unknown tokens pass through verbatim, so a raw `supports_*` key can be supplied directly when you need one the groups don't cover.

**How to choose from the Hugging Face model card** (the four facts from Step 1 plus a quick card read):

| HF signal | token |
|---|---|
| `pipeline_tag: image-text-to-text`, a "Vision Encoder" / multimodal section | `vision` — **but only when you actually serve the projector** (see caveat) |
| "thinking" / "reasoning mode", `--reasoning-parser`, a `<think>` template | `thinking` |
| "function calling" / "tool use", `--tool-call-parser`, tool-calling chat template | `tools` |
| pure embedding / reranker model | `embedding` (also set `MODEL_MODE=embedding`) |
| no extra capabilities | `embedding` (or leave it as the lone fallback token) |

**Model capability vs deployed capability — the key caveat.** `MODEL_SUPPORTS` must describe what *this deployment actually exposes*, not what the upstream model can do in theory. A natively multimodal model served as a **text-only GGUF** does **not** expose vision, so it must **not** claim `vision`. To genuinely enable vision on llama.cpp you must also pull the mmproj projector via the indexed multi-source form ([download multiple models](#download-multiple-models-at-once)(b), `MODEL_SOURCE_<i>_ROLE=mmproj`); only then add `vision`.

Example — Qwen3.6-27B and Gemma4-26B-A4B-it are both multimodal (`image-text-to-text`) + reasoning + function-calling upstream, but when deployed as a plain text GGUF (no mmproj) the correct value is:

```bash
--env MODEL_SUPPORTS=tools,thinking      # NOT vision: no mmproj projector served
```

## Context length sizing

The context window is set **only** through `ENGINE_ARGS` (llama.cpp `-c` / `LLAMA_ARG_CTX_SIZE`, vLLM `--max-model-len`, SGLang `--context-length`, Ollama `OLLAMA_NUM_CTX`); llm-init passes it verbatim and does **not** validate it. There is no separate context env.

**Default policy: pick the largest context that fits the GPU-memory budget with stable headroom — longer context is better for agent use, so do not default to a small `-c`.** How to size it:

1. **Upper bound = the model's trained context** (HF `config.json` `max_position_embeddings`, or the model card). Never exceed it — it buys nothing and can silently trigger RoPE extension. Modern models train long (Qwen3.6 / Gemma4 are 100K+), so VRAM is usually the real limit, not this cap.
2. **Budget for the KV cache** = `*_REQUIRED_GPU_MEMORY` − weights (quantized size, see the Accelerator sizing §C) − ~1–2Gi runtime overhead. The KV cache grows roughly linearly with context length, so fit the largest `-c` into that remaining budget.
3. **Stretch the budget** so a longer window fits: llama.cpp `-fa on -ctk q8_0 -ctv q8_0` (flash-attention + 8-bit KV ≈ halves KV memory vs f16); vLLM `--kv-cache-dtype fp8`; SGLang `--mem-fraction-static`.
4. **Leave stable headroom** — size for the working peak (concurrent requests, long generations), take the largest value that runs *stably*, not the absolute byte-max. Verify on the real node and raise until it stops fitting.

Notes:

- `-c 0` means "use the model's full trained context" — convenient but it will OOM on a 100K+ model under a fixed GPU budget, so **don't** use it as the default; compute the fitting value instead.
- **CPU / unified-memory modes** (no discrete GPU): the window is bounded by host RAM rather than VRAM — more forgiving, same "as long as it stays stable" policy.
- Keep the advertised `context_size` in sync — see [manage / switch the model](#manage--switch-the-model).

Worked example (Qwen2.5-7B four ways):

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

## Manage / switch the model

- Change the model/tuning later: edit the envs and re-apply via the Market lifecycle ([`../../olares-market/SKILL.md`](../../olares-market/SKILL.md)); the shared HF cache (`appCommon/huggingface`) keeps old snapshots so swapping `MODEL_SOURCE` back is instant.
- The capability card (`mode` / `supports` / `context_size` / pricing) is editable at runtime on the llm-init dashboard (its `/v1/*` entrance) via `PUT /api/model-spec`.
- **`context_size` is decoupled from the engine window** — an auto-generated spec defaults it to `0` (unknown), and nothing syncs it to the engine's actual `-c`. Set it (via `PUT /api/model-spec` or a disk spec at `MODEL_SPEC_PATH`) to **match — and never exceed — the real `-c`** from [context length sizing](#context-length-sizing); otherwise clients trust the advertised window and send prompts the engine truncates or rejects.

## Download-only — pre-warm the shared cache

Leaving `ENGINE_KIND` **empty/unset** puts llm-init in *download-only* mode: it runs the full download lifecycle, writes the readiness sentinel, and flips `phase=ready` (so `/readyz` reports done), but starts **no engine** and mounts **no `/v1/*` data plane** — only the control plane (`/livez` `/readyz` `/api/progress` `/api/config`) stays up. The model bytes just land in the shared HF cache for a sibling container (ComfyUI, an audio app, etc.) to consume. It does **not** exit after downloading; it idles serving the control plane.

**This is the general pattern for any non-LLM app that needs Hugging Face weights** — image generation (ComfyUI / Stable Diffusion), TTS / ASR, rerankers, embedding pipelines, OCR, etc. Instead of baking multi-GB weights into your app image or writing a bespoke downloader, run an llm-init download-only instance (as a sidecar in your chart, or a separately-cloned base instance) to populate the shared, cross-app App Common store `appCommon/huggingface` (mounted at `/cache/hf/hub`), then have your app **mount that same volume read-only** and load the model from the standard `models--<owner>--<repo>/snapshots/<sha>/` layout. Benefits: model provisioning is decoupled from the app image, you get the HF mirror/token auto-injection (`HF_ENDPOINT` / `HF_TOKEN` from `OLARES_USER_HUGGINGFACE_*`) and the progress/sentinel machinery for free, and the cache is deduped/reused across every HF-based app on the box.

The four base charts each hard-wire their engine, so download-only is reached by clearing `ENGINE_KIND` (or via a sibling app that embeds llm-init purely as a downloader). Rules in this mode:

- `MODEL_NAME` is still required; `MODEL_SOURCE` must be `hf://` or `https?://`. An `ollama://` source is **rejected at boot** (exit 2) — there is no Ollama daemon to pull into.
- `ENGINE_ARGS` is ignored (no engine to launch); the dashboard hides the engine-status / args / perf cards.

```bash
# Pre-populate the shared HF cache with a repo for a sibling app — no engine
MODEL_SOURCE=hf://Qwen/Qwen2.5-7B-Instruct  MODEL_NAME=Qwen/Qwen2.5-7B-Instruct
MODEL_MODE=chat  ENGINE_KIND=        # empty == download-only
```

## Download multiple models at once

One llm-init instance can fetch several sources in a single deploy. There are two syntaxes — pick one (they are mutually exclusive):

**(a) Comma form** — list extra sources in `MODEL_SOURCE`, comma-separated. The **first** segment is `role=main` (binds the engine, gets the sentinel, registered as `MODEL_NAME`); every later segment is `role=extra` (downloaded/pre-pulled into the cache only — no sentinel, not registered). A comma is only a separator when it directly precedes a new scheme (`hf://` / `ollama://` / `http(s)://`), so **don't put a bare comma inside a flag value** like `--include`. `MODEL_SOURCE_LOCAL` (only for `https?://`) aligns 1:1 with the comma segments (leave a slot empty for `hf://` / `ollama://`).

```bash
# Main model + an extra model pre-pulled into the same cache
MODEL_SOURCE=ollama://qwen3:0.6b,ollama://llama3:8b   # 1st=main, 2nd=extra
```

**(b) Indexed form** — `MODEL_SOURCE_NUM=N` plus `MODEL_SOURCE_<i>` (1-based, contiguous), each with optional `MODEL_SOURCE_<i>_LOCAL` and `MODEL_SOURCE_<i>_ROLE`. Use this for heterogeneous multi-file models (a main weight + a vision projector), since flag values can carry commas safely here.

| `_ROLE` | meaning |
|---|---|
| `main` | primary weight; **exactly one** required (the default when `_ROLE` is omitted). Multiple `main` → fail-fast. |
| `mmproj` | llama.cpp vision projector. lifecycle writes a `${RUN_DIR}/mmproj_path` sentinel and the wrapper appends `--mmproj <path>`. |
| `extra` | pre-download only; **auto-assigned by the comma form** — don't hand-write it in the indexed form. |

```bash
# VLM main repo + an mmproj projector file (llama.cpp)
MODEL_SOURCE_NUM=2
MODEL_SOURCE_1=hf://InternVL/InternVL3-8B               # _ROLE defaults to main
MODEL_SOURCE_2=https://example.com/mmproj-v3.gguf#sha256=...
MODEL_SOURCE_2_LOCAL=/cache/internvl/mmproj.gguf
MODEL_SOURCE_2_ROLE=mmproj
```

Notes:

- All `hf://` repos share the one HF cache root `/cache/hf/hub` (the `appCommon/huggingface` volume); each repo lands under `models--<owner>--<repo>/snapshots/<sha>/`.
- Sources are fetched **sequentially** (extras first, then the main dispatch; an HF `mmproj` is a second `hf` call). `MAX_CONCURRENT_DOWNLOADS` (default `4`, range `[1,16]`, wired to the HF CLI `--max-workers`) only caps parallelism **within** one fetch — push past ~8–16 and HF Hub starts rate-limiting.

## Errors → fixes

| Symptom | Cause | Fix |
|---|---|---|
| `missing required env var(s): ...` | a required env was omitted | add `--env KEY=VALUE` for each |
| `app '<base>' is not cloneable` | not a template/multi-instance row | confirm `market get <base> -s upload -o json` shows `templateOnly`/`allowMultipleInstall` |
| engine pod OOM / CUDA OOM at load | `*_REQUIRED_GPU_MEMORY` or model too big for the GPU | smaller quant, lower `--max-model-len` / `--gpu-memory-utilization`, or a smaller model |
| llama.cpp won't start, bad `-hf` | `MODEL_NAME` not `owner/repo[:quant]` | set `MODEL_NAME` to the repo (and quant) matching `MODEL_SOURCE` |
| SGLang/vLLM can't load a GGUF | wrong engine for the format | use `llamacpp`/`ollama` for GGUF; safetensors for vLLM/SGLang |
| uploaded chart invisible | wrong source | always pair `market upload` with `-s upload` on install/clone |
| `download-only mode ... cannot use an ollama:// source` | empty `ENGINE_KIND` + `ollama://` | use `hf://` / `https://` for download-only, or set `ENGINE_KIND=ollama` |
| `multiple sources have ROLE=main` / `no source has ROLE=main` | indexed multi-source mis-roled | mark exactly one `MODEL_SOURCE_<i>_ROLE=main` (or omit `_ROLE` on it) |
