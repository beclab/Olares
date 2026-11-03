---
outline: [2, 3]
description: Run a larger MoE model in Colibri on Olares with a lower resident memory budget, turn meeting notes into action items, and export the conversation.
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, local LLM, GGUF, NVIDIA GPU, meeting notes, Olares Router
app_version: "0.2.22"
doc_version: "2.1"
doc_updated: "2026-11-03"
---

# Run a larger MoE model with Colibri on Olares

Colibri is a local AI inference engine with a web chat interface. It loads supported GGUF models. For mixture-of-experts (MoE) models, it can read expert weights from disk as needed, reducing the need to keep every expert in memory.

This is useful when a model's complete weights would take too much resident memory. In exchange, disk reads and transfers can make replies slower. System file cache can also use additional, reclaimable memory, so an engine budget is not a cap on the whole application's memory use.

On Olares, you can reuse existing model files, run inference with an NVIDIA GPU, and access the chat interface through your Olares sign-in. This tutorial uses the low-memory MoE profile with Qwen3.6-35B-A3B to turn short meeting notes into action items, then exports the conversation to save the result.

Qwen3.6-35B-A3B is used only as a convenient example for demonstrating the workflow. Colibri does not require this specific model. You can substitute another supported MoE model in GGUF format that fits your available resources. Memory use and response speed will vary with the model, quantization, and context length.

## Learning objectives

In this tutorial, you will learn how to:

- Choose the low-memory MoE profile and select an existing GGUF model.
- Organize meeting notes with a local model and export the conversation.
- Switch the chat connection to the default chat model in Olares Router.

## Prerequisites

- An amd64 device running Olares 1.12.7 or later, with AVX2 support.
- An NVIDIA GPU compatible with this Blackwell sm_120 build and an NVIDIA driver that supports CUDA 13.0. This preview package has been tested on an RTX 5090 Laptop GPU. The build does not support older NVIDIA GPU architectures.
- Colibri installed, with enough available system memory and GPU memory to load your chosen model.
- A downloaded GGUF model supported by Colibri. This tutorial uses the `UD-IQ4_XS` quantization of Qwen3.6-35B-A3B, whose weight file is about 16.5 GiB. Keep enough disk space for the complete weights even when using a smaller resident memory budget.

## Prepare your model files

Open Files and go to `Common/huggingface`. Make sure your GGUF file is in this directory or one of its subdirectories. You can reuse an existing Hugging Face cache. Keep its `snapshots` structure and relative links intact, without moving files or copying paths.

Keep any matching `mmproj` vision projector in the same directory as the main weights. Colibri displays it as a vision companion under its model, rather than as a second chat model. Finding a vision file does not enable image input. Check the loaded engine's reported capabilities.

Colibri reads the original weights without modifying them. Imported indexes and your model selection are saved in the application's own data directory. The package does not include or automatically download model weights.

## Choose the low-memory MoE profile

The default `standard` profile uses the application's general RAM budget. For this workflow, switch to `low-memory-moe` in Colibri's environment settings in Olares Settings. Set `INFERENCE_PROFILE` to `low-memory-moe` and save. This restarts the application, so export any conversation you want to keep first.

This profile sets the engine RAM budget to 4 GB and the expert-cache budget to 2 GiB, with staged GPU execution. It does not limit the entire application to those amounts. Model size, context, runtime overhead, and reclaimable file cache still affect actual resource use. It is intended for supported MoE models, rather than as a universal setting for every GGUF file.

To return to the general profile later, set `INFERENCE_PROFILE` back to `standard`.

## Scan for and select a local model

1. Open Colibri from Launchpad. If no model is loaded, the **Local models** page opens. Use the language selector at the top to choose **English** if needed.
2. Click **Rescan** to find models in `Common/huggingface`. The list shows each model's name, main weight path, architecture, and size. Vision companions appear under their model. Split GGUF files form one entry and require every shard before loading.
3. Select **Qwen3.6-35B-A3B**. Unavailable entries explain why they cannot be selected, such as an architecture unsupported by the current Colibri engine.
4. Click **Load selected model**, review the confirmation message, then click **Confirm**.
5. Wait for **Model ready**, then click **Open chat**. Large models can take several minutes to load. Expand **Loading log** to check progress while you wait.

![Select a local model in Colibri, with its vision file grouped as a companion](/images/manual/use-cases/colibri-moe-models-en.jpg#bordered)

Your selection is saved in the application data. After an application restart, Colibri reloads the same model instead of choosing the first file found during a scan. If the saved file has been moved or deleted, rescan and select another model.

To switch models, click the local model icon on the left side of the chat page. Model management opens in a new tab. Wait for the current reply to finish and export any conversation you want to keep before selecting another model. Once the new model is ready, click **Open chat** to start a new conversation.

Click **Stop local model** and confirm to stop the inference process and release its GPU memory. Olares still manages the application's GPU allocation. You can select a model again later. The stopped state is also saved, so restarting the application does not automatically reload the model.

## Organize meeting notes with your local model

### Connect to the local engine

1. The following steps use English interface labels. To change the chat interface language, click the settings icon in the bottom-left corner and select **English** under **Settings** > **General** > **Language**.
2. Under **Settings** > **Connection**, check that **Connection service** is **Colibri local engine (default)**. Keep the automatically filled **API endpoint**.
3. Click **Probe server**. When **Engine reachable** appears, the connection is ready. If it fails, return to the local model page and check the loading status.
4. Open the **Model** tab and select the loaded local model. For this example, set **Max output tokens** to `512` and leave **Reasoning** off to keep the meeting summary brief.

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

This preview package uses a local context length of 4096 tokens, shared by the input, conversation history, and output. For longer notes, work through one topic at a time, then continue with a shorter summary. Avoid pasting an entire long document at once.

### Export the conversation

1. After the reply finishes, click the **Export conversation** icon in the top-right corner.
2. Select **Markdown (.md)** to download the conversation to your computer. The file contains your input and the model's replies for later editing or sharing. You can also choose **JSON (.json)** for structured storage.

:::warning Export before refreshing
The current version does not retain chat history or generation settings after a page refresh. Export any conversation you want to keep before closing or refreshing the page.
:::

## Optional: Switch to Olares Router

Colibri can also use models provided through Router. Before you start, configure an available chat model in Router and set it as the default chat model on the **Default models** page. For background, see [Use Olares Router as your AI gateway](/use-cases/olares-router.md).

:::info The local engine keeps using GPU resources
Changing the connection service changes where chat requests go. It does not stop Colibri's local engine or release its GPU memory. Before calling a Router model on the same GPU, make sure both services fit in available GPU memory. If they do not, choose a Router model service that does not compete with Colibri for that GPU.
:::

1. In Colibri, open **Settings** > **Connection**.
2. Change **Connection service** to **Olares Router** and keep the automatically filled address. You do not need to create or enter an API key.
3. Click **Probe server**, then select `default-chat` on the **Model** tab.
4. Click **New chat** and send the same meeting notes in a new conversation. Router's current default chat model generates the reply. If you change that default model in Router later, this connection follows the change.

Router replies do not appear in Colibri's local engine performance statistics. To return to local inference, select **Colibri local engine (default)** under **Connection service**, probe the server again, select the local model, and start a new conversation.

## Learn more

- [Connect AI applications to Olares models](/manual/best-practices/connect-ai-apps.md): Learn about Router model names, connection addresses, and authentication.
- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Manage model services and default chat models.
