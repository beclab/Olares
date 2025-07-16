---
description: Administrators' guide on how to manage ComfyUI on Olares using ComfyUI Launcher, covering controling the service, managing models, plugins, and python environments, troubleshooting and maintenance.
---

# Manage ComfyUI using ComfyUI Launcher

ComfyUI Launcher is the core management tool of ComfyUI for **administrator users**. You can use it to control the running status of the ComfyUI service within the cluster, while easily managing models, plugins, runtime environment, and network configurations.

This document guides you on how to use ComfyUI Launcher for ComfyUI service management and routine maintenance.

## Start and stop the ComfyUI service

As the administrator, you must start the ComfyUI service before you and other members can access it using the client interface.

![Start/Stop ComfyUI service](/images/manual/use-cases/comfyui-start.png#bordered)

* **Start the ComfyUI Service**: Click the **START** button in the upper-right corner to start the ComfyUI service.
    
    ::: tip Notes on first run
    * Initial startup of ComfyUI Launcher typically takes 10-20 seconds for environment initialization.
    * If the system prompts that essential models are missing, you can click the **START ANYWAY** button to launch the service. However, note that workflows may not run correctly if basic models are missing. It is recommended to download the essential model package before starting the service.
    :::

* **Stop the ComfyUI Service**: If you are not using ComfyUI for the moment, click the **STOP** button to stop the service. This releases the VRAM and memory resources occupied by ComfyUI.

## Manage models

:::tip Note
Before installing models, ensure your host can access GitHub and HuggingFace. For details, refer to [Configure network](#configure-network).
:::

ComfyUI Launcher provides flexible ways to install models. You can install the essential models with one click, manually install from Hugging Face, or copy from external sources.

### Install essential models

Essential models are basic models required for ComfyUI to run, including SD large models, VAE, preview decoders, and auxiliary tools models. It is recommended to install the essential package when running ComfyUI for the first time.

1. Access the essential model package page in either of the following methods:
    - In the **Missing essential models** prompt window that appears when starting the service for the first time, click **INSTALL MODELS**.
    * In the **Package installation** section on the homepage, find **Essential model package**, and click **VIEW**.

2. On the essential model installation page, click **GET ALL RESOURCES** to start the automatic installation. You can view the installation status via the progress bar below.
    
    ![Install resource package](/images/manual/use-cases/comfyui-install-essential.png#bordered)

### Manually download models

In addition to the essential models, ComfyUI Launcher also supports downloading models from the HuggingFace model library, allowing you to easily get the models you need.

<Tabs>
<template #Download-from-builtin-library>
Follow these steps to download a model the built-in HuggingFace library:

1. Navigate to **Model management**.
2. Scroll down to the **Available models** section, and find the desired model by category or keyword.
3. Click the <i class="material-symbols-outlined">download</i> button to download the model.

    ![Library download](/images/manual/use-cases/comfyui-model-library-download.png#bordered)

</template>
<template #Custom-download-via-link>
If you can't find a specific model in the built-in library, you can still download it via the model URL on Hugging Face:

1. Navigate to **Model management** > **Custom Download**.
2. Fill in the model URL and select the destination path.
3. Click **DOWNLOAD MODEL**.

    ![Custom download](/images/manual/use-cases/comfyui-custom-model-download.png#bordered)
</template>
</Tabs>

:::tip Upload external models
If you can't find the desired model on Hugging Face, you can use **Files** to upload external models to the following directory:

 `External Devices/ai/model`
:::

### Delete a model

To delete a model:

1. Navigate to **Model management** > **Model library**.
2. Under the **Installed models** section, find the model you want to delete, and click the <i class="material-symbols-outlined">delete</i> button on the right to delete it permanently.

## Manage plugins

You can manage plugins through ComfyUI's built-in ComfyUI-Manager or use **Plugin management** in the Launcher.

![Manage plugins](/images/manual/use-cases/comfyui-manage-plugin.png#bordered)

### Manage available plugins

To manage available plugins registered in ComfyUI Manager:

1. Navigate to **Plugin management** > **Plugin Library**.
2. Under **Available plugins**, select the target plugin:
    * Click the <i class="material-symbols-outlined">pause</i> button to disable the currently running plugin, and click the <i class="material-symbols-outlined">play_circle</i> button to resume running.
    * Click the <i class="material-symbols-outlined">delete</i> button to delete the plugin.
    * Click the <i class="material-symbols-outlined">download</i> button to download the plugin.
    * Click the <i class="material-symbols-outlined">visibility</i> button to view plugin details.
    * Click the **UPDATE ALL PLUGINS** button to update all installed plugins.
    * Click the **REFRESH** button to refresh the plugin status.

### Download plugin from GitHub

To install plugins directly from GitHub repositories:

1. Navigate to **Plugin management** > **Custom Install**.
2. Enter the URL of the repository. 
3. Specify the branch (optional). Use the default (usually 'main' or 'master') if not specified.
4. Click **INSTALL PLUGIN**.

## Manage Python environment

ComfyUI's operation relies on a set of complex Python dependency libraries. You can manage these libraries easily on the **Python dependency management** page.

![Manage Python libraries](/images/manual/use-cases/comfyui-manage-python.png#bordered)

### Install new dependency libraries

To install a new dependency library:

1. Navigate to **Python dependencies**.
2. click **INSTALL NEW PACKAGE**.
3. In the pop-up window, enter the library name and version number (optional), and then click **INSTALL**.

### Manage installed dependency libraries

1. Under **Installed Python packages**, find the Python library you want to manage.
2. Click the <i class="material-symbols-outlined">arrow_upward</i> button on the right to update the library, or the <i class="material-symbols-outlined">delete</i> button to remove it.

### Analyze dependency installation status

1. Under the **Dependency analysis** tab, click **ANALYZE NOW** to start analyzing.
2. From the plugins list on the left, find the problematic plugin highlighted in red, and click on it.
3. From **Dependency list**, find the missing library for the plugin, and click the **Install** button on the right. You can also click **FIX ALL** to automatically install all missing libraries.
    ![Analyze dependencies](/images/manual/use-cases/comfyui-analyze-dependency.png#bordered)

## Troubleshooting and maintenance

ComfyUI Launcher provides tools to help diagnose and maintain the ComfyUI service.

### Configure network

Network connection issues can affect the download of models and plugins. Before using ComfyUI, it's recommended to check the connection status to GitHub, PyPI, and HuggingFace on the Launcher homepage.

![Check network status](/images/manual/use-cases/comfyui-check-network.png#bordered)

For example, if GitHub is inaccessible:

1. Navigate to the **Network configuration**.
2. Switch the provided URLs for GitHub.
3. Click the **CHECK** button on the right to refresh the connection status. 

Repeat steps 2 and 3 until the network becomes "Accessible".

![Switch network](/images/manual/use-cases/comfyui-change-network.png#bordered)

### Export ComfyUI logs

You can export logs to diagnose the current running status of ComfyUI:

![Export Logs](/images/manual/use-cases/comfyui-log.png#bordered)

1. In the upper-right corner of the Launcher homepage, click <i class="material-symbols-outlined">more_vert</i>, then click **View logs** to view the current running log.
2. Click the <i class="material-symbols-outlined">refresh</i> button to refresh the log, and the <i class="material-symbols-outlined">download</i> button to download the log.

### Reset ComfyUI configuration

To reset ComfyUI to its initial state:

![Reset ComfyUI](/images/manual/use-cases/comfyui-reset.png#bordered)

1. In the upper-right corner of the Launcher's homepage, click <i class="material-symbols-outlined">more_vert</i>, and select **Wipe and restore**. After a second confirmation, the Launcher will begin the wiping and restoration process.

2. After the restoration operation is complete, restart ComfyUI for the changes to take effect.

:::warning Exercise caution
Restoring ComfyUI is an irreversible operation. Please operate carefully.
:::

## Discover inspirations

Enter the **Discover** page from the left sidebar to view the latest and most trending ComfyUI models and workflows on [civit.ai](civit.ai).

::: tip Note
Please ensure your network can properly access [civit.ai](civit.ai).
:::