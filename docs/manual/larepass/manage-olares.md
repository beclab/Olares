---
outline: [2, 3]
description: Learn how to use the LarePass app to manage your Olares device.
---
# Manage Olares with LarePass

The LarePass app allows you to easily manage your Olares device. You can upgrade Olares, monitor system status, manage network connections, perform remote controls, and check device information from your phone.

## Prerequisites

Before you begin, ensure the following:

- You have a valid Olares ID and an activated Olares device.
- Your Olares device is powered on and connected to a network.
- Your current account has administrator permissions.

## Access Olares management

The **Olares management** page is the central hub for using LarePass to manage your Olares device. To access **Olares management**:

1. Open LarePass app and go to **Settings**.
2. In the **My Olares** card, tap **System** to enter the **Olares management** page.

![Access Olares management](/images/manual/larepass/system.png#bordered)

On this page, you can:
- View device information
- Upgrade Olares
- Restart or shut down Olares remotely
- Modify network settings
- Restore Olares to factory settings
- Reset the SSH password <Badge type="tip" text="Olares One Only" />

![Olares management](/images/manual/larepass/olares-management1.png#bordered)

### View device information

Tap the Device information area at the top to view detailed information, including:
- Hardware details
- System version
- Current network connection status, including internal and external IP addresses

![View device information](/images/manual/larepass/view-device-information.png#bordered)

### Upgrade Olares
:::warning Olares admin required
Only Olares admin can perform system updates. Updates will apply to all members within the same Olares cluster.
:::

To install a system update:

1. On the **Olares management** page, tap **System update**.

2. On the **System update** page, confirm the available version in the **New version** field, then click **Upgrade**.
   ![Check for available version](/images/manual/larepass/check-version.png#bordered)

3. In the pop-up dialog, select how you want to upgrade:

   - **Download only**<br>
     Olares downloads the update package in the background while you continue using the system.

   - **Download and upgrade**<br>
     Olares downloads the update package and will install it after you confirm a restart.
   
   ![Choose upgrade method](/images/manual/larepass/olares-upgrade1.png#bordered)

4. If you selected **Download only**, click **Upgrade now** on the **System update** page to initiate the process.
5. If you selected **Download and upgrade**, confirm the restart when prompted to begin installation.
6. Wait for the upgrade and restart to finish. When Olares starts up again, you'll see a success message and can continue using it.

### Restart or shut down Olares remotely

In the upper-right corner of the **Olares management** page, tap the <i class="material-symbols-outlined">power_settings_new</i> icon to access remote control options:
- **Remote shutdown**: Your Olares will power off. If your phone is on the same network as Olares, its status will show `Powered off`. Remote operations are unavailable after shutdown, and you must turn the device back on manually.
- **Restart device**: Your Olares will restart. If your phone is on the same network as Olares, its status will show `Restarting` and will return to `Olares running` in approximately 5–8 minutes.
  ::: tip Note
  If you restart from a network outside Olares, the **My Olares** card will be inaccessible during the restart and will return to normal after startup is complete.
  :::

![Restart or shut down Olares remotely](/images/manual/larepass/device-control1.png#bordered)

### Modify network settings
:::tip Same network required
Make sure your phone and Olares are connected to the same network.
:::
:::tip Wired network recommended
To ensure an optimal and stable connection, we recommend using a wired network whenever possible.
:::

You can switch your Olares device's connection between its wired and Wi-Fi network options.

1. On the **Olares management** page, tap **Wi-Fi configuration**.

2. A list of available networks will appear. Tap the network you want to switch to:
   - To switch to Wi-Fi, tap the desired Wi-Fi network.
   - To switch to wired, tap the Wired network option.
   ![Modify network settings](/images/manual/larepass/switch-wifi1.png#bordered)

3. If you selected a password-protected Wi-Fi network, enter the password and tap **Confirm**.

4. Olares will begin switching to the new network. The transition takes approximately 5 minutes. The Olares status will change to `IP changing` before it reverts to `Olares Running`.

### Restore Olares to factory settings
:::warning Proceed with caution
This will permanently delete all your accounts info and data. Proceed with caution.
:::
:::tip Same network required
Make sure your phone and Olares are connected to the same network.
:::

This action will reset your device to the prepare phase, where you can scan the LAN to re-install and activate Olares.

1. On the **Olares management** page, tap **Restore to factory settings**.

2. Review the risk prompt, and tap **Restore to factory settings**.
   ![Review risk prompt](/images/manual/larepass/review-risk-prompt.png#bordered)

3. Enter your local LarePass lock screen password, and tap **Confirm**. If you haven't set a local password, you will be prompted to set one first.
   ![Enter local unlock password](/images/manual/larepass/enter-password-to-uninstall.png#bordered)

4. Wait for the reset to complete. Once finished, you will return to the account activation page.

### Reset SSH password <Badge type="tip" text="Olares One Only" />
:::warning Required action
This dialog will continue to appear until you reset the password. While you can reset later from the [My hardware](../olares/settings/my-olares.md#reset-ssh) settings page, we strongly recommend doing it immediately.
:::
After you activate **Olares One** with LarePass, a **Reset SSH Password** dialog will appear automatically. Use it to change the default SSH password and prevent unintended SSH access.

![Reset SSH Password](/images/manual/larepass/change-ssh-pw.png#bordered)

1. In the **Reset SSH Password** dialog, enter a new password. Make sure it meets all strength requirements.
2. Click **Confirm**.

## FAQ
### How do I reconnect Olares after moving it to a new network?
If your Olares device has been moved to a different network, it will become temporarily unavailable. To reconnect it:

1.  Connect the Olares device to your new wired network and power it on.
2.  Ensure your phone is connected to the same network's Wi-Fi.
3.  Open LarePass on your phone and go to **Settings** > **My Olares** > **System** > **Olares management** page.
4.  LarePass will automatically scan for the Olares device within the network. When found, Olares will appear as `IP changing` in LarePass.
5.  Once the IP update finishes, the status will revert to `Olares running`. This process may take 5–10 minutes.

### I moved the Olares device to a new network that has no wired connection. How do I connect it to Wi-Fi?
If your activated Olares device is moved to a new location without a wired connection, it will show as offline.

To reconnect it, look for the **Bluetooth network setup** option. This feature allows you to use your phone's Bluetooth to find the Olares and connect it to the new Wi-Fi network.

This process may take longer than the initial activation as the network switch needs to update all existing services on the device.