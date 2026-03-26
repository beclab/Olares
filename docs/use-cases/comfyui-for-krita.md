---
outline: [2, 3]
description: Integrate ComfyUI with Krita for AI-powered digital art creation. Connect your Olares-hosted ComfyUI to Krita and generate AI artwork seamlessly.
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, Krita, AI art, digital painting, Krita AI Diffusion, image generation
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-03-20"
---

# Create AI art with ComfyUI and Krita

ComfyUI provides powerful AI image generation, but to make it truly useful, you need to integrate it into your creative workflow. This guide shows you how to connect ComfyUI running on Olares to Krita on your computer, so you can generate AI artwork directly within your digital painting environment.

## Learning objectives

In this guide, you will learn how to:
- Deploy and configure ComfyUI in Olares to maximize performance and resource efficiency.
- Install and configure the Krita AI Diffusion plugin.
- Connect Krita to your Olares-hosted ComfyUI instance.
- Generate AI artwork using text prompts in Krita.

## Prerequisites

- A working Olares installation with [ComfyUI Shared installed and running](comfyui.md).
- [Krita](https://krita.org/en/download/) installed on your computer.
- Sufficient system resources on your Olares device to download models.

## Set up ComfyUI

Open ComfyUI Launcher and click **START** to make sure the service is running.

:::tip Maximize GPU performance
You can set the GPU mode to **App exclusive** and assign ComfyUI full GPU access in **Settings** > **GPU** to ensure maximum performance.
:::

## Get the endpoint for ComfyUI

1. On Olares, open Settings, then go to **Application** > **ComfyUI Shared**.
2. In **Entrances**, click **ComfyUI**.
3. Make sure its **Authentication level** is set to **Internal**.
4. Click **Set up endpoint**, then copy the endpoint URL displayed.

   ![Set up endpoint](/images/manual/use-cases/comfyui-set-up-endpoint.png#bordered)

## Download and enable AI Diffusion plugin
1. Download the [Krita AI Diffusion plugin](https://github.com/Acly/krita-ai-diffusion/releases).
2. Launch Krita, and from the toolbar, select **Tools** > **Scripts** > **Import Python Plugin from File**.

   ![Import AI plugin](/images/manual/use-cases/krita-import-plugin1.png#bordered)

3. Select the downloaded ZIP package.
4. When prompted, confirm the plugin activation and restart Krita.

   ![Confirm plugin activation](/images/manual/use-cases/krita-comfirm-plugin-activation.png#bordered){width=40%}

5. After restarting, verify the installation in **Krita** > **Preferences** > **Python Plugin Manager**.

   ![Verify AI plugin](/images/manual/use-cases/krita-verify-plugin.png#bordered)

## Connect Krita to ComfyUI

The connection steps depend on whether your computer and Olares device are on the same network.

<tabs>
<template #Use-.local-domain-(LAN,-recommended)>

If your computer is on the same local network as Olares, you can use the `.local` domain to connect without LarePass VPN. The steps below use macOS as an example, where `.local` domains work natively with no additional setup.

:::info Windows users
On Windows, multi-level `.local` domains need a bit of extra setup. Try one of these:
- **Import hosts in LarePass**: Open the LarePass desktop app and use the built-in option to import Olares hosts to your system.
- **Use the single-level domain**: Change `https://806ba3e40.{username}.olares.com` to `http://806ba3e40-{username}-olares.local`.  

For details, see [Access Olares services locally](../manual/best-practices/local-access.md).
:::

1. Create a new document in Krita.

   :::tip Canvas size
   Start with a 512 x 512 pixel canvas to optimize performance and manage graphics memory efficiently.
   :::

2. Click **Settings** > **Dockers** > **AI Image Generation** to enable the plugin. You can position the panel wherever it's convenient.

   ![Enable AI plugin](/images/manual/use-cases/krita-enable-plugin.png#bordered)

3. Click **Configure** to access the plugin settings.

   ![Configure AI plugin](/images/manual/use-cases/krita-configure-plugin1.png#bordered){width=70%}

4. Set up the ComfyUI connection:

   a. In **Connection**, select **Custom Server**, and paste your ComfyUI URL.

   b. Change the URL to use the `.local` domain and `http`. For example, if the copied URL is:
      ```plain
      https://806ba3e40.laresprime.olares.com
      ```
      Change it to:
      ```plain
      http://806ba3e40.laresprime.olares.local
      ```

   c. Click **Connect** to verify the connection.

   ![Connect ComfyUI](/images/manual/use-cases/krita-missing-required-nodes-local.png#bordered)

   You might see an error message indicating the connection is established but the server is missing required custom nodes or models. This is expected. Proceed to [Prepare models and plugins](#prepare-models-and-plugins) to download the required resources.

</template>

<template #Use-.com-domain>

If your computer is not on the same local network as Olares, enable LarePass VPN to ensure a secure connection.

1. Enable LarePass VPN on the LarePass desktop client:

   a. Open the LarePass app and click your avatar in the top-left corner to open the user menu.

   b. Toggle on the switch for **VPN connection**.

   Once enabled, make sure the connection status is either **Intranet** (LAN) or **P2P** (outside LAN).

2. Create a new document in Krita.

   :::tip Canvas size
   Start with a 512 x 512 pixel canvas to optimize performance and manage graphics memory efficiently.
   :::

3. Click **Settings** > **Dockers** > **AI Image Generation** to enable the plugin. You can position the panel wherever it's convenient.

   ![Enable AI plugin](/images/manual/use-cases/krita-enable-plugin.png#bordered)

4. Click **Configure** to access the plugin settings.

   ![Configure AI plugin](/images/manual/use-cases/krita-configure-plugin1.png#bordered){width=70%}

5. Set up the ComfyUI connection:

   a. In **Connection**, select **Custom Server**, and paste your ComfyUI URL.

   b. Click **Connect** to verify the connection.

   ![Connect ComfyUI](/images/manual/use-cases/krita-missing-required-nodes-com.png#bordered)

   You might see an error message indicating the connection is established but the server is missing required custom nodes or models. This is expected. Proceed to [Prepare models and plugins](#prepare-models-and-plugins) to download the required resources.

</template>
</tabs>

## Prepare models and plugins

### Install required custom nodes

The Krita AI Diffusion plugin requires the following custom nodes:
- ControlNet preprocessors: https://github.com/Fannovel16/comfyui_controlnet_aux
- IP-Adapter: https://github.com/cubiq/ComfyUI_IPAdapter_plus
- Inpaint nodes: https://github.com/Acly/comfyui-inpaint-nodes
- External tooling nodes: https://github.com/Acly/comfyui-tooling-nodes

To install them:

1. Open ComfyUI Launcher, and go to **Plugins** > **Custom Install**.
2. Paste the GitHub URL of the custom node (e.g., `https://github.com/Acly/comfyui-tooling-nodes`), and click **INSTALL PLUGIN**.
3. Repeat for each of the remaining custom nodes listed above.
   ![Download custom nodes](/images/manual/use-cases/comfyui-download-custom-nodes.png#bordered)
4. Go back to the **Home** page in ComfyUI Launcher, then click **RESTART** for the changes to take effect.

5. Optional: If you go back to Krita and click **Connect** again, you should see an error message indicating that required models are still missing.

   ![Missing required models](/images/manual/use-cases/krita-missing-required-models.png#bordered)

### Install required models

The plugin needs these utility models to work properly. Without them, some features will not function correctly. Pre-installing them ensures a smoother experience.

- NMKD Superscale: https://huggingface.co/gemasai/4x_NMKD-Superscale-SP_178000_G/resolve/main/4x_NMKD-Superscale-SP_178000_G.pth
- OmniSR X2: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X2_DIV2K.safetensors
- OmniSR X3: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X3_DIV2K.safetensors
- OmniSR X4: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X4_DIV2K.safetensors
- MAT Inpaint: https://huggingface.co/Acly/MAT/resolve/main/MAT_Places512_G_fp16.safetensors

To download them via ComfyUI Launcher:

1. Open ComfyUI Launcher, and go to **Models** > **Custom Download**.
2. Download upscale models:

   a. Paste the NMKD Superscale URL, set **Destination folder** to **Upscale Models**, and click **DOWNLOAD MODEL**.
      ![Download required upscale models](/images/manual/use-cases/comfyui-download-upscale-models.png#bordered)

   b. Repeat for the three OmniSR models using the same destination folder.

3. Download the inpaint model:

   a. Paste the MAT Inpaint URL.

   b. Set **Destination folder** to **Custom Directory**.

   c. Enter `inpaint` as the **Directory Name**.

   d. Click **DOWNLOAD MODEL**.
      ![Download required inpaint model](/images/manual/use-cases/comfyui-download-inpaint-model.png#bordered)
4. Go back to the **Home** page in ComfyUI Launcher, then click **RESTART** for the changes to take effect.
5. Optional: If you go back to Krita and click **Connect** again, you should see an error message indicating that base models are still missing.

   ![Missing required models](/images/manual/use-cases/krita-missing-base-models.png#bordered)

### Install a base diffusion model

At least one diffusion model (commonly called a “checkpoint”) is required. This guide uses Z-Image Turbo as an example. Z-Image Turbo is a medium-sized model that balances quality and speed, producing realistic images without requiring excessive memory.

1. Open ComfyUI Launcher, and scroll down to the **Package installation** section.
2. Find **Z-Image Turbo Package** and click **VIEW**.

   ![Z-Image Turbo Package](/images/manual/use-cases/comfyui-zimage-turbo-package.png#bordered)

3. On the package details page, click **GET ALL** to start downloading. You can track the progress in the status bar.

   ![Download progress](/images/manual/use-cases/comfyui-download-progress-z-image.png#bordered)

4. Go back to the **Home** page in ComfyUI Launcher, then click **RESTART** for the changes to take effect.
5. In Krita, go to **Connection** > **Server Configuration** and click **Connect** again. A green “Connected” indicator confirms a successful connection. You should see Z-Image marked as “supported” in the base model list.

   ![Z-Image detected](/images/manual/use-cases/comfyui-z-image-detected.png#bordered)

## Add a style

Before generating images, you need to create a Style Preset that tells Krita which model to use.

1. Open the **Configure Image Diffusion** dialog in Krita, and go to the **Styles** tab.
2. For **Style Presets**, select **Z-Image Turbo** from the built-in styles.

   ![Select built-in Z-Image Turbo style](/images/manual/use-cases/krita-select-built-in-style.png#bordered)

3. Click the duplicate icon to create a duplicate of the current style.

   ![Duplicate style](/images/manual/use-cases/krita-duplicate-style.png#bordered)

4. For **Model Checkpoint**, select the Z-Image model. The model name should be `public/z_image_turbo_bf16`.

   ![Select Z-Image model](/images/manual/use-cases/krita-select-z-image-model.png#bordered)

5. Click the refresh icon to refresh the available styles.
   ![Refresh style list](/images/manual/use-cases/krita-refresh-style-list.png#bordered)

6. Keep default values for other settings, and click **Ok** to save changes.
   :::warning
   It is recommended to use the default settings if you are not familiar with Krita. Changing the default settings might generate unexpected results.
   :::

## Create AI art with text prompts

1. In the **AI Image Generation** panel, confirm that the Z-Image Turbo style is selected.

2. Enter your prompts in the text box. For example: 

   ```plain
   A person relaxing on a sandy beach, basking in the warm sunlight, with the calm blue ocean in the background.
   ```

3. Click **Generate**. The generated image appears on the canvas.
   ![Generate image](/images/manual/use-cases/krita-generated-image-1.png#bordered)

4. Click **Generate** again to generate a new image.
   ![Generate image](/images/manual/use-cases/krita-generated-image-2.png#bordered)

5. Select a preferred result, and click **Apply** to add it to the layers.

## Inpaint

To refine specific areas of a generated image, use inpainting. This lets you modify parts of the image while keeping the rest intact.

1. Select the freehand selection tool and draw around the area you want to modify.
   ![Use the freehand selection tool](/images/manual/use-cases/krita-use-selection-tool.png#bordered)

2. Enter a description for what you want in the selected area. For example:
   ```plain
   Seagulls can be seen flying in the distant sky.
   ```

3. Click **Fill**. Several fill candidates will appear in the panel.
   ![Fill candidates](/images/manual/use-cases/krita-fill.png#bordered)
4. Click each candidate to preview it on the canvas.

5. When you find a result you like, click **Apply** to add it to the layers. 
   ![Select inpaint candidate](/images/manual/use-cases/krita-select-inpaint-candidate.png#bordered)

## Troubleshooting

### Cannot connect to ComfyUI from Krita

If Krita shows a connection error:

| Check | What to do |
|:------|:-----------|
| Network connectivity | Make sure your computer and Olares are on the same network. |
| ComfyUI authentication level | In **Settings** > **Application** > **ComfyUI Shared**, confirm it is set to **Internal**. |
| LarePass VPN for `.com` URLs | Enable **VPN connection** in the LarePass desktop app. |
| Interfering proxy/VPN | Temporarily disable other VPN or proxy software. |
| ComfyUI service status | Open ComfyUI Launcher and verify the service is **Running**. |
| GPU access | In **Settings** > **GPU**, verify ComfyUI is bound to the GPU with enough<br> VRAM allocated. |
| Required plugins and models | Make sure all custom nodes, utility models, and base diffusion models<br> are downloaded and ComfyUI has been restarted. |

## Learn more

- [ComfyUI quick start guide](comfyui.md): Install ComfyUI and generate your first image.
- [Manage ComfyUI using ComfyUI Launcher](comfyui-launcher.md): Control the ComfyUI service, manage models, and configure the environment.
- [Krita AI Diffusion documentation](https://github.com/Acly/krita-ai-diffusion/wiki): Explore advanced features and workflows.
