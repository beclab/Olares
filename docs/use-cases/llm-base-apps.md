---
outline: [2, 3]
description: Learn how to use the Engine Base applications on Olares to self-host large language models and run different inference engines by cloning the base apps.
---

# Host local large language models with Engine Base apps

Olares v1.12.6 introduces **Model Console**, a platform that manages the full lifecycle of local large language models (LLMs). This platform provides four Engine Base applications, each built on a different inference engine: **Ollama Engine Base**, **vLLM Engine Base**, **llama.cpp Engine Base**, and **SGLang Engine Base**.

Choose the base app for the engine you want, clone it to deploy a model, then run and manage that model from its dedicated console.

## Before you start

- Your Olares system has been upgraded to v1.12.6 or later.

## Locate Engine Base apps

1. Open Market and search for "Engine Base".

    Four engine base apps appear: vLLM Engine Base, SGLang Engine Base, Ollama Engine Base, and llama.cpp Engine Base.

    ![Engine Base apps in Market](/images/manual/olares/llm-base-apps.png#bordered)

2. Choose the engine base that fits your needs. Each one is optimized for a different inference scenario:

    | Engine Base | When to choose |
    | :--- | :--- |
    | **llama.cpp Engine Base** | Choose llama.cpp when you are running lightweight<br> GGUF models or deploying with limited GPU memory.<br>It is the recommended engine on Olares One. |
    | **Ollama Engine Base** | Choose Ollama when you want to get started quickly with<br> broad model compatibility. It pulls models automatically<br> using native model tags, making it ideal for chat and <br>embedding tasks. |
    | **SGLang Engine Base** | Choose SGLang when you need efficient structured<br> generation or advanced reasoning optimizations. |
    | **vLLM Engine Base** | Choose vLLM when you need high-throughput serving<br> of Hugging Face models under heavy concurrent load. |

## Create a new model instance

An Engine Base app serves as a template. To run a model, you must first clone the base app into an independent running instance.

1. Select the base app that matches your preferred inference engine, and then click **View** on it. For example, **llama.cpp Engine Base**.
2. Click **Create** to initialize a new instance.

    ![Create a model instance](/images/manual/olares/llm-base-apps-create-instance2.png#bordered)

3. Select the hardware accelerator for the instance to run on, and then click **Confirm**.
4. Specify the instance identity settings:

    - **New app name**: Enter a unique name for the instance. This name is displayed as the app name in Market and Settings. For example, `Qwen3.6-35B-A3B`.
    - **Shortcut name for [client]**: Enter a unique shortcut name for the instance. This name is displayed on the Launchpad. For example, `qwen3.6-35b-a3b`.

5. Click **Create** to proceed to the environment configuration.

## Configure engine environment variables

After creating the instance, the configuration window opens. Define where your engine pulls the model, how much memory it uses, and what capabilities it exposes to other client apps.

1. In the **Configure environment variables for [New-app-name]** window, fill in the following details according to the target model and engine.

    <tabs>
    <template #llama-cpp>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model.<br><br>Format: `hf://<repo> --include <file>.gguf`<br><br>To download multiple files, separate each entry with a comma:<br>`hf://<repo> --include <file1>.gguf,hf://<repo> --include <file2>.gguf`.<br><br>Example:<ul><li>Model page: `https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF`</li><li>`MODEL_SOURCE`: `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE` and use this format:<br> `<repo>:<quantization>` (one quantization per instance).<br><br>Example:<ul><li>`MODEL_SOURCE`: `hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li><li>`MODEL_NAME`: `unsloth/Qwen3.6-35B-A3B-GGUF:UD-Q4_K_XL`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> or **Thinking**. For an embedding model, select **None**. |
    | **ENGINE_ARGS** | Set the engine startup parameters. The context size (`-c`) is required.<br>Separate multiple parameters with spaces.<br><br>Example:<ul><li>`-c 65536`</li><li>`-c 65536 -ngl all`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **LLAMACPP<br>_REQUIRED<br>_GPU_MEMORY** | Enter the minimum GPU memory the instance needs to start, <br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li><li>In CPU mode, set it to `0`.</li></ul> |

    </template>
    <template #Ollama>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model.<br><br>Format: `ollama://<model>:<size-tag>`<br><br>Example:<ul><li>Model page: `https://ollama.com/library/qwen3.5`</li><li>`MODEL_SOURCE`: `ollama://qwen3.5:2b`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE`: Use the string after `ollama://`.<br><br>Example:<ul><li>`MODEL_SOURCE`: `ollama://qwen3.5:2b`</li><li>`MODEL_NAME`: `qwen3.5:2b`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> or **Thinking**. For an embedding model, select **None**. |
    | **ENGINE_ARGS** | Set the engine startup parameters. The context size is required. <br>Separate multiple parameters with spaces.<br><br>Example:<ul><li>`OLLAMA_CONTEXT_LENGTH=8192`</li><li>`OLLAMA_CONTEXT_LENGTH=8192 OLLAMA_KV_CACHE_TYPE=q8_0`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **OLLAMA<br>_REQUIRED<br>_GPU_MEMORY** | Enter the minimum GPU memory the instance needs to start,<br>in MB or Gi. For example, `8Gi` and `8192Mi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li><li>In CPU mode, set it to `0`.</li></ul> |

    </template>
    <template #vLLM-or-SGLang>

    | Variable | Description |
    | :--- | :--- |
    | **MODEL_SOURCE** | Specify where the engine pulls the model. Choose a repository<br>that contains `.safetensors` weight files.<br><br>Format: `hf://<repo>`<br><br>Example:<ul><li>Model page: `https://huggingface.co/Qwen/Qwen3.5-2B`</li><li>`MODEL_SOURCE`: `hf://Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_NAME** | Define the name that client apps use to call this instance.<br><br>Derive it from `MODEL_SOURCE`: Use the string after `hf://`.<br><br>Example:<ul><li>`MODEL_SOURCE`: `hf://Qwen/Qwen3.5-2B`</li><li>`MODEL_NAME`: `Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_MODE** | Select **Chat** or **Embedding**. |
    | **MODEL_SUPPORTS** | Select the capabilities the model supports: **Vision**, **Tools**,<br> or **Thinking**. For an embedding model, select **None**. |
    | **ENGINE_ARGS** | Set the engine startup parameters. The context size is required. <br>Separate multiple parameters with spaces.<br><br>Example:<ul><li>vLLM: `--max-model-len 65536`</li><li>SGLang: `--context-length 65536`</li></ul>For more arguments, see [Engine tuning arguments](#reference-engine-tuning-arguments). |
    | **VLLM/SGLANG<br>_REQUIRED_GPU<br>_MEMORY** | Enter the minimum GPU memory the instance needs to start, <br>in MB or Gi. For example, `20Gi`.<ul><li>In time slicing or exclusive mode, set it below your total VRAM.</li><li>In memory slicing mode, set it below your remaining VRAM.</li><li>In CPU mode, set it to `0`.</li></ul> |

    </template>
    </tabs>

2. Click **Confirm** to save the configuration and start the instance installation.

    An **Instances** panel appears on the right side of the page, showing the installation progress. Once the setup completes, the instance's operation button changes to **Open**, indicating that the base service is running. A model app with the same name also appears on the Launchpad.

    ![Model instance installed](/images/manual/olares/llm-base-model-instance-installed1.png#bordered)

    :::info
    Model instances created from Engine Base apps show a `From template` tag next to the app name. You can see this tag when viewing the app in Market or Settings.

    ![Model instance tag](/images/manual/olares/llm-base-model-instance-tag1.png#bordered){width=70%}
    :::

:::tip Update environment variables later
To change these variables after installation, go to Olares **Settings** > **Applications** > **[App-Name]** > **Manage environment variables**. Click the edit icon next to a variable, update its value, save your change, and then click **Apply**.
:::

### Reference: Engine tuning arguments

Use the `ENGINE_ARGS` variable to add custom settings that adjust memory usage, context limits, and processing behaviors. Separate multiple arguments with spaces. Select your inference engine below to view some commonly used tuning arguments.

<tabs>
<template #Llama-cpp>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `-c` | Sets the maximum context length<br> in tokens. | `65536` |
| `-ngl` | Offloads all model layers to the GPU<br> to avoid CPU-bound slowdowns. | `all` |
| `-fa` | Enables Flash Attention to speed up<br> attention computation. | `on` |
| `-ctk` / `-ctv` | Quantizes the KV Cache to 8-bit,<br> balancing GPU memory use and precision. | `q8_0` |

For other llama.cpp arguments, see the [official documentation](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).
</template>
<template #Ollama>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | Sets the default context window<br> size in tokens. <br><br>Default scales by VRAM:<ul><li>Less than 24G: 4096</li><li>Between 24G and 48G: 32768</li><li>48G and more: 262144</li></ul> | `8192` to `131072` |
| `OLLAMA_KEEP_ALIVE` | Sets how long the model stays resident in<br> GPU memory after the last request. When the<br> time expires, the weights are swapped to<br> system RAM.<ul><li>`-1` keeps it in GPU memory permanently.</li><li>`3m` keeps it for 3 minutes.</li></ul>Default: `-1`. | `-1` or `30m` |
| `OLLAMA_FLASH_ATTENTION` | Enables Flash Attention. It must be enabled<br> to use `OLLAMA_KV_CACHE_TYPE` for KV cache<br> quantization.<br><br>Default: `1`. | `1` (Enabled) |
| `OLLAMA_KV_CACHE_TYPE` | Sets the key-value (KV) cache quantization<br> type to save video memory. <br><br>Default: `f16`. | `q8_0` (minor precision loss) or `q4_0` |
| `OLLAMA_NUM_PARALLEL` | Sets the number of concurrent<br> requests processed per model. <br><br>Default: `1`. | `1` |

For other Ollama arguments, see the [official documentation](https://github.com/ollama/ollama/blob/main/docs/faq.mdx).
</template>
<template #SGLang>

| Argument | Purpose | Recommended |
| :--- | :--- | :--- |
| `--context-length` | Sets the maximum context length. | `65536` |
| `--mem-fraction-static` | Sets the fraction of GPU memory<br> pre-allocated for static usage, similar<br> to vLLM's `--gpu-memory-utilization`. | `0.85` |
| `--chunked-prefill-size` | Splits very long inputs into chunks so<br> they don't block the GPU for long,<br> keeping concurrent requests' streaming<br> smooth. | `4096` |

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

For other vLLM arguments, see the [official documentation](https://docs.vllm.ai/en/v0.17.0/configuration/engine_args/).
</template>
</tabs>

### Recommended models and parameters

The following per-engine recommendations are validated best practices. Use them as a starting point, then adjust for your hardware.

<tabs>
<template #Llama-cpp>

- **Recommended model**: [`unsloth/Qwen3.6-27B-GGUF`](https://huggingface.co/unsloth/Qwen3.6-27B-GGUF), quantized to `Q4_K_M`
- **MODEL_SOURCE**: `hf://unsloth/Qwen3.6-27B-GGUF --include Qwen3.6-27B-Q4_K_M.gguf`

    :::tip Multimodal models
    If the model has multimodal capabilities, include the `mmproj-F16.gguf` file in `MODEL_SOURCE`:

    `hf://unsloth/Qwen3.6-27B-GGUF --include Qwen3.6-27B-Q4_K_M.gguf,hf://unsloth/Qwen3.6-27B-GGUF --include mmproj-F16.gguf`
    :::

- **MODEL_NAME**: `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`
- **ENGINE_ARGS**: `-c 131072 -ngl all -fa on -ctk q8_0 -ctv q8_0`

</template>
<template #Ollama>

- **Recommended model**: `gemma4-26b` from the Ollama library, which defaults to Q4_K_M quantization
- **MODEL_SOURCE**: `ollama://gemma4-26b`
- **MODEL_NAME**: `gemma4-26b`
- **ENGINE_ARGS**: `OLLAMA_KEEP_ALIVE=-1 OLLAMA_CONTEXT_LENGTH=131072 OLLAMA_FLASH_ATTENTION=1 OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_NUM_PARALLEL=1`

</template>
<template #SGLang>

:::info
SGLang models can take a while to load before the engine reaches `RUNNING`.
:::

- **Recommended model**: [`cyankiwi/Ornith-1.0-9B-AWQ-FP8`](https://huggingface.co/cyankiwi/Ornith-1.0-9B-AWQ-FP8) from Hugging Face
- **MODEL_SOURCE**: `hf://cyankiwi/Ornith-1.0-9B-AWQ-FP8`
- **MODEL_NAME**: `cyankiwi/Ornith-1.0-9B-AWQ-FP8`
- **ENGINE_ARGS**: `--context-length 131072 --mem-fraction-static 0.85 --chunked-prefill-size 4096`

</template>
<template #vLLM>

:::info
vLLM models can take a while to load before the engine reaches `RUNNING`.
:::

- **Recommended model**: [`cyankiwi/Ornith-1.0-9B-AWQ-FP8`](https://huggingface.co/cyankiwi/Ornith-1.0-9B-AWQ-FP8) from Hugging Face
- **MODEL_SOURCE**: `hf://cyankiwi/Ornith-1.0-9B-AWQ-FP8`
- **MODEL_NAME**: `cyankiwi/Ornith-1.0-9B-AWQ-FP8`
- **ENGINE_ARGS**: `--max-model-len 131072 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 --max-num-batched-tokens 8192 --tool-call-parser qwen3_coder --reasoning-parser qwen3 --enable-prefix-caching --enable-auto-tool-choice`

</template>
</tabs>

## Monitor deployment and configure the model service

Open the built-in model console to track the model download, confirm the model and engine readiness, configure client access, and inspect GPU usage and performance.

1. Locate the model instance in the **Instances** panel on the Engine Base app details page, or find it on the Launchpad.
2. Open it to launch the dedicated model console.

    The console opens on the **Status** tab by default, and the model files start downloading automatically.

3. Under **Service status**, track the readiness of the model and the engine:

    - **Model**: Shows **Ready** after the files are downloaded and verified.
    - **Engine**: Shows **Running** after the inference service is online.

    ![Model console ready](/images/manual/olares/llm-base-model-console-status.png#bordered)

4. When the engine shows **Running**, configure how client apps reach the service.

    - **Connection source**: Select where the client runs.
        - **Apps in Olares**: For apps running in Olares.
        - **Devices on your network**: For devices on the same local network.
        - **Remote**: For access over the public internet, which requires you to enable the VPN in LarePass first.
    - **API format**: Select the API style that matches your client: **Ollama**, **OpenAI-Compatible**, or **Anthropic-Compatible**.
    - **Base URL**: Copy the URL to use for client app connections.
    - **Supported endpoints**: Expand this list to see every endpoint the selected API format exposes, with its HTTP method, path, and purpose.

5. Select the **Configuration** tab to review the model's details:

    ![Configuration tab in model console](/images/manual/olares/llm-base-model-console-config.png#bordered)

    - **Model**: Shows the model name, mode, and the capability tags.
    - **Parameters**: View the engine parameters. Expand **Advanced parameters** for the full set, and switch the view between **Form** and **Raw**.

6. In the **GPU residency** section, click **Detect**, and then:

    - **Check the mode**: Confirm where the model is running.
        - **Full GPU**: The entire model runs on the GPU. This is the fastest state, and is expected when you selected the GPU accelerator during installation.
        - **CPU** or **Split**: Part or all of the model runs on the CPU, which makes inference slower.
            - If you chose the CPU accelerator during installation, `CPU` is expected. 
            - If you chose the GPU accelerator, review your `[ENGINE]_REQUIRED_GPU_MEMORY` setting and engine arguments.
    - **Check the memory usage**: Review the **VRAM**, **KV cache used**, and **GPU memory utilization** to see how much memory the model occupies and how much memory is left for longer contexts or more concurrent requests.

7. In the **Performance** section, click **Run test** to measure two response-speed metrics.

    Use them to compare quantization levels, context sizes, or engine arguments, and to verify that a change actually improved speed before you use it:

    - **TTFT** (Time To First Token): How long you wait before the first word appears. A lower value means the model responds faster.
    - **Cold start**: How long the engine takes to load the model from scratch, for example after a restart. A lower value means the model is ready to serve sooner.

## Connect client apps to the model service

Once the model instance is running, any client app that speaks the OpenAI-compatible API can connect to it through the base URL.

The following example uses [OpenCode](./opencode.md) as the client.

1. In the model console, go to the **Status** tab. Under **Service status**:

    - **Connection source**: Select **Apps in Olares**, because OpenCode runs in Olares.
    - **API format**: Select **OpenAI-Compatible**.
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
