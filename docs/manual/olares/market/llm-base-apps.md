---
outline: [2, 3]
description: Learn how to use the LLM Base applications on Olares to self-host large language models and run different inference engines by cloning the base apps.
---

# Host local large language models with LLM Base apps

Olares V1.12.6 introduces the local hosting and management platform for large language models (LLMs), a self-hosting solution powered by the `llm-init` project. This platform provides four LLM Base applications, each for one inference engine: **Ollama LLM Base**, **vLLM LLM Base**, **llama.cpp LLM Base**, and **SGLang LLM Base**. Select the base app for the engine you want, use it to deploy different models, and then monitor model performance through a dedicated dashboard. 

## Before you start

- Your Olares system has been upgraded to V1.12.6 or later.

## Locate LLM Base apps

1. Open Market and search for "LLM Base". Four base apps appear: vLLM LLM Base (llm-init), SGLang LLM Base (llm-init), Ollama LLM Base (llm-init), and llama.cpp LLM Base (llm-init).

    ![LLM Base apps in Market](/images/manual/olares/llm-base-apps.png#bordered)

2. Each base app is optimized for a different inference scenario. Choose one based on your model source, performance needs, and hardware:

    | Base app | Best for |
    | :--- | :--- |
    | **Ollama LLM Base (llm-init)** | Getting started quickly and broad model compatibility. <br>Ollama pulls models automatically using native model<br> tags, making it ideal for chat and embedding tasks. |
    | **vLLM LLM Base (llm-init)** | High-throughput serving of Hugging Face models<br> under heavy concurrent load. |
    | **SGLang LLM Base (llm-init)** | Efficient structured generation and advanced<br> reasoning optimizations. |
    | **llama.cpp LLM Base (llm-init)** | Lightweight GGUF models or deployments with<br> limited GPU memory. |

## Create a new model instance

An LLM Base app serves as a template. To run a model, you must first clone the base app into an independent running instance.

1. Select the base app that matches your preferred inference engine, and then click **View** on it. For example, **Ollama LLM Base (llm-init)**.
2. Click **Create** to initialize a new instance.
3. Specify the instance identity settings:

    - **New app name**: Enter a unique name for the instance. This name is displayed as the app name in **Market** > **My Olares**.
    - **Shortcut name for {client}**: Enter a unique shortcut name to be displayed on the Launchpad.

4. Select **Create** to proceed to the environment configuration.

## Configure engine environment variables

After creating the instance, the configuration window opens. Define where your engine pulls the model, how much memory it uses, and what capabilities it exposes to other client apps.

1. In the **Configure environment variables for {New-app-name}** window, fill in the following details according to the target model and engine:

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify the model source address. <br>The format depends on the selected engine.<br>Example: `ollama://qwen3.5:0.8b` or `hf://Qwen/Qwen3.5-2B`. |
    | **MODEL_NAME** | Specify the exact model name for client application connections. <br>Example: `qwen3.5-2b`. |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. <br>Example: `chat`. |
    | **MODEL_SUPPORTS** | Enter a comma-separated list of model capabilities:<ul><li>**Reasoning models**: Include `supports_reasoning`.</li><li>**Tool calling models**: Include `supports_function_calling`.<br>Add `supports_parallel_function_calling` for simultaneous tasks.</li><li>**Vision models**: Include `supports_vision`.</li><li>**Embedding models**: Leave this field empty. Do not include chat capability flags.</li></ul>Example: `supports_function_calling,supports_tool_choice`.<br>Reference: [Model capability flags](#reference-model-capability-flags).|
    | **ENGINE_ARGS** | Specify the startup parameters for the engine.<br>Separate multiple parameters with spaces.<br>Example: `OLLAMA_CONTEXT_LENGTH=4096`.<br>Reference: [Ollama tuning arguments](#reference-ollama-tuning-arguments). |
    | **{ENGINE}_REQUIRED<br>_GPU_MEMORY** | Sets the minimum video memory required by the instance in MB or Gi. <br>Example: `8192`. |

2. Click **Confirm** to save the configuration and start the instance installation.

    The **Instances** panel appears on the right side of the page and displays the installation progress. After the setup completes, the instance action button changes to **Open**, indicating the base service is running.

    ![LLM Base instances installed](/images/manual/olares/llm-base-model-instance-installed.png#bordered)

### Reference: Model capability flags

The `MODEL_SUPPORTS` variable declares what features the model exposes to external clients. These flags apply universally across all inference engines.

| Category | Supported Flags |
| --- | --- |
| **Core** | `supports_vision`, `supports_function_calling`,<br>`supports_reasoning`, `supports_native_streaming`,<br>`supports_response_schema`, `supports_prompt_caching`, <br>`supports_web_search`, `supports_parallel_function_calling` |
| **Multimodal** | `supports_audio_input`, `supports_audio_output`,<br> `supports_video_input`, `supports_pdf_input`,<br> `supports_computer_use`, `supports_url_context` |
| **Reasoning + control tokens** | `supports_reasoning_effort`, `supports_thinking`,<br> `supports_assistant_prefill`, `supports_tool_choice`,<br> `supports_tokenizer` |
| **Sampling controls** | `supports_system_messages`, `supports_temperature`,<br> `supports_top_p`, `supports_top_k`,<br> `supports_stop_sequences`, `supports_frequency_penalty`,<br> `supports_presence_penalty` |
| **Response shape** | `supports_n`, `supports_logprobs`, `supports_seed`,<br> `supports_response_format`, `supports_logit_bias`, `supports_user` |

### Reference: Engine tuning arguments

Use the `ENGINE_ARGS` variable to add custom settings that adjust memory usage, context limits, and processing behaviors. Select your inference engine below to view the available tuning arguments.

<tabs>
<template #Ollama>

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | Sets the default context window<br> size in tokens. <br><br>Default scales by VRAM:<ul><li>Less than 24G: 4096</li><li>Between 24G and 48G: 32768</li><li>48G and more: 262144</li></ul> | `8192` to `131072` |
| `OLLAMA_NUM_PARALLEL` | Sets the number of concurrent<br> requests processed per model. <br><br>Ollama determines this automatically<br> based on your available VRAM, <br>typically `1` or `4`.<br><br>Default: `0` | `1` |
| `OLLAMA_KV_CACHE_TYPE` | Sets the KV cache quantization type<br> to save video memory. <br><br>Default: `f16`. | `q8_0` (minor precision loss) or `q4_0` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention to optimize<br> memory efficiency during long-context<br> operations. <br><br>Default: `0` (Disabled). | `1` (Enabled) |
| `OLLAMA_MAX_LOADED_MODELS` | Sets the maximum number of models<br> kept loaded in memory simultaneously.<br>It automatically scales to roughly 3<br> models per available GPU<br><br>Default: `0`. | `1` |
| `OLLAMA_MAX_QUEUE` | Sets the maximum number of incoming<br> requests allowed in the processing queue. <br><br>Default: `512`. | `512` |
| `OLLAMA_KEEP_ALIVE` | Sets model resident duration in memory<br> after the last request. Use `-1` <br>for permanent retention. <br><br>Default: `5m`. | `30m` or `-1` |
| `OLLAMA_LOAD_TIMEOUT` | Sets the maximum duration to wait for<br> a model to finish loading before giving up. <br><br>Default: `5m`. | `5m` |
| `OLLAMA_GPU_OVERHEAD` | Sets the amount of video memory <br>in bytes reserved as a safety margin<br> overhead. <br><br>Default: `0`. | `0` |
| `OLLAMA_DEBUG` | Sets the system log level for <br>troubleshooting. <br><br>Default: `0` (Info). | `1` (Debug) |

</template>
<template #vLLM>

**!!Keep tuning: the following are some placeholders**

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `--max-model-len` | Sets the maximum context window size. If you hit out-of-memory errors, lower this value. | `8192` |
| `--gpu-memory-utilization` | Sets the memory utilization cap for the model. | `0.9` |
| `--tensor-parallel-size` | Sets the number of GPUs to use for tensor parallelism. | `1` |
</template>
<template #Llama-cpp>

**!!Keep tuning: the following are some placeholders**

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `-c` | Sets the maximum context window size in tokens. | `65536` |
| `-ngl` | Offloads model layers to the GPU. | `all` |
| `-fa` | Enables Flash Attention to reduce the memory footprint. | `on` |
</template>
<template #SGLang>

**!!Keep tuning: the following are some placeholders**

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `--context-length` | Sets the maximum context length for processing. | `32768` |
| `--mem-fraction-static` | Sets the fraction of GPU memory to allocate for static usage (model weights and KV cache). | `0.85` |
| `--max-running-requests` | Sets the maximum number of requests being processed concurrently. | `64` |
| `--reasoning-parser` | Configures the parser for reasoning models. | `qwen3` |
</template>
</tabs>

<!--
| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | Sets the default context window<br> size in tokens. <br><br>Default scales by VRAM:<ul><li>Less than 24G: 4096</li><li>Between 24G and 48G: 32768</li><li>48G and more: 262144</li></ul> | `8192` to `131072` |
| `OLLAMA_NUM_PARALLEL` | Sets the number of concurrent<br> requests processed per model. <br><br>Ollama determines this automatically<br> based on your available VRAM, <br>typically `1` or `4`.<br><br>Default: `0` | `1` |
| `OLLAMA_KV_CACHE_TYPE` | Sets the KV cache quantization type<br> to save video memory. <br><br>Default: `f16`. | `q8_0` (minor precision loss) or `q4_0` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention to optimize<br> memory efficiency during long-context<br> operations. <br><br>Default: `0` (Disabled). | `1` (Enabled) |
| `OLLAMA_MAX_LOADED_MODELS` | Sets the maximum number of models<br> kept loaded in memory simultaneously.<br>It automatically scales to roughly 3<br> models per available GPU<br><br>Default: `0`. | `1` |
| `OLLAMA_MAX_QUEUE` | Sets the maximum number of incoming<br> requests allowed in the processing queue. <br><br>Default: `512`. | `512` |
| `OLLAMA_KEEP_ALIVE` | Sets model resident duration in memory<br> after the last request. Use `-1` <br>for permanent retention. <br><br>Default: `5m`. | `30m` or `-1` |
| `OLLAMA_LOAD_TIMEOUT` | Sets the maximum duration to wait for<br> a model to finish loading before giving up. <br><br>Default: `5m`. | `5m` |
| `OLLAMA_GPU_OVERHEAD` | Sets the amount of video memory <br>in bytes reserved as a safety margin<br> overhead. <br><br>Default: `0`. | `0` |
| `OLLAMA_DEBUG` | Sets the system log level for <br>troubleshooting. <br><br>Default: `0` (Info). | `1` (Debug) |
-->

### Engine configuration examples

<tabs>
<template #Ollama>

Ollama pulls models automatically using native model tags.

**Chat Model Example**
```text
MODEL_SOURCE=ollama://qwen3.5:2b
MODEL_NAME=qwen3.5-2b
MODEL_MODE=chat
MODEL_SUPPORTS=supports_function_calling,supports_tool_choice
ENGINE_ARGS=OLLAMA_CONTEXT_LENGTH=8192
OLLAMA_REQUIRED_GPU_MEMORY=4096
```

**Embedding Model Example**
```text
MODEL_SOURCE=ollama://nomic-embed-text
MODEL_NAME=nomic-embed-text
MODEL_MODE=embedding
MODEL_SUPPORTS=
ENGINE_ARGS=OLLAMA_KEEP_ALIVE=-1
OLLAMA_REQUIRED_GPU_MEMORY=4096
```

</template>
<template #vLLM>

```text
MODEL_SOURCE=hf://Qwen/Qwen3.5-2B
MODEL_NAME=Qwen/Qwen3.5-2B
MODEL_MODE=chat
MODEL_SUPPORTS=supports_reasoning
ENGINE_ARGS=--max-model-len 8192 --gpu-memory-utilization 0.9 --tensor-parallel-size 1
VLLM_REQUIRED_GPU_MEMORY=10Gi
```

</template>
<template #Llama-cpp>

```text
MODEL_SOURCE=hf://unsloth/Qwen3.5-2B-GGUF --include Qwen3.5-2B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.5-2B-GGUF --include mmproj-F16.gguf
MODEL_NAME=unsloth/Qwen3.5-2B-GGUF:UD-Q4_K_XL
MODEL_MODE=chat
MODEL_SUPPORTS=supports_vision,supports_reasoning
ENGINE_ARGS=-c 65536 -ngl all -fa on
LLAMACPP_REQUIRED_GPU_MEMORY=8192
```

</template>
<template #SGLang>

```text
MODEL_SOURCE=hf://Qwen/Qwen3.5-2B
MODEL_NAME=Qwen/Qwen3.5-2B
MODEL_MODE=chat
MODEL_SUPPORTS=supports_function_calling,supports_tool_choice,supports_reasoning,supports_thinking
ENGINE_ARGS=--context-length 32768 --mem-fraction-static 0.85 --max-running-requests 64 --reasoning-parser qwen3
SGLANG_REQUIRED_GPU_MEMORY=8192
```
</template>
</tabs>

## Monitor download and initialization status

Track model downloads, monitor performance metrics, and retrieve the API connection details through the built-in instance dashboard.

1. Locate your deployment on the **Instances** panel on the right side of the base app details page. 
2. Select **Open** once the status indicates the base service is running. The **llm-init** page of the model instance is opened.
3. On the **STATUS** tab, verify your deployment progress:
    - **DOWNLOAD**: Displays real-time download percentages, speeds, and an estimated arrival time (ETA).
    - **STATUS**: Tracks failures or retry counts. If a network drop happens or your model source syntax is incorrect, select **Retry** after fixing the error.
    - **ENGINE**: Displays the initialization state.

        - Verify that both tracking labels read **Engine alive: yes** and **Model exists: yes**. This confirms your engine is online and ready to accept requests.
        - Copy the model name and the OpenAI-compatible API base URL.

4. Go to the **CONFIG** tab to review operational limits, run performance probes, view benchmark history, or update your variables.

## Connect client apps to the model service

With your local model instance successfully running, other client apps can connect with it using standard OpenAI API patterns.

1. Open Olares Settings, and then go to **Applications** > **{Your-New-Model-Instance}** > **Shared entrance** > **{Engine} LLM API**.
2. Copy the endpoint URL, and enter it along with your defined `MODEL_NAME` into the model configuration section of the client app to connect.

## Uninstall model instances

1. Open the target base app from Market.
2. In the **Instances** section, locate the target model instance, click the drop-down arrow next to the operation button, and then click **Uninstall**.
