---
outline: [2, 3]
description: Use the Common directory in Olares to manage AI models shared across applications and users.
---

# Manage shared AI models with the Common directory

The Common directory provides a system-level shared space to store AI models. Multiple applications and users access the exact same model files simultaneously to save storage space and eliminate duplicate downloads.

## Understand the Common directory

### How the Common directory works

Understand Olares's file storage mechanisms to see the benefits of the Common directory. 

The following table shows the storage strategies for different locations in the Files app:

| Node | File management strategy | Cross-user sharing | Cross-app sharing |
| :------| :------------| :---------| :------- |
| **Storage** | Isolated by user | ❌ | ✅ |
| **Data** | Isolated by application | ✅ | ❌ |
| **Common** | Shared across users and applications | ✅ | ✅ |

**Storage** and **Data** isolate files, which creates challenges in AI scenarios:
- **Difficulty sharing models between applications**: Multiple AI applications might require the same underlying models. For example, ComfyUI and Stable Diffusion share the same image model files, and vLLM and llama.cpp can read from the same Hugging Face cache directory. Isolated storage prevents direct reuse of these files across applications.
- **Wasted space from repeated downloads**: Storing massive model files in isolated user or application directories forces different applications or users to download the same model repeatedly. This wastes disk space and network bandwidth.

The Common directory resolves these issues. It provides a centralized space where different apps and users read the same model files, eliminating storage redundancy.

![Common directory interface](/images/manual/olares/files-common.png#bordered)

### Default directory structure

The Common directory includes three default subdirectories. Each follows the official storage structure of its corresponding platform:

| Subdirectory | Function | Directory structure |
|:------|:----|:-----------|
| **huggingface** | Stores cached models downloaded<br> via the Hugging Face CLI. Applications<br> such as vLLM, transformers, and llama.cpp <br>read from this unified cache. | Mirrors the official cache structure exactly.<br><br>For details on file organization, see the [Hugging Face cache management guide](https://huggingface.co/docs/huggingface_hub/guides/manage-cache). |
| **comfyui** | Stores models shared between ComfyUI <br>and related applications, such as<br> Checkpoint, LoRA, and VAE. | Follows the standard ComfyUI `models` folder structure.<br><br>For details on file organization, see the [ComfyUI models documentation](https://docs.comfy.org/development/core-concepts/models).|
| **ollama** | Stores models pulled and managed by Ollama. | Uses Ollama's unique manifests and blobs storage mechanism.<br><br>For more information, see the [Ollama FAQ](https://docs.ollama.com/faq#where-are-models-stored). |

## Upgrade to use the Common directory

Olares V1.12.6 includes the Common directory by default.

If you are using an older version of Olares, do not manually migrate existing model files. Manual migration involves complex steps and risks breaking application dependencies. Instead, follow these steps to upgrade:

1. Upgrade Olares to V1.12.6.
2. Uninstall your existing AI applications.
3. Install the V3 versions of these applications from the Market.
4. Re-download your required models. The V3 applications automatically store models in the new Common directory.

## Find and manage shared models

Access the Common directory in the Files app and manage shared models centrally.

### Access the Common directory

1. Open the Files app from the Launchpad.
2. Select **Application** > **Common** in the left sidebar.
3. Open the `huggingface`, `ollama`, or `comfyui` subdirectory.
4. Find your target model file.

### Manage files

You can add or delete model files in the Common directory, but you must maintain the official recommended structure for each subdirectory. Applications might fail to load models if you change the required hierarchy.

- **Add models**: Drag and drop model files directly into the corresponding subdirectory.
- **Delete models**: Right-click the model folder and select **Delete** to free up space. After deletion, related applications immediately lose access to the model and require a re-download to use it again.
