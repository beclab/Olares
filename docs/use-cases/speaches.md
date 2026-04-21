---
outline: [2, 3]
description: Install Speaches on Olares for speech-to-text, text-to-speech, and AI voice chat. Use the OpenAI-compatible API to integrate speech services with other apps.
head:
  - - meta
    - name: keywords
      content: Olares, Speaches, speech-to-text, text-to-speech, STT, TTS, voice chat, OpenAI-compatible, Whisper, Kokoro
app_version: "1.0.7"
doc_version: "1.0"
doc_updated: "2026-04-14"
---

# Set up speech services with Speaches

Speaches is an OpenAI-compatible speech server for speech-to-text (STT) and text-to-speech (TTS). With pre-loaded models, you can use it right out of the box, or easily integrate it as a drop-in backend for any app supporting the OpenAI SDK.

This guide walks you through installing and using Speaches on Olares, including speech-to-text, text-to-speech, Audio Chat, API access, and basic model management.

## Learning objectives

In this guide, you will learn how to:

- Install Speaches on Olares.
- Transcribe or translate audio files using speech-to-text.
- Generate speech from text using text-to-speech.
- Have voice conversations with an AI model using Audio Chat.
- Access the Speaches API from other apps.
- Manage speech models.

## Prerequisites

- Olares is running on a device with an NVIDIA GPU.
- [Ollama installed and running](ollama.md) with at least one chat model downloaded (required for Audio Chat only).

## Install Speaches

1. Open Market and search for "Speaches".

   ![Speaches in Market](/images/manual/use-cases/speaches.png#bordered){width=95%}

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, you will see two icons on Launchpad:
- Speaches: The main interface for speech-to-text, text-to-speech, and audio chat.
- Speaches Terminal: A command-line terminal for managing models.

:::info  Model setup on first launch
When you open Speaches for the first time, it downloads and initializes its built-in models. Depending on your network connection, this process may take some time.

If initialization does not finish within 30 minutes, it may time out and be canceled automatically. If this happens, wait until your network connection is stable, then open Speaches again to retry initialization.
:::

## Use Speaches

Speaches ships with two models ready to use out of the box:

| Model | Type | Purpose |
|:------|:-----|:--------|
| `Systran/faster-whisper-small` | STT | Speech recognition and translation |
| `speaches-ai/Kokoro-82M-v1.0-ONNX` | TTS | Speech synthesis |

### Transcribe audio

1. Open Speaches and click the **Speech-to-Text** tab.
2. Under **Model**, select a STT model, such as `Systran/faster-whisper-small`.
3. Under **Task**, select **transcribe**.
4. Upload an audio file or click <i class="material-symbols-outlined">mic</i> to record audio from your microphone.
5. (Optional) Enable **Stream** if you want to receive partial results while transcription is still in progress.
6. Click **Generate**.

   ![Speech-to-text transcription](/images/manual/use-cases/speaches-stt-transcribe.png#bordered){width=90%}

The transcription appears in  **Textbox** after processing completes.

### Translate audio to English

Speaches can automatically detect the language of the audio and translate it into English.

1. Open Speaches and click the **Speech-to-Text** tab.
2. Under **Model**, select a STT model, such as `Systran/faster-whisper-small`.
3. Under **Task**, select **translate**.
4. Upload an audio file or click <i class="material-symbols-outlined">mic</i> to record audio from your microphone.
5. (Optional) Enable **Stream** if you want to receive partial results while translation is still in progress. 
6. Click **Generate**.
   ![Speech-to-text translation](/images/manual/use-cases/speaches-stt-translate.png#bordered){width=90%}

The English translation appears in **Textbox** after processing completes.

### Generate speech from text

1. Open Speaches and click the **Text-to-Speech** tab.
2. Enter the text you want to convert in **Input Text**.
3. Under **Model**, select a TTS model.
4. Select a voice from **Voice**.
5. Under **Response Format**, select an output format.
6. Click **Generate Speech**.

   ![Text-to-speech generation](/images/manual/use-cases/speaches-tts.png#bordered){width=90%}

7. Play the generated audio and download it if needed.

### Chat with AI using voice

Use **Audio Chat** to talk to an AI model with voice, text, or an audio file. Speaches first converts your voice to text, sends the text to the chat model, and can convert the reply back to speech.


:::info
- Audio Chat requires Ollama to be installed, with at least one chat model downloaded.
- Audio playback is currently available for English replies only. For other languages, the reply is shown as text only.
:::

#### Start a voice conversation

1. Open Speaches and click the **Audio Chat** tab.
2. Under **Chat Model**, select an Ollama model, such as `qwen2.5:7b`.
3. Send a message using one of these methods:
   - **Audio file**: Upload an audio file.
   - **Text**: Type your message in the input field next to the microphone icon and send it.
   - **Voice**: Click <i class="material-symbols-outlined">mic</i> to record your message, then click <i class="material-symbols-outlined">send</i> to send it.

   ![Audio Chat interface](/images/manual/use-cases/speaches-audio-chat.png#bordered){width=90%}

4. Wait for Speaches to generate the reply.

:::warning
The full voice pipeline (STT, LLM, TTS) takes time to complete. Do not refresh the page while a reply is being generated, as you might see UI flickering during processing.
:::

#### Optional: Improve transcription accuracy for Audio Chat

Audio Chat uses the pre-installed `Systran/faster-whisper-small` speech-to-text model by default. For better transcription accuracy, you can switch to a larger model such as `Systran/faster-whisper-large-v3`.

:::info More GPU resources may be required
Larger models require more GPU resources. If generation tasks start failing after switching to a larger model, see [Why do tasks fail after switching to a larger model](#why-do-tasks-fail-after-switching-to-a-larger-model).
:::

1. Open Speaches Terminal and download the model:

   ```bash
   hf download Systran/faster-whisper-large-v3
   ```
   
   If you see a warning about `HF_TOKEN`, you can ignore it. The model download can still continue without this setting.

2. Go to **Settings** > **Applications** > **Speaches** > **Manage environment variables**.
3. Click <i class="material-symbols-outlined">edit_square</i> next to `SPEACHES_WHISPER_MODEL`.
4. Set the value as the model you downloaded, for example, `Systran/faster-whisper-large-v3`, then click **Confirm**.
   ![Update STT model](/images/manual/use-cases/speaches-update-stt-model.png#bordered){width=90%}

5. Click **Apply** to save the changes.

Speaches restarts automatically to apply the change.

:::tip Wait for service initialization
After the app shows as running again, wait a little longer before using it, as the service may still be initializing.
:::

<!-- ## Use the Speaches API

This section is for connecting other apps to Speaches. If you only want to use Speaches in its own interface, you can skip this section.

Speaches is fully compatible with the OpenAI API format. Any app that supports the OpenAI SDK can call it directly.-->

<!-- ### Get the endpoint

<Tabs>
<template #Access-within-Olares>

1. Go to **Settings** > **Applications** > **Speaches**.
2. Under **Shared entrances**, click **Speaches API**.
3. On the **Set up endpoint** page, copy the URL next to **Endpoint**. 

   For example:
   ```
   http://edd26bab0.shared.olares.com
   ```

![Speaches shared entrance](/images/manual/use-cases/speaches-shared-entrance.png#bordered){width=90%}

</template>
<template #Access-outside-of-Olares>

Before you start, make sure [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) is enabled on your device.
1. Go to **Settings** > **Applications** > **Speaches**.
2. Under **Entrances**, click **Speaches API**.
3. On the **Endpoint settings** panel, copy the URL next to **Endpoint**. 

   For example:
   ```
   https://a8259cf22.laresprime.olares.com
   ```

![Speaches API entrance](/images/manual/use-cases/speaches-api-entrance.png#bordered){width=90%}

</template>
</Tabs> -->

<!-- ### Connect Open Notebook to Speaches

You can use Speaches as the STT/TTS backend for Open Notebook.

1. Open Open Notebook and navigate to its settings page.
2. Add a new model provider:
   - **Name**: Name your configuration (for example, `Speaches`).
   - **Base URL**: Enter the shared entrance URL with `/v1` appended. For example:
     ```plain
     http://d54536a50.shared.olares.com/v1
     ```

3. Configure STT and TTS separately under the new provider, selecting the models you have downloaded in Speaches.

   ![Configure STT/TTS in Open Notebook](/images/manual/use-cases/speaches-open-notebook-config.png#bordered) -->

## Manage models

Manage models when you want to use a different model, improve quality, or free up storage space.

### Check downloaded models
To see all downloaded models, open Speaches Terminal and run:

```bash
hf cache list
```

### Download a new model

1. Open Speaches Terminal and run:

   ```bash
   hf download <model-name>
   ```

   For example:

   ```bash
   # Download a larger Whisper model for higher accuracy
   hf download Systran/faster-whisper-medium
   # Highest accuracy Whisper model, requires more memory
   hf download Systran/faster-whisper-large-v3
   ```

   :::tip Shared model storage
   Models are downloaded to Olares Files, at `/Home/Huggingface/speaches/`. If other apps on your Olares also use Hugging Face models, they share this directory.
   :::

2. Refresh the Speaches page to load the new model into the list.

### Remove a model

To free up storage space, you can remove models you no longer need:

1. Open Speaches Terminal and run:
```bash
hf cache rm model/<model_name>
```

For example:

```bash
hf cache rm model/Systran/faster-whisper-medium
```

2. Refresh the Speaches page to update the model list.

## Switch to CPU mode

Speaches uses GPU mode by default. If needed, you can switch it to CPU mode instead. CPU mode is slower and is mainly suitable for small tasks.

To switch to CPU mode:

1. Go to **Settings** > **Applications** > **Speaches** > **Manage environment variables**.
2. Click <i class="material-symbols-outlined">edit_square</i> next to `SPEACHES_GPU`, change its value to `false`, then click **Confirm**.

   ![Switch to CPU mode](/images/manual/use-cases/speaches-cpu-mode.png#bordered){width=90%}

3. Click **Apply** to save the changes.

Speaches automatically redeploys in CPU mode. Processing will be slower compared to GPU mode.

## FAQs

### Why does Audio Chat show an error?

Audio Chat requires Ollama to be running with at least one chat model downloaded. If Ollama is not installed or has no models available, Audio Chat displays an error. 

To fix this issue, install Ollama and download a chat model by following the [Ollama guide](ollama.md). Speaches detects Ollama automatically, so you do not need to restart Speaches.

### Why do tasks fail after switching to a larger model?

This issue usually happens when the GPU is in **Memory slicing** mode.

Larger models require more VRAM. If Speaches is assigned only a small amount of VRAM, generation tasks may fail after you switch to a larger model.

To fix this issue:
- Increase the VRAM assigned to Speaches in **Memory slicing** mode.
- Or switch the GPU to another mode.

For detailed instructions, see [Manage GPU resources](/manual/olares/settings/single-gpu.md).

### Can I use a different Ollama instance for Audio Chat?

Yes. Update the `CHAT_COMPLETION_BASE_URL` in the deployment configuration:

1. Open Control Hub and navigate to **Browse** > **System** > **speachesserver-shared** > **Deployments** > **speaches**.
2. Click <i class="material-symbols-outlined">edit_square</i> to edit the YAML file.

   ![Navigate to Speaches deployment](/images/manual/use-cases/speaches-controlhub-deployment.png#bordered){width=90%}

3. In **Edit YAML**, find `CHAT_COMPLETION_BASE_URL`, and update its value to your Ollama endpoint. Make sure the URL ends with `/v1`.
   
   ![Edit CHAT_COMPLETION_BASE_URL](/images/manual/use-cases/speaches-edit-base-url.png#bordered){width=90%}

4. Go to **Settings** > **Applications** > **Speaches**, click **Stop**, then click **Resume** to restart Speaches.

## Learn more

- [Speaches official documentation](https://speaches.ai/): Full API reference and model compatibility.
- [Ollama](ollama.md): Download and run local AI models.
- [Open WebUI](openwebui.md): Chat interface that can use Speaches as a speech backend.