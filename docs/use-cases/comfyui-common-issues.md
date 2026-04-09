---
outline: deep
description: Resolve common ComfyUI Launcher issues on Olares, including startup problems, missing models outside of ComfyUI Launcher, workflow errors, and troubleshooting steps.
---

# Resolve common issues in ComfyUI Launcher

Use this page to identify and troubleshoot common issues with ComfyUI Launcher on Olares.

For routine setup and maintenance tasks, see [Manage ComfyUI using ComfyUI Launcher](./comfyui-launcher).

**Common issues**
- [ComfyUI cannot start](#comfyui-cannot-start)
- [Why does the Launcher log show errors?](#why-does-the-launcher-log-show-errors)
- [Download missing models outside of ComfyUI Launcher](#download-missing-models-outside-of-comfyui-launcher)
- [A workflow reports an error](#a-workflow-reports-an-error)
- [CPU temperature rises unusually high on Olares One](#cpu-temperature-rises-unusually-high-on-olares-one)

**General recovery flow**

- [Troubleshooting flow](#troubleshooting-flow)

## ComfyUI cannot start

If ComfyUI does not start, keeps stopping, or behaves unexpectedly when you try to launch it, start with the checks below.

1. Check whether enough CPU, memory, and VRAM are available. Stop other resource-intensive apps if needed.
2. Check whether the current [GPU mode](/manual/olares/settings/single-gpu.md) is suitable for the workload.
3. If the issue started after installing new plugins, continue with [Check dependency conflicts](#check-dependency-conflicts).
4. If ComfyUI still cannot start, continue with [Troubleshooting flow](#troubleshooting-flow).

## Why does the Launcher log show errors?

Log messages in ComfyUI Launcher do not always mean that ComfyUI is broken.

Some `Error` messages may appear during startup, plugin scanning, or environment checks even when ComfyUI is working normally.

You usually only need to investigate the logs if:

- ComfyUI does not start
- a workflow cannot run
- a plugin stops working after installation

If you need to escalate the issue, see [Collect information for support](#collect-information-for-support).

## Download missing models outside of ComfyUI Launcher

Some workflows require models that cannot be downloaded directly in ComfyUI Launcher.

This usually happens when a model requires login, a token, approval, manual download, or comes from a source not supported by ComfyUI Launcher.

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
    ![ComfyUI download URL helper](/images/manual/use-cases/comfyui-download-url-helper.png#bordered){width=90%}

4. Copy the URL, then use it in your downloader or save it for manual download.

### Method 3: Inspect the page in browser developer tools

If the URL is not shown in the template notes or dialog, inspect the page in your browser developer tools and look for network requests triggered by the template or missing-model dialog.

![Inspect url](/images/manual/use-cases/comfyui-inspect-url.png#bordered){width=90%}

After downloading the model manually, upload it to the correct folder in Olares Files. For detailed steps, see [Upload local models](/use-cases/comfyui-launcher.md#upload-local-models).

## A workflow reports an error

If ComfyUI starts successfully but a workflow fails during execution, find out why by checking the error report in the client.

1. In the ComfyUI client, click **Active** to open the **Job Queue**.
2. Select the failed task from the list.
3. Click **Report error**, then click **Show Report** to expand the details.

   ![Workflow error report](/images/manual/use-cases/comfyui-workflow-error.png#bordered){width=90%}

Once you have the error details, decide your next step:

- If the error points to a missing model, see [Download missing models outside of ComfyUI Launcher](#download-missing-models-outside-of-comfyui-launcher).
- If the error points to a missing Python module or node, see [Check dependency conflicts](#check-dependency-conflicts).
- If the cause is still unclear, continue with the [Troubleshooting flow](#troubleshooting-flow).

## CPU temperature rises unusually high on Olares One

If CPU temperature rises unusually high while running some ComfyUI workloads on Olares One, follow the workaround below.

This issue typically occurs when running large workflows that require more memory (VRAM) than your graphics card has available.

When this happens, the system may place unusually heavy load on a single CPU core and drive the reported CPU temperature very high.

To reduce the temperature, you can temporarily lower the CPU frequency.

1. Open the Control Hub app.
2. In the left sidebar, under the **Terminal** section, click **Olares**.
    ![Open terminal](/images/manual/use-cases/comfyui-ts-terminal.png#bordered){width=90%}

3. Run the following command to lower the maximum CPU frequency to a lower value.

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

## Troubleshooting flow

Use the following flow when the issue is not resolved by the problem-specific guidance above, or when you are not sure which problem category applies.

Start with the step that best matches your situation:

- If a workflow that previously worked starts failing after installing new plugins, check dependency conflicts first.
- If the issue is still not resolved, continue with reset or reinstall.
- If you need help from teammates or support, collect diagnostic information before escalating the issue.

### Check dependency conflicts

If problems start after installing new plugins, the issue may be caused by dependency conflicts.

Run a dependency analysis to identify and fix the problem. For detailed steps, see [Analyze dependency installation status](./comfyui-launcher#analyze-dependency-installation-status).

### Reset ComfyUI configuration

If the issue is still not resolved after the checks above, reset ComfyUI to its initial state.

:::warning Perform with caution
Resetting ComfyUI is irreversible. All plugins, custom configurations, and Python dependencies will be removed. Models stored in the shared `model` folder are not affected.
:::
:::tip Get diagnostic details
If you plan to contact support, export your ComfyUI logs before resetting, as this action will erase the current system state. See [Collect information for support](#collect-information-for-support).
:::

To reset ComfyUI:

1. In ComfyUI Launcher, go to **Home** and click <i class="material-symbols-outlined">more_vert</i> in the upper-right corner, then click **Wipe and restore**.
2. In the prompt window, click **WIPE AND RESTORE**.
    ![Wipe and restore](/images/manual/use-cases/comfyui-wipe-and-restore.png#bordered){width=50%}

3. Enter `CONFIRM`, then click **CONFIRM**.
    ![Second confirmation](/images/manual/use-cases/comfyui-second-confirm.png#bordered){width=50%}

After the reset is complete, restart ComfyUI for the changes to take effect.

### Reinstall ComfyUI completely

If the issue persists after the wipe and restore, uninstall and reinstall ComfyUI completely.

1. Go to **Market** > **My Olares**.
2. Click the dropdown arrow next to ComfyUI's operation button and select **Uninstall**.
3. In the **Uninstall** window, select **Also remove all local data**, then click **Confirm**.
4. Open Files from the Launchpad and go to `External/olares/ai`.
5. Delete the `comfyui` folder.
6. Reinstall ComfyUI from Market.
7. Once installation is complete, open ComfyUI Launcher and start the service.

### Collect information for support

If you cannot resolve the issue and need to escalate it to the support team, prepare the following diagnostic information.

#### Export ComfyUI logs

Logs contain the backend running status and error traces.

1. In ComfyUI Launcher, go to **Home** and click <i class="material-symbols-outlined">more_vert</i> in the upper-right corner, then click **View logs**.
   ![View Logs](/images/manual/use-cases/comfyui-view-logs1.png#bordered){width=90%}
2. Click the <i class="material-symbols-outlined">refresh</i> button to ensure you have the latest output.
3. Click the <i class="material-symbols-outlined">download</i> button to save the log file.
   ![Export Logs](/images/manual/use-cases/comfyui-export-logs.png#bordered){width=90%}

#### Get the workflow error report (optional)

If a specific workflow is failing, include a screenshot of the workflow error report.

For detailed steps, see [A workflow reports an error](#a-workflow-reports-an-error).