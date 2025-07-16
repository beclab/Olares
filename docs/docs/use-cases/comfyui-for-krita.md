---
description: Learn how to integrate ComfyUI with Krita for AI-powered digital art creation. Step-by-step guide to setting up and connecting ComfyUI in Olares with Krita for seamless creative workflows.
---
# AI art creation with ComfyUI and Krita
Running ComfyUI locally on Olares gives you the flexibility of server-side AI processing, but making it work seamlessly with your creative tools requires additional steps. Instead of confining ComfyUI to a single device, Olares allows you to extend its functionality to other machines, enabling smooth integration with tools such as Krita for editing and refinement.

This tutorial will show you how to connect a locally hosted ComfyUI instance on Olares to Krita running on a separate computer. By combining the power of ComfyUI with Krita, you'll be able to create a streamlined, AI-driven workflow that fits naturally into your creative process.

## Objectives
In this tutorial, you will learn how to:
- Deploy and configure ComfyUI in Olares to maximize performance and resource efficiency.
- Integrate ComfyUI with Krita to create AI-generated artwork seamlessly.

## Understanding the components
Your AI art studio consists of three key pieces working together:

* **ComfyUI**: The AI engine running in your Olares environment that powers image generation.
* **Krita**: Professional-grade digital art software where you'll create and edit your artwork.
* **Krita AI Diffusion Plugin**: The connector that enables seamless communication between Krita and ComfyUI.

## What you'll need
Before starting, ensure you have:
* A working Olares installation with internet access
* A computer connected to the same local network as Olares
* Sufficient system resources (recommended: 16GB RAM for optimal performance)

## Set up ComfyUI

1. Install ComfyUI Shared from Market. 
   - For administrator users, this installs both ComfyUI Launcher (the management UI) and ComfyUI (the client UI).
   - For member users, this only installs ComfyUI.
    
   ![Intall ComfyUI](/images/manual/use-cases/install-comfyui.png){width=40%}

2. The administrator configures and launches the ComfyUI service from the ComfyUI Launcher.

3. Configure ComfyUI access policy.

   a. Open Settings, navigate to **Applications** > **ComfyUI Shared** > **Entrances**.

   b. Set the **Authentication level** for ComfyUI to **Internal**.
   ![ComfyUI authentication level](/images/manual/use-cases/comfyui-authentication-level.png#bordered){width=70%}
4. Launch ComfyUI from your desktop, and verify the installation by generating a sample image.
5. Copy the address of ComfyUI for next steps.
:::tip
For security, you should always run AI applications with [LarePass VPN enabled](../manual/larepass/private-network.md).
:::

## Set up Krita

1. Download [Krita](https://krita.org/en/download/).
2. Download the [Krita AI Diffusion plugin](https://github.com/Acly/krita-ai-diffusion/releases).
3. Launch Krita, and navigate to **Tools** > **Scripts** > **Import Python Plugin from File**, and select the downloaded ZIP package.
   ![Import AI plugin](/images/manual/use-cases/krita-import-plugin.png#bordered){width=70%}
4. Confirm the plugin activation and restart Krita.
5. Open Krita, and verify the installation in **Settings** > **Configure Krita** > **Python Plugin Manager**.
   ![Verify AI plugin](/images/manual/use-cases/krita-verify-plugin.png#bordered)
## Connect Krita to ComfyUI
Establish a secure connection between Krita and ComfyUI:
1. Create a new document in Krita.
   :::tip
   Start with a 512 x 512 pixel canvas to optimize performance and manage graphics memory efficiently.
   :::
2. Click **Settings** > **Dockers** > **AI Image Generation** to enable the plugin. You could position the panel to where it's convenient.
   ![Enable AI plugin](/images/manual/use-cases/krita-enable-plugin.png#bordered)
3. Click **Configure** to access the plugin settings.
   ![Configure AI plugin](/images/manual/use-cases/krita-configure-plugin.png#bordered){width=70%}
4. Set up ComfyUI connection.

   a. In **Connection**, select **Custom Server**, and paste your ComfyUI URL.
   
   b. Click **Connect** to verify the connection. A green "Connected" indicator confirms successful connection.
   ![Connect ComfyUI](/images/manual/use-cases/krita-comfyui-connected.png#bordered)
   :::info
   If connection fails:
   - Verify network connectivity between your computer and Olares.
   - Confirm ComfyUI's authentication level is set to "Internal".
   - Confirm LarePass VPN is enabled.
   - Check for and disable any interfering proxy services.
   - Ensure ComfyUI is running correctly on your Olares.
   :::
5. Adjust ComfyUI settings.

   a. In **Styles**, configure your preferred style templates and select appropriate model checkpoints.

   b. Keep default values for other settings unless you need specific optimizations.

## Create AI art with text prompts
Now comes the exciting part - creating AI-generated artwork using natural language prompts.

1. Enter your prompts in the text box, and click **Generate**. 
2. Browse through the generated image variations.
3. Select a preferred result, and click **Apply** to save it to the canvas.
   ![Generate AI art](/images/manual/use-cases/krita-generate-ai-art.png#bordered)
:::tip
If the results aren't quite what you want, you could:
- Create additional variations with new generations.
- Fine-tune the generation parameters.
- Refine your text prompt for more precise results.
- Experiment with different style settings.
:::
