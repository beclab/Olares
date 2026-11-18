---
outline: [2, 3]
description: Run a larger MoE model in Colibri on Olares, chat with it, and adjust memory settings for your device.
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, local LLM, GGUF, MoE, RAM, GPU memory, CPU inference, Python code review, Olares Router
app_version: "0.2.35"
doc_version: "3.0"
doc_updated: "2026-11-18"
---

# Run a larger MoE model with Colibri on Olares

Colibri lets you chat with supported GGUF models on your Olares device. It can read parts of a large mixture-of-experts (MoE) model from disk as needed, so the whole model does not have to fit in memory.

You can choose how much system RAM and GPU memory to give Colibri, or use the CPU for inference. Less memory means more reading from disk and can make replies slower, especially with long inputs.

This guide uses Qwen3.6-35B-A3B as an example. You can follow the same steps with another supported GGUF model and adjust the memory settings for your device.

## Learning objectives

In this tutorial, you will learn how to:

- Add a GGUF model file, set RAM and GPU memory limits, and choose an inference profile.
- Chat with a local model and export the conversation.
- Adjust resources when you try another model.

## Prerequisites

- An Intel or AMD 64-bit computer running Olares 1.12.7 or later, with an Intel Core 4th-generation or newer processor, an AMD Ryzen processor, or an equivalent processor.
- An NVIDIA GPU on the Olares device. This build supports GPU acceleration on GeForce RTX 50-series GPUs and has been tested on an RTX 5090 Laptop GPU. GPU acceleration also requires a driver that supports CUDA 13.0.
- Enough disk space for your model file. The Qwen3.6-35B-A3B `UD-IQ4_XS` file used below is about 16.5 GiB.

## Add a model file

Colibri scans `Common/huggingface` for GGUF files. If your model is already under this folder, go to [install Colibri](#install-colibri-and-set-memory-limits).

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

Store models on an SSD where possible. When memory is limited, disk speed affects how quickly Colibri can read model data and respond.

## Install Colibri and set memory limits

If your device can spare `32Gi` of RAM and `8Gi` of GPU memory, start with these values. You can adjust the RAM limit later when trying another model.

1. Open Market and search for "Colibri". Click **Get**, then **Install**.
2. In **Configure environment variables for Colibri**, enter:

   | Setting | Start with | What it controls |
   |---|---|---|
   | `MEMORY_LIMIT` | `32Gi` | The system RAM available to Colibri |
   | `GPU_MEMORY_LIMIT` | `8Gi` | The GPU memory allocated to Colibri |

   :::tip Choose an allocation for your device
   Check available RAM in Dashboard and the GPU allocation in **Settings** > **GPU**. If your device has less memory available, enter amounts it can spare and try a smaller model first. Set GPU memory to `0` to use the CPU.
   :::

   Leave the other settings at their defaults.

3. Click **Confirm** and wait for installation to finish.

## Scan for and select a local model

1. Open Colibri from Launchpad. The **Local models** page opens when no model is loaded.
2. Click **Rescan** to find GGUF files in `Common/huggingface`.
3. Select **Qwen3.6-35B-A3B** and click **Load selected model** > **Confirm**.
4. When the status changes to **Model ready**, click **Open chat**. You can expand **Loading log** to check progress while the model loads.

![Select a local model in Colibri](/images/manual/use-cases/colibri-moe-models-en.jpg#bordered)

Colibri remembers the selected model and loads it again when the app restarts.

## Chat with your local model

1. Type a question into the chat box and send it. For example, paste this prompt to ask the model to fix a short Python function:

   ```text
   This function should return the highest scores first.
   Explain the bug in one sentence, then rewrite the function.

   def top_scores(scores, limit=3):
       ranked = sorted(scores)
       return ranked[:limit]
   ```

2. Read the reply. If the suggested fix is unclear, ask a follow-up question in the same chat.

For longer files, send one function or section at a time. The default 16K context is shared by your input, conversation history, and the reply. Long inputs may take longer to process when memory is limited.

### Export the conversation

:::warning Export before refreshing
Colibri does not retain chat history or generation settings after a page refresh. Export any conversation you want to keep before closing or refreshing the page.
:::

1. After the reply finishes, click the **Export conversation** icon in the top-right corner.
2. Select **Markdown (.md)** to download the conversation to your computer. The file contains your input and the model's replies for later editing or sharing. You can also choose **JSON (.json)** for structured storage.

## Manage local models

### Switch models

1. Finish the current reply and export the conversation if you want to keep it.
2. Click the local model icon on the left of the chat page to open model management.
3. Select another model and click **Load selected model** > **Confirm**.
4. When the model is ready, click **Open chat** to start a new conversation.

### Stop the local model

1. Open model management from the local model icon on the chat page.
2. Click **Stop local model** and confirm. This releases the RAM and GPU memory used by the inference process.

You can load a model again from the same page.

## Adjust resources for your model

To tune Colibri for another model, first try a short prompt with the current settings. If Colibri reports insufficient RAM or responds too slowly, raise the RAM limit if your device has room. More GPU memory can also improve speed. If the device cannot spare more memory, try a smaller model or quantization.

### Change the RAM limit

1. Export any conversation you want to keep. Changing this setting restarts Colibri.
2. Open **Settings** > **Applications** > **Colibri** > **Manage environment variables**.
3. Change `MEMORY_LIMIT` to an amount your device can spare, then click **Confirm** > **Apply**.
4. Return to Colibri and wait for the model to be ready.

### Change the GPU memory allocation

1. Export any conversation you want to keep.
2. Uninstall Colibri and choose to retain its application data.
3. Reinstall Colibri and enter the new `GPU_MEMORY_LIMIT` value. Your model files in `Common/huggingface` remain available.

### Choose an inference profile

Colibri starts in `standard`. If you want to try a supported MoE model with less memory available, you can switch to `low-memory-moe`. It uses the same RAM and GPU memory limits, but may respond more slowly, especially to long inputs.

1. Export any conversation you want to keep. Switching profiles restarts Colibri.
2. Open **Settings** > **Applications** > **Colibri** > **Manage environment variables**.
3. Set `INFERENCE_PROFILE` to `low-memory-moe`, then click **Confirm** > **Apply**.
4. Return to Colibri and wait for **Model ready** before starting a new chat.

To switch back, set `INFERENCE_PROFILE` to `standard` and apply the change.

## Resource configuration and performance reference

These measurements show how memory settings affected the example model on Olares One with an RTX 5090 Laptop GPU. Each request used about 1K input tokens and 128 output tokens with a 16K context. The figures are from the second request at each setting.

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

For this model, `20Gi` of GPU memory produced a large speed increase.

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

For this model, raising the RAM limit from `24Gi` to `32Gi` made the first response much faster. Use the tables to see what may change, then test the model you want to run on your own device.

## Optional: Switch to Olares Router

If you have set a default chat model in Router, you can use it from Colibri. See [Use Olares Router as your AI gateway](/use-cases/olares-router.md) to set one up.

1. In Colibri, open **Settings** > **Connection**.
2. Change **Connection service** to **Olares Router**.
3. On the **Model** tab, select `default-chat`.
4. Start a new chat. To return to your local model, select **Colibri local engine (default)** under **Connection service** and start another new chat.

Switching the chat connection does not stop a loaded local model. To free its memory, [stop the local model](#stop-the-local-model).

## Learn more

- [Save resources to Olares](/manual/larepass/save-resources-to-olares.md): Download a model file directly to Olares with LarePass Fetch.
- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Manage model services and default chat models.
- [Colibri official documentation (this build)](https://github.com/JustVugg/colibri/blob/0a5504a1dc64f86fcd0fe7db298b0448c730e7e6/README.md): Read about model compatibility, how the engine works, and tuning.
