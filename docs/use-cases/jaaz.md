---
outline: [2, 3]
description: Set up Jaaz on Olares as a private AI design assistant. Connect it to Ollama for text generation and ComfyUI for image generation to create a fully local, end-to-end design workflow.
head:
  - - meta
    - name: keywords
      content: Olares, Jaaz, AI design, Canva alternative, self-hosted, Ollama, ComfyUI, image generation, privacy
app_version: "1.0.13"
doc_version: "1.0"
doc_updated: "2026-03-26"
---

# Create AI-powered designs with Jaaz

Jaaz is an open-source, multimodal AI design assistant that serves as a privacy-first alternative to tools like Canva. By running Jaaz on Olares, you can connect it to locally hosted AI models through Ollama and ComfyUI, creating a fully private, end-to-end AI design workflow.

## Learning objectives

In this guide, you will learn how to:
- Connect Jaaz to a local Ollama service for text generation.
- Connect Jaaz to a local ComfyUI service for image generation.
- Configure and use a complete local workflow to generate your first AI design.

## Prerequisites

- [Ollama installed](ollama.md) with a Qwen3 model downloaded.

  :::info Recommended text models
  This guide uses Qwen3 as an example. For GPUs with more than 12 GB of VRAM, use `qwen3:14b`. For 12 GB or less, use `qwen3:8b`. We recommend using models from the Qwen series as the text model for Jaaz, as they offer the best compatibility with Jaaz's design instruction parsing. Other models might produce inconsistent results or fail to generate proper design task breakdowns.
  :::

- [ComfyUI Shared installed](comfyui.md) with a working image generation workflow.

  :::tip Model setup for this guide
  This guide uses the Flux.1 Dev FP8 quantized workflow as an example. Make sure you have placed the model file at the following path, which corresponds to the `checkpoints` folder in the ComfyUI configuration:
  ```
  External/ai/model/main/flux1-dev-fp8.safetensors
  ```
  :::

## Install Jaaz

1. Open **Market** and search for "Jaaz".
2. Click **Get**, then click **Install**, and wait for installation to complete.

   <!-- ![Install Jaaz](/images/manual/use-cases/jaaz-install.png#bordered) -->

3. Once installed, click **Open** to launch the app.

:::info
Since Jaaz runs locally, if a login prompt appears, close it by clicking the **X** in the upper-right corner of the login screen.
:::

## Connect Ollama

1. Open Olares Settings, then go to **Applications** > **Ollama**.
2. Under **Shared entrances**, click **Ollama API**, and then copy the endpoint address.

   ![Get Ollama shared endpoint](/images/manual/use-cases/obtain-ollama-hosturl2.png#bordered){width=70%}

3. In the Jaaz interface, click the **Settings** icon in the upper-right toolbar.
4. In the Ollama section, paste the endpoint address into the **API URL** field.
5. Leave the **API Key** field empty.
6. Click **Save settings**.

Return to the Jaaz main page and refresh. You should now see your local Ollama models in the **Model** dropdown.

<!-- ![Ollama models available in Jaaz](/images/manual/use-cases/jaaz-ollama-models.png#bordered) -->

## Prepare a ComfyUI workflow

Before connecting ComfyUI to Jaaz, you need an API-format workflow file.

1. Open ComfyUI Shared from Launchpad. If the service is stopped, click **Start**, then click **Open**.

2. Load a working workflow and click **Run** to verify it generates images correctly.

3. Once confirmed, click the **ComfyUI** icon in the upper-right corner, then select **File** > **Export (API)** to save the workflow as a JSON file.

## Connect ComfyUI

1. Open Olares Settings, then go to **Applications** > **ComfyUI Shared**.
2. In **Entrances**, click **ComfyUI**.
3. Make sure its **Authentication level** is set to **Internal**.
4. Click **Set up endpoint**, then copy the endpoint URL displayed.

   <!-- ![Get ComfyUI shared endpoint](/images/manual/use-cases/jaaz-comfyui-shared-endpoint.png#bordered) -->

5. In the Jaaz interface, click the **Settings** icon in the upper-right toolbar.
6. In the ComfyUI section, paste the endpoint URL into the **API URL** field.
7. Click **Save settings**.

## Configure the workflow

1. On the Jaaz settings page, click **Add workflow** and upload the JSON file you exported earlier.
2. Enter a workflow description, such as `flux_dev_checkpoint_example`.
3. Click **Add input** to define the parameters Jaaz can control in the workflow. For example, you can add the following inputs:

   | Input | Description | Required |
   |:------|:-----------|:---------|
   | Text prompt | The text prompt for image generation | Yes |
   | Width | Image width in pixels | No |
   | Height | Image height in pixels | No |
   | Quantity | Number of images to generate | No |

   The text prompt input must be added and marked as **Required**. Other inputs are optional and depend on your workflow.

   <!-- ![Configure workflow inputs](/images/manual/use-cases/jaaz-comfyui-workflow-inputs.png#bordered) -->

4. Click **Submit** to save the settings.

Return to the main page. You should now see the workflow you added in the **Model** dropdown.

<!-- ![ComfyUI workflow available in Jaaz](/images/manual/use-cases/jaaz-comfyui-workflow-available.png#bordered) -->

## Generate your first design

1. On the main page, select the **Qwen3** text model and your image generation workflow from the dropdowns.

2. Enter a design prompt, for example: `Draw a cute corgi`.

3. Press **Enter**.

Jaaz opens the design canvas. On the right panel, you can see Jaaz analyzing your request and breaking down the task.

After a moment, the generated image appears on the canvas. You can move, resize, and continue building your design.

<!-- ![Generated design on canvas](/images/manual/use-cases/jaaz-generate-design.png#bordered) -->

### Iterate on your design

Select a generated image and click **Add to chat** to let Jaaz iterate and refine the design based on your follow-up instructions.

<!-- ![Iterate on design](/images/manual/use-cases/jaaz-iterate-design.png#bordered) -->

### Export design assets

When your design is ready, select the assets you need and click **Export** in the upper-right corner to download them.

<!-- ![Export design assets](/images/manual/use-cases/jaaz-export-design.png#bordered) -->

## Troubleshooting

### Configuration lost after Jaaz restarts

If your Ollama and ComfyUI connection settings disappear after a Jaaz restart, reconfigure them by following the [Connect Ollama](#connect-ollama) and [Connect ComfyUI](#connect-comfyui) sections above.

### Image generation errors

If you encounter errors or long response times during image generation:

1. Open **Control Hub**.
2. Check the container logs for Ollama and ComfyUI to identify the specific error.

## Learn more

- [Manage ComfyUI using ComfyUI Launcher](comfyui-launcher.md): Control the ComfyUI service, manage models, plugins, and dependencies.
- [ComfyUI](comfyui.md): Install ComfyUI, download models, and generate your first AI image.