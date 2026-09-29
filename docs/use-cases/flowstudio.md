---
outline: [2, 3]
description: Create images and videos directly in FlowStudio or call its workflows through Lares. Learn how FlowStudio, Router, and Lares work together.
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Lares, Olares Router, image generation, video generation
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

# FlowStudio

FlowStudio lets you generate images, edit existing images, and create videos using ready-made workflows. Choose a workflow to create a scene, and FlowStudio prepares the required models and runtime for you.

## Choose how to create

Create directly in FlowStudio, or ask Lares to call its capabilities through Router.

| Method | How it works | Guide |
| --- | --- | --- |
| Create in FlowStudio | Choose a scene, supply a prompt and any required media, and configure generation parameters in FlowStudio. | [Generate images and videos in FlowStudio](flowstudio-create.md) |
| Create through Lares | Describe your request in a conversation. Lares calls the available FlowStudio capabilities through Router. | [Use FlowStudio through Lares](flowstudio-lares.md) |

## How the apps work together

Whether you create directly in FlowStudio or through Lares, FlowStudio runs the generation workflow. Lares provides a conversational interface and sends requests through Router. Before using either method, create the required scenes in FlowStudio and wait for model downloads and initialization to finish.

| App | Role |
| --- | --- |
| FlowStudio | Prepares workflow scenes and runs image and video generation tasks. |
| Router | Makes FlowStudio capabilities available to Lares and routes generation requests. |
| Lares | Uses an LLM to interpret your request and call the available generation capabilities. |

## Get started

Start with [direct creation in FlowStudio](flowstudio-create.md) to install the app and prepare image and video scenes. To create through a conversation, see [Use FlowStudio through Lares](flowstudio-lares.md).

## Learn more

- [Lares](lares.md): Get started with the Olares AI assistant.
- [Olares Router](olares-router.md): Learn about model and tool routing.
