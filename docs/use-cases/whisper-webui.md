---
outline: [2, 3]
description: Learn how to use Whisper-WebUI on Olares for speech-to-text transcription, subtitle generation, real-time recording, subtitle translation, and vocal separation across 96 languages.
head:
  - - meta
    - name: keywords
      content: Olares, Whisper-WebUI, speech to text, transcription, subtitles, AI, self-hosted, vocal separation
app_version: "1.0.14"
doc_version: "1.0"
doc_updated: "2026-03-30"
---

# Transcribe audio and video with Whisper-WebUI

Whisper-WebUI is an open-source speech-to-text tool powered by OpenAI's Whisper model, supporting 96 languages. It accepts audio files, video files, YouTube links, and live microphone input, and generates timestamped subtitles in formats like SRT and TXT. Beyond transcription, it can also translate subtitle files and separate vocals from background music.

## Install Whisper-WebUI

1. Open Market and search for "Whisper-WebUI".
   <!-- ![Install Whisper-WebUI](/images/manual/use-cases/whisper-webui.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Interface overview

The main interface has five tabs, each corresponding to a different task:

| Tab | Input | Output | Best for |
|:-------|:------|:-------|:---------|
| File | Audio/video files | Subtitle files (SRT/TXT) | Extracting subtitles from media |
| YouTube | Video URL | Multi-language subtitles | Language learning, video review |
| Mic | Microphone input | Live transcript | Meeting notes, lecture recording |
| T2T Translation | SRT subtitle files | Translated subtitles | Localizing video content |
| BGM Separation | Audio files | Vocal and instrumental tracks | Remixing, content repurposing |

<!-- ![Whisper-WebUI interface](/images/manual/use-cases/whisper-webui-interface.png#bordered) -->

## Use Whisper-WebUI

### Transcribe local files

1. Click the upload area and select an audio or video file.
2. Under **Model**, select a transcription model (e.g., V3 for better accuracy).
3. Under **Language**, specify the source language.
4. Under **File Format**, choose your preferred output format (e.g., SRT).
5. Click **GENERATE SUBTITLE FILE**.

Once complete, preview the result in the left panel and download the subtitle file.

<!-- ![File transcription result](/images/manual/use-cases/whisper-webui-file-result.png#bordered) -->

### Transcribe YouTube videos

1. Paste the YouTube video URL into the input field. Whisper-WebUI automatically detects the video's thumbnail, title, and description.

   <!-- ![YouTube URL input](/images/manual/use-cases/whisper-webui-youtube-input.png#bordered) -->

2. Under **Model**, select a transcription model.
3. Under **Language**, specify the video's language.
4. Under **File Format**, choose your preferred output format.
5. Adjust additional settings if needed, such as filtering background music from the audio before transcription.
6. Click **GENERATE SUBTITLE FILE**.

Once complete, preview the result in the left panel and download the subtitle file.

<!-- ![YouTube transcription result](/images/manual/use-cases/whisper-webui-youtube-result.png#bordered) -->

### Record and transcribe with microphone

1. Click the record button to start recording. You can pause at any time.

   <!-- ![Mic recording](/images/manual/use-cases/whisper-webui-mic-recording.png#bordered) -->

2. Click **Stop** to end the recording. You can preview and trim the audio.

   <!-- ![Mic recorded](/images/manual/use-cases/whisper-webui-mic-recorded.png#bordered) -->

3. Select the **Model** and **File Format** for transcription.
4. Click **GENERATE SUBTITLE FILE**.

Once complete, preview and download the transcription.

### Translate subtitles

1. Upload the subtitle file you want to translate.
2. Under the **NLLB** tab, select a translation model.

   <!-- ![T2T translation model](/images/manual/use-cases/whisper-webui-t2t-model.png#bordered) -->

3. Set the **Source Language** and **Target Language**.
4. Click **Generate Translation File**.

Once complete, preview and download the translated subtitle file.

### Separate vocals from background music

1. Upload the audio file you want to process.
2. Under **Device**, select a processing device based on your hardware. Choose CUDA if you have an NVIDIA GPU.
3. Under **Model**, select a separation model. The default segment size is `256`.
4. Click **SEPARATE BACKGROUND MUSIC**.

Once complete, download the separated vocal and instrumental tracks.

<!-- ![BGM separation result](/images/manual/use-cases/whisper-webui-bgm-result.png#bordered) -->
