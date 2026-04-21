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

:::tip Need more help?
If you are encountering an issue that is not listed here, refer to [Troubleshooting flow](./comfyui-launcher#troubleshooting-flow).
:::

## ComfyUI cannot start

ComfyUI does not start, keeps stopping, or behaves unexpectedly when you try to launch it.

This is usually caused by incorrect GPU allocation or insufficient resources. To resolve this:

1. Go to **Settings** > **GPU** and check your GPU mode:
   - If you are using **Memory slicing**, make sure ComfyUI is bound to the GPU and has enough VRAM allocated.
   - If you are using **App exclusive**, make sure the exclusive app is set to ComfyUI.
2. Check your system resources. If your CPU or memory usage is maxed out, stop other resource-intensive apps.
3. Wait a moment, then try to launch ComfyUI again.

## Launcher log shows errors

`Error` messages in the Launcher logs do not necessarily indicate a system failure. During startup and plugin scanning, ComfyUI often logs non-fatal errors for missing optional dependencies or environment checks, even when running perfectly.

If ComfyUI starts successfully, many of these messages require action. Investigate logs only if ComfyUI fails to start, a workflow cannot run, or a plugin stops working.

## Models cannot be downloaded in ComfyUI Launcher

Some workflows require models that need a login, access token, or approval to download. ComfyUI Launcher cannot download these models directly.

To solve it, find the download link using one of the methods below, download the model manually, and [upload it](/use-cases/comfyui-launcher.md#upload-local-models) to the correct folder in Olares Files. 

### Method 1: Check the template notes or Model Links section

Some official templates include notes or a **Model Links** section that lists:

- The required model file
- The download URL
- The expected storage location

If available, copy the download URL or open the model page directly.

![Model links](/images/manual/use-cases/comfyui-model-links.png#bordered){width=90%}

### Method 2: Use a browser helper extension

If the template shows a missing-model dialog and does not expose the full URL, use a browser helper extension like [WAN Download URL Helper](https://github.com/carlric/wan-download-url-helper):

1. Open the missing-model dialog in ComfyUI.
2. Hover over a download icon.
3. Right-click the icon and choose **Show download URL**.
4. Copy the URL, then use it in your downloader or save it for manual download.

![ComfyUI download URL helper](/images/manual/use-cases/comfyui-download-url-helper.png#bordered){width=80%}

### Method 3: Inspect the page in browser developer tools

If the URL is not shown in the template notes or dialog, inspect the page in your browser developer tools and look for network requests triggered by the template or missing-model dialog.

![Inspect URL](/images/manual/use-cases/comfyui-inspect-url.png#bordered){width=80%}

## CPU temperature rises unusually high on Olares One

CPU temperature rises unusually high when running certain ComfyUI workloads on Olares One.

This issue typically occurs when the workflow requires more VRAM than your graphics card has. When this happens, the system will place heavy load on a single CPU core to swap data, driving the temperature high.

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