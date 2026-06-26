---
outline: [2, 3]
description: Learn how to use the LLM Base applications on Olares to self-host large language models and run different inference engines by cloning the base apps.
---

# Host local large language models with LLM Base apps

Olares V1.12.6 introduces the local hosting and management platform for large language models (LLMs), a self-hosting solution powered by the `llm-init` project. This platform provides four LLM Base applications, each for one inference engine: **Ollama LLM Base**, **vLLM LLM Base**, **llama.cpp LLM Base**, and **SGLang LLM Base**. Select the base app for the engine you want, use it to deploy different models, and then monitor model performance through a dedicated console. 

## Before you start

- Your Olares system has been upgraded to V1.12.6 or later.

## Locate LLM Base apps

1. Open Market and search for "LLM Base". Four base apps appear: vLLM LLM Base (llm-init), SGLang LLM Base (llm-init), Ollama LLM Base (llm-init), and llama.cpp LLM Base (llm-init).

    ![LLM Base apps in Market](/images/manual/olares/llm-base-apps.png#bordered)

2. Each base app is optimized for a different inference scenario. Choose one based on your model source, performance needs, and hardware.

    | Base app | When to choose |
    | :--- | :--- |
    | **llama.cpp LLM Base (llm-init)** | Choose llama.cpp when you are running lightweight<br> GGUF models or deploying with limited GPU memory. |    
    | **Ollama LLM Base (llm-init)** | Choose Ollama when you want to get started quickly with<br> broad model compatibility. It pulls models automatically<br> using native model tags, making it ideal for chat and embedding tasks. |
    | **vLLM LLM Base (llm-init)** | Choose vLLM when you need high-throughput serving <br>of Hugging Face models under heavy concurrent load. |
    | **SGLang LLM Base (llm-init)** | Choose SGLang when you need efficient structured<br> generation or advanced reasoning optimizations. |

## Create a new model instance

An LLM Base app serves as a template. To run a model, you must first clone the base app into an independent running instance.

1. Select the base app that matches your preferred inference engine, and then click **View** on it. For example, **llama.cpp LLM Base (llm-init)**.
2. Click **Create** to initialize a new instance.

    ![Create a model instance](/images/manual/olares/llm-base-apps-create-instance1.png#bordered)

3. Specify the instance identity settings:

    - **New app name**: Enter a unique name for the instance. This name is displayed as the app name in Market and Settings. For example, `Qwen3.6-35B-A3B`.
    - **Shortcut name for {client}**: Enter a unique shortcut name for the instance. This name is displayed on the Launchpad. For example, `qwen3.6-35b-a3b`.

4. Click **Create** to proceed to the environment configuration.

## Configure engine environment variables

After creating the instance, the configuration window opens. Define where your engine pulls the model, how much memory it uses, and what capabilities it exposes to other client apps.

1. In the **Configure environment variables for {New-app-name}** window, fill in the following details according to the target model and engine:

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model.<br><br>The format depends on the selected engine:<ul><li>**Ollama**: `ollama://<model>:<size>`<br>Example: `ollama://qwen3.5:2b`</li><li>**llama.cpp**: `hf://<repo> --include <file>.gguf`<br>Example: `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li><li>**vLLM** / **SGLang**: `hf://<repo>`<br>Example: `hf://Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br> Derive it from `MODEL_SOURCE` per engine:<ul><li>**Ollama**: Use the string after `ollama://`.<br>`MODEL_SOURCE`: `ollama://qwen3.5:2b`<br>`MODEL_NAME`: `qwen3.5:2b`</li><li>**llama.cpp**: Use the repo name plus the quantization tag (one quantization per instance).<br>`MODEL_SOURCE`: `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`<br>`MODEL_NAME`: `unsloth/Qwen3.6-35B-A3B-GGUF:UD-Q4_K_XL`</li><li>**vLLM** / **SGLang**: Use the string after `hf://`.<br>`MODEL_SOURCE`: `hf://Qwen/Qwen3.5-2B`<br>`MODEL_NAME`: `Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br>**Thinking**, or **Embedding**. |
    | **ENGINE_ARGS** | Specify the engine startup parameters, separated by spaces.<br><br>The format depends on the engine:<ul><li>**Ollama**: `OLLAMA_CONTEXT_LENGTH=8192`</li><li>**llama.cpp**: `-c 65536 -ngl all`</li><li>**vLLM**: `--max-model-len 65536`</li><li>**SGLang**: `--context-length 65536`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **{ENGINE}_REQUIRED<br>_GPU_MEMORY** | Enter the minimum GPU memory the instance needs to start,<br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li></ul> |

2. Click **Confirm** to save the configuration and start the instance installation.

    An **Instances** panel appears on the right side of the page, showing the installation progress. Once the setup completes, the instance's operation button changes to **Open**, indicating that the base service is running. A model app with the same name also appears on the Launchpad.

    ![Model instance installed](/images/manual/olares/llm-base-model-instance-installed1.png#bordered)

    :::info
    Model instances created from LLM Base apps show a `From template` tag next to the app name. You can see this tag when viewing the app in Market or Settings.

    ![Model instance tag](/images/manual/olares/llm-base-model-instance-tag1.png#bordered){width=70%}   
    :::

:::tip Update variables later
To change these variables after installation, go to Olares **Settings** > **Applications** > **[App-Name]** > **Manage environment variables**. Click the edit icon next to a variable, update its value, save your change, and then click **Apply**.
:::

### Reference: Engine tuning arguments

Use the `ENGINE_ARGS` variable to add custom settings that adjust memory usage, context limits, and processing behaviors. Separate multiple arguments with spaces. Select your inference engine below to view the available tuning arguments.

<tabs>
<template #Ollama>

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | Sets the default context window<br> size in tokens. <br><br>Default scales by VRAM:<ul><li>Less than 24G: 4096</li><li>Between 24G and 48G: 32768</li><li>48G and more: 262144</li></ul> | `8192` to `131072` |
| `OLLAMA_KEEP_ALIVE` | Sets model resident duration in memory<br> after the last request. Use `-1` <br>for permanent retention. <br><br>Default: `5m`. | `30m` or `-1` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention to optimize<br> memory efficiency during long-context<br> operations. <br><br>Default: `0` (Disabled). | `1` (Enabled) |
| `OLLAMA_KV_CACHE_TYPE` | Sets the KV cache quantization type<br> to save video memory. <br><br>Default: `f16`. | `q8_0` (minor precision loss) or `q4_0` |
| `OLLAMA_NUM_PARALLEL` | Sets the number of concurrent<br> requests processed per model. <br><br>Ollama determines this automatically<br> based on your available VRAM, <br>typically `1` or `4`.<br><br>Default: `0` | `1` |

For other Ollama arguments, see the [official documentation](https://github.com/ollama/ollama/blob/main/docs/faq.mdx).
</template>
<template #vLLM>

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `--max-model-len` | Sets the maximum context length.<br>Lower it if you hit out-of-memory errors. | `65536` |
| `--gpu-memory-utilization` | Sets the fraction of GPU memory the<br> vLLM engine may use. | `0.9` |
| `--tensor-parallel-size` | Sets the tensor-parallel size, that is,<br> how many GPUs split and run one<br> model together. | `1` |
| `--max-num-batched-tokens` | Caps the number of tokens processed<br> per batch, preventing sharp latency<br> spikes. | `8192` |
| `--enable-prefix-caching` | Caches the KV Cache of shared prompt<br> prefixes and reuses it across requests. | Enabled |
| `--kv-cache-dtype` | Sets the KV Cache data type. Using<br> `fp8` raises throughput while preserving<br> quality. <br><br>Accepted values: `auto`, `bfloat16`,<br> `fp8`, `fp8_ds_mla`, `fp8_e4m3`,<br> `fp8_e5m2`, `fp8_inc`. <br><br>With `auto` (default), the KV Cache type<br> matches the model weights (usually<br> `float16` or `bfloat16`). | `fp8` |

For other vLLM arguments, see the [official documentation](https://docs.vllm.ai/en/v0.17.0/configuration/engine_args/).
</template>
<template #Llama-cpp>

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `-c` | Sets the maximum context length<br> in tokens. | `65536` |
| `-ngl` | Offloads all model layers to the GPU<br> to avoid CPU-bound slowdowns. | `all` |
| `-fa` | Enables Flash Attention to speed up<br> attention computation. | `on` |
| `-ctk` / `-ctv` | Quantizes the KV Cache to 8-bit,<br> balancing GPU memory use and precision. | `q8_0` |
| `--spec-type` | Enables MTP (speculative decoding). | `draft-mtp` |
| `--spec-draft-n-max` | Sets the maximum number of tokens<br> the drafter guesses ahead per<br> speculative step. | `3` |

For other llama.cpp arguments, see the [official documentation](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).
</template>
<template #SGLang>

| Argument | Purpose | Recommended Example |
| :--- | :--- | :--- |
| `--context-length` | Sets the maximum context length. | `65536` |
| `--mem-fraction-static` | Sets the fraction of GPU memory<br> pre-allocated for static usage, similar<br> to vLLM's `--gpu-memory-utilization`. | `0.85` |
| `--chunked-prefill-size` | Splits very long inputs into chunks so<br> they don't block the GPU for long,<br> keeping concurrent requests' streaming<br> smooth. | `4096` |
| `--reasoning-parser` | Separates chain-of-thought output:<br> writes the model's reasoning to the<br> `reasoning_content` field and the final<br> answer to the `content` field. Set it<br> to match the model. | `gpt-oss` |
| `--tool-call-parser` | Enables parsing of function-call<br> (tool use) output. Set it to match<br> the model. | `gpt-oss` |

For other SGLang arguments, see the [official documentation](https://docs.sglang.io/docs/advanced_features/server_arguments).
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

### Sample engine configurations

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
MODEL_MODE=supports_embedding
MODEL_SUPPORTS=embedding
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

## Monitor deployment and configure settings

Track model downloads, verify engine readiness, and manage operational parameters through the built-in model console.

1. Locate the model instance in the **Instances** panel on the LLM Base app details page, or find it on the Launchpad.
2. Open it to launch the dedicated model console.

    The console opens on the **Status** tab by default. The model files are start downloading automatically.

    ![Model console status tab](/images/manual/olares/llm-base-model-console-status.png#bordered)

3. Tracks the readiness of the model and the engine:

    - **Model**: Shows `READY` once the files are downloaded and verified. Copy the **Model name** to use for client app connections.
    - **Engine**: Shows `RUNNING` once the inference service is online. Configure how client apps reach it:
        - **WHO IS CALLING**: Select who can access the API, **Apps in Olares**, **Devices in LAN**, or **Remote**.
        - **WHAT API FORMAT**: Select the API format. The available options depend on the engine, for example **OpenAI-Compatible**, **Anthropic-Compatible**, or **Ollama**.
        - **Base URL**: Copy this URL to use for client app connections.
        - **Supported Endpoints**: Expand to see the available API endpoints.

    ![Model console ready](/images/manual/olares/llm-base-model-console-ready.png#bordered)

4. Select the **Config** tab to review the model's capabilities and parameters, check how it sits on the GPU, and measure its performance.

    ![Model console, config page](/images/manual/olares/llm-base-model-console-config.png#bordered)

    - **Model card** (top): Shows the model name, mode (**Chat** or **Embedding**), and the capability tags the instance exposes, such as `function_calling`, `parallel_function_calling`, `reasoning`, `reasoning_effort`, and `tool_choice`.
    - **Parameters**: View the engine parameters. Expand **Advanced parameters** for the full set, and use the **Form** / **Raw** toggle to switch the view.
    - **GPU Residency**: Confirm whether the model is actually running on the GPU. Click **Detect** to refresh, then check **Mode**:
        - `full GPU`: All layers run on the GPU. This is the expected, fastest state.
        - `partial` or `cpu_only`: Part or all of the model fell back to the CPU, which makes inference much slower. On a GPU host this usually means an environment mis-mount, so review your `{ENGINE}_REQUIRED_GPU_MEMORY` setting and engine arguments.

        The panel also reports **VRAM**, **KV cache used**, and **GPU mem util**, so you can see how much memory the model occupies and how much headroom is left for longer contexts or more concurrent requests.
    - **Performance**: Click **Run test** to benchmark responsiveness:
        - **TTFT** (Time To First Token): How long a user waits before the first word appears. Lower means a snappier experience.
        - **perf.code**: How long the engine takes to load the model from scratch, for example after a restart.

        Use these numbers to compare quantization levels, context sizes, or engine arguments, and to confirm that a change actually improved speed before you rely on it.

## Connect client apps to the model service

Once the model instance is running, any client app that speaks the OpenAI-compatible API can connect to it through the base URL. 

The following example uses [OpenCode](../../../use-cases/opencode.md) as the client.

1. In the model console, go to the **Status** tab. Under **Service status**:

    - **WHO IS CALLING**: Select **Apps in Olares**, because OpenCode runs inside Olares.
    - **WHAT API FORMAT**: Select **OpenAI-Compatible**.
    - Copy the **Base URL** and note down the **Model name**.

2. In OpenCode, click <i class="material-symbols-outlined">settings</i> in the bottom-left corner, select **Providers**, then scroll down and select **Connect** next to **Custom Provider**.

3. Enter the following details:

    - **Provider ID**: A unique identifier for this provider. For example, `olares-llm`.
    - **Display name**: The name shown in the provider list. For example, `Olares LLM`.
    - **Base URL**: The **Base URL** you copied from the model console.
    - **Models**:
        - **Model ID**: Your `MODEL_NAME`. For example, `Qwen3.6-35B-A3B`.
        - **Display Name**: The name shown for this model. For example, `Qwen3.6 35B A3B`.

4. Click **Submit** to save the configuration. The provider appears in the provider list.
5. Run a task to test the connection. This example uses the Olares skill to upload and deploy an app to Olares.

    a. At the top, click the **Search** field and select **Toggle terminal** to open a terminal.

    b. Log in to the Olares CLI. Replace `alice123@olares.com` with your own Olares ID: `olares-cli profile login --olares-id alice123@olares.com`.

    c. When prompted, type your Olares password and press **Enter**. The input stays hidden.

    d. Below the chat box, select **Big Pickle** to open the model selector, and select **Qwen3.6 35B A3B** from the list.

    e. Send a task:

    ```text
    Upload and deploy this app to Olares:
    https://github.com/chandruk4321/dockerize-static-web-project
    ```

    f. Respond to OpenCode's questions, decisions, and approvals until the task finishes.

    ![Running the task in OpenCode](/images/manual/olares/llm-base-model-inst-test.png#bordered)

    In this example, the Todo app is uploaded and deployed to **My Olares**. Open it from **My Olares** or the Launchpad to use the running app.

    ![Todo app deployed to My Olares](/images/manual/olares/llm-base-model-inst-task.png#bordered)

## Uninstall model instances

1. Open Market, go to **My Olares**, and then locate the model instance app.
2. Click the drop-down arrow next to the operation button, and then select **Uninstall**.
