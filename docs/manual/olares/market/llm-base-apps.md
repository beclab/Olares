---
outline: [2, 3]
description: Learn how to use the LLM Base applications on Olares to self-host large language models and run different inference engines by cloning the base apps.
---

# Host local large language models with LLM Base apps

Olares V1.12.6 introduces the local hosting and management platform for large language models (LLMs), a self-hosting solution powered by the `llm-init` project. This platform provides four LLM Base applications, each for one inference engine: **Ollama LLM Base**, **vLLM LLM Base**, **llama.cpp LLM Base**, and **SGLang LLM Base**. Select the base app for the engine you want, use it to deploy different models, and then monitor model performance through a dedicated console.

## Before you start

- Your Olares system has been upgraded to V1.12.6 or later.

## Locate LLM Base apps

1. Open Market and search for "LLM Base".

    Four base apps appear: vLLM LLM Base (llm-init), SGLang LLM Base (llm-init), Ollama LLM Base (llm-init), and llama.cpp LLM Base (llm-init).

    ![LLM Base apps in Market](/images/manual/olares/llm-base-apps.png#bordered)

2. Choose the base app that fits your needs. Each one is optimized for a different inference scenario:

    | Base app | When to choose |
    | :--- | :--- |
    | **llama.cpp LLM Base (llm-init)** | Choose llama.cpp when you are running lightweight<br> GGUF models or deploying with limited GPU memory.<br>It is the recommended engine on Olares One. |
    | **Ollama LLM Base (llm-init)** | Choose Ollama when you want to get started quickly with<br> broad model compatibility. It pulls models automatically<br> using native model tags, making it ideal for chat and embedding tasks. |
    | **SGLang LLM Base (llm-init)** | Choose SGLang when you need efficient structured<br> generation or advanced reasoning optimizations. |
    | **vLLM LLM Base (llm-init)** | Choose vLLM when you need high-throughput serving<br> of Hugging Face models under heavy concurrent load. |

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

1. In the **Configure environment variables for [New-app-name]** window, fill in the following details according to the target model and engine.

    <tabs>
    <template #llama-cpp>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model.<br><br>Format: `hf://<repo> --include <file>.gguf`.<br>To download more than one file, separate each entry with a comma.<br><br>Example:<ul><li>Model page: `https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF`</li><li>`MODEL_SOURCE` (single file): `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li><li>`MODEL_SOURCE` (multiple files): `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.6-35B-A3B-GGUF --include mmproj-F16.gguf`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE` and use this format:<br> `<repo>:<quantization>` (one quantization per instance).<br><br>Example:<ul><li>`MODEL_SOURCE`: `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li><li>`MODEL_NAME`: `unsloth/Qwen3.6-35B-A3B-GGUF:UD-Q4_K_XL`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> **Thinking**, or **Embedding**. |
    | **ENGINE_ARGS** | Specify the engine startup parameters, separated by spaces.<br><br>Example:<ul><li>`-c 65536`</li><li>`-c 65536 -ngl all`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **LLAMACPP<br>_REQUIRED<br>_GPU_MEMORY** | Enter the minimum GPU memory the instance needs to start, <br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li></ul> |

    </template>
    <template #Ollama>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model.<br><br>Format: `ollama://<model>:<size-tag>`.<br><br>Example:<ul><li>Model page: `https://ollama.com/library/qwen3.5`</li><li>`MODEL_SOURCE`: `ollama://qwen3.5:2b`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE`: Use the string after `ollama://`.<br><br>Example:<ul><li>`MODEL_SOURCE`: `ollama://qwen3.5:2b`</li><li>`MODEL_NAME`: `qwen3.5:2b`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> **Thinking**, or **Embedding**. |
    | **ENGINE_ARGS** | Specify the engine startup parameters, separated by spaces.<br><br>Example:<ul><li>`OLLAMA_CONTEXT_LENGTH=8192`</li><li>`OLLAMA_CONTEXT_LENGTH=8192 OLLAMA_KV_CACHE_TYPE=q8_0`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **OLLAMA<br>_REQUIRED<br>_GPU_MEMORY** | Enter the minimum GPU memory the instance needs to start,<br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li></ul> |

    </template>
    <template #vLLM-or-SGLang>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model. Choose a repository<br>that contains `.safetensors` weight files.<br><br>Format: `hf://<repo>`.<br><br>Example:<ul><li>Model page: `https://huggingface.co/Qwen/Qwen3.5-2B`</li><li>`MODEL_SOURCE`: `hf://Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE`: Use the string after `hf://`.<br><br>Example:<ul><li>`MODEL_SOURCE`: `hf://Qwen/Qwen3.5-2B`</li><li>`MODEL_NAME`: `Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> **Thinking**, or **Embedding**. |
    | **ENGINE_ARGS** | Specify the engine startup parameters, separated by spaces.<br><br>Example:<ul><li>vLLM: `--max-model-len 65536`</li><li>SGLang: `--context-length 65536`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **VLLM/SGLANG<br>_REQUIRED_GPU<br>_MEMORY** | Enter the minimum GPU memory the instance needs to start, <br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li></ul> |

    </template>
    </tabs>

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
<template #Llama-cpp>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `-c` | Sets the maximum context length<br> in tokens. | `65536` |
| `-ngl` | Offloads all model layers to the GPU<br> to avoid CPU-bound slowdowns. | `all` |
| `-fa` | Enables Flash Attention to speed up<br> attention computation. | `on` |
| `-ctk` / `-ctv` | Quantizes the KV Cache to 8-bit,<br> balancing GPU memory use and precision. | `q8_0` |
| `--spec-type` | Enables MTP (Multi-Token Prediction)<br>speculative decoding. | `draft-mtp` |
| `--spec-draft-n-max` | Sets the maximum number of tokens<br> the drafter guesses ahead per<br> speculative step. | `3` |

For other llama.cpp arguments, see the [official documentation](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).
</template>
<template #Ollama>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | Sets the default context window<br> size in tokens. <br><br>Default scales by VRAM:<ul><li>Less than 24G: 4096</li><li>Between 24G and 48G: 32768</li><li>48G and more: 262144</li></ul> | `8192` to `131072` |
| `OLLAMA_KEEP_ALIVE` | Sets how long the model stays resident in<br> GPU memory after the last request. When the<br> time expires, the weights are swapped to<br> system RAM.<ul><li>`-1` keeps it in GPU memory permanently.</li><li>`3m` keeps it for 3 minutes.</li></ul>Default: `-1`. | `-1` or `30m` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention. It must be enabled<br> to use `OLLAMA_KV_CACHE_TYPE` for KV cache<br> quantization.<br><br>Default: `1`. | `1` (Enabled) |
| `OLLAMA_KV_CACHE_TYPE` | Sets the key-value (KV) cache quantization<br> type to save video memory. <br><br>Default: `f16`. | `q8_0` (minor precision loss) or `q4_0` |
| `OLLAMA_NUM_PARALLEL` | Sets the number of concurrent<br> requests processed per model。 <br><br>Default: `1`. | `1` |

For other Ollama arguments, see the [official documentation](https://github.com/ollama/ollama/blob/main/docs/faq.mdx).
</template>
<template #SGLang>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `--context-length` | Sets the maximum context length. | `65536` |
| `--mem-fraction-static` | Sets the fraction of GPU memory<br> pre-allocated for static usage, similar<br> to vLLM's `--gpu-memory-utilization`. | `0.85` |
| `--chunked-prefill-size` | Splits very long inputs into chunks so<br> they don't block the GPU for long,<br> keeping concurrent requests' streaming<br> smooth. | `4096` |
| `--reasoning-parser` | Separates chain-of-thought output:<br> writes the model's reasoning to the<br> `reasoning_content` field and the final<br> answer to the `content` field. Set it<br> to match the model. | `gpt-oss` |
| `--tool-call-parser` | Enables parsing of function calling output.<br> Set it to match the model. | `gpt-oss` |

For other SGLang arguments, see the [official documentation](https://docs.sglang.io/docs/advanced_features/server_arguments).
</template>
<template #vLLM>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `--max-model-len` | Sets the maximum context length. | `65536` |
| `--gpu-memory-utilization` | Sets the fraction of GPU memory that the<br> vLLM engine uses. | `0.9` |
| `--tensor-parallel-size` | Sets the tensor-parallel size, that is,<br> how many GPUs split and run one<br> model together. | `1` |
| `--max-num-batched-tokens` | Limits the maximum number of tokens<br> processed per batch, so response time<br> stays stable when a very long request<br> arrives. | `8192` |
| `--enable-prefix-caching` | Caches the KV Cache of shared prompt<br> prefixes and reuses it across requests. | Enabled |
| `--kv-cache-dtype` | Sets the KV Cache data type. Using<br> `fp8` raises throughput while preserving<br> quality. <br><br>Accepted values: `auto`, `bfloat16`,<br> `fp8`, `fp8_ds_mla`, `fp8_e4m3`,<br> `fp8_e5m2`, `fp8_inc`. <br><br>Default: `auto`. | `fp8` |

For other vLLM arguments, see the [official documentation](https://docs.vllm.ai/en/v0.17.0/configuration/engine_args/).
</template>
</tabs>

### Recommended models and parameters

The following per-engine recommendations are validated best practices. Use them as a starting point, then adjust for your hardware.

<tabs>
<template #Llama-cpp>

For the best speed, pick an MTP model so the engine can apply speculative decoding.

- **Recommended model**: [`unsloth/gemma-4-26B-A4B-it-GGUF`](https://huggingface.co/unsloth/gemma-4-26B-A4B-it-GGUF), an MTP model, and the `MODEL_SOURCE` must include the `mtp-gemma-4-26B-A4B-it.gguf` file
- **MODEL_SOURCE**: `hf://unsloth/gemma-4-26B-A4B-it-GGUF --include gemma-4-26B-A4B-it-UD-Q4_K_XL.gguf,hf://unsloth/gemma-4-26B-A4B-it-GGUF --include mmproj-F16.gguf,hf://unsloth/gemma-4-26B-A4B-it-GGUF --include mtp-gemma-4-26B-A4B-it.gguf`
- **MODEL_NAME**: `unsloth/gemma-4-26B-A4B-it-GGUF:UD-Q4_K_XL`
- **ENGINE_ARGS**: `-c 65536 -ngl all -fa on -ctk q8_0 -ctv q8_0 --spec-type draft-mtp --spec-draft-n-max 3`

</template>
<template #Ollama>

- **Recommended model**: `gemma4-26b` from the Ollama library, which defaults to Q4_K_M quantization
- **MODEL_SOURCE**: `ollama://gemma4-26b`
- **MODEL_NAME**: `gemma4-26b`
- **ENGINE_ARGS**: `OLLAMA_KEEP_ALIVE=-1 OLLAMA_CONTEXT_LENGTH=65536 OLLAMA_FLASH_ATTENTION=1 OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_NUM_PARALLEL=1`

</template>
<template #SGLang>

:::info
SGLang models can take a while to load before the engine reaches `RUNNING`.
:::

- **Recommended model**: [`gpt-oss-20b`](https://huggingface.co/openai/gpt-oss-20b) from Hugging Face
- **MODEL_SOURCE**: `hf://openai/gpt-oss-20b`
- **MODEL_NAME**: `openai/gpt-oss-20b`
- **ENGINE_ARGS**: `--context-length 65536 --mem-fraction-static 0.85 --chunked-prefill-size 4096 --reasoning-parser gpt-oss --tool-call-parser gpt-oss`

</template>
<template #vLLM>

:::info
vLLM models can take a while to load before the engine reaches `RUNNING`.
:::

- **Recommended model**: [`gpt-oss-20b`](https://huggingface.co/openai/gpt-oss-20b) from Hugging Face
- **MODEL_SOURCE**: `hf://openai/gpt-oss-20b`
- **MODEL_NAME**: `openai/gpt-oss-20b`
- **ENGINE_ARGS**: `--max-model-len 65536 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 --max-num-batched-tokens 8192 --enable-prefix-caching --kv-cache-dtype fp8`

</template>
</tabs>

## Monitor deployment and configure the model service

Open the built-in model console to track the model download, confirm the model and engine readiness, configure client access, and inspect GPU usage and performance.

1. Locate the model instance in the **Instances** panel on the LLM Base app details page, or find it on the Launchpad.
2. Open it to launch the dedicated model console.

    The console opens on the **Status** tab by default, and the model files start downloading automatically.

    ![Model console status tab](/images/manual/olares/llm-base-model-console-status.png#bordered)

3. Under **Service status**, track the readiness of the model and the engine:

    - **Model**: Shows `READY` after the files are downloaded and verified.
    - **Engine**: Shows `RUNNING` after the inference service is online.

    ![Model console ready](/images/manual/olares/llm-base-model-console-ready.png#bordered)

4. When the engine shows `RUNNING`, configure how client apps reach the service.

    - **WHO IS CALLING**: Select where the client runs.
        - **Apps in Olares**: For apps running in Olares.
        - **Devices in LAN**: For devices on the same local network.
        - **Remote**: For access over the public internet, which requires you to enable the VPN in LarePass first.
    - **WHAT API FORMAT**: Select the API style your client requires.
        - **OpenAI-Compatible**: Exposes OpenAI-style endpoints, such as `/v1/chat/completions` and `/v1/embeddings`.
        - **Anthropic-Compatible**: Exposes Anthropic Messages endpoints, such as `/v1/messages`.
    - **Base URL**: Copy the URL to use for client app connections.
    - **Supported Endpoints**: Expand to see the endpoints available for the selected API format.

5. Select the **Config** tab to review the model's details:

    ![Config tab in model console](/images/manual/olares/llm-base-model-console-config.png#bordered)

    - **Model**: Shows the model name, mode, and the capability tags the instance exposes.
    - **Parameters**: View the engine parameters. Expand **Advanced parameters** for the full set, and switch the view between **Form** and **Raw**.

6. In the **GPU Residency** section, click **Detect**, and then:

    - **Check the mode**: Confirm the model runs on the GPU.
        - `full GPU`: All layers run on the GPU. This is the expected, fastest state.
        - `partial` or `cpu_only`: Part or all of the model fell back to the CPU, which makes inference much slower. On a GPU host, this usually means an environment mis-mount; review your `{ENGINE}_REQUIRED_GPU_MEMORY` setting and engine arguments.
    - **Check the memory usage**: Review the **VRAM**, **KV cache used**, and **GPU mem util** to see how much memory the model occupies and how much headroom is left for longer contexts or more concurrent requests.

7. In the **Performance** section, click **Run test** to measure two response-speed metrics. Use them to compare quantization levels, context sizes, or engine arguments, and to verify that a change actually improved speed before you use it:

    - **TTFT** (Time To First Token): How long you wait before the first word appears. A lower value means the model responds faster.
    - **perf.cold**: How long the engine takes to load the model from scratch, for example after a restart. A lower value means the model is ready to serve sooner.

## Connect client apps to the model service

Once the model instance is running, any client app that speaks the OpenAI-compatible API can connect to it through the base URL.

The following example uses [OpenCode](../../../use-cases/opencode.md) as the client.

1. In the model console, go to the **Status** tab. Under **Service status**:

    - **WHO IS CALLING**: Select **Apps in Olares**, because OpenCode runs in Olares.
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
