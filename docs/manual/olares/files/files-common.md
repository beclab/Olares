---
outline: [2, 3]
description: Use the Common directory in Olares to manage AI models shared across applications and users.
---

# Manage shared AI models with the Common directory

The Common directory is a system-level shared space for storing AI models, allowing multiple applications and users to access the same model files simultaneously.

## How the Common directory works

To understand the purpose of the Common directory, you first need to understand Olares's file storage mechanism. The following table shows the storage strategies for different locations in the Files app:

| Node | File management strategy | Cross-user sharing | Cross-app sharing |
| :------| :------------| :---------| :------- |
| **Storage** | Isolated by user | ❌ | ✅ |
| **Data** | Isolated by application | ✅ | ❌ |
| **Common** | Shared across users and applications | ✅ | ✅ |

Both **Storage** and **Data** isolate files to some extent, but this causes two problems in AI scenarios:
- **Difficulty sharing models between applications**: Multiple AI applications may require the same underlying models. For example, ComfyUI and Stable Diffusion can share the same image model files, and vLLM and llama.cpp can use the same Hugging Face cache directory. With isolated storage, these files cannot be reused directly across applications.
- **Wasted space from repeated downloads**: If model files are stored separately in isolated storage directories or application directories, different applications or users download the same model repeatedly, wasting time and space.

The Common directory solves these problems by providing a shared storage across users and applications, allowing different applications and users to access the same model files and avoid duplicate storage.

![Common directory interface](/images/manual/olares/files-common.png#bordered)

### Understand the directory structure

The Common directory includes the following three subdirectories by default, each following the official storage structure of the corresponding platform:

| Subdirectory | Function | Directory structure |
|:------|:----|:-----------|
| **huggingface** | Stores cached models downloaded<br> via the Hugging Face CLI. Applications<br> such as vLLM, transformers, and llama.cpp <br>read from this unified cache. | Mirrors the official cache structure exactly. For details on file organization, see the [Hugging Face cache management guide](https://huggingface.co/docs/huggingface_hub/guides/manage-cache). |
| **comfyui** | Stores models shared between ComfyUI <br>and related applications, such as<br> Checkpoint, LoRA, and VAE. | Follows the standard ComfyUI `models` folder structure. You can move or rename models here just like regular files. |
| **ollama** | Stores models pulled and managed by Ollama. | Uses Ollama's unique manifests and blobs storage mechanism. For more information, see the [Ollama FAQ](https://docs.olama.com/faq#where-are-models-stored). |

## Upgrade to use the Common directory

Olares V1.12.6 includes the Common directory by default.

If you are using an older version of Olares, do not manually migrate existing model files. Manual migration is complex and can break application dependencies. Follow these steps instead:

1. Upgrade Olares to V1.12.6.
2. Uninstall your existing AI applications.
3. Reinstall the V3 versions of these applications from the Market.
4. Download the models again. V3 applications automatically store models in the Common directory.

## Find and manage shared models

Access the Common directory in the Files app and manage the shared models centrally.

### Access the Common directory

1. Open the Files app from the Launchpad.
2. Select **Application** > **Common** in the left sidebar.
3. Enter the `huggingface`, `ollama`, or `comfyui` subdirectory.
4. Find the target model file.

### Manage files

You can manage files in the Common directory just like a regular folder.

:::warning Important notes
- Always maintain the folder structure recommended by the official documentation; otherwise, applications may fail to load models correctly.
- Stop any running AI applications before managing model files to prevent system errors.
:::

- **Add models**: Drag and drop model files directly into the corresponding subdirectory.
- **Delete models**: Right-click the model folder and select **Delete** to free up space. After deletion, related applications immediately lose access to the model and must re-download it to use it again.
- **Organize files**: Move or rename model files within the directory.
