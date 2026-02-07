---
outline: [2, 3]
description: Step-by-step guide to installing ACE-Step AI on Olares, generating songs with lyrics or instrumentals, optimizing audio with retake and repainting, and using Audio2Audio to transform reference audio into new music.
head:
  - - meta
    - name: keywords
      content: ACE-Step, AI music
---
# Create AI-generated music with ACE-Step <Badge type="tip" text="15 min" />

ACE-Step is an open-source music generation model developed by ACE Studio and StepFun. It generates music from lyrics and style tags, and supports editing tools such as retake, repainting, and Audio2Audio.

This guide walks you through installation, first-time generation, and basic editing workflows on Olares One.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Generate songs with lyrics, tags, and stylistic controls.
- Locate and download your generated audio files.
- Refine tracks by adjusting the style, editing sections, extending the song, or using a reference clip to reshape it.

## Prerequisites

Before you begin, make sure you have:
- Olares One connected to a stable network.
- Sufficient disk space to download models.

## Install and set up ACE-Step

With your Olares device ready, follow these steps to install ACE-Step and begin generating music.

### Install ACE-Step

Follow these steps to install ACE-Step.

1. Open Market, and search for "ACE-Step". 
    ![Install ACE-Step](/images/one/ace-market.png#bordered)

2. Click **Get**, then **Install**.  
3. Wait a few minutes for the installation to complete.

### Download required models on first-time launch

Once installation finishes, open ACE-Step from Launchpad.

Olares will automatically download and install required models. A **Download Manager** window will appear, showing model size and download progress.  
   ![ACE-Step Download Manager](/images/manual/use-cases/ace-step-download-manager.png#bordered){width=500}

After the download completes, the ACE-Step generation interface will open automatically.

## Generate your first track

Follow these steps to set your parameters and begin music generation.

### Set basic parameters

- **Audio Duration**: Drag the slider to choose the track length (up to 240 seconds). If left at the default (`-1`), the audio length will be random.
- **Format**: Select the audio format from `MP3`, `ogg`, `wav`, and `flac`.
    :::tip MP3 recommended
    It's recommended to change the default output format to MP3. This results in much smaller file sizes, faster loading, and a better overall experience.
    :::
- **Lora Name or Path**: Select a LoRA model if available. Currently, only a Chinese rap LoRA is available by default.
- **Tags**: Enter descriptors for style, mood, rhythm, or instruments, separated by commas. For example:

    ```plain
    Chinese Rap, J-Pop, Anime, kawaii future bass, Female vocals, EDM, Super Fast
    ```
- **Lyrics**: Enter your lyrics. Use structural tags to organize the song flow:
    - `[verse]` for the main verse.
    - `[chorus]` for the chorus.
    - `[bridge]` for the bridge.
    
    :::tip Generate an **instrumental-only** track
    Enter the tag `[instrumental]` or `[inst]` in the **Lyrics** area to generate music without vocals.
    :::

### Start generation

1.  Click **Generate** when all parameters are set. 
2.  Once generation is complete, click the **Play** button to preview your track.
   ![Generate the audio](/images/manual/use-cases/ace-step-generate.png#bordered){width=80%}

### Save the generated music

You can save your generated music via two methods.

<tabs>
<template #Direct-download>

Click the <i class="material-symbols-outlined">download</i> button in the upper right corner to save the audio file directly to your local device.
</template>
<template #From Olares Files>

1. Open Files.
2. Navigate to: `/Home/AI/output/acestepv2`.
3. Right-click the generated audio file and save it to your local device.
</template>
</tabs>

## Optimize your audio

ACE-Step offers powerful tools to refine and modify specific parts of your generated track.

### Regenerate the entire segment

You can generate a new version of the entire track.

1. Click the **retake** tab.
2. Adjust the **variance** slider:
    - Higher values: Create a significantly different version.
    - Lower values: Keep the new version closer to the original.
3. Click **Retake** and wait for the generation.

    ![Preview the retake](/images/manual/use-cases/ace-step-retake.png#bordered){width=90%}

### Regenerate a specific section

You can update only a selected time range while keeping the rest of the track unchanged.

1. Click the **repainting** tab.
2. Adjust the **variance** slider:
    - Higher values: Create a significantly different version of the selected section.
    - Lower values: Keep the new section closer to the original.
3. Adjust the **Repaint Start Time** and **Repaint End Time** sliders to select the section you want to regenerate.
4. Select the source for repainting:
    - `text2music`: The original song generated via Text2Music.
    - `last_repaint`: The previous repainted version.
    - `upload`: The audio you uploaded.
5. Click **Repaint** and wait for the generation.

    ![Preview the repaint](/images/manual/use-cases/ace-step-repaint.png#bordered){width=90%}

### Edit lyrics

You can edit lyrics to modify specific lines without affecting the rest of the track.

1. Click the **edit** tab.
2. Copy the original lyrics and paste them into the **Edit Lyrics** area.
3. Modify only the specific lines you wish to change.
4. Under **Edit Type**, select `only_lyrics`.
5. Click **Edit** and wait for the generation.

    ![Edit lyrics](/images/manual/use-cases/ace-step-edit-lyrics.png#bordered){width=90%}

### Edit tags

You can edit tags to reset the style or timbre of the track.

1. Click the **edit** tab.
2. Enter the new style or timbre tags (e.g., `hard rock` or `male tenor vocals`) in the **Edit Tags** area.
3. In **Edit Type**, select `remix`.
4. Click **Edit** and wait for the generation.

    ![Edit tags](/images/manual/use-cases/ace-step-edit-tags.png#bordered)

### Extend the audio

You can extend the length of the original track by adding new audio before or after it.

1. Click the **extend** tab.
2. Adjust the sliders:
    - **Left Extend Length**: Add new audio *before* the original audio.
    - **Right Extend Length**: Add new audio *after* the original audio.
3. Select the source to extend:
    - `text2music`: The original song generated via Text2Music.
    - `last_extend`: The previous extended version.
    - `upload`: The audio you uploaded.
4. Click **Extend** and wait for the generation.

    ![Extend tags](/images/manual/use-cases/ace-step-extend.png#bordered)

## Audio2Audio

You can create a new track based on a **reference audio** clip you upload. The AI analyzes characteristics such as timbre, rhythm, and style to produce a track with a similar feel.
1. Check the box to **Enable Audio2Audio**.
2. Upload an existing music clip to serve as the reference.
3. Adjust the **Refer audio strength** slider. A higher value results in music more similar to the reference track.
4. Select a **Preset** style, or keep the default.
5. Set other parameters as needed.
6. Click **Generate** to create new music with an atmosphere similar to the reference audio.

    ![Audio2Audio](/images/manual/use-cases/ace-step-audio2audio.png#bordered)