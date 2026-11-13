---
outline: [2, 3]
description: Run a larger MoE model in Colibri on Olares with less resident memory, turn meeting notes into action items, and export the conversation.
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, local LLM, GGUF, MoE, RAM, GPU memory, CPU inference, meeting notes, Olares Router
app_version: "0.2.34"
doc_version: "2.2"
doc_updated: "2026-11-12"
---

# Run a larger MoE model with Colibri on Olares

Colibri is a local AI inference engine with a web chat interface. It loads supported GGUF models. For mixture-of-experts (MoE) models, it can read expert weights from disk as needed, reducing the need to keep every expert in memory.

This can make larger models accessible when keeping all their weights in memory would require too much RAM. You can set the system RAM limit and GPU memory allocation. Setting GPU memory to 0 runs inference on the CPU. With fewer resources, the first response and subsequent generation can become much slower, especially with long inputs.

On Olares, you can reuse existing model files and access the chat interface through your Olares sign-in. This tutorial uses Qwen3.6-35B-A3B to turn meeting notes into action items, then exports the conversation to save the result.

Qwen3.6-35B-A3B is only an example for demonstrating the workflow. Colibri does not require this specific model. You can substitute another supported MoE model in GGUF format that fits your available resources. Memory use and response speed vary with the model, quantization, and context length.

## Learning objectives

In this tutorial, you will learn how to:

- Set RAM and GPU memory limits for your available resources and select an existing GGUF model.
- Organize meeting notes with a local model and export the conversation.
- Switch the chat connection to the default chat model in Olares Router.

## Prerequisites

- An amd64 device running Olares 1.12.7 or later, with AVX2 support.
- Colibri 0.2.34, with enough available resources to load your chosen model.
- An NVIDIA device on the host. GPU acceleration requires an NVIDIA GPU compatible with this Blackwell sm_120 build and a driver that supports CUDA 13.0. This build has been tested on an RTX 5090 Laptop GPU. It does not support acceleration on older NVIDIA GPU architectures.
- A downloaded GGUF model supported by Colibri. This tutorial uses the `UD-IQ4_XS` quantization of Qwen3.6-35B-A3B, whose weight file is about 16.5 GiB. Keep enough disk space for the complete weights even when using a smaller resident memory budget.

## Prepare your model files

Open Files and go to `Common/huggingface`. Make sure your GGUF file is in this directory or one of its subdirectories. You can reuse an existing Hugging Face cache. Keep its `snapshots` structure and relative links intact, without moving files or copying paths.

Keep any matching `mmproj` vision projector in the same directory as the main weights. Colibri displays it as a vision companion under its model, rather than as a second chat model. Finding a vision file does not enable image input. Check the loaded engine's reported capabilities.

Colibri reads the original weights without modifying them. Imported indexes and your model selection are saved in the application's own data directory. The package does not include or automatically download model weights.

Your disk must have room for the complete model files. Colibri reads expert weights as needed, so disk read speed can also affect response times when RAM is limited. Store your models on an SSD where possible. More disk space lets you store larger models, but does not guarantee that they can run. Available RAM, GPU memory, and model compatibility still matter.

## Set RAM and GPU memory limits

During installation, enter the RAM limit and GPU memory allocation for your available resources. Leave the other settings at their defaults. This example uses a `32Gi` RAM limit and an `8Gi` GPU memory allocation. These are tested example values, not minimum requirements for every model.

| Setting | Example value | Description |
|---|---|---|
| `MEMORY_LIMIT` | `32Gi` | System RAM limit for Colibri's main container, not GPU memory |
| `GPU_MEMORY_LIMIT` | `8Gi` | Total GPU memory allocated to the application |

`MEMORY_LIMIT` and `GPU_MEMORY_LIMIT` are required at installation and have no preset values. Positive values must include `Mi` or `Gi`, such as `32Gi`, `48Gi`, or `49152Mi`. Do not enter `48GB` or a bare `48`. `1Gi` equals `1024Mi`.

Colibri converts 75% of the RAM limit into its engine budget, leaving headroom for the runtime, context, and file cache. For example, a `32Gi` limit gives the engine a budget of about 25 decimal GB. The limit does not preallocate that memory or increase the machine's physical RAM. A limit that is too low can still cause model loading or generation to fail.

The GPU allocation covers model weights, expert cache, and runtime buffers. It is not an additional expert-cache allowance. The minimum positive allocation is `4Gi`. Colibri keeps `2Gi` of the allocation for CUDA and temporary buffers, and uses the remainder for model placement. GPU memory is not necessarily filled immediately at startup.

Set GPU memory to `0` to disable CUDA and run inference on the CPU. Set it to `4Gi` or more to enable GPU acceleration. Both use the same installation entry, and Olares still records an NVIDIA device binding.

### Adjust resources after installation

To change the RAM limit, open Colibri's environment variable settings in Olares Settings and update `MEMORY_LIMIT`. Saving restarts the application and reloads the selected model. Export any conversation you want to keep first.

`GPU_MEMORY_LIMIT` cannot be edited directly after installation. To change the allocation, uninstall the application with its data retained, then reinstall and enter the new value. Enter `0` to switch to CPU inference. You do not need to download the weights in the shared directory again.

## Scan for and select a local model

1. Open Colibri from Launchpad. If no model is loaded, the **Local models** page opens. Use the language selector at the top to choose **English** if needed.
2. Click **Rescan** to find models in `Common/huggingface`. The list shows each model's name, main weight path, architecture, and size. Vision companions appear under their model. Split GGUF files form one entry and require every shard before loading.
3. Select **Qwen3.6-35B-A3B**. Unavailable entries explain why they cannot be selected, such as an architecture unsupported by the current Colibri engine.
4. Click **Load selected model**, review the confirmation message, then click **Confirm**.
5. Wait for **Model ready**, then click **Open chat**. Large models can take several minutes to load. Expand **Loading log** to check progress while you wait.

![Select a local model in Colibri, with its vision file grouped as a companion](/images/manual/use-cases/colibri-moe-models-en.jpg#bordered)

Your selection is saved in the application data. After an application restart, Colibri reloads the same model instead of choosing the first file found during a scan. If the saved file has been moved or deleted, rescan and select another model.

To switch models, click the local model icon on the left side of the chat page. Model management opens in a new tab. Wait for the current reply to finish and export any conversation you want to keep before selecting another model. Once the new model is ready, click **Open chat** to start a new conversation.

Click **Stop local model** and confirm to stop the inference process and release the RAM and GPU memory it was using. When GPU acceleration is enabled, Olares still retains the application's GPU allocation. You can select a model again later. The stopped state is also saved, so restarting the application does not automatically reload the model.

## Organize meeting notes with your local model

### Connect to the local engine

1. The following steps use English interface labels. To change the chat interface language, click the settings icon in the bottom-left corner and select **English** under **Settings** > **General** > **Language**.
2. Under **Settings** > **Connection**, check that **Connection service** is **Colibri local engine (default)**. Keep the automatically filled **API endpoint**.
3. Click **Probe server**. When **Engine reachable** appears, the connection is ready. If it fails, return to the local model page and check the loading status.
4. Open the **Model** tab and select the loaded local model. For this example, set **Max output tokens** to `512` and leave **Reasoning** off to keep the meeting summary brief.

`512` is only the maximum output length for this short-summary example, not a fixed model output limit. You can adjust it for your task, but the input, conversation history, and output must fit within the current context limit.

The web interface authenticates through your Olares sign-in. You do not need to enter an API key in Colibri.

### Generate a meeting summary

1. Click the **New chat** icon on the left to start a conversation.
2. Paste the following fictional meeting notes into the input box and click the send arrow. You can also substitute a short text of your own.

   ```text
   Turn the following meeting notes into a concise meeting summary.
   List confirmed decisions first, then use a table for tasks, owners,
   and deadlines. If the notes do not provide a detail, write "To be
   confirmed" instead of guessing.

   Meeting notes:
   The team decided to launch the updated help center on November 10.
   Alex will organize the frequently asked questions by November 6.
   Casey will check the page links by November 8.
   The product screenshots also need updating, but the owner and
   deadline have not been decided.
   ```

3. Wait for the reply to appear. Compare the action items with the original notes, especially the owners and deadlines. Check that missing details remain marked as "To be confirmed."
4. In the same conversation, enter "Rewrite these action items as a short message to the team" to turn the summary into a message you can use.

   ![Colibri chat interface showing a streamed reply from a local model](/images/manual/use-cases/colibri-moe-chat-en.jpg#bordered)

   The screenshot shows the same meeting task in the English chat interface.

This version provides a 16K context by default, shared by the input, conversation history, and output. For longer notes, work through one topic at a time, then continue with a shorter summary. Even when the input fits within the context limit, lower RAM or GPU memory allocations can increase waiting time.

### Export the conversation

1. After the reply finishes, click the **Export conversation** icon in the top-right corner.
2. Select **Markdown (.md)** to download the conversation to your computer. The file contains your input and the model's replies for later editing or sharing. You can also choose **JSON (.json)** for structured storage.

:::warning Export before refreshing
The current version does not retain chat history or generation settings after a page refresh. Export any conversation you want to keep before closing or refreshing the page.
:::

## Resource configuration and performance reference

For the Qwen3.6-35B-A3B `UD-IQ4_XS` model in this example, a `32Gi` RAM limit and an `8Gi` GPU memory allocation are a tested starting point. More RAM or GPU memory can generally reduce waiting time when resources allow. Lower allocations can still complete requests, but might not maintain the same speed.

The following measurements were made on Olares One with an RTX 5090 Laptop GPU, a 16K context limit, approximately 1K input tokens, and 128 output tokens. The tables show the second request at each setting, when caches might already be warm. Time to first token means the time from starting a request to receiving the first text fragment.

### GPU memory allocations

With the system RAM limit fixed at `48Gi`:

| GPU memory allocation | Time to first token | Generation speed |
|---|---:|---:|
| 0, CPU only | 8.57 s | 20.53 token/s |
| 4Gi | 5.89 s | 30.57 token/s |
| 8Gi | 1.87 s | 39.81 token/s |
| 12Gi | 0.98 s | 46.71 token/s |
| 16Gi | 0.73 s | 68.80 token/s |
| 20Gi | 0.25 s | 171.13 token/s |

At `20Gi`, all of this model's expert weights can reside in GPU memory, which explains the larger speed increase. Other models might not reach that point at the same allocation.

### System RAM limits

With the GPU memory allocation fixed at `8Gi`:

| RAM limit | Time to first token | Generation speed |
|---|---:|---:|
| 48Gi | 1.88 s | 42.92 token/s |
| 32Gi | 2.01 s | 35.30 token/s |
| 24Gi | 30.98 s | 22.80 token/s |
| 16Gi | 33.82 s | 20.45 token/s |
| 12Gi | 58.05 s | 16.32 token/s |
| 8Gi | 68.61 s | 11.18 token/s |

Both requests at every RAM setting completed without an out-of-memory termination or engine restart. The `32Gi` and `48Gi` settings kept all expert weights resident in RAM and produced the first response much sooner than the lower RAM settings.

The GPU and RAM comparisons were run separately, so results for the same configuration can vary. The machine and GPU were shared, with only two requests at each setting. Actual performance also depends on the model, input content, and other workloads.

These are container RAM limits. The host had approximately 94 GiB of total RAM, and shared file cache can remain between requests and containers. These results are not equivalent to running on a device with only 8 GiB or 12 GiB of physical RAM.

## Optional: Switch to Olares Router

Colibri can also use models provided through Router. Before you start, configure an available chat model in Router and set it as the default chat model on the **Default models** page. For background, see [Use Olares Router as your AI gateway](/use-cases/olares-router.md).

:::info Switching the connection does not stop the local engine
Changing the connection service only changes where chat requests go. It does not stop Colibri's local engine. When GPU acceleration is enabled, the engine can still use GPU memory. Before calling a Router model on the same GPU, make sure both services fit in available GPU memory. If they do not, choose a Router model service that does not compete with Colibri for that GPU.
:::

1. In Colibri, open **Settings** > **Connection**.
2. Change **Connection service** to **Olares Router** and keep the automatically filled address. You do not need to create or enter an API key.
3. Click **Probe server**, then select `default-chat` on the **Model** tab.
4. Click **New chat** and send the same meeting notes in a new conversation. Router's current default chat model generates the reply. If you change that default model in Router later, this connection follows the change.

Router replies do not appear in Colibri's local engine performance statistics. To return to local inference, select **Colibri local engine (default)** under **Connection service**, probe the server again, select the local model, and start a new conversation.

## Learn more

- [Connect AI applications to Olares models](/manual/best-practices/connect-ai-apps.md): Learn about Router model names, connection addresses, and authentication.
- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Manage model services and default chat models.
- [Colibri official documentation (this build)](https://github.com/JustVugg/colibri/blob/0a5504a1dc64f86fcd0fe7db298b0448c730e7e6/README.md): Read about model compatibility, how the engine works, and tuning.
