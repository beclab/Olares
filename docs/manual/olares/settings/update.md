---
description: Learn how to update your Olares system from LarePass or Settings, and how to manually update olaresd for optimal performance and security.
---
# Update Olares

Olares regularly releases new versions with feature improvements and usability enhancements. This guide explains how to check for and install system updates.

:::info Olares admin required
Only Olares admin can perform system updates. Updates will apply to all members within the same Olares cluster.
:::

:::tip
For details on Olares versioning practices and the current limitations regarding cross-minor version upgrades (e.g. `1.10.5` to `1.11.0`), see [Olares versioning](/developer/install/versioning.md).
:::

## Check and install updates
:::tip
Review the release notes before updating to learn about new features and important changes.
:::

You can update Olares from the LarePass app or from Olares Settings.

<Tabs>
<template #From-LarePass>

1. Open LarePass on your phone and go to **Settings**.
2. In the **My Olares** card, tap **System** to enter the **Olares management** page.
3. Tap **System update**.
4. Confirm the available version in the **New version** field, then tap **Upgrade**.
   ![Check for available version](/images/one/check-version1.png#bordered)
5. In the pop-up dialog, select how you want to upgrade:
   - **Download only**: Olares downloads the update package in the background while you continue using the system.
   - **Download and upgrade**: Olares downloads the update package and will install it after you confirm a restart.
   ![Choose upgrade method](/images/one/olares-upgrade1.png#bordered)
6. If you selected **Download only**, tap **Upgrade now** on the **System update** page to initiate the process. If you selected **Download and upgrade**, confirm the restart when prompted to begin installation.
7. Wait for the upgrade and restart to finish. A success message indicates the upgrade is completed.
   ![Upgrade success message](/images/one/upgrade-success.png#bordered)
8. Refresh your Olares desktop to sync the latest system changes.

</template>
<template #From-Olares-Settings>

1. Open **Settings**, and click **System** > **My Olares** > Current version.
2. Click **Upgrade now** when there is an available new version.

You'll see a confirmation message when update completed.

</template>
</Tabs>

## Update `olaresd` manually

`olaresd` is the core daemon process of the Olares system responsible for various key management functions. In some cases, after your update Olares via the Settings app, a manual update of `olaresd` may be required to resolve issues where certain services fail to operate correctly.

To confirm if this step is required, refer to the [Release notes](https://github.com/beclab/Olares/releases/).

To update `olaresd` manually:

1. Open Control Hub, and navigate to **Terminal** > **Olares**.
   ![Open terminal in Olares](/images/manual/olares/olares-terminal-in-control-hub.png#bordered)
2. Execute the following command in the terminal:
   ```bash
   curl -SsfL https://dc3p1870nn3cj.cloudfront.net/upgrade_1_11_6.sh | bash -
   ```
   where:
   - `1_11_6` means to upgrade `olaresd` and `olares-cli` to version `1.11.6`.