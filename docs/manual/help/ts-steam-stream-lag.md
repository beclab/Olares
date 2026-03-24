---
outline: [2,4]
description: Troubleshoot slow or delayed Steam streaming on Olares by checking nginx restarts, network connection, GPU mode, system mode, game compatibility, and Steam Headless resource limits.
---
# Slow or delayed Steam streaming

Use this guide when Steam streaming from Olares feels slow or delayed during gameplay.

## Condition

Steam streaming feels slow or delayed during gameplay, which may appear as high latency, stuttering, sudden disconnections, or delayed input response.

## Cause

Slow or delayed Steam streaming may be related to one or more of the following:

- **Nginx restarts**: An nginx restart interrupts the streaming session.
- **Network**: The Olares device is using Wi-Fi instead of Ethernet, causing lower bandwidth or higher latency.
- **GPU allocation**: The GPU resources are occupied by other apps.
- **GPU power**: The GPU is not running at full power.
- **Compatibility**: The game is not performing well with the current Proton version.
- **Resource limits**: Steam Headless is reaching its configured CPU or memory limits.

## Solution

Check the following items in order of priority.

### Avoid interruptions from nginx restarts

- If you are running Olares 1.12.4, update to 1.12.5.
- Avoid installing, uninstalling, or updating apps while streaming a game.

### Use a wired connection

If your Olares device is connected through Wi‑Fi, use a wired connection instead.

### Set the GPU to App exclusive mode

Grant Steam Headless full GPU access to maximize performance.

1. Go to **Settings** > **GPU** and select **App exclusive** from the **GPU mode** dropdown.
2. In the **Select exclusive app** section, check whether Steam Headless is selected. If it is, continue to the next check.

   ![Check GPU mode](/images/manual/help/ts-steam-stream-gpu-mode.png#bordered)

If another app is currently selected:

1. Go to **Settings** > **Applications**, select the app, and click **Stop**.
2. Return to **Settings** > **GPU** and click <i class="material-symbols-outlined">link_off</i>.
3. Go back to **Settings** > **Applications**, select Steam Headless, and click **Resume**.
4. Return to **Settings** > **GPU**, click <i class="material-symbols-outlined">sync</i> to refresh the app list, and if Steam Headless is still not selected, click **Bind app** to set it manually.

   ![Set the GPU to App exclusive mode](/images/manual/help/ts-steam-stream-exclu.png#bordered)

### Switch to Performance mode (Olares One only)

If you are using Olares One, try switching to **Performance mode**.

1. Go to **Settings** > **My Olares** > **My hardware**.
2. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to **Power mode** and select **Performance mode**.
   ![Set power mode to Performance mode](/images/manual/help/ts-steam-stream-lag-power-mode.png#bordered)

### Try a different Proton version

If the issue only happens with a specific game, try changing the game's Proton version in Steam.

:::info What is Proton?
Proton is Steam’s compatibility layer used to run Windows games on Linux-based systems. Since different games interact with hardware differently, the version of Proton you use can directly impact game compatibility and streaming performance.
:::

1. In Steam, open the target game in your **Library**, then click  <i class="material-symbols-outlined">settings</i>  > **Properties...**
2. Go to **Compatibility**.
3. Enable **Force the use of a specific Steam Play compatibility tool**.
4. Select a Proton version from the dropdown list.
   :::tip
   You can check [ProtonDB](https://www.protondb.com/) for compatibility reports and recommended Proton versions for specific games.
   :::
   ![Proton version](/images/manual/help/ts-steam-stream-lag-proton.png#bordered)
   
5. Launch the game again and check whether streaming performance improves.

### Check CPU and memory usage

#### Check runtime usage

Launch the game, then check how much CPU and memory Steam Headless is actively using.

1. Open Control Hub from the Launchpad.
2. In the left sidebar, click **Browse**.
3. In the resource tree, expand your project and then **Deployments**.
4. Select the Steam Headless deployment.
5. In the upper-right corner of the details pane, click <i class="material-symbols-outlined">more_vert</i>.
6. Click **Details** and note the highest CPU and memory usage values while the game runs.
   - CPU usage is shown in `m` (millicores), where 1000 m = 1 CPU core.
   - Memory usage is shown in `Gi`.
   
   ![Check Steam resource usage](/images/manual/help/ts-steam-stream-details.png#bordered)

#### Compare usage with the configured limits

Check if the runtime usage is hitting the configured limits.

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
4. Compare these limits with the actual usage you noted earlier.
5. If the usage is consistently close to the current limit, increase the `cpu` or `memory` value based on your device capacity.
6. Click **Confirm** to save the changes, then test the game again.

## If the issue persists

If the issue persists after completing the steps above, create an issue in the [Olares GitHub repository](https://github.com/beclab/Olares/issues) and share the results of your checks, including:

- The current `cpu` and `memory` limits configured for Steam Headless.
- A screenshot showing Steam Headless CPU and memory usage while the game is running.
- The game title and Proton version in use.
- A short description of the issue.
- The device you are streaming to, such as a laptop, handheld device, phone, or TV.

Providing this information helps the team narrow down the cause more quickly.