---
outline: [2, 3]
description: Administrators' guide for managing ComfyUI on Olares with ComfyUI Launcher—covering a quick preflight network check, service control, network configuration, model storage and installation, plugin management, Python dependencies, and maintenance.
---

# Manage ComfyUI using ComfyUI Launcher

ComfyUI Launcher is the core management tool of ComfyUI for **administrator users**. You can use it to control the running status of the ComfyUI service within the cluster, while easily managing models, plugins, runtime environment, and network configurations.

This guide walks you through using ComfyUI Launcher for service management and routine maintenance.

## Learning objectives

By the end of this tutorial, you will learn how to:

- Start and stop the ComfyUI service for all users in the cluster.
- Verify and adjust network settings so GitHub, PyPI, and Hugging Face are accessible.
- Install essential models and additional models, and understand where models are stored in **Files**.
- Manage plugins registered in ComfyUI Manager as well as from GitHub.
- Manage Python dependencies by installing, updating, removing packages, and analyzing missing dependencies for plugins.
- Troubleshoot common issues using logs and safely reset ComfyUI to its initial state when necessary.

## Start and stop the service

As the administrator, you must start the ComfyUI service before you or other members can access it via the client interface.

- **Start ComfyUI service**

    Go to **Home** and click the **START** button in the upper-right corner.
    ![Start ComfyUI service](/images/manual/use-cases/comfyui-start-service.png#bordered)    
    
    ::: tip Notes on first run
    - Initial startup of ComfyUI Launcher typically takes 10-20 seconds for environment initialization.
    - If the system prompts that essential models are missing, you can click **START ANYWAY**. However, workflows may fail without these base models. We recommend downloading the essential model package before starting the service.
    :::

- **Stop ComfyUI service**

    Go to **Home** and click the **STOP** button when ComfyUI is not in use. This releases VRAM and memory resources for other applications.
    ![Stop ComfyUI service](/images/manual/use-cases/comfyui-stop-service.png#bordered)   

## Configure network

ComfyUI relies heavily on external resources such as GitHub (plugins), PyPI (Python packages), and Hugging Face (models). Before installing components, verify the connection status in **Network Config**.

1. Go to **Network Config**.
2. If any service shows `Inaccessible`, select an alternative URL for that service from the provided list.
3. Click **SAVE & CHECK** to re-test connectivity.
    ![Configure network](/images/manual/use-cases/comfyui-network-config.png#bordered)   
4. Repeat until the status of each service becomes `Accessible`.

![Network recheck](/images/manual/use-cases/comfyui-network-accessible.png#bordered){width=300}   

## Manage models

ComfyUI Launcher provides flexible ways to manage models: install the essential package with one click, download from the built-in library or a Hugging Face link, upload directly to **Files**, and delete models when needed.

### Install essential models

Essential models are basic resources required for ComfyUI to run, including SD large models, VAE, preview decoders, and auxiliary tools models. We recommend installing them on first run.

1. Open the package page in either of the following ways:
    - In the **Missing essential models** prompt window that appears on first start, click **INSTALL MODELS**.
    - Go to **Home**, scroll to the **Package installation** section, find **Stable Diffusion base package**, and click **VIEW**.
    ![Install basic package](/images/manual/use-cases/comfyui-base-package.png#bordered)

2. On the Package Details page, click **GET ALL** to start the automatic installation. You can track progress via the status bar.
    
    ![Install essential models](/images/manual/use-cases/comfyui-install-essential.png#bordered)

### Install additional models

In addition to the essential models, ComfyUI Launcher supports installing additional models from the built-in library, via a Hugging Face link, or by uploading files manually.

**Download from built-in library**

Follow these steps to download a model from the built-in Hugging Face library:

1. Go to **Model management**.
2. Scroll down to the **Available models** section, and find the desired model by category or keyword.
3. Click the <i class="material-symbols-outlined">download</i> button to install the model.

    ![Library download](/images/manual/use-cases/comfyui-model-built-in.png#bordered)

**Download via link**

If you can't find a specific model in the built-in library, you can install it via the model URL on Hugging Face:

1. Go to **Model management** > **Custom Download**.
2. Fill in the model URL and select the destination path based on the model type.
3. Click **DOWNLOAD MODEL**.

    ![Custom download](/images/manual/use-cases/comfyui-model-link.png#bordered)

**Upload external models**

If you can't find the desired model on Hugging Face, you can upload external models via **Files**.

1. Open **Files** from the launchpad.
2. Go to **External** > **olares** > **ai** > **model**.
3. Find the folder that your downloaded model belongs to and upload the file directly inside the target folder.

:::tip Folder structure guide
Ensure files are placed in the correct subfolders so ComfyUI recognizes them:
```plain
ai/model
  ├─ checkpoints/       → Base models (SD 1.5 / SDXL / Flux)  [.safetensors/.ckpt]
  ├─ vae/               → VAE                                 [.vae/.safetensors]
  ├─ loras/             → LoRA                                [.safetensors]
  ├─ controlnet/        → ControlNet models
  ├─ clip/              → CLIP text encoders
  ├─ ipadapter/         → IP-Adapter models
  ├─ upscalers/         → (Real)ESRGAN and other upscalers
  ├─ unet/              → UNet weights (e.g., SDXL/Flux)
  ├─ style_models/      → Style/special models (plugin-dependent)
  └─ others/            → Misc. or plugin-specific weights
  ```
:::

### Delete models

To delete a model:

1. Go to **Model management** > **Model library**.
2. Under the **Installed models** section, find the model you want to delete, and click the <i class="material-symbols-outlined">delete</i> button on the right to delete it.
![Delete a model](/images/manual/use-cases/comfyui-delete-model.png#bordered)

## Manage plugins

ComfyUI Launcher provides flexible ways to manage plugins in **Plugin management**.

![Plugin status](/images/manual/use-cases/comfyui-plugin-status.png#bordered)

### Manage available plugins

To manage available plugins registered in ComfyUI Manager:

1. Go to **Plugin management** > **Plugin library**.
2. Under **Available Plugins**, find the target plugin. 

   For any plugin, you can:
   - Click the <i class="material-symbols-outlined">visibility</i> button to view plugin details.
   - Visit the GitHub repository if it is available.

   Depending on the plugin status, additional actions are available:

    - **`Not installed`**:  
      - Switch version if multiple versions are provided.  
      - Click <i class="material-symbols-outlined">download</i> to install the plugin.

    - **`Installed`**:  
      - Switch version to upgrade or roll back. 
      - Click <i class="material-symbols-outlined">pause</i> to disable the plugin.  
      - Click <i class="material-symbols-outlined">delete</i> to uninstall the plugin.

    - **`Disabled`**:  
      - Switch version.  
      - Click <i class="material-symbols-outlined">play_circle</i> to enable the plugin.  
      - Click <i class="material-symbols-outlined">delete</i> to uninstall the plugin.

    - **`Banned`**:  
        Banned plugins cannot be installed or enabled.

At the top of the section, you can also:
   - Click **UPDATE ALL PLUGINS** to update all installed plugins.
   - Click **REFRESH** to refresh the plugin list.



### Install plugins from GitHub

To install plugins directly from GitHub repositories:

1. Go to **Plugin management** > **Custom Install**.
2. Enter the GitHub repository URL of the plugin.
3. (Optional) Specify the branch. If you are not sure, keep the default value.
4. Click **INSTALL PLUGIN**.
![Download plugin](/images/manual/use-cases/comfyui-plugin-install.png#bordered)

## Manage Python environment

ComfyUI's operation relies on a set of Python dependency libraries. You can manage these libraries easily on the **Python dependency management** page.

### Install dependency libraries

1. Go to **Python dependencies** > **INSTALL NEW PACKAGE**.
2. In the pop-up window, enter the library name and version number (optional), and then click **INSTALL**.
![Install new package](/images/manual/use-cases/comfyui-python-install.png#bordered)

### Manage installed dependency libraries

1. Go to **Python dependencies**.
2. Under the **Installed Python packages** tab, find the Python library you want to manage.
3. Click the <i class="material-symbols-outlined">arrow_upward</i> button on the right to update the library, or the <i class="material-symbols-outlined">delete</i> button to remove it.
![Manage installed packages](/images/manual/use-cases/comfyui-python-manage.png#bordered)

### Analyze dependency installation status

1. Go to **Python dependencies** > **Dependency analysis**. 
2. Click **ANALYZE NOW** to start analyzing.
3. From the plugins list on the left, find the problematic plugin highlighted in red, and click on it.
4. From **Dependency list**, find the missing library for the plugin, and click the **Install** button on the right. You can also click **FIX ALL** to automatically install all missing libraries.
![Analyze dependencies](/images/manual/use-cases/comfyui-dependency-analy.png#bordered)

## Troubleshoot and maintain ComfyUI

ComfyUI Launcher provides tools to help diagnose and maintain the ComfyUI service.

### Export ComfyUI logs

You can export logs to diagnose the current running status of ComfyUI:

1. Go to **Home** and click <i class="material-symbols-outlined">more_vert</i> in the upper-right corner, then click **View logs** to view the current running log.
![View Logs](/images/manual/use-cases/comfyui-view-logs.png#bordered)
2. Click the <i class="material-symbols-outlined">refresh</i> button to refresh the log, and the <i class="material-symbols-outlined">download</i> button to download the log.
![Export Logs](/images/manual/use-cases/comfyui-export-logs.png#bordered){width=450}

### Reset ComfyUI configuration

To reset ComfyUI to its initial state:

1. Go to **Home** and click <i class="material-symbols-outlined">more_vert</i> in the upper-right corner, then click **Wipe and restore**. 
2. In the prompt window, click **WIPE AND RESTORE**.
![Wipe and restore](/images/manual/use-cases/comfyui-wipe-and-restore.png#bordered){width=350}
3. Enter "CONFIRM", then click **CONFIRM**.
![Second confirmation](/images/manual/use-cases/comfyui-second-confirm.png#bordered){width=350}

After the restoration operation is complete, restart ComfyUI for the changes to take effect.

:::warning Exercise caution
Restoring ComfyUI is an irreversible operation. Please operate carefully.
:::