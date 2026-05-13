---
outline: deep
description: Enable speech-to-text and text-to-speech in Open WebUI using the Speaches app on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, STT, TTS, voice, Speaches, audio
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

# Configure audio

Open WebUI supports voice interactions through speech-to-text (STT) and text-to-speech (TTS). This guide shows you how to connect Open WebUI to the Speaches app for voice input and output.

## Prerequisites

- [Open WebUI and a model backend installed](openwebui.md) on Olares
- Approximately 14 GB of VRAM for LLM, STT, and TTS models to run simultaneously
- Admin privileges on Open WebUI

## Install Speaches

1. Open Market and search for "Speaches".
   ![Speaches in Market](/images/manual/use-cases/speaches.png#bordered){width=95%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Note models and voice

1. Open the Speaches app.
2. In the Playground, note the default model names and voice:
   - **STT model**: For example, `Systran/faster-whisper-small`
   - **TTS model**: For example, `speaches-ai/Kokoro-82M-v1.0-ONNX`
   - **Voice**: For example, `am_eric`
   <!-- ![Speaches Playground](/images/manual/use-cases/openwebui/speaches-playground.png#bordered) -->

   You will need these values in a later step.

## Get the Speaches endpoint URL

1. Open Olares Settings and navigate to **Applications** > **Speaches**.
2. In **Shared entrances**, copy the endpoint URL.
   <!-- ![Speaches shared entrance](/images/manual/use-cases/openwebui/speaches-shared-entrance.png#bordered) -->

   For example:
   ```plain
   http://a1b2c3d40.shared.olares.com
   ```

## Configure audio settings

1. In Open WebUI, click your **profile icon** and select **Admin Panel**.
2. Navigate to **Settings** > **Audio**.
3. Set both **Speech-to-Text Engine** and **Text-to-Speech Engine** to **OpenAI**.
4. Fill in the following fields:
   - **API Base URL**: Paste the Speaches endpoint URL and append `/v1` to the end. For example:
     ```plain
     http://a1b2c3d40.shared.olares.com/v1
     ```
   - **STT Model**: Enter the STT model name you noted earlier.
   - **TTS Model**: Enter the TTS model name you noted earlier.
   - **TTS Voice**: Enter the voice name you noted earlier.
   <!-- ![Audio settings](/images/manual/use-cases/openwebui/audio-settings.png#bordered) -->
5. Click **Save**.

## Verify the configuration

### Test speech-to-text

1. Open a chat in Open WebUI.
2. Click the **dictate** button (microphone icon) next to the message input field.
   <!-- ![Dictate button](/images/manual/use-cases/openwebui/dictate-button.png#bordered) -->
3. Allow browser microphone access when prompted.
   <!-- ![Mic permission](/images/manual/use-cases/openwebui/mic-permission.png#bordered) -->
4. Speak into your microphone. Your speech should be transcribed into the text box.
   <!-- ![STT result](/images/manual/use-cases/openwebui/stt-result.png#bordered) -->

### Test text-to-speech

1. Send a message to the model and wait for a response.
2. Click the **Read Aloud** button below the response.
   <!-- ![Read aloud](/images/manual/use-cases/openwebui/read-aloud.png#bordered) -->
3. You should hear the response spoken aloud.

### Test voice mode

1. In the chat interface, click the **Voice Mode** button.
   <!-- ![Voice mode](/images/manual/use-cases/openwebui/voice-mode.png#bordered) -->
2. The first load might take a few moments as models initialize.
3. Speak naturally. The system will transcribe your speech, generate a response, and read it back automatically.

:::warning Resource usage
Using audio features invokes the LLM, STT, and TTS models simultaneously. Make sure your device has enough VRAM and memory for all three models to load and switch smoothly. If resources run low, Olares might stop apps to protect the system, causing brief unavailability.

For production use, consider GPU time slicing mode to prevent resource contention between models.
:::

:::tip Non-English speech
The default STT and TTS models might not perform well for non-English languages. You can switch to different models in the Speaches Playground if needed. For instructions on changing models, see [Manage models in Speaches](speaches.md#manage-models).
:::
