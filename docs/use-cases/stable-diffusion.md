---
description: Comprehensive guide to using Stable Diffusion in Olares. Learn about text-to-image generation, and how to optimize your SD Web UI deployment for multi-user environments.
---
# Stable Diffusion
Stable Diffusion represents a groundbreaking implementation of latent diffusion models (LDMs) in AI image synthesis. This deep learning architecture operates by decomposing the image generation process into a lower-dimensional latent space, significantly reducing computational requirements while maintaining high-fidelity output.

Olares simplifies the deployment and management of Stable Diffusion. Unlike traditional deployments that require manual configuration of file systems and databases, Olares shields developers from these infrastructure complexities, allowing you to focus solely on using the model for image generation.

With Olares's multi-user support, team members can share a single Stable Diffusion deployment while maintaining individual data privacy. This approach eliminates the need for redundant system installations that would otherwise consume excessive hardware resources.
## What can Stable Diffusion do?
Whether you're an artist looking to expand your creative toolkit, a developer integrating AI imaging into your workflow, or simply curious about AI art generation, Stable Diffusion offers:

* Text-to-image generation
* Image-to-image transformation
* Inpainting and outpainting capabilities
* Style transfer and artistic modifications
* High-resolution image creation and upscaling

## Install SD Web UI

:::info
Starting from Olares 1.11.6, if "SD Web UI For Cluster" or "SD Web UI" was previously installed, uninstall them before proceeding.
:::

1. Install SD Web UI Shared directly from Market. 

2. Launch SD Web UI from your desktop. Please ensure the admin has already installed SD Web UI Shared.

Now you are ready to unleash your creativity and start generating stunning images!

## Prevent conflicts among members
In SD Web UI, checkpoint settings are globally applied to all users by default. When one user switches to a different checkpoint, all subsequent image generations by other users will also use this newly selected checkpoint. To prevent workflow disruptions in multi-user environments, you could specify checkpoints for individual tasks.

![Checkpoint settings](/images/manual/use-cases/sd-checkpoint.png)
1. Global checkpoint
2. Per-task checkpoint

## Adjust system settings
When launching SD Web UI, the `--xformers` flag is enabled by default to:
- Reduce VRAM usage
- Accelerate image generation
- Support higher resolution outputs

While this optimization allows for higher resolution image generation, it comes with some trade-offs, such as less stylistic variation between generated images, and slightly reduced prompt interpretation accuracy.

If you need to remove `--xformers`:

:::info
Only Olares admin can adjust system parameters through the Control Hub app.
:::

1. Open Control Hub, and navigate to **Browse**.
2. Locate **sdwebui** under the admin's namespace.
3. Under **Deployments**, click **sdwebui**.
4. Click <i class="material-symbols-outlined">more_vert</i> in the top right corner, and click **Edit YAML**.
5. In the YAML editor, locate `--xformers` and remove it. The default YAML file should look similar to the following.

    ```yaml {5}
    env:
      - name: CLI_ARGS
        value: >-
          --allow-code --enable-insecure-extension-access --api
          --no-hashing --gradio-queue --xformers
     ```

6. Click **OK** to apply the system settings.

## Gallery

<table>
  <tr>
    <td><img src="/images/manual/use-cases/sd-example1.png" alt="Image 1" width="200" /></td>
    <td><img src="/images/manual/use-cases/sd-example2.png" alt="Image 2" width="200" /></td>
  </tr>
  <tr>
    <td><img src="/images/manual/use-cases/sd-example3.png" alt="Image 3" width="200" /></td>
    <td><img src="/images/manual/use-cases/sd-example4.png" alt="Image 4" width="200" /></td>
  </tr>
</table>