---
outline: [2, 3]
description: Run, fine-tune, compare, and export open models on Olares with Unsloth Studio, and work with the same environment in JupyterLab.
head:
  - - meta
    - name: keywords
      content: Olares, Unsloth, Unsloth Studio, fine-tuning, QLoRA, LoRA, GGUF, JupyterLab, local LLM
app_version: "0.0.5"
doc_version: "1.0"
doc_updated: "2026-10-14"
---

# Fine-tune and run open models with Unsloth

Unsloth Studio is an open-source, no-code web UI for training, running, and exporting open models in one local interface. You can download GGUF and safetensors models from Hugging Face, chat with them, fine-tune text models with QLoRA, LoRA, or full fine-tuning, and export the result to GGUF or safetensors.

On Olares, the app includes Unsloth Studio and JupyterLab. Both use the same GPU, model cache, and persistent storage, so models, training runs, and exports remain available after the app restarts.

This guide uses `Qwen/Qwen3.5-2B` and the `mlabonne/FineTome-100k` dataset as a small example.

## Learning objectives

In this guide, you will learn how to:
- Assign GPU memory to Unsloth.
- Load a model and chat with it in Unsloth Studio.
- Run a short QLoRA fine-tuning job and compare the result with the base model.
- Export the fine-tuned model to GGUF and save it to your Home folder.

## Prerequisites

- An x86-64 (amd64) Olares device with an NVIDIA GPU.
- Internet access to Hugging Face for downloading models and datasets.

## Install Unsloth

1. Open Market and search for "Unsloth".
   <!-- ![Unsloth](/images/manual/use-cases/unsloth.png#bordered) -->

2. Click **Get**, then **Install**. When prompted, set the following environment variables, and wait for installation to complete:
    - **UNSLOTH_STUDIO_PASSWORD:** Enter the Unsloth Studio password. It must contain at least 8 characters.
    - **JUPYTER_PASSWORD:** Enter the JupyterLab password. It must contain at least 8 characters.

:::warning Keep the Studio password
Unsloth Studio saves its password when the app starts for the first time. Changing **UNSLOTH_STUDIO_PASSWORD** afterward does not change the saved password. Store the password in Vault or another password manager.
:::

## Assign GPU memory

Unsloth runs on one NVIDIA GPU. The VRAM available to the app depends on the GPU mode:

| GPU mode | VRAM for Unsloth |
|:---|:---|
| Time slicing or Exclusive | The app can use the GPU's full VRAM. |
| Memory slicing | The app uses the quota you assign when you launch it. The default quota is the 4 GiB minimum, and you can assign up to 23 GiB. |

In Memory slicing mode, the default quota is only enough to start the app. Assign a larger quota before you train a model or load a large one. You cannot change the quota while the app is assigned to a GPU, so remove it first:

1. Go to **Settings** > **Accelerator**, then click **Manage** on the resource card.

2. Under **Assigned apps**, find Unsloth, click **Remove**, and confirm. This stops the app.

3. Go to **Settings** > **Applications**, find Unsloth, and click **Resume**.

4. Under **Select GPUs**, choose the GPU. In **Quota for current app**, enter the VRAM for your workload, then click **Launch**.

Training runs, exports, and downloaded models remain in storage while the app is stopped. For details, see [Manage GPU resources](../manual/olares/settings/gpu-resource.md#resume-an-app-and-assign-resources).

## Log in to Unsloth Studio

1. Open Unsloth from Launchpad.

2. Log in with the username `unsloth` and the password you set for **UNSLOTH_STUDIO_PASSWORD**.
   <!-- ![Unsloth Studio login](/images/manual/use-cases/unsloth-login.png#bordered) -->

JupyterLab is a separate entrance. You can find both entrances in **Settings** > **Applications** > **Unsloth** > **Entrances**.

## Use Unsloth Studio

### Chat with a model

1. On the chat page, open the **Select model** dropdown in the top-left corner. If you are on another page, click **New Chat** in the left sidebar first.

2. Search for a model. For example, enter `unsloth/Qwen3.5-2B-GGUF`.

3. For a GGUF repository, choose a quantization that fits your assigned VRAM. A 4-bit variant, such as `Q4_K_M` or `UD-Q4_K_XL`, is a common starting point.

4. Wait for the model to download and load. When loading finishes, enter a prompt to start chatting.
   <!-- ![Chat with a model in Unsloth Studio](/images/manual/use-cases/unsloth-chat.png#bordered) -->

Studio automatically sets inference parameters such as temperature, top-p, and top-k for supported models. You can change them and edit the system prompt in the chat settings.

:::tip Reuse downloaded models
Unsloth stores Hugging Face downloads in the shared Hugging Face cache on Olares. A model that is already in this cache is detected without downloading it again. Loading a large model still takes time because the weights must be read into memory.
:::

### Fine-tune a model

The Studio homepage has four areas: **Model**, **Dataset**, **Parameters**, and **Training/Config**.

1. Under **Model**, configure the model:
    - **Model Type:** Select **Text**.
    - **Training Method:** Select **QLoRA**. It uses the least VRAM of the three methods.
    - **Model:** Search for and select `Qwen/Qwen3.5-2B`. GGUF models are inference-only and cannot be trained.

2. Under **Dataset**, open the **HuggingFace Hub** tab, search for `mlabonne/FineTome-100k`, and select it. Keep the format set to **auto**.

3. For a quick trial run, set **Dataset slice** to a small range, such as rows `0` to `1000`.

4. Under **Parameters**, set **Max Steps** to `30`. Keep the default values for the other parameters.

5. Click **Start Training**. The button is available after both a model and a dataset are configured.

6. Wait for the model and dataset to download and load. The training view then shows the epoch, progress, loss, learning rate, and GPU usage. Training is complete when the progress bar reaches 100%.
   <!-- ![Unsloth Studio training progress](/images/manual/use-cases/unsloth-training.png#bordered) -->

A 30-step run checks that the workflow works on your device. It does not materially change model behavior. For real training, use a dataset prepared for your task, and increase the steps or epochs.

To stop a run early, click **Stop Training**. Select **Stop & Save** to keep a checkpoint, or **Cancel** to stop without saving one.

:::tip Prepare your own dataset
In the **Dataset** area, open the **Local** tab to upload PDF, DOCX, JSONL, JSON, CSV, or Parquet files. To turn documents into a training dataset, use Data Recipes. Create a recipe, add blocks, click **Validate**, run a preview, and then run a full dataset build.
:::

:::tip Gated models
For a gated model, such as some Llama or Gemma models, paste a Hugging Face access token into the **HuggingFace Token** field under **Model**. The token must have access to that model.
:::

### Compare the fine-tuned model

1. On the chat page, click **Compare** to compare two models side by side.

2. Select `Qwen/Qwen3.5-2B` as one model. For the other, open the **Fine-tuned** tab in the model selector and select the training run you just created.

3. Send a prompt. Studio sends the same prompt to both models.

Studio loads the two models one after another. Use prompts from your target task to check whether the responses improved.

### Export the model to GGUF

1. Open **Export** in Unsloth Studio.

2. Select the training run and its checkpoint.

3. Under **Choose a format**, select **GGUF / Llama.cpp**, then choose a quantization, such as `Q4_K_M`.

4. Keep the default save location, click **Export**, and wait for the export to finish. Conversion and quantization might take several minutes.

5. To check the exported file, open the **Fine-tuned** tab in the **Select model** dropdown and load the GGUF export.

Exports are saved in Unsloth's application data. To copy one to your Home folder, open JupyterLab, select **File** > **New** > **Terminal**, and run:

```bash
ls /opt/unsloth-studio/exports
cp -r /opt/unsloth-studio/exports/<export-folder> /workspace/host/
```

Replace `<export-folder>` with the folder name from the first command. The files then appear in the Home folder in Files.

## Use JupyterLab

1. Open the JupyterLab entrance and log in with the password you set for **JUPYTER_PASSWORD**.

2. Create notebooks in the default `/workspace` folder. Files in this folder persist in Unsloth's application data.

3. Use the `host` folder to read and write files in your Olares Home folder.

The notebook environment includes the Unsloth Python package and uses the same GPU and Hugging Face cache as Studio. Training jobs in Studio and JupyterLab share the GPU, so do not run both at the same time on a small VRAM quota.

## FAQs

### Where does Unsloth store its data?

| Data | Location |
|:---|:---|
| Hugging Face models and datasets | Shared Hugging Face cache on Olares |
| Training runs | `/opt/unsloth-studio/outputs` in application data |
| Exports | `/opt/unsloth-studio/exports` in application data |
| Notebooks | `/workspace` in application data |
| Your Home folder | `/workspace/host` in JupyterLab |

Large models use significant disk space. Delete models you no longer need from the bin icon in Studio's model search.

### Why does a model fail to load or training run out of memory?

The model or training job needs more VRAM than the app has. In Memory slicing mode, assign a larger quota as described in [Assign GPU memory](#assign-gpu-memory). You can also choose a smaller model, a lower quantization, or QLoRA instead of LoRA or full fine-tuning.

### Can I generate images or videos?

Image and video models in Studio are much larger than text models. The smallest image model setup is a download of about 7 GB and needs more than the default 4 GiB quota. The smallest video model setup is a download of about 61 GB and exceeds the app's resource limits, so video generation is not supported on Olares.

### I forgot the Unsloth Studio password. How do I reset it?

Changing **UNSLOTH_STUDIO_PASSWORD** does not reset a saved password. Open a JupyterLab terminal and run:

```bash
/opt/unsloth-studio/unsloth_studio/bin/unsloth studio reset-password --username unsloth
```

The command generates a new password and shows it in the terminal. The change takes effect immediately, without restarting the app. Log in with the new password, and store it in Vault.

## Learn more

- [Manage GPU resources](../manual/olares/settings/gpu-resource.md): Change GPU modes and assign VRAM quotas.
- [Unsloth Studio documentation](https://unsloth.ai/docs/new/studio): Upstream reference for training, Data Recipes, and export options.
