---
outline: [2, 3]
title: Manage shared AI models in Common
description: Use the Common directory in Olares to manage AI models shared across applications and users.
head:
  - - meta
    - name: keywords
      content: Olares, Files, Common directory, AI models, shared models, Ollama, ComfyUI
---

# Manage shared AI models with the Common directory

The Common directory stores AI models that applications can reuse, saving storage space and avoiding duplicate downloads. Access to these files through Files is controlled separately for each Olares member.

## Understand the Common directory

### How the Common directory works

Multiple AI applications may need the same model files. For example, vLLM and llama.cpp can use the same Hugging Face cache. Common provides a shared location so applications can reuse models instead of keeping separate copies.

Application access to shared models is separate from member access through Files. Granting a member access to a Common folder lets them browse or manage its contents in Files. It does not configure how applications load models.

![Common directory interface](/images/manual/olares/files-common.png#bordered)

### Default directory structure

The Common directory includes three default subdirectories. Each follows the official storage structure of its corresponding platform:

| Subdirectory | Function | Directory structure |
|:------|:----|:-----------|
| **huggingface** | Stores cached models downloaded<br> via the Hugging Face CLI. Applications<br> such as vLLM, transformers, and llama.cpp <br>read from this unified cache. | Mirrors the official cache structure exactly.<br><br>For details on file organization, see the [Hugging Face cache management guide](https://huggingface.co/docs/huggingface_hub/guides/manage-cache). |
| **comfyui** | Stores models shared between ComfyUI <br>and related applications, such as<br> Checkpoint, LoRA, and VAE. | Follows the standard ComfyUI `models` folder structure.<br><br>For details on file organization, see the [ComfyUI models documentation](https://docs.comfy.org/development/core-concepts/models).|
| **ollama** | Stores models pulled and managed by Ollama. | Uses Ollama's unique manifests and blobs storage mechanism.<br><br>For more information, see the [Ollama FAQ](https://docs.ollama.com/faq#where-are-models-stored). |

## Migrate models to the Common directory

Olares v1.12.6 includes the Common directory by default. If you are using an older version of Olares, follow the steps below to migrate your existing models.

1. Upgrade Olares to v1.12.6.
2. Uninstall your existing AI model-serving applications:

    - Pre-packaged model applications
    - Ollama (installed as a base engine with models pulled manually). Do not select **Also remove all local data**.
    - ComfyUI. Do not select **Also remove all local data**.

3. Install the new versions and restore your models:

    <Tabs>
    <template #Model-apps>

    Deploy the model you need using the new [Engine Base apps](/use-cases/llm-base-apps.md) (llama.cpp Engine Base, Ollama Engine Base, SGLang Engine Base, or vLLM Engine Base).

    The model files will be saved into the `huggingface` or `ollama` subdirectories accordingly.

    </template>

    <template #Ollama-models>

    For models pulled via Ollama:

    - To re-download models, deploy the models you need using the new Ollama [Engine Base app](/use-cases/llm-base-apps.md). The model files will be saved into the `ollama` subdirectory automatically.
    - To reuse existing models, manually copy them to the `ollama` subdirectory.

    </template>

    <template #ComfyUI>

    Install the new version of ComfyUI from the Market. The system will automatically move your existing model files to the `comfyui` subdirectory.

    :::tip Identify v2 and new versions
    On the app details page, check the **Compatibility** field in the **Information** panel:
    - The v2 shared app shows `Olares >=1.12.3-0, <1.12.6`.
    - The new shared app shows `Olares >=1.12.6-0`.
    :::

    </template>
    </Tabs>

## Find and manage shared models

In Olares 1.12.8, members have no access to Common folders by default. Ask the Super Admin or an admin with management access to grant **Can view** or **Can edit** on the folder you need. See [Manage access to Common and External](manage-common-external-access.md).

### Access the Common directory

1. Open the Files app from the Launchpad.
2. Select **Application** > **Common** in the left sidebar.
3. Open the `huggingface`, `ollama`, or `comfyui` folder you have access to.
4. Find your target model files.

### Manage model files

Adding or deleting model files requires **Can edit** or management access to the Common folder. With **Can view**, you can browse and download files but cannot change them.

Do not change the required folder structure. Applications might fail to load models if you change it.

- **Add models**: Drag and drop model files directly into the corresponding subdirectory.
:::warning Before deleting models
Deleting shared model files affects every application that uses them. Keep a backup or confirm that you can download the model again before deleting it.
:::

- **Delete models**: Open the Common folder you can edit, right-click the model file or folder inside it, and select **Delete**. Edit access lets members delete files and subfolders inside an authorized folder, but not the folder itself.
