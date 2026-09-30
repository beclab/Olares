---
outline: [2, 3]
description: Generate images and videos through Lares conversations using workflows prepared in FlowStudio.
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Lares, Olares Router, conversational image generation, video generation
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

# Use FlowStudio through Lares

Lares lets you request image and video generation in a conversation. Router connects the request to an available FlowStudio capability, and FlowStudio runs the generation workflow.

## Prerequisites

- Olares OS 1.12.7 or later.
- FlowStudio, Router, and Lares installed from Market. Update existing installations to the latest available versions.
- Qwen3.8-27B (llama.cpp) installed, with the model download complete.
- Required FlowStudio scenes created, with model downloads and initialization complete. For setup instructions, see [Generate images and videos in FlowStudio](flowstudio-create.md).

:::warning Important: Run one task at a time
When running Qwen3.8-27B (llama.cpp) on Olares One, we recommend running only one request at a time to ensure the best experience with the 102K context window and model precision.
:::

:::warning Avoid running FlowStudio generation and Qwen3.8-27B chats at the same time
When using Qwen3.8-27B (llama.cpp) in Lares, wait for any image or video generation task started directly in FlowStudio to finish before sending a message. Running both at the same time can cause GPU memory conflicts and make the model unresponsive.
:::

## Get started

1. Open Lares, select **Qwen3.8-27B (llama.cpp)**, and confirm the permission level.

   ![Lares chat interface](/images/manual/use-cases/lares1.png#bordered)

2. Ask Lares which image and video generation capabilities are available:

   ```text
   What image and video generation capabilities are currently available through FlowStudio?
   ```

3. Review the available capabilities, then describe what you want to create. To use a specific generation model, include its name in your prompt. Otherwise, let Lares choose based on your request and the available models.

## Generate an image

1. Describe the image you want. This example specifies the generation model:

   ```text
   Use Qwen-Image 2.1 NVFP4 · Text to Image to generate a 16:9 automotive poster. Show a red vintage coupe on a coastal road, with a seaside village and turquoise sea in the background. Use warm late-afternoon lighting and add the headline "TIMELESS DRIVE".
   ```

2. Wait for generation to finish, then open the image to review the result.

## Generate a video

1. Describe the scene, motion, and sound you want. This example lets Lares select an available generation model:

   ```text
   Generate a 5-second video at 480p. A red vintage coupe drives slowly along a coastal road at sunset. The camera tracks alongside it, keeping the whole car visible as its wheels turn. Add a gentle engine sound and soft instrumental music. No dialogue or text.
   ```

2. Wait for generation to finish, then play the video to review the result.

## Learn more

- [FlowStudio](flowstudio.md#how-the-apps-work-together): Understand how FlowStudio, Router, and Lares work together.
- [Lares](lares.md): Learn how to use Lares to manage Olares and run research tasks.
- [Generate images and videos in FlowStudio](flowstudio-create.md): Use the scenes directly and prepare example media.
