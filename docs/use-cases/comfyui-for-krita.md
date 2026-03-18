---
outline: [2, 3]
description: Integrate ComfyUI with Krita for AI-powered digital art creation. Connect your Olares-hosted ComfyUI to Krita and generate AI artwork seamlessly.
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, Krita, AI art, digital painting, Krita AI Diffusion, image generation
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-03-18"
---

# Create AI art with ComfyUI and Krita

Running ComfyUI locally on Olares gives you the flexibility of server-side AI processing, but making it work seamlessly with your creative tools requires additional steps. Instead of confining ComfyUI to a single device, Olares allows you to extend its functionality to other machines, enabling smooth integration with tools such as Krita for editing and refinement.

This tutorial will show you how to connect a locally hosted ComfyUI instance on Olares to Krita running on a separate computer. By combining the power of ComfyUI with Krita, you'll be able to create a streamlined, AI-driven workflow that fits naturally into your creative process.

## Learning objectives

In this guide, you will learn how to:
- Deploy and configure ComfyUI in Olares to maximize performance and resource efficiency.
- Install and configure the Krita AI Diffusion plugin.
- Connect Krita to your Olares-hosted ComfyUI instance.
- Generate AI artwork using text prompts in Krita.

## Before you begin

- A working Olares installation with ComfyUI Shared installed and running.
- A computer connected to the same local network as Olares.
- Krita installed on your computer.
- Sufficient system resources (recommended: 16GB RAM for optimal performance).

## Understanding the components

Your AI art studio consists of three key pieces working together:

* **ComfyUI**: The AI engine running in your Olares environment that powers image generation.
* **Krita**: Professional-grade digital art software where you'll create and edit your artwork.
* **Krita AI Diffusion Plugin**: The connector that enables seamless communication between Krita and ComfyUI.

## Set up ComfyUI

Open ComfyUI Launcher and click **START** to make sure the service is running.

::: tip Maximize GPU performance
You can set the GPU mode to **App exclusive** and assign ComfyUI full GPU access in **Settings** > **GPU** to ensure maximum performance.
:::

## Get the endpoint for ComfyUI

1. On Olares, open Settings, then go to **Application** > **ComfyUI Shared**.
2. In **Entrances**, click **ComfyUI**.
3. Make sure its **Authentication level** is set to **Internal**.

   ![ComfyUI authentication level](/images/manual/use-cases/comfyui-authentication-level.png#bordered){width=70%}

4. Click **Set up endpoint**, then copy the endpoint URL displayed.

   ![Set up endpoint](/images/manual/use-cases/comfyui-set-up-endpoint.png#bordered)

## Set up Krita
1. Download and install [Krita](https://krita.org/en/download/).
2. Download the [Krita AI Diffusion plugin](https://github.com/Acly/krita-ai-diffusion/releases).
3. Launch Krita, and navigate to **Tools** > **Scripts** > **Import Python Plugin from File**. Select the downloaded ZIP package.

   ![Import AI plugin](/images/manual/use-cases/krita-import-plugin.png#bordered){width=70%}

4. Confirm the plugin activation and restart Krita.

   ![Confirm plugin activation](/images/manual/use-cases/krita-confirm-plugin.png#bordered){width=70%}

5. After restarting, verify the installation in **Krita** > **Preferences** > **Python Plugin Manager**.

   ![Verify AI plugin](/images/manual/use-cases/krita-verify-plugin.png#bordered)

## Connect Krita to ComfyUI

<tabs>
<template #Use-.local-domain-(LAN,-recommended)>

If your client device is on the same local network as Olares, you can use the `.local` domain. You will need to slightly modify the URL you copied earlier.

1. Create a new document in Krita.

   ::: tip Canvas size
   Start with a 512 x 512 pixel canvas to optimize performance and manage graphics memory efficiently.
   :::

2. Click **Settings** > **Dockers** > **AI Image Generation** to enable the plugin. You can position the panel wherever it's convenient.

   ![Enable AI plugin](/images/manual/use-cases/krita-enable-plugin.png#bordered)

3. Click **Configure** to access the plugin settings.

   ![Configure AI plugin](/images/manual/use-cases/krita-configure-plugin.png#bordered){width=70%}

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

   ![Connect ComfyUI](/images/manual/use-cases/krita-comfyui-connected.png#bordered)

   You might see an error message indicating the connection is established but the server is missing required custom nodes or models. This is expected. Proceed to [Prepare models and plugins](#prepare-models-and-plugins) to download the required resources.

</template>

<template #Use-.com-domain>

If your client device is not on the same local network as Olares, you need to enable LarePass VPN to ensure a secure connection.

1. Enable LarePass VPN on the LarePass desktop client:

   a. Open the LarePass app and click your avatar in the top-left corner to open the user menu.

   b. Toggle on the switch for **VPN connection**.

   Once enabled, make sure the connection status is either **Intranet** (LAN) or **P2P** (outside LAN).

2. Create a new document in Krita.

   ::: tip Canvas size
   Start with a 512 x 512 pixel canvas to optimize performance and manage graphics memory efficiently.
   :::

3. Click **Settings** > **Dockers** > **AI Image Generation** to enable the plugin. You can position the panel wherever it's convenient.

   ![Enable AI plugin](/images/manual/use-cases/krita-enable-plugin.png#bordered)

4. Click **Configure** to access the plugin settings.

   ![Configure AI plugin](/images/manual/use-cases/krita-configure-plugin.png#bordered){width=70%}

5. Set up the ComfyUI connection:

   a. In **Connection**, select **Custom Server**, and paste your ComfyUI URL.

   b. Click **Connect** to verify the connection.

   ![Connect ComfyUI](/images/manual/use-cases/krita-comfyui-connected.png#bordered)

   You might see an error message indicating the connection is established but the server is missing required custom nodes or models. This is expected. Proceed to [Prepare models and plugins](#prepare-models-and-plugins) to download the required resources.

</template>
</tabs>

## Prepare models and plugins

### Install required custom nodes

The Krita AI Diffusion plugin requires the following custom nodes:
- [ControlNet preprocessors](https://github.com/Fannovel16/comfyui_controlnet_aux)
- [IP-Adapter](https://github.com/cubiq/ComfyUI_IPAdapter_plus)
- [Inpaint nodes](https://github.com/Acly/comfyui-inpaint-nodes)
- [External tooling nodes](https://github.com/Acly/comfyui-tooling-nodes)

To install them:

1. Open ComfyUI Launcher, and go to **Plugins** > **Custom Install**.
2. Paste the GitHub URL of the custom node (e.g., `https://github.com/Acly/comfyui-tooling-nodes`), and click **INSTALL PLUGIN**.
3. Repeat for each of the remaining custom nodes listed above.

### Install required models

Some utility models are required for the plugin to function properly. Without them, the connection might fail or certain features will not work. Pre-installing them ensures a smoother experience.

| Model | URL | Destination |
|:------|:----|:------------|
| NMKD Superscale | `https://huggingface.co/gemasai/4x_NMKD-Superscale-SP_178000_G/resolve/main/4x_NMKD-Superscale-SP_178000_G.pth` | Upscale Models |
| OmniSR X2 | `https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X2_DIV2K.safetensors` | Upscale Models |
| OmniSR X3 | `https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X3_DIV2K.safetensors` | Upscale Models |
| OmniSR X4 | `https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X4_DIV2K.safetensors` | Upscale Models |
| MAT Inpaint | `https://huggingface.co/Acly/MAT/resolve/main/MAT_Places512_G_fp16.safetensors` | Custom: `inpaint` |

To download them via ComfyUI Launcher:

1. Open ComfyUI Launcher, and go to **Models** > **Custom Download**.
2. Download upscale models:

   a. Paste the NMKD Superscale URL, set **Destination folder** to **Upscale Models**, and click **DOWNLOAD MODEL**.

   b. Repeat for the three OmniSR models using the same destination folder.

3. Download the inpaint model:

   a. Paste the MAT Inpaint URL.

   b. Set **Destination folder** to **Custom Directory**.

   c. Enter `inpaint` as the **Directory Name**.

   d. Click **DOWNLOAD MODEL**.

### Install a base diffusion model

At least one diffusion model (commonly called a “checkpoint”) is required. This guide uses Z-Image Turbo as an example. Z-Image is a medium-sized diffusion model that falls between Flux 1 and Flux 2 4B in terms of memory requirements and speed. The Turbo variant delivers convincing realistic images with reasonable performance.

1. Open ComfyUI Launcher, and scroll down to the **Package installation** section.
2. Find **Z-Image Turbo Package** and click **VIEW**.

   ![Z-Image Turbo Package](/images/manual/use-cases/comfyui-zimage-turbo-package.png#bordered)

3. On the package details page, click **GET ALL** to start downloading. You can track the progress in the status bar.

   ![Download progress](/images/manual/use-cases/comfyui-download-progress1.png#bordered)

### Verify plugins and models in Krita

1. In ComfyUI Launcher, restart the ComfyUI service.
2. Go back to the **Connection** > **Server Configuration** page in Krita, and click **Connect** again. A green “Connected” indicator confirms a successful connection. In the detected base model list, you should see Z-Image marked as “supported”.

3. Adjust ComfyUI settings:

   a. In **Styles**, select **Z-image Turbo** from **Style Presets**.

   b. Keep default values for other settings unless you need specific optimizations.

   c. Click **Ok** to exit.

## Create AI art with text prompts

Now comes the exciting part — creating AI-generated artwork using natural language prompts.

1. Enter your prompts in the text box, and click **Generate**.

2. Browse through the generated image variations.

3. Select a preferred result, and click **Apply** to add it to the canvas.

   ![Generate AI art](/images/manual/use-cases/krita-generate-ai-art.png#bordered)

::: tip Refining results
If the results aren't quite what you want, you can:
- Create additional variations with new generations.
- Fine-tune the generation parameters.
- Refine your text prompt for more precise results.
- Experiment with different style settings.
:::

## FAQ

### Connection cannot be established
If the connection fails:
- Verify network connectivity between your computer and Olares.
- Confirm ComfyUI's authentication level is set to "Internal".
- If you are using `.com` URL, confirm LarePass VPN is enabled.
- Check for and disable any interfering proxy services.
- Ensure ComfyUI is running correctly on Olares.
- Check whether ComfyUI has GPU access.

## Learn more

- [ComfyUI quick start guide](comfyui.md): Install ComfyUI and generate your first image.
- [Manage ComfyUI using ComfyUI Launcher](comfyui-launcher.md): Control the ComfyUI service, manage models, and configure the environment.
- [Krita AI Diffusion documentation](https://github.com/Acly/krita-ai-diffusion/wiki): Explore advanced features and workflows.
