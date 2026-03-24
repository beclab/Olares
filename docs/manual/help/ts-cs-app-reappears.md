---
outline: [2, 3]
description: Troubleshoot when a stopped app cannot be removed in App exclusive mode in Olares 1.12.5.
---

# Cannot remove a stopped app in App exclusive mode

Use this guide when an app still appears in **App exclusive** mode after you stop it and remove it.

## Condition

- In **App exclusive** mode, after you stop an app and remove it, it appears again in the **Select exclusive app** section.
- The app is still shown as **Stopped** in Settings, Market, and Launchpad.

## Cause

In Olares 1.12.5, when you switch the GPU mode to **App exclusive**, the system automatically selects an app as the exclusive app.

To remove that app, you must stop it first. However, if the app uses a client/server (C/S) architecture, stopping it may stop only the client side while the server side continues running in the background. As a result, the app may still occupy GPU resources and appear again in the **Select exclusive app** section.

## Solution

1. Open **Settings** > **GPU** and check which app is currently shown in the **Select exclusive app** section.

    ![App shown in App exclusive mode](/images/manual/help/ts-cs-app-reappears-stopped-app.png#bordered){width=90%}

2. Go to **Settings** > **Applications**, find the same app, and then click **Resume**.

    ![Resume the app](/images/manual/help/ts-cs-app-reappears-resume.png#bordered){width=90%}

3. Wait a few moments for the app to start, and then click **Stop**.

    ![Stop the app](/images/manual/help/ts-cs-app-reappears-stop.png#bordered){width=90%}

4. Go back to **Settings** > **GPU**, and then click <i class="material-symbols-outlined">link_off</i> to remove the app again.

5. Refresh the list and verify that the app no longer appears in the **Select exclusive app** section.

    ![Check App exclusive mode again](/images/manual/help/ts-cs-app-reappears-no-apps.png#bordered){width=90%}

6. You can now resume the app you want to use and select it as the exclusive app.