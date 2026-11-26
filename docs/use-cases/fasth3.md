---
outline: [2, 3]
description: Generate videos with synchronized audio using FastH3 on Olares. Create clips from text, frames, and reference media.
head:
  - - meta
    - name: keywords
      content: Olares, FastH3, MiniMax H3, AI video generation, text to video, image to video, reference to video, self-hosted
app_version: "0.1.15"
doc_version: "1.0"
doc_updated: "2026-11-26"
---

<script setup>
import { withBase } from 'vitepress'
</script>

# Generate videos with FastH3

FastH3 is a local video generation app on Olares that uses the eight-step distilled models of MiniMax H3 to generate synchronized video and audio. You can create videos from text, animate existing images, or guide video generation with reference images, videos, and audio.

Through a visual interface, you can create and refine videos without building ComfyUI workflows yourself. AI prompt optimization uses a chat model connected through Router to help refine your descriptions.

## Learning objectives

- Install FastH3 and prepare the video generation models.
- Generate videos from text, frames, and reference media.
- Extract and reuse video frames to create new videos.
- Find and manage your media assets.

## Prerequisites

- Olares 1.12.8 or later and an administrator account for app installation.
- An NVIDIA GPU supported by the app. The package requests 22 GiB of GPU memory and 32 GiB of system memory.
- Sufficient disk space for the app, media assets, and model weights. The default NVFP4 weights require approximately 47 GB of downloads.

## Install FastH3

1. Open Market and search for "FastH3".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Prepare the models

Open FastH3 from Launchpad. On first startup, the app automatically downloads the required NVFP4 model weights by default.

To check the model download progress, open **Settings** > **Models** in FastH3 and click **Open the Model Console**. Wait until the console shows **Ready** and the generation service is ready before submitting a task.

FastH3 shares model files with other ComfyUI apps and FlowStudio on your Olares, allowing existing model files to be reused.

FastH3 is a [shared app](../manual/olares/market/shared-apps.md). An administrator only needs to install it once for all users on the same Olares. Each user's personal generation queue, uploads, and results remain private.

## Generate videos

FastH3 supports three video generation modes:

| Mode | Input | When to use |
| --- | --- | --- |
| Text | Text prompt | Generate a video from an idea without existing media. |
| Frames | Start frame, optional end frame, and prompt | Animate an image or guide the transition between two frames. |
| References | Reference images, video, audio clips, and prompt | Guide video generation using existing media assets. |

Use the perfume commercial examples below to create a video from text, reuse a video frame to create a different shot, and combine the product image with new reference media. Substitute your own product and desired changes as needed.

The three examples use **Random** for **Seed** and do not use **AI optimize**. Your generated videos will vary.

### Generate a video from text

Use **Text** mode to generate a video from a description without uploading reference media.

1. Go to **Create** and select **Text**.
2. In **Prompt**, describe the subject, setting, movement, and visual style of the video. For example:

   ```text
   A cinematic product shot of a minimalist transparent glass perfume bottle filled with amber liquid, standing on a cream-colored stone pedestal.

   The front of the bottle features only one word: "PERFUME", printed in simple, clean, black uppercase sans-serif letters. The text is centered, clearly legible, and remains unchanged throughout the video. No brand names, logos, additional text, numbers, or decorative typography.

   Warm sunlight casts soft shadows across the background. The camera slowly pushes toward the bottle as delicate reflections shift across the glass. Minimalist composition, realistic glass materials, refined luxury aesthetic, and soft studio lighting.
   ```

3. Optionally, use **AI optimize** to refine your description. It supports three options: structuring and translating the prompt into English, structuring it while keeping the original language, or translating it into English only.

   This feature requires a chat model available in [Olares Router](olares-router.md), which you can select under **Settings** > **AI optimize** in FastH3. To prepare a local chat model, see [Run local AI models with Engine Base apps](llm-base-apps.md).

4. Configure the generation settings:

   | Setting | This tutorial | How to choose |
   | --- | --- | --- |
   | Duration | 5 s | Set the video duration from 2 to 15 seconds. |
   | Aspect | 16:9 | Select an aspect ratio that suits your intended use, or use Auto. |
   | Quality | 480p | Select 720p for higher-resolution videos. |
   | Video count | ×1 | Generate one to four videos per submission. |
   | Seed | Random | Under Advanced settings, enable Lock seed to reuse the same seed across generations. |

5. Click **Generate** to submit the task. When available, the button displays the estimated generation time.
6. In **Results**, monitor the generation progress, elapsed time, and estimated time remaining.
7. Once the task is complete, open the result to preview the video.

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="Perfume video generated from a text prompt">
  <source :src="withBase('/images/manual/use-cases/fasth3-text.mp4')" type="video/mp4" />
  Your browser does not support embedded video. <a :href="withBase('/images/manual/use-cases/fasth3-text.mp4')">Download the example video</a>.
</video>

### Animate an image with Frames

Use **Frames** mode to animate an existing image. In this example, reuse a frame from the previous video to create a different camera movement.

1. Open the video generated in the previous section and navigate to a frame that clearly shows the perfume bottle.
2. Click **Grab frame** and select **Use as first frame** to use the captured image as the start frame.
3. In **Prompt**, describe the new camera movement:

   ```text
   The camera slowly moves from left to right around the perfume bottle. Warm sunlight creates subtle shifting reflections across the glass and amber liquid. Keep the bottle's original shape, proportions, color, and label consistent throughout the video. Elegant product cinematography with smooth, controlled camera movement.
   ```

4. Set **Duration** to 5 s, **Quality** to 480p, and **Video count** to ×1.

   In **Frames** mode, the output aspect ratio follows the start frame.

5. Click **Generate** and preview the result.

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="Perfume video animated from a captured frame">
  <source :src="withBase('/images/manual/use-cases/fasth3-frames.mp4')" type="video/mp4" />
  Your browser does not support embedded video. <a :href="withBase('/images/manual/use-cases/fasth3-frames.mp4')">Download the example video</a>.
</video>

You can also upload your own image as the start frame. Each frame supports one image of up to 512 MB in JPG, PNG, WebP, still GIF, or HEIC/HEIF format.

To guide how the video ends, add an optional **End frame** and describe the transition between the start and end frames in your prompt.

### Generate a video from references

Use **References** mode to guide the content of a video with existing media assets rather than specifying its exact start and end frames.

In this example, combine the perfume bottle from the previous video with a reference image of white jasmine flowers to create a new advertising scene.

References mode supports the following media:

| Media type | Maximum count | Duration |
| --- | --- | --- |
| Images | 5 | — |
| Video | 1 | 2–15 s |
| Audio clip | 1 | 2–15 s |

You can add up to seven reference files, with a maximum size of 512 MB per file. To view supported file formats, click **File formats** under **Add media**.

1. Open the video generated in the previous section and navigate to a frame that clearly shows the perfume bottle.
2. Click **Grab frame** and select **Use as reference image**.
3. Download the jasmine flower image attached below. Under **Add media**, click **+** and upload the downloaded image.

   <a :href="withBase('/images/manual/use-cases/fasth3-jasmine.jpg')" download="fasth3-jasmine.jpg">Download the jasmine flower reference image (JPG)</a>

4. In **Prompt**, type `@` and select a reference asset. Describe how each asset should be used in the video. For example:

   ```text
   Create an elegant cinematic perfume commercial featuring the perfume bottle from @image_1 and the white jasmine flowers from @image_2.

   Preserve the original design, shape, and amber color of the perfume bottle. Arrange the jasmine flowers naturally around its base on a cream-colored stone surface. A few delicate white petals drift slowly through the air.

   Warm golden sunlight, soft shadows, realistic glass reflections, and a slow cinematic camera pullback. Refined luxury fragrance advertising aesthetic.
   ```

5. Replace `@image_1` and `@image_2` with the actual media references by typing `@` and selecting the corresponding assets.
6. Set **Duration** to 5 s, **Aspect** to 16:9, **Quality** to 480p, and **Video count** to ×1.
7. Click **Generate** and preview how the reference assets are incorporated into the video.

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="Perfume video generated using bottle and jasmine references">
  <source :src="withBase('/images/manual/use-cases/fasth3-references.mp4')" type="video/mp4" />
  Your browser does not support embedded video. <a :href="withBase('/images/manual/use-cases/fasth3-references.mp4')">Download the example video</a>.
</video>

You can also add video or audio clips as references to guide the visual content, movement, or sound of a new video. Describe how each reference should be used in your prompt.

## Manage your videos and assets

### Reuse and extend videos

Open a completed video to access the following actions:

- **Reuse**: Restore the inputs and generation settings of a previous task. Modify them as needed and generate again.
- **Continue**: Use the last frame of a generated video as the start frame for a new video.
- **Grab frame**: Extract the selected frame and choose how to use it:
  - **Use as first frame**: Use it as a start frame.
  - **Use as last frame**: Use it as an end frame.
  - **Use as reference image**: Use it as a reference image.
  - **Library only**: Save it to Library for later use.
- **Download**: Save the generated video.

### Browse and download media

The **Results** panel displays generation tasks and their outputs. Use search or **Filters** to find tasks by generation mode, resolution, or status.

Go to **Library** to browse uploaded media and generated content. You can filter assets by media type and source.

Your generation queue, uploaded media, and results are stored on your Olares and remain accessible across devices. Uploaded media and generated videos are retained across app restarts.

### Use personal and project spaces

Use the workspace selector to switch between personal and project spaces.

- **My space**: A private workspace containing your personal generation queue, uploads, and results.
- **New project**: Create a project and share its content with selected users or everyone on your Olares.

FastH3 stores files under `Common/Studio`, with a separate folder for each user.

## Manage model weights

FastH3 provides two model weight formats:

| Precision | Download size | Description |
| --- | --- | --- |
| NVFP4 | Approximately 47 GB | The default model set, optimized for faster generation. |
| INT8 | Approximately 63 GB | An alternative model weight format. |

In FastH3, go to **Settings** > **Models** to check the installation status of each model set.

You can download additional model weights to switch between available precisions or delete unused model sets to free up disk space.

Before deleting a model set, check whether other ComfyUI apps or FlowStudio need its weights. Deleting the set removes its weights from the shared model folder even if another app uses them.

## Open workflows in ComfyUI

For advanced customization, go to **Settings** > **Open in ComfyUI** in FastH3 and select **Text**, **Frames**, or **References** to open the corresponding workflow in the node editor.

For more information, see [Run ComfyUI locally for AI image generation](comfyui.md).

## Common issues

### Why can't I generate videos after opening the app?

The required models might still be downloading. In FastH3, go to **Settings** > **Models** and click **Open the Model Console** to check the download status. Wait until the console shows **Ready** before generating videos.

### Why is AI optimize unavailable?

Make sure a chat model is available in [Router](olares-router.md). In FastH3, go to **Settings** > **AI optimize**, click **Refresh**, and select an available chat model.

You can still generate videos using your own prompts without AI optimization.
