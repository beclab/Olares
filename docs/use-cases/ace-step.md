---
outline: [2, 3]
description: Step-by-step guide to installing ACE-Step AI on Olares, generating songs with lyrics or instrumentals, optimizing audio with retake and repainting, and using Audio2Audio to transform reference audio into new music.
---

# Create your own AI-generated music with ACE-Step

ACE-Step, developed by ACE Studio and StepFun, is an open-source model that generates music from the lyrics and style tags you provide, allowing you to create songs, vocals, and instrumentals from simple text inputs. With its built-in tools, you can also refine your tracks by adjusting or regenerating specific parts without starting over.

This guide shows you how to install ACE-Step on Olares, generate your first track, explore different musical styles, and enhance your audio using the editing features built into the app.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install ACE-Step on your Olares device.
- Generate songs with lyrics, tags, and stylistic controls.
- Locate and download your generated audio files.
- Refine tracks by adjusting the style, editing sections, extending the song, or using a reference clip to reshape it.

## Prerequisites

Before you begin, make sure:
- Olares running on a machine equipped with an NVIDIA GPU.

## Install and set up ACE-Step

With your Olares device ready, follow these steps to install ACE-Step and begin generating music.

### Install ACE-Step

Follow these steps to install ACE-Step.

1. Open the **Market** app in your Olares web interface.  
2. Use the search bar and type "ACE-Step". 
3. Click **Get**, then click **Install**.  
   ![ACE-Step install](../public/images/manual/use-cases/ace-step-install.png#bordered)
4. Wait a few minutes for the installation to complete.

### Download required models on first-time launch

Once installation finishes, open ACE-Step from Launchpad.

Olares will automatically download and install required models. A **Download Manager** window will appear, showing model size and download progress.  
   ![ACE-Step Download Manager](../public/images/manual/use-cases/ace-step-download-manager.png#bordered){width=500}

After the download completes, the ACE-Step generation interface will open automatically.

## Generate your first track

Follow these steps to set your parameters and begin music generation.

### Set basic parameters

- **Audio Duration**: Drag the slider to choose the track length (up to **240 seconds**).
- **Format**: Select the audio format from `MP3`, `ogg`, `wav`, and `flac`.
    :::tip MP3 recommended
    It's recommended to change the default output format to MP3. This will result in much smaller file sizes, faster loading, and a better user experience.
    :::
- **Lora Name or Path**: Select a LoRA model if available. Currently, only a Chinese rap LoRA is supported.
- **Tags**: Enter descriptors for style, mood, rhythm, or instruments, separated by commas. For example:
- 
    ```plain
    Chinese Rap, J-Pop, Anime, kawaii future bass, Female vocals, EDM, Super Fast`
    ```
- **Lyrics**: Enter your lyrics, ensuring you use structural tags for optimal organization and flow:
    - `[verse]` for the main verse part
    - `[chorus]` for the chorus part
    - `[bridge]` for the bridge
    
    :::tip Generate an **instrumental-only** track
    Enter the tag `[instrumental]` or `[inst]` in the Lyrics area.
    :::
    :::tip Inspirations for genre or lyrics
    Use an AI assistant to help generate style prompts or lyrical content.
    :::

### Start generation

1.  Click **Generate** when all parameters are set. 
2.  Once generation is complete, click the **Play** button to preview your track.
   ![Generate the audio](../public/images/manual/use-cases/ace-step-generate.png#bordered)

### Save the generated music

You can save your generated music via two methods:

- **Direct download**: Click the <i class="material-symbols-outlined">download</i> button in the upper right corner to save the audio file directly to your local device.
    
- **From Olares Files**:
    1. Open **Files**.
    2. Go to the following path: `/Home/AI/output/acestepv2`.
    3. Right-click the generated audio file and save it to your local device.


## Optimize your audio

ACE-Step offers powerful tools to refine and modify specific parts of your generated track.

### Regenerate the entire segment

You can generate a new version of the entire track.

1. Click the **retake** tab.
2. Adjust the **variance** slider to control how different the new version will be. The higher the value, the more different the song will be.
3. Click **Retake** and wait for the generation.
4. Click the **Play** button below to preview the style change.
    ![Preview the retake](../public/images/manual/use-cases/ace-step-retake.png#bordered)

### Regenerate a specific section

You can update only a selected time range while keeping the rest of the track unchanged.

1. Click the **repainting** tab.
2. Adjust the **Variance** slider to control the degree of change in the new generation. The higher the value, the more different the song will be.
3. Adjust the slider under **Repaint Start Time** and **Repaint End Time** to set the period for the section you want to regenerate.
4. Select the source for repainting:
    - `text2music`: The original song generated via Text2Music.
    - `last_repaint`: The previous repainted version.
    - `upload`: The audio you uploaded.
5. Click **Repaint** and wait for the generation.
6. Click the **Play** button below to preview the result.
    ![Preview the repaint](../public/images/manual/use-cases/ace-step-repaint.png#bordered)

### Edit lyrics

You can edit lyrics to modify specific lines without affecting the rest of the track.

1. Click the **edit** tab.
2. Copy the original lyrics and paste them into the **Edit Lyrics** area.
3. Modify only the specific lines of the lyrics you wish to change.
4. Under **Edit Type**, select `only_lyrics`.
5. Click **Edit** and wait for the generation.
6. Click the **Play** button below to preview the change.
    ![Edit lyrics](../public/images/manual/use-cases/ace-step-edit-lyrics.png#bordered)

### Edit tags

You can edit tags to reset the style or timbre of the track.

1. Click the **edit** tab.
2. Enter the new style or timbre tags (e.g., `hard rock` or `male tenor vocals`) in the **Edit Tags** area.
3. In **Edit Type**, select `remix`.
4. Click **Edit** and wait for the generation.
5. Click the **Play** button below to preview the change.
    ![Edit tags](../public/images/manual/use-cases/ace-step-edit-tags.png#bordered)

### Extend the audio

You can extend the length of the original track by adding new audio before or after it.

1. Click the **extend** tab.
2. Adjust the slider under **Left Extend Length** to add new generation *before* the original audio.
3. Adjust the slider under **Right Extend Length** to add new generation *after* the original audio.
4. Select the source to extend:
    - `text2music`: The original song generated via Text2Music.
    - `last_extend`: The previous extended version.
    - `upload`: The audio you uploaded.
5. Click **Extend** and wait for the generation.
6. Click the **Play** button below to preview the change.
    ![Extend tags](../public/images/manual/use-cases/ace-step-extend.png#bordered)

## Audio2Audio

You can create a new track based on a **reference audio** clip you upload. It analyzes characteristics such as timbre, rhythm, and style to produce a track with a similar feel.
1. Check the box to **Enable Audio2Audio**.
2. Upload an existing music clip to serve as the reference.
3. Adjust the **Refer audio strength** slider. A higher value results in music more similar to the reference track.
4. Select a **Preset** style, or keep the default.
5. Set other parameters as needed.
6. Click **Generate** to create new music with an atmosphere similar to the reference audio.
    ![Audio2Audio](../public/images/manual/use-cases/ace-step-audio2audio.png#bordered)
