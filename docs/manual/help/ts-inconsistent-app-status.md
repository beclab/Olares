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

Use Settings or Market to perform the corresponding action and sync the app status:

- If you stopped the app in Control Hub, select **Stop**.
- If you started the app in Control Hub, select **Resume**.

**From Settings**

1. Go to **Settings** > **Applications**.
2. Find the affected app.
3. Click **Stop** or **Resume**.

**From Market**

1. Go to **Market** > **My Olares**.
2. Find the affected app.
3. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button, and select **Stop** or **Resume**.

Wait a few moments. The app status in Launchpad, Settings, and Market should now match the app state set from Control Hub.