---
outline: [2, 3]
description: Troubleshoot inconsistent app status between Control Hub and Launchpad, Settings, or Market.
---
# App status differs after starting or stopping an app in Control Hub

Use this guide when an app's status in Control Hub does not match what you see in Launchpad, Settings, or Market after you start or stop the app from Control Hub.

## Condition

- After you stop an app in Control Hub:
  - In **Settings** > **Applications**, the app still appears as **Running**.
  - In **Market** > **My Olares**, the app still shows the **Open** button.
  - In Launchpad, the app still appears available to open, without the orange paused indicator.
  - When you open the app from Launchpad or Market, the app fails to load.

- After you start an app in Control Hub:
  - In **Settings** > **Applications**, the app still appears as **Stopped**.
  - In **Market** > **My Olares**, the app still shows the **Stopped** button.
  - In Launchpad, the app still appears paused, with an orange indicator near the app name.
  - When you open the app from Launchpad, Olares asks you to resume the app first.

## Cause

Settings, Market, and Launchpad manage app status through app-service, while Control Hub operates directly on the underlying Kubernetes resources.

When you start or stop an app from Control Hub, the app workload is changed directly at the Kubernetes level. However, this operation does not update the app status maintained by app-service, which is shown in Settings, Market, and Launchpad. This can cause the app status in Settings, Market, and Launchpad to differ from its status in Control Hub.

## Solution

To sync the app status, repeat the same operation from Settings or Market.

The following steps use the case where you have stopped the app in Control Hub, but the app still appears as running in Launchpad, Settings, or Market.

1. Open **Settings** > **Applications**, or open **Market** > **My Olares**.

2. Find the affected app.

3. Stop the app again from Settings or Market:

   - In Settings, click **Stop**.
   - In Market, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button, and select **Stop**.

4. Wait a few moments. The app status in Launchpad, Settings, and Market should now match the status in Control Hub.

If you have started the app in Control Hub, but the app still appears as stopped in Launchpad, Settings, or Market, follow the same steps and select **Resume** instead.