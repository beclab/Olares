---
outline: [2, 3]
description: Learn how to use Whisper-WebUI on Olares for speech-to-text transcription, subtitle generation, real-time recording, subtitle translation, and vocal separation across 96 languages.
head:
  - - meta
    - name: keywords
      content: Olares, Whisper-WebUI, speech to text, transcription, subtitles, AI, self-hosted, vocal separation
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-28"
---

# Transcribe audio and video with Whisper-WebUI

Whisper-WebUI is a browser-based speech-to-text tool for generating transcripts and subtitle files from audio, video, YouTube links, and microphone recordings. It also includes standalone tools for subtitle translation and vocal/background music separation.

Use this guide to transcribe media files on Olares, improve transcription results with optional filters, translate existing subtitle files, and separate vocals from background music when needed.

## Learning objectives

In this guide, you will learn how to:

- Install Whisper-WebUI on Olares.
- Transcribe local files, YouTube videos, and microphone recordings.
- Improve transcription results with background music removal, VAD, and speaker diarization.
- Translate subtitles and separate vocals from background music.
- Find generated files and handle model downloads if automatic downloads fail.

## Install Whisper-WebUI

1. Open Market and search for "Whisper-WebUI".
   ![Install Whisper-WebUI](/images/manual/use-cases/whisper-webui.png#bordered){width=80%}

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, you will see two icons on Launchpad:

- Whisper-WebUI: The main interface for transcription, subtitle translation, and background music separation.
- Whisper-WebUI Terminal: A command-line terminal for managing models.

## Understand the basics

### Main workflows

Whisper-WebUI includes five main tabs, which fall into two categories: transcription and standalone tools.

| Tab | Type | Best for |
| :--- | :--- | :--- |
| **File** | Transcription | Generating subtitles for local media files, up to 500 MB. |
| **Youtube** | Transcription | Transcribing online videos via URL without downloading them manually. |
| **Mic** | Transcription | Recording and transcribing audio directly in the browser. |
| **T2T Translation** | Standalone tool | Translating existing subtitle files. |
| **BGM Separation** | Standalone tool | Exporting separate vocal and instrumental tracks. |

![Whisper-WebUI interface](/images/manual/use-cases/whisper-webui-interface.png#bordered)

### Choose an output format

When using the transcription tabs, choose the output format based on how you plan to use the result.

| Format | Best for |
|:--|:--|
| SRT | Standard subtitle files for video players and editors. |
| WebVTT | Web video subtitles and browser-based playback. |
| TXT | Plain text transcripts without timestamps. |
| LRC | Synchronized lyrics for music players and audio applications. |

### Choose a transcription model

For most tasks, start with `large-v2`. It is pre-installed and works well for general transcription.

Change the model when you need a different balance between speed, accuracy, and resource usage:

| Need | Recommended model | Notes |
| :--- | :--- | :--- |
| Faster processing | `small` or `medium` | Use when large models are too slow. Accuracy may be lower. |
| Lowest resource usage | `tiny` or `base` | Use only for quick tests or simple audio. Accuracy is limited. |
| Higher accuracy | `large-v3` | Better for complex, noisy, or non-English audio, but uses more resources. |
| Faster large-model transcription | `large-v3-turbo` | Faster than `large-v3`, with some accuracy tradeoff. |
| English-only audio | Models ending in `.en` | Use only when the source audio is English. |

:::info First-time downloads
Only `large-v2` is pre-installed. Other models are downloaded automatically the first time you select them. The download may take some time, depending on your network and the model size.
:::

## Transcribe audio and video

The **File**, **Youtube**, and **Mic** tabs follow the same transcription workflow and share core settings such as model, language, output format, and advanced settings.

After each transcription task finishes, Whisper-WebUI shows the transcript in the output area and provides a downloadable subtitle or text file. Generated files are also saved in Files under `/External/olares/ai/output/whisperwebui/`.

### Transcribe local files

1. Click the **File** tab.
2. Click the upload area and select an audio or video file. The file size limit is 500 MB.
3. Under **Model**, select a transcription model.
4. Under **Language**, specify the source language or use **Automatic Detection**.
    :::tip
    Specifying the language can improve accuracy, especially for short audio or non-English content.
    :::
5. Under **File Format**, choose your preferred output format.
6. Optional: Expand panels below to [remove background music](#remove-background-music-before-transcription), [detect speech with VAD](#detect-speech-segments-with-vad), or [identify speakers](#identify-speakers-in-multi-speaker-audio).
7. Click **GENERATE SUBTITLE FILE**.

![File transcription result](/images/manual/use-cases/whisper-webui-file-result.png#bordered){width=90%}

### Transcribe YouTube videos

:::warning YouTube access limitation
YouTube transcription may fail if YouTube blocks automated access or download requests, or if the network environment cannot access the video.
:::

1. Click the **Youtube** tab.
2. Paste the YouTube video URL into the input field. Whisper-WebUI detects the video's thumbnail, title, and description when available.

   ![YouTube URL input](/images/manual/use-cases/whisper-webui-youtube-input.png#bordered){width=90%}

3. Under **Model**, select a transcription model.
4. Under **Language**, specify the video's language.
5. Under **File Format**, choose your preferred output format.
6. Optional: Expand panels below to [remove background music](#remove-background-music-before-transcription), [detect speech with VAD](#detect-speech-segments-with-vad), or [identify speakers](#identify-speakers-in-multi-speaker-audio).
7. Click **GENERATE SUBTITLE FILE**.

![YouTube transcription result](/images/manual/use-cases/whisper-webui-youtube-result.png#bordered){width=90%}

### Record and transcribe with microphone

:::info Microphone access requirement
Microphone recording requires browser microphone permission and HTTPS or localhost access. If recording does not work, check browser permissions and how you opened Whisper-WebUI.
:::

1. Click the **Mic** tab.
2. Click the record button to start recording. You can pause at any time.

   ![Mic recording](/images/manual/use-cases/whisper-webui-mic-recording.png#bordered){width=90%}

3. Click **Stop** to end the recording. You can preview and trim the audio.

   ![Mic recorded](/images/manual/use-cases/whisper-webui-mic-recorded.png#bordered){width=90%}

4. Select the **Model**, **Language**, and **File Format** for transcription.
5. Optional: Expand panels below to [remove background music](#remove-background-music-before-transcription), [detect speech with VAD](#detect-speech-segments-with-vad), or [identify speakers](#identify-speakers-in-multi-speaker-audio).
6. Click **GENERATE SUBTITLE FILE**.

### Optional transcription filters

The **File**, **YouTube**, and **Mic** tabs include optional filters that can improve results for specific audio types. Configure these filters before clicking **GENERATE SUBTITLE FILE**.

#### Remove background music before transcription

Use this feature when speech is mixed with music or background audio.

To enable it:

1. Expand **Background Music Remover Filter**.
2. Check **Enable Background Music Remover Filter**.
3. Keep the default model and segment size unless you need a custom setup.

Whisper-WebUI separates the vocal track first, then transcribes the processed audio.

:::tip BGM Separation vs. background music removal
The **BGM Separation** tab only separates audio into vocal and instrumental tracks. It does not transcribe the result.

Background music removal is part of the transcription workflow. It separates vocals first, then transcribes the vocal track.
:::

#### Detect speech segments with VAD

Use VAD for long recordings, meetings, podcasts, or audio with long silent sections. VAD can skip silence, speed up transcription, and reduce hallucinated text from silent audio.

To enable it:

1. Expand **Voice Detection Filter**.
2. Check **Enable Silero VAD Filter**.

#### Identify speakers in multi-speaker audio

Speaker diarization labels different speakers in the transcript, such as `SPEAKER_00` and `SPEAKER_01`.

Before using it for the first time, complete the Hugging Face setup so Whisper-WebUI can download the required models:

1. Expand the **Diarization** panel.
2. Under the **HuggingFace Token** field, click the two provided `pyannote` model links.
3. On Hugging Face, log in or create a free account, then accept the conditions to access both models.
4. In your Hugging Face account settings, create an access token with **Read** permissions.
5. Back in Whisper-WebUI, check **Enable Diarization**.
6. Paste your Hugging Face token into the **HuggingFace Token** input field.
7. Run transcription as usual.

:::info First-time diarization download
The first time you enable speaker diarization, Whisper-WebUI uses your token to download the required models. This may take some time. After the models are downloaded, they can be reused for future transcriptions.
:::

## Use standalone tools

Besides transcription, Whisper-WebUI provides dedicated tabs for subtitle translation and audio separation.

### Translate subtitles

Use the **T2T Translation** tab to translate existing subtitle files.

Whisper-WebUI provides two translation methods:

| Method | Requirement | Best for |
| :--- | :--- | :--- |
| **NLLB** | Downloads a local translation model on first use. | Local translation without an external API key. |
| **DeepL API** | Requires a DeepL API key. | Online translation using DeepL. |

<Tabs>
<template #Translate-with-NLLB>

The first translation may take several minutes because Whisper-WebUI downloads the selected NLLB model first. After the model is downloaded, it can be reused later.

1. Click the **T2T Translation** tab.
2. Upload the subtitle file you want to translate.
3. Under the **NLLB** subtab, select a translation model.
   
   ![T2T translation model](/images/manual/use-cases/whisper-webui-t2t-model.png#bordered){width=90%}

4. Set the **Source Language** and **Target Language**.
5. Click **TRANSLATE SUBTITLE FILE**.

Once complete, you can preview the result, download the generated file, or check it in Files at `/External/olares/ai/output/whisperwebui/`.

</template>

<template #Translate-with-DeepL-API>

:::info API key required
You must have a valid DeepL API key to use this method. If the key is missing or invalid, translation will fail.
:::

1. Click the **T2T Translation** tab.
2. Upload the subtitle file you want to translate.
3. Under the **DeepL API** subtab, enter your **DeepL API key**.
4. If you are using a DeepL Pro subscription, check **Pro account**.
5. Set the **Source Language** and **Target Language**.
6. Click **TRANSLATE SUBTITLE FILE**.


Once complete, you can preview the result, download the generated file, or check it in Files at `/External/olares/ai/output/whisperwebui/`.

</template>

</Tabs>

### Separate vocals from background music

Use the **BGM Separation** tab to split an audio file into separate vocal and instrumental tracks. This standalone tool does not transcribe the result.

1. Click the **BGM Separation** tab.
2. Upload the audio file you want to process.
3. Under **Device**, select a processing device based on your hardware.
4. Under **Model**, select a separation model.
5. Click **SEPARATE BACKGROUND MUSIC**.

![BGM separation result](/images/manual/use-cases/whisper-webui-bgm-result.png#bordered){width=90%}

Once complete, you can preview the result, download the generated file, or check it in Files:

- Instrumental: 
`/External/olares/ai/output/whisperwebui/UVR/instrumental`
- Vocals: `/External/olares/ai/output/whisperwebui/UVR/vocals`

## Advanced: Manage model downloads from Terminal

Most users do not need to manage models manually. Use Whisper-WebUI Terminal only when automatic model downloads fail, time out, or you need to check whether a model has already been downloaded.

Click the **Whisper-WebUI Terminal** icon on the Launchpad to open the web terminal.

### Check downloaded models

Open Whisper-WebUI Terminal, then run:

```bash
find /Whisper-WebUI/models -maxdepth 3 -type d
```

To check specific model folders:

```bash
# Whisper transcription models
ls -la /Whisper-WebUI/models/Whisper/faster-whisper/

# NLLB translation models
ls -la /Whisper-WebUI/models/NLLB/

# UVR background music separation models
ls -la /Whisper-WebUI/models/UVR/MDX_Net_Models/

# Speaker diarization models
ls -la /Whisper-WebUI/models/Diarization/
```

### Manually download models

While Whisper-WebUI downloads models automatically, you can trigger downloads manually if the UI download times out:

For example, to download a Whisper transcription model, replace the repository name with the model you need:
```bash
hf download Systran/faster-whisper-large-v3 \
  --cache-dir /Whisper-WebUI/models/Whisper/faster-whisper
```

After the download completes, refresh Whisper-WebUI and select the model from the model list.

## FAQs

### Why does T2T Translation fail with NLLB?

NLLB translation may fail if the model download was interrupted or the model folder is incomplete.

To reset the NLLB download:

1. Open `External/olares/ai/whisperwebui/NLLB/` in Files.
2. Delete the contents inside the folder, but keep the folder itself.
3. Return to Whisper-WebUI, and download the model again.

### Why does speaker diarization fail?

Speaker diarization may fail if:

- The Hugging Face token is missing or invalid.
- The required pyannote model terms were not accepted.
- The model download failed because of network issues.

Check that:
- The Hugging Face token has **Read** permissions.
- You accepted the terms for both pyannote models using the same Hugging Face account.
- Your network is stable while Whisper-WebUI downloads the models.

### Why do tasks fail after switching to a larger model?

A task may fail after you switch models for one of these reasons:

- The selected model has not finished downloading.
- The model requires more GPU memory than Whisper-WebUI currently has.

To fix this issue:

- Wait for the first-time model download to complete, then retry the task.
- [Assign more VRAM](/docs/manual/olares/settings/single-gpu.md#adjust-vram-allocation) to Whisper-WebUI in Memory slicing mode.
- Switch the GPU to another suitable mode.
- Choose a smaller model.

## Learn more
- [Open WebUI](openwebui.md): Use Whisper-WebUI as a speech-to-text backend for chat input.