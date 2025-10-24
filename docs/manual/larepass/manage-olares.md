# Manage Olares with LarePass

The LarePass app allows you to easily manage your Olares device. You can upgrade Olares，monitor system status, manage network connections, perform remote controls, and check device information from your phone.

This guide walks you through the core management features available on LarePass.

## Prerequisites

Before you begin, ensure the following:

- You have a valid Olares ID and an activated Olares device.
- Your Olares device is powered on and connected to a network.
- Your current account has administrator permissions.

## Access Olares management

The **Olares management** page is the central hub for using LarePass to manage your Olares device. To access **Olares management**:

![Olares management](/images/manual/larepass/olares-management.png#bordered)

1. Open LarePass app and go to **Settings**.
2. In the **My Olares** card, tap **System** to enter the **Olares management** page.

## Upgrade Olares

:::warning Olares admin required
Only Olares admin can perform system updates. Updates will apply to all members within the same Olares cluster.
:::

When a new Olares version is available, you will see a "New version found" prompt under **Olares management** > **System update**. To install a system update:

![Upgrade Olares](/images/manual/larepass/olares-upgrade.png#bordered)

1. On the **Olares management** page, tap **System update**.

2. On the **System update** page, confirm the available version in the **New version** field, then click **Upgrade**.

3. In the pop-up dialog, choose your upgrade method and tap **Confirm**:

   - **Download only**: Olares will only download the update package. After the download completes, click **Upgrade** on the **System update** page to start the installation. You can continue using Olares during the download process. 

    - **Download and upgrade**: Olares will immediately download and install the update package. The system will be temporarily unavailable during the upgrade process.

4. Wait for the upgrade to finish. You will see a success message, and Olares will automatically resume normal operation.

## Remote device control

In the upper-right corner of the Olares management page, tap the <i class="material-symbols-outlined">power_settings_new</i> icon to access remote control options:

 ![Device control](/images/manual/larepass/device-control.png)

- **Restart device** – Your Olares will restart. If your phone is on the same network as Olares, its status will show `Restarting` and will return to `Olares running` in approximately 5–8 minutes.
- **Remote shutdown** – Your Olares will power off. If your phone is on the same network as Olares, its status shows `Powered off`.  Remote operations are unavailable after shutdown, and you must turn the device back on manually.

::: tip Note
If you restart from a network outside Olares, the **My Olares** card will be inaccessible during the restart and will return to normal after startup is complete.
:::

## Configure network

Tap **Wi-Fi configuration** to view or modify the current network settings of your Olares device.

:::tip Same network required
Make sure your phone and Olares are connected to the same network.
:::

### Switch from wired to wireless network

If Olares was activated over wired Ethernet, you can switch it to the Wi-Fi on the same network:

![Wi-Fi switch](/images/manual/larepass/switch-wifi.jpg)

1. Tap the **Wi-Fi configuration** option to enter the network selection page.
2. Tap the Wi-Fi network from the list. If the network is password-protected, enter the password and tap **Confirm**.
3. Once connected, the network switches to Wi-Fi automatically. The transition takes approximately 5 minutes. The Olares status will change to `IP changing` before it reverts to `Olares Running`.

You can switch back to the wired network following the same steps.

::: tip Wired network recommended
To ensure an optimal and stable connection, we recommend using a wired network whenever possible.
:::

### Update IP address

If your Olares device moves to a different network:

1. Connect the Olares device to your wired network via Ethernet and power it on. Ensure your phone is connected to the Wi-Fi for that same network.
2. Open LarePass on your phone and go to **Settings** > **My Olares** > **System** > **Olares management** page.
3. LarePass will automatically scan Olares device within the network. When found, Olares will appear as `IP changing` in LarePass.
4. Once IP update finishes, the status will revert to `Olares running`. This process may take 5–10 minutes.

### Set Wi-Fi via Bluetooth

If your Olares cannot connect to a wired network during activation, or if it's connected to a wired network different from your phone, activation with LarePass may fail, and you cannot perform device management tasks that require the same network (such as Wi-Fi configuration or a factory reset). In this case, use the Bluetooth network setup feature to connect Olares to your phone's Wi-Fi network.
 ![Bluetooth network](/images/manual/larepass/bluetooth-network.png)

1. On the **Olares not found** page, tap the **Bluetooth network setup** option. LarePass will use your phone's Bluetooth to scan for the nearby Olares device.
2. When your device appears in the list, tap **Network setup**.
3. Select the Wi-Fi network your phone is currently connected to. If it's password-protected, enter the password and tap **Confirm**.

4. Olares will start switching the network.Once the process is complete, a success message will appear. If you return to the Bluetooth network setup page, you'll see that Olares' IP address has changed to your phone's Wi-Fi subnet. 
   ::: tip Note
   The process takes longer if your Olares was activated earlier as the network switch affects more services.
   :::
5. Go back to the device scan page and tap **Discover nearby Olares** to find your device. You can now proceed with [device activation](activate-olares.md) or device management as instructed in this document.

## View device information

On the **Olares management** page, tap the device information area at the top to access device details, including:

- Hardware details 
- System version 
- Current network connection status, including internal and external IP addresses.

## Uninstall Olares

:::tip Same network required
Make sure your phone and Olares are connected to the same network.
:::

This will reset your device to the prepared installation phase, where you can scan the LAN to re-install and activate Olares. 

![Uninstall Olares](/images/manual/larepass/restore-to-factory.png)

::: warning Proceed with caution
This will permanently delete all your accounts info and data. Proceed with caution.
:::

1. On the **Olares management** page, tap **Restore to factory settings**.

2. Review the risk prompt and enter your local LarePass lock screen password. If not set, you will be prompted to set one first.

3. Wait for the uninstallation to complete. You will return to the Olares login screen when it's done.

## Reset SSH Password <Badge type="tip" text="Olares One Only" />

After you activate **Olares One** with LarePass, a **Reset SSH Password** dialog will appear automatically. Use it to change the default SSH password and prevent unintended SSH access.

![Reset SSH Password](/images/manual/larepass/change-ssh-pw.png)

1. In the **Reset SSH Password** dialog, enter a new password. Make sure it meets all strength requirements.
2. Click **Confirm**.

:::warning Required action
This dialog will continue to appear until you reset the password. While you can reset later from the [My hardware](../olares/settings/my-olares.md#reset-ssh) settings page, we strongly recommend doing it immediately. 
:::


 


