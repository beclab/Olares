---
outline: [2,3]
description: Troubleshoot slow or delayed Steam streaming on Olares by checking network connection, GPU allocation, and Steam resource configuration.
---
# Slow or delayed Steam streaming on Olares

Use this guide when Steam streaming on Olares feels slow or delayed during gameplay.

## Condition

- Steam streaming feels slow or delayed during gameplay.
- The issue may appear as high latency, stuttering, or delayed input response while streaming.

## Cause

Slow or delayed Steam streaming may be related to one or more of the following:

- The Olares device is using Wi-Fi instead of Ethernet, which may result in lower bandwidth or higher latency.
- The GPU is not set to **App exclusive** mode for Steam Headless, so it may need to compete with other apps for GPU resources.
- Steam Headless does not have enough CPU or memory resources allocated for the game being streamed.
- A specific game may perform better with a different Proton version.

## Solution

### Step 1: Use a wired connection

Check whether your Olares device is connected through **Ethernet** instead of Wi-Fi.

### Step 2: Set the GPU to App exclusive mode

Set the GPU to **App exclusive** mode, and make sure Steam Headless is selected as the exclusive app.

1. Go to **Settings** > **GPU**.
2. From the **GPU mode** dropdown, select **App exclusive**.
3. In the **Select exclusive app** section, check whether Steam Headless is selected.
   ![Check GPU mode](/images/manual/help/ts-steam-stream-gpu-mode.png#bordered)

If Steam Headless is already selected, continue to [Step 3](#step-3-check-runtime-cpu-and-memory-usage).

If another app is selected, switch it to Steam Headless as follows:

1. Stop that app.
   - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
   - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**, then click <i class="material-symbols-outlined">link_off</i> to unbind the current app.
3. Resume Steam Headless and make sure it is running.
   - Go to **Market** > **My Olares**, then select **Resume** from the dropdown list.
   - Or go to **Settings** > **Applications**, select Steam Headless, then click **Resume**.
4. Wait a few seconds, then click <i class="material-symbols-outlined">sync</i> to refresh the app list.
5. If Steam Headless is still not selected automatically, click **Bind app** to set it manually.
   ![Set the GPU to App exclusive mode](/images/manual/help/ts-steam-stream-exclu.png#bordered
)

### Step 3: Check runtime CPU and memory usage

Launch the game, then check Steam Headless runtime CPU and memory usage while it is running.

1. Open Control Hub from the Launchpad.
2. In the left sidebar, click **Browse**.
3. In the resource tree, expand your project and then **Deployments**.
4. Select the Steam Headless deployment.
5. In the upper-right corner of the details pane, click <i class="material-symbols-outlined">more_vert</i>.
6. Click **Details** and note the highest CPU and memory usage values while the game runs.
   - CPU usage is shown in `m` (millicores), where 1000 m = 1 CPU core.
   - Memory usage is shown in `Gi`.
   ![Check Steam resource usage](/images/manual/help/ts-steam-stream-details.png#bordered)

### Step 4: Compare usage with resource limits

After checking the runtime usage, open the Steam Headless deployment YAML file and compare the configured limits with the usage you noted in Step 3.
1. Close the **Details** page and return to the Steam Headless deployment page in Control Hub.
2. In the details pane on the right, click <i class="material-symbols-outlined">edit_square</i> to edit the YAML file.
3. Find `cpu` and `memory` under `limits`.
   
   For example:
   ```yaml
   limits:
      cpu: '18'
      memory: 64Gi
   ```
   ![Check Steam CPU and memory limits](/images/manual/help/ts-steam-stream-limit.png#bordered)
4. Compare these configured limits with the runtime usage you noted in Step 3.
5. If the runtime usage is consistently close to the current limit, increase the corresponding `cpu` or `memory` value based on your device capacity.

6. Click **Confirm** to save the changes, then test the game again.

### Optional: Try a different Proton version

If the issue only happens with a specific game after you complete the checks above, try changing the game's Proton version in Steam.

1. In Steam, open the target game in your **Library**, then click  <i class="material-symbols-outlined">settings</i>  > **Properties...**
2. Go to **Compatibility**.
3. Enable **Force the use of a specific Steam Play compatibility tool**.
4. Select a Proton version from the dropdown list.
   :::tip
   You can check [ProtonDB](https://www.protondb.com/) for compatibility reports and recommended Proton versions for specific games.
   :::
   ![Proton version](/images/manual/help/ts-steam-stream-lag-proton.png#bordered)
   
5. Launch the game again and check whether streaming performance improves.

## If the issue persists

If the issue persists after completing the steps above, create an issue in the [Olares GitHub repository](https://github.com/beclab/Olares/issues) and share the results of your checks, including:

- The current `cpu` and `memory` limits configured for Steam Headless.
- A screenshot showing Steam Headless CPU and memory usage while the game is running.
- The game title and Proton version in use.
- A short description of the issue.

Providing this information helps the team narrow down the cause more quickly.