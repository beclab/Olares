---
title: Run local LLMs on Olares
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/llm-base-apps
outline: [2, 3]
description: Deploy local LLMs on Olares with Ollama, llama.cpp, vLLM, SGLang, or Strata. Configure model sources, GPU memory, and context length.
head:
  - - meta
    - name: keywords
      content: Olares, Router, Engine Base, self-hosted LLM, Ollama, vLLM, llama.cpp, SGLang, Strata, run LLM locally
doc_version: "2.0"
doc_updated: "2026-11-27"
---

# Run local LLMs with Ollama, llama.cpp, vLLM, SGLang, and Strata

<VersionRouteSelect />

Olares supports five inference engines: Ollama, llama.cpp, vLLM, SGLang, and Strata. Each engine has a base app for creating an independent local model service.

Choose the base app for your model, configure its source and startup settings, and connect AI apps through Router.

## Learning objectives

- Choose an Engine Base app for your model format.
- Create and configure a model instance, then check that it is ready.
- Adjust GPU memory requirements, context length, and engine startup arguments.

## Prerequisites {#before-you-start}

Use Olares 1.12.7 or later. For Olares 1.12.6, select the older guide at the top of this page.

:::tip Access more pre-built models
[Upgrade to the latest Olares version](/manual/olares/settings/update.md) for access to a wider range of pre-built model apps.
:::

## Step 1: Choose a local LLM inference engine {#locate-engine-base-apps}

Choose an Engine Base app based on the model's file format and your workload.

| Model source or need | Engine Base app |
| --- | --- |
| GGUF files from Hugging Face. Recommended for Olares One | llama.cpp Engine Base |
| Model tags from the Ollama library | Ollama Engine Base |
| Hugging Face `.safetensors` weights, serving multiple concurrent requests | vLLM Engine Base |
| Hugging Face `.safetensors` weights, structured generation or reasoning optimizations | SGLang Engine Base |
| Run large MoE models such as Hy4-Flash on a consumer GPU by using system RAM and SSD storage alongside GPU memory | Strata Engine Base |

:::info Choosing vLLM or SGLang
The engine needs to support the model architecture and quantization method. Check the deployment instructions on the Hugging Face model page for a recommended engine and version. For example, an AWQ model requires support for that model's AWQ implementation.
:::

1. Open Market and search for "Engine Base".
2. Find your chosen engine in the results.

   ![Engine Base apps in Market](/images/manual/olares/llm-base-apps1.png#bordered)

## Step 2: Create and configure a local model {#create-a-new-model-instance}

1. Click **View** on your chosen Engine Base app, then click **Create**.

   ![Create a model instance](/images/manual/olares/llm-base-apps-create-instance2.png#bordered)

2. Enter the instance names. These examples use Qwen3.6-27B MTP:

   | Field | What to enter | Where it appears |
   | --- | --- | --- |
   | **New app name** | A unique app name, such as `Qwen3.6-27B-MTP` | Market and Settings |
   | **Shortcut name for [client]** | A unique shortcut name, such as `qwen3.6-27b-mtp` | Launchpad |

   Strata also provides **Shortcut name for [strataui]** for naming its Strata UI shortcut.

3. Click **Create** to open **Configure environment variables for [New-app-name]**.

:::info Hardware accelerator selection
Market prompts you to choose when several hardware acceleration modes are available on your device. With one available mode, Olares selects it automatically and continues to instance configuration.
:::

<a id="configure-engine-environment-variables"></a>

<a id="custom-model"></a>

<a id="recommended-models-and-parameters"></a>

Fill in the fields for your chosen engine below. Expand an example to see a complete configuration.

### Configure llama.cpp for GGUF models {#llama-cpp}

Choose a GGUF file from Hugging Face. `<repo>` includes the user or organization and repository name. `<quantization>` is the quantization type, with one type per instance.

| Field | Format and purpose |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo> --include <file>.gguf` |
| **MODEL_NAME** | `<repo>:<quantization>` |
| **ENGINE_ARGS** | Include `-c <tokens>` for context length in tokens. Separate arguments with spaces |
| **MODEL_MODE** | **Chat** for chat models or **Embedding** for embedding models |
| **MODEL_SUPPORTS** | Select supported capabilities: **Vision** for images, **Tools** for tool calling, or **Thinking** for reasoning. Use **None** for embedding models |
| **LLAMACPP_REQUIRED_GPU_MEMORY** | Minimum GPU memory needed to start, in MB or Gi, such as `23Gi` |
| **LOG_LEVEL** | Logging level, such as `Info` |

For multiple GGUF files, separate complete `hf://... --include ...` entries with commas.

<details>
<summary>Example: Qwen3.6-27B MTP</summary>

Model: [`unsloth/Qwen3.6-27B-MTP-GGUF`](https://huggingface.co/unsloth/Qwen3.6-27B-MTP-GGUF), quantized to `UD-Q4_K_XL`

**MODEL_SOURCE**

```text
hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf
```

For image input, add the matching vision projector to **MODEL_SOURCE**:

```text
hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.6-27B-MTP-GGUF --include mmproj-F16.gguf
```

**MODEL_NAME**

```text
unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL
```

- **MODEL_MODE**: `Chat`
- **MODEL_SUPPORTS**: `Thinking`, `Tools`

**ENGINE_ARGS**

```text
-c 131072 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2
```

This model includes MTP weights. `--spec-type draft-mtp` enables MTP speculative decoding, and `--spec-draft-n-max 2` caps each draft at two tokens.

- **LOG_LEVEL**: `Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**: `23Gi`

</details>

<details>
<summary>Alternative: Qwopus3.6-27B MTP</summary>

Model: [`Jackrong/Qwopus3.6-27B-v2-MTP-GGUF`](https://huggingface.co/Jackrong/Qwopus3.6-27B-v2-MTP-GGUF), quantized to `Q4_K_M`

**MODEL_SOURCE**

```text
hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf
```

For image input, add the matching vision projector to **MODEL_SOURCE**:

```text
hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf,hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include mmproj-F32.gguf
```

**MODEL_NAME**

```text
Jackrong/Qwopus3.6-27B-v2-MTP-GGUF:Q4_K_M
```

- **MODEL_MODE**: `Chat`
- **MODEL_SUPPORTS**: `Thinking`, `Tools`

**ENGINE_ARGS**

```text
-c 131072 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2
```

This model includes MTP weights. `--spec-type draft-mtp` enables MTP speculative decoding, and `--spec-draft-n-max 2` caps each draft at two tokens.

- **LOG_LEVEL**: `Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**: `23Gi`

</details>

### Configure models with Ollama {#ollama}

Choose a model and tag from the Ollama library. `<model>` is the model name, and `<tag>` identifies its version or quantization.

| Field | Format and purpose |
| --- | --- |
| **MODEL_SOURCE** | `ollama://<model>:<tag>` |
| **MODEL_NAME** | `<model>:<tag>` |
| **ENGINE_ARGS** | Include `OLLAMA_CONTEXT_LENGTH=<tokens>` for context length in tokens. Separate arguments with spaces |
| **MODEL_MODE** | **Chat** for chat models or **Embedding** for embedding models |
| **MODEL_SUPPORTS** | Select supported capabilities: **Vision** for images, **Tools** for tool calling, or **Thinking** for reasoning. Use **None** for embedding models |
| **OLLAMA_REQUIRED_GPU_MEMORY** | Minimum GPU memory needed to start, in MB or Gi, such as `23Gi` |
| **LOG_LEVEL** | Logging level, such as `Info` |

<details>
<summary>Example: Gemma 4 26B</summary>

Model: [`gemma4:26b`](https://ollama.com/library/gemma4:26b) from the Ollama library, which defaults to Q4_K_M quantization

**MODEL_SOURCE**

```text
ollama://gemma4:26b
```

**MODEL_NAME**

```text
gemma4:26b
```

- **MODEL_MODE**: `Chat`
- **MODEL_SUPPORTS**: Thinking, Tools, Vision

**ENGINE_ARGS**

```text
OLLAMA_KEEP_ALIVE=-1 OLLAMA_CONTEXT_LENGTH=131072 OLLAMA_FLASH_ATTENTION=1 OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_NUM_PARALLEL=1
```

- **OLLAMA_REQUIRED_GPU_MEMORY**: `23Gi`

</details>

### Configure SGLang for Hugging Face models {#sglang}

Choose a Hugging Face repository with `.safetensors` weights supported by SGLang. `<repo>` includes the user or organization and repository name.

| Field | Format and purpose |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo>` |
| **MODEL_NAME** | `<repo>` |
| **ENGINE_ARGS** | Include `--context-length <tokens>` for context length in tokens. Separate arguments with spaces |
| **MODEL_MODE** | **Chat** for chat models or **Embedding** for embedding models |
| **MODEL_SUPPORTS** | Select supported capabilities: **Vision** for images, **Tools** for tool calling, or **Thinking** for reasoning. Use **None** for embedding models |
| **SGLANG_REQUIRED_GPU_MEMORY** | Minimum GPU memory needed to start, in MB or Gi, such as `23Gi` |
| **LOG_LEVEL** | Logging level, such as `Info` |

<details>
<summary>Example: Qwen3.5-9B AWQ</summary>

Model: [`Qwen3.5-9B-AWQ-4bit`](https://huggingface.co/cyankiwi/Qwen3.5-9B-AWQ-4bit) from Hugging Face

**MODEL_SOURCE**

```text
hf://cyankiwi/Qwen3.5-9B-AWQ-4bit
```

**MODEL_NAME**

```text
cyankiwi/Qwen3.5-9B-AWQ-4bit
```

- **MODEL_MODE**: `Chat`
- **MODEL_SUPPORTS**: `Thinking`, `Tools`, `Vision`

**ENGINE_ARGS**

```text
--context-length 65536 --mem-fraction-static 0.85 --chunked-prefill-size 4096 --reasoning-parser qwen3 --tool-call-parser qwen3_coder
```

This example uses `--reasoning-parser qwen3` to extract reasoning and `--tool-call-parser qwen3_coder` to parse tool calls. When changing model families, choose parsers from the new model's deployment instructions.

- **LOG_LEVEL**: `Info`
- **SGLANG_REQUIRED_GPU_MEMORY**: `23Gi`

</details>

### Configure vLLM for Hugging Face models {#vllm}

Choose a Hugging Face repository with `.safetensors` weights supported by vLLM. `<repo>` includes the user or organization and repository name.

| Field | Format and purpose |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo>` |
| **MODEL_NAME** | `<repo>` |
| **ENGINE_ARGS** | Include `--max-model-len <tokens>` for context length in tokens. Separate arguments with spaces |
| **MODEL_MODE** | **Chat** for chat models or **Embedding** for embedding models |
| **MODEL_SUPPORTS** | Select supported capabilities: **Vision** for images, **Tools** for tool calling, or **Thinking** for reasoning. Use **None** for embedding models |
| **VLLM_REQUIRED_GPU_MEMORY** | Minimum GPU memory needed to start, in MB or Gi, such as `23Gi` |
| **LOG_LEVEL** | Logging level, such as `Info` |

<details>
<summary>Example: Gemma-4-12B AWQ</summary>

Model: [`gemma-4-12B-it-AWQ-INT4`](https://huggingface.co/cyankiwi/gemma-4-12B-it-AWQ-INT4) from Hugging Face

**MODEL_SOURCE**

```text
hf://cyankiwi/gemma-4-12B-it-AWQ-INT4
```

**MODEL_NAME**

```text
cyankiwi/gemma-4-12B-it-AWQ-INT4
```

- **MODEL_MODE**: `Chat`
- **MODEL_SUPPORTS**: `Thinking`, `Tools`, `Vision`

**ENGINE_ARGS**

```text
--max-model-len 65536 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 --max-num-batched-tokens 8192 --tool-call-parser gemma4 --reasoning-parser gemma4 --enable-prefix-caching --enable-auto-tool-choice
```

This example uses `gemma4` for both reasoning and tool-call parsing. `--enable-auto-tool-choice` lets the model choose tool calls. When changing model families, update the parsers using the new model's deployment instructions.

- **LOG_LEVEL**: `Info`
- **VLLM_REQUIRED_GPU_MEMORY**: `23Gi`

</details>

### Configure models with Strata {#strata}

Choose a Strata-supported model pack and its companion files from Hugging Face. `<repo>` includes the user or organization and repository name. `<file>` is the filename, and `<model-name>` is a custom name for the model service.

:::info Strata configuration fields
Use `MODEL_SOURCE`, `MODEL_NAME`, and `MODEL_SUPPORTS` to configure the model and its capabilities. Olares allocates GPU resources based on the app's supported hardware and the resources available on your device. The base app generates model paths and service addresses. Use `ENGINE_ARGS` for inference options.
:::

| Field | Format and purpose |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo> --include <file>`. Include the model pack and its required companion files |
| **MODEL_NAME** | `<model-name>`, which identifies the model service |
| **ENGINE_ARGS** | Include `--context <tokens>` for context length in tokens. Separate arguments with spaces. See [Strata startup parameters](#strata-tuning) for common options |
| **MODEL_MODE** | Select **Chat** |
| **MODEL_SUPPORTS** | Select supported capabilities: **Vision** for images, **Tools** for tool calling, or **Thinking** for reasoning. Use **None** for a model with no additional capabilities |
| **STRATA_CPU_LIMIT** | Maximum CPU cores the container can use, such as `19` |
| **STRATA_MEMORY_LIMIT** | Container system memory limit, such as `56Gi`. Exceeding it can terminate and restart the container due to an out-of-memory error |
| **STRATA_REQUIRED_GPU_MEMORY** | Minimum GPU memory needed to start, in MB or Gi, such as `23Gi` |
| **LOG_LEVEL** | Logging level, such as `Info` |

Separate complete `hf://... --include ...` entries with commas. Include the model pack and companion files for the same model version.

Set CPU and RAM limits and the GPU memory requirement to suit your device, with room for model loading and inference.

<details>
<summary>Example: Hy4-Flash SQ-64</summary>

Model: [strata-engine/Hy4-Flash-Strata](https://huggingface.co/strata-engine/Hy4-Flash-Strata), using the SQ-64 pack. The model pack and required Engram file occupy about 130 GB of disk space together.

:::info Example resource budget
These values are estimates for a device with 64 GiB of system RAM, 24 GiB of GPU memory, and 19 available CPU cores, using a 128K text context. Adjust the limits using memory measurements during model loading and long-prompt processing.
:::

**MODEL_SOURCE**

```text
hf://strata-engine/Hy4-Flash-Strata --include Hy4-Flash-SQ-64.strata,hf://strata-engine/Hy4-Flash-Strata --include Hy4-Flash-Engram-Q8.strata
```

**MODEL_NAME**

```text
hy4-flash-sq-64
```

- **MODEL_SUPPORTS**: `Tools`, `Thinking`

**ENGINE_ARGS**

```text
--context 131072
```

This configuration keeps the default DFly2 drafter. Use these options to adjust this model:

| Goal | What to enter |
| --- | --- |
| Disable speculative decoding | `--context 131072 --draft off` |
| Use Q8 KV caching to reduce cache memory | `--context 131072 -- --kv q8` |
| Use FP16 KV caching | `--context 131072 -- --kv fp16` |
| Set GPU and system RAM budgets for expert data | Add `--vram-experts-gb <GiB>` and `--ram-experts-gb <GiB>` to `ENGINE_ARGS`, with headroom for other model components and buffers |

The standalone `--` passes the following KV options directly to the Hy4-Flash engine. `--kv` also accepts `auto` to keep the engine default. Expert budgets control model placement. The resource fields below set container limits and the GPU memory requirement.

| Resource field | Example value | Purpose |
| --- | --- | --- |
| **STRATA_CPU_LIMIT** | `19` | Allow up to 19 CPU cores on this example device |
| **STRATA_MEMORY_LIMIT** | `56Gi` | Allow memory growth while leaving 8 GiB of the device RAM for Olares and other services |
| **STRATA_REQUIRED_GPU_MEMORY** | `23Gi` | Schedule on a GPU with at least 23 GiB available |

The budget is informed by Strata's [Hy4 memory planner](https://github.com/Niko1221/Strata/blob/main/terminal/strata_control/hy4_fit.py). Adjust the CPU limit and GPU memory requirement for your device. With less GPU memory, Strata keeps more model data in RAM or reads it from the SSD.

</details>

## Step 3: Install the model app

1. After filling in the configuration for your chosen engine, click **Confirm** to save it and start installation.
2. Follow installation progress in the **Instances** panel on the right. Wait for the button to change to **Open**.

   ![Model instance installed](/images/manual/olares/llm-base-model-instance-installed1.png#bordered)

## Step 4: Check model download and engine status {#monitor-deployment-and-configure-the-model-service}

After installation, open the model app to follow its download and engine startup progress.

1. Open the new model app from **Instances** or Launchpad.
2. On the default **Status** tab, follow the download progress. Model files download automatically.
3. Wait for both entries under **Service status**:

   | Entry | Ready to continue | Meaning |
   | --- | --- | --- |
   | **Model** | **Ready** | Model files are downloaded and verified |
   | **Engine** | **Running** | The inference service is online |

   ![Model console ready](/images/manual/olares/llm-base-model-console-status.png#bordered)

vLLM and SGLang can take a while to load a model. If either status does not become ready, see [Model or engine is not ready](/manual/help/ts-model-engine-not-ready.md).

## Connect AI apps through Olares Router {#connect-client-apps-to-the-model-service}

Once the model is ready, follow [Get connection details from Olares Router](olares-router.md#get-connection-details-from-router) to obtain the URL, model name, and authentication details for your client.

## Adjust GPU memory and model context length {#references}

### Adjust the instance's GPU memory requirement

Use the model instance's environment settings to adjust its minimum GPU memory requirement.

1. Go to **Settings** > **Applications** and select your model instance.
2. Open **Manage environment variables** and find the field for your engine:

   | Engine | Field |
   | --- | --- |
   | llama.cpp | `LLAMACPP_REQUIRED_GPU_MEMORY` |
   | Ollama | `OLLAMA_REQUIRED_GPU_MEMORY` |
   | SGLang | `SGLANG_REQUIRED_GPU_MEMORY` |
   | vLLM | `VLLM_REQUIRED_GPU_MEMORY` |
   | Strata | `STRATA_REQUIRED_GPU_MEMORY` |

3. Click the edit icon beside the field and enter a memory value in MB or Gi, such as `23Gi`. Keep it below total VRAM in time slicing or exclusive mode, or below remaining VRAM in memory slicing mode. Use `0` for CPU mode.
4. Save the change, then click **Apply**.
5. Open the model app and check that **Model** shows **Ready** and **Engine** shows **Running**.

### Check the engine's context length {#view-model-details-in-router}

If your client asks for a context length, check the actual setting in Router. These screenshots use Qwen3.8-27B (llama.cpp). Read the value for your own instance.

<!--@include: ../reusables/ai-service-connections.md#model-context-window-->

### Change context length or GPU memory usage

1. On Router's **LLM** page, find the model and click its information icon to open **Model card**.
2. Click **Edit** and update the engine arguments. For example, llama.cpp uses `-c` for context length, and vLLM uses `--gpu-memory-utilization` for GPU memory fraction. For other engines, see [Inference engine startup parameters](#engine-tuning-arguments) below.
3. Save your changes. This restarts the engine. Wait for the model to become available before sending requests.
4. If you changed the context length, set the client's value to the engine's configured length or lower.

## Inference engine startup parameters {#engine-tuning-arguments}

Use `ENGINE_ARGS` to adjust context length, memory use, and concurrency. The tables below cover common tuning options. Model-specific settings are explained in the deployment examples above.

### llama.cpp {#llama-cpp-tuning}

Enter space-separated flags, such as `-c 65536 -np 1 -ngl all -fa on`. The base app also accepts `LLAMA_ARG_*=value` environment settings. See the [llama.cpp official reference](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md) for the full parameter list and supported configurations.

| Argument | Purpose | Example value |
| --- | --- | --- |
| `-c` / `--ctx-size` | Context token budget. With multiple slots, check the per-slot context reported by the server. | `65536` |
| `-np` / `--parallel` | Number of parallel request slots. Use `1` to dedicate the context budget to one request. | `1` |
| `-ngl` / `--gpu-layers` | Number of layers to offload to the GPU. `all` requests full offload. A smaller count reduces GPU memory use. | `all` |
| `-fa` / `--flash-attn` | Flash Attention mode. | `on`, `off`, or `auto` |
| `-ctk` / `-ctv` | K and V cache data types. Quantized caches use less memory. Supported combinations depend on the model and backend. | `q8_0` |
| `-t` / `--threads` | CPU threads used during generation. | `8` |
| `--jinja` | Enables Jinja chat templates, including templates used for tool calling. | Flag only |

### Ollama {#Ollama-tuning}

Enter space-separated environment assignments, such as `OLLAMA_CONTEXT_LENGTH=8192 OLLAMA_NUM_PARALLEL=1`. These configure the Ollama server. Request options such as `temperature` belong in the client or API request. See the [Ollama official reference](https://docs.ollama.com/faq) for the full parameter list and supported configurations.

| Setting | Purpose | Example value |
| --- | --- | --- |
| `OLLAMA_CONTEXT_LENGTH` | Default context length in tokens. | `8192` |
| `OLLAMA_KEEP_ALIVE` | How long a model stays loaded. The base app defaults to `-1`, which keeps it loaded. | `-1` or `5m` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention, required for KV cache quantization. | `1` |
| `OLLAMA_KV_CACHE_TYPE` | KV cache precision. Lower precision reduces memory use and can affect output quality. | `f16`, `q8_0`, or `q4_0` |
| `OLLAMA_NUM_PARALLEL` | Requests processed concurrently per model. More parallel requests require more context memory. | `1` |
| `OLLAMA_MAX_QUEUE` | Maximum queued requests before the server rejects additional work. | `64` |

### SGLang {#SGLang-tuning}

Enter space-separated flags, such as `--context-length 65536 --mem-fraction-static 0.85`. See the [SGLang official reference](https://docs.sglang.io/docs/advanced_features/server_arguments) for the full parameter list and supported configurations.

:::warning SGLang 0.5.23 parameter change
This version removes `--speculative-skip-dp-mlp-sync`. Older configurations containing this flag cause startup to fail.
:::

| Argument | Purpose | Example value |
| --- | --- | --- |
| `--context-length` | Maximum sequence length in tokens. | `65536` |
| `--mem-fraction-static` | Fraction of GPU memory reserved for model weights and the KV cache, leaving the rest for dynamic allocations. | `0.85` |
| `--chunked-prefill-size` | Tokens per prefill chunk. Smaller chunks reduce prefill memory peaks but can reduce prompt-processing throughput. | `4096` |
| `--max-running-requests` | Maximum requests running concurrently. | `4` |
| `--tp-size` | Number of GPUs used for tensor parallelism. | `1` |
| `--kv-cache-dtype` | KV cache data type, subject to model and GPU support. | `auto` or `fp8_e4m3` |

### vLLM {#vLLM-tuning}

Enter space-separated flags, such as `--max-model-len 65536 --gpu-memory-utilization 0.9`. GPU utilization is relative to the memory visible to the engine. On Olares, `VLLM_REQUIRED_GPU_MEMORY` sets that GPU memory budget. See the [vLLM official reference](https://docs.vllm.ai/en/latest/configuration/engine_args/) for the full parameter list and supported configurations.

:::tip Leave GPU memory headroom for vLLM
The base app using vLLM 0.34.0 enables extensible KV caching and defaults GPU utilization to `1.0`. A lower `--gpu-memory-utilization` value, such as `0.9`, leaves headroom within the allocated GPU budget.
:::

| Argument | Purpose | Example value |
| --- | --- | --- |
| `--max-model-len` | Maximum sequence length, including input and output tokens. | `65536` |
| `--gpu-memory-utilization` | Fraction of visible GPU memory allocated to the engine. | `0.9` |
| `--tensor-parallel-size` | Number of GPUs used for tensor parallelism. | `1` |
| `--max-num-seqs` | Maximum sequences processed in one iteration. | `4` |
| `--max-num-batched-tokens` | Maximum tokens processed in one iteration. Tune with concurrency and prefill chunking to balance throughput and latency. | `8192` |
| `--enable-prefix-caching` | Reuses cached KV data for shared prompt prefixes. | Flag only |
| `--kv-cache-dtype` | KV cache data type, subject to model and GPU support. | `auto` or `fp8` |

### Strata {#strata-tuning}

Separate arguments with spaces. Drafter, KV cache, and expert-loading options depend on the model family. See the [Strata model family guide](https://github.com/Niko1221/Strata/blob/main/docs/MODELS.md) for available models, memory requirements, and configuration options by family. For launch argument usage, see the [Strata launch reference](https://github.com/Niko1221/Strata/blob/main/docs/TERMINAL_LAUNCH.md).

| Argument | Purpose and scope | Example value |
| --- | --- | --- |
| `--context` | Context length in tokens. The base app also accepts `--max-context` and `--ctx` as aliases. | `131072` |
| `--draft` | Controls speculative decoding. `off` disables drafting. Choose an enabled drafter for the model family. | `off` |

## Learn more

- [Connect AI apps through Olares Router](olares-router.md#get-connection-details-from-router)
- [Troubleshoot model or engine readiness](/manual/help/ts-model-engine-not-ready.md)
