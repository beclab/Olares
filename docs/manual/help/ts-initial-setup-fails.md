---
outline: [2, 3]
description: Troubleshoot Olares One initial setup failure where installation stops at about 9% and prompts for uninstall or reinstall.
---

# Olares One initial setup fails (9% error)

Use this guide when your Olares One fails during the initial setup process with the installation stopping at around 9% and prompting you to uninstall or reinstall.

## Condition

- Device has completed initial pairing via Bluetooth or the LarePass app.
- Successfully connected to Wi-Fi network.
- Installation progress stops at about 9% with an error message indicating installation failed and suggesting you to uninstall and try again.
- Multiple reinstallation attempts fail with the same issue.
- Restarting the device does not fix the issue.

## Cause

Olares One devices are shipped with the default timezone set to East Eight Time (Beijing time). If you start the installation process immediately after powering on the device, certain components that rely on timestamp validation might fail due to incorrect system time.

This issue usually occurs during first boot, before the device has had a chance to automatically synchronize time via NTP.

## Solution

Follow these steps to uninstall the incomplete installation and then re-activate the device.

### Step 1: Attempt SSH connection

1. Follow instructions [on this page](/one/access-terminal-ssh.md#method-2-access-via-ssh) to access Olares One via SSH.
2. If the connection is successful, skip to [Step 3](#step-3-run-the-uninstall-command).
3. If the connection times out or is refused, proceed to **Step 2**.

### Step 2: Log in locally

When the remote SSH access is blocked or fails, you must log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares One.
2. Power on Olares One and wait for the system to boot. A text-based login prompt is displayed on your screen automatically:

    ```text
    olares login:
    ```
3. Log in with the username `olares` and the SSH password you obtained in **Step 1**.

### Step 3: Run the uninstall command

1. Once logged in, type the following command and then press **Enter**. This command removes all installed components and data, restoring the device to the unactivated state.

```bash
sudo olares-cli uninstall
```
2. When the uninstallation is completed, the device automatically reboots.

### Step 4: Reinstall and activate using LarePass

:::tip Before reinstallation
To ensure accurate time synchronization, let the device remain powered on for a few minutes before reinstalling, allowing it to automatically calibrate time via NTP.
:::

1. Open the LarePass app on your mobile and restart the device discovery and activation process.
2. Follow the app instructions to complete Wi-Fi configuration, account binding, and other steps. The installation should now pass the 9% mark and complete successfully.


