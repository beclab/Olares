---
outline: [2, 3]
description: Learn how to use Whisper-WebUI on Olares for speech-to-text transcription, subtitle generation, real-time recording, subtitle translation, and vocal separation across 96 languages.
head:
  - - meta
    - name: keywords
      content: Olares, Whisper-WebUI, speech to text, transcription, subtitles, AI, self-hosted, vocal separation
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-04-10"
---

# Transcribe audio and video with Whisper-WebUI

Whisper-WebUI is an open-source speech-to-text tool powered by OpenAI's Whisper model, supporting 96 languages. It accepts audio files, video files, YouTube links, and live microphone input, and generates timestamped subtitles in formats like SRT and TXT. Beyond transcription, it can also translate subtitle files and separate vocals from background music.

## Install Whisper-WebUI

1. Open Market and search for "Whisper-WebUI".
   ![Install Whisper-WebUI](/images/manual/use-cases/whisper-webui.png#bordered){width=80%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Understand the basics

### Main workflows

Whisper-WebUI includes five main tabs. Each tab is designed for a different workflow.

| Tab | Input | Output | Best for |
| :--- | :--- | :--- | :--- |
| File | Local audio/video files | Transcripts & subtitle files | Generating subtitles for podcasts, interviews, or local media. |
| YouTube | YouTube video URL | Transcripts & subtitle files | Transcribing online videos without downloading them first. |
| Mic | Microphone recording | Transcripts from recorded audio | Dictating voice notes, spoken drafts, or short speeches. |
| T2T Translation | Subtitle files | Translated subtitle files | Localizing video content into other languages. |
| BGM Separation | Audio files | Isolated vocal & instrumental tracks | Removing background music for better transcription, or remixing. |

![Whisper-WebUI interface](/images/manual/use-cases/whisper-webui-interface.png#bordered)

### Choose an output format

When using the **File**, **YouTube**, or **Mic** tabs, choose the output format based on how you plan to use the result.

| Format | Best for |
|:--|:--|
| SRT | Standard subtitle files for video players and editors. |
| WebVTT | Web video subtitles and browser-based playback. |
| TXT | Plain text transcripts without timestamps. |
| LRC | Synchronized lyrics for music players and audio applications. |

### Choose a transcription model

When using the **File**, **YouTube**, or **Mic** tabs, the exact list of available models may vary slightly, but they generally follow the same Whisper naming patterns.

- **Smaller models**: models starting with `tiny` or `base`  
  Choose these when you want faster results or have limited hardware resources.

- **Mid-size models**: models starting with `small` or `medium`  
  Choose these for a balance between speed and accuracy.

- **Larger models**: `large-v1`, `large-v2`, `large-v3`, `large`  
  Choose these when accuracy matters more than speed.

- **Distilled models**: models starting with `distil-`  
  Choose these when you want a lighter and faster alternative.

- **Turbo models**: `large-v3-turbo`, `turbo`  
  Choose these for fast transcription, but not for speech-to-English translation.

- **English-only models**: models ending in `.en`  
  Choose these when the source audio is English only.

- **Multilingual models**: models without `.en`  
  Choose these for non-English audio, mixed-language audio, or speech-to-English translation.

For most transcription tasks, start with `small` or `medium`. If the result is not accurate enough, move to a larger model.

## Use Whisper-WebUI

### Transcribe local files

1. Click the upload area and select an audio or video file.
2. Under **Model**, select a transcription model (e.g., V3 for better accuracy).
3. Under **Language**, specify the source language.
4. Under **File Format**, choose your preferred output format (e.g., SRT).
5. Click **GENERATE SUBTITLE FILE**.

Once complete, preview the result in the left panel and download the subtitle file.

![File transcription result](/images/manual/use-cases/whisper-webui-file-result.png#bordered){width=80%}

### Transcribe YouTube videos

1. Paste the YouTube video URL into the input field. Whisper-WebUI automatically detects the video's thumbnail, title, and description.

   ![YouTube URL input](/images/manual/use-cases/whisper-webui-youtube-input.png#bordered){width=80%}

2. Under **Model**, select a transcription model.
3. Under **Language**, specify the video's language.
4. Under **File Format**, choose your preferred output format.
5. Adjust additional settings if needed, such as filtering background music from the audio before transcription.
6. Click **GENERATE SUBTITLE FILE**.

Once complete, preview the result in the left panel and download the subtitle file.

![YouTube transcription result](/images/manual/use-cases/whisper-webui-youtube-result.png#bordered){width=80%}

### Record and transcribe with microphone

1. Click the record button to start recording. You can pause at any time.

   ![Mic recording](/images/manual/use-cases/whisper-webui-mic-recording.png#bordered){width=80%}

2. Click **Stop** to end the recording. You can preview and trim the audio.

   ![Mic recorded](/images/manual/use-cases/whisper-webui-mic-recorded.png#bordered){width=80%}

3. Select the **Model** and **File Format** for transcription.
4. Click **GENERATE SUBTITLE FILE**.

Once complete, preview and download the transcription.

### Translate subtitles

1. Upload the subtitle file you want to translate.
2. Under the **NLLB** tab, select a translation model.
   
   ![T2T translation model](/images/manual/use-cases/whisper-webui-t2t-model.png#bordered){width=80%}

3. Set the **Source Language** and **Target Language**.
4. Click **Generate Translation File**.

Once complete, preview and download the translated subtitle file.

### Separate vocals from background music

1. Upload the audio file you want to process.
2. Under **Device**, select a processing device based on your hardware.
3. Under **Model**, select a separation model. The default segment size is `256`.
4. Click **SEPARATE BACKGROUND MUSIC**.

![BGM separation result](/images/manual/use-cases/whisper-webui-bgm-result.png#bordered){width=90%}

Once complete, find the separated audio tracks in the Files app at the following paths.
- Instrumental: 
`External/olares/ai/output/whisperwebui/UVR/instrumental`
- Vocal: `External/olares/ai/output/whisperwebui/UVR/vocals`
