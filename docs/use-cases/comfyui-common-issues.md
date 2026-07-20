---
outline: [2, 3]
description: Troubleshoot ComfyUI on Olares, including startup, launcher logs, model downloads, workflow errors, and high CPU temperature on Olares One.
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, troubleshooting, common issues, self-hosted
---
# ComfyUI common issues

Use this page to identify and resolve common issues with ComfyUI on Olares.

:::tip Need more help?
If you are encountering an issue that is not listed here, refer to [Troubleshooting flow](./comfyui-launcher#troubleshooting-flow).
:::

## How to migrate to the new ComfyUI after upgrading to Olares 1.12.6

Use this section if you upgraded to Olares 1.12.6 and already had ComfyUI Shared installed. If you are installing ComfyUI for the first time on Olares 1.12.6 or later, install ComfyUI directly from Market.

Olares 1.12.6 updates the shared app architecture. The old ComfyUI Shared app can still run after the upgrade, but it cannot receive future updates. To keep ComfyUI up to date, uninstall the old app without deleting local data, then install the new ComfyUI app from Market.

:::warning
When uninstalling the old app, do not select **Also remove all local data**. Selecting this option may delete your models, plugins, workflows, and input/output files.
:::

### Migration steps

1. Open Market and go to **My Olares**.
2. Find ComfyUI Shared, click the dropdown arrow next to its operation button, and select **Uninstall**.
3. In the **Uninstall** window, leave **Also remove all local data** unselected, then click **Confirm**.
4. In Market, search for "ComfyUI" and click **Install**.
5. On the app details page, check **Information** > **Compatibility**. If it shows `Olares >=1.12.6-0`, you are installing the new version.
6. After installation, open ComfyUI and check that your models, plugins, workflows, and input/output files are available.

### What gets migrated

After the new ComfyUI is installed, data is migrated automatically as follows:

| Data type | Old location | New location |
|:---|:---|:---|
| ComfyUI core (plugins, workflows, etc.) | `External/<your_hostname>/ai/comfyui/` | `Data/comfyuisharev3/comfyui/` |
| Models | `External/<your_hostname>/ai/model/` | `Common/comfyui/model/` |
| Output files | `External/<your_hostname>/ai/output/comfyui/` | `Common/comfyui/output/` |
| Input files | `External/<your_hostname>/ai/comfyui/ComfyUI/input/` | `Common/comfyui/input/` |

:::warning
After migration, upload new models and input files to the new locations under `Common/comfyui/`. The new ComfyUI no longer uses `External/<your_hostname>/ai/` as its active file location.
:::

The migration runs each time ComfyUI restarts. If files are later added to the old locations, ComfyUI will move them to the new locations on the next restart and delete the originals from `External/<your_hostname>/ai/`. To avoid confusion, upload new files directly to the new locations.

### Setting up `extra_model_paths.yaml` after migration

After migration, ComfyUI automatically generates the `extra_model_paths.yaml` configuration file, which tells ComfyUI where to find models in the `Common` library.

The file is pre-configured with:
- **`base_path`**: Points to `Common/comfyui/model` (the shared model directory).

In most cases, you do not need to edit this file manually. However, if your models are stored in a non-default folder, you can manually edit and add the path at the following location:

```
Data/comfyuisharev3/comfyui/user/extra_model_paths.yaml
```

For details on editing this file, see [Manage files and directories](/use-cases/comfyui-launcher#about-extra_model_pathsyaml).

## ComfyUI cannot start

ComfyUI fails to start, stops unexpectedly, or behaves abnormally.

This is usually caused by insufficient resources or incorrect GPU allocation. To resolve this:

1. Check your system resources. If your CPU or memory usage is maxed out, stop other resource-intensive apps.
2. If system resources look fine, go to **Settings** > **Accelerator** and check your GPU mode:
   - If you are using **Memory slicing**, make sure ComfyUI is bound to the GPU and has enough VRAM allocated.
   - If you are using **Exclusive**, make sure the exclusive app is set to ComfyUI.
3. Wait a moment, then try to launch ComfyUI again.

## Launcher log shows errors

`Error` messages in the Launcher logs do not necessarily indicate a system failure. During startup and plugin scanning, ComfyUI often logs non-fatal errors for missing optional dependencies or environment checks, even when running normally.

If ComfyUI starts successfully, most of these messages do not require action. Investigate logs only if ComfyUI fails to start, a workflow cannot run, or a plugin stops working.

## Workflow cannot find models stored in `Common/comfyui/model/`

After migrating to ComfyUI v3 (Olares 1.12.6+), a workflow may report missing models even though the model files exist in `Common/comfyui/model/`. There are two main causes:

- **Cause 1**: The model's subdirectory is not registered in `extra_model_paths.yaml`, so ComfyUI does not scan this folder.
- **Cause 2**: The model is in ComfyUI's model directory, but a custom node's defined search path does not match.

In a example below, the `Common` directory contains the following two models. However, neither is recognized by the workflow node.
   ```
   /Common/comfyui/model/
   ├── detection/
   │   └── mediapipe_face_fp32.safetensors
   └── ultralytics/
       └── bbox/
           └── face_yolov8m.pt
   ```

### Step 1: Check if the model's subdirectory is recognized

1. In ComfyUI, open the **Model Library** sidebar and search for the model file name.
2. If the model does not appear in the list, its subdirectory has not been registered in `extra_model_paths.yaml`.
   In the example case,
   `ultralytics/bbox/face_yolov8m.pt` is detected, but `detection/mediapipe_face_fp32.safetensors` is not:
   ![Model detected](../public/images/manual/use-cases/comfyui-common-model-detected.png#bordered){width=49%}
   ![Model not detected](../public/images/manual/use-cases/comfyui-common-model-missing.png#bordered){width=49%}

### Step 2: Add the missing subdirectory to `extra_model_paths.yaml`

1. Open Files and navigate to `/Data/comfyuisharev3/comfyui/user/`.
2. Open `extra_model_paths.yaml`. The `base_path` at the top points to `/mnt/olares-shared-model`, which corresponds to `/Common/comfyui/model/` inside the container.

   In this example, the `detection` subdirectory is not yet registered in `extra_model_paths.yaml`, so you need to add `detection: detection` at the end of the file. Here, the key (`detection`) represents the name of the subdirectory under the ComfyUI model directory, which is also the model search path needed by the workflow node; the value (`detection`) refers to the corresponding subdirectory name under `Common/comfyui/model/`.

   ![Adding detection mapping](../public/images/manual/use-cases/comfyui-extra-model-paths-add-detection.png#bordered)

3. Save the file and restart ComfyUI from ComfyUI Launcher.
4. In the startup log, look for a line like `Adding extra search path detection /mnt/olares-shared-model/detection` to confirm the path was registered.

   ![Startup log confirming detection path](../public/images/manual/use-cases/comfyui-detection-path-added-log.png#bordered)

5. Refresh the ComfyUI page and check the Model Library again.

   ![Model now recognized after restart](../public/images/manual/use-cases/comfyui-model-recognized-after-restart.png#bordered)

### Step 3: Check for custom node path mismatches

If the model appears in the Model Library but a specific workflow node still cannot find it, the custom node may be looking in a different subdirectory.

1. Check the custom node's documentation or source code to find the exact search path it uses.
   
   For example, the `ImpactPack/UltralyticsDetectorProvider` node listens for models under `ultralytics_bbox` and `ultralytics_segm` — not the standard `ultralytics/` folder.

2. Add the required path mapping to `extra_model_paths.yaml`. For instance, to make bbox YOLO models available to the node, adding following path:

   ```yaml
   ultralytics_bbox: ultralytics/bbox
   ```

   ![Adding ultralytics_bbox mapping](../public/images/manual/use-cases/comfyui-ultralytics-bbox-mapping.png#bordered)

3. Restart ComfyUI and reload the workflow.

   ![face_yolov8m.pt now recognized](../public/images/manual/use-cases/comfyui-face-yolov8m-recognized.png#bordered)

## ComfyUI fails to start after upgrading to v1.0.37 or later

This issue may occur after upgrading to ComfyUI v1.0.37 or later.

After the upgrade, ComfyUI app may fail to start and show an error like:

```
main.py: error: unrecognized arguments: --normalvram
```

This means a custom launch argument from a previous version is still in use, but the new version no longer supports it.

To fix this:

1. Open **ComfyUI Launcher** and go to **Lab** from the left sidebar.
2. In the **Manually edit extra arguments** field, remove `--normalvram` manually and click **SAVE MANUAL ARGS**. Alternatively, click **RESTORE DEFAULT** to reset to the default launch arguments.
3. Verify that **Current full launch command** at the top no longer contains `--normalvram`.
4. Go back to **Home** in ComfyUI Launcher, then click **Start** to launch ComfyUI.

## Models cannot be downloaded directly to Olares

Some models require login, access approval, a token, or manual confirmation before they can be downloaded. These models cannot be downloaded directly to Olares through ComfyUI Launcher or Server Download.

To solve it, find the download link using one of the methods below. Then download the model manually and [upload it](/use-cases/comfyui-launcher.md#upload-local-models) to the correct folder in Olares Files.

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

When a workflow requires more VRAM than your graphics card has, the system places heavy load on a single CPU core to swap data, driving the temperature high.

The long-term fix is to reduce the VRAM footprint of your workflow (for example, lower resolution, use a smaller model, or enable model offloading). As a temporary workaround, limit the maximum CPU frequency while the workload is running.

### Olares OS 1.12.6 or later

Olares One ships with a CPU whose default maximum frequency is 5.4 GHz. Use the **Limit CPU frequency** switch to lower it to 5.0 GHz during the workload, then turn the switch off when the workload completes.

1. Open **Settings**.
2. Select your avatar in the top-left corner to open **My Olares**.
3. Under **My hardware**, turn on **Limit CPU frequency**.
4. Run your task in ComfyUI.
5. After the workload completes, turn off **Limit CPU frequency**.

For more details, see [Limit CPU frequency](/manual/olares/settings/my-olares#limit-cpu-frequency).

### Olares OS 1.12.5 or earlier

If your device is running Olares OS 1.12.5 or earlier, use terminal commands to lower the maximum CPU frequency during the workload, then restore it afterward.

1. Open the Control Hub app.
2. In the left sidebar, under **Terminal**, click **Olares**.

   ![Open terminal](/images/manual/use-cases/comfyui-ts-terminal.png#bordered){width=90%}

3. Run the following command to lower the maximum CPU frequency to 5.0 GHz:
    ```bash
    echo 5000000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```
   On other devices, adjust the target value based on your CPU's maximum frequency. Run `cat /sys/devices/system/cpu/cpufreq/policy0/cpuinfo_max_freq` to check it first.
4. Run your task in ComfyUI.
5. After the workload completes, run the following command to restore the default maximum CPU frequency of 5.4 GHz:
    ```bash
    echo 5400000 | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_max_freq
    ```
