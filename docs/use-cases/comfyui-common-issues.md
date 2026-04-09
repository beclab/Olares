---
outline: [2, 3]
description: Common issues and solutions for ComfyUI on Olares, including startup problems, launcher log messages, missing models, workflow errors, and high CPU temperature on Olares One
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, troubleshooting, common issues, self-hosted
---
# Common issues

Use this page to identify and resolve common issues with ComfyUI on Olares.

## ComfyUI cannot start

ComfyUI does not start, keeps stopping, or behaves unexpectedly when you try to launch it.

This is often related to insufficient system resources, an unsuitable GPU mode, or plugin dependency conflicts.

To resolve this:

1. Check whether enough CPU, memory, and VRAM are available. Stop other resource-intensive apps if needed.
2. Check whether the current [GPU mode](/manual/olares/settings/single-gpu.md) is suitable for your workload.
3. If the issue still persists, follow the [Troubleshooting flow](/use-cases/comfyui-launcher#troubleshooting-flow).

## Launcher log shows unexpected errors

Log messages in ComfyUI Launcher do not always mean that ComfyUI is broken. Some `Error` messages may appear during startup, plugin scanning, or environment checks even when ComfyUI is still usable.

If ComfyUI starts successfully, many of these warnings may not require action. You only need to investigate the logs if ComfyUI fails to start, a workflow cannot run, or a plugin stops working.

If you need to escalate the issue, see [Collect information for support](/use-cases/comfyui-launcher.md#collect-information-for-support). 

## Models cannot be downloaded in ComfyUI Launcher

A workflow may require a model that needs a login, access token, or approval to download. ComfyUI Launcher cannot download these models directly.

To solve it, find the download link using one of the methods below, download the model manually, and [upload it](/use-cases/comfyui-launcher.md#upload-local-models) to the correct folder in Olares Files. 

### Method 1: Check the template notes or Model Links section

Some official templates include notes or a **Model Links** section that lists:

- the required model file
- the download URL
- the expected storage location

If available, you can use this information to copy the download URL or open the model page directly.

![Model links](/images/manual/use-cases/comfyui-model-links.png#bordered){width=90%}

### Method 2: Use a browser helper extension

If the template shows a missing-model dialog but does not expose the full URL clearly, you can use a browser helper extension.

For example, with [WAN Download URL Helper](https://github.com/carlric/wan-download-url-helper):

1. Open the missing-model dialog in ComfyUI.
2. Hover over a download icon.
3. Right-click the icon and choose **Show download URL**.
4. Copy the URL, then use it in your downloader or save it for manual download.

![ComfyUI download URL helper](/images/manual/use-cases/comfyui-download-url-helper.png#bordered){width=80%}

### Method 3: Inspect the page in browser developer tools

If the URL is not shown in the template notes or dialog, inspect the page in your browser developer tools and look for network requests triggered by the template or missing-model dialog.

![Inspect url](/images/manual/use-cases/comfyui-inspect-url.png#bordered){width=80%}

## Workflow fails during execution

ComfyUI starts successfully, but a workflow halts and reports an error.

This happens when the workflow encounters missing models, missing custom nodes, or Python dependency conflicts during the generation process.

To resolve this, find out the exact cause by checking the error report:

1. In the ComfyUI client, click **Active** to open the **Job Queue**.
2. Select the failed task from the list.
3. Click **Report error**, then click **Show Report** to expand the details.

   ![Workflow error report](/images/manual/use-cases/comfyui-workflow-error.png#bordered){width=80%}

Once you have the error details, decide your next step:

- If the error points to a missing model, see [Models cannot be downloaded in ComfyUI Launcher](#models-cannot-be-downloaded-in-comfyui-launcher).
- If the error points to a missing Python module or node, see [Analyze dependency installation status](/use-cases/comfyui-launcher#analyze-dependency-installation-status).
- If the cause is still unclear, follow the [Troubleshooting flow](/use-cases/comfyui-launcher#troubleshooting-flow).

## CPU temperature rises unusually high on Olares One

CPU temperature rises unusually high while running certain ComfyUI workloads on Olares One.

This issue typically occurs when running large workflows that require more memory (VRAM) than your graphics card has available. When this happens, the system may place an unusually heavy load on a single CPU core to swap data, driving the reported CPU temperature very high.

**Workaround**: Temporarily lower the CPU frequency.

1. Open the Control Hub app.
2. In the left sidebar, under **Terminal**, click **Olares**.

   ![Open terminal](/images/manual/use-cases/comfyui-ts-terminal.png#bordered){width=90%}

3. Run the following command to lower the maximum CPU frequency.

   For example, to set it to 5.0 GHz, run:
    ```bash
    echo 5000000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```
4. Run your task in ComfyUI.
5. After the workload completes, run the following command to restore the normal maximum CPU frequency of 5.4 GHz.
    ```bash
    echo 5400000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```

This is a temporary workaround. A long-term fix is still under investigation.