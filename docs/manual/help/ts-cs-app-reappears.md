---
outline: [2, 3]
description: Troubleshoot when a stopped app reappears in App exclusive mode after you remove it in Olares 1.12.5.
---

# A stopped app reappears in App exclusive mode

Use this guide when an app reappears in **App exclusive** mode after you remove it, even though the app is shown as **Stopped** in Settings, Market, or Launchpad.

## Condition

- Your device is running Olares 1.12.5.
- You switched the GPU from **Time-slicing** or **Memory slicing** to **App exclusive** mode.
- You removed the app that the system automatically selected, but after refreshing the page, it appears again in the **Select exclusive app** section.
- The app is shown as **Stopped** in Settings, Market, or Launchpad.

## Cause

This usually happens with apps that use a client/server (C/S) architecture.

In some cases, only the client side of the app has stopped, while the server side is still running in the background. Because of this, the system still treats the app as active and shows it again in the **Select exclusive app** section.

Before the GPU can be assigned to another app, the app must be fully stopped.

## Solution

1. Open **Settings** > **GPU** and check which app is currently shown in the **Select exclusive app** section.

    ![App shown in App exclusive mode](/images/manual/help/ts-cs-app-reappears-stopped-app.png#bordered){width=90%}

2. Go to **Settings** > **Applications**, find the same app, and then click **Resume**.

    ![Resume the app](/images/manual/help/ts-cs-app-reappears-resume.png#bordered){width=90%}

3. Wait a few moments for the app to start, and then click **Stop**.

    ![Stop the app](/images/manual/help/ts-cs-app-reappears-stop.png#bordered){width=90%}

4. Go back to **Settings** > **GPU**, and then click <i class="material-symbols-outlined">link_off</i> to remove the app again.

5. Refresh the page and verify that the app no longer appears in the **Select exclusive app** section.

    ![Check App exclusive mode again](/images/manual/help/ts-cs-app-reappears-no-apps.png#bordered){width=90%}

6. You can now resume the app you want to use and select it as the exclusive app.