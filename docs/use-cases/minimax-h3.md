---
outline: [2, 3]
description: Create videos with sound locally using MiniMax H3 on Olares, animate images, continue a shot, and guide new clips with reference media.
head:
  - - meta
    - name: keywords
      content: Olares, MiniMax H3, local video generation, image to video, reference video, AI audio, ComfyUI
app_version: "0.1.7"
doc_version: "1.0"
doc_updated: "2026-11-25"
---

# Create videos with sound using MiniMax H3

MiniMax H3 is a video creation app maintained by Olares. It runs MiniMax H3 models locally to generate video and sound together. Start with a text description, animate a still image, or use reference media to guide a new clip.

AI prompt optimization uses a chat model connected through Router to refine and structure your descriptions before video generation.

## Learning objectives

- Install the app and check that its video models are ready.
- Create a short fashion film from a text description.
- Use frames and reference media to guide another shot.
- Save still frames, reuse settings, and download your videos.

:::info MiniMax H3 or FastH3?
MiniMax H3 offers two speed settings:

- **Quality priority**: Uses 20 generation steps for all three input modes: **Text**, **Frames**, and **References**.
- **Fast**: Uses Lightning LoRA to reduce generation time, with 8 steps for **Text** and **Frames**, or 4 steps for **References**.

FastH3 is a separate app that uses eight-step distilled models.

Choose MiniMax H3 when you want to try ideas quickly, then spend more time on a version you plan to keep. Choose FastH3 when your priority is to try many different prompts and movements quickly. Both support text, frames, and reference media.
:::

## Prerequisites

- Olares 1.12.8 or later and an administrator account for installation.
- A supported NVIDIA GPU. The package requests 22 GiB of GPU memory and 32 GiB of system memory.
- Space for the app, generated videos, and approximately 65 GB of model files. Models already available in the shared model storage can be reused.

## Install MiniMax H3 and prepare the models

1. Open Market and search for "MiniMax H3".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open MiniMax H3 from Launchpad. On first start, the app prepares its video models. Wait for the workspace to open.
4. In **Settings**, click **Open the Model Console**. Wait until it shows **Ready** and ComfyUI shows **up**.

::: tip First generation takes longer
The first generation also loads the models into memory, so it takes longer than later runs. Follow its progress in **Results**.
:::

## Configure a chat model for AI optimization

You can generate videos with your own prompts. To have **AI optimize** structure and refine them, connect a chat model through Router. This guide uses Qwen3.8-27B (llama.cpp), a local chat model.

1. Install "Qwen3.8-27B (llama.cpp)" from Market. When the installation completes, open its Model Console and wait for **Model** to show **Ready** and **Engine** to show **Running**.

   For another model, install its packaged app or follow [Run local LLMs with Engine Base apps](llm-base-apps.md).
2. Open Router from Launchpad, go to **Default models**, and select the Qwen3.8-27B model for `default-chat`. Wait until its row shows **Callable**.
3. In MiniMax H3, go to **Settings** > **AI optimize**, and click **Refresh** to reload the chat models from Router.
4. In **Chat model**, keep **Router default (default-chat)** to follow Router's default, or select a specific model from the list. Confirm that the status shows **Ready**.

   Models marked **sees images** can use your reference images when rewriting the prompt. With a text-only model, enhancement uses the text alone. A specific model selection applies independently of Router's default.

## Generate a video from text

Use **Text** to start with a description alone. This example creates a five-second fashion shot with an ivory outfit, flowing orange silk, and a sunlit sandstone courtyard. The following sections reuse its frames to continue the shot and create a new scene. Substitute your own subject, setting, and desired movement.

1. Go to **Create** and select **Text**.
2. In **Prompt**, describe the subject, movement, camera, lighting, and sound. For example:

   ```text
   A cinematic luxury fashion editorial, one continuous shot. An adult woman with short wavy dark hair and small sculptural gold earrings stands in a sunlit sandstone courtyard. She wears an elegant ivory silk dress. Behind her, a long translucent burnt-orange silk ribbon catches a gentle breeze and flows in broad graceful waves, glowing amber where the sunlight passes through it.

   Medium portrait from the waist up, face clearly visible, subject on the right third and flowing fabric filling the left. She slowly turns her face toward the warm sunlight with a calm expression. The camera makes a slow subtle dolly-in.

   Pale sandstone arches, a clear powder-blue sky, warm late-afternoon backlight, soft reflected light on her face, realistic skin and delicate fabric texture, restrained ivory, terracotta and gold palette, shallow depth of field, 35mm film cinematography. The fabric stays behind her and never covers her face.

   Sound: soft breeze and delicate silk rustling. No dialogue, no music, no text, no logos.
   ```

3. Optionally, click **AI optimize** to refine and structure your description. Review the rewritten prompt and edit any details you do not want before generating.
4. Optionally, to keep the original language or translate only, click the arrow beside **AI optimize** and select **Structure, keep the language** or **Translate to English only**.
5. Set the basic generation options:

   | Setting | <nobr>This example</nobr> | How to choose |
   | --- | --- | --- |
   | **Duration** | **5 s** | Start with a short shot to refine the motion. You can type an exact duration or use a preset. |
   | **Aspect** | **16:9** | Use a wide composition for the subject and flowing fabric. |
   | **Quality** | **720p** | Use **480p** for a smaller preview. |
   | **Video count** | **×1** | Generate one clip while refining your idea. |

6. Optionally, expand **Advanced settings** to adjust generation speed or the seed. You can keep the defaults for this example:

   | Setting | <nobr>This example</nobr> | How to choose |
   | --- | --- | --- |
   | **Speed** | **Quality priority** | Select **Fast** for a quicker preview with fewer generation steps. |
   | **Seed** | **Random** | Keep the default for your first generation. |

7. Click **Generate** and wait for the task to complete.
8. Open the result and check that the face, fabric, camera movement, and sound are as expected before using the clip or continuing from it.

   ![Preview the generated fashion video and its playback, frame capture, and continuation controls](/images/manual/use-cases/minimax-h3-text-result.jpg#bordered)

9. Optionally, click **Reuse** to make another version. The prompt and settings return to the input panel. Adjust them, then click **Generate** again.

## Animate images with Frames

Use **Frames** to animate a photograph or illustration. A start frame sets the opening image, and an optional end frame guides where the shot ends.

1. Go to **Create** and select **Frames**. Click **Start frame** and upload an image or choose one from the library.
2. To use a frame from a generated video instead, pause the video at the desired moment, fine-tuning with **Previous frame** and **Next frame**, then choose **Grab frame** > **Use as first frame** or **Use as last frame**.
3. To guide where the shot ends, add an image under **End frame (Optional)**. Keep the subject and composition compatible with the start frame.
4. In **Prompt**, describe how the subject and camera should move and what should stay the same, such as the person's appearance, outfit, and lighting.
5. Choose the duration and quality, then click **Generate**. **Aspect** follows the start frame.

### Continue from the last frame

To continue the courtyard video from the previous section, use **Continue** to load its last frame as the next clip's start frame.

1. Open the finished video and click **Continue** below the player. This switches to **Frames** and adds the video's last frame as **Start frame**.
2. Replace the prompt with what should happen next. For example:

   ```text
   Continue from this frame in one continuous fashion editorial shot. Keep the same woman, short wavy dark hair, gold earrings, ivory silk outfit, and orange scarf. The camera slowly pulls back to reveal more of the sunlit sandstone arches. She gently turns her gaze toward the courtyard while the translucent orange silk flows to the left in the breeze. Preserve the warm golden light and natural skin detail. Sound: soft wind and silk rustling. No dialogue, no music, no text.
   ```

3. Keep **End frame (Optional)** empty for an open-ended continuation. **Aspect** follows the first frame.
4. Set **Duration** to **5 s** and **Quality** to **720p**. For a quicker motion preview, select **Fast** under **Advanced settings** > **Speed**, then click **Generate**.
5. Play the new result and compare its opening with the previous clip's ending. The continuation is a separate video that you can download and join to the first clip in a video editor.

   ![Continue the courtyard shot with its last frame as the start image](/images/manual/use-cases/minimax-h3-frames-result.jpg#bordered)

## Create a video with References

Use **References** to carry the subject or visual style into a different scene. In this example, a frame from the courtyard video guides a new shot on a seaside terrace.

::: tip Reference limits and multiple inputs
The reference tray accepts up to five images, one video, and one audio clip, with a maximum of 512 MB per file. Reference video and audio clips must be 2–15 seconds long. Click **File formats** beside **Add media** to check supported formats.

When combining inputs, mention each one where you describe its purpose, such as a person image for appearance or a video for motion. Include at least an image or a video. An audio clip alone is not enough for **References**.
:::

1. Open the courtyard video, pause at a clear view of the person, and choose **Grab frame** > **Library only**.
2. Select **References**. Click **Add media** > **From library**, select the saved frame, and click **Choose**. Use **Upload files** instead to start with your own material.
3. In **Prompt**, type `@` and choose the image. Explain what to keep from it and what to change:

   ```text
   Use @[person reference] as the reference for the woman, her short wavy dark hair, gold earrings, ivory silk outfit, and translucent orange scarf. Create a new luxury fashion shot on a white limestone terrace overlooking a calm turquoise Mediterranean sea. She stands on the right of the frame beside a pale stone balustrade, with orange silk flowing gracefully to the left in the sea breeze. The camera slowly glides sideways at eye level as she looks toward the horizon. Clear warm sunlight, luminous natural skin, powder-blue sky, elegant ivory and terracotta palette, realistic fabric, one continuous shot. Sound: gentle sea waves and soft wind. No dialogue, no music, no text or logos.
   ```

   Insert the actual image mention with `@` in place of the bracketed placeholder.

4. Review the prompt and confirm that the image mention points to your selected reference.
5. Set **Aspect** to **16:9**, **Duration** to **5 s**, and **Quality** to **720p**. Choose **Fast** for this preview, then click **Generate**.
6. Open the result and compare the person's face, outfit, and scarf with the reference.

   ![Use a courtyard frame as a reference for a new fashion video on a seaside terrace](/images/manual/use-cases/minimax-h3-reference-result.jpg#bordered)

## Find and download your videos

### Find a video

Choose where to search based on what you remember:

- By prompt: Go to **Create** and use **Search prompts** in **Results**. Use **Filters** to narrow the results.
- By name: Go to **Library**, select **Videos**, and use **Search names**. Use the source filter to narrow the results.

### Download videos

- A single video: Open it from **Results** or **Library**, then click **Download** below the player to save it to your computer.
- Multiple videos: Go to **Library**, select **Videos**, select the checkboxes for the videos you want, and click **Download**.

## Common issues

### Why is the first generation taking longer?

The app loads its models before generating. Check **Settings** > **Open the Model Console** if model preparation has not finished, and follow the task status in **Results**. Higher resolution and longer clips also take more work. Use **Fast** to preview a shot before refining it.

### Why is Generate unavailable in References mode?

Add at least one reference image or video and enter a prompt. Check the file count, size, and duration limits shown beside **Add media**.

### Can I adjust the underlying workflow?

Open **Settings** and find **Open workflows in ComfyUI**. Choose the workflow for **Text**, **Frames**, or **References**, including its fast variant when needed, to open the node editor.

## Learn more

- [Generate and edit images with Qwen Image 2.1](qwen-image-2.1.md): Prepare still images to animate or use as visual references.
- [Olares Router](olares-router.md): Manage the chat model used for prompt optimization.
