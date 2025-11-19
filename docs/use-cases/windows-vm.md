---
outline: [2, 3]
description: A comprehensive guide to installing and running a Windows virtual machine on Olares. Learn how to configure initial credentials, connect via browser-based VNC or Microsoft Remote Desktop (RDP), and transfer files between your computer and the VM.
---

# Run a Windows VM on your Olares device

Olares lets you run a full Windows virtual machine directly on your device, giving you a personal, always-available Windows environment accessible from macOS, Windows, or Linux.

This guide walks you through installing the Windows VM, enabling secure networking, and connecting using Remote Desktop for the best experience.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up the Windows VM on your Olares device.
- Access the Windows VM using the browser-based VNC viewer or Microsoft Remote Desktop (RDP).
- Change your Windows login password from inside the VM.
- Transfer files seamlessly between your computer and the Windows VM.

## Prerequisites

Before you begin, make sure:
- LarePass desktop client installed on your Mac or PC.
- Olares ID imported into the LarePass client.

:::tip About LarePass VPN
- For the best experience (RDP), **LarePass VPN is required**.  
You can perform the initial setup via the browser (VNC) without VPN, but for smooth daily usage and file transfer, you must enable the VPN on the device running the Remote Desktop client.
- Enable LarePass VPN on the same device where you run the Microsoft Remote Desktop client.

Learn how to enable VPN:   [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass)  
:::

## Install and configure Windows VM
Windows is available as an app in the Olares Market.

### Install Windows
1. Open the **Market** app in your Olares web interface.
2. Use the search bar and type "Windows".
3. Click **Get**, then click **Install**.
    ![Install Windows](../public/images/manual/use-cases/win-install.png#bordered)
4. When prompted, set your Windows login credentials:
    - `WINDOWS_USERNAME`: Create a username for Windows access.
    - `WINDOWS_PASSWORD`: Set the corresponding password.
    ![Set environment variables](../public/images/manual/use-cases/win-set-env-var.png#bordered)

  These credentials are required for your first login.  
  You can change them later inside Windows.

5. Wait a few minutes for the installation and initialization to complete.

### Launch Windows from Launchpad

Once installation finishes, open Windows from Launchpad to start the VM for the first time.

Olares will automatically download and install the Windows 11 system image. This file is approximately 5.5 GB, so the process may take several minutes depending on your network speed.

## Connect via browser-based VNC viewer

You can open Windows 11 from Launchpad. The Windows VM opens in your browser using a VNC connection.

:::tip About VNC access
VNC is designed for:
- Initial setup  
- Troubleshooting or emergency access 

For daily use, we strongly recommend switching to **Microsoft Remote Desktop (RDP)** for a fast, smooth, native-like Windows experience.
:::

## Connect via Remote Desktop (RDP)

To use RDP, first [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

When the VPN connection status shows `DERP`, `P2P`, or `Intranet`, the secure network is active and ready for remote access.

This guide demonstrates connecting from macOS, but the process is the same on Windows.

### Install the Remote Desktop client

**On Windows**:
No installation needed since the Remote Desktop client is built in.

**On other platforms**:
- Install the official Windows App from your platform's app store:
    - **macOS**: Mac App Store  
    - **iOS**: App Store  
    - **Android**: Google Play Store  
- You can also find the download link on the Windows app page inside Olares Market.
    ![Install Windows](../public/images/manual/use-cases/win-app-on-mac.png#bordered)

### Find your Windows VM address

You need your VM's address to connect via RDP.

There are two ways to find it. Choose whichever is more convenient for you.

<tabs>
<template #From-the-VNC-browser-URL>

1. In your Olares web interface, open Windows from Launchpad.
2. In the browser URL bar, copy the main domain:

    - Remove `https://`
    - Remove slashes `/`
    - Remove anything after your Olares ID
3. Append the RDP port: `46879`
</template>
<template #From-the-Olares-web-interface>

1. In your Olares web interface, go to **Settings** > **Application** > **Windows 11**> **Entrances** > **Set up endpoint**.
2. Copy the domain shown next to **Endpoint** (remove `https://`)
3. Append the RDP port: `46879`
</template>
</tabs>

Your Windows VM address will look like `7e89d2a1.laresprime.olares.com:46879`.

### Add your Windows VM as an RDP connection

With your VM address ready, you're now set to add the Windows VM to your Remote Desktop client.

1. Open the **Windows App** / **Microsoft Remote Desktop** on your device.
2. Click the **＋** icon and select **Add PC**.
3. In **PC name**, enter your Windows VM address you get from the previous step.
    ![Add PC](../public/images/manual/use-cases/win-add-pc.png#bordered)
4. Click **Add**.
5. Double click your saved PC entry, or click **⋯** and choose **Connect**.
    ![Connect to PC](../public/images/manual/use-cases/win-connect-device.png#bordered)
6. When prompted, enter the `WINDOWS_USERNAME` and `WINDOWS_PASSWORD` you created earlier.
    ![Log in to PC](../public/images/manual/use-cases/win-log-in.png#bordered)
7. If a security warning appears, click **Continue**.
    ![Continue to log in](../public/images/manual/use-cases/win-confirm-connect.png#bordered)

You are now connected to your Windows VM via RDP.
![Windows VM](../public/images/manual/use-cases/win-vm-interface.png#bordered)

## Change your Windows VM password (optional)

You can update your Windows login password directly from inside the VM:
1. Click the search bar in the Windows taskbar and type "password".  
2. Select **Change your password**.  
    ![Change your password](../public/images/manual/use-cases/win-change-pw.png#bordered)
3. Click **Change** to set your new password.
    ![Set new password](../public/images/manual/use-cases/win-set-pw.png#bordered)

Your updated password is used for both VNC and RDP logins.

## Transfer files between your computer and Windows

RDP supports clipboard-based file transfers.

You can: 
- Copy any file on your Mac or PC.
- Paste it directly into the Windows VM.

The file appears immediately in Windows and is ready to use.

## Disconnect from the Windows VM

To end your RDP session, simply close the RDP window.

The Windows VM continues running on your Olares device and is always ready for you to reconnect.

---
Your Olares device is now a fully functional Windows machine you can access securely from anywhere using the LarePass private network.