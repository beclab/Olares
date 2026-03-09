---
outline: deep
description: Run macOS as a virtual machine on Olares. Learn how to install macOS from the Market, configure initial settings, and connect via browser-based VNC or VNC client.
app_version: "1.0.8"
doc_updated: "2025-03-09"
---

# Run macOS on your Olares device

Olares lets you run macOS as a virtual machine directly on your device. This gives you access to macOS-specific applications and workflows from any computer with a web browser or VNC client.

:::info System capabilities
- This VM provides access to macOS applications that require the Apple ecosystem.
- Performance depends on your hardware's CPU capabilities. GPU acceleration is not available.
:::

This guide walks you through installing macOS, completing the initial setup, and connecting to your VM.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and configure the macOS VM on your Olares device.
- Access the macOS VM using the browser-based VNC viewer or a VNC client.
- Complete the macOS initial setup and system configuration.

## Before you begin

- Olares admin privileges
- [LarePass installed](../manual/larepass/index.md) on your device
- Ensure your Olares device meets the minimum requirements:
  - **CPU**: 4 cores
  - **Memory**: 6GB RAM minimum
  - **Disk**: 128GB free space minimum

## Install and configure macOS VM

macOS is available as an app in the Olares Market.

### Install macOS

1. Open **Market**, and search for "macOS".
2. Click **Get**, then click **Install**.
3. When prompted, set environment variables:
    - **VERSION:** Select your preferred macOS version from the dropdown list (macOS 11, 12, 13, or 14).
    - **DISK_SIZE:** Allocate disk space for macOS (e.g., `128G`).

    <!-- ![Set environment variables](/images/manual/use-cases/macos-set-env-var.png#bordered){width=70%} -->

4. Wait for the installation and initialization to complete.

### Format the virtual disk

When you first launch macOS, you'll see the macOS Recovery screen.

1. Select **Disk Utility** from the menu, then click **Continue**.
2. In Disk Utility, select **Apple Inc. VirtIO Block Media** (choose the one with the largest capacity).
3. Click **Erase** in the toolbar.
4. Configure the format:

    - **Name:** Enter any name you prefer (e.g., "Macintosh HD").
    - **Format:** Select **APFS**.
5. Click **Erase** to format the disk.
6. Once complete, click **Done**, then close Disk Utility to return to the main menu.

### Install macOS system

1. From the main menu, select **Reinstall macOS**, then click **Continue**.
2. Accept the license agreement.
3. Select the disk you just formatted, then click **Continue**.
4. Wait for the installation to complete. This might take 20-40 minutes depending on your network speed and hardware.

### Complete initial setup

After the system installation finishes:

1. **Select your region** and **preferred language**.
2. When you reach **Migration Assistant**, select **Not Now** to skip migrating data from another Mac.
3. When prompted for **Apple ID**, select **Set Up Later**.
4. Create your account:

    a. Set up a username and password for the macOS account.
    
    b. Complete the remaining setup steps.

## Access the macOS VM

You can access your VM in two ways:
- [**Browser:**](#method-1-access-from-the-browser) for quick access without additional software
- [**VNC Client:**](#method-2-access-using-a-vnc-client) for better performance and features

### Method 1: Access from the browser

Open the macOS app from Launchpad to launch the VM directly in your browser.

:::info
The browser-based viewer provides immediate access without requiring any additional software installation. It is ideal for initial setup and quick tasks. However, it might feel less responsive than a dedicated VNC client.
:::

### Method 2: Access using a VNC client

A dedicated VNC client provides a significantly better experience:
- **Better performance**: Smoother screen updates and lower latency
- **Clipboard sharing**: Copy and paste text and files between your computer and the VM
- **Persistent connections**: Maintain stable connections for extended sessions
- **Additional features**: Screen scaling, quality adjustment, and multi-monitor support

:::tip Browser vs. VNC Client
- **Browser**: Use for quick access or when you cannot install software. Best for initial setup and troubleshooting.
- **VNC Client**: Use for daily work, better performance, and full feature support including clipboard and file transfer.
:::

#### Locate port number for macOS

:::warning Multiple macOS instances
Each macOS instance uses a unique port. If you have cloned the macOS app, ensure you check the **ACLs** section for the specific instance you want to access.
:::

1. Open **Settings**, and navigate to **Application** > **macOS**.
2. Under **Permissions**, click **ACLs**.
3. Note the port number listed in the **Port** column. You will need this for the connection step.

#### Connect via VNC client

<Tabs>
<template #macOS>

1. [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

   When the VPN connection status shows **P2P** or **Intranet**, the secure network is active.

2. Install VNC Viewer:
   ```bash
   brew install --cask vnc-viewer
   ```

3. Open VNC Viewer and create a new connection:

   a. Click **File** > **New Connection**.

   b. Enter the address and port number you obtained from the ACLs section.

   c. Save the connection.

4. Double-click the saved connection to connect.

</template>
<template #Windows>

1. [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

   When the VPN connection status shows **P2P** or **Intranet**, the secure network is active.

2. Download and install [RealVNC Viewer](https://www.realvnc.com/en/connect/download/viewer/).

3. Open RealVNC Viewer and create a new connection:

   a. Click **File** > **New Connection**.

   b. Enter the address and port number you obtained from the ACLs section.

   c. Save the connection.

4. Double-click the saved connection to connect.

</template>
</Tabs>

When prompted, enter the username and password you created during the macOS setup.

## Disconnect from the macOS VM

To end your session:
- **Browser:** Simply close the browser tab.
- **VNC Client:** Close the VNC viewer window.

The macOS VM continues running on your Olares device and remains ready for you to reconnect.

## FAQ

### The macOS VM shows a blank screen or no desktop

The browser might have suspended the VNC connection due to inactivity to conserve system resources. Refresh the page or click **Connect** to restore the session.

### Can I use my Apple ID with this VM?

While you can sign in with an Apple ID during setup, some Apple services may not function correctly in a virtualized environment. For best results, use local accounts or skip Apple ID setup.

### What macOS versions are supported?

Currently supported versions:
- macOS 14 (Sonoma)
- macOS 13 (Ventura)
- macOS 12 (Monterey)
- macOS 11 (Big Sur)

## Learn more

- [dockur/macos GitHub repository](https://github.com/dockur/macos)
- [Run a Windows VM on your Olares device](./windows.md)
