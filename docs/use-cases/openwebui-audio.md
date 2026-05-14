---
outline: deep
description: Enable speech-to-text and text-to-speech in Open WebUI using the Speaches app on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, STT, TTS, voice, Speaches, audio
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-14"
---

# Configure voice interactions in Open WebUI

Enable hands-free communication in Open WebUI by connecting it to the Speaches application. This integration provides speech-to-text (STT) for dictating prompts and text-to-speech (TTS) for hearing AI responses aloud.


## Learning objectives

In this guide, you will learn how to:

- Retrieve the required STT and TTS configuration details from Speaches.
- Configure Open WebUI to use Speaches as the audio backend.
- Verify speech-to-text, text-to-speech, and continuous voice modes.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one active model backend.
- [Speaches](speaches.md#install-speaches) installed.
- Administrator privileges for the Open WebUI instance.
- Approximately 14 GB of available VRAM to run the LLM, STT, and TTS models simultaneously.

## Retrieve Speaches configuration details

To link Open WebUI and Speaches, you must identify specific model names and obtain the Speaches shared endpoint URL.

### Find model and voice names

1. Open Speaches from the Launchpad.
2. Go to the **Speech-to-Text** tab, click the **Model** drop-down list, and then note down the default STT model name `Systran/faster-whisper-small`.
3. Go to the **Text-to-Speech** tab, click the **Model** drop-down list, select the default TTS model name `speaches-ai/Kokoro-82M-v1.0-ONNX`, and then note down the model name.
4. From the **Voice** drop-down list, select a voice for the AI to use when reading responses aloud, and then note down the voice name. For example, `am_eric`.

   ![Text-to-speech generation](/images/manual/use-cases/speaches-tts.png#bordered){width=90%}   

### Get the shared endpoint URL

1. Open Olares Settings, and then go to **Applications** > **Speaches**.
2. In **Shared entrances**, click **Speaches API**, and then note down the endpoint URL. For example, `http://edd26bab0.shared.olares.com`.

   ![Speaches shared entrance](/images/manual/use-cases/openwebui-speaches-shared-entrance.png#bordered){width=70%}

## Configure audio settings in Open WebUI

1. In Open WebUI, click your profile icon, and then go to **Admin Panel** > **Settings** > **Audio**.
2. In the **Speech-to-Text** section, specify the following settings:

   - **Speech-to-Text Engine**: Select **OpenAI**.
   - **API Base URL**: Enter the Speaches shared endpoint URL and append `/v1` to the end. For example, `http://edd26bab0.shared.olares.com/v1`.
   - **API Key**: Enter any text. Do not leave it empty.
   - **STT Model**: Enter the STT model name you noted down earlier. That is `Systran/faster-whisper-small`.

3. In the **Text-to-Speech** section, specify the following settings:

   - **Text-to-Speech Engine**: Select **OpenAI**.
   - **API Base URL**: Enter the Speaches shared endpoint URL and append `/v1` to the end. For example, `http://edd26bab0.shared.olares.com/v1`.
   - **API Key**: Enter any text. Do not leave it empty.
   - **TTS Voice**: Enter the voice name you noted down earlier. For example, `am_eric`.
   - **TTS Model**: Enter the TTS model name you noted earlier. That is, `speaches-ai/Kokoro-82M-v1.0-ONNX`.

   ![Audio settings in Open WebUI](/images/manual/use-cases/openwebui-audio-settings.png#bordered)

4. Click **Save**.

## Verify the configuration

Test the individual audio features to ensure the integration works correctly.

:::tip Run Open WebUI in a new tab for audio
Modern web browsers block microphone access for applications running inside the Olares desktop window. To use voice features without receiving a "Permission denied" error, select <i class="material-symbols-outlined">open_in_new</i> in the top-right corner of the Open WebUI window to open it in a new browser tab. Perform the following tests in that new browser tab.
:::

### Test speech-to-text

1. Start a new chat in Open WebUI.
2. Select a chat model.
3. Click <i class="material-symbols-outlined">mic</i> next to the message input field.

   ![Dictate button](/images/manual/use-cases/openwebui-dictate-button.png#bordered)

4. Allow browser microphone access when prompted.
5. Speak into your microphone. Your speech is transcribed into the text box.

### Test text-to-speech

1. Send a message to the model and wait for a response.
2. Click <i class="material-symbols-outlined">volume_up</i> under the response. The response is spoken aloud.
   
   ![Read aloud](/images/manual/use-cases/openwebui-read-aloud.png#bordered)

### Test continuous voice mode

1. In the chat interface, click <i class="material-symbols-outlined">graphic_eq</i>. The first load might take a few moments as models initialize.

   ![Voice mode](/images/manual/use-cases/openwebui-voice-mode.png#bordered)

2. Speak naturally. The system will transcribe your speech, generate a response, and read it back automatically.

:::warning Resource usage
Using audio features invokes the LLM, STT, and TTS models simultaneously. Make sure your device has enough VRAM and memory for all three models to load and switch smoothly. If resources run low, Olares might stop apps to protect the system, causing brief unavailability.

For production use, consider setting **GPU mode** to **Time slicing** to prevent resource contention between models.
:::

:::tip Non-English speech
The default STT and TTS models might not perform well for non-English languages. You can switch to different models in the Speaches Playground if needed. For instructions on changing models, see [Manage models in Speaches](speaches.md#manage-models).
:::
