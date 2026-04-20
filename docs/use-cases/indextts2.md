---
outline: deep
description: Use IndexTTS2 on Olares to generate natural-sounding speech from text with zero-shot voice cloning. Provide a short audio sample, type your text, and get speech that matches the reference voice.
head:
  - - meta
    - name: keywords
      content: Olares, IndexTTS2, text-to-speech, TTS, voice cloning, zero-shot, AI speech synthesis
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-20"
---

# Clone voices with IndexTTS2

IndexTTS2 is a zero-shot text-to-speech (TTS) system that generates natural-sounding speech from a short audio reference. It separates speaker identity from emotion, giving you independent control over voice timbre, speaking style, and speech duration.

Running IndexTTS2 on Olares keeps your voice data and generated audio entirely on your own hardware.

Use this guide to quickly test IndexTTS2 with a built-in example voice, clone speech from your own reference audio, or adjust emotion settings for the generated result.

## Learning objectives

In this guide, you will learn how to:

- Install IndexTTS2.
- Generate speech from text by using either a built-in example voice or your own reference audio.
- Adjust emotion settings to change the emotional tone of the generated speech.

## Prerequisites

- Olares is running on a device with an NVIDIA GPU (minimum 9 GB VRAM).
- The device uses an x86_64 (amd64) processor.
- You have a stable network connection for the initial model download.

## Install IndexTTS2

1. Open Market and search for "IndexTTS2".

   ![IndexTTS2](/images/manual/use-cases/indextts2.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open IndexTTS2.

On first launch, IndexTTS2 downloads the required model files from Hugging Face and then initializes them locally. This may take several minutes, depending on your network speed and device performance.

   ![IndexTTS2 model download](/images/manual/use-cases/indextts2-model-download.png#bordered){width=90%}

If the download does not complete or the page appears stuck for a long time, check whether your network can access Hugging Face, then reopen IndexTTS2 and try again.

## Generate speech

IndexTTS2 provides two ways to get started:
- Use a built-in example voice to test the app quickly.
- Use your own reference audio to clone a specific voice.

![IndexTTS2 interface](/images/manual/use-cases/indextts2-interface.png#bordered){width=90%}

### Use an example voice

Use the built-in examples to quickly test voice synthesis without preparing your own audio.

1. In **Examples**, select a sample voice.
2. In the **Text** field, keep the default text or enter your own.
3. (Optional) To change the emotion, see [Adjust emotion](#optional-adjust-emotion).
4. Click **Synthesize**.

   ![Generate from example](/images/manual/use-cases/indextts2-example.png#bordered){width=90%}

When generation finishes, the result appears in the output audio player. You can play it directly in the browser or click <i class="material-symbols-outlined">download</i> to download it.

### Use your own reference audio

Use this option when you want the generated speech to match a specific speaker.

1. Upload a reference audio file to **Voice Reference** area, or click <i class="material-symbols-outlined">mic</i> to record one.

   :::tip Choose a good reference clip
   For best results, use a clean recording of 5 to 15 seconds with minimal background noise and a single speaker.
   :::

2. In the **Text** field, enter the text you want to synthesize.
3. (Optional) To change the emotion, see [Adjust emotion](#optional-adjust-emotion).
4. Click **Synthesize**.
   
   ![Generate from custom voice](/images/manual/use-cases/indextts2-custom.png#bordered){width=90%}

When generation finishes, the result appears in the output audio player. You can play it directly in the browser or click <i class="material-symbols-outlined">download</i> to download it.

### Optional: Adjust emotion

By default, IndexTTS2 uses the emotion from the main reference audio. You can generate speech without changing any emotion settings.

If you want to change the emotion, expand **Settings**, then choose a method under **Emotion control method**.

#### Use an emotion reference audio

Use this option when you want to keep one speaker's voice but borrow the emotion from another clip.

1. Select **Use emotion reference audio**.
2. Upload an audio clip in **Upload emotion reference audio** or click <i class="material-symbols-outlined">mic</i> to record one.
3. Adjust **Emotion control weight** to control how strongly the emotion reference affects the generated speech.

#### Use emotion vectors

Use this option when you want direct control over emotional intensity.

1. Select **Use emotion vectors**.
2. Adjust one or more emotion sliders, such as **Happy**, **Angry**, **Sad**, **Afraid**, **Disgusted**, **Melancholic**, **Surprised**, and **Calm**.
3. Adjust **Emotion control weight** to control how strongly the emotion is applied.

   ![Use emotion vectors](/images/manual/use-cases/indextts2-emotion-vectors.png#bordered){width=90%}

:::tip Adjust emotion weight gradually
We recommend starting with a value around 0.6. Higher values increase emotional intensity, while lower values preserve the original voice's natural tone.
:::

## FAQs

### Why is the audio cut off before the text finishes?

If the generated audio stops before the full text is spoken, **max_mel_tokens** in **Advanced generation parameter settings** may be too low.

To fix this issue:
1. Expand **Advanced generation parameter settings**.
2. Increase **max_mel_tokens**.
3. Generate the audio again.
4. If the text is very long, also increase **Max tokens per generation segment** slightly and try again.

### Why does long text sound choppy or pause at awkward places?

Long text is split into smaller segments before generation. If the text is split too aggressively, the result may sound less smooth or pause in unnatural places.

To improve continuity:
1. Expand **Advanced generation parameter settings**.
2. Review **Preview of the audio generation segments** to see how the text is being split.
3. Increase **Max tokens per generation segment** gradually.
4. Generate the audio again and compare the result.

:::tip
If the text is very long, consider manually breaking it into smaller paragraphs before generating.
:::

### Why does the output contain repeated words or phrases?

If the generated speech repeats words or phrases unnaturally, the current decoding settings may be causing too much variation.

To reduce repetition:
1. Expand **Advanced generation parameter settings**.
2. Increase **repetition_penalty** slightly.
3. Generate the audio again.
4. If repetition continues, try lowering **temperature** slightly and test again.

Adjust these values gradually. Large changes may make the result sound less natural.

## Learn more

- [IndexTTS2 on GitHub](https://github.com/index-tts/index-tts): Source code and technical details.
