---
outline: [2, 3]
description: Prepare FlowStudio scenes, Router, and Lares to request image and video generation through a conversation.
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Lares, Olares Router, conversational image generation, video generation
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

# Use FlowStudio through Lares (in Draft)

Lares lets you request image and video generation in a conversation. Router connects the request to an available FlowStudio capability, and FlowStudio runs the generation workflow. For the roles of all three apps, see [FlowStudio](flowstudio.md#how-the-apps-work-together).

## Prerequisites

- Olares System 1.12.7.
- FlowStudio, Router, and Lares installed from Market. For existing installations, update the apps to the latest available versions before continuing.
- Qwen3.8-27B (llama.cpp) installed for Lares, as described in the [Lares guide](lares.md#prerequisites).
- The FlowStudio scenes you want to call created and initialized, with their required models downloaded.

The LLM used by Lares interprets your request. The image and video models used by FlowStudio perform the generation. Installing the Lares LLM alone does not prepare FlowStudio's generation capabilities.

## Prepare the FlowStudio scenes

1. Open FlowStudio and choose the image or video workflow you need in **Workflow Plaza**.
2. Review **Pre-run checks**, then click **Create scene** when the checks pass.
3. In the **Initialization progress** panel that opens on the right, wait for **Init finished**. Model downloads marked **Completed** alone do not mean the scene is ready.
4. Open **Create** > **Scenes** and check that the scene appears.

For the full setup and example inputs, see [Generate images and videos in FlowStudio](flowstudio-create.md).

## Request generation in Lares

The next part of this guide will cover how to check that the prepared capabilities are available through Router, request image generation in Lares, and request video generation with the required inputs.

<!-- TODO: Verify Router capability availability, Lares input controls, image and video prompts, task progress, and result handling with screenshots before adding actionable steps. Do not infer Lares controls from the FlowStudio UI. -->

## Learn more

- [Lares](lares.md#get-started): Select a model and confirm its permission level.
- [Generate images and videos in FlowStudio](flowstudio-create.md): Use the scenes directly and prepare example media.
