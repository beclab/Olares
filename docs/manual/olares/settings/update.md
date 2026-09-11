---
outline: [2, 3]
description: Check the installed Olares OS version in Settings or LarePass, then download and install available updates from LarePass.
---
# Check and update Olares

You can check the installed Olares OS version in either Settings or LarePass. System upgrades are currently performed in the LarePass mobile app.

:::info Olares admin required
Only Olares admin can perform system updates. Updates will apply to all members within the same Olares cluster.
:::

:::tip
For details on Olares versioning practices and the current limitations regarding cross-minor version upgrades (e.g. `1.10.5` to `1.11.0`), see [Olares versioning](/developer/install/versioning.md).
:::

## Check the installed version

<Tabs>
<template #In-Olares-Settings>

1. Open **Settings**.
2. Click your avatar in the upper-left corner.
3. Find **Current version** to view the installed Olares OS version.

</template>
<template #In-LarePass>

1. Open LarePass on your phone and go to **Settings**.
2. In the **My Olares** card, tap **System** to enter the **Olares management** page.
3. Tap the device information area at the top. The **System version** field shows the installed Olares OS version.

</template>
</Tabs>

## Install an update in LarePass

:::tip
Review the [Olares release notes](https://github.com/beclab/Olares/releases/) before updating to learn about important changes and version-specific instructions.
:::

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

After the restart, [check the installed version](#check-the-installed-version) again and confirm it matches the version you selected.
