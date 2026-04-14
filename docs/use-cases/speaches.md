---
outline: [2, 3]
description: Install Speaches on Olares for speech-to-text, text-to-speech, and AI voice chat. Use the OpenAI-compatible API to integrate speech services with other apps.
head:
  - - meta
    - name: keywords
      content: Olares, Speaches, speech-to-text, text-to-speech, STT, TTS, voice chat, OpenAI-compatible, Whisper, Kokoro
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-14"
---

# Set up speech services with Speaches

Speaches is an OpenAI API-compatible speech server that provides speech-to-text (STT) and text-to-speech (TTS) capabilities. It comes with pre-loaded models so you can start transcribing audio and generating speech right away.

Because Speaches exposes an OpenAI-compatible API, any app that supports the OpenAI SDK can use it as a drop-in speech backend.

## Learning objectives

In this guide, you will learn how to:

- Install Speaches on Olares.
- Transcribe or translate audio files using speech-to-text.
- Generate speech from text using text-to-speech.
- Have voice conversations with an AI model using Audio Chat.
- Access the Speaches API from other apps.
- Download and manage speech models.

## Prerequisites

- An NVIDIA GPU is recommended for faster processing. CPU mode is also available. See [Switch to CPU mode](#switch-to-cpu-mode).
- [Ollama installed and running](ollama.md) with at least one chat model downloaded (required for Audio Chat only).

## Install Speaches

1. Open Market and search for "Speaches".

   <!-- ![Speaches in Market](/images/manual/use-cases/speaches.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, you will see two icons on Launchpad:
- Speaches: The main interface for speech-to-text, text-to-speech, and audio chat.
- Speaches Terminal: A command-line terminal for managing models.

## Use Speaches

Speaches ships with two models ready to use out of the box:

| Model | Type | Purpose |
|:------|:-----|:--------|
| `Systran/faster-whisper-small` | STT | Speech recognition and translation |
| `speaches-ai/Kokoro-82M-v1.0-ONNX` | TTS | Speech synthesis |

### Transcribe audio

1. Open Speaches and click the **Speech-to-Text** tab.
2. For **STT model**, select a model.
3. For **Task**, select **Transcribe**.
4. Upload a `.wav` file or click **Record** to capture audio directly from your microphone.
5. Click **Transcribe** to start the conversion.

<!-- ![Speech-to-text transcription](/images/manual/use-cases/speaches-stt-transcribe.png#bordered) -->

### Translate audio to English

Speaches can automatically detect the language of the audio and translate it into English.

1. Open Speaches and click the **Speech-to-Text** tab.
2. For **STT model**, select a model.
3. For **Task**, select **Translate**.
4. Upload a `.wav` file or click **Record** to capture audio directly from your microphone.
5. Click **Translate** to start the conversion.

<!-- ![Speech-to-text translation](/images/manual/use-cases/speaches-stt-translate.png#bordered) -->

### Generate speech from text

1. Open Speaches and click the **Text-to-Speech** tab.
2. Enter the text you want to convert.
3. For **TTS model**, select a model.
4. Select a **Voice** style from the dropdown.
5. Select an output **Format**.
6. Click **Generate Speech**.

<!-- ![Text-to-speech generation](/images/manual/use-cases/speaches-tts.png#bordered) -->

### Chat with AI using voice

Audio Chat lets you have a spoken conversation with an AI model. It combines STT, an LLM, and TTS into a single pipeline: your voice is transcribed to text, sent to the LLM for a response, and the reply is converted back to speech.

:::info
- Audio Chat requires Ollama to be installed with at least one chat model downloaded.
- Full text and audio output is currently supported for English queries. For other languages, only text output is available.
:::

#### Start a voice conversation

1. Open Speaches and click the **Audio Chat** tab.
2. For **Chat Model**, select a model from Ollama (for example, `qwen2.5:7b`).
3. Send a message using one of these methods:
   - **Text**: Type your message in the input field next to the microphone icon and send it.
   - **Voice**: Click **Record** to capture your message, then click the send button.

   <!-- ![Audio Chat interface](/images/manual/use-cases/speaches-audio-chat.png#bordered) -->

4. Wait for the response. The system transcribes your input, sends it to the LLM, and converts the reply to speech.

:::warning
The full voice pipeline (STT, LLM, TTS) takes time to complete. Do not refresh the page while a response is being generated, as you might see UI flickering during processing.
:::

#### Optional: Use a larger STT model

Audio Chat uses the pre-installed `Systran/faster-whisper-small` model by default. For better transcription accuracy, you can switch to a larger model such as `Systran/faster-whisper-large-v3`.

1. Open the Speaches terminal and download the model:

   ```bash
   hf download Systran/faster-whisper-large-v3
   ```

2. Open Settings, then go to **Applications** > **Speaches** > **Manage environment variables**.

3. Change `SPEACHES_WHISPER_MODEL` to the model you downloaded (for example, `Systran/faster-whisper-large-v3`), and click **Apply**.
   <!-- ![Update STT model](/images/manual/use-cases/speaches-update-stt-model.png#bordered) -->

Speaches restarts automatically to apply the changes.

## Use the Speaches API

Speaches is fully compatible with the OpenAI API format. Any app that supports the OpenAI SDK can call it directly.

### Access the API

<!-- The following content is temporary. Will update with two separate examples (access from within Olares, and outside Olares). -->

- **From other Olares apps**: Use the shared entrance endpoint. Go to **Settings** > **Applications** > **Speaches**, and copy the URL under **Shared entrances**. For example:

  ```
  http://d54536a50.shared.olares.com
  ```

  <!-- ![Speaches shared entrance](/images/manual/use-cases/speaches-shared-entrance.png#bordered) -->

- **From external apps**: Use the Speaches API entrance URL. Go to **Settings** > **Applications** > **Speaches** > **Speaches API**, and copy the URL under **Entrances**. For example:

  ```
  https://39975b9a1.laresprime.olares.com
  ```

  <!-- ![Speaches API entrance](/images/manual/use-cases/speaches-api-entrance.png#bordered) -->

  :::tip
  External access requires [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) to be enabled on your device.
  :::

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

## Switch to CPU mode

Speaches installs in GPU mode by default, which requires an NVIDIA GPU. If your device does not have an NVIDIA GPU, or you prefer to use CPU, change the `SPEACHES_GPU` setting to `false`:

1. Go to **Settings** > **Applications** > **Speaches** > **Manage environment variables**.
2. Find `SPEACHES_GPU` and change its value to `false`.

   <!-- ![Switch to CPU mode](/images/manual/use-cases/speaches-cpu-mode.png#bordered) -->

The app automatically redeploys in CPU mode. Processing will be slower compared to GPU mode.

## Manage models

### Check downloaded models
To see all downloaded models, open the Speaches terminal and run:

```bash
hf cache list
```

### Download a new model

1. Open the Speaches terminal and run:

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
   Models are stored in Olares Files, at `Home/Huggingface/`. If other apps on your Olares also use Hugging Face models, they share this directory.
   :::

2. Go to **Settings** > **Applications** > **Speaches**, and click **Stop** then **Resume** to restart Speaches to load the new model.

   <!-- ![Restart Speaches](/images/manual/use-cases/speaches-restart.png#bordered) -->

### Remove a model

To free up storage space, you can remove models you no longer need:

```bash
hf cache rm model/<model_name>
```

For example:

```bash
hf cache rm model/speaches-ai/Kokoro-82M-v1.0-ONNX
```

After removing a model, restart Speaches from **Settings** > **Applications** > **Speaches** to refresh the model list.

## FAQs

### Can I use a different Ollama instance for Audio Chat?

Yes. Update the `CHAT_COMPLETION_BASE_URL` in the deployment configuration:

1. Open Control Hub and navigate to **Browse** > **System** > **speachesserver-shared** > **Deployments** > **speaches**.

   <!-- ![Navigate to Speaches deployment](/images/manual/use-cases/speaches-controlhub-deployment.png#bordered) -->

2. Click **Edit YAML**, find `CHAT_COMPLETION_BASE_URL`, and update its value to your Ollama endpoint. Make sure the URL ends with `/v1`.

   <!-- ![Edit CHAT_COMPLETION_BASE_URL](/images/manual/use-cases/speaches-edit-base-url.png#bordered) -->

### Why does Audio Chat show an error?

Audio Chat requires Ollama to be running with at least one chat model downloaded. If Ollama is not installed or has no models available, Audio Chat displays an error. Install Ollama and download a chat model by following the [Ollama guide](ollama.md). Speaches detects Ollama automatically, so you do not need to restart Speaches.

### Why won't Speaches start?

If your device uses a non-NVIDIA GPU, Speaches cannot use the default GPU acceleration. Switch to CPU mode by setting `SPEACHES_GPU` to `false` in the app's environment variables. See [Switch to CPU mode](#switch-to-cpu-mode).

## Learn more

- [Speaches official documentation](https://speaches.ai/): Full API reference and model compatibility.
- [Ollama](ollama.md): Download and run local AI models.
- [Open WebUI](openwebui.md): Chat interface that can use Speaches as a speech backend.
