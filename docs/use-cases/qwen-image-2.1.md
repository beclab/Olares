---
outline: [2, 3]
description: Install Qwen Image 2.1 on Olares, connect a chat model for prompt enhancement, and generate, edit, extend, and download images locally.
head:
  - - meta
    - name: keywords
      content: Olares, Qwen Image 2.1, AI image generation, reference images, image editing, inpainting, transparent PNG
app_version: "0.1.19"
doc_version: "1.0"
doc_updated: "2026-11-20"
---

# Generate and edit images with Qwen Image 2.1

Qwen Image 2.1 is an image creation app maintained by Olares. It runs the Qwen Image 2.1 model locally to generate pictures from text and edit existing images. You can combine reference images, change a selected area, extend a composition, and remove backgrounds through a visual interface.

AI prompt enhancement uses a chat model connected through Router to help refine your descriptions.

## Learning objectives

- Install the app and download its image models.
- Configure a chat model for AI prompt enhancement.
- Create a product image, then refine its colors, composition, and background.
- Find, reuse, and download your results.

## Prerequisites

- Olares 1.12.8 or later and an administrator account for installation and model configuration.
- An NVIDIA GPU supported by the app. The package requests 22 GiB of GPU memory and 32 GiB of system memory.
- Space for the app, generated images, and model weights. The default INT8 weights require approximately 25 GB of downloads.

## Install the app and prepare the image models

1. Open Market and search for "Qwen Image 2.1".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open Qwen Image 2.1 from Launchpad. On first start, the app downloads the INT8 image models by default. Wait for the download screen to finish and the workspace to open.
4. In **Settings** > **Models**, check that the INT8 weights show **Installed**. Click **Open the Model Console** to view file download progress and checksums. Once the models have downloaded and the generation service is ready, you can submit your first task.

The examples below use INT8. To try NVFP4, download its weights in **Settings** > **Models**, then select that precision when generating an image.

## Configure a chat model for AI enhancement

Use your own prompts directly to generate and edit pictures, or connect a chat model to refine your descriptions with **AI enhance**.

This guide uses Qwen3.8-27B (llama.cpp), a local chat model that can also receive reference images.

1. Make the chat model available on Olares. If it is not installed, search for "Qwen3.8-27B (llama.cpp)" in Market, click **Get**, then **Install**, and complete the installation. Open its Model Console and wait for **Model** to show **Ready** and **Engine** to show **Running**.

   For another model, install its packaged app or follow [Run local LLMs with Engine Base apps](llm-base-apps.md).

2. Open Router from Launchpad. On **Default models**, select the chat model for `default-chat`.
3. Return to Qwen Image 2.1 and open **Settings**. Under **AI optimize**, click **Refresh** to reload the chat models from Router.
4. In **Chat model**, keep **Router default (default-chat)** to follow Router's default, or select a specific model from the list. Confirm that the status shows **Ready**.

   Models marked **sees images** can use your reference images when rewriting the prompt. With a text-only model, enhancement uses the text alone. A specific model selection applies independently of Router's default.

The app connects to Router automatically.

## Create and refine images

Use the coffee mug examples below to create a product photograph and adapt it for different uses. Substitute your own product and desired changes as needed.

The **Apps** page and the menu above the input panel provide **Generate image**, **Inpaint**, **Extend**, and **Remove background**. You can also open a result and choose an action directly below it.

### Generate an image from text

1. Select **Generate image**. For a new image from text alone, remove any existing reference images with their remove buttons.
2. Enter a description in **Prompt**. For example:

   ```text
   A clean product photograph of a matte teal ceramic coffee mug on a warm ivory tabletop, soft morning window light from the left, a few coffee beans beside the mug, minimalist composition, no text, no logo.
   ```

3. Optionally, use the arrow beside **AI enhance** to choose **Enhance and translate to English**, **Enhance, keep the language**, or **Translate to English only**. Click **AI enhance** and review the rewritten prompt. Edit any details you do not want before generating.
4. Set the aspect ratio to **1:1** and the image count to **1**. Under **Advanced settings**, use **Resolution** of **1K**, **Precision** of **INT8**, and a random seed.

   | Setting | How to choose |
   | --- | --- |
   | **Aspect** | Use 1:1 for a square product image, 16:9 for a banner, or 9:16 for a vertical composition. |
   | **Images** | Generate 1–4 candidates. Start with one while refining the description. |
   | **Resolution** | Use 1K for initial drafts. Select 2K when generating a higher-resolution version. |
   | **Speed** | Available on supported editing paths. Text-to-image and reference composition have one speed. |
   | **Seed** | Leave it random for new variations. Lock a seed when you want to compare prompt changes with the same starting seed. |

5. Click **Generate**. The task appears in **Image** with its progress. Use **View queue** to inspect active and waiting tasks.
6. Open the completed image to inspect it. **Details** shows its prompt, references, and generation parameters.

### Edit an existing image

Use a text instruction to change an image while keeping its composition. For example, make a second color option for the same mug.

1. Open the generated mug image and click **Edit**. It becomes the first reference image, and the aspect setting changes to **Image 1**.
2. Describe the change and what should remain unchanged:

   ```text
   Change only the mug color to terracotta orange. Keep its shape, handle, position, the coffee beans, tabletop, lighting and framing unchanged.
   ```

3. Keep **Aspect** set to **Image 1**, then click **Generate**.
4. Open the result and enable **Compare with the original**. Move the divider to inspect the edited and original areas.

![Compare the original teal mug with the edited orange mug](/images/manual/use-cases/qwen-image-2.1-edit-compare.jpg#bordered)

To start with your own photo, select **Generate image**, use **Add images** to upload it or choose it from the library, and select **Image 1** for the aspect. The first reference is the image being edited.

### Create a new composition from references

Reference generation combines elements from images into a new composition. Add references in the input panel and choose an output ratio for the new image.

1. Select **Generate image** and add your reference images. You can add up to 10, including images already in **Library**.
2. Open a reference thumbnail to set its **Image role**, such as **Product**, **Person**, **Style**, or **Scene**. For a mug photograph, choose **Product** to indicate which design to preserve.
3. In **Prompt**, type `@` and select an image from the list. Refer to each image where you explain how it should be used. For example, identify the mug image and a café interior image, then ask for the mug on a table in that café.

   ```text
   Place the mug from [product reference] on a table in [café reference]. Keep the mug's color, shape and handle. Use warm morning light and leave empty space on the right for a headline. Do not add text.
   ```

   Replace the bracketed descriptions with the image references inserted through `@`.

4. Select an output ratio, such as **16:9** for a wide composition.
5. Choose the resolution and image count, then click **Generate**. Check the product's shape and details in the result before using it.

You can drag reference thumbnails to change their order and remove a reference directly from the tray. When enhancing a prompt that contains references, check that the intended images are still referenced in the rewritten text.

### Change a selected area

Use **Inpaint** for a change limited to part of an image, such as removing the coffee beans beside the mug.

1. Open the orange mug result and click **Inpaint**.
2. Click **Paint mask**. Paint over the coffee beans and their immediate shadows. Adjust **Size**, use **Eraser** to refine the selection, or use **Undo** and **Clear** to correct it.
3. Click **Done**. The input panel now offers **Edit mask** and **Remove mask**.
4. Under **What should change?**, enter:

   ```text
   Remove the coffee beans and fill their area with the same warm ivory tabletop texture. Keep the mug and the rest of the photograph unchanged.
   ```

5. Click **Apply edit**. When the result is ready, compare it with the original and inspect the selection's edges.

![Paint over the coffee beans before applying the edit](/images/manual/use-cases/qwen-image-2.1-inpaint-mask.jpg#bordered)

The model uses the whole image as context, while the final result retains the original outside the mask. Give the selection enough room for the desired change and its transition into the surroundings.

### Extend the composition

Use **Extend** to make room around an existing image, such as adding space for a headline beside the mug.

1. Open the image and click **Extend**.
2. In the canvas editor, choose **16:9** under **Shape**, or use **Same shape** and an **Amount** preset to expand in the original proportions.
3. Drag the original image within the canvas to control where new content is added. For space on the right, move the original toward the left. Check the preview and displayed dimensions.
4. Click **Extend** in the editor to submit the task.
5. Open the result and inspect the join between the original and the new area. Plain backgrounds can reveal seams or changes in texture, so adjust the expansion and try again when needed.

When opening **Extend** from **Apps**, first add a **Source image**. **Adjust position and size** opens the same canvas editor.

### Remove a background

1. Open the image and click **Remove background**.
2. Confirm the intended picture appears under **Source image**, then click **Remove background** in the input panel.
3. Open the completed result. The checkerboard indicates transparent areas. Inspect fine edges and the inside of the mug handle.
4. Click **Download** to save the PNG with transparency.

![The mug and coffee beans on a transparent background](/images/manual/use-cases/qwen-image-2.1-remove-background.jpg#bordered)

For a cutout of the mug alone, first use **Inpaint** to remove the coffee beans, then remove the background.

### Continue from a result

Choose an action according to what you want to carry forward:

| Action | Use |
| --- | --- |
| **Reuse settings** | Return the prompt and generation settings to the input panel so you can change them before submitting another task. |
| **Generate variations** | Generate alternatives from the same prompt and settings with new random seeds. |
| **Use as reference** | Add the result as an input to another composition. |
| **Edit** | Make a described change to this image. |

For uploaded photos, use **Edit** to make changes or **Use as reference** to include them in a new composition.

## Find and download your work

**Image** groups results by generation task. Use **Search prompts** to find a description or **Filters** to narrow results by type and date. **Layout** changes how results are displayed.

**Library** separates **Generations**, **Uploads**, and **Favorites**. Click an image's heart icon to keep it in **Favorites**, or use **Search names** to find an asset.

To download one image, open it and click **Download**. To download several, select their checkboxes and click **Download** in the selection toolbar. The app packages the selected images into a ZIP file. The downloaded images retain their output dimensions and format.

## Common issues

### Why can't I generate images after opening the app?

Open **Settings** > **Models** and the Model Console to check download and startup progress. Once the weights show **Installed** and the generation service is ready, submit your task.

### How do I resolve "No chat model in Router"?

Confirm that the chat model is available in Router, then click **Refresh** under **AI optimize** in the app's **Settings**. While setting up enhancement, you can enter your own prompts to generate images.

### What should I do if prompt enhancement reports a lost image reference?

The app keeps the original prompt. Continue with that prompt, or simplify it and try enhancement again. Check that the rewritten prompt references the intended images before generating.

### How do I preserve the composition when editing an image?

Place the image you want to edit first in the reference tray and set **Aspect** to **Image 1**. To change a specific area, use **Inpaint** and paint a mask over that area.

### What should I do if generation runs out of GPU memory?

Read the failed task's error message, then try generating at **1K** with fewer reference images.

## Learn more

- [Run local LLMs with Engine Base apps](llm-base-apps.md): Prepare another local chat model for prompt enhancement.
- [Olares Router](olares-router.md): Manage the models available to Olares apps.
