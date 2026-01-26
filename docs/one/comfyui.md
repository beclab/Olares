---
outline: [2,3]
description: Learn how to install ComfyUI, manage models via the ComfyUI Launcher, and generate high-performance images and videos on Olares One.
head:
  - - meta
    - name: keywords
      content: local ai, comfyui, nunchaku, wan
---
# Generate images and videos with ComfyUI <Badge type="tip" text="1 h" />
ComfyUI is a powerful, node-based interface for Stable Diffusion that transforms AI media generation into a visual programming experience.

Olares provides ComfyUI Shared to allow multiple users to share models, plugins, and workflow resources within the cluster. It also features ComfyUI Launcher, providing administrator users with a simple way to manage ComfyUI resources and runtime environments.

## Learning objectives
- Install and configure the ComfyUI service.
- Use ComfyUI Launcher to download optimized model packages.
- Generate high-quality images using the Nunchaku Flux.1-dev workflow.
- Generate videos using the Wan 2.2 model.

## Before you begin
For image generation:
- Olares One is equipped with an NVIDIA RTX 5090 mobile GPU. This allows you to leverage the Nunchaku Flux.1-dev model to generate images significantly faster than standard FP16 or FP8 versions.
- Nunchaku Flux.1-dev is an optimized model utilizing SVDQuant quantization (NVFP4). It is designed to deliver high-performance inference on this specific hardware while maintaining minimal loss in visual quality.

## Prerequisites
**Hardware** <br>
- Olares One connected to a stable network.
- Sufficient disk space to download models.

**User permissions**
- Admin privileges to install ComfyUI from the Market, and to start or stop the ComfyUI service for the cluster.

## Step 1: Install ComfyUI
1. Open Market, and search for "ComfyUI."
2. Click **Get**, then **Install**, and wait for installation to complete.

   ![Install ComfyUI](/images/one/comfyui.png#bordered)

After it finished, you can see two icons for ComfyUI on Launchpad:
- **ComfyUI**: The client interface for ComfyUI.
- **ComfyUI Launcher**: The core management dashboard. You must use this tool to start the ComfyUI service before you can use the client.

## Step 2: Download model package
To generate content, you must first download the specific model package.

1. Open ComfyUI Launcher, and scroll down to **Package installation**.
2. For image generation, select **Nunchaku Flux FP4 Package** and click **GET ALL**. This downloads the necessary Nunchaku plugin and quantized models.
3. For video generation, select **Wan 2.2 Text to Video 14B Package** and click **GET ALL** to download necessary models.

   ![Install model package](/images/one/comfyui-install-model-package.png#bordered)

:::info ComfyUI service restart required
Whenever you install a new model package or custom node, you must restart the ComfyUI service for the changes to take effect.
:::

## Step 3: Start the ComfyUI service
1. In ComfyUI Launcher, click **START** in the upper-right corner.
   ![Start ComfyUI](/images/one/comfyui-start.png#bordered)
   :::tip Initialization time
   The initial startup typically takes 10–20 seconds as the environment initializes.
   :::
2. Once the status changes to "Running," click **OPEN**. This will launch the ComfyUI client in a new browser tab.

## Step 4: Generate images
This section uses the `nunchaku-flux.1-dev-qencoder` workflow to help you get started.

1. In the ComfyUI client, click the **ComfyUI** icon in the top-left corner to open the menu.
2. Select **View** > **Browse Templates**.
3. Under **EXTENSIONS**, select **ComfyUI-nunchaku**. 
4. Choose the template: **nunchaku-flux.1-dev-qencoder**.
   ![Open Nunchaku workflow](/images/one/comfyui-nunchaku-templates.png#bordered)

5. Prepend `public/` to the filename in each model loader node. For example:
   - **Default**: `clip_I.safetensors`
   - **Change to**: `public/clip_I.safetensors`
   :::info Shared model path
   ComfyUI in Olares uses a file structure that differs from the standard installation. This change allows models to be shared between ComfyUI and SD Web UI.
   :::
   ![Change model path](/images/one/comfyui-nunchaku-change-model-path.png#bordered)

6. Replace the text in **CLIP Text Encode** to update the prompt for your image. For example:
    ```plain
   8-bit cyberpunk: Blocky pixel cat holds "olares is fast!" neon on pixel street.
    ```
7. Click **Run** in the toolbar to start generation.

## Step 5: Generate videos
This section uses the Wan 2.2 workflow.

1. In ComfyUI, open the ComfyUI menu and select **View** > **Browse Templates**.
2. Under the **GENERATION TYPE** category, select **Video**. 
3. Select the template: **Wan 2.2 14B Text to Video**.
   ![Open Wan 2.2 workflow](/images/one/comfyui-video-templates.png#bordered)

4. Prepend `public/` to the filename in each model loader node. For example:
   - **Default**: `wan_2.1_vae.safetensors`
   - **Change to**: `public/wan_2.1_vae.safetensors`
     :::info Shared model path
     ComfyUI in Olares uses a file structure that differs from the standard installation. This change allows models to be shared between ComfyUI and SD Web UI.
     :::
   ![Change model path](/images/one/comfyui-wan-change-model-path.png#bordered){width=90%}

5. Locate the **CLIP Text Encode** node and enter a detailed prompt for your video.
    ```plain
    A woman with long brown hair and light skin smiles at another woman with long blonde hair. The woman with brown hair wears a black jacket and has a small, barely noticeable mole on her right cheek. The camera angle is a close-up, focused on the woman with brown hair's face. The lighting is warm and natural, likely from the setting sun, casting a soft glow on the scene. The scene appears to be real-life footage.
    ```
6. Click **Run** in the toolbar to start generation. Video generation takes significantly longer than image generation.


## Step 6: Download output files
You can download all output images and videos from Olares One to your local computer.
1. Open the Files app.
2. Navigate to the following directory:
    ```plain
    External/olares/ai/output/comfyui
    ```
   ![View output files](/images/one/comfyui-output.png#bordered)
3. Select the files you wish to save.
4. Right-click and select **Download** to save them to your local computer.

## Resources
- [ComfyUI official docs](https://docs.comfy.org/)
- [ComfyUI-nunchaku docs](https://nunchaku.tech/docs/ComfyUI-nunchaku/index.html)
- [Manage ComfyUI using ComfyUI Launcher](../use-cases/comfyui-launcher.md)