---
outline: [2, 3]
description: Use IndexTTS2 on Olares to generate natural-sounding speech from text with zero-shot voice cloning. Provide a short audio sample, type your text, and get speech that matches the reference voice.
head:
  - - meta
    - name: keywords
      content: Olares, IndexTTS2, text-to-speech, TTS, voice cloning, zero-shot, AI speech synthesis
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

# Clone voices with IndexTTS2

IndexTTS2 is a zero-shot text-to-speech (TTS) system that generates natural-sounding speech from a short audio reference. It separates speaker identity from emotion, giving you independent control over voice timbre, speaking style, and speech duration.

Running IndexTTS2 on Olares keeps your voice data and generated audio entirely on your own hardware.

## Prerequisites

- Olares is running on a device with an NVIDIA GPU (minimum 9 GB VRAM).
- The device uses an x86_64 (amd64) processor.
- You have a stable network connection for the initial model download.

## Install IndexTTS2

1. Open Market and search for "IndexTTS2".
   <!-- ![IndexTTS2](/images/manual/use-cases/indextts2.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

Open IndexTTS2 from Launchpad. On first launch, the app downloads the required model from Hugging Face. This might take a few minutes depending on your network speed.

<!-- ![IndexTTS2 interface](/images/manual/use-cases/indextts2.png#bordered) -->

## Use IndexTTS2

IndexTTS2 provides a Gradio-based interface with two ways to get started: use a built-in example voice, or upload your own audio sample.

### Generate speech from an example voice

Use the built-in examples to quickly test voice synthesis without preparing your own audio.

1. In the **Examples** section at the bottom of the page, select a sample voice.
2. Review the pre-filled text, or replace it with your own in the **Text** field.
3. Click **Synthesize**.

<!-- ![Generate from example](/images/manual/use-cases/indextts2-example.png#bordered) -->

The generated audio appears in the output player. You can play it directly in the browser or download it.

### Generate speech from a custom voice

Provide your own reference audio to clone a specific voice.

1. In the **Speech Synthesis** area, click **Upload** and select an audio file from your computer.

   :::tip Choose a good reference clip
   For best results, use a clean recording of 5 to 15 seconds with minimal background noise and a single speaker.
   :::

2. Enter the text you want to convert in the **Text** field.
3. Click **Synthesize**.

<!-- ![Generate from custom voice](/images/manual/use-cases/indextts2-custom.png#bordered) -->

The result preserves the timbre and speaking style of your reference audio while reading the new text content.

## Learn more

- [IndexTTS2 on GitHub](https://github.com/index-tts/index-tts): Source code and technical details.
