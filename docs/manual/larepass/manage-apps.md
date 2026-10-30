---
outline: [2, 3]
description: Use the Apps tab in LarePass to open Olares apps, install and update apps through Market, and stop, resume, or uninstall installed apps.
---

# Manage apps with LarePass

Use the **Apps** tab in the LarePass mobile app to access and manage applications on your Olares. Apps run on Olares; installing an app through LarePass installs it on Olares, not on your phone.

## Before you begin

- Your Olares runs version 1.12.8 or later.
- Your LarePass mobile app includes the **Apps** tab.
- Your Olares account is signed in to LarePass, and your phone can connect to Olares. For remote access, see [LarePass VPN](private-network.md).

<!-- Release check: Confirm the minimum LarePass mobile version before publishing. -->

## Find and open apps

1. Open LarePass and tap **Apps** in the bottom navigation bar.
2. Find an app in the grid, or enter its name in **Search apps**.
3. Tap the app icon to open it.

![Apps tab in LarePass](/images/manual/larepass/apps-home.png#bordered){width=40%}

The list shows running apps by default. Tap <i class="material-symbols-outlined">more_horiz</i> beside the search box to select **All**, **Running**, or **Stopped**. Select **Refresh** to reload the list. To find an app that has stopped, select **All** or **Stopped** before searching.

![Filter and refresh apps](/images/manual/larepass/apps-filter.png#bordered){width=40%}

The search box finds installed apps. To find new apps to install, open **Market**.

### Mobile support

System apps available from the Apps tab support mobile use. Other installed apps may or may not support mobile use.

An entrance that does not support mobile use appears grayed out with a monitor badge. Tapping it displays a message asking you to open it on desktop. You can still tap **Manage**, or long-press its icon, to access management actions allowed for your account.

### Go back

Inside an open app, tap its name at the top, such as **Market**, to show the toolbar. Tap **Back** in the upper-left corner to return to the previous level.

## Install apps

1. In **Apps**, tap **Market**.
2. Browse the **Discover** page, or open the upper-left menu and select **Search** to find an app.
3. Open the app details and check its hardware and resource requirements.
4. Tap **Get**, then **Install**. If prompted, complete the required configuration.
5. Wait for installation to finish. Return to **Apps** to find the app. If needed, select **Refresh** from the more menu.

Installing an app does not guarantee that it can open on your phone. Check its mobile support as described above.

For environment variables and other installation options, see [Manage applications in Market](../olares/market/market.md).

## Stop, resume, or uninstall apps

1. In **Apps**, long-press an app icon.
2. In the **Manage [app name]** panel at the bottom, select an action:

   | Action | Result |
   | --- | --- |
   | **Stop app** | Stop the app on Olares. |
   | **Resume app** | Resume a stopped app on Olares. |
   | **Uninstall** | Uninstall the app from Olares. |

   Available actions depend on the app's state and your account permissions. System apps cannot be managed in the same way as installed apps.

3. Review any confirmation dialog before continuing.

   :::warning Uninstalling apps
   Uninstalling removes the app from Olares, not just its icon from LarePass. If you choose to delete app data, that data cannot be recovered through reinstallation. Back up any data you need before confirming. Shared-service operations may affect other users; check the scope shown in the confirmation dialog.
   :::

4. Check the updated app state in **Apps**. After stopping an app, select **Stopped** or **All** to find it again.

## Update apps

1. In **Apps**, open **Market**.
2. Open the upper-left menu and select **Updates**.
3. Update an individual app, or select **Update all** to update all available apps.

For more Market operations, see [Manage applications in Market](../olares/market/market.md) and [Clone apps](../olares/market/clone-apps.md).
