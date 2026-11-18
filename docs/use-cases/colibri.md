---
outline: [2, 3]
description: Run a larger MoE model in Colibri on Olares, use it to review a short Python function, and export the chat.
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, local LLM, GGUF, MoE, RAM, GPU memory, CPU inference, Python code review, Olares Router
app_version: "0.2.35"
doc_version: "3.0"
doc_updated: "2026-11-18"
---

# Run a larger MoE model with Colibri on Olares

Colibri runs supported GGUF models locally and gives you a web chat interface. A mixture-of-experts (MoE) model uses only some of its experts for each reply. Colibri can keep other expert weights in the model file and read them from disk when needed, so you can try a larger model without keeping all its weights in memory.

You can choose how much system RAM and GPU memory to give Colibri, or use the CPU for inference. Less memory means more reading from disk and can make replies slower, especially with long inputs.

On Olares, you can reuse existing model files and access the chat interface through your Olares sign-in. This tutorial uses Qwen3.6-35B-A3B to find a bug in a short Python function, then exports the conversation. You can follow the same steps with another supported GGUF model.

## Learning objectives

In this tutorial, you will learn how to:

- Add a GGUF model file, set RAM and GPU memory limits, and choose an inference profile.
- Ask a local model to review a Python function and export the conversation.
- Switch the chat connection to the default chat model in Olares Router.

## Prerequisites

- An amd64 device running Olares 1.12.7 or later, with AVX2 support. See [Check and update Olares](/manual/olares/settings/update.md) if you need to update.
- An NVIDIA device on the host. GPU acceleration requires a Blackwell sm_120 GPU and a driver that supports CUDA 13.0. To check the GPU model, open Dashboard, select the GPU card under **Cluster's physical resources**, then open **Graphics management**. This build has been tested on an RTX 5090 Laptop GPU.
- Enough disk space for your model file. The Qwen3.6-35B-A3B `UD-IQ4_XS` file used below is about 16.5 GiB.

## Add a model file

Colibri scans `Common/huggingface` for GGUF files. It does not download model weights during installation. If your model is already in this folder or in a Hugging Face cache under it, continue to [install Colibri](#install-colibri-and-set-memory-limits). Keep the cache's `snapshots` structure and relative links intact.

To add the example model:

1. Download [Qwen3.6-35B-A3B-UD-IQ4_XS.gguf](https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF/blob/main/Qwen3.6-35B-A3B-UD-IQ4_XS.gguf) to your computer.
2. Open Files from Launchpad and go to **Common** > **huggingface**.
3. Drag the GGUF file into the Files window, or click the upload icon in the top-right corner. See [Upload files](/manual/olares/files/add-edit-download.md#upload-files).

:::tip Save the model directly to Olares
In the LarePass desktop client, go to **Transfer** > **Fetch** and select **Save to Olares**. Paste the address below, choose `Common/huggingface` as the destination, and start the download. See [Save resources to Olares](/manual/larepass/save-resources-to-olares.md).

```text
https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF/resolve/main/Qwen3.6-35B-A3B-UD-IQ4_XS.gguf
```
:::

Keep any matching `mmproj` vision projector beside the main weights. Colibri lists it under the model as a vision companion. The loaded engine determines whether image input is available.

Colibri reads the original weights without modifying them. It saves imported indexes and your model selection in its own application data. Store models on an SSD where possible: Colibri reads expert weights from disk as needed, so disk speed affects replies when RAM is limited.

## Install Colibri and set memory limits

Before installing, check the available RAM in Dashboard and the GPU allocation in **Settings** > **GPU**. If your device can spare `32Gi` of RAM and `8Gi` of GPU memory, start with these values:

1. Open Market and search for "Colibri". Click **Get**, then **Install**.
2. In **Configure environment variables for Colibri**, enter:

   | Setting | Start with | What it controls |
   |---|---|---|
   | `MEMORY_LIMIT` | `32Gi` | The most system RAM Colibri's main service can use |
   | `GPU_MEMORY_LIMIT` | `8Gi` | The GPU memory allocated to Colibri |

   :::tip Choose an allocation for your device
   If your device has less memory available, enter an amount it can spare and start with a smaller model. Set GPU memory to `0` to use the CPU instead. CPU inference still requires an NVIDIA-equipped Olares device for this package.
   :::

   Leave the other settings at their defaults.

3. Click **Confirm** and wait for installation to finish.

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

## Adjust resources for your model

### Change memory limits

For another model, load it and try a short prompt with the current settings. If Colibri reports insufficient RAM, raise `MEMORY_LIMIT` when your device has room. If the model loads but replies are slow, try a higher RAM limit first. If the GPU also has room, a larger GPU allocation can help. When neither resource has room, choose a smaller model or quantization.

To change the RAM limit, open **Settings** > **Applications** > **Colibri** > **Manage environment variables**, edit `MEMORY_LIMIT`, and apply the change. Colibri restarts and reloads the selected model. Export any conversation you want to keep before changing it.

To change `GPU_MEMORY_LIMIT`, uninstall Colibri while retaining its application data, then reinstall it with the new value. This setting cannot be edited after installation. The model files in `Common/huggingface` remain available.

### Choose an inference profile

Colibri starts in `standard`. If you want to try a supported MoE model with less memory available, you can switch to `low-memory-moe`. It uses staged GPU execution and may be slower, especially with long inputs. Both profiles use the RAM and GPU memory limits you entered during installation.

1. Open **Settings** > **Applications** > **Colibri** > **Manage environment variables**.
2. Set `INFERENCE_PROFILE` to `low-memory-moe`, then click **Confirm** and **Apply**. Colibri restarts.
3. Return to Colibri and wait for **Model ready** before starting a new chat.

To switch back, set `INFERENCE_PROFILE` to `standard` and apply the change. Export a conversation you want to keep before either restart.

## Review a Python function with your local model

### Connect to the local engine

1. The following steps use English interface labels. To change the chat interface language, click the settings icon in the bottom-left corner and select **English** under **Settings** > **General** > **Language**.
2. Under **Settings** > **Connection**, check that **Connection service** is **Colibri local engine (default)**. Keep the automatically filled **API endpoint**.
3. Click **Probe server**. When **Engine reachable** appears, the connection is ready. If it fails, return to the local model page and check the loading status.
4. Open the **Model** tab and select the loaded local model. Set **Max output tokens** to `1024` and leave **Reasoning** off for this short code review.

The web interface authenticates through your Olares sign-in. You do not need to enter an API key in Colibri.

### Find and fix the sorting bug

1. Click the **New chat** icon on the left to start a conversation.
2. Paste this short Python example into the input box and click the send arrow. You can use a function of your own instead.

   ```text
   This function should return the highest scores first.
   Explain the bug in one sentence, then rewrite the function.

   def top_scores(scores, limit=3):
       ranked = sorted(scores)
       return ranked[:limit]
   ```

3. Compare the suggested fix with the function you sent. The result should contain the highest scores first. If the explanation is unclear, ask a follow-up question in the same chat.

This version provides a 16K context by default for the input, conversation history, and output together. For longer files, send one function at a time. Lower RAM or GPU memory allocations can increase the wait before a reply.

### Export the conversation

:::warning Export before refreshing
Colibri does not retain chat history or generation settings after a page refresh. Export any conversation you want to keep before closing or refreshing the page.
:::

1. After the reply finishes, click the **Export conversation** icon in the top-right corner.
2. Select **Markdown (.md)** to download the conversation to your computer. The file contains your input and the model's replies for later editing or sharing. You can also choose **JSON (.json)** for structured storage.

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
4. Click **New chat** and send the same `top_scores` prompt in a new conversation. Router's current default chat model generates the reply. If you change that default model in Router later, this connection follows the change.

Router replies do not appear in Colibri's local engine performance statistics. To return to local inference, select **Colibri local engine (default)** under **Connection service**, probe the server again, select the local model, and start a new conversation.

## Learn more

- [Connect AI applications to Olares models](/manual/best-practices/connect-ai-apps.md): Learn about Router model names, connection addresses, and authentication.
- [Save resources to Olares](/manual/larepass/save-resources-to-olares.md): Download a model file directly to Olares with LarePass Fetch.
- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Manage model services and default chat models.
- [Colibri official documentation (this build)](https://github.com/JustVugg/colibri/blob/0a5504a1dc64f86fcd0fe7db298b0448c730e7e6/README.md): Read about model compatibility, how the engine works, and tuning.
