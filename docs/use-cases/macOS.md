---
outline: deep
description: Run macOS as a virtual machine on Olares. Learn how to install macOS from the Market, configure initial settings, and connect via browser-based VNC or VNC client.
head:
  - - meta
    - name: keywords
      content: Olares, macOS, virtual machine, VNC, macOS VM
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-03-25"
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

## Prerequisites

Before you begin, make sure:
- You have Olares admin privileges.
- [LarePass is installed](../manual/larepass/index.md) on your device.
- Your Olares device meets the minimum requirements:
  - **CPU**: 4 cores
  - **Memory**: 6GB RAM minimum
  - **Disk**: 128GB free space minimum

## Install and configure macOS VM

macOS is available as an app in the Olares Market.

### Install macOS

1. Open Market and search for "macOS".
2. Click **Get**, then click **Install**.
3. When prompted, set environment variables:
    - **VERSION:** Select your preferred macOS version from the dropdown list.
    - **DISK_SIZE:** Allocate disk space for macOS (e.g., `128G`).

    <!-- ![Set environment variables](/images/manual/use-cases/macos-set-env-var.png#bordered){width=70%} -->

4. Wait for the installation and initialization to complete.

### Format the virtual disk

When you first launch macOS, you'll see the macOS Recovery screen.

<!-- ![macOS Recovery screen](/images/manual/use-cases/macos-recovery.png#bordered) -->

1. Select **Disk Utility** from the menu, then click **Continue**.

   <!-- ![Select Disk Utility](/images/manual/use-cases/macos-select-disk-utility.png#bordered) -->

2. In Disk Utility, select **Apple Inc. VirtIO Block Media** (choose the one with the largest capacity).

   <!-- ![Select VirtIO Block Media](/images/manual/use-cases/macos-select-virtio.png#bordered) -->

3. Click **Erase** in the toolbar.

   <!-- ![Click Erase](/images/manual/use-cases/macos-click-erase.png#bordered) -->

4. Configure the format:

    - **Name:** Enter any name you prefer (e.g., "Macintosh HD").
    - **Format:** Select **APFS**.

   <!-- ![Configure format](/images/manual/use-cases/macos-configure-format.png#bordered) -->

5. Click **Erase** to format the disk.

   <!-- ![Erase disk](/images/manual/use-cases/macos-erase-disk.png#bordered) -->

6. Once complete, click **Done**, then close Disk Utility to return to the main menu.

   <!-- ![Disk formatted](/images/manual/use-cases/macos-disk-formatted.png#bordered) -->

### Install macOS system

1. From the main menu, select **Reinstall macOS**, then click **Continue**.

   <!-- ![Reinstall macOS](/images/manual/use-cases/macos-reinstall.png#bordered) -->

2. Accept the license agreement.
3. Select the disk you just formatted, then click **Continue**.
4. Wait for the installation to complete. This might take 20-40 minutes depending on your network speed and hardware.

   <!-- ![macOS installation progress](/images/manual/use-cases/macos-installing.png#bordered) -->

### Complete initial setup

After the system installation finishes:

1. Select your region.
2. Select your preferred language.
3. When you reach **Migration Assistant**, select **Not Now** to skip migrating data from another Mac.
4. When prompted for **Apple ID**, select **Set Up Later**.
5. Set up a username and password for the macOS account. For the remaining setup steps, you can skip or accept the defaults.

   <!-- ![macOS desktop](/images/manual/use-cases/macos-desktop.png#bordered) -->

## Access the macOS VM

You can access your VM in two ways:
- [**Browser:**](#method-1-access-from-the-browser) for quick access without additional software
- [**VNC Client:**](#method-2-access-using-a-vnc-client) for better performance and features

:::tip Browser vs. VNC Client
- **Browser**: Use for quick access or when you cannot install software. Best for initial setup and troubleshooting.
- **VNC Client**: Use for daily work. Offers better performance, smoother display, and more stable connections.
:::

### Method 1: Access from the browser

Open the macOS app from Launchpad to launch the VM directly in your browser. 

<!-- ![macOS VM in browser](/images/manual/use-cases/macos-browser.png#bordered) -->

You can close the browser tab when you're done. The macOS VM continues running on your Olares device and remains ready for you to reconnect.

### Method 2: Access using a VNC client

#### Locate port number for macOS

:::warning Multiple macOS instances
Each macOS instance uses a unique port. If you have cloned the macOS app, ensure you check the **ACLs** section for the specific instance you want to access.
:::

1. Open Settings and navigate to **Application** > **macOS**.
2. Under **Permissions**, click **ACLs**.
3. Note the port number listed in the **Port** column. You will need this for the connection step.

   <!-- ![Locate port number](/images/manual/use-cases/macos-port-number.png#bordered) -->

#### Connect via VNC client

<Tabs>
<template #macOS>

1. Open LarePass and enable VPN on your device.

   ![Enable VPN on LarePass desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

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

1. Open LarePass and enable VPN on your device.

   ![Enable VPN on LarePass desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

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

To disconnect from the macOS VM, close the VNC viewer window. The macOS VM continues running on your Olares device and remains ready for you to reconnect.

## FAQ

### The macOS VM shows a blank screen or no desktop

The browser might have suspended the VNC connection due to inactivity to conserve system resources. Refresh the page or click **Connect** to restore the session.

### Can I use my Apple ID with this VM?

While you can sign in with an Apple ID during setup, some Apple services might not function correctly in a virtualized environment. For best results, use local accounts or skip Apple ID setup.

### What macOS versions are supported?

Currently supported versions:
- macOS 14 (Sonoma)
- macOS 13 (Ventura)
- macOS 12 (Monterey)
- macOS 11 (Big Sur)

## Learn more

- [dockur/macos GitHub repository](https://github.com/dockur/macos)
- [Run a Windows VM on your Olares device](./windows.md)
