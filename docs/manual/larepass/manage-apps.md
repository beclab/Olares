---
outline: [2, 3]
description: Open and manage Olares apps from your mobile device with LarePass. Install, update, clone, stop, resume, and uninstall apps.
---

# Manage Olares apps with LarePass

Use the **Apps** tab in LarePass to open, install, and manage Olares apps from your mobile device. The apps are installed and run on your Olares.

## Before you begin

Check the following requirements before accessing apps from your mobile device.

- Your Olares runs version 1.12.8 or later.
- Your LarePass mobile app includes the **Apps** tab.
- Your Olares account is signed in to LarePass, and your mobile device can connect to Olares. For remote access, see [LarePass VPN](private-network.md).

<!-- Release check: Confirm the minimum LarePass mobile version before publishing. -->

## Mobile support

All the Olares system apps support mobile use. Whether other installed apps can open on your mobile device depends on their mobile support. An app can install successfully but still not open on your mobile device.

An entrance that does not support mobile use appears grayed out with a monitor badge. Tapping it displays a message asking you to open it on desktop. You can still tap **Manage**, or long-press its icon, to access management actions allowed for your account.

## Access installed apps

The Apps tab shows apps installed on your Olares. By default, it shows running apps. Use the filter menu to view stopped apps or all apps.

1. Open LarePass and tap **Apps** in the bottom navigation bar.

   ![Apps tab in LarePass](/images/manual/larepass/apps-home.png#bordered){width=40%}

2. Find an app in the grid, or enter its name in **Search apps**.
3. Tap the app icon to open it.
4. To view stopped apps or all apps, tap <i class="material-symbols-outlined">more_horiz</i> beside the search box, and then select **Stopped** or **All**.

   ![Filter and refresh apps](/images/manual/larepass/apps-filter.png#bordered){width=40%}

5. To reload the app list, open the same menu and select **Refresh**.

### Go back

Use the app toolbar to return to the previous level after opening an app.

1. Inside an open app, tap its name at the top, such as **Market**, to show the toolbar.
2. Tap **Back** in the upper-left corner to return to the previous level.

## Install apps

Use Market in LarePass to find and install new apps on your Olares.

1. In **Apps**, tap **Market**.
2. Find the app you want to install, open its details page, and check its hardware and resource requirements.
3. Tap **Get**, then **Install**. If prompted, complete the required configuration.
4. Once installation is complete, return to **Apps** to find the app. If needed, tap <i class="material-symbols-outlined">more_horiz</i> beside the search box and select **Refresh**.

## Update apps

Market lists available updates for your installed apps. You can update apps individually or all at once.

1. In **Apps**, open **Market**.
2. Tap <i class="material-symbols-outlined">menu_open</i> in the upper-left corner and select **Updates**.
3. Update an individual app, or select **Update all** to update all available apps.

## Clone apps

Cloning creates another instance of an installed app with its own configuration. Only apps that support multiple instances can be cloned.

1. In **Apps**, open **Market**.
2. Tap <i class="material-symbols-outlined">menu_open</i> in the upper-left corner and select **My Olares**.
3. Find the app, tap the drop-down arrow beside its operation button, and select **Clone**.
4. Follow the prompts to name and configure the new instance, then confirm.
5. Once cloning is complete, return to **Apps** to find the new instance.

For cloning requirements and configuration details, see [Clone applications](../olares/market/clone-apps.md).

## Stop and resume apps

Stop apps you are not using to free up resources, and resume them when needed. Both actions are available from Apps and Market. Available actions depend on the app's state and your account permissions. System apps cannot be managed in the same way as installed apps.

<tabs>
<template #From-Apps>

- **To stop an app**: Long-press the app icon to open the **Manage [app name]** panel, then select **Stop app**.
- **To resume an app**: Tap <i class="material-symbols-outlined">more_horiz</i> beside the search box and select **Stopped**, then long-press the app icon and select **Resume app**.
</template>
<template #From-Market>

1. In **Apps**, open **Market** and go to **My Olares**.
2. Find the app and tap the drop-down arrow beside its operation button.
3. Select **Stop** or **Resume**.
</template>
</tabs>

For more information about stopping and resuming apps, see the [Market FAQ](../olares/market/market.md#faqs).

## Uninstall apps

Remove apps you no longer need from your Olares through Apps or Market.

:::warning Uninstalling apps
Uninstalling removes the app from Olares, not just its icon from LarePass. If you choose to delete app data, that data cannot be recovered through reinstallation. Back up any data you need before confirming. Uninstalling a shared service may affect other users. Check the scope shown in the confirmation dialog. For details about uninstall options, see [Uninstall applications](../olares/market/market.md#uninstall-applications).
:::

Uninstall is available only for apps that your account has permission to remove. System apps cannot be uninstalled.

<tabs>
<template #From-Apps>

1. In **Apps**, find the app. If it is stopped, tap <i class="material-symbols-outlined">more_horiz</i> beside the search box and select **All** or **Stopped** first.
2. Long-press its icon and select **Uninstall** in the **Manage [app name]** panel.
3. Review the uninstall options, including whether to delete app data or remove a shared service, then confirm.
</template>
<template #From-Market>

1. In **Apps**, open **Market** and go to **My Olares**.
2. Find the app, tap the drop-down arrow beside its operation button, and select **Uninstall**.
3. Review the uninstall options, including whether to delete app data or remove a shared service, then confirm.
</template>
</tabs>
