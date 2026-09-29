---
outline: [2, 3]
description: Generate and edit images with Qwen-Image 2.1 in FlowStudio, then create videos from first and last frames or reference media.
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Qwen-Image 2.1, FastVideo FastH3, MiniMax H3, image generation, image editing, image to video, reference to video, NVFP4
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

# Generate images and videos in FlowStudio

Generate images, edit them, and create videos using ready-made workflows in FlowStudio.

## Tutorial overview

This tutorial uses a coastal car poster as a running example across four scenarios: image generation, image editing, first-and-last-frame video generation, and reference-based video generation.

| Scenario | Workflow | What you will create |
| --- | --- | --- |
| Text-to-image generation | Qwen-Image 2.1 NVFP4 · Text to Image | A coastal poster featuring a red vintage car. |
| Image editing | Qwen-Image 2.1 NVFP4 · Image Edit | A version of the poster with blue car paint. |
| First-and-last-frame video generation | FastVideo FastH3: Image to Video | A video using the blue and red posters as its first and last frames. |
| Reference-based video generation | MiniMax H3 NVFP4 · Reference to Audio-Video | A cartoon coastal drive using the original poster and generated video as references. |

## Prerequisites

- Olares OS 1.12.7 or later

## Install FlowStudio

If you installed FlowStudio previously, update it to the latest version before continuing.

1. Open Market and search for "FlowStudio".
   ![FlowStudio in Market](/images/manual/use-cases/flowstudio.png#bordered)

2. Click **Get**, then **Install**, and wait for the installation to complete.

## Generate an image from text

Use **Qwen-Image 2.1 NVFP4 · Text to Image** to generate images from text prompts, with support for text rendering, realistic textures, and fine details.

### Prepare the scene

1. Open FlowStudio, go to **Workflow Plaza**, and select **Image**.
2. Search for `Qwen-Image 2.1`, and then select **Qwen-Image 2.1 NVFP4 · Text to Image**.
3. Review **Pre-run checks**. When it shows **Checks passed**, click **Create scene**.

   ![Text to Image workflow dependencies and passing pre-run checks](/images/manual/use-cases/flowstudio-create-scene.png#bordered)

   :::tip Insufficient GPU memory
   If the check reports insufficient GPU memory, stop other GPU-heavy apps or unassign their GPU resources, then check again.
   :::

4. The **Initialization progress** panel opens on the right. Wait until it shows **Init finished**.

   Model downloads marked **Completed** do not by themselves mean initialization is finished. Wait if the panel still shows **Provisioning in progress**.

   ![Initialization progress panel showing Init finished](/images/manual/use-cases/flowstudio-init-finished.png#bordered)

### Configure and generate the image

1. Open **Create** > **Scenes**, and then select **Qwen-Image 2.1 NVFP4 · Text to Image**.
2. Enter a prompt. The example below creates a coastal poster featuring a red vintage car and the headline "TIMELESS DRIVE":

   ```text
   Create a cinematic 16:9 automotive travel poster.

   A cherry-red vintage two-door coupe is parked at a scenic coastal overlook, shown in full from a low front three-quarter angle, with its front pointing left. Position the car across the lower center of the frame.

   Show finely detailed chrome trim, realistic tire tread, silver alloy wheels, subtle leather upholstery visible through the windows, and reflections of the sky and coastline in the glossy paint.

   In the middle distance, a winding road follows rugged cliffs past a small seaside village with colorful facades. A weathered stone wall borders the overlook. In the foreground, textured pavement and a few tufts of coastal grass add detail. Beyond the cliffs, turquoise water reveals submerged rocks, with white surf along the shore.

   Late-afternoon sunlight creates warm highlights on the car, cool shadows, and atmospheric depth in the distant coastline. Keep the car sharply defined and the background slightly softer, but recognizable.

   Place the exact headline "TIMELESS DRIVE" in elegant, clearly legible ivory uppercase letters in the upper-left sky area. Keep the headline modest in size and clear of the car and village.

   Photorealistic automotive photography with refined travel-poster typography. No people, brand logos, or additional text.
   ```

3. Click **Parameters** and configure the image:

   | Parameter | Setting |
   | --- | --- |
   | **Aspect ratio** | Select **16:9** to match the prompt. |
   | **Megapixels** | Choose an output resolution from the list. |
   | **Batch** | Select **1** for this example. |

4. Click **Generate** and wait for the task to finish.
5. Open the generated image to check the details.

   ![Generated poster with a red vintage car on a coastal overlook and the headline TIMELESS DRIVE](/images/manual/use-cases/flowstudio-generated-poster.png#bordered)

6. Hover over the generated image and click the download icon to download it for the next section.

## Edit an image

Use **Qwen-Image 2.1 NVFP4 · Image Edit** to modify an existing image with text instructions. You can use a result from FlowStudio or another existing image.

### Prepare the scene

1. Open FlowStudio, go to **Workflow Plaza**, and select **Image**.
2. Search for `Qwen-Image 2.1`, and then select **Qwen-Image 2.1 NVFP4 · Image Edit**.
3. Review **Pre-run checks**. When it shows **Checks passed**, click **Create scene**.
4. In the **Initialization progress** panel that opens on the right, wait for **Init finished**. Text to Image and Image Edit use the same three model files. Models already installed in the library are reused rather than downloaded again.

### Configure and edit the image

This example uses the red-car poster from the previous section and asks the model to change only the car's paint color.

1. Open **Create** > **Scenes**, and then select **Qwen-Image 2.1 NVFP4 · Image Edit**.
2. Click **Parameters** and configure the image:

   | Parameter | Setting |
   | --- | --- |
   | **Editable image** | Click **Choose reference**, then select **Upload** to upload the red-car poster. |
   | **Aspect ratio** | Select **16:9** to match the original poster. |
   | **Megapixels** | Choose an output resolution from the list. |
   | **Batch** | Select **1** for this example. |

3. Enter a prompt. The example below changes the car's paint from red to blue:

   ```text
   Change only the car's red paint to a rich metallic blue. Preserve the car's body shape, chrome trim, wheels, windows, and interior. Keep the coastal village, sea, mountains, road, camera angle, and composition unchanged. Preserve the exact text "TIMELESS DRIVE" in its original position and style. Maintain the warm late-afternoon lighting, with natural highlights and reflections on the new blue paint.
   ```

4. Click **Generate**. When the task finishes, open the edited image and check that the requested changes have been applied.

   ![Edited poster with blue car paint and the coastal setting retained](/images/manual/use-cases/flowstudio-edited-poster.png#bordered)

5. Hover over the edited image and click the download icon. Use this blue-car poster as the first frame in the next section.

## Generate a video from first and last frames

Use **FastVideo FastH3: Image to Video** to create a video with a specified starting and ending image.

### Prepare the scene

1. Open FlowStudio, go to **Workflow Plaza**, and select **Video**.
2. Search for `FastVideo FastH3`, and then select **FastVideo FastH3: Image to Video**.
3. Review **Pre-run checks**. When it shows **Checks passed**, click **Create scene**.
4. In the **Initialization progress** panel that opens on the right, wait for **Init finished**. As with the image scenes, completed model downloads alone do not mean the scene is ready.

### Configure and generate the video

This example uses the blue-car poster as the first frame and the red-car poster as the last frame. The prompt asks the car to move forward while its paint changes from blue to red.

1. Open **Create** > **Scenes**, and then select **FastVideo FastH3: Image to Video**.
2. Click **Parameters** and configure the video:

   | Parameter | Setting |
   | --- | --- |
   | **First & last frames** | Click **Choose image** to upload the blue-car poster under **First frame** and the red-car poster under **Last frame**. JPG and PNG images are supported. |
   | **Megapixels** | **480p** |
   | **Duration** | **5** seconds |

3. Enter a prompt. The example below asks the car to move forward while its paint changes from blue to red:

   ```text
   Create a continuous tracking shot, starting with the blue car in the first frame and ending with the red car in the last frame.

   The car drives slowly forward along the coastal road, toward the left side of the scene. Its wheels rotate naturally. The camera tracks alongside it at the same speed, keeping the car at a consistent size and position in the frame while the road and coastal scenery move gently behind it.

   As the car moves, its paint gradually changes from metallic blue to the red shown in the last frame. Complete the color transition by the end of the clip.

   Preserve the car's body shape and details. Maintain warm late-afternoon lighting and realistic reflections. Keep the "TIMELESS DRIVE" headline fixed and legible as a screen overlay. Smooth motion, no cuts, no sudden acceleration, and no body deformation.
   ```

4. Click **Generate** and wait for the task to finish.

   The following video is the example output:

   <video controls playsinline preload="none" poster="/images/manual/use-cases/flowstudio-edited-poster.png" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="Example video generated from the blue and red car posters">
   <source src="/images/manual/use-cases/flowstudio-car-color-transition.mp4" type="video/mp4" />
   Your browser does not support embedded video. <a href="/images/manual/use-cases/flowstudio-car-color-transition.mp4">Download the example video</a>.
   </video>

5. Download the generated video to use as a reference in the next section.

## Generate a video from reference media

Use **MiniMax H3 NVFP4 · Reference to Audio-Video** to generate video with audio using reference media and a text prompt.

### Prepare the scene and references

1. Open FlowStudio, go to **Workflow Plaza**, and select **Video**.
2. Search for `MiniMax H3`, and then select **MiniMax H3 NVFP4 · Reference to Audio-Video**.
3. Review **Pre-run checks**. When it shows **Checks passed**, click **Create scene**.
4. In the **Initialization progress** panel that opens on the right, wait for **Init finished**.

### Configure the reference inputs

This example reuses the car video from the previous section and the original red-car poster to request a cartoon version of the coastal drive. The prompt uses the video as a motion reference and the poster as a reference for the car and setting, rather than specifying first and last frames.

1. Open **Create** > **Scenes**, and then select **MiniMax H3 NVFP4 · Reference to Audio-Video**.
2. Click **Parameters** and configure the video:

   | Parameter | Setting |
   | --- | --- |
   | **Reference image group** | Click the add button and select the original red-car poster. |
   | **Reference videos** | Click the add button and select the video generated in the previous section. |
   | **Reference audio** | Leave empty for this example. |
   | **Aspect ratio** | **16:9 (Widescreen)**, matching the poster's landscape composition. |
   | **Megapixels** | **480p** |
   | **Frame rate** | **24** |
   | **Duration** | **5** seconds |
   | **Quick Generation** | Leave enabled for this example. |

   :::warning Important: Video duration and GPU memory
   Use 480p for reference-based video generation. In the setup used for this guide, generating more than 5 seconds with a reference video can exhaust GPU memory. Keep the output duration at 5 seconds or less when supplying a reference video. With image and audio references but no reference video, the reported duration limit for this setup is 15 seconds. These are setup-specific limits, not universal model limits.
   :::

   <!-- Before publication, confirm the tested hardware and image/audio-only duration limit. Do not infer them from model specifications. -->

### Prompt a cartoon scene

1. Enter a prompt. The example below requests a 3D cartoon version of the coastal drive:

   ```text
   Create a playful 5-second 3D animated short of a vintage two-door coupe driving along a coastal road, in a 16:9 landscape composition.

   Use the reference video as a guide for the car's forward motion and the tracking camera. Use the reference image for the car's body shape, red paint, chrome trim, silver wheels, and coastal village setting. Keep the paint red throughout; do not repeat the video's blue-to-red color transition.

   Reimagine the entire scene as a colorful 3D cartoon world, including the car, winding road, stone wall, seaside village, cliffs, and turquoise sea. Use softly rounded forms, stylized materials, vibrant colors, and warm late-afternoon lighting rather than photorealism.

   The car drives slowly forward toward the left, with its wheels visibly rotating. The camera tracks alongside it from a front three-quarter angle, keeping the whole car visible while the coastal scenery moves gently behind it. Add a subtle suspension bounce as the car travels along the road.

   Use a single continuous shot with no cuts or sudden camera movements. Keep the car's proportions consistent. Do not add eyes, a mouth, people, or other vehicles. Omit the poster's headline and add no text.

   Add a cheerful engine sound and light, playful background music. No dialogue, captions, or added text.
   ```

2. Click **Generate** and wait for the task to finish.
3. Play the video to review the result.

<!-- The cartoon prompt is a proposed example. Add an output video only after the result has been reviewed. -->

## Learn more

- [Qwen-Image 2.1](https://huggingface.co/Qwen/Qwen-Image-2.1): Read the official model description.
- [FlowStudio overview](flowstudio.md): Compare direct creation with calling FlowStudio through Lares.
- [Use FlowStudio through Lares](flowstudio-lares.md): Prepare the apps and scenes for conversational creation.
