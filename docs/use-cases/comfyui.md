---
outline: [2, 3]
description: Install ComfyUI on Olares, download essential models via ComfyUI Launcher, and generate your first AI image using the default workflow.
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, Stable Diffusion, AI image generation, self-hosted, ComfyUI Launcher
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-03-17"
---
# ComfyUI

ComfyUI is a powerful, node-based interface for Stable Diffusion that transforms AI image generation into a visual programming experience. By connecting different nodes together like building blocks, you gain precise control over every aspect of the generation process, from prompts and models to post-processing effects.

## Learning objectives

In this guide, you will learn how to:
- Install ComfyUI Shared and understand its components.
- Download the essential Stable Diffusion model package via ComfyUI Launcher.
- Start the ComfyUI service and generate your first image using the default workflow.

## Prerequisites
- A working Olares installation with a GPU and sufficient disk space to download models.
- Admin privileges to install ComfyUI from Market and to start the ComfyUI service.

## Install ComfyUI

1. Open **Market** and search for "ComfyUI".
2. Click **Get**, then click **Install**, and wait for installation to complete.

   ![Install ComfyUI](/images/one/comfyui.png#bordered)

After installation, you will see two icons on Launchpad:
- **ComfyUI**: The client interface where you build workflows and generate images.
- **ComfyUI Launcher**: The management dashboard for the administrator. You must use this tool to start the ComfyUI service before anyone in the cluster can use the client.

:::info Member users
Member users will only see the ComfyUI client icon. The administrator must start the service from the Launcher before members can access ComfyUI.
:::

## Download the essential model package

Before generating images, you need to prepare models. This guide uses Stable Diffusion v1.5 as an example. ComfyUI Launcher provides a one-click package that includes VAEs, utility models, and preview decoders.

1. Open **ComfyUI Launcher** from Launchpad.
2. Scroll down to the **Package installation** section.
3. Find **Stable Diffusion base package** and click **VIEW**.

   ![Stable Diffusion base package](/images/manual/use-cases/comfyui-base-package1.png#bordered)

4. On the package details page, click **GET ALL** to start downloading. You can track the progress in the status bar.

   ![Download progress](/images/manual/use-cases/comfyui-download-progress1.png#bordered)

## Start the ComfyUI service

1. In ComfyUI Launcher, click **START** in the upper-right corner.

   ![Start ComfyUI](/images/manual/use-cases/comfyui-start-service.png#bordered)

   :::tip Initialization time
   The initial startup typically takes 10–20 seconds as the environment initializes.
   :::

2. Once the status changes to "Running", click **OPEN** to launch the ComfyUI client in a new browser tab.

## Generate your first image

The ComfyUI client loads with a default text-to-image workflow. This workflow contains all the basic nodes you need to generate an image.

![Default workflow](/images/manual/use-cases/comfyui-default-workflow.png#bordered)

The key nodes in the default workflow:

| Node | Purpose |
|:-----|:--------|
| **Load Checkpoint** | Loads the base Stable Diffusion model. |
| **CLIP Text Encode** (positive) | Your text description of what you want to generate. |
| **CLIP Text Encode** (negative) | Text describing what to avoid in the generated image. |
| **Empty Latent Image** | Sets the image dimensions and batch size. |
| **KSampler** | Controls the generation process (steps, CFG scale, sampler). |
| **Save Image** | Displays and stores the output image. |

### Set your prompts

1. In the **positive prompt** node (CLIP Text Encode), enter a description of what you want to generate. For example:
   ```text
   a purple glass bottle, studio lighting, high detail, product photography
   ```

2. In the **negative prompt** node, enter elements you want to avoid. For example:
   ```text
   blurry, low quality, distorted
   ```

### Run the workflow

1. Click **Run** in the toolbar to start generation.
2. Wait for the process to complete. The generated image appears in the **Save Image** node.

   ![Generated image](/images/manual/use-cases/comfyui-generated-image.png#bordered)

   You can right-click the image in the **Save Image** node to save it locally, or find all output files in the Files app at `External/olares/ai/output/comfyui`.

   ![Check generated image in Files](/images/manual/use-cases/comfyui-check-generated-image-in-files.png#bordered)


## Troubleshooting

### Cannot access ComfyUI Launcher

If you open ComfyUI Launcher and see an error message saying the connection cannot be established:

- Go to **Settings** > **GPU**.
- If you are using **Memory slicing**, make sure ComfyUI is bound to the GPU and has enough VRAM allocated.
- If you are using **App exclusive**, make sure the exclusive app is set to ComfyUI.

Wait a while and then open ComfyUI Launcher from Launchpad again.

## Learn more

- [Manage ComfyUI using ComfyUI Launcher](comfyui-launcher.md): Control the ComfyUI service, manage models, plugins, and Python dependencies.
- [AI art creation with ComfyUI and Krita](comfyui-for-krita.md): Integrate ComfyUI with Krita for AI-powered digital art workflows.
